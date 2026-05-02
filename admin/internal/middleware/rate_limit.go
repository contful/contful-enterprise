// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter Admin API 限流中间件（基于 IP）
type RateLimiter struct {
	rdb *redis.Client
}

// NewRateLimiter 创建限流中间件
func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// LoginRateLimit 登录接口限流：5次/分钟/IP
func (rl *RateLimiter) LoginRateLimit() gin.HandlerFunc {
	return rl.rateLimitByIP(5, time.Minute, "login")
}

// rateLimitByIP 基于 IP 的滑动窗口限流
func (rl *RateLimiter) rateLimitByIP(maxRequests int, window time.Duration, limitType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端 IP
		clientIP := getClientIP(c)
		key := "ratelimit:admin:" + limitType + ":" + clientIP

		ctx := c.Request.Context()
		now := time.Now().UnixNano()
		windowStart := now - window.Nanoseconds()

		// 使用 Redis ZSET 实现滑动窗口
		pipe := rl.rdb.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: strconv.FormatInt(now, 10)})
		pipe.ZCard(ctx, key)
		pipe.Expire(ctx, key, window)

		results, err := pipe.Exec(ctx)
		if err != nil {
			// Redis 故障时允许请求通过（fail open）
			c.Next()
			return
		}

		// 获取当前窗口内的请求数
		count := results[2].(*redis.IntCmd).Val()
		if count > int64(maxRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP 获取客户端真实 IP
func getClientIP(c *gin.Context) string {
	// 优先使用 X-Forwarded-For
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For: client, proxy1, proxy2
		parts := splitComma(xff)
		if len(parts) > 0 {
			return trimSpace(parts[0])
		}
	}
	// 其次使用 X-Real-IP
	if xrip := c.GetHeader("X-Real-IP"); xrip != "" {
		return xrip
	}
	// 最后使用直接连接的 IP
	return c.ClientIP()
}

// splitComma 分割逗号分隔的字符串
func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if start < i {
				parts = append(parts, s[start:i])
			}
			start = i + 1
		}
	}
	return parts
}

// trimSpace 去除前后空格
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
