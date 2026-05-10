// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"

	"github.com/google/uuid"
)

// ============================================
// 系统角色 DTO
// ============================================

// SystemRoleCreateRequest 创建系统角色请求
type SystemRoleCreateRequest struct {
	Name        string   `json:"name" binding:"required,min=1,max=100"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// SystemRoleUpdateRequest 更新系统角色请求
type SystemRoleUpdateRequest struct {
	Name        *string  `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string  `json:"description"`
	Permissions []string `json:"permissions"`
}

// ============================================
// 站点角色 DTO
// ============================================

// SiteRoleCreateRequest 创建站点角色请求
type SiteRoleCreateRequest struct {
	Name               string   `json:"name" binding:"required,min=1,max=100"`
	Description        string   `json:"description"`
	Permissions        []string `json:"permissions"`
	ContentPermissions []string `json:"content_permissions"`
	ChannelPermissions []string `json:"channel_permissions"`
	SortOrder          int      `json:"sort_order"`
}

// SiteRoleUpdateRequest 更新站点角色请求
type SiteRoleUpdateRequest struct {
	Name               *string  `json:"name" binding:"omitempty,min=1,max=100"`
	Description        *string  `json:"description"`
	Permissions        []string `json:"permissions"`
	ContentPermissions []string `json:"content_permissions"`
	ChannelPermissions []string `json:"channel_permissions"`
}

// ============================================
// 站点成员 DTO
// ============================================

// AddSiteMemberRequest 邀请成员加入站点请求
type AddSiteMemberRequest struct {
	Email  string    `json:"email" binding:"required,email"`
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

// UpdateSiteMemberRoleRequest 更换成员角色请求
type UpdateSiteMemberRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

// UpdateSiteMemberStatusRequest 更新成员状态请求
type UpdateSiteMemberStatusRequest struct {
	Status UserStatus `json:"status" binding:"required,oneof=active inactive"`
}

// SiteMemberResponse 成员列表条目
type SiteMemberResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Email     string     `json:"email"`
	Nickname  string     `json:"nickname"`
	AvatarURL string     `json:"avatar_url,omitempty"`
	RoleID    uuid.UUID  `json:"role_id"`
	RoleName  string     `json:"role_name"`
	Status    UserStatus `json:"status"`
	JoinedAt  time.Time  `json:"joined_at"`
}

// SiteMemberListResponse 成员列表响应
type SiteMemberListResponse struct {
	Items    []SiteMemberResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}
