// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

export interface AuditQueryTemplate {
  id: string
  name: string
  conditions: AuditTemplateConditions
  created_by?: string
  created_time: string
}

export interface AuditTemplateConditions {
  keyword?: string
  combinator?: string
  time_preset?: string
  category?: string
  level?: string
  action?: string
  start_time?: string
  end_time?: string
}
