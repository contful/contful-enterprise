package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentTypeRepository 内容类型仓储
type ContentTypeRepository struct {
	db *gorm.DB
}

// NewContentTypeRepository 新建仓储
func NewContentTypeRepository(db *gorm.DB) *ContentTypeRepository {
	return &ContentTypeRepository{db: db}
}

// Create 创建内容类型
func (r *ContentTypeRepository) Create(ctx context.Context, ct *model.ContentType) error {
	return r.db.WithContext(ctx).Create(ct).Error
}

// GetByID 根据 ID 获取
func (r *ContentTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ContentType, error) {
	var ct model.ContentType
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// GetByIDWithFields 获取内容类型及其字段
func (r *ContentTypeRepository) GetByIDWithFields(ctx context.Context, id uuid.UUID) (*model.ContentType, error) {
	var ct model.ContentType
	err := r.db.WithContext(ctx).
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("id = ?", id).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// GetBySlug 根据 slug 获取
func (r *ContentTypeRepository) GetBySlug(ctx context.Context, siteID uuid.UUID, slug string) (*model.ContentType, error) {
	var ct model.ContentType
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND slug = ?", siteID, slug).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

// ListBySite 列出站点的内容类型
func (r *ContentTypeRepository) ListBySite(ctx context.Context, siteID uuid.UUID, page, pageSize int) ([]model.ContentType, int64, error) {
	var cts []model.ContentType
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ContentType{}).Where("site_id = ?", siteID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("sort_order ASC, created_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&cts).Error
	if err != nil {
		return nil, 0, err
	}

	return cts, total, nil
}

// Update 更新内容类型
func (r *ContentTypeRepository) Update(ctx context.Context, ct *model.ContentType) error {
	return r.db.WithContext(ctx).Save(ct).Error
}

// Delete 删除内容类型（软删除）
func (r *ContentTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ContentType{}, "id = ?", id).Error
}

// ExistsSlug 检查 slug 是否存在
func (r *ContentTypeRepository) ExistsSlug(ctx context.Context, siteID uuid.UUID, slug string, excludeID *uuid.UUID) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.ContentType{}).
		Where("site_id = ? AND slug = ?", siteID, slug)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
