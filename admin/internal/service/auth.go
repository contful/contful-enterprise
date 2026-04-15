package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	jwtSecret []byte
	accessTTL time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		jwtSecret: []byte(jwtSecret),
		accessTTL: 15 * time.Minute, // Access Token 15分钟
	}
}

// JWT Claims
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	IsSuperAdmin bool `json:"is_super_admin"`
	jwt.RegisteredClaims
}

// Register 注册新用户
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest, ip string) (*model.UserResponse, error) {
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

	user := &model.GlobalUser{
		ID:           uuid.New(),
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
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest, ip string) (*model.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, repository.ErrInvalidPassword // 防止邮箱枚举攻击
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, repository.ErrInvalidPassword
	}

	// 检查用户状态
	switch user.Status {
	case model.UserStatusInactive:
		return nil, repository.ErrUserInactive
	case model.UserStatusSuspended:
		return nil, repository.ErrUserSuspended
	}

	// 更新最后登录信息
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ip); err != nil {
		log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("failed to update last login")
	}

	// 生成 Access Token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// 生成并存储 Refresh Token
	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// 审计日志
	s.createAuditLog(ctx, user.ID, nil, "login", "user", user.ID, model.AuditLevelInfo, model.AuditTypeAuth, ip)

	return &model.LoginResponse{
		User:        s.toUserResponse(user),
		AccessToken: accessToken + "." + refreshToken, // accessToken.refreshToken 格式
	}, nil
}

// Refresh 刷新 Token
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	// 解析 token（refreshToken 是单独存储的，格式为 accessToken.refreshToken）
	parts := splitToken(refreshToken)
	if len(parts) != 2 {
		return "", errors.New("invalid token format")
	}

	accessTokenPart := parts[0]
	refreshTokenPart := parts[1]

	// 验证 Refresh Token
	userID, err := s.userRepo.ValidateRefreshToken(ctx, refreshTokenPart)
	if err != nil {
		return "", err
	}

	// 查找用户
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	// 生成新的 Access Token（保留旧的 refresh token）
	accessToken, err := s.generateAccessTokenWithClaims(user, accessTokenPart)
	if err != nil {
		return "", err
	}

	return accessToken + "." + refreshTokenPart, nil
}

// Logout 登出
func (s *AuthService) Logout(ctx context.Context, refreshToken string, ip string) error {
	parts := splitToken(refreshToken)
	if len(parts) != 2 {
		return nil // 无效 token 直接返回
	}

	refreshTokenPart := parts[1]
	
	// 删除 Refresh Token
	if err := s.userRepo.DeleteRefreshToken(ctx, refreshTokenPart); err != nil {
		return err
	}

	// 审计日志（从 accessToken 提取 userID）
	accessTokenPart := parts[0]
	claims, err := s.parseAccessToken(accessTokenPart)
	if err != nil {
		return nil // token 已无效
	}

	s.createAuditLog(ctx, claims.UserID, nil, "logout", "user", claims.UserID, model.AuditLevelInfo, model.AuditTypeAuth, ip)

	return nil
}

// GetUser 获取当前用户
func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*model.UserWithSites, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.UserWithSites{
		UserResponse: *s.toUserResponse(user),
		Sites:        []model.SiteBasic{}, // TODO: 查询用户关联的站点
	}, nil
}

// ListUsers 获取用户列表
func (s *AuthService) ListUsers(ctx context.Context, page, pageSize int) (*model.PageResponse, error) {
	users, total, err := s.userRepo.List(ctx, page, pageSize)
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
		Data:       userResponses,
	}, nil
}

// Helper functions

func (s *AuthService) generateAccessToken(user *model.GlobalUser) (string, error) {
	claims := JWTClaims{
		UserID:       user.ID,
		Email:        user.Email,
		IsSuperAdmin: user.IsSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generateAccessTokenWithClaims(user *model.GlobalUser, existingToken string) (string, error) {
	claims := JWTClaims{
		UserID:       user.ID,
		Email:        user.Email,
		IsSuperAdmin: user.IsSuperAdmin,
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

func (s *AuthService) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
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

func (s *AuthService) toUserResponse(user *model.GlobalUser) *model.UserResponse {
	return &model.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Nickname:     user.Nickname,
		AvatarURL:    user.AvatarURL,
		Status:       user.Status,
		IsSuperAdmin: user.IsSuperAdmin,
		CreatedAt:    user.CreatedAt,
	}
}

func (s *AuthService) createAuditLog(ctx context.Context, userID uuid.UUID, siteID *uuid.UUID, action, resourceType string, resourceID uuid.UUID, level model.AuditLevel, category model.AuditType, ip string) {
	// 异步创建审计日志
	go func() {
		auditLog := &model.AuditLog{
			ID:           uuid.New(),
			UserID:       &userID,
			SiteID:       siteID,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   &resourceID,
			Level:        level,
			Category:     category,
			IPAddress:    ip,
		}
		if err := s.auditRepo.Create(ctx, auditLog); err != nil {
			log.Error().Err(err).Msg("failed to create audit log")
		}
	}()
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
