// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/gorm"
)

var (
	ErrSystemRoleNotFound     = errors.New("system role not found")
	ErrSystemRoleAlreadyExists = errors.New("system role already exists")
	ErrSystemRoleIsSystem     = errors.New("cannot delete system built-in role")
)

// SystemRoleRepository 系统角色仓储
type SystemRoleRepository struct {
	db *gorm.DB
}

// NewSystemRoleRepository 新建系统角色仓储
func NewSystemRoleRepository(db *gorm.DB) *SystemRoleRepository {
	return &SystemRoleRepository{db: db}
}

// Create 创建系统角色
func (r *SystemRoleRepository) Create(ctx context.Context, role *model.SystemRole) error {
	result := r.db.WithContext(ctx).Create(role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrSystemRoleAlreadyExists
		}
		return result.Error
	}
	return nil
}

// GetByID 根据 ID 获取系统角色
func (r *SystemRoleRepository) GetByID(ctx context.Context, id uid.UID) (*model.SystemRole, error) {
	var role model.SystemRole
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_time IS NULL", id).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSystemRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

// GetByName 根据名称获取系统角色
func (r *SystemRoleRepository) GetByName(ctx context.Context, name string) (*model.SystemRole, error) {
	var role model.SystemRole
	err := r.db.WithContext(ctx).
		Where("name = ? AND deleted_time IS NULL", name).
		First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSystemRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

// List 列出所有系统角色（含系统内置）
func (r *SystemRoleRepository) List(ctx context.Context) ([]model.SystemRole, error) {
	var roles []model.SystemRole
	err := r.db.WithContext(ctx).
		Where("deleted_time IS NULL").
		Order("created_time ASC").
		Find(&roles).Error
	return roles, err
}

// Update 更新系统角色
func (r *SystemRoleRepository) Update(ctx context.Context, role *model.SystemRole) error {
	role.UpdatedTime = time.Now()
	return r.db.WithContext(ctx).
		Where("id = ?", role.ID).
		Updates(map[string]interface{}{
			"name":         role.Name,
			"description":  role.Description,
			"permissions":  role.Permissions,
			"updated_time": role.UpdatedTime,
		}).Error
}

// Delete 软删除系统角色（系统内置角色不可删除）
func (r *SystemRoleRepository) Delete(ctx context.Context, id uid.UID) error {
	// 先检查是否为系统内置角色
	var role model.SystemRole
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_time IS NULL", id).
		First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSystemRoleNotFound
		}
		return err
	}
	if role.IsSystem {
		return ErrSystemRoleIsSystem
	}
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.SystemRole{}).
		Where("id = ?", id).
		Update("deleted_time", &now).Error
}
