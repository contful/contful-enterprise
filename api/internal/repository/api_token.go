package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APITokenRepository Token 数据访问层
type APITokenRepository struct {
	db *gorm.DB
}

// NewAPITokenRepository 创建 Token Repository
func NewAPITokenRepository(db *gorm.DB) *APITokenRepository {
	return &APITokenRepository{db: db}
}

// FindByHash 根据 Token Hash 查找有效的 Token
func (r *APITokenRepository) FindByHash(ctx context.Context, tokenHash string) (*APIToken, error) {
	var token APIToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		Where("status = ?", TokenStatusActive).
		Where("deleted_time IS NULL").
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateLastUsedTime 更新最后使用时间
func (r *APITokenRepository) UpdateLastUsedTime(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&APIToken{}).
		Where("id = ?", tokenID).
		Update("last_used_time", &now).Error
}

// TokenStatusActive 活跃状态常量
const TokenStatusActive = "active"

// StringArray PostgreSQL JSONB []string 类型
type StringArray []string

// Scan 实现 sql.Scanner
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
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
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// APIToken 对应 DB 表结构（与 Admin API 共享同一 DB）
// 表结构见 contful/sql/init.sql api_tokens 表定义
type APIToken struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID       uuid.UUID   `gorm:"type:uuid;not null;index"`
	Name         string      `gorm:"size:200;not null"`
	Description  string      `gorm:"type:text"`
	TokenPrefix  string      `gorm:"size:20;not null;index"`
	TokenHash    string      `gorm:"size:64;not null;uniqueIndex"`
	Scopes       StringArray `gorm:"type:jsonb;default:'[]'"`        // 权限范围，如 ["read"] 或 ["read","write"]
	SiteScope    StringArray `gorm:"type:jsonb;default:'[]'"`        // 站点范围，["*"] 表示全部
	ChannelScope StringArray `gorm:"type:jsonb;default:'[]'"`        // 渠道范围
	RateLimit    int         `gorm:"column:rate_limit;default:60"`   // 速率限制（次/分钟）
	ExpiresTime  *time.Time  `gorm:"column:expires_time;type:timestamptz"`
	Status       string      `gorm:"type:token_status;not null;default:'active'"`
	LastUsedTime *time.Time  `gorm:"column:last_used_time;type:timestamptz"`
	CreatedTime  time.Time   `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime  time.Time   `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime  *time.Time  `gorm:"column:deleted_time;type:timestamptz"` // 与 Admin API 规范一致，手动软删除
}

func (APIToken) TableName() string {
	return "api_tokens"
}
