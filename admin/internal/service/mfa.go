// Package service MFA（多因子认证）服务 — TOTP + Recovery Code
package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/crypto"
	"github.com/contful/contful/admin/internal/repository"
)

const (
	mfaPendingPrefix = "mfa_pending:"
	mfaPendingTTL    = 5 * time.Minute

	recoveryCodeGroups = 10  // 10 组恢复码
	recoveryCodeLength = 12  // 每组 12 字符（XXXX-XXXX-XXXX）
)

var (
	ErrMFAAlreadyEnabled      = errors.New("MFA is already enabled")
	ErrMFANotEnabled          = errors.New("MFA is not enabled")
	ErrMFANotSetup            = errors.New("MFA has not been set up, please call setup first")
	ErrMFAInvalidCode         = errors.New("invalid or expired TOTP code")
	ErrMFAPendingTokenInvalid = errors.New("MFA pending token is invalid or expired")
	ErrRecoveryCodeInvalid    = errors.New("recovery code is invalid or has already been used")
	ErrRecoveryCodesExhausted = errors.New("all recovery codes have been used")
)

// MFAPendingData mfa_pending Redis 中存储的数据
type MFAPendingData struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}

// MFAService MFA 双因子认证服务
type MFAService struct {
	userRepo *repository.UserRepository
	redis    *redis.Client
	crypter  crypto.Crypter
}

func NewMFAService(userRepo *repository.UserRepository, redisClient *redis.Client, crypter crypto.Crypter) *MFAService {
	return &MFAService{
		userRepo: userRepo,
		redis:    redisClient,
		crypter:  crypter,
	}
}

// Setup 生成 TOTP Secret，临时存储，返回 otpauth URI 和 QR Code URL
// 调用此接口后 mfa_enabled 仍为 false，需调用 Enable 完成激活
func (s *MFAService) Setup(ctx context.Context, userID uuid.UUID) (*model.MFASetupResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.MFAEnabled {
		return nil, ErrMFAAlreadyEnabled
	}

	// 生成 TOTP Secret（20 字节，base32 编码后 32 字符）
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Contful",
		AccountName: user.Email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	secretPlain := key.Secret()

	// 生成 10 组 Recovery Code
	recoveryCodes, err := generateRecoveryCodes(recoveryCodeGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}

	// 加密存储 TOTP Secret
	encryptedSecret, err := s.crypter.Encrypt([]byte(secretPlain))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}

	// 加密存储 Recovery Code
	codesJSON, err := marshalRecoveryCodes(recoveryCodes)
	if err != nil {
		return nil, err
	}
	encryptedCodes, err := s.crypter.Encrypt(codesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt recovery codes: %w", err)
	}

	// 写入数据库（mfa_enabled 保持 false，等待 Enable 确认）
	if err := s.userRepo.UpdateMFASecret(ctx, userID, encryptedSecret, encryptedCodes); err != nil {
		return nil, err
	}

	// 构造 QR Code URL
	otpauthURI := key.URL()
	qrCodeURL := "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" + url.QueryEscape(otpauthURI)

	return &model.MFASetupResponse{
		TOTPSecret: secretPlain,
		OTPAuthURI: otpauthURI,
		QRCodeURL:  qrCodeURL,
	}, nil
}

// Enable 验证 TOTP 码，启用 MFA，返回明文 Recovery Code（仅此次）
func (s *MFAService) Enable(ctx context.Context, userID uuid.UUID, totpCode string) (*model.MFAEnableResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.MFAEnabled {
		return nil, ErrMFAAlreadyEnabled
	}
	if user.TOTPSecret == nil {
		return nil, ErrMFANotSetup
	}

	// 解密 TOTP Secret
	secretPlain, err := s.decryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return nil, err
	}

	// 验证 TOTP（容差 ±1 步，共 3 个 30s 窗口）
	if !validateTOTP(secretPlain, totpCode) {
		return nil, ErrMFAInvalidCode
	}

	// 解密 Recovery Codes（用于返回）
	plainCodes, err := s.decryptRecoveryCodes(user.RecoveryCodes)
	if err != nil {
		return nil, err
	}

	// 设置 mfa_enabled = true
	if err := s.userRepo.UpdateMFAEnabled(ctx, userID, true); err != nil {
		return nil, err
	}

	// 收集明文恢复码
	codeStrings := make([]string, len(plainCodes))
	for i, rc := range plainCodes {
		codeStrings[i] = rc.Code
	}

	return &model.MFAEnableResponse{
		MFAEnabled:    true,
		RecoveryCodes: codeStrings,
	}, nil
}

// Disable 验证 TOTP 码，关闭 MFA
func (s *MFAService) Disable(ctx context.Context, userID uuid.UUID, totpCode string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.MFAEnabled {
		return ErrMFANotEnabled
	}

	secretPlain, err := s.decryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return err
	}

	if !validateTOTP(secretPlain, totpCode) {
		return ErrMFAInvalidCode
	}

	return s.userRepo.ClearMFA(ctx, userID)
}

// StoreMFAPendingToken 步骤 1 登录成功后写入 Redis mfa_pending token
func (s *MFAService) StoreMFAPendingToken(ctx context.Context, mfaToken string, data MFAPendingData) error {
	key := mfaPendingPrefix + mfaToken
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, key, string(b), mfaPendingTTL).Err()
}

