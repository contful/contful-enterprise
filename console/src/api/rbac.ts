// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post, put, del } from '@/utils/request'

// ─────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────

export interface SystemRole {
  id: string
  name: string
  description: string
  is_system: boolean
  permissions: string[]
  created_time: string
  updated_time: string
}

export interface PermissionsMeta {
  [group: string]: Record<string, string>
}

// ─────────────────────────────────────────────────────────────
// 系统角色管理
// ─────────────────────────────────────────────────────────────

/** 获取系统角色列表 */
export function listSystemRoles() {
  return get<SystemRole[]>('/system/roles')
}

/** 获取系统角色详情 */
export function getSystemRole(id: string) {
  return get<SystemRole>(`/system/roles/${id}`)
}

/** 创建系统角色 */
export function createSystemRole(data: { name: string; description?: string; permissions: string[] }) {
  return post<SystemRole>('/system/roles', data)
}

/** 更新系统角色 */
export function updateSystemRole(id: string, data: { name?: string; description?: string; permissions?: string[] }) {
  return put<SystemRole>(`/system/roles/${id}`, data)
}

/** 删除系统角色 */
export function deleteSystemRole(id: string) {
  return del(`/system/roles/${id}`)
}

/** 获取系统级权限树元数据 */
export function getSystemPermissions() {
  return get<PermissionsMeta>('/system/roles/permissions')
}

// ─────────────────────────────────────────────────────────────
// 权限元数据
// ─────────────────────────────────────────────────────────────

/** 获取完整权限元数据（供前端渲染权限树） */
export function getPermissionsMeta() {
  return get<PermissionsMeta>('/permissions')
}

// ─────────────────────────────────────────────────────────────
// 用户-角色关联管理
// ─────────────────────────────────────────────────────────────

/** 获取用户的系统角色列表 */
export function getUserRoles(userId: string) {
  return get<SystemRole[]>(`/users/${userId}/roles`)
}

/** 为用户分配角色 */
export function assignUserRole(userId: string, roleId: string) {
  return put(`/users/${userId}/roles/${roleId}`)
}

/** 移除用户的角色 */
export function removeUserRole(userId: string, roleId: string) {
  return del(`/users/${userId}/roles/${roleId}`)
}
