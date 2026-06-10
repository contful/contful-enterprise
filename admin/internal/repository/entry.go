// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"

	"github.com/contful/contful/admin/pkg/uid"
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
func (r *EntryRepository) GetByID(ctx context.Context, id uid.UID) (*model.Entry, error) {
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
func (r *EntryRepository) GetByIDWithValues(ctx context.Context, id uid.UID) (*model.Entry, error) {
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
func (r *EntryRepository) GetByIDWithType(ctx context.Context, id uid.UID) (*model.Entry, error) {
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
func (r *EntryRepository) ListBySite(ctx context.Context, siteID uid.UID, filter *model.EntryListFilter, page, pageSize int) ([]model.Entry, int64, error) {
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
func (r *EntryRepository) ListByContentSchema(ctx context.Context, siteID uid.UID, contentTypeID uid.UID, filter *model.EntryListFilter, page, pageSize int) ([]model.Entry, int64, error) {
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
func (r *EntryRepository) Delete(ctx context.Context, id uid.UID) error {
	return r.db.WithContext(ctx).Delete(&model.Entry{}, "id = ?", id).Error
}

// CountByContentSchema 统计内容类型的条目数量
func (r *EntryRepository) CountByContentSchema(ctx context.Context, contentTypeID uid.UID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("schema_id = ?", contentTypeID).
		Count(&count).Error
	return count, err
}

// ============ 批量操作 ============

// BatchDelete 批量删除
func (r *EntryRepository) BatchDelete(ctx context.Context, ids []uid.UID) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&model.Entry{}, "id IN ?", ids)
	return result.RowsAffected, result.Error
}

// BatchUpdateStatus 批量更新状态
func (r *EntryRepository) BatchUpdateStatus(ctx context.Context, ids []uid.UID, status model.EntryStatus) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id IN ?", ids).
		Update("status", status)
	return result.RowsAffected, result.Error
}

// BatchPublish 批量发布
func (r *EntryRepository) BatchPublish(ctx context.Context, ids []uid.UID) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.Entry{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"status":          model.EntryStatusPublished,
			"published_time":  gorm.Expr("NOW()"),
		})
	return result.RowsAffected, result.Error
}

// BatchUnpublish 批量取消发布
func (r *EntryRepository) BatchUnpublish(ctx context.Context, ids []uid.UID) (int64, error) {
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
func (r *EntryRepository) GetValuesByEntry(ctx context.Context, entryID uid.UID) ([]model.EntryValue, error) {
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
func (r *EntryRepository) DeleteValues(ctx context.Context, entryID uid.UID) error {
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
func (r *EntryRepository) GetVersions(ctx context.Context, entryID uid.UID) ([]model.EntryVersion, error) {
	var versions []model.EntryVersion
	err := r.db.WithContext(ctx).
		Where("entry_id = ?", entryID).
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

// GetVersion 获取指定版本
func (r *EntryRepository) GetVersion(ctx context.Context, entryID uid.UID, version int) (*model.EntryVersion, error) {
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

// ============ 排期相关查询 ============

// FindDuePublish 查找到期待发布的条目
func (r *EntryRepository) FindDuePublish(ctx context.Context) ([]model.Entry, error) {
	var entries []model.Entry
	err := r.db.WithContext(ctx).
		Where("scheduled_publish_time <= NOW()").
		Where("status = ?", model.EntryStatusDraft).
		Where("deleted_time IS NULL").
		Find(&entries).Error
	return entries, err
}

// FindDueUnpublish 查找到期待下架的条目
func (r *EntryRepository) FindDueUnpublish(ctx context.Context) ([]model.Entry, error) {
	var entries []model.Entry
	err := r.db.WithContext(ctx).
		Where("scheduled_unpublish_time <= NOW()").
		Where("status = ?", model.EntryStatusPublished).
		Where("deleted_time IS NULL").
		Find(&entries).Error
	return entries, err
}

// ListScheduled 查询排期条目列表
func (r *EntryRepository) ListScheduled(ctx context.Context, siteID uid.UID, filter *model.ScheduledEntryFilter) ([]model.Entry, int64, error) {
	var entries []model.Entry
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Entry{}).Where("site_id = ?", siteID)

	if filter == nil {
		filter = &model.ScheduledEntryFilter{}
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 50
	}

	// 根据 status 筛选
	if filter.Status != nil {
		switch *filter.Status {
		case "pending_publish":
			query = query.Where("scheduled_publish_time IS NOT NULL AND status = ?", model.EntryStatusDraft)
		case "pending_unpublish":
			query = query.Where("scheduled_unpublish_time IS NOT NULL AND status = ?", model.EntryStatusPublished)
		case "all":
			query = query.Where("(scheduled_publish_time IS NOT NULL OR scheduled_unpublish_time IS NOT NULL)")
		}
	} else {
		query = query.Where("(scheduled_publish_time IS NOT NULL OR scheduled_unpublish_time IS NOT NULL)")
	}

	if filter.From != nil {
		query = query.Where("(scheduled_publish_time >= ? OR scheduled_unpublish_time >= ?)", *filter.From, *filter.From)
	}
	if filter.To != nil {
		query = query.Where("(scheduled_publish_time <= ? OR scheduled_unpublish_time <= ?)", *filter.To, *filter.To)
	}

	query = query.Where("deleted_time IS NULL")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := query.
		Preload("ContentSchema").
		Order("COALESCE(scheduled_publish_time, scheduled_unpublish_time) ASC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}
