// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SchemaRepository 内容模型仓储
type SchemaRepository struct {
	db *gorm.DB
}

// NewSchemaRepository 新建仓储
func NewSchemaRepository(db *gorm.DB) *SchemaRepository {
	return &SchemaRepository{db: db}
}

// Create 创建内容模型
func (r *SchemaRepository) Create(ctx context.Context, ct *model.ContentSchema) error {
	return r.db.WithContext(ctx).Create(ct).Error
}

// GetByID 根据 ID 获取
func (r *SchemaRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ContentSchema, error) {
	var ct model.ContentSchema
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// GetByIDWithFields 获取内容模型及其字段
func (r *SchemaRepository) GetByIDWithFields(ctx context.Context, id uuid.UUID) (*model.ContentSchema, error) {
	var ct model.ContentSchema
	err := r.db.WithContext(ctx).
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("id = ?", id).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// GetBySlug 根据 slug 获取
func (r *SchemaRepository) GetBySlug(ctx context.Context, siteID uuid.UUID, slug string) (*model.ContentSchema, error) {
	var ct model.ContentSchema
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND slug = ?", siteID, slug).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// ListBySite 列出站点的内容模型
func (r *SchemaRepository) ListBySite(ctx context.Context, siteID uuid.UUID, page, pageSize int) ([]model.ContentSchema, int64, error) {
	var cts []model.ContentSchema
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ContentSchema{}).Where("site_id = ?", siteID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("sort_order ASC, created_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&cts).Error
	if err != nil {
		return nil, 0, err
	}

	return cts, total, nil
}

// Update 更新内容模型
func (r *SchemaRepository) Update(ctx context.Context, ct *model.ContentSchema) error {
	return r.db.WithContext(ctx).Save(ct).Error
}

// Delete 删除内容模型（软删除）
func (r *SchemaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ContentSchema{}, "id = ?", id).Error
}

// ExistsSlug 检查 slug 是否存在
func (r *SchemaRepository) ExistsSlug(ctx context.Context, siteID uuid.UUID, slug string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.ContentSchema{}).
		Where("site_id = ? AND slug = ?", siteID, slug)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
