package model

import (
	"github.com/google/uuid"
)

// TokenContext Token 验证通过后存入 Context 的信息
type TokenContext struct {
	TokenID      uuid.UUID       `json:"token_id"`
	SiteID       uuid.UUID       `json:"site_id"`
	Name         string          `json:"name"`
	Permissions  TokenPermission `json:"permissions"`
	RateLimits   RateLimitConfig `json:"rate_limits"`
	ExpiresAt    *int64          `json:"expires_at,omitempty"` // Unix 时间戳
}

// TokenPermission Token 权限范围
type TokenPermission struct {
	ContentTypes []string `json:"content_types"` // 允许的内容类型 slug，* 表示全部
	AllowRead    bool     `json:"allow_read"`    // 允许读取
	AllowWrite   bool     `json:"allow_write"`   // 允许写入
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerDay     int `json:"requests_per_day"`
}

// DefaultRateLimit 默认速率限制（Open API 标准 Token）
var DefaultRateLimit = RateLimitConfig{
	RequestsPerMinute: 100,
	RequestsPerDay:    10000,
}

// Context keys
const (
	TokenContextKey = "api_token"
)
