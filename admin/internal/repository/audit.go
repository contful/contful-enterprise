// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/admin/pkg/uid"
	"github.com/contful/contful/admin/internal/audit"
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
	ctx = audit.WithSigningKey(ctx, signingKey)
	return r.db.WithContext(ctx).Create(log).Error
}

// ListByUser 查询用户的审计日志
func (r *AuditRepository) ListByUser(ctx context.Context, userID uid.UID, page, pageSize int) ([]model.AuditLog, int64, error) {
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
func (r *AuditRepository) GetByID(ctx context.Context, id uid.UID) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// List 通用列表查询（支持筛选和分页）
func (r *AuditRepository) List(ctx context.Context, filter *model.AuditLogFilter, page, pageSize int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.buildFilterQuery(ctx, filter)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_time DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ExportAll 查询满足筛选条件的所有记录（导出用，支持 maxRows 限制）
func (r *AuditRepository) ExportAll(ctx context.Context, filter *model.AuditLogFilter, maxRows int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.buildFilterQuery(ctx, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(maxRows).Order("created_time DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditRepository) buildFilterQuery(ctx context.Context, filter *model.AuditLogFilter) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	if filter != nil {
		if filter.SiteID != nil {
			query = query.Where("site_id = ?", *filter.SiteID)
		}
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Action != "" {
			query = query.Where("action = ?", filter.Action)
		}
		if filter.ResourceType != "" {
			query = query.Where("resource_type = ?", filter.ResourceType)
		}
		if filter.Category != "" {
			query = query.Where("category = ?", filter.Category)
		}
		if filter.Level != "" {
			query = query.Where("level = ?", filter.Level)
		}
		if !filter.StartTime.IsZero() {
			query = query.Where("created_time >= ?", filter.StartTime)
		}
		if !filter.EndTime.IsZero() {
			query = query.Where("created_time <= ?", filter.EndTime)
		}
	}

	return query
}
