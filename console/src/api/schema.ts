// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post, put, del } from '@/utils/request'

// ============ ContentSchema API ============

export interface ContentSchema {
  id: string
  site_id: string
  name: string
  slug: string
  description: string
  kind: 'collection' | 'single'
  versioning_enabled: boolean
  draft_autosave_interval: number | null
  is_active: boolean
  sort_order: number
  created_by: string | null
  created_time: string
  updated_time: string
  fields?: Field[]
}

export interface Field {
  id: string
  schema_id: string
  name: string
  label: string
  description: string
  field_type: FieldType
  config: Record<string, any>
  validation: Record<string, any>
  display: Record<string, any>
  default_value: any
  sort_order: number
  conditional_display: any
  created_time: string
  updated_time: string
}

export type FieldType =
  | 'text'
  | 'rich_text'
  | 'number'
  | 'boolean'
  | 'date'
  | 'datetime'
  | 'email'
  | 'url'
  | 'json'
  | 'media'
  | 'relation'
  | 'enum'
  | 'password'

export const FIELD_TYPES: Record<FieldType, { labelKey: string; icon: string }> = {
  text: { labelKey: 'fieldTypes.text', icon: 'text' },
  rich_text: { labelKey: 'fieldTypes.richText', icon: 'richtext' },
  number: { labelKey: 'fieldTypes.number', icon: 'number' },
  boolean: { labelKey: 'fieldTypes.boolean', icon: 'boolean' },
  date: { labelKey: 'fieldTypes.date', icon: 'calendar' },
  datetime: { labelKey: 'fieldTypes.dateTime', icon: 'calendar' },
  email: { labelKey: 'fieldTypes.email', icon: 'mail' },
  url: { labelKey: 'fieldTypes.url', icon: 'link' },
  json: { labelKey: 'fieldTypes.json', icon: 'code' },
  media: { labelKey: 'fieldTypes.media', icon: 'image' },
  relation: { labelKey: 'fieldTypes.relation', icon: 'relation' },
  enum: { labelKey: 'fieldTypes.enum', icon: 'list' },
  password: { labelKey: 'fieldTypes.password', icon: 'lock' },
}

export interface ContentSchemaCreate {
  name: string
  slug: string
  description?: string
  kind: 'collection' | 'single'
  versioning_enabled?: boolean
  draft_autosave_interval?: number
  sort_order?: number
}

export interface ContentSchemaUpdate {
  name?: string
  slug?: string
  description?: string
  kind?: 'collection' | 'single'
  versioning_enabled?: boolean
  draft_autosave_interval?: number
  is_active?: boolean
  sort_order?: number
}

export interface FieldCreate {
  name: string
  label: string
  description?: string
  field_type: FieldType
  config?: Record<string, any>
  validation?: Record<string, any>
  display?: Record<string, any>
  default_value?: any
  sort_order?: number
  conditional_display?: Record<string, any>
}

export interface FieldUpdate {
  name?: string
  label?: string
  description?: string
  field_type?: FieldType
  config?: Record<string, any>
  validation?: Record<string, any>
  display?: Record<string, any>
  default_value?: any
  sort_order?: number
  conditional_display?: Record<string, any>
}

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

// 创建内容模型
export function createContentSchema(data: ContentSchemaCreate) {
  return post<ContentSchema>('/content/schemas', data)
}

// 获取内容模型列表
export function getContentSchemas(params?: { page?: number; page_size?: number }) {
  return get<ListResponse<ContentSchema>>('/content/schemas', { params })
}

// 获取内容模型详情
export function getContentSchema(id: string) {
  return get<ContentSchema>(`/content/schemas/${id}`)
}

// 更新内容模型
export function updateContentSchema(id: string, data: ContentSchemaUpdate) {
  return put<ContentSchema>(`/content/schemas/${id}`, data)
}

// 删除内容模型
export function deleteContentSchema(id: string) {
  return del(`/content/schemas/${id}`)
}

// ============ Field API ============

// 创建字段
export function createField(schemaId: string, data: FieldCreate) {
  return post<Field>(`/content/schemas/${schemaId}/fields`, data)
}

// 获取字段列表
export function getFields(schemaId: string) {
  return get<{ items: Field[] }>(`/content/schemas/${schemaId}/fields`)
}

// 更新字段
export function updateField(schemaId: string, fieldId: string, data: FieldUpdate) {
  return put<Field>(`/content/schemas/${schemaId}/fields/${fieldId}`, data)
}

// 删除字段
export function deleteField(schemaId: string, fieldId: string) {
  return del(`/content/schemas/${schemaId}/fields/${fieldId}`)
}

// 重新排序字段
export function reorderFields(schemaId: string, orders: Record<string, number>) {
  return post(`/content/schemas/${schemaId}/fields/reorder`, { orders })
}
