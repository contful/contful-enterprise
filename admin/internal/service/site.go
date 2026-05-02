// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSiteNotFound     = errors.New("site not found")
	ErrSiteSlugExists   = errors.New("site slug already exists")
	ErrSiteInvalidSlug  = errors.New("invalid site slug")

	// siteSlugRegex 站点 slug 规则：字母开头，小写字母+数字+连字符，2-100字符
	siteSlugRegex = regexp.MustCompile(`^[a-z][a-z0-9\-]{0,98}[a-z0-9]$`)
)

// SiteService 站点服务
type SiteService struct {
	db         *gorm.DB
	siteRepo   *repository.SiteRepository
	configRepo *repository.SiteConfigRepository
}

// NewSiteService 新建站点服务
func NewSiteService(db *gorm.DB, siteRepo *repository.SiteRepository, configRepo *repository.SiteConfigRepository) *SiteService {
	return &SiteService{db: db, siteRepo: siteRepo, configRepo: configRepo}
}

// Create 创建站点（同时创建默认角色 + 关联创建者到 site_users）
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
		IsActive:  true,
		CreatedBy: &userID,
	}

	if req.Config != nil {
		site.Config = *req.Config
	}
	if req.IsActive != nil {
		site.IsActive = *req.IsActive
	}

	// 事务：创建站点 + 默认角色 + site_users 关联
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建站点
		if err := tx.Create(site).Error; err != nil {
			return fmt.Errorf("create site failed: %w", err)
		}

		// 2. 创建默认站点角色（Owner）
		ownerRole := &model.SiteRole{
			ID:          uuid.New(),
			SiteID:      site.ID,
			Name:        "Owner",
			Description: "站点所有者，拥有全部权限",
			IsSystem:    true,
			Permissions: []string{
				"content_type:read", "content_type:write", "content_type:delete",
				"entry:read", "entry:write", "entry:delete", "entry:publish",
				"asset:read", "asset:write", "asset:delete",
				"media:read", "media:write", "media:delete",
				"site:read", "site:write", "site:delete",
				"user:read", "user:write", "user:delete",
				"api_token:read", "api_token:write", "api_token:delete",
			},
			ContentPermissions: []string{"*"},
			ChannelPermissions: []string{"*"},
			SortOrder:          0,
		}
		if err := tx.Create(ownerRole).Error; err != nil {
			return fmt.Errorf("create owner role failed: %w", err)
		}

		// 3. 创建默认站点角色（Editor）
		editorRole := &model.SiteRole{
			ID:          uuid.New(),
			SiteID:      site.ID,
			Name:        "Editor",
			Description: "编辑者，可管理内容和媒体",
			IsSystem:    true,
			Permissions: []string{
				"content_type:read", "content_type:write",
				"entry:read", "entry:write", "entry:publish",
				"asset:read", "asset:write",
				"media:read", "media:write",
				"site:read",
			},
			ContentPermissions: []string{"*"},
			ChannelPermissions: []string{},
			SortOrder:          1,
		}
		if err := tx.Create(editorRole).Error; err != nil {
			return fmt.Errorf("create editor role failed: %w", err)
		}

		// 4. 关联创建者到 site_users（Owner 角色）
		siteUser := &model.SiteUser{
			ID:               uuid.New(),
			SiteID:           site.ID,
			UserID:           userID,
			RoleID:           ownerRole.ID,
			Status:           model.UserStatusActive,
			ExtraPermissions: []string{}, // 避免 JSONB NOT NULL 约束
		}
		if err := tx.Create(siteUser).Error; err != nil {
			return fmt.Errorf("create site_user failed: %w", err)
		}

		// 5. 初始化站点默认配置（storage、integrity 等）
		if err := s.initDefaultConfigs(tx, site.ID); err != nil {
			return fmt.Errorf("init default configs failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
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

// ListMySites 列出当前用户所属的站点
func (s *SiteService) ListMySites(ctx context.Context, userID uuid.UUID, page, pageSize int) (*model.SiteListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var sites []model.Site
	var total int64

	subQuery := s.db.WithContext(ctx).
		Model(&model.SiteUser{}).
		Select("site_id").
		Where("user_id = ? AND status = ?", userID, model.UserStatusActive)

	query := s.db.WithContext(ctx).
		Model(&model.Site{}).
		Where("id IN (?) AND deleted_time IS NULL", subQuery)

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count my sites failed: %w", err)
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("created_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&sites).Error
	if err != nil {
		return nil, fmt.Errorf("list my sites failed: %w", err)
	}

	items := make([]model.SiteResponse, len(sites))
	for i, site := range sites {
		items[i] = site.ToResponse()
	}

	return &model.SiteListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// List 列出站点
func (s *SiteService) List(ctx context.Context, page, pageSize int, isActive *bool) (*model.SiteListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sites, total, err := s.siteRepo.List(ctx, page, pageSize, isActive)
	if err != nil {
		return nil, fmt.Errorf("list sites failed: %w", err)
	}

	items := make([]model.SiteResponse, len(sites))
	for i, site := range sites {
		items[i] = site.ToResponse()
	}

	return &model.SiteListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
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
		if !siteSlugRegex.MatchString(*req.Slug) {
			return nil, ErrSiteInvalidSlug
		}
		exists, err := s.siteRepo.SlugExists(ctx, *req.Slug, &id)
		if err != nil {
			return nil, fmt.Errorf("check slug failed: %w", err)
		}
		if exists {
			return nil, ErrSiteSlugExists
		}
		site.Slug = *req.Slug
	}
	if req.SiteURL != nil {
		site.SiteURL = req.SiteURL
	}
	if req.Config != nil {
		site.Config = *req.Config
	}
	if req.IsActive != nil {
		site.IsActive = *req.IsActive
	}

	if err := s.siteRepo.Update(ctx, site); err != nil {
		return nil, fmt.Errorf("update site failed: %w", err)
	}

	return site, nil
}

// Delete 删除站点
func (s *SiteService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.siteRepo.GetByID(ctx, id); err != nil {
		return ErrSiteNotFound
	}
	return s.siteRepo.Delete(ctx, id)
}

// initDefaultConfigs 初始化新站点的默认配置
func (s *SiteService) initDefaultConfigs(tx *gorm.DB, siteID uuid.UUID) error {
	configs := []model.SiteConfig{
		{SiteID: siteID, ConfigKey: "storage.driver", ConfigValue: "local", ConfigType: "string", ConfigGroup: "storage", IsEncrypted: false, IsReadonly: false, Description: "存储驱动类型: local/oss/cos/obs/s3"},
		{SiteID: siteID, ConfigKey: "storage.local.root", ConfigValue: "uploads", ConfigType: "string", ConfigGroup: "storage", IsEncrypted: false, IsReadonly: false, Description: "本地存储根目录"},
		{SiteID: siteID, ConfigKey: "storage.local.base_url", ConfigValue: "/uploads", ConfigType: "string", ConfigGroup: "storage", IsEncrypted: false, IsReadonly: false, Description: "本地存储访问路径"},
		{SiteID: siteID, ConfigKey: "integrity.enabled", ConfigValue: "false", ConfigType: "boolean", ConfigGroup: "integrity", IsEncrypted: false, IsReadonly: false, Description: "是否启用数据签名"},
		{SiteID: siteID, ConfigKey: "integrity.algorithm", ConfigValue: "HMAC-SHA256", ConfigType: "string", ConfigGroup: "integrity", IsEncrypted: false, IsReadonly: false, Description: "签名算法"},
		{SiteID: siteID, ConfigKey: "integrity.signing_key", ConfigValue: "", ConfigType: "encrypted", ConfigGroup: "integrity", IsEncrypted: true, IsReadonly: false, Description: "签名密钥（AES-256-GCM 加密存储）"},
	}
	for _, cfg := range configs {
		if err := tx.Create(&cfg).Error; err != nil {
			return err
		}
	}
	return nil
}
