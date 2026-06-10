// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/openapi/pkg/uid"
	"gorm.io/gorm"
)

// FieldRepository 字段定义数据访问层（只读）
type FieldRepository struct {
	db *gorm.DB
}

// NewFieldRepository 创建 FieldRepository
func NewFieldRepository(db *gorm.DB) *FieldRepository {
	return &FieldRepository{db: db}
}

// ListByContentSchemaID 列出某个 Content Schema 的所有 Fields
func (r *FieldRepository) ListByContentSchemaID(ctx context.Context, schemaID uid.UID) ([]Field, error) {
	var fields []Field
	err := r.db.WithContext(ctx).
		Where("schema_id = ?", schemaID).
		Where("deleted_time IS NULL").
		Order("sort_order ASC, created_time ASC").
		Find(&fields).Error
	if err != nil {
		return nil, err
	}
	return fields, nil
}