// VerifyMFALogin 步骤 2 — 验证 mfa_token + TOTP 码，返回 user（供 AuthService 生成 JWT）
func (s *MFAService) VerifyMFALogin(ctx context.Context, mfaToken, totpCode string) (*model.SystemUser, error) {
	key := mfaPendingPrefix + mfaToken
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrMFAPendingTokenInvalid
		}
		return nil, err
	}

	var pending MFAPendingData
	if err := json.Unmarshal([]byte(val), &pending); err != nil {
		return nil, ErrMFAPendingTokenInvalid
	}

	userID, err := uuid.Parse(pending.UserID)
	if err != nil {
		return nil, ErrMFAPendingTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.MFAEnabled || user.TOTPSecret == nil {
		return nil, ErrMFANotEnabled
	}

	secretPlain, err := s.decryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return nil, err
	}

	if !validateTOTP(secretPlain, totpCode) {
		return nil, ErrMFAInvalidCode
	}

	// 验证通过，删除 mfa_pending token（一次性）
	s.redis.Del(ctx, key)

	return user, nil
}

// Recover 使用 Recovery Code 登录，一次性使用
func (s *MFAService) Recover(ctx context.Context, email, recoveryCode string) (*model.SystemUser, int, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, 0, repository.ErrUserNotFound
	}

	if !user.MFAEnabled {
		return nil, 0, ErrMFANotEnabled
	}

	if user.RecoveryCodes == nil {
		return nil, 0, ErrRecoveryCodesExhausted
	}

	codes, err := s.decryptRecoveryCodes(user.RecoveryCodes)
	if err != nil {
		return nil, 0, err
	}

	normalizedInput := strings.ToUpper(strings.TrimSpace(recoveryCode))
	found := false
	remaining := 0

	for i, rc := range codes {
		if !rc.Used && rc.Code == normalizedInput {
			now := time.Now().Format(time.RFC3339)
			codes[i].Used = true
			codes[i].UsedAt = &now
			found = true
		}
		if !codes[i].Used {
			remaining++
		}
	}

	if !found {
		return nil, 0, ErrRecoveryCodeInvalid
	}

	// 所有恢复码用完，自动关闭 MFA
	if remaining == 0 {
		log.Warn().Str("user_id", user.ID.String()).Msg("all MFA recovery codes exhausted, disabling MFA")
		if err := s.userRepo.ClearMFA(ctx, user.ID); err != nil {
			return nil, 0, err
		}
		return user, 0, nil
	}

	// 重新加密存储已更新的恢复码
	codesJSON, err := marshalRecoveryCodes(codes)
	if err != nil {
		return nil, 0, err
	}
	encryptedCodes, err := s.crypter.Encrypt(codesJSON)
	if err != nil {
		return nil, 0, err
	}
	if err := s.userRepo.UpdateRecoveryCodes(ctx, user.ID, encryptedCodes); err != nil {
		return nil, 0, err
	}

	return user, remaining, nil
}

// GetMFAStatus 获取用户 MFA 状态（剩余恢复码数量）
func (s *MFAService) GetMFAStatus(ctx context.Context, userID uuid.UUID) (bool, int, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, 0, err
	}
	if !user.MFAEnabled || user.RecoveryCodes == nil {
		return user.MFAEnabled, 0, nil
	}

	codes, err := s.decryptRecoveryCodes(user.RecoveryCodes)
	if err != nil {
		return user.MFAEnabled, 0, nil // 解密失败不阻断
	}

	remaining := 0
	for _, rc := range codes {
		if !rc.Used {
			remaining++
		}
	}
	return user.MFAEnabled, remaining, nil
}

// ==============================
// 内部辅助函数
// ==============================

func (s *MFAService) decryptTOTPSecret(encryptedSecret *string) (string, error) {
	if encryptedSecret == nil {
		return "", ErrMFANotSetup
	}
	plainBytes, err := s.crypter.Decrypt(*encryptedSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}
	return string(plainBytes), nil
}

func (s *MFAService) decryptRecoveryCodes(encryptedCodes *string) ([]model.RecoveryCode, error) {
	if encryptedCodes == nil {
		return nil, nil
	}
	plainBytes, err := s.crypter.Decrypt(*encryptedCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt recovery codes: %w", err)
	}
	var codes []model.RecoveryCode
	if err := json.Unmarshal(plainBytes, &codes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recovery codes: %w", err)
	}
	return codes, nil
}

// validateTOTP 验证 TOTP 码（容差 ±1 步）
func validateTOTP(secretBase32, code string) bool {
	valid, err := totp.ValidateCustom(code, secretBase32, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1, // 允许 ±1 步（共 3 个窗口）
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Error().Err(err).Msg("TOTP validation error")
		return false
	}
	return valid
}

// generateRecoveryCodes 生成 n 组恢复码，格式 XXXX-XXXX-XXXX（全大写字母+数字）
func generateRecoveryCodes(n int) ([]model.RecoveryCode, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codes := make([]model.RecoveryCode, n)
	for i := 0; i < n; i++ {
		var parts [3]string
		for j := 0; j < 3; j++ {
			var sb strings.Builder
			for k := 0; k < 4; k++ {
				idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
				if err != nil {
					return nil, err
				}
				sb.WriteByte(charset[idx.Int64()])
			}
			parts[j] = sb.String()
		}
		codes[i] = model.RecoveryCode{
			Code: parts[0] + "-" + parts[1] + "-" + parts[2],
			Used: false,
		}
	}
	return codes, nil
}

// marshalRecoveryCodes 序列化恢复码为 JSON
func marshalRecoveryCodes(codes []model.RecoveryCode) ([]byte, error) {
	b, err := json.Marshal(codes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recovery codes: %w", err)
	}
	return b, nil
}
