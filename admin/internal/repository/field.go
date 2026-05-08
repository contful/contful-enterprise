// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FieldRepository 字段仓储
type FieldRepository struct {
	db *gorm.DB
}

// NewFieldRepository 新建仓储
func NewFieldRepository(db *gorm.DB) *FieldRepository {
	return &FieldRepository{db: db}
}

// Create 创建字段
func (r *FieldRepository) Create(ctx context.Context, field *model.Field) error {
	return r.db.WithContext(ctx).Create(field).Error
}

// CreateBatch 批量创建字段
func (r *FieldRepository) CreateBatch(ctx context.Context, fields []*model.Field) error {
	return r.db.WithContext(ctx).CreateInBatches(fields, 100).Error
}

// GetByID 根据 ID 获取
func (r *FieldRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Field, error) {
	var field model.Field
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&field).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

// ListByContentSchema 列出内容模型的字段
func (r *FieldRepository) ListByContentSchema(ctx context.Context, contentSchemaID uuid.UUID) ([]model.Field, error) {
	var fields []model.Field
	err := r.db.WithContext(ctx).
		Where("schema_id = ?", contentSchemaID).
		Order("sort_order ASC").
		Find(&fields).Error
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// Update 更新字段
func (r *FieldRepository) Update(ctx context.Context, field *model.Field) error {
	return r.db.WithContext(ctx).Save(field).Error
}

// Delete 删除字段（软删除）
func (r *FieldRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Field{}, "id = ?", id).Error
}

// DeleteByContentSchema 删除内容模型的所有字段
func (r *FieldRepository) DeleteByContentSchema(ctx context.Context, contentSchemaID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Field{}, "schema_id = ?", contentSchemaID).Error
}

// ExistsName 检查字段名是否已存在
func (r *FieldRepository) ExistsName(ctx context.Context, contentSchemaID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Field{}).
		Where("schema_id = ? AND name = ?", contentSchemaID, name)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMaxSortOrder 获取最大排序号
func (r *FieldRepository) GetMaxSortOrder(ctx context.Context, contentSchemaID uuid.UUID) (int, error) {
	var maxOrder int
	err := r.db.WithContext(ctx).
		Model(&model.Field{}).
		Where("schema_id = ?", contentSchemaID).
		Select("COALESCE(MAX(sort_order), -1)").
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	return maxOrder, nil
}

// ReorderFields 重新排序字段
func (r *FieldRepository) ReorderFields(ctx context.Context, fieldOrders map[uuid.UUID]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for id, order := range fieldOrders {
			if err := tx.Model(&model.Field{}).Where("id = ?", id).Update("sort_order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
