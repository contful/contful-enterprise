package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

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

// JWTAuth JWT 认证中间件
func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "missing authorization header"))
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid authorization format"))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		
		// 解析 token
		parts := strings.Split(tokenString, ".")
		if len(parts) < 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid token"))
			return
		}

		claims, err := authService.ValidateAccessToken(parts[0])
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "token expired"))
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "invalid token"))
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
		claims, exists := c.Get(ClaimsContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		jwtClaims, ok := claims.(*service.JWTClaims)
		if !ok || !jwtClaims.IsSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "super admin only"))
			return
		}

		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 自定义日志中间件（使用 zerolog）
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现 zerolog 日志
		c.Next()
	}
}
