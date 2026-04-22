package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 注册
// POST /admin/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	ip := c.ClientIP()
	user, err := h.authService.Register(c.Request.Context(), &req, ip)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, "email already exists"))
			return
		}
		// P1-004: 密码强度错误
		if errors.Is(err, service.ErrPasswordTooShort) ||
			errors.Is(err, service.ErrPasswordNoUppercase) ||
			errors.Is(err, service.ErrPasswordNoLowercase) ||
			errors.Is(err, service.ErrPasswordNoDigit) ||
			errors.Is(err, service.ErrPasswordNoSpecialChar) {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(user))
}

// Login 登录
// POST /admin/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request: "+err.Error()))
		return
	}

	ip := c.ClientIP()
	resp, err := h.authService.Login(c.Request.Context(), &req, ip)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidPassword):
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid email or password"))
		case errors.Is(err, repository.ErrUserInactive):
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "user is inactive"))
		case errors.Is(err, repository.ErrUserSuspended):
			c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "user is suspended"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
}

// Refresh 刷新 Token
// POST /admin/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	// 从 Authorization Header 获取 token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		// 也可以从 Cookie 获取
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "missing refresh token"))
			return
		}
		authHeader = "Bearer " + refreshToken
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	newAccessToken, newRefreshToken, err := h.authService.Refresh(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid refresh token"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(model.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}))
}

// Logout 登出
// POST /admin/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	ip := c.ClientIP()

	if err := h.authService.Logout(c.Request.Context(), token, ip); err != nil {
		// 登出错误不阻断流程
	}

	// 清除 Cookie
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// Me 获取当前用户
// GET /admin/v1/users/me
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		// 尝试从 uuid.UUID 获取
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userIDStr = uid.String()
		} else {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
			return
		}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid user id"))
		return
	}

	user, err := h.authService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))
}

// GetClaims 实现 middleware.claimsGetter 接口，供 JWT 中间件使用
func (h *AuthHandler) GetClaims(token string) (*middleware.Claims, error) {
	// JWT 格式: header.payload.signature，直接解析即可
	claims, err := h.authService.ParseAccessTokenInternal(token)
	if err != nil {
		return nil, err
	}

	return &middleware.Claims{
		UserID:       claims.UserID,
		Email:        claims.Email,
		IsSuperAdmin: claims.IsSuperAdmin,
	}, nil
}

// ListUsers 获取用户列表
// GET /admin/v1/users
func (h *AuthHandler) ListUsers(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid request"))
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	page, err := h.authService.ListUsers(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(page))
}
