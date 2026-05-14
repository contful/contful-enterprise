// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import request from '@/utils/request'

// ============ Types ============

export type TokenStatus = 'active' | 'expired' | 'revoked'

export interface APITokenResponse {
  id: string
  site_id: string
  name: string
  description?: string
  token_prefix: string
  expires_time?: string
  status: TokenStatus
  last_used_time?: string
  last_used_ip?: string
  created_by?: string
  created_time: string
  updated_time: string
}

export interface APITokenCreateResponse extends APITokenResponse {
  token: string
}

export interface APITokenCreate {
  name: string
  description?: string
  expires_in_days?: number
  expires_time?: string
}

export interface APITokenUpdate {
  name?: string
  description?: string
  expires_time?: string
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
  create: async (data: APITokenCreate): Promise<APITokenCreateResponse> => {
    const response = await request.post<{ data: APITokenCreateResponse }>(
      '/tokens',
      data
    )
    return response.data.data
  },

  list: async (params?: {
    page?: number
    page_size?: number
    filter?: APITokenListFilter
  }): Promise<APITokenListResponse> => {
    const { page = 1, page_size = 20, filter } = params || {}
    const queryParams = new URLSearchParams()
    queryParams.append('page', String(page))
    queryParams.append('page_size', String(page_size))
    if (filter?.status) queryParams.append('status', filter.status)
    if (filter?.name) queryParams.append('name', filter.name)
    const response = await request.get<{ data: APITokenListResponse }>(
      `/tokens?${queryParams.toString()}`
    )
    return response.data.data
  },

  get: async (id: string): Promise<APITokenResponse> => {
    const response = await request.get<{ data: APITokenResponse }>(`/tokens/${id}`)
    return response.data.data
  },

  update: async (id: string, data: APITokenUpdate): Promise<APITokenResponse> => {
    const response = await request.put<{ data: APITokenResponse }>(
      `/tokens/${id}`,
      data
    )
    return response.data.data
  },

  delete: async (id: string): Promise<void> => {
    await request.delete(`/tokens/${id}`)
  },

  regenerate: async (id: string): Promise<APITokenCreateResponse> => {
    const response = await request.post<{ data: APITokenCreateResponse }>(
      `/tokens/${id}/regenerate`
    )
    return response.data.data
  },

  revoke: async (id: string): Promise<void> => {
    await request.post(`/tokens/${id}/revoke`)
  },

  export: async (id: string): Promise<APITokenCreateResponse> => {
    const response = await request.post<{ data: APITokenCreateResponse }>(
      `/tokens/${id}/export`
    )
    return response.data.data
  },
}

// 便捷导出
export const {
  create: createApiToken,
  list: listApiTokens,
  get: getApiToken,
  update: updateApiToken,
  delete: deleteApiToken,
  regenerate: regenerateApiToken,
  revoke: revokeApiToken,
  export: exportApiToken,
} = {
  create: apiTokenApi.create.bind(apiTokenApi),
  list: apiTokenApi.list.bind(apiTokenApi),
  get: apiTokenApi.get.bind(apiTokenApi),
  update: apiTokenApi.update.bind(apiTokenApi),
  delete: apiTokenApi.delete.bind(apiTokenApi),
  regenerate: apiTokenApi.regenerate.bind(apiTokenApi),
  revoke: apiTokenApi.revoke.bind(apiTokenApi),
  export: apiTokenApi.export.bind(apiTokenApi),
}

export default apiTokenApi
