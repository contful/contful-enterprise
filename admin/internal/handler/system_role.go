// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// SystemRoleHandler 系统角色 Handler
type SystemRoleHandler struct {
	rbacService *service.RBACService
	auditService *service.AuditService
}

// NewSystemRoleHandler 新建系统角色 Handler
func NewSystemRoleHandler(rbacService *service.RBACService, auditService *service.AuditService) *SystemRoleHandler {
	return &SystemRoleHandler{rbacService: rbacService, auditService: auditService}
}

// List GET /admin/api/v1/system/roles
func (h *SystemRoleHandler) List(c *gin.Context) {
	roles, err := h.rbacService.ListSystemRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(roles))
}

// Get GET /admin/api/v1/system/roles/:id
func (h *SystemRoleHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid role id"))
		return
	}

	role, err := h.rbacService.GetSystemRole(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "role not found"))
		} else {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(role))
}

// Create POST /admin/api/v1/system/roles
func (h *SystemRoleHandler) Create(c *gin.Context) {
	var req model.SystemRoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	role, err := h.rbacService.CreateSystemRole(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrRoleNameExists:
			c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, "role name already exists"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}
	c.JSON(http.StatusCreated, model.NewSuccessResponse(role))

	// 审计日志：记录角色创建操作
	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, "role:create",
			service.WithResource("role", role.ID))
	}
}

// Update PUT /admin/api/v1/system/roles/:id
func (h *SystemRoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid role id"))
		return
	}

	var req model.SystemRoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	role, err := h.rbacService.UpdateSystemRole(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrRoleNotFound:
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "role not found"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(role))

	// 审计日志：记录角色更新操作
	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, "role:update",
			service.WithResource("role", id))
	}
}

// Delete DELETE /admin/api/v1/system/roles/:id
func (h *SystemRoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid role id"))
		return
	}

	if err := h.rbacService.DeleteSystemRole(c.Request.Context(), id); err != nil {
		switch err {
		case service.ErrRoleNotFound:
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "role not found"))
		case service.ErrRoleIsSystem:
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "cannot delete system built-in role"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}
	c.JSON(http.StatusNoContent, nil)

	// 审计日志：记录角色删除操作
	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, "role:delete",
			service.WithResource("role", id))
	}
}

// Permissions GET /admin/api/v1/system/roles/permissions（权限树元数据）
func (h *SystemRoleHandler) Permissions(c *gin.Context) {
	meta := h.rbacService.GetPermissionsMeta()
	c.JSON(http.StatusOK, model.NewSuccessResponse(meta))
}

// ─────────────────────────────────────────────────────────────
// 从 gin context 获取当前用户 ID 的辅助函数（跨 handler 复用）
// ─────────────────────────────────────────────────────────────

// getCurrentUserID 从 JWT Claims 获取当前用户 ID
func getCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	claimsVal, exists := c.Get(middleware.ClaimsContextKey)
	if !exists {
		return uuid.Nil, false
	}
	claims, ok := claimsVal.(*middleware.Claims)
	if !ok {
		return uuid.Nil, false
	}
	return claims.UserID, true
}
