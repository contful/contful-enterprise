import request from '@/utils/request'

export interface Site {
  id: string
  name: string
  slug: string
  description?: string
  logo_url?: string
  favicon_url?: string
  config: Record<string, any>
  seo: Record<string, any>
  custom_domains: string[]
  is_active: boolean
  plan: string
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

// SEO 配置（seo JSONB）
export interface SiteSEO {
  meta_title?: string
  meta_description?: string
  keywords?: string
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
  seo?: SiteSEO
  custom_domains?: string[]
  is_active?: boolean
  plan?: string
}

// 获取当前用户所属站点
export function getMySites(params?: { page?: number; page_size?: number }) {
  return request.get<SiteListResponse>('/sites/mine', { params })
}

// 创建站点
export function createSite(data: CreateSiteParams) {
  return request.post<{ items: Site[]; total: number }>('/sites', data)
}

// 获取站点详情
export function getSite(id: string) {
  return request.get<Site>(`/sites/${id}`)
}

// 更新站点
export function updateSite(id: string, data: UpdateSiteParams) {
  return request.put<Site>(`/sites/${id}`, data)
}

// 删除站点
export function deleteSite(id: string) {
  return request.delete(`/sites/${id}`)
}
