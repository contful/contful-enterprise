// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post, put, del, patch } from '@/utils/request'

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

export interface SiteRole {
  id: string
  site_id: string
  name: string
  description: string
  is_system: boolean
  permissions: string[]
  content_permissions: string[]
  sort_order: number
  created_time: string
  updated_time: string
}

export interface SiteMember {
  id: string
  user_id: string
  email: string
  nickname: string
  avatar_url?: string
  role_id: string
  role_name: string
  status: 'active' | 'inactive'
  joined_at: string
}

export interface SiteMemberListResponse {
  items: SiteMember[]
  total: number
  page: number
  page_size: number
}

export interface PermissionsMeta {
  system: Record<string, Record<string, string>>
  content_schema: Record<string, string>
  entry: Record<string, string>
  asset: Record<string, string>
  media: Record<string, string>
  site: Record<string, string>
  api_token: Record<string, string>
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
// 站点角色管理
// ─────────────────────────────────────────────────────────────

/** 获取站点角色列表 */
export function listSiteRoles(siteId: string) {
  return get<SiteRole[]>(`/sites/${siteId}/roles`)
}

/** 获取站点角色详情 */
export function getSiteRole(siteId: string, id: string) {
  return get<SiteRole>(`/sites/${siteId}/roles/${id}`)
}

/** 创建自定义站点角色 */
export function createSiteRole(
  siteId: string,
  data: {
    name: string
    description?: string
    permissions?: string[]
    content_permissions?: string[]
    channel_permissions?: string[]
    sort_order?: number
  },
) {
  return post<SiteRole>(`/sites/${siteId}/roles`, data)
}

/** 更新站点角色 */
export function updateSiteRole(
  siteId: string,
  id: string,
  data: {
    name?: string
    description?: string
    permissions?: string[]
    content_permissions?: string[]
    channel_permissions?: string[]
  },
) {
  return put<SiteRole>(`/sites/${siteId}/roles/${id}`, data)
}

/** 删除站点角色 */
export function deleteSiteRole(siteId: string, id: string) {
  return del(`/sites/${siteId}/roles/${id}`)
}

/** 获取站点级权限树元数据 */
export function getSitePermissions(siteId: string) {
  return get(`/sites/${siteId}/roles/permissions`)
}

// ─────────────────────────────────────────────────────────────
// 站点成员管理
// ─────────────────────────────────────────────────────────────

/** 获取站点成员列表（分页） */
export function listSiteMembers(siteId: string, params?: { page?: number; page_size?: number }) {
  return get<SiteMemberListResponse>(`/sites/${siteId}/members`, { params })
}

/** 邀请用户加入站点 */
export function addSiteMember(siteId: string, data: { email: string; role_id: string }) {
  return post<SiteMember>(`/sites/${siteId}/members`, data)
}

/** 更换成员角色 */
export function updateSiteMemberRole(siteId: string, userId: string, roleId: string) {
  return put(`/sites/${siteId}/members/${userId}`, { role_id: roleId })
}

/** 更新成员状态（active/inactive） */
export function updateSiteMemberStatus(siteId: string, userId: string, status: 'active' | 'inactive') {
  return patch(`/sites/${siteId}/members/${userId}/status`, { status })
}

/** 移除站点成员 */
export function removeSiteMember(siteId: string, userId: string) {
  return del(`/sites/${siteId}/members/${userId}`)
}

// ─────────────────────────────────────────────────────────────
// 权限元数据
// ─────────────────────────────────────────────────────────────

/** 获取完整权限元数据（供前端渲染权限树） */
export function getPermissionsMeta() {
  return get<PermissionsMeta>('/permissions')
}
