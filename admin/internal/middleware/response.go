package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/model"
)

// OK 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, model.NewSuccessResponse(data))
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, model.NewSuccessResponse(data))
}

// BadRequest 错误请求响应
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, msg))
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, msg))
}

// Forbidden 禁止响应
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, msg))
}

// NotFound 未找到响应
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, msg))
}

// Conflict 冲突响应
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, msg))
}

// InternalError 内部错误响应
func InternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, msg))
}

// GetSiteID 获取站点 ID
// 优先从 X-Site-ID 请求头读取，否则返回 nil
func GetSiteID(c *gin.Context) (uuid.UUID, bool) {
	siteHeader := c.GetHeader("X-Site-ID")
	if siteHeader != "" {
		siteID, err := uuid.Parse(siteHeader)
		if err == nil {
			return siteID, true
		}
	}
	return uuid.Nil, false
}

// GetUserID 获取用户 ID
func GetUserID(c *gin.Context) (*uuid.UUID, bool) {
	userID, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		return nil, false
	}
	return &uid, true
}

// GetClaims 获取 JWT Claims
func GetClaims(c *gin.Context) (*Claims, bool) {
	claims, exists := c.Get(ClaimsContextKey)
	if !exists {
		return nil, false
	}
	return claims.(*Claims), true
}
