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

// EntryHandler 条目处理器
type EntryHandler struct {
	entryService *service.EntryService
}

// NewEntryHandler 新建处理器
func NewEntryHandler(entryService *service.EntryService) *EntryHandler {
	return &EntryHandler{entryService: entryService}
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
	}
}

// Create 创建条目
func (h *EntryHandler) Create(c *gin.Context) {
	siteID, _ := middleware.GetSiteID(c)
	if siteID == uuid.Nil {
		siteID = uuid.Nil // 使用默认站点
	}
	userID, _ := middleware.GetUserID(c)

	var req model.EntryCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	entry, err := h.entryService.Create(c.Request.Context(), siteID, userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.Created(c, entry.ToResponse())
}

// Get 获取条目
func (h *EntryHandler) Get(c *gin.Context) {
	siteID, _ := middleware.GetSiteID(c)

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
	siteID, _ := middleware.GetSiteID(c)

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析过滤参数
	filter := &model.EntryListFilter{}
	if ctID := c.Query("content_type_id"); ctID != "" {
		if id, err := uuid.Parse(ctID); err == nil {
			filter.ContentTypeID = &id
		}
	}
	if status := c.Query("status"); status != "" {
		s := model.EntryStatus(status)
		filter.Status = &s
	}
	if locale := c.Query("locale"); locale != "" {
		filter.Locale = &locale
	}

	entries, total, err := h.entryService.List(c.Request.Context(), siteID, filter, page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 转换为响应
	items := make([]model.EntryResponse, len(entries))
	for i, e := range entries {
		items[i] = e.ToResponse()
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
	siteID, _ := middleware.GetSiteID(c)
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

	entry, err := h.entryService.Update(c.Request.Context(), siteID, userID, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// Delete 删除条目
func (h *EntryHandler) Delete(c *gin.Context) {
	siteID, _ := middleware.GetSiteID(c)

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
	siteID, _ := middleware.GetSiteID(c)
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

	entry, err := h.entryService.Publish(c.Request.Context(), siteID, userID, id, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, entry.ToResponse())
}

// Unpublish 取消发布
func (h *EntryHandler) Unpublish(c *gin.Context) {
	siteID, _ := middleware.GetSiteID(c)

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

// GetVersions 获取版本历史
func (h *EntryHandler) GetVersions(c *gin.Context) {
	siteID, _ := middleware.GetSiteID(c)

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

// handleError 处理错误
func (h *EntryHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEntryNotFound):
		middleware.NotFound(c, "entry not found")
	case errors.Is(err, service.ErrContentTypeNotFound):
		middleware.BadRequest(c, "content type not found")
	default:
		middleware.InternalError(c, err.Error())
	}
}
