import request from '@/utils/request'

// ============ Entry API ============

export interface Entry {
  id: string
  content_type_id: string
  site_id: string
  locale: string
  status: 'draft' | 'published' | 'archived'
  version: number
  version_history?: EntryVersion[]
  published_at?: string
  published_by?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string[]
  sort_weight: number
  created_by?: string
  created_at: string
  updated_at: string
  values?: Record<string, any>
}

export interface EntryVersion {
  version: number
  created_by?: string
  created_at: string
  change_summary?: string
}

export interface EntryCreate {
  content_type_id: string
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
  content_type_id?: string
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
  return request.get<{ code: number; data: ListResponse<Entry> }>('/content/entries', { params })
}

// 获取条目详情
export function getEntry(id: string) {
  return request.get<{ code: number; data: Entry }>(`/content/entries/${id}`)
}

// 创建条目
export function createEntry(data: EntryCreate) {
  return request.post<{ code: number; data: Entry }>('/content/entries', data)
}

// 更新条目
export function updateEntry(id: string, data: EntryUpdate) {
  return request.put<{ code: number; data: Entry }>(`/content/entries/${id}`, data)
}

// 删除条目
export function deleteEntry(id: string) {
  return request.delete(`/content/entries/${id}`)
}

// 发布条目
export function publishEntry(id: string, data?: EntryPublish) {
  return request.post<{ code: number; data: Entry }>(`/content/entries/${id}/publish`, data || {})
}

// 取消发布
export function unpublishEntry(id: string) {
  return request.post<{ code: number; data: Entry }>(`/content/entries/${id}/unpublish`)
}

// 获取版本历史
export function getEntryVersions(id: string) {
  return request.get<{ code: number; data: EntryVersion[] }>(`/content/entries/${id}/versions`)
}
