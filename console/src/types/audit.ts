// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

export type AuditLevel = 'debug' | 'info' | 'warn' | 'error'

export type AuditType = 'auth' | 'content' | 'media' | 'settings' | 'user' | 'system'

export interface AuditLog {
  id: string
  site_id?: string
  user_id?: string
  action: string
  resource_type?: string
  resource_id?: string
  level: AuditLevel
  category: AuditType
  details?: string
  ip_address?: string
  user_agent?: string
  created_time: string
  data_signature?: Record<string, any>
}

export interface AuditLogListResponse {
  items: AuditLog[]
  total: number
  page: number
  page_size: number
}
