<template>
  <div class="anomaly-tab">
    <div class="anomaly-header">
      <t-button theme="primary" :loading="scanning" @click="handleScan">{{ scanning ? t('audit.anomaly.scanning') : t('audit.anomaly.manualScan') }}</t-button>
    </div>
    <AnomalySummary :summary="summary" :loading="summaryLoading" />
    <AnomalyTrend :trend="summary?.trend || []" :loading="summaryLoading" />
    <AnomalyList />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { getAnomalySummary, triggerScan } from '@/api/audit-anomaly'
import type { AnomalySummary as AnomalySummaryType } from '@/types/audit-anomaly'
import AnomalySummary from './AnomalySummary.vue'
import AnomalyTrend from './AnomalyTrend.vue'
import AnomalyList from './AnomalyList.vue'
const { t } = useI18n()
const summary = ref<AnomalySummaryType|null>(null)
const summaryLoading = ref(false); const scanning = ref(false)
async function fetchSummary(){ summaryLoading.value=true; try{ const res = await getAnomalySummary(); summary.value = res.data||null } catch(e:unknown){ MessagePlugin.error(e instanceof Error?e.message:String(e)) } finally { summaryLoading.value=false } }
async function handleScan(){ scanning.value=true; try{ await triggerScan(); MessagePlugin.success(t('audit.anomaly.scanSuccess')); await fetchSummary() } catch(e:unknown){ MessagePlugin.error(e instanceof Error?e.message:String(e)) } finally { scanning.value=false } }
onMounted(fetchSummary)
</script>

<style scoped>
.anomaly-tab { padding-top: 8px; }
.anomaly-header { display: flex; justify-content: flex-end; margin-bottom: 16px; }
</style>
