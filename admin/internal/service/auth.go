// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/contful/contful/admin/pkg/uid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/contful/contful/admin/internal/audit"
	"github.com/contful/contful/admin/internal/config"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
)

// P1-003: Token Rotation — 每次 Refresh 时撤销旧 refresh token，生成新的
// P1-004: 密码强度验证 — 注册时强制复杂度要求
// P2-002: 账户锁定机制 — 连续失败登录后锁定账户
// P2-004: 登录失败记录审计日志

// ErrAccountLocked 账户因多次登录失败被临时锁定
var ErrAccountLocked = errors.New("账户已被临时锁定，请稍后再试")

// AuthService 认证服务
type AuthService struct {
	userRepo        *repository.UserRepository
	siteRepo        *repository.SiteRepository
	configRepo      *repository.SystemConfigRepository
	auditRepo       *repository.AuditRepository
	configSvc       *ConfigService // 用于 AuditLog 数据签名
	mfaService      *MFAService    // MFA 双因子认证（可空，启用时注入）
	redis           *redis.Client
	jwtSecret       []byte
	accessTTL       time.Duration
	// P2-002: 账户锁定默认值（实际从 system_config 动态读取）
}

func NewAuthService(
	userRepo *repository.UserRepository,
	siteRepo *repository.SiteRepository,
	configRepo *repository.SystemConfigRepository,
	auditRepo *repository.AuditRepository,
	redisClient *redis.Client,
	jwtSecret string,
	configSvc *ConfigService,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		siteRepo:        siteRepo,
		configRepo:      configRepo,
		auditRepo:       auditRepo,
		configSvc:       configSvc,
		redis:           redisClient,
		jwtSecret:     []byte(jwtSecret),
		accessTTL:     15 * time.Minute,
	}
}

// SetMFAService 注入 MFAService（避免循环依赖，setup 后调用）
func (s *AuthService) SetMFAService(mfaSvc *MFAService) {
	s.mfaService = mfaSvc
}

// JWT Claims
type JWTClaims struct {
	UserID            uid.UID `json:"user_id"`
	Email             string    `json:"email"`
	IsSuperAdmin      bool      `json:"is_super_admin"`
	MFASetupRequired  bool      `json:"mfa_setup_required,omitempty"`
	jwt.RegisteredClaims
}

// P1-004: 密码复杂度验证（导出供 handler 层使用）
var (
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters")
	ErrPasswordNoUppercase   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit       = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecialChar = errors.New("password must contain at least one special character")
)

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return ErrPasswordNoUppercase
	}
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return ErrPasswordNoLowercase
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return ErrPasswordNoDigit
	}
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return ErrPasswordNoSpecialChar
	}
	return nil
}

// P2-002: 账户锁定 Redis key
func loginAttemptsKey(email string) string {
	return "login_attempts:" + email
}

// isAccountLocked 检查账户是否被锁定（从 system_config 动态读取阈值）
func (s *AuthService) isAccountLocked(ctx context.Context, email string) (bool, time.Duration, error) {
	key := loginAttemptsKey(email)
	count, err := s.redis.Get(ctx, key).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, 0, err
	}

	maxAttempts := s.configRepo.GetInt(ctx, "login_max_attempts", 5)
	if count >= maxAttempts {
		ttl := s.redis.TTL(ctx, key).Val()
		return true, ttl, nil
	}
	return false, 0, nil
}

// incrementLoginAttempts 增加失败计数（从 system_config 动态读取锁定时长）
func (s *AuthService) incrementLoginAttempts(ctx context.Context, email string) error {
	key := loginAttemptsKey(email)
	lockMinutes := s.configRepo.GetInt(ctx, "login_lock_duration", 30)
	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, time.Duration(lockMinutes)*time.Minute)
	_, err := pipe.Exec(ctx)
	return err
}

// clearLoginAttempts 清除失败计数（登录成功时）
func (s *AuthService) clearLoginAttempts(ctx context.Context, email string) error {
	return s.redis.Del(ctx, loginAttemptsKey(email)).Err()
}

