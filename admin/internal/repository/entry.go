package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EntryRepository 条目仓储
type EntryRepository struct {
	db *gorm.DB
}

// NewEntryRepository 新建仓储
func NewEntryRepository(db *gorm.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

// Create 创建条目
func (r *EntryRepository) Create(ctx context.Context, entry *model.Entry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// GetByID 根据 ID 获取
func (r *EntryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Entry, error) {
	var entry model.Entry
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetByIDWithValues 根据 ID 获取条目及其字段值
func (r *EntryRepository) GetByIDWithValues(ctx context.Context, id uuid.UUID) (*model.Entry, error) {
	var entry model.Entry
	err := r.db.WithContext(ctx).
		Preload("Values", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Field")
		}).
		Where("id = ?", id).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetByIDWithType 根据 ID 获取条目及其内容类型
func (r *EntryRepository) GetByIDWithType(ctx context.Context, id uuid.UUID) (*model.Entry, error) {
	var entry model.Entry
	err := r.db.WithContext(ctx).
		Preload("ContentType", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Fields", func(db *gorm.DB) *gorm.DB {
				return db.Order("sort_order ASC")
			})
		}).
		Preload("Values", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Field")
		}).
		Where("id = ?", id).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListBySite 列出站点的条目
func (r *EntryRepository) ListBySite(ctx context.Context, siteID uuid.UUID, filter *model.EntryListFilter, page, pageSize int) ([]model.Entry, int64, error) {
	var entries []model.Entry
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Entry{}).Where("site_id = ?", siteID)

	// 应用过滤条件
	if filter != nil {
		if filter.ContentTypeID != nil {
			query = query.Where("content_type_id = ?", *filter.ContentTypeID)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Locale != nil {
			query = query.Where("locale = ?", *filter.Locale)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("updated_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// ListByContentType 列出内容类型的条目
func (r *EntryRepository) ListByContentType(ctx context.Context, siteID uuid.UUID, contentTypeID uuid.UUID, filter *model.EntryListFilter, page, pageSize int) ([]model.Entry, int64, error) {
	var entries []model.Entry
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("site_id = ? AND content_type_id = ?", siteID, contentTypeID)

	// 应用过滤条件
	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Locale != nil {
			query = query.Where("locale = ?", *filter.Locale)
		}
		// TODO: Keyword 搜索需要关联 entry_values 表，支持文本搜索
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 构建排序
	orderClause := "sort_weight DESC, updated_time DESC"
	if filter != nil && filter.SortField != "" {
		validFields := map[string]string{
			"updated_time":   "updated_time",
			"created_time":   "created_time",
			"published_time": "published_time",
			"sort_weight":    "sort_weight",
		}
		if field, ok := validFields[filter.SortField]; ok {
			order := "DESC"
			if filter.SortOrder == "asc" {
				order = "ASC"
			}
			orderClause = field + " " + order
		}
	}

	offset := (page - 1) * pageSize
	err := query.
		Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// Update 更新条目
func (r *EntryRepository) Update(ctx context.Context, entry *model.Entry) error {
	return r.db.WithContext(ctx).Save(entry).Error
}

// Delete 删除条目（软删除）
func (r *EntryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Entry{}, "id = ?", id).Error
}

// CountByContentType 统计内容类型的条目数量
func (r *EntryRepository) CountByContentType(ctx context.Context, contentTypeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("content_type_id = ?", contentTypeID).
		Count(&count).Error
	return count, err
}

// ============ 批量操作 ============

// BatchDelete 批量删除
func (r *EntryRepository) BatchDelete(ctx context.Context, ids []uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&model.Entry{}, "id IN ?", ids)
	return result.RowsAffected, result.Error
}

// BatchUpdateStatus 批量更新状态
func (r *EntryRepository) BatchUpdateStatus(ctx context.Context, ids []uuid.UUID, status model.EntryStatus) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id IN ?", ids).
		Update("status", status)
	return result.RowsAffected, result.Error
}

// BatchPublish 批量发布
func (r *EntryRepository) BatchPublish(ctx context.Context, ids []uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":          model.EntryStatusPublished,
			"published_time":  gorm.Expr("NOW()"),
		})
	return result.RowsAffected, result.Error
}

// BatchUnpublish 批量取消发布
func (r *EntryRepository) BatchUnpublish(ctx context.Context, ids []uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id IN ?", ids).
		Update("status", model.EntryStatusDraft)
	return result.RowsAffected, result.Error
}

// ============ EntryValue 操作 ============

// CreateValues 批量创建字段值
func (r *EntryRepository) CreateValues(ctx context.Context, values []model.EntryValue) error {
	if len(values) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&values).Error
}

// GetValuesByEntry 获取条目的所有字段值
func (r *EntryRepository) GetValuesByEntry(ctx context.Context, entryID uuid.UUID) ([]model.EntryValue, error) {
	var values []model.EntryValue
	err := r.db.WithContext(ctx).
		Preload("Field").
		Where("entry_id = ?", entryID).
		Find(&values).Error
	return values, err
}

// UpdateValue 更新字段值
func (r *EntryRepository) UpdateValue(ctx context.Context, value *model.EntryValue) error {
	return r.db.WithContext(ctx).Save(value).Error
}

// DeleteValues 删除条目的所有字段值
func (r *EntryRepository) DeleteValues(ctx context.Context, entryID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.EntryValue{}, "entry_id = ?", entryID).Error
}

// UpsertValue 插入或更新字段值
func (r *EntryRepository) UpsertValue(ctx context.Context, value *model.EntryValue) error {
	return r.db.WithContext(ctx).
		Where("entry_id = ? AND field_id = ?", value.EntryID, value.FieldID).
		Assign(model.EntryValue{
			Value: value.Value,
		}).
		FirstOrCreate(value).Error
}

// ============ EntryVersion 操作 ============

// CreateVersion 创建版本记录
func (r *EntryRepository) CreateVersion(ctx context.Context, version *model.EntryVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersions 获取条目的版本历史
func (r *EntryRepository) GetVersions(ctx context.Context, entryID uuid.UUID) ([]model.EntryVersion, error) {
	var versions []model.EntryVersion
	err := r.db.WithContext(ctx).
		Where("entry_id = ?", entryID).
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

// GetVersion 获取指定版本
func (r *EntryRepository) GetVersion(ctx context.Context, entryID uuid.UUID, version int) (*model.EntryVersion, error) {
	var v model.EntryVersion
	err := r.db.WithContext(ctx).
		Where("entry_id = ? AND version = ?", entryID, version).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// WithTransaction 执行事务
func (r *EntryRepository) WithTransaction(fn func(repo *EntryRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := &EntryRepository{db: tx}
		return fn(repo)
	})
}
