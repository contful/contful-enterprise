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
