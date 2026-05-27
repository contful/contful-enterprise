// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// csrfTokens 内存存储 CSRF token（key: session_id, value: token）
var csrfTokens sync.Map

// GenerateCSRFToken 生成 32 字节随机 hex token
func GenerateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// SetupCSRF 安装向导 CSRF 防护中间件
// GET 请求：发放 token（Set-Cookie + X-CSRF-Token header）
// POST 请求：校验 token（Cookie session_id → 内存取出 token → 比对 Header X-CSRF-Token）
func SetupCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			token := GenerateCSRFToken()
			sessionID := GenerateCSRFToken()
			csrfTokens.Store(sessionID, token)

			c.SetCookie("setup_session", sessionID, 3600, "/", "", false, true)
			c.Header("X-CSRF-Token", token)
			c.Next()
			return
		}

		// POST 请求：校验
		sessionID, err := c.Cookie("setup_session")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing session"})
			return
		}

		expectedToken, ok := csrfTokens.Load(sessionID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid session"})
			return
		}

		headerToken := c.GetHeader("X-CSRF-Token")
		if headerToken != expectedToken.(string) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf token mismatch"})
			return
		}

		// 一次性使用：校验通过后立即销毁
		csrfTokens.Delete(sessionID)

		c.Next()
	}
}

// InvalidateCSRFToken 手动销毁 CSRF session
func InvalidateCSRFToken(c *gin.Context) {
	sessionID, err := c.Cookie("setup_session")
	if err == nil {
		csrfTokens.Delete(sessionID)
	}
	c.SetCookie("setup_session", "", -1, "/", "", false, true)
}
