// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"strconv"
	"time"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/contful/contful/admin/pkg/uid"
)

// ScheduleHandler 排期处理器
type ScheduleHandler struct {
	scheduleService *service.ScheduleService
}

// NewScheduleHandler 新建处理器
func NewScheduleHandler(scheduleService *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

// RegisterRoutes 注册路由
func (h *ScheduleHandler) RegisterRoutes(rg *gin.RouterGroup) {
	entries := rg.Group("/content/entries")
	{
		// 静态路径必须在 :id 之前注册
		entries.GET("/scheduled", h.ListScheduled)
	}
}

// ListScheduled 查询排期条目列表
func (h *ScheduleHandler) ListScheduled(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	if siteID == uid.Nil {
		middleware.BadRequest(c, "X-Site-ID header is required")
		return
	}

	filter := &model.ScheduledEntryFilter{}

	// 解析分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	filter.Page = page
	filter.PageSize = pageSize

	// 解析 status 筛选
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	// 解析时间范围
	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			filter.From = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			filter.To = &t
		}
	}

	entries, total, err := h.scheduleService.ListScheduled(c.Request.Context(), siteID, filter)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	items := make([]model.EntryResponseWithType, len(entries))
	for i, e := range entries {
		items[i] = e.ToResponseWithType()
	}

	middleware.OK(c, model.EntryListResponse{
		Items:    items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	})
}
