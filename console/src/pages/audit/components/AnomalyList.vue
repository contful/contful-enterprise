<template>
  <t-card :title="t('audit.anomaly.title')">
    <t-table :data="anomalies" :columns="columns" :loading="loading" :pagination="pagination" row-key="id" @page-change="onPageChange" hover stripe size="medium">
      <template #anomaly_type="{ row }"><t-tag variant="light">{{ t('audit.anomaly.type.'+row.anomaly_type) }}</t-tag></template>
      <template #severity="{ row }"><t-tag :theme="sevTheme(row.severity)" variant="light">{{ t('audit.anomaly.severity.'+row.severity) }}</t-tag></template>
      <template #detected_time="{ row }">{{ row.detected_time ? new Date(row.detected_time).toLocaleString() : '-' }}</template>
      <template #operation="{ row }"><t-button variant="text" theme="primary" @click="handleViewDetail(row)">{{ t('common.view') }}</t-button></template>
    </t-table>
    <t-dialog v-model:visible="detailVisible" :header="t('audit.detailTitle')" :width="700" :footer="false">
      <div class="detail-content" v-if="current">
        <t-list :split="true">
          <t-list-item><t-list-item-meta :title="t('common.id')" :description="current.id"/></t-list-item>
          <t-list-item><t-list-item-meta :title="t('common.type')"><template #description><t-tag variant="light">{{ t('audit.anomaly.type.'+current.anomaly_type) }}</t-tag></template></t-list-item-meta></t-list-item>
          <t-list-item><t-list-item-meta :title="t('audit.anomaly.bySeverity')"><template #description><t-tag :theme="sevTheme(current.severity)" variant="light">{{ t('audit.anomaly.severity.'+current.severity) }}</t-tag></template></t-list-item-meta></t-list-item>
          <t-list-item><t-list-item-meta :title="t('audit.anomaly.score')" :description="String(current.score)"/></t-list-item>
          <t-list-item><t-list-item-meta :title="t('audit.anomaly.description')" :description="current.description||'-'"/></t-list-item>
          <t-list-item><t-list-item-meta :title="t('audit.anomaly.detectedTime')" :description="current.detected_time?new Date(current.detected_time).toLocaleString():'-'"/></t-list-item>
          <t-list-item><t-list-item-meta :title="t('audit.anomaly.resolvedTime')" :description="current.resolved_time?new Date(current.resolved_time).toLocaleString():'-'"/></t-list-item>
          <t-list-item v-if="current.baseline_value"><t-list-item-meta :title="t('audit.anomaly.baselineValue')"><template #description><pre class="json-pre">{{ JSON.stringify(current.baseline_value,null,2) }}</pre></template></t-list-item-meta></t-list-item>
          <t-list-item v-if="current.actual_value"><t-list-item-meta :title="t('audit.anomaly.actualValue')"><template #description><pre class="json-pre">{{ JSON.stringify(current.actual_value,null,2) }}</pre></template></t-list-item-meta></t-list-item>
        </t-list>
      </div>
    </t-dialog>
  </t-card>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { getAnomalies } from '@/api/audit-anomaly'
import type { AuditAnomaly } from '@/types/audit-anomaly'
const { t } = useI18n()
const anomalies = ref<AuditAnomaly[]>([])
const loading = ref(false); const detailVisible = ref(false); const current = ref<AuditAnomaly|null>(null)
const pagination = reactive({ current:1, pageSize:20, total:0, showJumper:true, showPageSize:true, pageSizeOptions:[10,20,50] })
const columns = computed(()=>[
  { colKey:'anomaly_type', title:t('common.type'), width:120 },
  { colKey:'severity', title:t('audit.anomaly.bySeverity'), width:80 },
  { colKey:'score', title:t('audit.anomaly.score'), width:80 },
  { colKey:'description', title:t('audit.anomaly.description'), ellipsis:true },
  { colKey:'detected_time', title:t('audit.anomaly.detectedTime'), width:170 },
  { colKey:'operation', title:t('common.actions'), width:80, fixed:'right' as const },
])
function sevTheme(s:string){ const m:Record<string,string>={low:'success',medium:'warning',high:'danger',critical:'danger'}; return m[s]||'primary' }
async function fetchAnomalies(){ loading.value=true; try{ const res = await getAnomalies({page:pagination.current,page_size:pagination.pageSize}); anomalies.value = res.data?.items||[]; pagination.total = res.data?.total||0 } catch(e:unknown){ MessagePlugin.error(e instanceof Error?e.message:String(e)) } finally { loading.value=false } }
function onPageChange(p:any){ pagination.current=p.current; pagination.pageSize=p.pageSize; fetchAnomalies() }
function handleViewDetail(row:AuditAnomaly){ current.value=row; detailVisible.value=true }
onMounted(fetchAnomalies)
</script>
<style scoped>.detail-content{max-height:500px;overflow-y:auto}.json-pre{margin:4px 0 0;padding:8px;background:#f5f5f5;border-radius:4px;max-height:150px;overflow:auto;font-size:12px}</style>
