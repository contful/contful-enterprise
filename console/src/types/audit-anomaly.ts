// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import type { AnomalyType, AnomalySeverity } from './audit'

export interface AuditAnomaly {
  id: string
  audit_log_id?: string
  anomaly_type: AnomalyType
  severity: AnomalySeverity
  score: number
  baseline_value?: Record<string, any>
  actual_value?: Record<string, any>
  description: string
  detected_time: string
  resolved_time?: string
  resolution_note?: string
  created_time: string
}

export interface AnomalySummary {
  total_anomalies: number
  by_type: Record<string, number>
  by_severity: Record<string, number>
  trend: Array<{ date: string; count: number }>
  latest_anomalies: AuditAnomaly[]
  last_scan_time: string
}

export interface AnomalyListResponse {
  items: AuditAnomaly[]
  total: number
  page: number
  page_size: number
}
