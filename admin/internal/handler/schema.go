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
	"github.com/contful/contful/admin/pkg/uid"
)

// SchemaHandler 内容模型处理器
type SchemaHandler struct {
	csService *service.SchemaService
}

// NewSchemaHandler 新建处理器
func NewSchemaHandler(csService *service.SchemaService) *SchemaHandler {
	return &SchemaHandler{csService: csService}
}

// RegisterRoutes 注册路由
func (h *SchemaHandler) RegisterRoutes(rg *gin.RouterGroup) {
	contentSchemas := rg.Group("/content/schemas")
	{
		contentSchemas.POST("", h.Create)
		contentSchemas.GET("", h.List)
		contentSchemas.GET("/:id", h.Get)
		contentSchemas.PUT("/:id", h.Update)
		contentSchemas.DELETE("/:id", h.Delete)

		// 字段管理
		contentSchemas.POST("/:id/fields", h.CreateField)
		contentSchemas.GET("/:id/fields", h.ListFields)
		contentSchemas.PUT("/:id/fields/:fieldId", h.UpdateField)
		contentSchemas.DELETE("/:id/fields/:fieldId", h.DeleteField)
		contentSchemas.POST("/:id/fields/reorder", h.ReorderFields)

		// 数据签名/验签
		contentSchemas.POST("/:id/sign", h.Sign)
		contentSchemas.POST("/:id/verify", h.Verify)
	}
}

// Create 创建内容模型
func (h *SchemaHandler) Create(c *gin.Context) {
	var req model.ContentSchemaCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 获取站点 ID（从上下文中）
	siteID := middleware.GetSiteID(c)

	// 获取用户 ID
	userID, _ := middleware.GetUserID(c)

	resp, err := h.csService.Create(c.Request.Context(), siteID, userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.Created(c, resp)
}

// Get 获取内容模型
func (h *SchemaHandler) Get(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	siteID := middleware.GetSiteID(c)
	resp, err := h.csService.Get(c.Request.Context(), siteID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// List 列出内容模型
func (h *SchemaHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	siteID := middleware.GetSiteID(c)
	resp, err := h.csService.List(c.Request.Context(), siteID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// Update 更新内容模型
func (h *SchemaHandler) Update(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	var req model.ContentSchemaUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID := middleware.GetSiteID(c)
	resp, err := h.csService.Update(c.Request.Context(), siteID, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// Delete 删除内容模型
func (h *SchemaHandler) Delete(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	siteID := middleware.GetSiteID(c)
	if err := h.csService.Delete(c.Request.Context(), siteID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ============ Field 操作 ============

// CreateField 创建字段
func (h *SchemaHandler) CreateField(c *gin.Context) {
	contentTypeID, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid content type id format")
		return
	}

	var req model.FieldCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID := middleware.GetSiteID(c)
	resp, err := h.csService.CreateField(c.Request.Context(), siteID, contentTypeID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.Created(c, resp)
}

// ListFields 列出字段
func (h *SchemaHandler) ListFields(c *gin.Context) {
	contentTypeID, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid content type id format")
		return
	}

	siteID := middleware.GetSiteID(c)
	fields, err := h.csService.ListFields(c.Request.Context(), siteID, contentTypeID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, gin.H{"items": fields})
}

// UpdateField 更新字段
func (h *SchemaHandler) UpdateField(c *gin.Context) {
	fieldID, err := uid.Parse(c.Param("fieldId"))
	if err != nil {
		middleware.BadRequest(c, "invalid field id format")
		return
	}

	var req model.FieldUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID := middleware.GetSiteID(c)
	resp, err := h.csService.UpdateField(c.Request.Context(), siteID, fieldID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// DeleteField 删除字段
func (h *SchemaHandler) DeleteField(c *gin.Context) {
	fieldID, err := uid.Parse(c.Param("fieldId"))
	if err != nil {
		middleware.BadRequest(c, "invalid field id format")
		return
	}

	siteID := middleware.GetSiteID(c)
	if err := h.csService.DeleteField(c.Request.Context(), siteID, fieldID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderFields 重新排序字段
func (h *SchemaHandler) ReorderFields(c *gin.Context) {
	contentTypeID, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid content type id format")
		return
	}

	var req struct {
		Orders map[uid.UID]int `json:"orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID := middleware.GetSiteID(c)
	if err := h.csService.ReorderFields(c.Request.Context(), siteID, contentTypeID, req.Orders); err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, nil)
}

// Sign 对内容模型数据重新签名
func (h *SchemaHandler) Sign(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id")
		return
	}

	if err := h.csService.SignSchema(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, gin.H{"message": "签名成功"})
}

// Verify 验签内容模型数据
func (h *SchemaHandler) Verify(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id")
		return
	}

	result, err := h.csService.VerifySchema(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, result)
}

// handleError 处理错误
func (h *SchemaHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrContentSchemaNotFound):
		middleware.NotFound(c, "content type not found")
	case errors.Is(err, service.ErrSlugAlreadyExists):
		middleware.Conflict(c, "slug already exists")
	case errors.Is(err, service.ErrInvalidSlug):
		middleware.BadRequest(c, "invalid slug format: must be lowercase letters, numbers, and hyphens")
	case errors.Is(err, service.ErrCannotChangeKind):
		middleware.BadRequest(c, "cannot change content type kind after creation")
	default:
		middleware.InternalError(c, err.Error())
	}
}
