// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful-enterprise/shared/uid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService *service.AuditService
	exportService *service.ExportService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditService *service.AuditService, exportService *service.ExportService) *AuditHandler {
	return &AuditHandler{auditService: auditService, exportService: exportService}
}

// List 获取审计日志列表（支持筛选和分页）
func (h *AuditHandler) List(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "搜索审计日志失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.PageResponse{
		Items:      logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
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
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "导出审计日志失败"))
		return
	}

	filename := buildExportFilename(filter)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("X-Export-Count", strconv.FormatInt(count, 10))
	c.Header("X-Export-Total", strconv.FormatInt(total, 10))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}

// ExportXLSX 导出审计日志为 XLSX 文件（含条件着色 + 完整性声明 sheet）
// GET /admin/api/v1/audit/logs/export/xlsx
func (h *AuditHandler) ExportXLSX(c *gin.Context) {
	filter := parseAuditFilter(c)

	maxRows, _ := strconv.Atoi(c.DefaultQuery("max_rows", "50000"))
	if maxRows < 1 || maxRows > 100000 {
		maxRows = 50000
	}

	xlsxBytes, count, total, err := h.auditService.ExportXLSX(c.Request.Context(), filter, maxRows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "导出审计日志失败"))
		return
	}

	filename := buildExportFilenameExt(filter, "xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("X-Export-Count", strconv.FormatInt(count, 10))
	c.Header("X-Export-Total", strconv.FormatInt(total, 10))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxBytes)
}

// ExportJSON 流式 JSON (NDJSON) 导出
// GET /admin/api/v1/audit/logs/export/json
func (h *AuditHandler) ExportJSON(c *gin.Context) {
	filter := parseAuditFilter(c)

	maxRows, _ := strconv.Atoi(c.DefaultQuery("max_rows", "50000"))
	if maxRows < 1 || maxRows > 100000 {
		maxRows = 50000
	}

	// 设置响应头：NDJSON 流式输出
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)

	count, _, err := h.exportService.ExportJSONStream(c.Request.Context(), filter, maxRows, c.Writer)
	if err != nil {
		// 注意：此时 header 已写入，无法再返回 JSON 错误响应
		// 仅记录日志
		return
	}

	// 在响应尾部追加计数元数据行
	metaLine := fmt.Sprintf(`{"_export_count":%d,"_export_complete":true}`, count)
	c.Writer.Write([]byte(metaLine))
	c.Writer.Write([]byte("\n"))
}

func buildExportFilename(filter *model.AuditLogFilter) string {
	return buildExportFilenameExt(filter, "csv")
}

func buildExportFilenameExt(filter *model.AuditLogFilter, ext string) string {
	if filter == nil {
		return "audit_export_all." + ext
	}
	category := "all"
	if filter.Category != "" {
		category = string(filter.Category)
	}
	return fmt.Sprintf("audit_export_%s_%s.%s", category, time.Now().Format("20060102"), ext)
}

// parseAuditFilter 从查询参数解析审计日志筛选条件
func parseAuditFilter(c *gin.Context) *model.AuditLogFilter {
	filter := &model.AuditLogFilter{}

	if siteIDStr := c.Query("site_id"); siteIDStr != "" {
		if siteID, err := uid.Parse(siteIDStr); err == nil {
			filter.SiteID = &siteID
		}
	}
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uid.Parse(userIDStr); err == nil {
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

	// 新增高级查询参数
	if keyword := c.Query("keyword"); keyword != "" {
		filter.Keyword = keyword
	}
	if combinator := c.Query("combinator"); combinator != "" {
		if combinator == "and" || combinator == "or" {
			filter.Combinator = combinator
		}
	}
	if timePreset := c.Query("time_preset"); timePreset != "" {
		filter.TimePreset = timePreset
		// time_preset 与 start_time/end_time 互斥
		// 如果提供了 time_preset，自动计算 start_time
		if filter.StartTime.IsZero() {
			now := time.Now()
			switch timePreset {
			case "1h":
				filter.StartTime = now.Add(-1 * time.Hour)
			case "24h":
				filter.StartTime = now.Add(-24 * time.Hour)
			case "7d":
				filter.StartTime = now.Add(-7 * 24 * time.Hour)
			}
		}
	}

	// 导出字段选择
	if fields := c.Query("fields"); fields != "" {
		filter.Fields = strings.Split(fields, ",")
	}

	return filter
}

func (h *AuditHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 ID"))
		return
	}

	log, err := h.auditService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "审计日志未找到"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(log))
}
