// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
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

func NewAPITokenRepository(db *gorm.DB) *APITokenRepository {
	return &APITokenRepository{db: db}
}

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

func (r *APITokenRepository) UpdateLastUsedTime(ctx context.Context, tokenID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&APIToken{}).
		Where("id = ?", tokenID).
		Update("last_used_time", &now).Error
}

const TokenStatusActive = "active"

// APIToken 对应 DB tokens 表
type APIToken struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Name         string     `gorm:"size:200;not null"`
	Description  string     `gorm:"type:text"`
	TokenPrefix  string     `gorm:"size:20;not null;index"`
	TokenHash    string     `gorm:"size:64;not null;uniqueIndex"`
	ExpiresTime  *time.Time `gorm:"column:expires_time;type:timestamptz"`
	Status       string     `gorm:"type:token_status;not null;default:'active'"`
	LastUsedTime *time.Time `gorm:"column:last_used_time;type:timestamptz"`
	CreatedTime  time.Time  `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime  time.Time  `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime  *time.Time `gorm:"column:deleted_time;type:timestamptz"`
}

func (APIToken) TableName() string {
	return "tokens"
}
