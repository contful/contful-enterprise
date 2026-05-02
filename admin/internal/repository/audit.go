// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/audit_callback"
	"github.com/contful/contful/admin/internal/model"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create 创建审计日志（自动将签名密钥注入 context，供 GORM callback 使用）
func (r *AuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// CreateWithSigningKey 创建审计日志并显式指定签名密钥（绕过自动查找）
func (r *AuditRepository) CreateWithSigningKey(ctx context.Context, log *model.AuditLog, signingKey string) error {
	ctx = audit_callback.WithSigningKey(ctx, signingKey)
	return r.db.WithContext(ctx).Create(log).Error
}

// ListByUser 查询用户的审计日志
func (r *AuditRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	db := r.db.WithContext(ctx).Model(&model.AuditLog{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_time DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID 根据 ID 获取审计日志
func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
