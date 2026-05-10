// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package audit

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// signAuditLog 在 AuditLog 插入前计算并写入 data_signature
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
		// 无密钥时写入空签名（降级，不阻塞操作）
		scope.Statement.SetColumn("data_signature", map[string]interface{}{})
		return
	}

	// 构建规范 payload
	payload := canonicalAuditPayload(&auditLog)

	// HMAC-SHA256 签名
	h := hmac.New(sha256.New, []byte(signingKey))
	h.Write([]byte(payload))
	signature := hex.EncodeToString(h.Sum(nil))

	sigData := &SignatureData{
		Alg:      "HMAC-SHA256",
		Sign:     signature,
		Entity:   "audit_log",
		EntityID: auditLog.ID.String(),
		IssuedAt: auditLog.CreatedTime.Format(time.RFC3339),
	}

	sigJSON, err := json.Marshal(sigData)
	if err != nil {
		scope.Statement.SetColumn("data_signature", map[string]interface{}{})
		return
	}

	// 反序列化回 map 以符合 GORM 写入格式
	var sigMap map[string]interface{}
	if err := json.Unmarshal(sigJSON, &sigMap); err != nil {
		scope.Statement.SetColumn("data_signature", map[string]interface{}{})
		return
	}
	scope.Statement.SetColumn("data_signature", sigMap)
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

// SignatureData 签名数据结构
type SignatureData struct {
	Alg      string `json:"alg"`
	Sign     string `json:"sign"`
	Entity   string `json:"entity"`
	EntityID string `json:"entity_id"`
	IssuedAt string `json:"issued_at"`
}
