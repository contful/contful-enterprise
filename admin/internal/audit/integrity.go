// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/gorm"

	"github.com/contful/contful/admin/internal/crypto"
	"github.com/contful/contful/admin/internal/model"
)

// signingKeyCtxKey context key 类型
type signingKeyCtxKey struct{}

// hasherCtxKey context key 类型（用于注入 HMAC 哈希器，支持 SM3 切换）
type hasherCtxKey struct{}

// WithSigningKey 将签名密钥注入 context（供 callback 使用）
func WithSigningKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, signingKeyCtxKey{}, key)
}

// WithHasher 将 HMAC 哈希器注入 context（供 callback 使用）。
// hasher 为 nil 时回退到 SHA-256（保持向后兼容）。
func WithHasher(ctx context.Context, hasher crypto.Hasher) context.Context {
	return context.WithValue(ctx, hasherCtxKey{}, hasher)
}

// getHasher 从 context 获取 Hasher，返回 nil 时调用方回退 SHA-256
func getHasher(ctx context.Context) crypto.Hasher {
	if ctx == nil {
		return nil
	}
	if h, ok := ctx.Value(hasherCtxKey{}).(crypto.Hasher); ok {
		return h
	}
	return nil
}

// Register 注册 GORM callbacks：audit_logs / system_users / schemas 的 BeforeCreate/BeforeUpdate 签名
func Register(db *gorm.DB) {
	if db.Dialector.Name() != "postgres" {
		return
	}
	// audit_logs
	db.Callback().Create().Before("gorm:create").Register("audit:sign_create", signBeforeCreate)
	// system_users + schemas
	db.Callback().Create().Before("gorm:create").Register("business:sign_create", signBusinessBeforeCreate)
	db.Callback().Update().Before("gorm:update").Register("business:sign_update", signBusinessBeforeUpdate)
}

// signBeforeCreate 处理 audit_logs 的 BeforeCreate 签名
func signBeforeCreate(scope *gorm.DB) {
	if scope.Statement.Table != "audit_logs" {
		return
	}

	auditLogValue := scope.Statement.ReflectValue.Interface()
	auditLog, ok := auditLogValue.(model.AuditLog)
	if !ok {
		return
	}

	signingKey := ""
	if ctx := scope.Statement.Context; ctx != nil {
		if key, ok := ctx.Value(signingKeyCtxKey{}).(string); ok {
			signingKey = key
		}
	}

	if signingKey == "" {
		scope.Statement.SetColumn("data_signature", "")
		return
	}

	payload := canonicalAuditPayload(&auditLog)

	// 优先使用 context 注入的 Hasher（国密模式时为 SM3），回退 SHA-256
	hasher := getHasher(scope.Statement.Context)
	var sigHex string
	if hasher != nil {
		sigHex = hex.EncodeToString(hasher.HMAC([]byte(signingKey), []byte(payload)))
	} else {
		h := hmac.New(sha256.New, []byte(signingKey))
		h.Write([]byte(payload))
		sigHex = hex.EncodeToString(h.Sum(nil))
	}

	scope.Statement.SetColumn("data_signature", sigHex)
}

// signBusinessBeforeCreate 处理 system_users / schemas 的 BeforeCreate 签名
func signBusinessBeforeCreate(scope *gorm.DB) {
	signBusiness(scope)
}

// signBusinessBeforeUpdate 处理 system_users / schemas 的 BeforeUpdate 签名
func signBusinessBeforeUpdate(scope *gorm.DB) {
	signBusiness(scope)
}

// signBusiness 对 system_users / schemas 执行签名
func signBusiness(scope *gorm.DB) {
	table := scope.Statement.Table
	if table != "system_users" && table != "schemas" {
		return
	}

	signer := GetSigner(scope.Statement.Context)
	if signer == nil {
		return
	}

	var entityType string
	var id uid.UID
	var payload string

	switch table {
	case "system_users":
		if u, ok := scope.Statement.ReflectValue.Interface().(model.SystemUser); ok {
			entityType = "system_users"
			id = u.ID
			payload = fmt.Sprintf("email=%s&password_hash=%s&nickname=%s&status=%s&is_super_admin=%t",
				u.Email, u.PasswordHash, u.Nickname, u.Status, u.IsSuperAdmin)
		} else {
			return
		}

	case "schemas":
		if s, ok := scope.Statement.ReflectValue.Interface().(model.ContentSchema); ok {
			entityType = "schemas"
			id = s.ID
			payload = fmt.Sprintf("name=%s&slug=%s&description=%s&kind=%s&versioning_enabled=%t",
				s.Name, s.Slug, s.Description, s.Kind, s.VersioningEnabled)
		} else {
			return
		}
	}

	sig, err := signer.Sign(entityType, id, payload)
	if err != nil {
		return
	}
	scope.Statement.SetColumn("data_signature", sig)
}

// canonicalAuditPayload 构建 AuditLog 规范 payload
func canonicalAuditPayload(a *model.AuditLog) string {
	parts := []string{
		"action=" + a.Action,
	}
	if a.SiteID != nil {
		parts = append(parts, fmt.Sprintf("site_id=%s", a.SiteID.String()))
	}
	if a.UserID != nil {
		parts = append(parts, fmt.Sprintf("user_id=%s", a.UserID.String()))
	}
	if a.ResourceID != nil {
		parts = append(parts, fmt.Sprintf("resource_id=%s", a.ResourceID.String()))
	}
	parts = append(parts, "level="+string(a.Level))
	parts = append(parts, "category="+string(a.Category))
	parts = append(parts, "created_time="+a.CreatedTime.Format(time.RFC3339))

	return strings.Join(parts, "&")
}
