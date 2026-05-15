// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// ListGroupsWithPermissions 获取所有权限分组及其权限项
func (r *PermissionRepository) ListGroupsWithPermissions(ctx context.Context) ([]model.PermissionGroup, []model.Permission, error) {
	var groups []model.PermissionGroup
	if err := r.db.WithContext(ctx).Order("sort_order ASC").Find(&groups).Error; err != nil {
		return nil, nil, err
	}

	var perms []model.Permission
	if err := r.db.WithContext(ctx).Order("sort_order ASC").Find(&perms).Error; err != nil {
		return nil, nil, err
	}

	return groups, perms, nil
}

// ── 权限分组 CRUD ──

func (r *PermissionRepository) CreateGroup(ctx context.Context, g *model.PermissionGroup) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *PermissionRepository) UpdateGroup(ctx context.Context, g *model.PermissionGroup) error {
	return r.db.WithContext(ctx).Model(g).Select("label", "label_en", "sort_order").Updates(g).Error
}

func (r *PermissionRepository) DeleteGroup(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.PermissionGroup{}, "id = ?", id).Error
}

// ── 权限项 CRUD ──

func (r *PermissionRepository) CreatePermission(ctx context.Context, p *model.Permission) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PermissionRepository) UpdatePermission(ctx context.Context, p *model.Permission) error {
	return r.db.WithContext(ctx).Model(p).Select("label", "label_en", "sort_order").Updates(p).Error
}

func (r *PermissionRepository) DeletePermission(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Permission{}, "id = ?", id).Error
}
