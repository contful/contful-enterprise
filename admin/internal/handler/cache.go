// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package handler

import (
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

// InvalidateSite 清除指定站点的所有内容缓存
// POST /admin/api/v1/cache/invalidate
func (h *CacheHandler) InvalidateSite(c *gin.Context) {
	siteID := middleware.GetSiteID(c)

	// uuid.UUID 不能直接和 "" 比较，需检查是否为 Nil
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
