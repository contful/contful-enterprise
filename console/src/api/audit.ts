// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get } from '@/utils/request'
import type { AuditLog, AuditLogListResponse, AuditLevel, AuditType } from '@/types/audit'

// 审计日志列表查询参数
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
}

// 获取审计日志列表
export function getAuditLogs(params?: AuditLogListParams) {
  return get<AuditLogListResponse>('/audit/logs', { params })
}

// 获取审计日志详情
export function getAuditLog(id: string) {
  return get<AuditLog>(`/audit/logs/${id}`)
}
