// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

// 共享常量
const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserContextKey      = "user"
	ClaimsContextKey    = "claims"
	SiteIDHeader        = "X-Site-ID"
	SiteContextKey      = "site_id"
)

// Claims JWT Claims（不含 site_id，由前端通过请求头传递）
type Claims struct {
	UserID       uuid.UUID
	Email        string
	IsSuperAdmin bool
}

// claimsGetter 接口：允许 handler 层注入 JWT 验证逻辑，避免循环依赖
type claimsGetter interface {
	GetClaims(string) (*Claims, error)
}

// JWTAuth JWT 认证中间件
func JWTAuth(getter claimsGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "invalid authorization format",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		claims, err := getter.GetClaims(tokenString)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  "token expired",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "invalid token",
			})
			return
		}

		// 将用户信息存入上下文
		c.Set(ClaimsContextKey, claims)
		c.Set(UserContextKey, claims.UserID)

		// 从请求头获取 site_id（前端必须传递）
		siteIDStr := c.GetHeader(SiteIDHeader)
		if siteIDStr != "" {
			if siteID, err := uuid.Parse(siteIDStr); err == nil {
				c.Set(SiteContextKey, siteID)
			}
		}

		c.Next()
	}
}

// SuperAdminOnly 超级管理员权限中间件
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "unauthorized",
			})
			return
		}

		claims, ok := claimsVal.(*Claims)
		if !ok || !claims.IsSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "super admin only",
			})
			return
		}

		c.Next()
	}
}

// GetSiteID 从上下文获取当前站点 ID（需要前端通过 X-Site-ID 头传递）
func GetSiteID(c *gin.Context) uuid.UUID {
	if siteID, exists := c.Get(SiteContextKey); exists {
		if id, ok := siteID.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

// MustSiteID 获取站点 ID，如果不存在则返回错误
func MustSiteID(c *gin.Context) (uuid.UUID, error) {
	siteID := GetSiteID(c)
	if siteID == uuid.Nil {
		return uuid.Nil, ErrSiteIDRequired
	}
	return siteID, nil
}

// ErrSiteIDRequired site_id 缺失错误
var ErrSiteIDRequired = &BizError{Code: "SITE_ID_REQUIRED", Message: "X-Site-ID header is required"}

// BizError 业务错误
type BizError struct {
	Code    string
	Message string
}

func (e *BizError) Error() string {
	return e.Message
}

// LoggerMiddleware 自定义日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
