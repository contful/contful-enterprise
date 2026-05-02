// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentTypeRepository 内容类型数据访问层（只读）
type ContentTypeRepository struct {
	db *gorm.DB
}

// NewContentTypeRepository 创建 ContentType Repository
func NewContentTypeRepository(db *gorm.DB) *ContentTypeRepository {
	return &ContentTypeRepository{db: db}
}

// ContentType 内容类型（与 Admin API 共享同一 DB）
type ContentType struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	SiteID      uuid.UUID   `gorm:"type:uuid;not null;index"`
	Name        string      `gorm:"size:200;not null"`
	Slug        string      `gorm:"size:200;not null;index"`
	Description string      `gorm:"type:text"`
	Status      string      `gorm:"type:content_type_status;not null;default:'draft'"`
	CreatedTime time.Time   `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime time.Time   `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime *time.Time  `gorm:"column:deleted_time;type:timestamptz"`
}

func (ContentType) TableName() string {
	return "content_types"
}

// FindBySlug 通过站点 ID 和 slug 查找内容类型
func (r *ContentTypeRepository) FindBySlug(ctx context.Context, siteID uuid.UUID, slug string) (*ContentType, error) {
	var ct ContentType
	err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Where("slug = ?", slug).
		Where("deleted_time IS NULL").
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}
