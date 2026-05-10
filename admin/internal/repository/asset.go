// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetRepository 资源仓储
type AssetRepository struct {
	db *gorm.DB
}

// NewAssetRepository 新建仓储
func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

// Create 创建资源
func (r *AssetRepository) Create(ctx context.Context, asset *model.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// GetByID 根据 ID 获取
func (r *AssetRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	var asset model.Asset
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetByUUID 根据 UUID 获取
func (r *AssetRepository) GetByUUID(ctx context.Context, uuid string) (*model.Asset, error) {
	var asset model.Asset
	err := r.db.WithContext(ctx).
		Where("uuid = ?", uuid).
		First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// GetByFileHash 根据文件哈希获取
func (r *AssetRepository) GetByFileHash(ctx context.Context, siteID uuid.UUID, hash string) (*model.Asset, error) {
	var asset model.Asset
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND file_hash = ?", siteID, hash).
		First(&asset).Error
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// List 列出资源
func (r *AssetRepository) List(ctx context.Context, siteID uuid.UUID, filter *model.AssetListFilter, page, pageSize int) ([]model.Asset, int64, error) {
	var assets []model.Asset
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Asset{}).Where("site_id = ?", siteID)

	// 应用过滤条件
	if filter != nil {
		if filter.FolderID != nil {
			query = query.Where("folder_id = ?", *filter.FolderID)
		}
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.Extension != nil {
			query = query.Where("extension = ?", *filter.Extension)
		}
		if filter.Tag != nil {
			query = query.Where("? = ANY(tags)", *filter.Tag)
		}
		if filter.Keyword != nil && *filter.Keyword != "" {
			query = query.Where("name ILIKE ?", "%"+*filter.Keyword+"%")
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
		Find(&assets).Error
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// ListByFolder 列出文件夹中的资源
func (r *AssetRepository) ListByFolder(ctx context.Context, siteID uuid.UUID, folderID *uuid.UUID, page, pageSize int) ([]model.Asset, int64, error) {
	var assets []model.Asset
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Asset{}).Where("site_id = ?", siteID)

	if folderID != nil {
		query = query.Where("folder_id = ?", *folderID)
	} else {
		query = query.Where("folder_id IS NULL")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("name ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&assets).Error
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

// Update 更新资源
func (r *AssetRepository) Update(ctx context.Context, asset *model.Asset) error {
	return r.db.WithContext(ctx).Save(asset).Error
}

// Delete 删除资源（软删除）
func (r *AssetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Asset{}, "id = ?", id).Error
}

// BatchDelete 批量删除
func (r *AssetRepository) BatchDelete(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Asset{}, "id IN ?", ids).Error
}

// IncrementUsedCount 增加引用计数
func (r *AssetRepository) IncrementUsedCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Asset{}).
		Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + ?", 1)).Error
}

// DecrementUsedCount 减少引用计数
func (r *AssetRepository) DecrementUsedCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Asset{}).
		Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("GREATEST(used_count - 1, 0)")).Error
}

// GetByIDs 根据 ID 列表获取
func (r *AssetRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Asset, error) {
	var assets []model.Asset
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&assets).Error
	return assets, err
}

// ============ Folder 操作 ============

// CreateFolder 创建文件夹
func (r *AssetRepository) CreateFolder(ctx context.Context, folder *model.AssetFolder) error {
	return r.db.WithContext(ctx).Create(folder).Error
}

// GetFolderByID 获取文件夹
func (r *AssetRepository) GetFolderByID(ctx context.Context, id uuid.UUID) (*model.AssetFolder, error) {
	var folder model.AssetFolder
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&folder).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

// ListFolders 列出文件夹
func (r *AssetRepository) ListFolders(ctx context.Context, siteID uuid.UUID, parentID *uuid.UUID) ([]model.AssetFolder, error) {
	var folders []model.AssetFolder
	query := r.db.WithContext(ctx).Model(&model.AssetFolder{}).Where("site_id = ?", siteID)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	err := query.Order("sort_order ASC, name ASC").Find(&folders).Error
	return folders, err
}

// GetFolderTree 获取完整文件夹树
func (r *AssetRepository) GetFolderTree(ctx context.Context, siteID uuid.UUID) ([]model.AssetFolder, error) {
	var folders []model.AssetFolder
	err := r.db.WithContext(ctx).
		Preload("Children").
		Where("site_id = ?", siteID).
		Where("parent_id IS NULL").
		Order("sort_order ASC, name ASC").
		Find(&folders).Error
	return folders, err
}

// UpdateFolder 更新文件夹
func (r *AssetRepository) UpdateFolder(ctx context.Context, folder *model.AssetFolder) error {
	return r.db.WithContext(ctx).Save(folder).Error
}

// DeleteFolder 删除文件夹（递归删除子文件夹和资源）
func (r *AssetRepository) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 获取所有子文件夹 ID
		var childIDs []uuid.UUID
		if err := tx.Model(&model.AssetFolder{}).Where("parent_id = ?", id).Pluck("id", &childIDs).Error; err != nil {
			return err
		}

		// 2. 递归删除子文件夹
		for _, childID := range childIDs {
			if err := tx.Delete(&model.AssetFolder{}, "id = ?", childID).Error; err != nil {
				return err
			}
		}

		// 3. 删除该文件夹下的所有资源
		if err := tx.Where("folder_id = ?", id).Delete(&model.Asset{}).Error; err != nil {
			return err
		}

		// 4. 删除文件夹本身
		return tx.Delete(&model.AssetFolder{}, "id = ?", id).Error
	})
}

// CountGlobal 统计全局资源总数（不限定站点）
func (r *AssetRepository) CountGlobal(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Asset{}).Count(&count).Error
	return count, err
}

// WithTransaction 执行事务
func (r *AssetRepository) WithTransaction(fn func(repo *AssetRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		repo := &AssetRepository{db: tx}
		return fn(repo)
	})
}
