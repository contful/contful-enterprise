// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post, del } from '@/utils/request'
import type { AuditQueryTemplate } from '@/types/audit-query-template'
import type { AuditLogListResponse } from '@/types/audit'

export function getQueryTemplates() {
  return get<AuditQueryTemplate[]>('/audit/query/templates')
}

export function createQueryTemplate(data: { name: string; conditions: Record<string, any> }) {
  return post<AuditQueryTemplate>('/audit/query/templates', data)
}

export function deleteQueryTemplate(id: string) {
  return del(`/audit/query/templates/${id}`)
}

export function applyQueryTemplate(id: string, params?: Record<string, any>) {
  return get<{ template: AuditQueryTemplate; results: AuditLogListResponse }>(`/audit/query/templates/${id}/apply`, { params })
}
