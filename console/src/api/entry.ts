// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post, put, del } from '@/utils/request'

// ============ Entry API ============

export interface Entry {
  id: string
  schema_id: string
  site_id: string
  locale: string
  status: 'draft' | 'published' | 'archived'
  version: number
  version_history?: EntryVersion[]
  published_time?: string
  published_by?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  sort_weight: number
  created_by?: string
  created_time: string
  updated_time: string
  values?: Record<string, any>
}

export interface EntryVersion {
  version: number
  created_by?: string
  created_time: string
  change_summary?: string
}

export interface EntryCreate {
  schema_id: string
  locale?: string
  values?: Record<string, any>
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  sort_weight?: number
}

export interface EntryUpdate {
  locale?: string
  status?: 'draft' | 'published' | 'archived'
  values?: Record<string, any>
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  sort_weight?: number
  change_summary?: string
}

export interface EntryPublish {
  change_summary?: string
}

export interface EntryListFilter {
  schema_id?: string
  status?: 'draft' | 'published' | 'archived'
  locale?: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

// 获取条目列表
export function getEntries(params?: EntryListFilter) {
  return get<ListResponse<Entry>>('/content/entries', { params })
}

// 获取条目详情
export function getEntry(id: string) {
  return get<Entry>(`/content/entries/${id}`)
}

// 创建条目
export function createEntry(data: EntryCreate) {
  return post<Entry>('/content/entries', data)
}

// 更新条目
export function updateEntry(id: string, data: EntryUpdate) {
  return put<Entry>(`/content/entries/${id}`, data)
}

// 删除条目
export function deleteEntry(id: string) {
  return del(`/content/entries/${id}`)
}

// 发布条目
export function publishEntry(id: string, data?: EntryPublish) {
  return post<Entry>(`/content/entries/${id}/publish`, data || {})
}

// 取消发布
export function unpublishEntry(id: string) {
  return post<Entry>(`/content/entries/${id}/unpublish`)
}

// 获取版本历史
export function getEntryVersions(id: string) {
  return get<EntryVersion[]>(`/content/entries/${id}/versions`)
}

// ============ 缓存 API ============

export interface CacheInvalidateResponse {
  message: string
  deleted: number
}

// 清除内容缓存
export function invalidateCache() {
  return post<CacheInvalidateResponse>('/cache/invalidate')
}

// 批量删除
export function batchDeleteEntries(ids: string[]) {
  return post('/content/entries/batch-delete', { ids })
}

// 批量发布
export function batchPublishEntries(ids: string[]) {
  return post('/content/entries/batch-publish', { ids })
}

// 批量取消发布
export function batchUnpublishEntries(ids: string[]) {
  return post('/content/entries/batch-unpublish', { ids })
}

// 批量移动到归档
export function batchArchiveEntries(ids: string[]) {
  return post('/content/entries/batch-archive', { ids })
}
