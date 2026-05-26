// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

type AuditExportHandler struct {
	exportService *service.AuditExportService
}

func NewAuditExportHandler(exportService *service.AuditExportService) *AuditExportHandler {
	return &AuditExportHandler{exportService: exportService}
}

// Create 创建导出任务（异步）
// POST /admin/api/v1/audit/exports
func (h *AuditExportHandler) Create(c *gin.Context) {
	var req struct {
		MaxRows int    `json:"max_rows"`
		Format  string `json:"format"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.MaxRows = 50000
		req.Format = "csv"
	}
	if req.MaxRows < 1 || req.MaxRows > 100000 {
		req.MaxRows = 50000
	}
	if req.Format == "" {
		req.Format = "csv"
	}

	filter := parseAuditFilter(c)
	userID, _ := c.Get("user")

	report, err := h.exportService.CreateExport(c.Request.Context(), filter, req.MaxRows, req.Format, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建导出任务失败"})
		return
	}

	c.JSON(http.StatusAccepted, model.NewSuccessResponse(report))
}

// List 导出任务列表
// GET /admin/api/v1/audit/exports
func (h *AuditExportHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	reports, total, err := h.exportService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询导出任务失败"})
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.AuditReportExportListResponse{
		Items:    reports,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}))
}

// Get 查询导出任务状态
// GET /admin/api/v1/audit/exports/:id
func (h *AuditExportHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	report, err := h.exportService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(report))
}

// Download 下载已完成的导出文件
// GET /admin/api/v1/audit/exports/:id/download
func (h *AuditExportHandler) Download(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	report, err := h.exportService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "export not found"})
		return
	}

	if report.Status != model.ReportStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "export not completed yet"})
		return
	}

	if _, err := os.Stat(report.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "export file not found on disk"})
		return
	}

	c.File(report.FilePath)
}

// Delete 删除导出任务及文件
// DELETE /admin/api/v1/audit/exports/:id
func (h *AuditExportHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.exportService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"deleted": id.String()}))
}
