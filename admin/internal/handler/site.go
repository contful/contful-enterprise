// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SiteHandler 站点处理器
type SiteHandler struct {
	siteService  *service.SiteService
	auditService *service.AuditService
}

// NewSiteHandler 新建处理器
func NewSiteHandler(siteService *service.SiteService, auditService *service.AuditService) *SiteHandler {
	return &SiteHandler{siteService: siteService, auditService: auditService}
}

// Create 创建站点
func (h *SiteHandler) Create(c *gin.Context) {
	var req model.SiteCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		middleware.Unauthorized(c, "unauthorized")
		return
	}

	resp, err := h.siteService.Create(c.Request.Context(), *userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeSetting, "site:create",
			service.WithResource("site", resp.ID))
	}

	middleware.Created(c, resp.ToResponse())
}

// Get 获取站点
func (h *SiteHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	resp, err := h.siteService.Get(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp.ToResponse())
}

// List 列出站点
func (h *SiteHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	sites, total, err := h.siteService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 转换为响应类型
	items := make([]model.SiteResponse, len(sites))
	for i, s := range sites {
		items[i] = s.ToResponse()
	}

	middleware.OK(c, model.SiteListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// MySites 列出当前用户所属站点（已废弃，直接返回所有站点）
func (h *SiteHandler) MySites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	sites, total, err := h.siteService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 转换为响应类型
	items := make([]model.SiteResponse, len(sites))
	for i, s := range sites {
		items[i] = s.ToResponse()
	}

	middleware.OK(c, model.SiteListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Update 更新站点
func (h *SiteHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	var req model.SiteUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	resp, err := h.siteService.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeSetting, "site:update",
			service.WithResource("site", id))
	}

	middleware.OK(c, resp.ToResponse())
}

// Delete 删除站点
func (h *SiteHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	if err := h.siteService.Delete(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelWarn, model.AuditTypeSetting, "site:delete",
			service.WithResource("site", id))
	}

	c.Status(http.StatusNoContent)
}

// handleError 处理错误
func (h *SiteHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSiteNotFound):
		middleware.NotFound(c, "site not found")
	case errors.Is(err, service.ErrSiteSlugExists):
		middleware.Conflict(c, "site slug already exists")
	case errors.Is(err, service.ErrSiteInvalidSlug):
		middleware.BadRequest(c, "invalid slug: must start with a letter, only lowercase letters, numbers, and hyphens")
	default:
		middleware.InternalError(c, err.Error())
	}
}
