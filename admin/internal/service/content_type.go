package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// ContentTypeService 内容类型服务
type ContentTypeService struct {
	ctRepo    *repository.ContentTypeRepository
	fieldRepo *repository.FieldRepository
	logger    zerolog.Logger
}

// NewContentTypeService 新建服务
func NewContentTypeService(
	ctRepo *repository.ContentTypeRepository,
	fieldRepo *repository.FieldRepository,
	logger zerolog.Logger,
) *ContentTypeService {
	return &ContentTypeService{
		ctRepo:    ctRepo,
		fieldRepo: fieldRepo,
		logger:    logger,
	}
}

// 创建内容类型错误
var (
	ErrContentTypeNotFound = errors.New("content type not found")
	ErrSlugAlreadyExists   = errors.New("slug already exists")
	ErrInvalidSlug         = errors.New("invalid slug format")
	ErrCannotChangeKind    = errors.New("cannot change content type kind after creation")
)

// slug 正则：只允许字母、数字、连字符
var slugRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Create 创建内容类型
func (s *ContentTypeService) Create(ctx context.Context, siteID uuid.UUID, userID *uuid.UUID, req *model.ContentTypeCreate) (*model.ContentTypeResponse, error) {
	// 验证 slug 格式
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if !slugRegex.MatchString(slug) {
		return nil, ErrInvalidSlug
	}

	// 检查 slug 是否已存在
	exists, err := s.ctRepo.ExistsSlug(ctx, siteID, slug, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("check slug exists failed")
		return nil, err
	}
	if exists {
		return nil, ErrSlugAlreadyExists
	}

	// 创建内容类型
	ct := &model.ContentType{
		ID:                   uuid.New(),
		SiteID:               siteID,
		Name:                 strings.TrimSpace(req.Name),
		Slug:                 slug,
		Description:          req.Description,
		Kind:                 req.Kind,
		DisplayConfig:        req.DisplayConfig,
		APISConfig:           req.APISConfig,
		PreviewConfig:        req.PreviewConfig,
		VersioningEnabled:    req.VersioningEnabled,
		DraftAutosaveInterval: req.DraftAutosaveInterval,
		IsActive:             true,
		SortOrder:            req.SortOrder,
		CreatedBy:            userID,
	}

	// 默认 API 配置
	if ct.APISConfig == nil {
		ct.APISConfig = model.JSONB{
			"publicRead":  false,
			"publicWrite": false,
		}
	}

	if err := s.ctRepo.Create(ctx, ct); err != nil {
		s.logger.Error().Err(err).Msg("create content type failed")
		return nil, err
	}

	s.logger.Info().
		Str("content_type", ct.Name).
		Str("slug", ct.Slug).
		Str("kind", string(ct.Kind)).
		Msg("content type created")

	result, err := s.ctRepo.GetByIDWithFields(ctx, ct.ID)
	if err != nil {
		return nil, err
	}
	resp := result.ToResponse()
	return &resp, nil
}

// Get 获取内容类型
func (s *ContentTypeService) Get(ctx context.Context, siteID uuid.UUID, id uuid.UUID) (*model.ContentTypeResponse, error) {
	ct, err := s.ctRepo.GetByIDWithFields(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentTypeNotFound
		}
		return nil, err
	}

	// 验证站点 ID
	if ct.SiteID != siteID {
		return nil, ErrContentTypeNotFound
	}

	resp := ct.ToResponse()
	return &resp, nil
}

// List 列出内容类型
func (s *ContentTypeService) List(ctx context.Context, siteID uuid.UUID, page, pageSize int) (*model.ContentTypeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	cts, total, err := s.ctRepo.ListBySite(ctx, siteID, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]model.ContentTypeResponse, len(cts))
	for i, ct := range cts {
		items[i] = ct.ToResponse()
	}

	return &model.ContentTypeListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update 更新内容类型
