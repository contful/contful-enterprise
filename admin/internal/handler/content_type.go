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

// ContentTypeHandler 内容类型处理器
type ContentTypeHandler struct {
	ctService *service.ContentTypeService
}

// NewContentTypeHandler 新建处理器
func NewContentTypeHandler(ctService *service.ContentTypeService) *ContentTypeHandler {
	return &ContentTypeHandler{ctService: ctService}
}

// RegisterRoutes 注册路由
func (h *ContentTypeHandler) RegisterRoutes(rg *gin.RouterGroup) {
	contentTypes := rg.Group("/content-types")
	{
		contentTypes.POST("", h.Create)
		contentTypes.GET("", h.List)
		contentTypes.GET("/:id", h.Get)
		contentTypes.PUT("/:id", h.Update)
		contentTypes.DELETE("/:id", h.Delete)

		// 字段管理
		contentTypes.POST("/:id/fields", h.CreateField)
		contentTypes.GET("/:id/fields", h.ListFields)
		contentTypes.PUT("/fields/:fieldId", h.UpdateField)
		contentTypes.DELETE("/fields/:fieldId", h.DeleteField)
		contentTypes.POST("/:id/fields/reorder", h.ReorderFields)
	}
}

// Create 创建内容类型
// @Summary 创建内容类型
// @Tags ContentTypes
// @Accept json
// @Produce json
// @Param request body model.ContentTypeCreate true "创建请求"
// @Success 201 {object} model.Response{data=model.ContentTypeResponse}
// @Router /admin/v1/content-types [post]
func (h *ContentTypeHandler) Create(c *gin.Context) {
	var req model.ContentTypeCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 获取站点 ID（从上下文中）
	siteID, ok := middleware.GetSiteID(c)
	if !ok {
		siteID = uuid.Nil // 使用默认站点
	}

	// 获取用户 ID
	userID, _ := middleware.GetUserID(c)

	resp, err := h.ctService.Create(c.Request.Context(), siteID, userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.Created(c, resp)
}

// Get 获取内容类型
// @Summary 获取内容类型
// @Tags ContentTypes
// @Produce json
// @Param id path string true "内容类型ID"
// @Success 200 {object} model.Response{data=model.ContentTypeResponse}
// @Router /admin/v1/content-types/{id} [get]
func (h *ContentTypeHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	resp, err := h.ctService.Get(c.Request.Context(), siteID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// List 列出内容类型
// @Summary 列出内容类型
// @Tags ContentTypes
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} model.Response{data=model.ContentTypeListResponse}
// @Router /admin/v1/content-types [get]
func (h *ContentTypeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	siteID, _ := middleware.GetSiteID(c)
	resp, err := h.ctService.List(c.Request.Context(), siteID, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// Update 更新内容类型
// @Summary 更新内容类型
// @Tags ContentTypes
// @Accept json
// @Produce json
// @Param id path string true "内容类型ID"
// @Param request body model.ContentTypeUpdate true "更新请求"
// @Success 200 {object} model.Response{data=model.ContentTypeResponse}
// @Router /admin/v1/content-types/{id} [put]
func (h *ContentTypeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	var req model.ContentTypeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	resp, err := h.ctService.Update(c.Request.Context(), siteID, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// Delete 删除内容类型
// @Summary 删除内容类型
// @Tags ContentTypes
// @Param id path string true "内容类型ID"
// @Success 204
// @Router /admin/v1/content-types/{id} [delete]
func (h *ContentTypeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	if err := h.ctService.Delete(c.Request.Context(), siteID, id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ============ Field 操作 ============

// CreateField 创建字段
func (h *ContentTypeHandler) CreateField(c *gin.Context) {
	contentTypeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid content type id format")
		return
	}

	var req model.FieldCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	resp, err := h.ctService.CreateField(c.Request.Context(), siteID, contentTypeID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.Created(c, resp)
}

// ListFields 列出字段
func (h *ContentTypeHandler) ListFields(c *gin.Context) {
	contentTypeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid content type id format")
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	fields, err := h.ctService.ListFields(c.Request.Context(), siteID, contentTypeID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, gin.H{"items": fields})
}

// UpdateField 更新字段
func (h *ContentTypeHandler) UpdateField(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		middleware.BadRequest(c, "invalid field id format")
		return
	}

	var req model.FieldUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	resp, err := h.ctService.UpdateField(c.Request.Context(), siteID, fieldID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, resp)
}

// DeleteField 删除字段
func (h *ContentTypeHandler) DeleteField(c *gin.Context) {
	fieldID, err := uuid.Parse(c.Param("fieldId"))
	if err != nil {
		middleware.BadRequest(c, "invalid field id format")
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	if err := h.ctService.DeleteField(c.Request.Context(), siteID, fieldID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderFields 重新排序字段
func (h *ContentTypeHandler) ReorderFields(c *gin.Context) {
	contentTypeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.BadRequest(c, "invalid content type id format")
		return
	}

	var req struct {
		Orders map[uuid.UUID]int `json:"orders" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	siteID, _ := middleware.GetSiteID(c)
	if err := h.ctService.ReorderFields(c.Request.Context(), siteID, contentTypeID, req.Orders); err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, nil)
}

// handleError 处理错误
func (h *ContentTypeHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrContentTypeNotFound):
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
