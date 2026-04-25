import { get, post, put, del } from '@/utils/request'

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
  created_time: string
  updated_time: string
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
  return post<ContentType>('/content/types', data)
}

// 获取内容类型列表
export function getContentTypes(params?: { page?: number; page_size?: number }) {
  return get<ListResponse<ContentType>>('/content/types', { params })
}

// 获取内容类型详情
export function getContentType(id: string) {
  return get<ContentType>(`/content/types/${id}`)
}

// 更新内容类型
export function updateContentType(id: string, data: ContentTypeUpdate) {
  return put<ContentType>(`/content/types/${id}`, data)
}

// 删除内容类型
export function deleteContentType(id: string) {
  return del(`/content/types/${id}`)
}

// ============ Field API ============

// 创建字段
export function createField(contentTypeId: string, data: FieldCreate) {
  return post<Field>(`/content/types/${contentTypeId}/fields`, data)
}

// 获取字段列表
export function getFields(contentTypeId: string) {
  return get<{ items: Field[] }>(`/content/types/${contentTypeId}/fields`)
}

// 更新字段
export function updateField(contentTypeId: string, fieldId: string, data: FieldUpdate) {
  return put<Field>(`/content/types/${contentTypeId}/fields/${fieldId}`, data)
}

// 删除字段
export function deleteField(contentTypeId: string, fieldId: string) {
  return del(`/content/types/${contentTypeId}/fields/${fieldId}`)
}

// 重新排序字段
export function reorderFields(contentTypeId: string, orders: Record<string, number>) {
  return post(`/content/types/${contentTypeId}/fields/reorder`, { orders })
}
