// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// List 获取审计日志列表（支持筛选和分页）
func (h *AuditHandler) List(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建筛选条件
	filter := &model.AuditLogFilter{}
	
	// 站点 ID（可选）
	if siteIDStr := c.Query("site_id"); siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			filter.SiteID = &siteID
		}
	}
	
	// 用户 ID（可选）
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}
	
	// 操作动作（可选）
	if action := c.Query("action"); action != "" {
		filter.Action = action
	}
	
	// 资源类型（可选）
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filter.ResourceType = resourceType
	}
	
	// 日志类别（可选）
	if category := c.Query("category"); category != "" {
		filter.Category = model.AuditType(category)
	}
	
	// 日志级别（可选）
	if level := c.Query("level"); level != "" {
		filter.Level = model.AuditLevel(level)
	}
	
	// 时间范围（可选）
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = endTime
		}
	}

	// 查询列表
	logs, total, err := h.auditService.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit logs"})
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.AuditLogListResponse{
		Items:    logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}))
}

// Get 获取审计日志详情
func (h *AuditHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	log, err := h.auditService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(log))
}
