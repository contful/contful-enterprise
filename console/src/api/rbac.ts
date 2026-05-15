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
// 权限分组与权限项管理
// ─────────────────────────────────────────────────────────────

export interface PermissionItem {
  id: string
  action: string
  label: string
  label_en?: string
  sort_order: number
}

export interface PermissionGroup {
  id: string
  group_key: string
  label: string
  label_en?: string
  sort_order: number
  permissions: PermissionItem[]
}

/** 获取权限分组及权限项列表 */
export function listPermissions() {
  return get<PermissionGroup[]>('/system/permissions')
}

/** 创建权限分组 */
export function createPermissionGroup(data: { group_key: string; label: string; label_en?: string; sort_order?: number }) {
  return post('/system/permissions/group', data)
}

/** 更新权限分组 */
export function updatePermissionGroup(id: string, data: { label?: string; label_en?: string; sort_order?: number }) {
  return put(`/system/permissions/group/${id}`, data)
}

/** 删除权限分组 */
export function deletePermissionGroup(id: string) {
  return del(`/system/permissions/group/${id}`)
}

/** 创建权限项 */
export function createPermission(data: { group_id: string; action: string; label: string; label_en?: string; sort_order?: number }) {
  return post('/system/permissions', data)
}

/** 更新权限项 */
export function updatePermission(id: string, data: { label?: string; label_en?: string; sort_order?: number }) {
  return put(`/system/permissions/${id}`, data)
}

/** 删除权限项 */
export function deletePermission(id: string) {
  return del(`/system/permissions/${id}`)
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
