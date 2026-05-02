// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"strings"
	"time"

	"github.com/contful/contful/admin/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// P1-002 修复：转义 LIKE 查询中的通配符字符，防止意外匹配
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// APITokenRepository API Token 仓储
type APITokenRepository struct {
	db *gorm.DB
}

// NewAPITokenRepository 新建仓储
func NewAPITokenRepository(db *gorm.DB) *APITokenRepository {
	return &APITokenRepository{db: db}
}

// Create 创建 Token
func (r *APITokenRepository) Create(ctx context.Context, token *model.APIToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByID 根据 ID 获取
func (r *APITokenRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.APIToken, error) {
	var token model.APIToken
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByTokenHash 根据 Token Hash 获取
func (r *APITokenRepository) GetByTokenHash(ctx context.Context, hash string) (*model.APIToken, error) {
	var token model.APIToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		Where("status = ?", model.TokenStatusActive).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByPrefix 根据 Token 前缀获取
func (r *APITokenRepository) GetByPrefix(ctx context.Context, prefix string) (*model.APIToken, error) {
	var token model.APIToken
	err := r.db.WithContext(ctx).
		Where("token_prefix = ?", prefix).
		Where("status = ?", model.TokenStatusActive).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// List 列出站点的 Token
func (r *APITokenRepository) List(ctx context.Context, siteID uuid.UUID, filter *model.APITokenListFilter, page, pageSize int) ([]model.APIToken, int64, error) {
	var tokens []model.APIToken
	var total int64

	query := r.db.WithContext(ctx).Model(&model.APIToken{}).Where("site_id = ?", siteID)

	// 应用过滤条件
	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Name != nil && *filter.Name != "" {
			escaped := escapeLikePattern(*filter.Name)
			query = query.Where("name ILIKE ?", "%"+escaped+"%")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("created_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tokens).Error
	if err != nil {
		return nil, 0, err
	}

	return tokens, total, nil
}

// Update 更新 Token
func (r *APITokenRepository) Update(ctx context.Context, token *model.APIToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

// Delete 删除 Token（软删除）
func (r *APITokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.APIToken{}, "id = ?", id).Error
}

// Revoke 撤销 Token
func (r *APITokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("id = ?", id).
		Update("status", model.TokenStatusRevoked).Error
}

// UpdateUsage 更新使用统计（自增 request_count）
func (r *APITokenRepository) UpdateUsage(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("id = ?", id).
		Update("request_count", gorm.Expr("request_count + 1")).Error
}

// UpdateLastUsed 更新最后使用时间
func (r *APITokenRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.APIToken{}).
		Where("id = ?", id).
		Update("last_used_time", &now).Error
}

// CountBySite 统计站点的 Token 数量
func (r *APITokenRepository) CountBySite(ctx context.Context, siteID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.APIToken{}).
		Where("site_id = ?", siteID).
		Where("status = ?", model.TokenStatusActive).
		Count(&count).Error
	return count, err
}
