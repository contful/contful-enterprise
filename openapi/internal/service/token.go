// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/contful/contful/openapi/internal/model"
	"github.com/contful/contful/openapi/internal/repository"
)

const (
	tokenPrefix = "ctf_"
	minLen     = 20 // Token 最小长度（不含前缀）
)

// ErrTokenNotFound Token 不存在
var ErrTokenNotFound = errors.New("token not found")

// ErrTokenExpired Token 已过期
var ErrTokenExpired = errors.New("token expired")

// ErrTokenRevoked Token 已撤销
var ErrTokenRevoked = errors.New("token revoked")

// ErrInvalidTokenFormat Token 格式无效
var ErrInvalidTokenFormat = errors.New("invalid token format")

// APITokenService Token 业务逻辑层
type APITokenService struct {
	repo *repository.APITokenRepository
}

// NewAPITokenService 创建 Token Service
func NewAPITokenService(repo *repository.APITokenRepository) *APITokenService {
	return &APITokenService{repo: repo}
}

// ValidateToken 验证 Token 并返回上下文信息
func (s *APITokenService) ValidateToken(ctx context.Context, rawToken string) (*model.TokenContext, error) {
	rawToken = strings.TrimSpace(rawToken)
	if !strings.HasPrefix(rawToken, tokenPrefix) {
		return nil, ErrInvalidTokenFormat
	}
	if len(rawToken) < len(tokenPrefix)+minLen {
		return nil, ErrInvalidTokenFormat
	}

	hash := sha256Hash(rawToken)

	token, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, ErrTokenNotFound
	}

	if token.Status == "revoked" {
		return nil, ErrTokenRevoked
	}

	if token.ExpiresTime != nil && token.ExpiresTime.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// 异步更新最后使用时间
	go func() {
		_ = s.repo.UpdateLastUsedTime(context.Background(), token.ID)
	}()

	var expiresAt *int64
	if token.ExpiresTime != nil {
		t := token.ExpiresTime.Unix()
		expiresAt = &t
	}

	return &model.TokenContext{
		TokenID:   token.ID,
		SiteID:    token.SiteID,
		Name:      token.Name,
		ExpiresAt: expiresAt,
	}, nil
}

// sha256Hash 计算 SHA-256 哈希
func sha256Hash(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
