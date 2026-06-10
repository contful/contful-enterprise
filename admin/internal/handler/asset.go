// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful/admin/pkg/uid"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// AssetHandler 资源处理器
type AssetHandler struct {
	assetService *service.AssetService
}

// NewAssetHandler 新建处理器
func NewAssetHandler(assetService *service.AssetService) *AssetHandler {
	return &AssetHandler{assetService: assetService}
}

// RegisterRoutes 注册路由
func (h *AssetHandler) RegisterRoutes(rg *gin.RouterGroup) {
	assets := rg.Group("/assets")
	{
		assets.POST("", h.Upload)
		assets.GET("", h.List)
		assets.GET("/:id", h.Get)
		assets.PUT("/:id", h.Update)
		assets.DELETE("/:id", h.Delete)
		assets.DELETE("/batch", h.BatchDelete)

		// 文件夹
		assets.POST("/folders", h.CreateFolder)
		assets.GET("/folders/tree", h.GetFolderTree)
		assets.GET("/folders", h.ListFolders)
		assets.GET("/folders/:id", h.GetFolder)
		assets.PUT("/folders/:id", h.UpdateFolder)
		assets.DELETE("/folders/:id", h.DeleteFolder)
	}
}

// getUserID 从上下文获取用户 ID
func getUserID(c *gin.Context) (uid.UID, error) {
	userIDVal, exists := c.Get("user")
	if !exists {
		return uid.Nil, errors.New("user not found")
	}

	switch v := userIDVal.(type) {
	case string:
		return uid.Parse(v)
	case uid.UID:
		return v, nil
	default:
		return uid.Nil, errors.New("invalid user id type")
	}
}

// getSiteID 从上下文获取站点 ID（通过 middleware.GetSiteID）
func getSiteID(c *gin.Context) (uid.UID, error) {
	siteID := middleware.GetSiteID(c)
	if siteID == uid.Nil {
		return uid.Nil, errors.New("site_id not found in context, X-Site-ID header required")
	}
	return siteID, nil
}

// Upload 上传资源
func (h *AssetHandler) Upload(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	siteID, err := getSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请选择要上传的文件"))
		return
	}

	// 获取其他参数
	folderIDStr := c.PostForm("folder_id")
	var folderID *uid.UID
	if folderIDStr != "" {
		id, err := uid.Parse(folderIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 folder_id"))
			return
		}
		folderID = &id
	}

	alt := c.PostForm("alt")
	title := c.PostForm("title")

	asset, err := h.assetService.Upload(c.Request.Context(), siteID, userID, file, folderID, alt, title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(asset.ToResponse()))
}

// List 列出资源
func (h *AssetHandler) List(c *gin.Context) {
	siteID, err := getSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建过滤条件
	filter := &model.AssetListFilter{}

	if folderIDStr := c.Query("folder_id"); folderIDStr != "" {
		id, err := uid.Parse(folderIDStr)
		if err == nil {
			filter.FolderID = &id
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		t := model.AssetType(typeStr)
		filter.Type = &t
	}

	if ext := c.Query("extension"); ext != "" {
		filter.Extension = &ext
	}

	if tag := c.Query("tag"); tag != "" {
		filter.Tag = &tag
	}

	if keyword := c.Query("keyword"); keyword != "" {
		filter.Keyword = &keyword
	}

	result, err := h.assetService.List(c.Request.Context(), siteID, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// Get 获取资源
func (h *AssetHandler) Get(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的资源 ID"))
		return
	}

	asset, err := h.assetService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "资源不存在"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(asset.ToResponse()))
}

// Update 更新资源
func (h *AssetHandler) Update(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的资源 ID"))
		return
	}

	var req model.AssetUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请求参数错误"))
		return
	}

	asset, err := h.assetService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(asset.ToResponse()))
}

// Delete 删除资源
func (h *AssetHandler) Delete(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的资源 ID"))
		return
	}

	if err := h.assetService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// ServeFile 提供静态文件服务（支持 /uploads/* 路径访问媒体文件）
// 该路由需要在 router 注册时放在 /assets/* 之前，因为 /uploads 是独立的路径
func (h *AssetHandler) ServeFile(c *gin.Context) {
	siteID, err := getSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	// 获取文件路径（从 URL 参数 uploads/* 获取）
	filePath := c.Param("filePath")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "缺少文件路径"))
		return
	}

	// 防止路径遍历攻击
	filePath = sanitizeFilePath(filePath)

	// 调用服务层获取文件
	reader, mimeType, err := h.assetService.ServeFile(c.Request.Context(), siteID, filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "文件不存在或无法访问"))
		return
	}
	defer reader.Close()

	// 设置响应头
	c.Header("Content-Type", mimeType)
	c.Header("Cache-Control", "public, max-age=31536000") // 缓存 1 年

	// 流式传输文件
	c.DataFromReader(http.StatusOK, -1, mimeType, reader, nil)
}

// sanitizeFilePath 清理文件路径，防止路径遍历攻击
func sanitizeFilePath(path string) string {
	// 移除开头的 /
	path = strings.TrimPrefix(path, "/")

	// 移除潜在的路径遍历字符
	path = strings.ReplaceAll(path, "..", "")
	path = strings.ReplaceAll(path, "\\", "")

	return path
}

// BatchDelete 批量删除
func (h *AssetHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uid.UID `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请提供要删除的资源 ID 列表"))
		return
	}

	if err := h.assetService.BatchDelete(c.Request.Context(), req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// ============ Folder Handlers ============

// CreateFolder 创建文件夹
func (h *AssetHandler) CreateFolder(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	siteID, err := getSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var req model.FolderCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请求参数错误"))
		return
	}

	folder, err := h.assetService.CreateFolder(c.Request.Context(), siteID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(folder.ToFolderResponse()))
}

// GetFolderTree 获取文件夹树
func (h *AssetHandler) GetFolderTree(c *gin.Context) {
	siteID, err := getSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	folders, err := h.assetService.GetFolderTree(c.Request.Context(), siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(folders))
}

// ListFolders 列出文件夹
func (h *AssetHandler) ListFolders(c *gin.Context) {
	siteID, err := getSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var parentID *uid.UID
	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		id, err := uid.Parse(parentIDStr)
		if err == nil {
			parentID = &id
		}
	}

	folders, err := h.assetService.ListFolders(c.Request.Context(), siteID, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(folders))
}

// GetFolder 获取文件夹
func (h *AssetHandler) GetFolder(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的文件夹 ID"))
		return
	}

	folder, err := h.assetService.GetFolder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "文件夹不存在"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(folder.ToFolderResponse()))
}

// UpdateFolder 更新文件夹
func (h *AssetHandler) UpdateFolder(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的文件夹 ID"))
		return
	}

	var req model.FolderUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请求参数错误"))
		return
	}

	folder, err := h.assetService.UpdateFolder(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(folder.ToFolderResponse()))
}

// DeleteFolder 删除文件夹
func (h *AssetHandler) DeleteFolder(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的文件夹 ID"))
		return
	}

	if err := h.assetService.DeleteFolder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
