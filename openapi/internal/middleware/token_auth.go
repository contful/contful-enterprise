package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/contful/contful/openapi/internal/model"
	"github.com/contful/contful/openapi/internal/service"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix       = "Bearer "
)

// TokenAuthMiddleware API Token 认证中间件
// 验证流程：
//   1. 从 Header 提取 Bearer Token
//   2. 校验格式（ctf_ 前缀 + 长度）
//   3. SHA-256 Hash 后查询数据库
//   4. 验证状态（未撤销、未过期）
//   5. 将 TokenContext 存入 Context
func TokenAuthMiddleware(tokenService *service.APITokenService, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "missing authorization header"))
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeInvalidToken, "invalid authorization format, expected: Bearer ctf_xxxxx"))
			return
		}

		rawToken := strings.TrimPrefix(authHeader, BearerPrefix)
		rawToken = strings.TrimSpace(rawToken)

		ctx := c.Request.Context()
		tokenCtx, err := tokenService.ValidateToken(ctx, rawToken)
		if err != nil {
			switch err {
			case service.ErrTokenNotFound, service.ErrInvalidTokenFormat:
				logger.Debug().Err(err).Str("token_prefix", rawToken[:minLen(rawToken)]).Msg("invalid token")
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeInvalidToken, "invalid token"))
			case service.ErrTokenExpired:
				logger.Debug().Err(err).Msg("token expired")
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeTokenExpired, "token expired"))
			case service.ErrTokenRevoked:
				logger.Debug().Err(err).Msg("token revoked")
				c.AbortWithStatusJSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeTokenRevoked, "token revoked"))
			default:
				logger.Error().Err(err).Msg("token validation error")
				c.AbortWithStatusJSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
			}
			return
		}

		// 存入 Context，供后续 Handler 和 Scope 检查使用
		c.Set(model.TokenContextKey, tokenCtx)

		// 将 SiteID 注入请求上下文（给 GORM 使用）
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "site_id", tokenCtx.SiteID))

		logger.Debug().
			Str("token_id", tokenCtx.TokenID.String()).
			Str("site_id", tokenCtx.SiteID.String()).
			Msg("token authenticated")

		c.Next()
	}
}

// RequireRead 只允许读操作（GET/HEAD）
func RequireRead() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "GET" && method != "HEAD" && method != "OPTIONS" {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorResponse(model.CodeInsufficientScope, "this endpoint requires read permission"))
			return
		}
		c.Next()
	}
}

// RequireWrite 只允许写操作（POST/PUT/PATCH/DELETE）
func RequireWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "GET" || method == "HEAD" {
			c.AbortWithStatusJSON(http.StatusForbidden, model.NewErrorResponse(model.CodeInsufficientScope, "this endpoint requires write permission"))
			return
		}
		c.Next()
	}
}

// minLen 安全截取 token 前缀用于日志（不打印完整 token）
func minLen(token string) int {
	if len(token) > 10 {
		return 10
	}
	return len(token)
}

// GetTokenContext 从 Gin Context 中取出 TokenContext，若不存在返回 nil
func GetTokenContext(c *gin.Context) *model.TokenContext {
	val, exists := c.Get(model.TokenContextKey)
	if !exists {
		return nil
	}
	tc, ok := val.(*model.TokenContext)
	if !ok {
		return nil
	}
	return tc
}
