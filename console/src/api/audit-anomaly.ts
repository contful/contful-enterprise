// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { get, post } from '@/utils/request'
import type { AuditAnomaly, AnomalySummary, AnomalyListResponse } from '@/types/audit-anomaly'

export function getAnomalies(params?: Record<string, any>) {
  return get<AnomalyListResponse>('/audit/anomalies', { params })
}

export function getAnomaly(id: string) {
  return get<AuditAnomaly>(`/audit/anomalies/${id}`)
}

export function getAnomalySummary() {
  return get<AnomalySummary>('/audit/anomalies/summary')
}

export function triggerScan() {
  return post<{ message: string }>('/audit/anomalies/scan')
}
