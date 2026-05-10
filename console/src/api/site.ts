// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post, put, del } from '@/utils/request'

// 站点配置类型（对应后端 settings JSONB）
export interface SiteConfig {
  [key: string]: any
}

// 站点（混合模式：固定列 + JSONB 动态配置）
export interface Site {
  id: string
  name: string
  slug: string
  description?: string
  site_url?: string

  // 混合模式：固定列
  locale?: string
  timezone?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  
  // 动态配置（JSONB）
  settings?: SiteConfig

  is_active: boolean
  created_by?: string
  created_time: string
  updated_time: string
}

export interface SiteListResponse {
  items: Site[]
  total: number
  page: number
  page_size: number
}

// 创建站点参数
export interface CreateSiteParams {
  name: string
  slug: string
  description?: string
  site_url?: string
  locale?: string
  timezone?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  settings?: Record<string, any>
}

// 更新站点参数
export interface UpdateSiteParams {
  name?: string
  slug?: string
  description?: string
  site_url?: string
  locale?: string
  timezone?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  settings?: Record<string, any>
  is_active?: boolean
}

// 获取当前用户所属站点
export function getMySites(params?: { page?: number; page_size?: number }) {
  return get<SiteListResponse>('/sites/mine', { params })
}

// 创建站点
export function createSite(data: CreateSiteParams) {
  return post<Site>('/sites', data)
}

// 获取站点详情
export function getSite(id: string) {
  return get<Site>(`/sites/${id}`)
}

// 更新站点
export function updateSite(id: string, data: UpdateSiteParams) {
  return put<Site>(`/sites/${id}`, data)
}

// 删除站点
export function deleteSite(id: string) {
  return del(`/sites/${id}`)
}
