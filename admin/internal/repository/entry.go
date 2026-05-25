// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"time"

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
		Preload("ContentSchema", func(db *gorm.DB) *gorm.DB {
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
		if filter.ContentSchemaID != nil {
			query = query.Where("schema_id = ?", *filter.ContentSchemaID)
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
		Preload("ContentSchema").
		Order("updated_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

// ListByContentSchema 列出内容类型的条目
func (r *EntryRepository) ListByContentSchema(ctx context.Context, siteID uuid.UUID, contentTypeID uuid.UUID, filter *model.EntryListFilter, page, pageSize int) ([]model.Entry, int64, error) {
	var entries []model.Entry
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("site_id = ? AND schema_id = ?", siteID, contentTypeID)

	// 应用过滤条件
	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.Locale != nil {
			query = query.Where("locale = ?", *filter.Locale)
		}
		// Keyword 搜索：关联 entry_values 表的 text_value 字段
		if filter.Keyword != nil && *filter.Keyword != "" {
			keyword := "%" + *filter.Keyword + "%"
			query = query.Where(`EXISTS (
				SELECT 1 FROM contful_entry_values ev
				WHERE ev.entry_id = entries.id
				AND (ev.text_value ILIKE ? OR CAST(ev.value AS TEXT) ILIKE ?)
			)`, keyword, keyword)
		}
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
		Preload("Values", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Field")
		}).
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

// CountByContentSchema 统计内容类型的条目数量
func (r *EntryRepository) CountByContentSchema(ctx context.Context, contentTypeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("schema_id = ?", contentTypeID).
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

// ============ 定时排期 ============

// FindScheduledToPublish 查找到达定时发布时间的草稿条目
func (r *EntryRepository) FindScheduledToPublish(ctx context.Context) ([]model.Entry, error) {
	var entries []model.Entry
	err := r.db.WithContext(ctx).
		Where("status = ?", model.EntryStatusDraft).
		Where("scheduled_publish_time IS NOT NULL").
		Where("scheduled_publish_time <= NOW()").
		Find(&entries).Error
	return entries, err
}

// FindScheduledToUnpublish 查找到达定时下架时间的已发布条目
func (r *EntryRepository) FindScheduledToUnpublish(ctx context.Context) ([]model.Entry, error) {
	var entries []model.Entry
	err := r.db.WithContext(ctx).
		Where("status = ?", model.EntryStatusPublished).
		Where("scheduled_unpublish_time IS NOT NULL").
		Where("scheduled_unpublish_time <= NOW()").
		Find(&entries).Error
	return entries, err
}

// ExecuteScheduledPublish 执行定时发布：状态→published，记录发布时间，清除定时
func (r *EntryRepository) ExecuteScheduledPublish(ctx context.Context, id uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":                 model.EntryStatusPublished,
			"published_time":         gorm.Expr("NOW()"),
			"scheduled_publish_time": nil,
		})
	return result.RowsAffected, result.Error
}

// ExecuteScheduledUnpublish 执行定时下架：状态→draft，清除发布时间和定时
func (r *EntryRepository) ExecuteScheduledUnpublish(ctx context.Context, id uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":                   model.EntryStatusDraft,
			"published_time":           nil,
			"scheduled_unpublish_time": nil,
		})
	return result.RowsAffected, result.Error
}

// SetScheduleTimes 设置排期时间
func (r *EntryRepository) SetScheduleTimes(ctx context.Context, id uuid.UUID, publishTime, unpublishTime *time.Time) error {
	updates := map[string]interface{}{
		"scheduled_publish_time":   publishTime,
		"scheduled_unpublish_time": unpublishTime,
	}
	return r.db.WithContext(ctx).Model(&model.Entry{}).Where("id = ?", id).Updates(updates).Error
}

// ClearScheduleTimes 清除所有排期时间
func (r *EntryRepository) ClearScheduleTimes(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"scheduled_publish_time":   nil,
			"scheduled_unpublish_time": nil,
		}).Error
}

// ListScheduled 列出有排期的条目（按站点过滤）
func (r *EntryRepository) ListScheduled(ctx context.Context, siteID uuid.UUID, page, pageSize int) ([]model.Entry, int64, error) {
	var entries []model.Entry
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("site_id = ?", siteID).
		Where("scheduled_publish_time IS NOT NULL OR scheduled_unpublish_time IS NOT NULL")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("ContentSchema").
		Order("COALESCE(scheduled_publish_time, scheduled_unpublish_time) ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}
