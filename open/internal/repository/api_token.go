package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/contful/contful/open/internal/model"
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
func (r *APITokenRepository) FindByHash(ctx context.Context, tokenHash string) (*model.APIToken, error) {
	var token model.APIToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		Where("status = ?", model.TokenStatusActive).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateLastUsedAt 更新最后使用时间
func (r *APITokenRepository) UpdateLastUsedAt(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("id = ?", tokenID).
		Update("last_used_at", &now).Error
}

// APIToken 对应 DB 表结构（与 Admin API 共享同一 DB）
type APIToken struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID      uuid.UUID       `gorm:"type:uuid;not null;index"`
	Name        string          `gorm:"size:100;not null"`
	TokenPrefix string          `gorm:"size:10;not null;index"` // ctg_ 前 10 位
	TokenHash   string          `gorm:"size:64;not null;uniqueIndex"`
	Permissions model.Permission `gorm:"type:jsonb;default:'{}'"`
	RateLimits  model.Limits     `gorm:"type:jsonb;default:'{\"requests_per_minute\": 100}'"`
	ExpiresAt   *time.Time      `gorm:"type:timestamptz"`
	Status      string          `gorm:"type:token_status;not null;default:'active'"`
	LastUsedAt  *time.Time      `gorm:"type:timestamptz"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"`
	DeletedAt   *gorm.DeletedAt `gorm:"index"`
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
