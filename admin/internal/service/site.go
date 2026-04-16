package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"

	"github.com/google/uuid"
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
	siteRepo *repository.SiteRepository
}

// NewSiteService 新建服务
func NewSiteService(siteRepo *repository.SiteRepository) *SiteService {
	return &SiteService{siteRepo: siteRepo}
}

// Create 创建站点
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
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		FaviconURL:  req.FaviconURL,
		IsActive:    true,
		Plan:        "free",
		CreatedBy:   &userID,
	}

	if req.Config != nil {
		site.Config = *req.Config
	}
	if req.SEO != nil {
		site.SEO = *req.SEO
	}
	if req.CustomDomains != nil {
		site.CustomDomains = *req.CustomDomains
	}
	if req.IsActive != nil {
		site.IsActive = *req.IsActive
	}
	if req.Plan != nil {
		site.Plan = *req.Plan
	}

	if err := s.siteRepo.Create(ctx, site); err != nil {
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
	if req.Description != nil {
		site.Description = *req.Description
	}
	if req.LogoURL != nil {
		site.LogoURL = req.LogoURL
	}
	if req.FaviconURL != nil {
		site.FaviconURL = req.FaviconURL
	}
	if req.Config != nil {
		site.Config = *req.Config
	}
	if req.SEO != nil {
		site.SEO = *req.SEO
	}
	if req.CustomDomains != nil {
		site.CustomDomains = *req.CustomDomains
	}
	if req.IsActive != nil {
		site.IsActive = *req.IsActive
	}
	if req.Plan != nil {
		site.Plan = *req.Plan
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
