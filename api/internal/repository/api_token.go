package repository

import (
	"context"
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

// APIToken 对应 DB 表结构（与 Admin API 共享同一 DB）
type APIToken struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID       uuid.UUID       `gorm:"type:uuid;not null;index"`
	Name         string          `gorm:"size:100;not null"`
	TokenPrefix  string          `gorm:"size:10;not null;index"` // ctg_ 前 10 位
	TokenHash    string          `gorm:"size:64;not null;uniqueIndex"`
	Permissions  Permission      `gorm:"type:jsonb;default:'{}'"`
	RateLimits   Limits          `gorm:"type:jsonb;default:'{\"requests_per_minute\": 100}'"`
	ExpiresTime  *time.Time      `gorm:"column:expires_time;type:timestamptz"`
	Status       string          `gorm:"type:token_status;not null;default:'active'"`
	LastUsedTime *time.Time      `gorm:"column:last_used_time;type:timestamptz"`
	CreatedTime  time.Time       `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime  time.Time       `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime  *gorm.DeletedAt `gorm:"index"`
}

func (APIToken) TableName() string {
	return "api_tokens"
}

// Permission 权限结构（与 Admin API 共享）
type Permission struct {
	ContentTypes []string `json:"content_types"`
	AllowRead    bool     `json:"allow_read"`
	AllowWrite   bool     `json:"allow_write"`
}

// Limits 速率限制配置
type Limits struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerDay    int `json:"requests_per_day"`
}