func (s *ContentTypeService) Update(ctx context.Context, siteID uuid.UUID, id uuid.UUID, req *model.ContentTypeUpdate) (*model.ContentTypeResponse, error) {
	ct, err := s.ctRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentTypeNotFound
		}
		return nil, err
	}

	// 验证站点 ID
	if ct.SiteID != siteID {
		return nil, ErrContentTypeNotFound
	}

	// 不能修改 kind
	if req.Kind != nil && *req.Kind != ct.Kind {
		return nil, ErrCannotChangeKind
	}
	req.Kind = nil

	// 检查 slug 冲突
	if req.Slug != nil {
		slug := strings.ToLower(strings.TrimSpace(*req.Slug))
		if !slugRegex.MatchString(slug) {
			return nil, ErrInvalidSlug
		}
		exists, err := s.ctRepo.ExistsSlug(ctx, siteID, slug, &id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrSlugAlreadyExists
		}
		ct.Slug = slug
	}

	// 更新字段
	if req.Name != nil {
		ct.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		ct.Description = *req.Description
	}
	if req.DisplayConfig != nil {
		ct.DisplayConfig = *req.DisplayConfig
	}
	if req.APISConfig != nil {
		ct.APISConfig = *req.APISConfig
	}
	if req.PreviewConfig != nil {
		ct.PreviewConfig = *req.PreviewConfig
	}
	if req.VersioningEnabled != nil {
		ct.VersioningEnabled = *req.VersioningEnabled
	}
	if req.DraftAutosaveInterval != nil {
		ct.DraftAutosaveInterval = req.DraftAutosaveInterval
	}
	if req.IsActive != nil {
		ct.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		ct.SortOrder = *req.SortOrder
	}

	if err := s.ctRepo.Update(ctx, ct); err != nil {
		return nil, err
	}

	// 获取完整信息
	result, err := s.ctRepo.GetByIDWithFields(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := result.ToResponse()
	return &resp, nil
}

// Delete 删除内容类型
func (s *ContentTypeService) Delete(ctx context.Context, siteID uuid.UUID, id uuid.UUID) error {
	ct, err := s.ctRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContentTypeNotFound
		}
		return err
	}

	// 验证站点 ID
	if ct.SiteID != siteID {
		return ErrContentTypeNotFound
	}

	// 删除内容类型（字段会自动级联删除）
	if err := s.ctRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info().
		Str("content_type", ct.Name).
		Str("id", id.String()).
		Msg("content type deleted")

	return nil
}

// ============ Field 操作 ============

// CreateField 创建字段
func (s *ContentTypeService) CreateField(ctx context.Context, siteID uuid.UUID, contentTypeID uuid.UUID, req *model.FieldCreate) (*model.FieldResponse, error) {
	// 验证内容类型存在且属于该站点
	ct, err := s.ctRepo.GetByID(ctx, contentTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentTypeNotFound
		}
		return nil, err
	}
	if ct.SiteID != siteID {
		return nil, ErrContentTypeNotFound
	}

	// 验证字段名格式（只能字母、数字、下划线）
	name := strings.TrimSpace(req.Name)
	if !isValidFieldName(name) {
		return nil, errors.New("invalid field name format")
	}

	// 检查字段名是否已存在
	exists, err := s.fieldRepo.ExistsName(ctx, contentTypeID, name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("field name already exists")
	}

	// 获取最大排序号
	maxOrder, _ := s.fieldRepo.GetMaxSortOrder(ctx, contentTypeID)
	sortOrder := req.SortOrder
	if sortOrder == 0 {
		sortOrder = maxOrder + 1
	}

	field := &model.Field{
		ID:            uuid.New(),
		ContentTypeID: contentTypeID,
		Name:          name,
		Label:         strings.TrimSpace(req.Label),
		Description:   req.Description,
		FieldType:     req.FieldType,
		Config:        req.Config,
		Validation:    req.Validation,
		Display:       req.Display,
		DefaultValue:  req.DefaultValue,
		SortOrder:     sortOrder,
	}

	if err := s.fieldRepo.Create(ctx, field); err != nil {
		return nil, err
	}

	s.logger.Info().
		Str("field", field.Name).
		Str("type", field.FieldType).
		Str("content_type", ct.Name).
		Msg("field created")

	resp := field.ToResponse()
	return &resp, nil
}

