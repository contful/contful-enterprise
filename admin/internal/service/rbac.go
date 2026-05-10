// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound     = errors.New("role not found")
	ErrRoleIsSystem     = errors.New("system built-in role cannot be deleted")
	ErrRoleNameExists   = errors.New("role name already exists")
	ErrPermissionDenied = errors.New("permission denied")
)

// permCacheEntry Redis 中缓存的权限条目
type permCacheEntry struct {
	SystemPerms []string `json:"system_perms"`
}

// RBACService RBAC 权限管理服务
type RBACService struct {
	db           *gorm.DB
	redis        *redis.Client
	systemRoleRepo *repository.SystemRoleRepository
	userRepo     *repository.UserRepository
}

// NewRBACService 创建 RBAC 服务
func NewRBACService(
	db *gorm.DB,
	redis *redis.Client,
	systemRoleRepo *repository.SystemRoleRepository,
	userRepo *repository.UserRepository,
) *RBACService {
	return &RBACService{
		db:             db,
		redis:          redis,
		systemRoleRepo: systemRoleRepo,
		userRepo:       userRepo,
	}
}

// ────────────────────────────────────────────────────────────
// 权限检查核心逻辑
// ────────────────────────────────────────────────────────────

const (
	rbacCachePrefix = "rbac:user:"
	rbacCacheSuffix = ":perms"
	rbacCacheTTL    = 15 * time.Minute
)

func rbacCacheKey(userID uuid.UUID) string {
	return rbacCachePrefix + userID.String() + rbacCacheSuffix
}

// HasPermission 检查用户是否拥有指定权限
// 超级管理员直接放行，否则从缓存或数据库加载权限列表
func (s *RBACService) HasPermission(ctx context.Context, userID uuid.UUID, isSuperAdmin bool, permission string) (bool, error) {
	// 超级管理员直接放行
	if isSuperAdmin {
		return true, nil
	}

	// 从缓存或数据库获取权限列表
	entry, err := s.getOrLoadPerms(ctx, userID)
	if err != nil {
		return false, err
	}

	// 检查系统级权限
	if matchPermission(entry.SystemPerms, permission) {
		return true, nil
	}

	return false, nil
}

// InvalidateUserCache 清除用户权限缓存（角色变更时调用）
func (s *RBACService) InvalidateUserCache(ctx context.Context, userID uuid.UUID) {
	s.redis.Del(ctx, rbacCacheKey(userID))
}

// getOrLoadPerms 从 Redis 缓存获取权限，不存在则从数据库加载
func (s *RBACService) getOrLoadPerms(ctx context.Context, userID uuid.UUID) (*permCacheEntry, error) {
	key := rbacCacheKey(userID)

	// 尝试从缓存读取
	raw, err := s.redis.Get(ctx, key).Bytes()
	if err == nil {
		var entry permCacheEntry
		if jsonErr := json.Unmarshal(raw, &entry); jsonErr == nil {
			return &entry, nil
		}
	}

	// 缓存未命中，从数据库加载
	entry, err := s.loadPermsFromDB(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if data, jsonErr := json.Marshal(entry); jsonErr == nil {
		s.redis.Set(ctx, key, data, rbacCacheTTL)
	}

	return entry, nil
}

// loadPermsFromDB 从数据库加载用户系统级权限（通过 system_user_roles 关联表）
func (s *RBACService) loadPermsFromDB(ctx context.Context, userID uuid.UUID) (*permCacheEntry, error) {
	entry := &permCacheEntry{
		SystemPerms: []string{},
	}

	// 加载用户所有系统角色的权限（通过 system_user_roles 关联表）
	type rolePermRow struct {
		Permissions []string `gorm:"type:jsonb;serializer:json"`
	}
	var rows []rolePermRow

	err := s.db.WithContext(ctx).
		Table("system_user_roles sur").
		Select("sr.permissions").
		Joins("JOIN system_roles sr ON sr.id = sur.role_id AND sr.deleted_time IS NULL").
		Where("sur.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("load system perms failed: %w", err)
	}

	// 合并去重所有角色的权限
	permMap := make(map[string]struct{})
	for _, row := range rows {
		for _, p := range row.Permissions {
			permMap[p] = struct{}{}
		}
	}

	perms := make([]string, 0, len(permMap))
	for p := range permMap {
		perms = append(perms, p)
	}
	entry.SystemPerms = perms

	return entry, nil
}

// matchPermission 检查权限列表是否包含指定权限（支持通配符 *）
func matchPermission(perms []string, required string) bool {
	for _, p := range perms {
		if p == "*" {
			return true
		}
		if p == required {
			return true
		}
		// 通配符匹配：content_schema:* 可以匹配 content_schema:read
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, ":*")
			if strings.HasPrefix(required, prefix+":") {
				return true
			}
		}
	}
	return false
}

// ────────────────────────────────────────────────────────────
// 系统角色管理
// ────────────────────────────────────────────────────────────

// ListSystemRoles 列出所有系统角色
func (s *RBACService) ListSystemRoles(ctx context.Context) ([]model.SystemRole, error) {
	return s.systemRoleRepo.List(ctx)
}

