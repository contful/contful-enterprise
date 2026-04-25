import { get, post, put, del } from '@/utils/request'

export interface Site {
  id: string
  name: string
  slug: string
  description?: string
  logo_url?: string
  favicon_url?: string
  config: Record<string, any>
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

export interface CreateSiteParams {
  name: string
  slug: string
  description?: string
}

// 站点配置（config JSONB）
export interface SiteConfig {
  timezone?: string
  locale?: string
  [key: string]: any
}

// 更新站点参数（对应后端 SiteUpdate）
export interface UpdateSiteParams {
  name?: string
  slug?: string
  description?: string
  logo_url?: string
  favicon_url?: string
  config?: SiteConfig
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
