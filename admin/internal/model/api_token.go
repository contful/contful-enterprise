package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TokenType Token 类型
type TokenType string

const (
	TokenTypeAPI     TokenType = "api"     // API Token (ctf_ 前缀)
	TokenTypeAccess  TokenType = "access"  // Access Token (不持久化)
	TokenTypeRefresh TokenType = "refresh" // Refresh Token
)

// TokenStatus Token 状态
type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "active"  // 激活
	TokenStatusExpired TokenStatus = "expired" // 过期
	TokenStatusRevoked TokenStatus = "revoked" // 已撤销
)

// StringArray PostgreSQL JSONB []string 类型
type StringArray []string

// Scan 实现 sql.Scanner
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// APIUsage 资源使用统计（保留扩展字段，当前 DB 用 request_count 列）
type APIUsage struct {
	RequestCount int `json:"request_count"`
}

// APIToken API Token
type APIToken struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID        uuid.UUID    `json:"site_id" gorm:"type:uuid;not null;index"`
	Name          string       `json:"name" gorm:"size:200;not null"`
	Description   string       `json:"description" gorm:"type:text"`
	TokenPrefix   string       `json:"token_prefix" gorm:"size:20;not null;index"` // ctg_ 前缀
	TokenHash     string       `json:"-" gorm:"size:64;not null;uniqueIndex"`
	EncryptedToken string       `json:"-" gorm:"type:text"`                           // AES-256-GCM 加密存储
	Scopes        StringArray  `json:"scopes" gorm:"type:jsonb;default:'[]'"`         // 权限范围
	SiteScope     StringArray  `json:"site_scope" gorm:"type:jsonb;default:'[]'"`     // 站点范围
	ChannelScope  StringArray  `json:"channel_scope" gorm:"type:jsonb;default:'[]'"`  // 频道范围
	AllowedIPs    *string      `json:"allowed_ips,omitempty" gorm:"type:inet"`
	RateLimit     int          `json:"rate_limit" gorm:"default:60"`
	ExpiresTime     *time.Time   `json:"expires_time,omitempty" gorm:"column:expires_time"`
	Status          TokenStatus  `json:"status" gorm:"type:token_status;not null;default:'active'"`
	LastUsedTime    *time.Time   `json:"last_used_time,omitempty" gorm:"column:last_used_time"`
	LastUsedIP      *string      `json:"last_used_ip,omitempty" gorm:"type:inet"`
	RequestCount    int64        `json:"request_count" gorm:"default:0"`
	CreatedBy       *uuid.UUID   `json:"created_by" gorm:"type:uuid"`
	CreatedTime     time.Time    `json:"created_time" gorm:"column:created_time;autoCreateTime"`
	UpdatedTime     time.Time    `json:"updated_time" gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime     *time.Time   `json:"deleted_time,omitempty" gorm:"type:timestamptz;index"`
}

// TableName 表名
func (APIToken) TableName() string {
	return "api_tokens"
}

// TokenHashPrefix 取 Token Hash 前缀用于日志脱敏
func TokenHashPrefix(fullHash string) string {
	if len(fullHash) < 8 {
		return strings.Repeat("*", len(fullHash))
	}
	return fullHash[:8] + "..."
}
