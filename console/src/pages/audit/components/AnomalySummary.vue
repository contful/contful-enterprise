<template>
  <t-row :gutter="16" class="anomaly-summary">
    <t-col :span="6">
      <t-card :loading="loading" class="summary-card">
        <div class="summary-number">{{ summary?.total_anomalies ?? 0 }}</div>
        <div class="summary-label">{{ t('audit.anomaly.totalEvents') }}</div>
      </t-card>
    </t-col>
    <t-col :span="6">
      <t-card :loading="loading">
        <div class="section-title">{{ t('audit.anomaly.bySeverity') }}</div>
        <div class="severity-tags">
          <t-tag theme="success" variant="light">{{ t('audit.anomaly.severity.low') }} {{ summary?.by_severity?.low ?? 0 }}</t-tag>
          <t-tag theme="warning" variant="light">{{ t('audit.anomaly.severity.medium') }} {{ summary?.by_severity?.medium ?? 0 }}</t-tag>
          <t-tag theme="danger" variant="light">{{ t('audit.anomaly.severity.high') }} {{ summary?.by_severity?.high ?? 0 }}</t-tag>
          <t-tag theme="danger" variant="light">{{ t('audit.anomaly.severity.critical') }} {{ summary?.by_severity?.critical ?? 0 }}</t-tag>
        </div>
      </t-card>
    </t-col>
    <t-col :span="6">
      <t-card :loading="loading">
        <div class="section-title">{{ t('audit.anomaly.byType') }}</div>
        <div class="type-list">
          <div v-if="summary" v-for="(count, key) in summary.by_type" :key="key" class="type-item">
            <span>{{ t('audit.anomaly.type.' + key) }}</span><span class="type-count">{{ count }}</span>
          </div>
          <div v-else class="no-data">{{ t('audit.anomaly.noAnomalies') }}</div>
        </div>
      </t-card>
    </t-col>
    <t-col :span="6">
      <t-card :loading="loading">
        <div class="section-title">{{ t('audit.anomaly.lastScan') }}</div>
        <div class="scan-time">{{ summary?.last_scan_time ? new Date(summary.last_scan_time).toLocaleString() : '-' }}</div>
      </t-card>
    </t-col>
  </t-row>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AnomalySummary } from '@/types/audit-anomaly'
const { t } = useI18n()
defineProps<{ summary: AnomalySummary | null; loading: boolean }>()
</script>

<style scoped>
.anomaly-summary { margin-bottom: 16px; }
.summary-card { text-align: center; }
.summary-number { font-size: 36px; font-weight: 700; color: var(--td-brand-color); }
.summary-label { font-size: 13px; color: var(--td-text-color-secondary); margin-top: 4px; }
.section-title { font-size: 13px; color: var(--td-text-color-secondary); margin-bottom: 10px; }
.severity-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.type-list { display: flex; flex-direction: column; gap: 6px; }
.type-item { display: flex; justify-content: space-between; font-size: 13px; }
.type-count { font-weight: 500; color: var(--td-brand-color); }
.scan-time { font-size: 14px; }
.no-data { font-size: 13px; color: var(--td-text-color-placeholder); }
</style>
