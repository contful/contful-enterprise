// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful-enterprise/shared/uid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// AnomalyHandler 异常事件处理器
type AnomalyHandler struct {
	anomalySvc *service.AnomalyService
	scanner    *service.Scanner
}

// NewAnomalyHandler 创建异常事件处理器
func NewAnomalyHandler(anomalySvc *service.AnomalyService, scanner *service.Scanner) *AnomalyHandler {
	return &AnomalyHandler{anomalySvc: anomalySvc, scanner: scanner}
}

// List 异常事件列表（分页）
func (h *AnomalyHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := parseAnomalyFilter(c)

	anomalies, total, err := h.anomalySvc.List(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "查询异常事件失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.PageResponse{
		Items:      anomalies,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}))
}

// Get 异常事件详情
func (h *AnomalyHandler) Get(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 ID"))
		return
	}

	anomaly, err := h.anomalySvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "异常事件未找到"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(anomaly))
}

// Summary 异常概览仪表盘
func (h *AnomalyHandler) Summary(c *gin.Context) {
	summary, err := h.anomalySvc.Summary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "获取异常概览失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(summary))
}

// TriggerScan 手动触发扫描
func (h *AnomalyHandler) TriggerScan(c *gin.Context) {
	if err := h.scanner.ScanNow(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "扫描失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "扫描已触发"}))
}

// parseAnomalyFilter 从查询参数解析异常事件筛选条件
func parseAnomalyFilter(c *gin.Context) *model.AnomalyFilter {
	filter := &model.AnomalyFilter{}

	if anomalyType := c.Query("anomaly_type"); anomalyType != "" {
		filter.AnomalyType = model.AnomalyType(anomalyType)
	}
	if severity := c.Query("severity"); severity != "" {
		filter.Severity = model.AnomalySeverity(severity)
	}
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = t
		}
	}

	return filter
}
