// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

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
