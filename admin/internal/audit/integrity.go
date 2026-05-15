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

	"gorm.io/gorm"

	"github.com/contful/contful/admin/internal/model"
)

// signingKeyCtxKey context key 类型
type signingKeyCtxKey struct{}

// WithSigningKey 将签名密钥注入 context（供 callback 使用）
func WithSigningKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, signingKeyCtxKey{}, key)
}

// CallbackName GORM callback 名称
const CallbackName = "audit:integrity_sign"

// Register 注册 BeforeCreate callback，自动为 AuditLog 生成数据签名
func Register(db *gorm.DB) {
	if db.Dialector.Name() == "postgres" {
		db.Callback().Create().Before("gorm:create").Register(CallbackName, signAuditLog)
	}
}

// signAuditLog 在 AuditLog 插入前计算并写入 data_signature（仅存 hex 签名）
func signAuditLog(scope *gorm.DB) {
	// 仅处理 audit_logs 表
	if scope.Statement.Table != "audit_logs" {
		return
	}

	auditLogValue := scope.Statement.ReflectValue.Interface()
	auditLog, ok := auditLogValue.(model.AuditLog)
	if !ok {
		return
	}

	// 获取签名密钥（从 context 注入）
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

	// 构建规范 payload 并计算 HMAC-SHA256 签名
	payload := canonicalAuditPayload(&auditLog)
	h := hmac.New(sha256.New, []byte(signingKey))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	scope.Statement.SetColumn("data_signature", signature)
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
