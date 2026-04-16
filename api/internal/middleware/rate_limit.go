package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/contful/contful/api/internal/model"
)

// RateLimiter Redis 滑动窗口速率限制
// 100次/分钟/Token
type RateLimiter struct {
	rdb *redis.Client
}

// NewRateLimiter 创建 RateLimiter
func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// RateLimitByToken 基于 Token 的速率限制
// 使用 Redis 滑动窗口算法，窗口大小 60 秒
func (rl *RateLimiter) RateLimitByToken(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 从 Context 获取 Token 信息（由 TokenAuthMiddleware 提供）
		tokenVal, exists := c.Get(model.TokenContextKey)
		if !exists {
			// 未经过认证，跳过限速
			c.Next()
			return
		}
		tokenCtx := tokenVal.(*model.TokenContext)

		// 使用 Token ID 作为限速 key
		key := fmt.Sprintf("ratelimit:openapi:%s", tokenCtx.TokenID.String())
		now := time.Now().Unix()
		windowStart := now - 60 // 60 秒滑动窗口

		pipe := rl.rdb.Pipeline()

		// 删除窗口外的记录
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// 当前窗口内的请求数
		countCmd := pipe.ZCard(ctx, key)

		_, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			// Redis 出错时记录日志，但不阻止请求（fail open）
			c.Next()
			return
		}

		count := countCmd.Val()
		if int(count) >= limit {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now+60))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, model.NewErrorResponse(
				model.CodeRateLimitExceeded,
				fmt.Sprintf("rate limit exceeded: %d requests per minute", limit),
			))
			return
		}

		// 记录本次请求
		rl.rdb.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d-%d", now, time.Now().UnixNano()),
		})
		rl.rdb.Expire(ctx, key, 2*time.Minute) // 多留 1 分钟防止并发问题

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-int(count)-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now+60))

		c.Next()
	}
}

// RateLimitByIP 基于 IP 的速率限制（用于未认证接口，如 health check）
func (rl *RateLimiter) RateLimitByIP(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:openapi:ip:%s", ip)
		now := time.Now().Unix()
		windowStart := now - 60

		pipe := rl.rdb.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
		countCmd := pipe.ZCard(ctx, key)
		_, _ = pipe.Exec(ctx)

		count := countCmd.Val()
		if int(count) >= limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, model.NewErrorResponse(
				model.CodeRateLimitExceeded,
				fmt.Sprintf("rate limit exceeded: %d requests per minute", limit),
			))
			return
		}

		rl.rdb.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", time.Now().UnixNano())})
		rl.rdb.Expire(ctx, key, 2*time.Minute)

		c.Next()
	}
}

// DailyLimit 每日调用次数限制
func (rl *RateLimiter) DailyLimit(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		tokenVal, exists := c.Get(model.TokenContextKey)
		if !exists {
			c.Next()
			return
		}
		tokenCtx := tokenVal.(*model.TokenContext)

		key := fmt.Sprintf("ratelimit:openapi:daily:%s:%s",
			tokenCtx.TokenID.String(),
			time.Now().Format("20060102"),
		)

		count, err := rl.rdb.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		if count >= limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, model.NewErrorResponse(
				model.CodeRateLimitExceeded,
				fmt.Sprintf("daily limit exceeded: %d requests per day", limit),
			))
			return
		}

		rl.rdb.Incr(ctx, key)
		rl.rdb.Expire(ctx, key, 25*time.Hour)

		c.Next()
	}
}

// contextWithSiteID 将 site_id 注入 context（供 repository 层使用）
func contextWithSiteID(ctx context.Context, c *gin.Context) context.Context {
	if tokenVal, exists := c.Get(model.TokenContextKey); exists {
		if tc, ok := tokenVal.(*model.TokenContext); ok {
			return context.WithValue(ctx, "site_id", tc.SiteID)
		}
	}
	return ctx
}
