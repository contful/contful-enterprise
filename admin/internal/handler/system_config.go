// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
)

type SystemConfigHandler struct {
	configRepo   *repository.SystemConfigRepository
	rbacService  *service.RBACService
	auditService *service.AuditService
}

func NewSystemConfigHandler(configRepo *repository.SystemConfigRepository, rbacService *service.RBACService, auditService *service.AuditService) *SystemConfigHandler {
	return &SystemConfigHandler{
		configRepo:   configRepo,
		rbacService:  rbacService,
		auditService: auditService,
	}
}

// GetPublicConfig 获取公开配置（无需认证）
func (h *SystemConfigHandler) GetPublicConfig(c *gin.Context) {
	configs, err := h.configRepo.ListByPublic(c.Request.Context(), true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to get public config"))
		return
	}

	// 转换为 map 方便前端使用
	result := make(map[string]interface{})
	for _, cfg := range configs {
		result[cfg.ConfigKey] = cfg.ConfigValue
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// List 获取所有配置（需要认证 + settings:read 权限）
func (h *SystemConfigHandler) List(c *gin.Context) {
	configs, err := h.configRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to list config"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(configs))
}

// Get 获取单个配置（需要认证 + settings:read 权限）
func (h *SystemConfigHandler) Get(c *gin.Context) {
	key := c.Param("key")

	config, err := h.configRepo.GetByKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "config not found"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(config))
}

// Update 更新配置（需要认证 + settings:write 权限，支持部分更新）
func (h *SystemConfigHandler) Update(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		ConfigValue *string `json:"config_value"`
		ValueType   *string `json:"value_type" binding:"omitempty,oneof=string number boolean json"`
		Description *string `json:"description"`
		IsPublic    *bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	// 读取当前配置，保留未传入的字段
	current, err := h.configRepo.GetByKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "config not found"))
		return
	}

	configValue := current.ConfigValue
	valueType := current.ValueType
	description := current.Description
	isPublic := current.IsPublic

	if req.ConfigValue != nil {
		configValue = *req.ConfigValue
	}
	if req.ValueType != nil {
		valueType = *req.ValueType
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	// 验证值类型
	if req.ConfigValue != nil && !h.validateValue(configValue, valueType) {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid value for type "+valueType))
		return
	}

	err = h.configRepo.Set(c.Request.Context(), key, configValue, valueType, description, isPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to update config"))
		return
	}

	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeSetting, "config:update",
			service.WithResource("config", current.ID))
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "config updated"}))
}

// GetSiteConfig 获取站点公开配置（登录页需要，无需认证）
// 返回：站点品牌、安全策略、密码强度规则
func (h *SystemConfigHandler) GetSiteConfig(c *gin.Context) {
	ctx := c.Request.Context()
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
		// 站点品牌
		"site_name":           h.configRepo.GetString(ctx, "site_name", "Contful"),
		"site_description":    h.configRepo.GetString(ctx, "site_description", ""),
		"logo_url":            h.configRepo.GetString(ctx, "logo_url", ""),
		"login_background_url": h.configRepo.GetString(ctx, "login_background_url", ""),
		// 安全策略
		"mfa_enforced":        h.configRepo.GetBool(ctx, "mfa_enforced", false),
		"login_max_attempts":  h.configRepo.GetInt(ctx, "login_max_attempts", 5),
		"login_lock_duration": h.configRepo.GetInt(ctx, "login_lock_duration", 30),
		// 密码策略
		"password_min_length":        h.configRepo.GetInt(ctx, "password_min_length", 8),
		"password_require_uppercase": h.configRepo.GetBool(ctx, "password_require_uppercase", true),
		"password_require_lowercase": h.configRepo.GetBool(ctx, "password_require_lowercase", true),
		"password_require_number":    h.configRepo.GetBool(ctx, "password_require_number", true),
		"password_require_special":   h.configRepo.GetBool(ctx, "password_require_special", false),
		"password_expire_days":       h.configRepo.GetInt(ctx, "password_expire_days", 90),
	}))
}

// Create 创建自定义配置（需要认证 + settings:write 权限）
func (h *SystemConfigHandler) Create(c *gin.Context) {
	var req struct {
		ConfigKey   string `json:"config_key" binding:"required"`
		ConfigValue string `json:"config_value"`
		ValueType   string `json:"value_type" binding:"required,oneof=string number boolean json"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	if !h.validateValue(req.ConfigValue, req.ValueType) {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid value for type "+req.ValueType))
		return
	}

	config, err := h.configRepo.Create(c.Request.Context(), req.ConfigKey, req.ConfigValue, req.ValueType, req.Description, req.IsPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to create config"))
		return
	}

	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeSetting, "config:create",
			service.WithResource("config", config.ID))
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(config))
}

// Delete 删除自定义配置（需要认证 + settings:write 权限，系统配置不可删除）
func (h *SystemConfigHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	err := h.configRepo.Delete(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "cannot delete system config or config not found"))
		return
	}

	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelWarn, model.AuditTypeSetting, "config:delete",
			service.WithDetails("key="+key))
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "config deleted"}))
}

// ClearCache 清除所有系统配置的 Redis 缓存
func (h *SystemConfigHandler) ClearCache(c *gin.Context) {
	if err := h.configRepo.ClearCache(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to clear cache"))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "cache cleared"}))
}

// validateValue 验证配置值是否符合类型
func (h *SystemConfigHandler) validateValue(value string, valueType string) bool {
	switch valueType {
	case "number":
		_, err := strconv.Atoi(value)
		return err == nil
	case "boolean":
		return value == "true" || value == "false"
	case "string", "json":
		return true
	default:
		return false
	}
}
