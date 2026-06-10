// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"time"

	"github.com/contful/contful/openapi/pkg/uid"
	"gorm.io/gorm"
)

// ContentSchemaRepository 内容模型数据访问层（只读）
type ContentSchemaRepository struct {
	db *gorm.DB
}

// NewContentSchemaRepository 创建 ContentSchema Repository
func NewContentSchemaRepository(db *gorm.DB) *ContentSchemaRepository {
	return &ContentSchemaRepository{db: db}
}

// ContentSchema 内容模型（与 Admin API 共享同一 DB）
type ContentSchema struct {
	ID          uid.UID   `gorm:"type:uuid;primaryKey"`
	SiteID      uid.UID   `gorm:"type:uuid;not null;index"`
	Name        string      `gorm:"size:200;not null"`
	Slug        string      `gorm:"size:200;not null;index"`
	Description string      `gorm:"type:text"`
	Status      string      `gorm:"type:content_schema_status;not null;default:'draft'"`
	CreatedTime time.Time   `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime time.Time   `gorm:"column:updated_time;autoUpdateTime"`
	DeletedTime *time.Time  `gorm:"column:deleted_time;type:timestamptz"`
}

func (ContentSchema) TableName() string {
	return "schemas"
}

// FindBySlug 通过站点 ID 和 slug 查找内容模型
func (r *ContentSchemaRepository) FindBySlug(ctx context.Context, siteID uid.UID, slug string) (*ContentSchema, error) {
	var ct ContentSchema
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

// ListBySiteID 列出站点下所有 Content Schema
func (r *ContentSchemaRepository) ListBySiteID(ctx context.Context, siteID uid.UID) ([]*ContentSchema, error) {
	var schemas []*ContentSchema
	err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Where("deleted_time IS NULL").
		Order("created_time ASC").
		Find(&schemas).Error
	if err != nil {
		return nil, err
	}
	return schemas, nil
}
