// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// MFAHandler MFA 双因子认证 Handler
type MFAHandler struct {
	mfaService  *service.MFAService
	authService *service.AuthService
}

func NewMFAHandler(mfaService *service.MFAService, authService *service.AuthService) *MFAHandler {
	return &MFAHandler{
		mfaService:  mfaService,
		authService: authService,
	}
}

// getUserIDFromContext 从 JWT 中间件注入的 context 获取当前用户 ID
func getUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("user")
	if !exists {
		return uuid.Nil, false
	}
	switch v := val.(type) {
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, false
		}
		return id, true
	case uuid.UUID:
		return v, true
	}
	return uuid.Nil, false
}

// Setup 生成 TOTP Secret + QR Code
// POST /admin/api/v1/auth/mfa/setup
func (h *MFAHandler) Setup(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	resp, err := h.mfaService.Setup(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrMFAAlreadyEnabled) {
			c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, "MFA is already enabled"))
			return
		}
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to setup MFA"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
}

// Enable 验证 TOTP 码，启用 MFA
// POST /admin/api/v1/auth/mfa/enable
func (h *MFAHandler) Enable(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var req model.MFAEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	resp, err := h.mfaService.Enable(c.Request.Context(), userID, req.TOTPCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, "MFA is already enabled"))
		case errors.Is(err, service.ErrMFANotSetup):
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "MFA has not been set up"))
		case errors.Is(err, service.ErrMFAInvalidCode):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid TOTP code"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to enable MFA"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
}

// Disable 验证 TOTP 码，关闭 MFA
// POST /admin/api/v1/auth/mfa/disable
func (h *MFAHandler) Disable(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var req model.MFADisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	err := h.mfaService.Disable(c.Request.Context(), userID, req.TOTPCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFANotEnabled):
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "MFA is not enabled"))
		case errors.Is(err, service.ErrMFAInvalidCode):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid TOTP code"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to disable MFA"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"mfa_enabled": false}))
}

// Verify 登录步骤 2 — 验证 mfa_token + TOTP 码，发放正式 JWT
// POST /admin/api/v1/auth/mfa/verify  （无需 JWT 认证）
func (h *MFAHandler) Verify(c *gin.Context) {
	var req model.MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	ip := c.ClientIP()
	user, err := h.mfaService.VerifyMFALogin(c.Request.Context(), req.MFAToken, req.TOTPCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAPendingTokenInvalid):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "MFA token is invalid or expired"))
		case errors.Is(err, service.ErrMFAInvalidCode):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid TOTP code"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "MFA verification failed"))
		}
		return
	}

	// 通过 AuthService 生成正式 JWT
	loginResp, err := h.authService.IssueTokens(c.Request.Context(), user, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to issue tokens"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(loginResp))
}

// Recover 使用 Recovery Code 恢复登录
// POST /admin/api/v1/auth/mfa/recover  （无需 JWT 认证）
func (h *MFAHandler) Recover(c *gin.Context) {
	var req model.MFARecoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	ip := c.ClientIP()
	user, remaining, err := h.mfaService.Recover(c.Request.Context(), req.Email, req.RecoveryCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFANotEnabled):
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "MFA is not enabled for this account"))
		case errors.Is(err, service.ErrRecoveryCodeInvalid):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid or already used recovery code"))
		case errors.Is(err, service.ErrRecoveryCodesExhausted):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "all recovery codes have been used"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "recovery failed"))
		}
		return
	}

	// 通过 AuthService 生成正式 JWT
	loginResp, err := h.authService.IssueTokens(c.Request.Context(), user, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "failed to issue tokens"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.MFARecoverResponse{
		AccessToken:    loginResp.AccessToken,
		RefreshToken:   loginResp.RefreshToken,
		RemainingCodes: remaining,
	}))
}
