// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful-enterprise/shared/uid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
)

// QueryTemplateHandler 查询模板处理器
type QueryTemplateHandler struct {
	templateRepo *repository.QueryTemplateRepository
	auditService *service.AuditService
}

// NewQueryTemplateHandler 创建查询模板处理器
func NewQueryTemplateHandler(templateRepo *repository.QueryTemplateRepository, auditService *service.AuditService) *QueryTemplateHandler {
	return &QueryTemplateHandler{templateRepo: templateRepo, auditService: auditService}
}

// List 获取用户的查询模板列表
func (h *QueryTemplateHandler) List(c *gin.Context) {
	userID := getCurrentUID(c)
	if userID.IsNil() {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "用户未认证"))
		return
	}

	templates, err := h.templateRepo.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "查询模板列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(templates))
}

// Create 创建查询模板
func (h *QueryTemplateHandler) Create(c *gin.Context) {
	userID := getCurrentUID(c)
	if userID.IsNil() {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "用户未认证"))
		return
	}

	var req struct {
		Name       string          `json:"name"`
		Conditions json.RawMessage `json:"conditions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请求参数无效"))
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "模板名称不能为空"))
		return
	}

	template := &model.AuditQueryTemplate{
		Name:       req.Name,
		Conditions: req.Conditions,
		CreatedBy:  &userID,
	}

	if err := h.templateRepo.Create(c.Request.Context(), template); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "创建模板失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(template))
}

// Delete 删除模板
func (h *QueryTemplateHandler) Delete(c *gin.Context) {
	userID := getCurrentUID(c)
	if userID.IsNil() {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "用户未认证"))
		return
	}

	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 ID"))
		return
	}

	if err := h.templateRepo.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "删除模板失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// Apply 应用模板（返回查询条件 + 结果第一页）
func (h *QueryTemplateHandler) Apply(c *gin.Context) {
	id, err := uid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 ID"))
		return
	}

	template, err := h.templateRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "模板未找到"))
		return
	}

	// 解析模板中的查询条件
	var conditions model.AuditTemplateConditions
	if err := json.Unmarshal(template.Conditions, &conditions); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "模板条件解析失败"))
		return
	}

	// 构建 filter
	filter := &model.AuditLogFilter{
		Keyword:    conditions.Keyword,
		Combinator: conditions.Combinator,
		TimePreset: conditions.TimePreset,
	}
	if conditions.Category != "" {
		filter.Category = model.AuditType(conditions.Category)
	}
	if conditions.Level != "" {
		filter.Level = model.AuditLevel(conditions.Level)
	}
	if conditions.Action != "" {
		filter.Action = conditions.Action
	}

	// time_preset 转换——使用当前时间计算时间范围
	if filter.TimePreset != "" && filter.StartTime.IsZero() {
		now := time.Now()
		switch filter.TimePreset {
		case "1h":
			filter.StartTime = now.Add(-1 * time.Hour)
		case "24h":
			filter.StartTime = now.Add(-24 * time.Hour)
		case "7d":
			filter.StartTime = now.Add(-7 * 24 * time.Hour)
		}
	}

	// 使用模板条件查询第一页
	logs, total, err := h.auditService.List(c.Request.Context(), filter, 1, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.TemplateApplyResponse{
		Template: template,
		Results: &model.AuditLogListResponse{
			Items:    logs,
			Total:    total,
			Page:     1,
			PageSize: 20,
		},
	}))
}

// getCurrentUID 从 context 中提取用户 ID（与 asset.go 的 getUserID 区分）
func getCurrentUID(c *gin.Context) uid.UID {
	if userIDVal, exists := c.Get("user"); exists {
		if uid, ok := userIDVal.(uid.UID); ok {
			return uid
		}
	}
	return uid.Nil
}

func parseTimeStr(s string) (time.Time, error) {
	// 支持 RFC3339 格式
	t, err := time.Parse(time.RFC3339, s)
	return t, err
}
