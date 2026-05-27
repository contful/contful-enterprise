// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"fmt"
	"net/http"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CacheHandler 缓存管理
type CacheHandler struct {
	cacheSvc *service.CacheService
}

// NewCacheHandler 创建 CacheHandler
func NewCacheHandler(cacheSvc *service.CacheService) *CacheHandler {
	return &CacheHandler{cacheSvc: cacheSvc}
}

// CacheResponse 缓存操作响应
type CacheResponse struct {
	Message string `json:"message"`
	Deleted int64  `json:"deleted"`
}

// InvalidateSite 清除指定站点的所有缓存
// POST /admin/api/v1/cache/invalidate
func (h *CacheHandler) InvalidateSite(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	if siteID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site id is required"})
		return
	}

	deleted, err := h.cacheSvc.InvalidateSite(c.Request.Context(), siteID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CacheResponse{
		Message: "Cache invalidated successfully",
		Deleted: deleted,
	})
}

// InvalidateAll 清除所有 contful 前缀的缓存（超管权限）
// POST /admin/api/v1/cache/invalidate/all
func (h *CacheHandler) InvalidateAll(c *gin.Context) {
	deleted, err := h.cacheSvc.InvalidateAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CacheResponse{
		Message: "All cache invalidated successfully",
		Deleted: deleted,
	})
}

// InvalidateSchema 清除指定内容模型的缓存
// POST /admin/api/v1/cache/invalidate/:slug
func (h *CacheHandler) InvalidateSchema(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	if siteID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "site id is required"})
		return
	}

	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	deleted, err := h.cacheSvc.InvalidateSchema(c.Request.Context(), siteID.String(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CacheResponse{
		Message: fmt.Sprintf("Cache invalidated for schema '%s'", slug),
		Deleted: deleted,
	})
}
