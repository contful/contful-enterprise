// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import request from '@/utils/request'

export interface SiteConfig {
  id: string
  site_id: string
  config_key: string
  config_value: string
  config_type: 'string' | 'number' | 'boolean' | 'json' | 'encrypted'
  config_group: string
  is_encrypted: boolean
  is_readonly: boolean
  description?: string
  created_time: string
  updated_time: string
  updated_by?: string
}

export interface ConfigListResponse {
  items: SiteConfig[]
  total: number
}

// 按分组列出配置
export function getConfigs(siteId: string): Promise<ConfigListResponse> {
  return request.get(`/sites/${siteId}/configs`)
}

// 读取单个配置
export function getConfig(siteId: string, key: string): Promise<{ data: SiteConfig }> {
  return request.get(`/sites/${siteId}/configs/${key}`)
}

// 设置配置（创建或更新）
export function setConfig(siteId: string, key: string, data: {
  config_value: string
  config_type?: string
  config_group?: string
  description?: string
}): Promise<{ data: SiteConfig }> {
  return request.put(`/sites/${siteId}/configs/${key}`, data)
}

// 删除配置
export function deleteConfig(siteId: string, key: string): Promise<void> {
  return request.delete(`/sites/${siteId}/configs/${key}`)
}
