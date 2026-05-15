// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
)

// AuditService 审计日志服务
type AuditService struct {
	auditRepo *repository.AuditRepository
	configSvc *ConfigService
}

func NewAuditService(auditRepo *repository.AuditRepository, configSvc *ConfigService) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
		configSvc: configSvc,
	}
}

// LogOption 审计日志记录选项
type LogOption func(*model.AuditLog)

// WithAuditSiteID 设置站点 ID
func WithAuditSiteID(siteID uuid.UUID) LogOption {
	return func(a *model.AuditLog) {
		a.SiteID = &siteID
	}
}

// WithResource 设置资源类型和资源 ID
func WithResource(resourceType string, resourceID uuid.UUID) LogOption {
	return func(a *model.AuditLog) {
		a.ResourceType = resourceType
		a.ResourceID = &resourceID
	}
}

// WithDetails 设置详细信息
func WithDetails(details string) LogOption {
	return func(a *model.AuditLog) {
		a.Details = details
	}
}

// WithIPAddress 设置 IP 地址
func WithIPAddress(ip string) LogOption {
	return func(a *model.AuditLog) {
		a.IPAddress = ip
	}
}

// WithUserAgent 设置 User-Agent
func WithUserAgent(ua string) LogOption {
	return func(a *model.AuditLog) {
		a.UserAgent = ua
	}
}

// Log 记录审计日志（高层接口）
func (s *AuditService) Log(ctx context.Context, userID uuid.UUID, level model.AuditLevel, category model.AuditType, action string, opts ...LogOption) error {
	auditLog := &model.AuditLog{
		UserID:   &userID,
		Action:   action,
		Level:    level,
		Category: category,
	}

	// 应用选项
	for _, opt := range opts {
		opt(auditLog)
	}

	// 获取签名密钥并记录
	signingKey, err := s.configSvc.GetAuditSigningKey()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get audit signing key, logging without signature")
		return s.auditRepo.Create(ctx, auditLog)
	}

	return s.auditRepo.CreateWithSigningKey(ctx, auditLog, signingKey)
}

// LogFromGin 从 Gin 上下文记录审计日志（自动提取用户信息）
func (s *AuditService) LogFromGin(c *gin.Context, level model.AuditLevel, category model.AuditType, action string, opts ...LogOption) error {
	// 从 Gin 上下文获取用户 ID（中间件存入 key="user"，类型 uuid.UUID）
	userIDVal, exists := c.Get("user")
	if !exists {
		log.Warn().Msg("user_id not found in gin context, skipping audit log")
		return nil
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		log.Warn().Msg("invalid user_id type in gin context, skipping audit log")
		return nil
	}

	// 自动提取 IP 和 User-Agent
	opts = append(opts, WithIPAddress(c.ClientIP()))
	opts = append(opts, WithUserAgent(c.GetHeader("User-Agent")))

	return s.Log(c.Request.Context(), userID, level, category, action, opts...)
}

// LogAuth 记录认证相关审计日志
func (s *AuditService) LogAuth(ctx context.Context, userID uuid.UUID, action string, ipAddress string, userAgent string, success bool) error {
	level := model.AuditLevelInfo
	details := "success"
	if !success {
		level = model.AuditLevelWarn
		details = "failed"
	}

	return s.Log(ctx, userID, level, model.AuditTypeAuth, action,
		WithIPAddress(ipAddress),
		WithUserAgent(userAgent),
		WithDetails(details),
	)
}

// LogUser 记录用户管理相关审计日志
func (s *AuditService) LogUser(ctx context.Context, operatorID uuid.UUID, action string, targetUserID uuid.UUID, opts ...LogOption) error {
	opts = append(opts, WithResource("user", targetUserID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeUser, action, opts...)
}

// LogRole 记录角色管理相关审计日志
func (s *AuditService) LogRole(ctx context.Context, operatorID uuid.UUID, action string, targetRoleID uuid.UUID, opts ...LogOption) error {
	opts = append(opts, WithResource("role", targetRoleID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeUser, action, opts...)
}

// LogSite 记录站点管理相关审计日志
func (s *AuditService) LogSite(ctx context.Context, operatorID uuid.UUID, action string, targetSiteID uuid.UUID, opts ...LogOption) error {
	opts = append(opts, WithResource("site", targetSiteID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeSetting, action, opts...)
}

// LogToken 记录 Token 管理相关审计日志
func (s *AuditService) LogToken(ctx context.Context, operatorID uuid.UUID, action string, targetTokenID uuid.UUID, opts ...LogOption) error {
	opts = append(opts, WithResource("token", targetTokenID))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeSystem, action, opts...)
}

// LogContent 记录内容管理相关审计日志
func (s *AuditService) LogContent(ctx context.Context, operatorID uuid.UUID, action string, siteID uuid.UUID, schemaID uuid.UUID, entryID uuid.UUID, opts ...LogOption) error {
	opts = append(opts, WithAuditSiteID(siteID))
	opts = append(opts, WithResource("entry", entryID))
	opts = append(opts, WithDetails("schema_id="+schemaID.String()))
	return s.Log(ctx, operatorID, model.AuditLevelInfo, model.AuditTypeContent, action, opts...)
}

// LogError 记录错误审计日志
func (s *AuditService) LogError(ctx context.Context, userID uuid.UUID, category model.AuditType, action string, err error, opts ...LogOption) error {
	opts = append(opts, WithDetails("error="+err.Error()))
	return s.Log(ctx, userID, model.AuditLevelError, category, action, opts...)
}

// List 查询审计日志列表（支持筛选和分页）
func (s *AuditService) List(ctx context.Context, filter *model.AuditLogFilter, page, pageSize int) ([]model.AuditLog, int64, error) {
	return s.auditRepo.List(ctx, filter, page, pageSize)
}

// GetByID 根据 ID 获取审计日志详情
func (s *AuditService) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	return s.auditRepo.GetByID(ctx, id)
}

// GetSigningKey 获取签名密钥（供其他服务使用）
func (s *AuditService) GetSigningKey(ctx context.Context) (string, error) {
	return s.configSvc.GetAuditSigningKey()
}

// Helper: 从 HTTP 请求中提取 IP 和 User-Agent
func getClientInfo(r *http.Request) (string, string) {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	ua := r.Header.Get("User-Agent")
	return ip, ua
}
