// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"fmt"
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

	filter := parseAuditFilter(c)

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

// ExportCSV 导出审计日志为 CSV 文件
// GET /admin/api/v1/audit/logs/export/csv
func (h *AuditHandler) ExportCSV(c *gin.Context) {
	filter := parseAuditFilter(c)

	maxRows, _ := strconv.Atoi(c.DefaultQuery("max_rows", "50000"))
	if maxRows < 1 || maxRows > 100000 {
		maxRows = 50000
	}

	csvBytes, count, total, err := h.auditService.ExportCSV(c.Request.Context(), filter, maxRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export audit logs"})
		return
	}

	filename := buildExportFilename(filter)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("X-Export-Count", strconv.FormatInt(count, 10))
	c.Header("X-Export-Total", strconv.FormatInt(total, 10))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

func buildExportFilename(filter *model.AuditLogFilter) string {
	category := "all"
	if filter.Category != "" {
		category = string(filter.Category)
	}
	return fmt.Sprintf("audit_export_%s_%s.csv", category, time.Now().Format("20060102"))
}

// parseAuditFilter 从查询参数解析审计日志筛选条件
func parseAuditFilter(c *gin.Context) *model.AuditLogFilter {
	filter := &model.AuditLogFilter{}

	if siteIDStr := c.Query("site_id"); siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			filter.SiteID = &siteID
		}
	}
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}
	if action := c.Query("action"); action != "" {
		filter.Action = action
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filter.ResourceType = resourceType
	}
	if category := c.Query("category"); category != "" {
		filter.Category = model.AuditType(category)
	}
	if level := c.Query("level"); level != "" {
		filter.Level = model.AuditLevel(level)
	}
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

	return filter
}
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