// GetSystemRole 获取系统角色详情
func (s *RBACService) GetSystemRole(ctx context.Context, id uuid.UUID) (*model.SystemRole, error) {
	return s.systemRoleRepo.GetByID(ctx, id)
}

// CreateSystemRole 创建系统角色
func (s *RBACService) CreateSystemRole(ctx context.Context, req *model.SystemRoleCreateRequest) (*model.SystemRole, error) {
	role := &model.SystemRole{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false,
		Permissions: req.Permissions,
	}
	if err := s.systemRoleRepo.Create(ctx, role); err != nil {
		if errors.Is(err, repository.ErrSystemRoleAlreadyExists) {
			return nil, ErrRoleNameExists
		}
		return nil, err
	}
	return role, nil
}

// UpdateSystemRole 更新系统角色
func (s *RBACService) UpdateSystemRole(ctx context.Context, id uuid.UUID, req *model.SystemRoleUpdateRequest) (*model.SystemRole, error) {
	role, err := s.systemRoleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrRoleNotFound
	}

	// 系统内置角色只允许修改 permissions，不允许改名/改描述
	if !role.IsSystem {
		if req.Name != nil {
			role.Name = *req.Name
		}
		if req.Description != nil {
			role.Description = *req.Description
		}
	}
	if req.Permissions != nil {
		role.Permissions = req.Permissions
	}

	if err := s.systemRoleRepo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// DeleteSystemRole 删除系统角色（系统内置角色不可删除）
func (s *RBACService) DeleteSystemRole(ctx context.Context, id uuid.UUID) error {
	err := s.systemRoleRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrSystemRoleNotFound) {
			return ErrRoleNotFound
		}
		if errors.Is(err, repository.ErrSystemRoleIsSystem) {
			return ErrRoleIsSystem
		}
	}
	return err
}

// ────────────────────────────────────────────────────────────
// 用户-角色关联管理
// ────────────────────────────────────────────────────────────

// AssignUserRole 为用户分配系统角色
func (s *RBACService) AssignUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	// 验证角色存在
	_, err := s.systemRoleRepo.GetByID(ctx, roleID)
	if err != nil {
		return ErrRoleNotFound
	}

	// 检查是否已关联
	var count int64
	s.db.WithContext(ctx).
		Model(&model.SystemUserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count)
	if count > 0 {
		return errors.New("user already has this role")
	}

	// 创建关联
	sur := &model.SystemUserRole{
		ID:     uuid.New(),
		UserID: userID,
		RoleID: roleID,
	}
	if err := s.db.WithContext(ctx).Create(sur).Error; err != nil {
		return err
	}

	// 清除用户权限缓存
	s.InvalidateUserCache(ctx, userID)
	return nil
}

// RemoveUserRole 移除用户的系统角色
func (s *RBACService) RemoveUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	result := s.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&model.SystemUserRole{})
	if result.RowsAffected == 0 {
		return errors.New("user-role association not found")
	}

	// 清除用户权限缓存
	s.InvalidateUserCache(ctx, userID)
	return nil
}

// GetUserRoles 获取用户的所有系统角色
func (s *RBACService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]model.SystemRole, error) {
	var roles []model.SystemRole
	err := s.db.WithContext(ctx).
		Table("system_user_roles sur").
		Joins("JOIN system_roles sr ON sr.id = sur.role_id AND sr.deleted_time IS NULL").
		Where("sur.user_id = ?", userID).
		Scan(&roles).Error
	return roles, err
}

// ────────────────────────────────────────────────────────────
// 权限元数据（供前端渲染权限树）
// ────────────────────────────────────────────────────────────

// GetPermissionsMeta 返回完整权限 Key 清单
func (s *RBACService) GetPermissionsMeta() map[string]interface{} {
	return map[string]interface{}{
		"system": map[string]interface{}{
			"users":   map[string]string{"read": "查看用户", "write": "管理用户", "delete": "删除用户"},
			"sites":   map[string]string{"read": "查看站点", "write": "管理站点", "delete": "删除站点"},
			"tokens":  map[string]string{"read": "查看 Token", "write": "管理 Token", "delete": "删除 Token"},
			"settings": map[string]string{"read": "查看设置", "write": "修改设置"},
			"audit":    map[string]string{"read": "查看审计日志", "export": "导出审计日志"},
			"roles":    map[string]string{"read": "查看角色", "write": "管理角色", "delete": "删除角色"},
			"dashboard": map[string]string{"read": "查看仪表盘"},
		},
		"content_schema": map[string]string{"read": "查看模型", "write": "管理模型", "delete": "删除模型"},
		"entry":          map[string]string{"read": "查看条目", "write": "编辑条目", "publish": "发布", "delete": "删除条目"},
		"asset":          map[string]string{"read": "查看文件", "write": "管理文件", "delete": "删除文件"},
		"media":          map[string]string{"read": "查看媒体库", "write": "管理媒体", "delete": "删除媒体"},
		"site":           map[string]string{"read": "查看站点", "write": "站点设置", "delete": "删除站点"},
		"api_token":      map[string]string{"read": "查看 Token", "write": "管理 Token", "delete": "删除 Token"},
	}
}
