import request from '@/utils/request'

// ============ Types ============

export type TokenStatus = 'active' | 'expired' | 'revoked'

export interface APIEndpoint {
  path: string
  method: string[]
}

export interface EndpointPermission {
  content_types?: string[]
  endpoints?: APIEndpoint[]
}

export interface APIEndpointLimits {
  requests_per_minute: number
  requests_per_day: number
}

export interface APIUsage {
  request_count: number
  daily_request_count: number
  bandwidth_used: number
  last_request_at?: string
}

export interface APITokenResponse {
  id: string
  site_id: string
  name: string
  description?: string
  token_prefix: string
  permissions: EndpointPermission
  rate_limits: APIEndpointLimits
  usage: APIUsage
  expires_at?: string
  status: TokenStatus
  last_used_at?: string
  created_by?: string
  created_at: string
  updated_at: string
}

export interface APITokenCreateResponse extends APITokenResponse {
  token: string // 仅在创建时返回
}

export interface APITokenCreate {
  name: string
  description?: string
  permissions?: EndpointPermission
  rate_limits?: APIEndpointLimits
  expires_at?: string
}

export interface APITokenUpdate {
  name?: string
  description?: string
  permissions?: EndpointPermission
  rate_limits?: APIEndpointLimits
  expires_at?: string
  status?: TokenStatus
}

export interface APITokenListFilter {
  status?: TokenStatus
  name?: string
}

export interface APITokenListResponse {
  items: APITokenResponse[]
  total: number
  page: number
  page_size: number
}

// ============ API ============

export const apiTokenApi = {
  /**
   * 创建 API Token
   * 注意: 返回的 token 仅在此接口返回一次，之后无法查看
   */
  create: async (data: APITokenCreate): Promise<APITokenCreateResponse> => {
    const response = await request.post<{ data: APITokenCreateResponse }>(
      '/admin/v1/api-tokens',
      data
    )
    return response.data.data
  },

  /**
   * 获取 Token 列表
   */
  list: async (params?: {
    page?: number
    page_size?: number
    filter?: APITokenListFilter
  }): Promise<APITokenListResponse> => {
    const { page = 1, page_size = 20, filter } = params || {}

    const queryParams = new URLSearchParams()
    queryParams.append('page', String(page))
    queryParams.append('page_size', String(page_size))

    if (filter) {
      if (filter.status) queryParams.append('status', filter.status)
      if (filter.name) queryParams.append('name', filter.name)
    }

    const response = await request.get<{ data: APITokenListResponse }>(
      `/admin/v1/api-tokens?${queryParams.toString()}`
    )

    return response.data.data
  },

  /**
   * 获取 Token 详情
   */
  get: async (id: string): Promise<APITokenResponse> => {
    const response = await request.get<{ data: APITokenResponse }>(
      `/admin/v1/api-tokens/${id}`
    )
    return response.data.data
  },

  /**
   * 更新 Token
   */
  update: async (id: string, data: APITokenUpdate): Promise<APITokenResponse> => {
    const response = await request.put<{ data: APITokenResponse }>(
      `/admin/v1/api-tokens/${id}`,
      data
    )
    return response.data.data
  },

  /**
   * 删除 Token
   */
  delete: async (id: string): Promise<void> => {
    await request.delete(`/admin/v1/api-tokens/${id}`)
  },

  /**
   * 重新生成 Token
   * 注意: 返回的新 token 仅在此接口返回一次
   */
  regenerate: async (id: string): Promise<APITokenCreateResponse> => {
    const response = await request.post<{ data: APITokenCreateResponse }>(
      `/admin/v1/api-tokens/${id}/regenerate`
    )
    return response.data.data
  },

  /**
   * 撤销 Token
   */
  revoke: async (id: string): Promise<void> => {
    await request.post(`/admin/v1/api-tokens/${id}/revoke`)
  },
}

// 便捷导出
export default apiTokenApi

// ============ 便捷函数导出 ============
export const getApiTokens = (params?: { page?: number; page_size?: number; status?: string; name?: string }) => {
  return apiTokenApi.list({ page: params?.page, page_size: params?.page_size, filter: params })
}

export const getApiToken = (id: string) => {
  return apiTokenApi.get(id)
}

export const createApiToken = (data: {
  name: string
  description?: string
  expires_in_days?: number
  permissions?: string[]
  rate_limit?: number
}) => {
  return apiTokenApi.create({
    name: data.name,
    description: data.description,
    expires_at: data.expires_in_days ? new Date(Date.now() + data.expires_in_days * 24 * 60 * 60 * 1000).toISOString() : undefined,
    permissions: data.permissions ? { content_types: data.permissions } : undefined,
    rate_limits: data.rate_limit ? { requests_per_minute: Math.floor(data.rate_limit / 60), requests_per_day: data.rate_limit } : undefined,
  })
}

export const updateApiToken = (id: string, data: {
  name?: string
  description?: string
  permissions?: string[]
  rate_limit?: number
}) => {
  return apiTokenApi.update(id, {
    name: data.name,
    description: data.description,
    permissions: data.permissions ? { content_types: data.permissions } : undefined,
    rate_limits: data.rate_limit ? { requests_per_minute: Math.floor(data.rate_limit / 60), requests_per_day: data.rate_limit } : undefined,
  })
}

export const deleteApiToken = (id: string) => {
  return apiTokenApi.delete(id)
}

export const regenerateApiToken = (id: string) => {
  return apiTokenApi.regenerate(id)
}

export const revokeApiToken = (id: string) => {
  return apiTokenApi.revoke(id)
}

// 类型导出
export type { APITokenResponse as ApiToken } from './api-token'
