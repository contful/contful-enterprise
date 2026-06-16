// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/contful/contful-enterprise/shared/uid"
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
	// 有关键词时走全文搜索
	if filter != nil && filter.Keyword != "" {
		return r.SearchFullText(ctx, filter, page, pageSize)
	}

	var logs []model.AuditLog
	var total int64

	query := r.buildFilterQuery(ctx, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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

	query := r.buildFilterQueryWithKeyword(ctx, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(maxRows).Order("created_time DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetRecentLogs 查询最近一段时间内的审计日志（异常扫描用）
func (r *AuditRepository) GetRecentLogs(ctx context.Context, since time.Duration) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := r.db.WithContext(ctx).
		Model(&model.AuditLog{}).
		Where("created_time >= ?", time.Now().Add(-since)).
		Order("created_time DESC").
		Find(&logs).Error
	return logs, err
}

// SearchFullText 全文搜索（GIN tsvector）
func (r *AuditRepository) SearchFullText(ctx context.Context, filter *model.AuditLogFilter, page, pageSize int) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.buildFilterQueryWithKeyword(ctx, filter)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_time DESC").Find(&logs).Error; err != nil {
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

// buildFilterQueryWithKeyword 构建带全文搜索的查询
func (r *AuditRepository) buildFilterQueryWithKeyword(ctx context.Context, filter *model.AuditLogFilter) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})
	hasKeyword := filter != nil && filter.Keyword != ""

	if filter == nil {
		return query
	}

	// 收集所有条件表达式
	var conds []string
	var args []interface{}

	if filter.SiteID != nil {
		conds = append(conds, "site_id = ?")
		args = append(args, *filter.SiteID)
	}
	if filter.UserID != nil {
		conds = append(conds, "user_id = ?")
		args = append(args, *filter.UserID)
	}
	if filter.Action != "" {
		conds = append(conds, "action = ?")
		args = append(args, filter.Action)
	}
	if filter.ResourceType != "" {
		conds = append(conds, "resource_type = ?")
		args = append(args, filter.ResourceType)
	}
	if filter.Category != "" {
		conds = append(conds, "category = ?")
		args = append(args, filter.Category)
	}
	if filter.Level != "" {
		conds = append(conds, "level = ?")
		args = append(args, filter.Level)
	}
	if !filter.StartTime.IsZero() {
		conds = append(conds, "created_time >= ?")
		args = append(args, filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		conds = append(conds, "created_time <= ?")
		args = append(args, filter.EndTime)
	}

	// 全文搜索条件
	if hasKeyword {
		// 使用 plainto_tsquery 处理用户输入（转义单引号防注入）
		keyword := strings.ReplaceAll(filter.Keyword, "'", "''")
		keywordCond := fmt.Sprintf("search_vector @@ plainto_tsquery('simple', '%s')", keyword)
		conds = append(conds, keywordCond)
	}

	if len(conds) == 0 {
		return query
	}

	// 组合条件逻辑
	combinator := "AND"
	if filter.Combinator == "or" {
		combinator = "OR"
	}

	whereClause := strings.Join(conds, " "+combinator+" ")
	query = query.Where(whereClause, args...)

	return query
}
