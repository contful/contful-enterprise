// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// SetupHandler 安装向导 HTTP Handler。
type SetupHandler struct {
	svc *service.SetupService
}

// NewSetupHandler 创建 SetupHandler 实例。
func NewSetupHandler(svc *service.SetupService) *SetupHandler {
	return &SetupHandler{svc: svc}
}

// Status 检查安装状态。
// GET /admin/api/v1/setup/status
func (h *SetupHandler) Status(c *gin.Context) {
	status, err := h.svc.CheckStatus()
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(status))
}

// TestDatabase 测试数据库连接。
// POST /admin/api/v1/setup/database
func (h *SetupHandler) TestDatabase(c *gin.Context) {
	var req model.SetupDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.TestDatabase(&req); err != nil {
		c.JSON(http.StatusOK, model.NewSuccessResponse(model.SetupDatabaseResponse{
			Success: false,
			Message: err.Error(),
		}))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.SetupDatabaseResponse{
		Success: true,
		Message: "数据库连接成功",
	}))
}

// Initialize 初始化数据库（执行 init_pg.sql）。
// POST /admin/api/v1/setup/initialize
func (h *SetupHandler) Initialize(c *gin.Context) {
	var req model.SetupDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.InitializeDatabase(&req); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
		"success": true,
		"message": "数据库初始化完成",
	}))
}

// CreateAdmin 创建管理员账号 + 默认站点 + 标记安装完成。
// POST /admin/api/v1/setup/admin
func (h *SetupHandler) CreateAdmin(c *gin.Context) {
	var req model.SetupAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	if err := h.svc.CreateAdmin(&req); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}

	// 销毁 CSRF session（安装完成，不再需要）
	middleware.InvalidateCSRFToken(c)

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.SetupCompleteResponse{
		Success: true,
		Message: "安装完成",
	}))
}