// ListFields 列出字段
func (s *ContentTypeService) ListFields(ctx context.Context, siteID uuid.UUID, contentTypeID uuid.UUID) ([]model.FieldResponse, error) {
	// 验证内容类型
	ct, err := s.ctRepo.GetByID(ctx, contentTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContentTypeNotFound
		}
		return nil, err
	}
	if ct.SiteID != siteID {
		return nil, ErrContentTypeNotFound
	}

	fields, err := s.fieldRepo.ListByContentType(ctx, contentTypeID)
	if err != nil {
		return nil, err
	}

	responses := make([]model.FieldResponse, len(fields))
	for i, f := range fields {
		responses[i] = f.ToResponse()
	}
	return responses, nil
}

// UpdateField 更新字段
func (s *ContentTypeService) UpdateField(ctx context.Context, siteID uuid.UUID, fieldID uuid.UUID, req *model.FieldUpdate) (*model.FieldResponse, error) {
	field, err := s.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		return nil, err
	}

	// 验证内容类型
	ct, err := s.ctRepo.GetByID(ctx, field.ContentTypeID)
	if err != nil {
		return nil, ErrContentTypeNotFound
	}
	if ct.SiteID != siteID {
		return nil, ErrContentTypeNotFound
	}

	// 检查字段名冲突
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if !isValidFieldName(name) {
			return nil, errors.New("invalid field name format")
		}
		exists, err := s.fieldRepo.ExistsName(ctx, field.ContentTypeID, name, &fieldID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("field name already exists")
		}
		field.Name = name
	}

	// 更新字段
	if req.Label != nil {
		field.Label = strings.TrimSpace(*req.Label)
	}
	if req.Description != nil {
		field.Description = *req.Description
	}
	if req.FieldType != nil {
		field.FieldType = *req.FieldType
	}
	if req.Config != nil {
		field.Config = *req.Config
	}
	if req.Validation != nil {
		field.Validation = *req.Validation
	}
	if req.Display != nil {
		field.Display = *req.Display
	}
	if req.DefaultValue != nil {
		field.DefaultValue = req.DefaultValue
	}
	if req.SortOrder != nil {
		field.SortOrder = *req.SortOrder
	}

	if err := s.fieldRepo.Update(ctx, field); err != nil {
		return nil, err
	}

	resp := field.ToResponse()
	return &resp, nil
}

// DeleteField 删除字段
func (s *ContentTypeService) DeleteField(ctx context.Context, siteID uuid.UUID, fieldID uuid.UUID) error {
	field, err := s.fieldRepo.GetByID(ctx, fieldID)
	if err != nil {
		return err
	}

	// 验证内容类型
	ct, err := s.ctRepo.GetByID(ctx, field.ContentTypeID)
	if err != nil {
		return ErrContentTypeNotFound
	}
	if ct.SiteID != siteID {
		return ErrContentTypeNotFound
	}

	return s.fieldRepo.Delete(ctx, fieldID)
}

// ReorderFields 重新排序字段
func (s *ContentTypeService) ReorderFields(ctx context.Context, siteID uuid.UUID, contentTypeID uuid.UUID, orders map[uuid.UUID]int) error {
	// 验证内容类型
	ct, err := s.ctRepo.GetByID(ctx, contentTypeID)
	if err != nil {
		return ErrContentTypeNotFound
	}
	if ct.SiteID != siteID {
		return ErrContentTypeNotFound
	}

	return s.fieldRepo.ReorderFields(ctx, orders)
}

// isValidFieldName 验证字段名格式
func isValidFieldName(name string) bool {
	if len(name) == 0 || len(name) > 100 {
		return false
	}
	// 必须是字母开头，只能包含字母、数字、下划线
	validName := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	return validName.MatchString(name)
}
