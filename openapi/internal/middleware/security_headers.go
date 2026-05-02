// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware HTTP 安全响应头
// 纵深防御：即使 Open API 本身无 XSS 风险，也要返回正确的安全头
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "accelerometer=(), camera=(), microphone=(), geolocation=()")
		// Content-Security-Policy：只允许同源，API 响应为 JSON，不需要 script
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}
