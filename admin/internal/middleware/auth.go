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
)

// Claims JWT Claims
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
				"code":    "UNAUTHORIZED",
				"message": "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid authorization format",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		claims, err := getter.GetClaims(tokenString)
		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    "UNAUTHORIZED",
					"message": "token expired",
				})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid token",
			})
			return
		}

		// 将用户信息存入上下文
		c.Set(ClaimsContextKey, claims)
		c.Set(UserContextKey, claims.UserID)

		c.Next()
	}
}

// SuperAdminOnly 超级管理员权限中间件
func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "unauthorized",
			})
			return
		}

		claims, ok := claimsVal.(*Claims)
		if !ok || !claims.IsSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "super admin only",
			})
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 自定义日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
