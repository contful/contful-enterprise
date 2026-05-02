// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConfigHandler 站点配置处理器
type ConfigHandler struct {
	configService *service.ConfigService
}

// NewConfigHandler 新建处理器
func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

// List 列出站点的所有配置
func (h *ConfigHandler) List(c *gin.Context) {
	siteID, err := h.getSiteID(c)
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	group := c.Query("group")

	var configs []model.SiteConfig
	if group == "" {
		configs, err = h.configService.ListAll(c.Request.Context(), siteID)
	} else {
		configs, err = h.configService.ListByGroup(c.Request.Context(), siteID, group)
	}
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.OK(c, configs)
}

// Get 获取单个配置值
func (h *ConfigHandler) Get(c *gin.Context) {
	siteID, err := h.getSiteID(c)
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	key := c.Param("key")
	value, err := h.configService.Get(c.Request.Context(), siteID, key)
	if err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			middleware.NotFound(c, "配置不存在")
			return
		}
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.OK(c, gin.H{"key": key, "value": value})
}

// Set 创建或更新配置
func (h *ConfigHandler) Set(c *gin.Context) {
	siteID, err := h.getSiteID(c)
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	key := c.Param("key")

	var req model.CreateSiteConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)

	if err := h.configService.Set(c.Request.Context(), siteID, key, req.ConfigValue, &req, userID); err != nil {
		if errors.Is(err, service.ErrConfigReadonly) {
			middleware.Forbidden(c, "配置为只读，禁止修改")
			return
		}
		if errors.Is(err, service.ErrCrypterEmpty) {
			middleware.InternalError(c, "服务器未配置 SECRET，无法加密存储敏感配置")
			return
		}
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.OK(c, gin.H{"key": key, "msg": "配置已保存"})
}

// Delete 删除配置
func (h *ConfigHandler) Delete(c *gin.Context) {
	siteID, err := h.getSiteID(c)
	if err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	key := c.Param("key")

	if err := h.configService.Delete(c.Request.Context(), siteID, key); err != nil {
		if errors.Is(err, service.ErrConfigNotFound) {
			middleware.NotFound(c, "配置不存在")
			return
		}
		if errors.Is(err, service.ErrConfigReadonly) {
			middleware.Forbidden(c, "配置为只读，禁止删除")
			return
		}
		middleware.InternalError(c, err.Error())
		return
	}

	middleware.OK(c, gin.H{"key": key, "msg": "配置已删除"})
}

// getSiteID 从 URL 参数或请求头获取 siteID
func (h *ConfigHandler) getSiteID(c *gin.Context) (uuid.UUID, error) {
	// 优先从 URL 参数
	if idStr := c.Param("id"); idStr != "" {
		return uuid.Parse(idStr)
	}
	// 其次从请求头 X-Site-ID
	if idStr := c.GetHeader("X-Site-ID"); idStr != "" {
		return uuid.Parse(idStr)
	}
	return uuid.Nil, errors.New("site_id 为必填参数")
}
