// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get } from '@/utils/request'
import type { AuditLog, AuditLogListResponse, AuditLevel, AuditType } from '@/types/audit'

export interface AuditLogListParams {
  site_id?: string
  user_id?: string
  action?: string
  resource_type?: string
  category?: AuditType
  level?: AuditLevel
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
  keyword?: string
  combinator?: string
  time_preset?: string
  fields?: string[]
  export_format?: string
}

export function getAuditLogs(params?: AuditLogListParams) {
  return get<AuditLogListResponse>('/audit/logs', { params })
}

export function getAuditLog(id: string) {
  return get<AuditLog>(`/audit/logs/${id}`)
}

export function exportAuditJSON(params?: Record<string, any>) {
  return get('/audit/logs/export/json', { params, responseType: 'blob' })
}
