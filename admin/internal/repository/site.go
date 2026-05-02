// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"strings"

	"github.com/contful/contful/admin/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SiteRepository 站点仓储
type SiteRepository struct {
	db *gorm.DB
}

// NewSiteRepository 新建仓储
func NewSiteRepository(db *gorm.DB) *SiteRepository {
	return &SiteRepository{db: db}
}

// Create 创建站点
func (r *SiteRepository) Create(ctx context.Context, site *model.Site) error {
	return r.db.WithContext(ctx).Create(site).Error
}

// GetByID 根据 ID 获取
func (r *SiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Site, error) {
	var site model.Site
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_time IS NULL", id).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

// GetBySlug 根据 slug 获取
func (r *SiteRepository) GetBySlug(ctx context.Context, slug string) (*model.Site, error) {
	var site model.Site
	err := r.db.WithContext(ctx).Where("slug = ? AND deleted_time IS NULL", slug).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

// List 列出站点
func (r *SiteRepository) List(ctx context.Context, page, pageSize int, isActive *bool) ([]model.Site, int64, error) {
	var sites []model.Site
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Site{}).Where("deleted_time IS NULL")

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("created_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sites).Error
	if err != nil {
		return nil, 0, err
	}

	return sites, total, nil
}

// Update 更新站点
func (r *SiteRepository) Update(ctx context.Context, site *model.Site) error {
	return r.db.WithContext(ctx).Save(site).Error
}

// Delete 软删除站点
func (r *SiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Site{}).Error
}

// SlugExists 检查 slug 是否已存在
func (r *SiteRepository) SlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Site{}).Where("slug = ? AND deleted_time IS NULL", slug)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// escapeLikePattern 转义 LIKE 通配符
func escapeSiteLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