// Register 注册新用户
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest, ip string) (*model.UserResponse, error) {
	// P1-004: 密码强度检查
	if err := validatePasswordStrength(req.Password); err != nil {
		return nil, err // 返回具体错误
	}

	// 检查邮箱是否已存在
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, repository.ErrUserAlreadyExists
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.SystemUser{
		ID:           uid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
		Status:       model.UserStatusActive,
		IsSuperAdmin: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 审计日志
	s.createAuditLog(ctx, user.ID, nil, "register", "user", user.ID, model.AuditLevelInfo, model.AuditTypeAuth, ip)

	return s.toUserResponse(user), nil
}

// Login 登录
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest, ip string) (interface{}, error) {
	// P2-002: 检查账户是否被锁定
	locked, ttl, err := s.isAccountLocked(ctx, req.Email)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("failed to check account lock status")
	}
	if locked {
		log.Warn().
			Str("email", req.Email).
			Str("remaining", ttl.String()).
			Msg("login blocked: account locked")
		return nil, ErrAccountLocked
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// P2-002/P2-004: 邮箱不存在时也增加计数 + 记录审计，防止枚举攻击
			s.incrementLoginAttempts(ctx, req.Email)
			s.createAuditLogLoginFail(ctx, req.Email, ip, "user not found")
			return nil, repository.ErrInvalidPassword // 返回相同错误，防止枚举
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// P2-002: 密码错误，增加失败计数
		s.incrementLoginAttempts(ctx, req.Email)
		// P2-004: 记录审计日志
		s.createAuditLogLoginFail(ctx, req.Email, ip, "invalid password")
		return nil, repository.ErrInvalidPassword
	}

	// P2-002: 登录成功，清除失败计数
	s.clearLoginAttempts(ctx, req.Email)

	// 检查用户状态
	switch user.Status {
	case model.UserStatusInactive:
		return nil, repository.ErrUserInactive
	case model.UserStatusSuspended:
		return nil, repository.ErrUserSuspended
	}

	// MFA 检测：如果用户启用了 MFA，返回 mfa_required + mfa_token
	if user.MFAEnabled && s.mfaService != nil {
		mfaToken, err := s.generateMFAToken()
		if err != nil {
			return nil, err
		}
		if err := s.mfaService.StoreMFAPendingToken(ctx, mfaToken, MFAPendingDataFromUser(user)); err != nil {
			return nil, err
		}
		return &model.MFARequiredResponse{
			MFARequired: true,
			MFAToken:    mfaToken,
		}, nil
	}

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("failed to update last login")
	}

	// 检查系统是否强制 MFA（用户未启用 MFA 时）
	mfaEnforced := s.configRepo.GetBool(ctx, "mfa_enforced", false)

	// 正常发放 JWT（携带 MFASetupRequired 标记）
	loginResp, err := s.IssueTokens(ctx, user, ip, mfaEnforced && !user.MFAEnabled)
	if err != nil {
		return nil, err
	}

	// 审计日志
	s.createAuditLog(ctx, user.ID, nil, "login", "user", user.ID, model.AuditLevelInfo, model.AuditTypeAuth, ip)

	return loginResp, nil
}

// Refresh 刷新 Token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	// P1-003: 验证 Refresh Token
	userID, err := s.userRepo.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	// P1-003: 撤销旧 refresh token（实现轮换）
	if err := s.userRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		log.Warn().Err(err).Msg("failed to delete old refresh token during rotation")
	}

	// 查找用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	// 生成新的 Access Token
	accessToken, err := s.generateAccessToken(user, false)
	if err != nil {
		return "", "", err
	}

	// 生成新的 Refresh Token（轮换）
	newRefreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// Logout 登出
func (s *AuthService) Logout(ctx context.Context, refreshToken string, ip string) error {
	// 删除 Redis 中的 Refresh Token
	if err := s.userRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return err
	}
	return nil
}

