// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"net/http"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	dashboardSvc *service.DashboardService
}

// NewDashboardHandler 新建仪表盘处理器
func NewDashboardHandler(dashboardSvc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardSvc: dashboardSvc}
}

// Stats 获取仪表盘统计（不依赖 X-Site-ID，如已选站点则附带站点相关统计）
func (h *DashboardHandler) Stats(c *gin.Context) {
	// 尝试从 header 读取可选 siteID
	var siteID *uuid.UUID
	if siteIDStr := c.GetHeader(middleware.SiteIDHeader); siteIDStr != "" {
		if id, err := uuid.Parse(siteIDStr); err == nil {
			siteID = &id
		}
	}

	stats := h.dashboardSvc.GetStats(c.Request.Context(), siteID)

	c.JSON(http.StatusOK, model.NewSuccessResponse(stats))
}
