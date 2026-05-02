// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/crypto"
	"github.com/contful/contful/admin/internal/repository"

	"github.com/google/uuid"
)

const (
	// TokenPrefix API Token 前缀
	TokenPrefix = "ctf_"
	// TokenLength Token 随机部分长度
	TokenLength = 32
)

// APITokenService API Token 服务
type APITokenService struct {
	tokenRepo *repository.APITokenRepository
	crypter   crypto.Crypter
}

// NewAPITokenService 新建服务
// crypter: 加密器，由 NewCrypter(algorithm, secret) 创建
func NewAPITokenService(tokenRepo *repository.APITokenRepository, crypter crypto.Crypter) *APITokenService {
	return &APITokenService{tokenRepo: tokenRepo, crypter: crypter}
}

// GenerateToken 生成新的 Token，返回完整 Token、Hash、前缀
func (s *APITokenService) GenerateToken() (fullToken string, tokenHash string, prefix string, err error) {
	randomBytes := make([]byte, TokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", "", fmt.Errorf("生成随机数失败: %w", err)
	}
	fullToken = TokenPrefix + hex.EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(fullToken))
	tokenHash = hex.EncodeToString(hash[:])
	prefix = fullToken[:10]
	return fullToken, tokenHash, prefix, nil
}

// Create 创建 API Token
func (s *APITokenService) Create(ctx context.Context, siteID, userID uuid.UUID, req *model.APITokenCreate) (*model.APIToken, string, error) {
	fullToken, tokenHash, prefix, err := s.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	// 加密存储 Token
	encryptedToken, err := s.crypter.Encrypt([]byte(fullToken))
	if err != nil {
		return nil, "", fmt.Errorf("加密 Token 失败: %w", err)
	}

	token := &model.APIToken{
		ID:             uuid.New(),
		SiteID:         siteID,
		Name:           req.Name,
		Description:    req.Description,
		TokenPrefix:    prefix,
		TokenHash:      tokenHash,
		EncryptedToken: encryptedToken,
		RateLimit:      60, // 默认每分钟 60 次
		Scopes:         model.StringArray{"read"}, // 默认只读
		SiteScope:      model.StringArray{"*"},    // 全部站点
		ChannelScope:   model.StringArray{},       // 空数组，非 nil
		Status:         model.TokenStatusActive,
		CreatedBy:      &userID,
	}

	token.ExpiresTime = nil
	if req.ExpiresTime != nil {
		token.ExpiresTime = req.ExpiresTime
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, "", fmt.Errorf("创建 Token 失败: %w", err)
	}

	return token, fullToken, nil
}

// Get 获取 Token
func (s *APITokenService) Get(ctx context.Context, id uuid.UUID) (*model.APIToken, error) {
	return s.tokenRepo.GetByID(ctx, id)
}

// List 列出 Token
func (s *APITokenService) List(ctx context.Context, siteID uuid.UUID, filter *model.APITokenListFilter, page, pageSize int) (*model.APITokenListResponse, error) {
	tokens, total, err := s.tokenRepo.List(ctx, siteID, filter, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]model.APITokenResponse, len(tokens))
	for i, token := range tokens {
		items[i] = token.ToResponse()
	}

	return &model.APITokenListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update 更新 Token
func (s *APITokenService) Update(ctx context.Context, id uuid.UUID, req *model.APITokenUpdate) (*model.APIToken, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		token.Name = *req.Name
	}
	if req.Description != nil {
		token.Description = *req.Description
	}
	if req.ExpiresTime != nil {
		token.ExpiresTime = req.ExpiresTime
	}
	if req.Status != nil {
		token.Status = model.TokenStatus(*req.Status)
	}

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// Delete 删除 Token
func (s *APITokenService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.tokenRepo.Delete(ctx, id)
}

// Revoke 撤销 Token
func (s *APITokenService) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.tokenRepo.Revoke(ctx, id)
}

// Validate 验证 Token
func (s *APITokenService) Validate(ctx context.Context, tokenStr string) (*model.APIToken, error) {
	if !strings.HasPrefix(tokenStr, TokenPrefix) {
		return nil, errors.New("invalid token format")
	}
	hash := sha256.Sum256([]byte(tokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	token, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	if token.Status != model.TokenStatusActive {
		return nil, fmt.Errorf("token is %s", token.Status)
	}
	if token.ExpiresTime != nil && !token.ExpiresTime.IsZero() && token.ExpiresTime.Before(time.Now()) {
		token.Status = model.TokenStatusExpired
		s.tokenRepo.Update(ctx, token)
		return nil, errors.New("token expired")
	}
	// 更新最后使用时间和 IP（由调用方注入 IP）
	s.tokenRepo.UpdateLastUsed(ctx, token.ID)

	return token, nil
}

// Regenerate 重新生成 Token（保留原 Token 的 ID 和权限）
func (s *APITokenService) Regenerate(ctx context.Context, id uuid.UUID) (*model.APIToken, string, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	fullToken, tokenHash, prefix, err := s.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	// 加密存储新 Token
	encryptedToken, err := s.crypter.Encrypt([]byte(fullToken))
	if err != nil {
		return nil, "", fmt.Errorf("加密 Token 失败: %w", err)
	}

	token.TokenHash = tokenHash
	token.TokenPrefix = prefix
	token.EncryptedToken = encryptedToken
	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, "", err
	}
	return token, fullToken, nil
}

// Export 导出 Token（解密并返回完整 Token）
func (s *APITokenService) Export(ctx context.Context, id uuid.UUID) (*model.APIToken, string, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	if token.EncryptedToken == "" {
		return nil, "", errors.New("token not found or not exportable")
	}

	// 解密获取完整 Token
	fullToken, err := s.crypter.Decrypt(token.EncryptedToken)
	if err != nil {
		return nil, "", fmt.Errorf("解密 Token 失败: %w", err)
	}

	return token, string(fullToken), nil
}