// GetUser 获取当前用户（含可访问站点列表）
func (s *AuthService) GetUser(ctx context.Context, userID uid.UID) (*model.UserWithSites, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 查询用户可访问的站点：super_admin 返回所有活跃站点
	var sites []model.SiteBasic
	if user.IsSuperAdmin {
		active := true
		allSites, _, err := s.siteRepo.List(ctx, 1, 100, &active)
		if err != nil {
			log.Warn().Err(err).Msg("query sites for super_admin failed, returning empty")
			sites = []model.SiteBasic{}
		} else {
			sites = make([]model.SiteBasic, len(allSites))
			for i, site := range allSites {
				sites[i] = model.SiteBasic{
					ID:   site.ID,
					Name: site.Name,
					Slug: site.Slug,
				}
			}
		}
	} else {
		// 非超管用户：未来可通过站点成员表扩展，当前返回空列表
		sites = []model.SiteBasic{}
	}

	return &model.UserWithSites{
		UserResponse: *s.toUserResponse(user),
		Sites:        sites,
	}, nil
}

// ListUsers 获取用户列表
func (s *AuthService) ListUsers(ctx context.Context, page, pageSize int) (*model.PageResponse, error) {
	users, total, err := s.userRepo.List(ctx, page, pageSize, false)
	if err != nil {
		return nil, err
	}

	userResponses := make([]model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.toUserResponse(&user)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &model.PageResponse{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Items:      userResponses,
	}, nil
}

// Helper functions

func (s *AuthService) generateAccessToken(user *model.SystemUser, mfaSetupRequired ...bool) (string, error) {
	mfaRequired := false
	if len(mfaSetupRequired) > 0 {
		mfaRequired = mfaSetupRequired[0]
	}

	claims := JWTClaims{
		UserID:           user.ID,
		Email:            user.Email,
		IsSuperAdmin:     user.IsSuperAdmin,
		MFASetupRequired: mfaRequired,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) parseAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// ParseAccessTokenInternal 暴露 parseAccessToken 供 handler.GetClaims 使用（避免循环依赖）
func (s *AuthService) ParseAccessTokenInternal(tokenString string) (*JWTClaims, error) {
	return s.parseAccessToken(tokenString)
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID uid.UID) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	if err := s.userRepo.StoreRefreshToken(ctx, userID, token); err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) toUserResponse(user *model.SystemUser) *model.UserResponse {
	resp := &model.UserResponse{
		ID:                user.ID,
		Email:             user.Email,
		Nickname:          user.Nickname,
		AvatarURL:         user.AvatarURL,
		Status:            user.Status,
		IsSuperAdmin:      user.IsSuperAdmin,
		MFAEnabled:        user.MFAEnabled,
		CreatedTime:       user.CreatedTime,
		PasswordChangedTime: user.PasswordChangedTime,
	}
	if user.DeletedAt.Valid {
		resp.DeletedAt = &user.DeletedAt.Time
	}
	return resp
}

// IssueTokens 生成并存储 Access Token + Refresh Token（供 AuthService 和 MFA 验证后复用）
func (s *AuthService) IssueTokens(ctx context.Context, user *model.SystemUser, ip string, mfaSetupRequired ...bool) (*model.LoginResponse, error) {
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("failed to update last login")
	}

	var mfaReq bool
	if len(mfaSetupRequired) > 0 {
		mfaReq = mfaSetupRequired[0]
	}

	accessToken, err := s.generateAccessToken(user, mfaReq)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 密码过期检查（从 system_config 读取配置）
	passwordExpireDays := s.configRepo.GetInt(ctx, "password_expire_days", 90)
	var passwordMaxAge time.Duration
	var passwordExpired bool
	var passwordExpireDaysCount *int

	if passwordExpireDays <= 0 {
		// 0 或负数表示永不过期
		passwordExpired = false
		passwordExpireDaysCount = nil
	} else {
		passwordMaxAge = time.Duration(passwordExpireDays) * 24 * time.Hour

		if user.PasswordChangedTime == nil {
			// 旧用户，从未修改过密码，使用创建时间
			elapsed := time.Since(user.CreatedTime)
			days := int(elapsed.Hours() / 24)
			passwordExpireDaysCount = &days
			passwordExpired = elapsed > passwordMaxAge
		} else {
			// 检查密码最后修改时间
			elapsed := time.Since(*user.PasswordChangedTime)
			days := int(elapsed.Hours() / 24)
			passwordExpireDaysCount = &days
			passwordExpired = elapsed > passwordMaxAge
		}
	}

	return &model.LoginResponse{
		User:              s.toUserResponse(user),
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		PasswordExpired:    passwordExpired,
		PasswordExpireDays: passwordExpireDaysCount,
		MFASetupRequired:   mfaReq,
	}, nil
}

