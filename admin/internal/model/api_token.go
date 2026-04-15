package model

import (
	"time"

	"github.com/google/uuid"
)

// TokenType Token 类型
type TokenType string

const (
	TokenTypeAPI    TokenType = "api"    // API Token (ctf_ 前缀)
	TokenTypeAccess TokenType = "access" // Access Token (不持久化)
	TokenTypeRefresh TokenType = "refresh" // Refresh Token
)

// TokenStatus Token 状态
type TokenStatus string

const (
	TokenStatusActive   TokenStatus = "active"   // 激活
	TokenStatusExpired  TokenStatus = "expired"  // 过期
	TokenStatusRevoked  TokenStatus = "revoked"  // 已撤销
)

// APIEndpoint API 端点权限
type APIEndpoint struct {
	Path   string   `json:"path"`   // 路径模式 (如 /content/*)
	Method []string `json:"method"` // HTTP 方法 (GET, POST, PUT, DELETE)
}

// APIEndpointPermission 端点权限
type EndpointPermission struct {
	ContentTypes []string `json:"content_types,omitempty"` // 允许的内容类型 (* 表示全部)
	Endpoints    []APIEndpoint `json:"endpoints,omitempty"` // 允许的端点
}

// APIEndpointLimits 速率限制
type APIEndpointLimits struct {
	RequestsPerMinute int `json:"requests_per_minute"` // 每分钟请求数
	RequestsPerDay    int `json:"requests_per_day"`    // 每天请求数
}

// APIUsage 资源使用统计
type APIUsage struct {
	RequestCount    int       `json:"request_count"`    // 总请求数
	DailyRequestCount int     `json:"daily_request_count"` // 今日请求数
	BandwidthUsed    int64    `json:"bandwidth_used"`     // 带宽使用 (bytes)
	LastRequestAt   *time.Time `json:"last_request_at,omitempty"` // 最后请求时间
}

// APIToken API Token
type APIToken struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID          uuid.UUID              `json:"site_id" gorm:"type:uuid;not null;index"`
	Name            string                 `json:"name" gorm:"size:100;not null"`
	Description     string                 `json:"description" gorm:"size:500"`
	TokenPrefix     string                 `json:"token_prefix" gorm:"size:10;not null;index"` // ctg_ 前 10 位
	TokenHash       string                 `json:"-" gorm:"size:64;not null;uniqueIndex"`     // SHA-256 哈希
	Permissions     EndpointPermission     `json:"permissions" gorm:"type:jsonb;default:'{}'"`
	RateLimits      APIEndpointLimits      `json:"rate_limits" gorm:"type:jsonb;default:'{\"requests_per_minute\": 60, \"requests_per_day\": 10000}'"`
	Usage           APIUsage               `json:"usage" gorm:"type:jsonb;default:'{\"request_count\": 0}'"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty" gorm:"type:timestamptz"`
	Status          TokenStatus            `json:"status" gorm:"type:token_status;not null;default:'active'"`
	LastUsedAt      *time.Time             `json:"last_used_at,omitempty" gorm:"type:timestamptz"`
	CreatedBy       *uuid.UUID             `json:"created_by" gorm:"type:uuid"`
	CreatedAt       time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       *time.Time             `json:"deleted_at" gorm:"index"`
}

// TableName 表名
func (APIToken) TableName() string {
	return "api_tokens"
}
