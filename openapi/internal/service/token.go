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
	// 1. 格式校验
	rawToken = strings.TrimSpace(rawToken)
	if !strings.HasPrefix(rawToken, tokenPrefix) {
		return nil, ErrInvalidTokenFormat
	}
	if len(rawToken) < len(tokenPrefix)+minLen {
		return nil, ErrInvalidTokenFormat
	}

	// 2. 计算 Hash（数据库只存 Hash，不存明文）
	hash := sha256Hash(rawToken)

	// 3. 查询数据库
	token, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, ErrTokenNotFound
	}

	// 4. 检查状态
	if token.Status == "revoked" {
		return nil, ErrTokenRevoked
	}

	// 5. 检查过期时间
	if token.ExpiresTime != nil && token.ExpiresTime.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// 6. 异步更新最后使用时间（不阻塞请求）
	go func() {
		_ = s.repo.UpdateLastUsedTime(context.Background(), token.ID)
	}()

	// 7. 构建上下文
	// Scopes 字段：["read"] 或 ["read","write"]，解析为读写权限
	allowRead := false
	allowWrite := false
	for _, scope := range token.Scopes {
		switch scope {
		case "read":
			allowRead = true
		case "write":
			allowWrite = true
		}
	}
	// SiteScope：["*"] 表示全站，否则为指定 content schema slug 列表
	var contentSchemas []string
	if len(token.SiteScope) == 1 && token.SiteScope[0] == "*" {
		contentSchemas = []string{"*"}
	} else {
		contentSchemas = token.SiteScope
	}
	perm := model.TokenPermission{
		AllowRead:     allowRead,
		AllowWrite:    allowWrite,
		ContentSchemas: contentSchemas,
	}

	rateCfg := model.RateLimitConfig{
		RequestsPerMinute: token.RateLimit,
	}

	var expiresAt *int64
	if token.ExpiresTime != nil {
		t := token.ExpiresTime.Unix()
		expiresAt = &t
	}

	return &model.TokenContext{
		TokenID:     token.ID,
		SiteID:      token.SiteID,
		Name:        token.Name,
		Permissions: perm,
		RateLimits:  rateCfg,
		ExpiresAt:   expiresAt,
	}, nil
}

// CheckScope 检查 Token 是否有权限访问指定内容和操作
func (s *APITokenService) CheckScope(tc *model.TokenContext, contentSlug string, method string) bool {
	// 检查内容类型权限
	types := tc.Permissions.ContentSchemas
	isWildcard := len(types) == 1 && types[0] == "*"

	if !isWildcard {
		allowed := false
		for _, t := range types {
			if t == contentSlug {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// 检查操作权限
	switch method {
	case "GET", "HEAD", "OPTIONS":
		return tc.Permissions.AllowRead
	case "POST", "PUT", "PATCH", "DELETE":
		return tc.Permissions.AllowWrite
	default:
		return false
	}
}

// sha256Hash 计算 SHA-256 哈希
func sha256Hash(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