// generateMFAToken 生成 mfa_pending token（随机 hex 字符串）
func (s *AuthService) generateMFAToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// MFAPendingDataFromUser 从 SystemUser 构造 MFAPendingData
func MFAPendingDataFromUser(user *model.SystemUser) MFAPendingData {
	return MFAPendingData{
		UserID:       user.ID.String(),
		Email:        user.Email,
		IsSuperAdmin: user.IsSuperAdmin,
	}
}

func (s *AuthService) createAuditLog(ctx context.Context, userID uid.UID, siteID *uid.UID, action, resourceType string, resourceID uid.UID, level model.AuditLevel, category model.AuditType, ip string) {
	go func() {
		// 使用 background context 避免 request context 被取消导致写入失败
		bgCtx := context.Background()
		auditLog := &model.AuditLog{
			ID:           uid.New(),
			UserID:       &userID,
			SiteID:       siteID,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   &resourceID,
			Level:        level,
			Category:     category,
			IPAddress:    ip,
		}
		// 注入签名密钥到 context
		if s.configSvc != nil && siteID != nil {
			if key, _ := s.configSvc.GetAuditSigningKey(); key != "" {
				bgCtx = audit.WithSigningKey(bgCtx, key)
			}
		}
		// 注入 Hasher（国密模式时自动使用 SM3）
		if provider := config.GetCryptoProvider(); provider != nil {
			bgCtx = audit.WithHasher(bgCtx, provider)
		}
		if err := s.auditRepo.Create(bgCtx, auditLog); err != nil {
			log.Error().Err(err).Msg("failed to create audit log")
		}
	}()
}

// P2-004: 登录失败审计日志（userID 为 nil）
func (s *AuthService) createAuditLogLoginFail(ctx context.Context, email string, ip string, reason string) {
	go func() {
		bgCtx := context.Background()
		maskedEmail := maskEmail(email)
		auditLog := &model.AuditLog{
			ID:           uid.New(),
			UserID:       nil,
			Action:       "login_failed:" + reason, // login_failed:invalid_password
			ResourceType: "user",
			ResourceID:   nil,
			Level:        model.AuditLevelWarn,
			Category:     model.AuditTypeAuth,
			IPAddress:    ip,
		}
		// 登录失败场景暂无 site_id，跳过签名
		if err := s.auditRepo.Create(bgCtx, auditLog); err != nil {
			log.Error().Err(err).Str("email", maskedEmail).Msg("failed to create login_fail audit log")
		}
	}()
}

// maskEmail 邮箱脱敏：show@example.com → s**w@example.com
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	local := parts[0]
	domain := parts[1]
	if len(local) <= 2 {
		return "***@" + domain
	}
	return string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1]) + "@" + domain
}

func splitToken(token string) []string {
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			return []string{token[:i], token[i+1:]}
		}
	}
	return nil
}

// ValidateAccessToken 验证 Access Token 并返回 Claims
func (s *AuthService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	parts := splitToken(tokenString)
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}
	return s.parseAccessToken(parts[0])
}

// GenerateRSAToken 生成一次性 Anti-Replay Token（存储 Redis，5 分钟过期）
func (s *AuthService) GenerateRSAToken(ctx context.Context) (token, tokenID string, err error) {
	tokenID = uid.New().String()
	token = uid.New().String()
	key := "rsa_token:" + tokenID
	err = s.redis.Set(ctx, key, token, 5*time.Minute).Err()
	if err != nil {
		return "", "", fmt.Errorf("failed to store RSA token: %w", err)
	}
	return token, tokenID, nil
}

// ValidateAndConsumeRSAToken 验证并销毁一次性 Anti-Replay Token
func (s *AuthService) ValidateAndConsumeRSAToken(ctx context.Context, tokenID, token string) bool {
	key := "rsa_token:" + tokenID
	stored, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return false
	}
	if stored != token {
		return false
	}
	// 一次性使用：验证后立即删除
	s.redis.Del(ctx, key)
	return true
}
