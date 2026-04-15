import request from '@/utils/request'

// ============ ContentType API ============

export interface ContentType {
  id: string
  site_id: string
  name: string
  slug: string
  description: string
  kind: 'collection' | 'single'
  display_config: Record<string, any>
  api_config: {
    publicRead: boolean
    publicWrite: boolean
  }
  preview_config: Record<string, any>
  versioning_enabled: boolean
  draft_autosave_interval: number | null
  is_active: boolean
  sort_order: number
  created_by: string | null
  created_at: string
  updated_at: string
  fields?: Field[]
}

export interface Field {
  id: string
  content_type_id: string
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
  created_at: string
  updated_at: string
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

export const FIELD_TYPES: Record<FieldType, { label: string; icon: string }> = {
  text: { label: '单行文本', icon: 'text' },
  rich_text: { label: '富文本', icon: 'richtext' },
  number: { label: '数字', icon: 'number' },
  boolean: { label: '布尔值', icon: 'boolean' },
  date: { label: '日期', icon: 'calendar' },
  datetime: { label: '日期时间', icon: 'calendar' },
  email: { label: '邮箱', icon: 'mail' },
  url: { label: 'URL', icon: 'link' },
  json: { label: 'JSON', icon: 'code' },
  media: { label: '媒体', icon: 'image' },
  relation: { label: '关联', icon: 'relation' },
  enum: { label: '枚举', icon: 'list' },
  password: { label: '密码', icon: 'lock' },
}

export interface ContentTypeCreate {
  name: string
  slug: string
  description?: string
  kind: 'collection' | 'single'
  display_config?: Record<string, any>
  api_config?: { publicRead?: boolean; publicWrite?: boolean }
  preview_config?: Record<string, any>
  versioning_enabled?: boolean
  draft_autosave_interval?: number
  sort_order?: number
}

export interface ContentTypeUpdate {
  name?: string
  slug?: string
  description?: string
  kind?: 'collection' | 'single'
  display_config?: Record<string, any>
  api_config?: { publicRead?: boolean; publicWrite?: boolean }
  preview_config?: Record<string, any>
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

// 创建内容类型
export function createContentType(data: ContentTypeCreate) {
  return request.post<{ code: number; data: ContentType }>('/content/types', data)
}

// 获取内容类型列表
export function getContentTypes(params?: { page?: number; page_size?: number }) {
  return request.get<{ code: number; data: ListResponse<ContentType> }>('/content/types', { params })
}

// 获取内容类型详情
export function getContentType(id: string) {
  return request.get<{ code: number; data: ContentType }>(`/content/types/${id}`)
}

// 更新内容类型
export function updateContentType(id: string, data: ContentTypeUpdate) {
  return request.put<{ code: number; data: ContentType }>(`/content/types/${id}`, data)
}

// 删除内容类型
export function deleteContentType(id: string) {
  return request.delete(`/content/types/${id}`)
}

// ============ Field API ============

// 创建字段
export function createField(contentTypeId: string, data: FieldCreate) {
  return request.post<{ code: number; data: Field }>(`/content/types/${contentTypeId}/fields`, data)
}

// 获取字段列表
export function getFields(contentTypeId: string) {
  return request.get<{ code: number; data: { items: Field[] } }>(`/content/types/${contentTypeId}/fields`)
}

// 更新字段
export function updateField(fieldId: string, data: FieldUpdate) {
  return request.put<{ code: number; data: Field }>(`/content/types/fields/${fieldId}`, data)
}

// 删除字段
export function deleteField(fieldId: string) {
  return request.delete(`/content/types/fields/${fieldId}`)
}

// 重新排序字段
export function reorderFields(contentTypeId: string, orders: Record<string, number>) {
  return request.post(`/content/types/${contentTypeId}/fields/reorder`, { orders })
}
