// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/audit"
)

// DefaultSigner 默认签名器：HMAC-SHA256，实现 audit.DataSigner 接口。
// 用户可通过实现 audit.DataSigner 并注入 context（audit.WithSigner）替换为自有签名方法。
type DefaultSigner struct {
	key []byte
}

// NewDefaultSigner 创建默认签名器（key 为 hex 编码的 32 字节密钥）
func NewDefaultSigner(keyHex string) (*DefaultSigner, error) {
	if keyHex == "" {
		return &DefaultSigner{key: nil}, nil
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("签名密钥格式错误（应为 hex）: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("签名密钥长度错误，需要 32 字节，当前 %d 字节", len(key))
	}
	return &DefaultSigner{key: key}, nil
}

// IsEnabled 签名器是否已启用
func (s *DefaultSigner) IsEnabled() bool {
	return s != nil && len(s.key) == 32
}

// Algorithm 实现 audit.DataSigner
func (s *DefaultSigner) Algorithm() string {
	if !s.IsEnabled() {
		return "none"
	}
	return "HMAC-SHA256"
}

// Sign 实现 audit.DataSigner
func (s *DefaultSigner) Sign(entityType string, entityID uuid.UUID, payload string) (string, error) {
	if !s.IsEnabled() {
		return "", nil
	}
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(entityType + ":" + entityID.String() + ":" + payload))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// Verify 实现 audit.DataSigner
func (s *DefaultSigner) Verify(entityType string, entityID uuid.UUID, payload string, signature string) (bool, error) {
	if !s.IsEnabled() {
		return true, nil
	}
	if signature == "" {
		return false, nil
	}
	expected, err := s.Sign(entityType, entityID, payload)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare([]byte(signature), []byte(expected)) == 1, nil
}

// InjectSigner 将 DataSigner 注入 context（便捷封装）
func InjectSigner(ctx context.Context, s audit.DataSigner) context.Context {
	return audit.WithSigner(ctx, s)
}

// ── 规范 Payload 构建 ──

// CanonicalSystemUserPayload 构建 system_users 规范 payload
func CanonicalSystemUserPayload(email, passwordHash, nickname, status string, isSuperAdmin bool) string {
	return fmt.Sprintf("email=%s&password_hash=%s&nickname=%s&status=%s&is_super_admin=%t",
		email, passwordHash, nickname, status, isSuperAdmin)
}

// CanonicalSchemaPayload 构建 schemas 规范 payload
func CanonicalSchemaPayload(name, slug, description, kind string, versioningEnabled bool) string {
	return fmt.Sprintf("name=%s&slug=%s&description=%s&kind=%s&versioning_enabled=%t",
		name, slug, description, kind, versioningEnabled)
}
