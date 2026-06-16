// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful-enterprise/shared/uid"
	"github.com/contful/contful/admin/internal/model"
	"gorm.io/gorm"
)

type QueryTemplateRepository struct {
	db *gorm.DB
}

func NewQueryTemplateRepository(db *gorm.DB) *QueryTemplateRepository {
	return &QueryTemplateRepository{db: db}
}

// Create 创建查询模板
func (r *QueryTemplateRepository) Create(ctx context.Context, template *model.AuditQueryTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetByID 根据 ID 获取模板
func (r *QueryTemplateRepository) GetByID(ctx context.Context, id uid.UID) (*model.AuditQueryTemplate, error) {
	var template model.AuditQueryTemplate
	if err := r.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// ListByUser 查询用户的查询模板
func (r *QueryTemplateRepository) ListByUser(ctx context.Context, userID uid.UID) ([]model.AuditQueryTemplate, error) {
	var templates []model.AuditQueryTemplate
	err := r.db.WithContext(ctx).
		Where("created_by = ?", userID).
		Order("created_time DESC").
		Find(&templates).Error
	return templates, err
}

// Delete 删除模板（仅创建者可删除）
func (r *QueryTemplateRepository) Delete(ctx context.Context, id uid.UID, userID uid.UID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND created_by = ?", id, userID).
		Delete(&model.AuditQueryTemplate{}).Error
}
