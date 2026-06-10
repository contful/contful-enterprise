// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"errors"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/gorm"
)

var (
	ErrSystemUserRoleNotFound = errors.New("system user role not found")
)

// SystemUserRoleRepository 系统用户-角色关联仓储
type SystemUserRoleRepository struct {
	db *gorm.DB
}

// NewSystemUserRoleRepository 创建系统用户-角色关联仓储
func NewSystemUserRoleRepository(db *gorm.DB) *SystemUserRoleRepository {
	return &SystemUserRoleRepository{db: db}
}

// Create 创建用户-角色关联
func (r *SystemUserRoleRepository) Create(ctx context.Context, sur *model.SystemUserRole) error {
	return r.db.WithContext(ctx).Create(sur).Error
}

// Delete 删除用户-角色关联
func (r *SystemUserRoleRepository) Delete(ctx context.Context, userID, roleID uid.UID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.SystemUserRole{})
	if result.RowsAffected == 0 {
		return ErrSystemUserRoleNotFound
	}
	return result.Error
}

// ListByUser 获取用户的所有角色关联
func (r *SystemUserRoleRepository) ListByUser(ctx context.Context, userID uid.UID) ([]model.SystemUserRole, error) {
	var items []model.SystemUserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&items).Error
	return items, err
}

// ListByRole 获取角色的所有用户关联
func (r *SystemUserRoleRepository) ListByRole(ctx context.Context, roleID uid.UID) ([]model.SystemUserRole, error) {
	var items []model.SystemUserRole
	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&items).Error
	return items, err
}

// Exists 检查用户-角色关联是否存在
func (r *SystemUserRoleRepository) Exists(ctx context.Context, userID, roleID uid.UID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.SystemUserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error
	return count > 0, err
}
