// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EntryHandler 条目处理器
type EntryHandler struct {
	entryService   *service.EntryService
	configService *service.ConfigService
}

// NewEntryHandler 新建处理器
func NewEntryHandler(entryService *service.EntryService, configService *service.ConfigService) *EntryHandler {
	return &EntryHandler{entryService: entryService, configService: configService}
}

// getIntegrityService 获取站点的签名服务
func (h *EntryHandler) getIntegrityService(ctx context.Context, siteID uuid.UUID) *service.IntegrityService {
	if h.configService == nil {
		return nil
	}
	signingKey, _ := h.configService.GetAuditSigningKey()
	alg := "HMAC-SHA256" // 默认算法
	svc, _ := service.NewIntegrityService(siteID, signingKey, alg)
	return svc
}

// RegisterRoutes 注册路由
func (h *EntryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	entries := rg.Group("/content/entries")
	{
		entries.GET("", h.List)
		entries.POST("", h.Create)
		entries.GET("/:id", h.Get)
		entries.PUT("/:id", h.Update)
		entries.DELETE("/:id", h.Delete)
		entries.POST("/:id/publish", h.Publish)
		entries.POST("/:id/unpublish", h.Unpublish)
		entries.GET("/:id/versions", h.GetVersions)
		// 批量操作
		entries.POST("/batch-delete", h.BatchDelete)
		entries.POST("/batch-publish", h.BatchPublish)
		entries.POST("/batch-unpublish", h.BatchUnpublish)
		// 定时排期
		entries.GET("/scheduled", h.ScheduledList)
		entries.PUT("/:id/schedule", h.Schedule)
		entries.DELETE("/:id/schedule", h.Unschedule)
	}
}

// Create 创建条目
func (h *EntryHandler) Create(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	if siteID == uuid.Nil {
		middleware.BadRequest(c, "X-Site-ID header is required")
		return
	}
	userID, _ := middleware.GetUserID(c)

	var req model.EntryCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	entry, err := h.entryService.Create(c.Request.Context(), siteID, userID, &req, h.getIntegrityService(c.Request.Context(), siteID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.Created(c, entry.ToResponse())
}

// Get 获取条目
func (h *EntryHandler) Get(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	entry, err := h.entryService.GetByID(c.Request.Context(), siteID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// List 列出条目
func (h *EntryHandler) List(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析过滤参数
	filter := &model.EntryListFilter{}
	if ctID := c.Query("schema_id"); ctID != "" {
		if id, err := uuid.Parse(ctID); err == nil {
			filter.ContentSchemaID = &id
		}
	}
	if status := c.Query("status"); status != "" {
		s := model.EntryStatus(status)
		filter.Status = &s
	}
	if locale := c.Query("locale"); locale != "" {
		filter.Locale = &locale
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filter.Keyword = &keyword
	}
	// 排序参数
	filter.SortField = c.DefaultQuery("sort_field", "updated_time")
	filter.SortOrder = c.DefaultQuery("sort_order", "desc")

	entries, total, err := h.entryService.List(c.Request.Context(), siteID, filter, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 转换为响应（包含 ContentSchema 信息）
	items := make([]model.EntryResponseWithType, len(entries))
	for i, e := range entries {
		items[i] = e.ToResponseWithType()
	}

	middleware.OK(c, model.EntryListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Update 更新条目
func (h *EntryHandler) Update(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	userID, _ := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	var req model.EntryUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	entry, err := h.entryService.Update(c.Request.Context(), siteID, userID, id, &req, h.getIntegrityService(c.Request.Context(), siteID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// Delete 删除条目
func (h *EntryHandler) Delete(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	if err := h.entryService.Delete(c.Request.Context(), siteID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Publish 发布条目
func (h *EntryHandler) Publish(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	userID, _ := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	var req model.EntryPublish
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body
		req = model.EntryPublish{}
	}

	entry, err := h.entryService.Publish(c.Request.Context(), siteID, userID, id, &req, h.getIntegrityService(c.Request.Context(), siteID))
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// Unpublish 取消发布
func (h *EntryHandler) Unpublish(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	entry, err := h.entryService.Unpublish(c.Request.Context(), siteID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// Schedule 设置定时排期
func (h *EntryHandler) Schedule(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	var req model.EntryScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	entry, err := h.entryService.SetSchedule(c.Request.Context(), siteID, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// Unschedule 清除定时排期
func (h *EntryHandler) Unschedule(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	entry, err := h.entryService.ClearSchedule(c.Request.Context(), siteID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// ScheduledList 列出有排期的条目
func (h *EntryHandler) ScheduledList(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	entries, total, err := h.entryService.ListScheduled(c.Request.Context(), siteID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]model.EntryResponseWithType, len(entries))
	for i, e := range entries {
		items[i] = e.ToResponseWithType()
	}

	middleware.OK(c, model.EntryListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetVersions 获取版本历史
func (h *EntryHandler) GetVersions(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid entry id")
		return
	}

	versions, err := h.entryService.GetVersions(c.Request.Context(), siteID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, versions)
}

// ============ 批量操作 ============

// BatchDelete 批量删除
func (h *EntryHandler) BatchDelete(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	var req model.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	result, err := h.entryService.BatchDelete(c.Request.Context(), siteID, req.IDs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, result)
}

// BatchPublish 批量发布
func (h *EntryHandler) BatchPublish(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	var req model.BatchPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	result, err := h.entryService.BatchPublish(c.Request.Context(), siteID, req.IDs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, result)
}

// BatchUnpublish 批量取消发布
func (h *EntryHandler) BatchUnpublish(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	var req model.BatchPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	result, err := h.entryService.BatchUnpublish(c.Request.Context(), siteID, req.IDs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, result)
}

// handleError 处理错误
func (h *EntryHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEntryNotFound):
		middleware.NotFound(c, "entry not found")
	case errors.Is(err, service.ErrContentSchemaNotFound):
		middleware.BadRequest(c, "content type not found")
	default:
		middleware.InternalError(c, err.Error())
	}
}
