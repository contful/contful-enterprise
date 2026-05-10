// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrSiteNotFound    = errors.New("site not found")
	ErrSiteSlugExists  = errors.New("site slug already exists")
	ErrSiteInvalidSlug = errors.New("invalid site slug")
)

// siteSlugRegex 站点 slug 规则：字母开头，小写字母+数字+连字符，2-100字符
var siteSlugRegex = regexp.MustCompile(`^[a-z][a-z0-9\-]{0,98}[a-z0-9]$`)

// SiteService 站点服务
type SiteService struct {
	db       *gorm.DB
	siteRepo *repository.SiteRepository
}

// NewSiteService 新建站点服务
func NewSiteService(db *gorm.DB, siteRepo *repository.SiteRepository) *SiteService {
	return &SiteService{db: db, siteRepo: siteRepo}
}

// Create 创建站点（不再创建站点级配置，由前端通过 settings JSONB 管理）
func (s *SiteService) Create(ctx context.Context, userID uuid.UUID, req *model.SiteCreate) (*model.Site, error) {
	// 验证 slug
	if !siteSlugRegex.MatchString(req.Slug) {
		return nil, ErrSiteInvalidSlug
	}

	// 检查 slug 唯一性
	exists, err := s.siteRepo.SlugExists(ctx, req.Slug, nil)
	if err != nil {
		return nil, fmt.Errorf("check slug failed: %w", err)
	}
	if exists {
		return nil, ErrSiteSlugExists
	}

	site := &model.Site{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		SiteURL:   req.SiteURL,
		Locale:    "zh-CN",  // 默认值
		Timezone:  "Asia/Shanghai",  // 默认值
		IsActive:  true,
		CreatedBy: &userID,
	}

	// 处理可选字段
	if req.Description != nil {
		site.Description = *req.Description
	}
	if req.Locale != nil {
		site.Locale = *req.Locale
	}
	if req.Timezone != nil {
		site.Timezone = *req.Timezone
	}
	if req.SeoTitle != nil {
		site.SeoTitle = *req.SeoTitle
	}
	if req.SeoDescription != nil {
		site.SeoDescription = *req.SeoDescription
	}
	if req.SeoKeywords != nil && len(req.SeoKeywords) > 0 {
		keywordsJSON, _ := json.Marshal(req.SeoKeywords)
		site.SeoKeywords = datatypes.JSON(keywordsJSON)
	}
	if req.Settings != nil {
		site.Settings = *req.Settings
	}
	if req.IsActive != nil {
		site.IsActive = *req.IsActive
	}

	// 直接创建站点（不再初始化默认配置）
	if err := s.db.WithContext(ctx).Create(site).Error; err != nil {
		return nil, fmt.Errorf("create site failed: %w", err)
	}

	return site, nil
}

// Get 获取站点
func (s *SiteService) Get(ctx context.Context, id uuid.UUID) (*model.Site, error) {
	site, err := s.siteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSiteNotFound
	}
	return site, nil
}

// List 列出所有站点（分页）
func (s *SiteService) List(ctx context.Context, page, pageSize int) ([]model.Site, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&model.Site{}).Where("deleted_time IS NULL").Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count sites failed: %w", err)
	}

	var sites []model.Site
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).
		Where("deleted_time IS NULL").
		Order("created_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sites).Error; err != nil {
		return nil, 0, fmt.Errorf("list sites failed: %w", err)
	}

	return sites, total, nil
}

// Update 更新站点
func (s *SiteService) Update(ctx context.Context, id uuid.UUID, req *model.SiteUpdate) (*model.Site, error) {
	site, err := s.siteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrSiteNotFound
	}

	if req.Name != nil {
		site.Name = *req.Name
	}
	if req.Slug != nil {
		// 验证 slug 格式
		if !siteSlugRegex.MatchString(*req.Slug) {
			return nil, ErrSiteInvalidSlug
		}
		// 检查 slug 唯一性（排除当前站点）
		exists, err := s.siteRepo.SlugExists(ctx, *req.Slug, &id)
		if err != nil {
			return nil, fmt.Errorf("check slug failed: %w", err)
		}
		if exists {
			return nil, ErrSiteSlugExists
		}
		site.Slug = *req.Slug
	}
	if req.Description != nil {
		site.Description = *req.Description
	}
	if req.SiteURL != nil {
		site.SiteURL = req.SiteURL
	}
	if req.Locale != nil {
		site.Locale = *req.Locale
	}
	if req.Timezone != nil {
		site.Timezone = *req.Timezone
	}
	if req.SeoTitle != nil {
		site.SeoTitle = *req.SeoTitle
	}
	if req.SeoDescription != nil {
		site.SeoDescription = *req.SeoDescription
	}
	if req.SeoKeywords != nil {
		keywordsJSON, _ := json.Marshal(req.SeoKeywords)
		site.SeoKeywords = datatypes.JSON(keywordsJSON)
	}
	if req.Settings != nil {
		site.Settings = *req.Settings
	}
	if req.IsActive != nil {
		site.IsActive = *req.IsActive
	}

	if err := s.siteRepo.Update(ctx, site); err != nil {
		return nil, err
	}

	return site, nil
}

// Delete 删除站点（软删除）
func (s *SiteService) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := s.db.WithContext(ctx).
		Model(&model.Site{}).
		Where("id = ?", id).
		Update("deleted_time", &now)

	if result.RowsAffected == 0 {
		return ErrSiteNotFound
	}

	return result.Error
}
