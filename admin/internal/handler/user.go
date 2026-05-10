// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	auditService *service.AuditService
}

func NewUserHandler(userService *service.UserService, auditService *service.AuditService) *UserHandler {
	return &UserHandler{userService: userService, auditService: auditService}
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, "user already exists"))
		case service.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "password must be at least 8 characters with uppercase, lowercase and numbers"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(user))

	// 审计日志：记录用户创建操作
	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, "user:create",
			service.WithResource("user", user.ID))
	}
}

// Get 获取单个用户
func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid user id"))
		return
	}

	user, err := h.userService.Get(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "user not found"))
		} else {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))
}

// List 分页列表（可包含已删除记录）
func (h *UserHandler) List(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	// 支持 ?include_deleted=true 参数
	includeDeleted := c.Query("include_deleted") == "true"

	pageResp, err := h.userService.List(c.Request.Context(), req.Page, req.PageSize, includeDeleted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(pageResp))
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid user id"))
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "user not found"))
		} else {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))

	// 审计日志：记录用户更新操作
	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, "user:update",
			service.WithResource("user", id))
	}
}

// Delete 删除用户（支持软删除和永久删除）
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid user id"))
		return
	}

	// 支持 ?permanent=true 参数进行永久删除
	permanent := c.Query("permanent") == "true"

	if err := h.userService.Delete(c.Request.Context(), id, permanent); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusNoContent, nil)

	// 审计日志：记录用户删除操作
	if h.auditService != nil {
		action := "user:delete"
		if permanent {
			action = "user:permanent_delete"
		}
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, action,
			service.WithResource("user", id))
	}
}

// Restore 恢复软删除的用户
func (h *UserHandler) Restore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid user id"))
		return
	}

	if err := h.userService.Restore(c.Request.Context(), id); err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "user not found"))
		} else {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))

	// 审计日志：记录用户恢复操作
	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, "user:restore",
			service.WithResource("user", id))
	}
}

// UpdateMe 用户更新自己的资料
// PATCH /users/me
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userIDStr, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var userID uuid.UUID
	switch v := userIDStr.(type) {
	case string:
		uid, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
			return
		}
		userID = uid
	case uuid.UUID:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
		return
	}

	var req model.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	user, err := h.userService.UpdateMe(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))
}

// UpdatePassword 用户修改自己的密码
// PUT /users/me/password
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userIDStr, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var userID uuid.UUID
	switch v := userIDStr.(type) {
	case string:
		uid, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
			return
		}
		userID = uid
	case uuid.UUID:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
		return
	}

	var req model.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	if err := h.userService.UpdatePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		switch err {
		case service.ErrInvalidPassword:
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid old password"))
		case service.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "password must be at least 8 characters with uppercase, lowercase and numbers"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// ResetPassword 管理员重置用户密码（不需要旧密码）
func (h *UserHandler) ResetPassword(c *gin.Context) {
	// 只有超级管理员可以重置其他用户的密码
	userIDStr, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var adminID uuid.UUID
	switch v := userIDStr.(type) {
	case string:
		uid, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
			return
		}
		adminID = uid
	case uuid.UUID:
		adminID = v
	default:
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
		return
	}

	// 获取目标用户 ID
	targetIDStr := c.Param("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid target user id"))
		return
	}

	// 不能重置自己的密码（应该用 UpdatePassword）
	if adminID == targetID {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "use UpdatePassword to change your own password"))
		return
	}

	// 检查操作者是否是超级管理员
	admin, err := h.userService.Get(c.Request.Context(), adminID)
	if err != nil || !admin.IsSuperAdmin {
		c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "only super admin can reset password"))
		return
	}

	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), targetID, req.NewPassword); err != nil {
		switch err {
		case service.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "password must be at least 8 characters with uppercase, lowercase and numbers"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	// 记录审计日志
	if h.auditService != nil {
		_ = h.auditService.LogFromGin(c, model.AuditLevelInfo, model.AuditTypeUser, "user:reset_password",
			service.WithResource("user", targetID),
			service.WithDetails("管理员重置用户密码"),
		)
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}
