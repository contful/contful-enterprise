<template>
    <PageHeader
      :title="t('audit.pageTitle')"
      :subtitle="t('audit.pageSubtitle')"
      :show-refresh="true"
      @refresh="handleSearch"
    />

    <t-tabs v-model="activeTab">
      <t-tab-panel value="logs" :label="t('audit.logsTab')">
        <AdvancedFilter @search="handleAdvancedSearch" />

        <t-card class="filter-card">
          <div class="filter-bar">
            <t-select v-model="filterForm.category" :placeholder="t('audit.filter.category')" clearable style="width: 150px">
              <t-option :label="t('audit.category.auth')" value="auth" />
              <t-option :label="t('audit.category.content')" value="content" />
              <t-option :label="t('audit.category.media')" value="media" />
              <t-option :label="t('audit.category.settings')" value="settings" />
              <t-option :label="t('audit.category.user')" value="user" />
              <t-option :label="t('audit.category.system')" value="system" />
            </t-select>
            <t-select v-model="filterForm.level" :placeholder="t('audit.filter.level')" clearable style="width: 150px">
              <t-option :label="t('audit.level.debug')" value="debug" />
              <t-option :label="t('audit.level.info')" value="info" />
              <t-option :label="t('audit.level.warn')" value="warn" />
              <t-option :label="t('audit.level.error')" value="error" />
            </t-select>
            <t-date-picker v-model="filterForm.startTime" enable-time-picker allow-input clearable style="width:150px" :placeholder="t('audit.filter.startTimePlaceholder')" />
            <t-date-picker v-model="filterForm.endTime" enable-time-picker allow-input clearable style="width:150px" :placeholder="t('audit.filter.endTimePlaceholder')" />
            <t-button theme="primary" @click="handleSearch"><template #icon><t-icon name="search" /></template></t-button>
            <t-button theme="default" @click="handleReset">{{ t('audit.filter.reset') }}</t-button>
            <div class="filter-spacer" />
            <QueryTemplate @apply="handleTemplateApply" />
            <t-button variant="outline" @click="showExportDialog = true"><template #icon><t-icon name="download" /></template>{{ t('audit.export.exportBtn') }}</t-button>
          </div>
        </t-card>

        <t-table :data="logs" :columns="columns" :loading="loading" :pagination="pagination" row-key="id" @page-change="onPageChange" hover stripe size="medium" class="log-table">
          <template #category="{ row }"><t-tag variant="light">{{ t('audit.category.' + row.category) }}</t-tag></template>
          <template #level="{ row }"><t-tag :theme="levelTagType(row.level)" variant="light">{{ t('audit.level.' + row.level) }}</t-tag></template>
          <template #created_time="{ row }">{{ formatTime(row.created_time) }}</template>
          <template #operation="{ row }"><t-button variant="text" theme="primary" @click="handleViewDetail(row)">{{ t('audit.detailBtn') }}</t-button></template>
        </t-table>

        <t-dialog v-model:visible="detailVisible" :header="t('audit.detailTitle')" :width="800" :footer="false">
          <div class="detail-content" v-if="currentLog">
            <t-list :split="true">
              <t-list-item><t-list-item-meta :title="t('audit.detail.id')" :description="currentLog.id" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.action')" :description="currentLog.action" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.category')" :description="t('audit.category.' + currentLog.category)" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.level')"><template #description><t-tag :theme="levelTagType(currentLog.level)" variant="light">{{ t('audit.level.' + currentLog.level) }}</t-tag></template></t-list-item-meta></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.resourceType')" :description="currentLog.resource_type || '-'" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.resourceId')" :description="currentLog.resource_id || '-'" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.userId')" :description="currentLog.user_id || '-'" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.siteId')" :description="currentLog.site_id || '-'" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.ipAddress')" :description="currentLog.ip_address || '-'" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.userAgent')" :description="currentLog.user_agent || '-'" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.details')" :description="currentLog.details || '-'" /></t-list-item>
              <t-list-item><t-list-item-meta :title="t('audit.detail.time')" :description="formatTime(currentLog.created_time)" /></t-list-item>
              <t-list-item><template #content><div><strong>{{ t('audit.dataSignature') }}：</strong><pre class="signature-pre">{{ currentLog.data_signature || '-' }}</pre></div></template></t-list-item>
            </t-list>
          </div>
        </t-dialog>
      </t-tab-panel>

      <t-tab-panel value="anomaly" :label="t('audit.anomalyTab')">
        <AnomalyTab />
      </t-tab-panel>
    </t-tabs>

    <ExportDialog :visible="showExportDialog" @close="showExportDialog = false" @export="handleExport" />
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { getAuditLogs, exportAuditJSON } from '@/api/audit'
import type { AuditLog, AuditLevel, AuditType } from '@/types/audit'
import type { AuditTemplateConditions } from '@/types/audit-query-template'
import PageHeader from '@/components/PageHeader.vue'
import AdvancedFilter from './components/AdvancedFilter.vue'
import QueryTemplate from './components/QueryTemplate.vue'
import ExportDialog from './components/ExportDialog.vue'
import AnomalyTab from './components/AnomalyTab.vue'

function handleError(err: unknown) { if (err instanceof Error) MessagePlugin.error(err.message); else MessagePlugin.error(String(err)) }

const { t, locale } = useI18n()
const logs = ref<AuditLog[]>([]); const loading = ref(false)
const detailVisible = ref(false); const currentLog = ref<AuditLog|null>(null)
const showExportDialog = ref(false); const activeTab = ref('logs')

const advancedFilter = reactive({ keyword: '', combinator: 'and', time_preset: '' })
const templateFilter = reactive({ keyword: '', combinator: '', time_preset: '', category: '' as string, level: '' as string, action: '', start_time: '', end_time: '' })

const filterForm = reactive({ action: '', resource_type: '', category: '' as AuditType|'', level: '' as AuditLevel|'', startTime: '', endTime: '' })

const pagination = reactive({ current:1, pageSize:20, total:0, showJumper:true, showPageSize:true, pageSizeOptions:[10,20,50] })

const columns = computed(() => [
  { colKey:'action', title:t('audit.column.action'), width:140 }, { colKey:'category', title:t('audit.column.category'), width:100 },
  { colKey:'level', title:t('audit.column.level'), width:80 }, { colKey:'resource_type', title:t('audit.column.resourceType'), width:120 },
  { colKey:'ip_address', title:t('audit.column.ipAddress'), width:140 }, { colKey:'created_time', title:t('audit.column.time'), width:170 },
  { colKey:'operation', title:t('audit.column.operation'), width:80, fixed:'right' as const },
])

function levelTagType(level:string) { const m:Record<string,string>={debug:'primary',info:'success',warn:'warning',error:'danger'}; return m[level]||'primary' }
function formatTime(time:string) { if(!time) return '-'; return new Date(time).toLocaleString(locale.value,{hour12:false}) }

async function fetchLogs() {
  loading.value=true
  try {
    const params: Record<string,any> = { page: pagination.current, page_size: pagination.pageSize }
    if(filterForm.category) params.category = filterForm.category
    if(filterForm.level) params.level = filterForm.level
    if(filterForm.startTime) params.start_time = filterForm.startTime
    if(filterForm.endTime) params.end_time = filterForm.endTime
    if(advancedFilter.keyword) params.keyword = advancedFilter.keyword
    if(advancedFilter.combinator) params.combinator = advancedFilter.combinator
    if(advancedFilter.time_preset) params.time_preset = advancedFilter.time_preset
    if(templateFilter.keyword) params.keyword = templateFilter.keyword
    if(templateFilter.combinator) params.combinator = templateFilter.combinator
    if(templateFilter.time_preset) params.time_preset = templateFilter.time_preset
    if(templateFilter.category) params.category = templateFilter.category
    if(templateFilter.level) params.level = templateFilter.level
    if(templateFilter.action) params.action = templateFilter.action
    if(templateFilter.start_time) params.start_time = templateFilter.start_time
    if(templateFilter.end_time) params.end_time = templateFilter.end_time
    const res = await getAuditLogs(params); logs.value = res.data?.items||[]; pagination.total = res.data?.total||0
  } catch(err:unknown) { handleError(err) } finally { loading.value=false }
}

function onPageChange(p:any) { pagination.current=p.current; pagination.pageSize=p.pageSize; fetchLogs() }
function handleSearch() { pagination.current=1; fetchLogs() }
function handleReset() { filterForm.category=''; filterForm.level=''; filterForm.startTime=''; filterForm.endTime=''; advancedFilter.keyword=''; advancedFilter.combinator='and'; advancedFilter.time_preset=''; Object.assign(templateFilter, { keyword:'', combinator:'', time_preset:'', category:'', level:'', action:'', start_time:'', end_time:'' }); pagination.current=1; fetchLogs() }
function handleViewDetail(row:AuditLog) { currentLog.value=row; detailVisible.value=true }
function handleAdvancedSearch(f: {keyword:string;combinator:string;time_preset:string}) { Object.assign(advancedFilter,f); pagination.current=1; fetchLogs() }
function handleTemplateApply(c: AuditTemplateConditions) {
  Object.assign(templateFilter, { keyword:c.keyword||'', combinator:c.combinator||'', time_preset:c.time_preset||'', category:c.category||'', level:c.level||'', action:c.action||'', start_time:c.start_time||'', end_time:c.end_time||'' })
  if(c.category) filterForm.category = c.category as AuditType; if(c.level) filterForm.level = c.level as AuditLevel
  if(c.start_time) filterForm.startTime = c.start_time; if(c.end_time) filterForm.endTime = c.end_time; pagination.current=1; fetchLogs()
}

async function handleExport(opts: {format:string;fields:string[]}) {
  try {
    const params: Record<string,any> = { fields: opts.fields, export_format: opts.format, page_size: pagination.total }
    if(advancedFilter.keyword) params.keyword = advancedFilter.keyword
    if(advancedFilter.combinator) params.combinator = advancedFilter.combinator
    if(advancedFilter.time_preset) params.time_preset = advancedFilter.time_preset
    const res = await exportAuditJSON(params)
    const blob = res as unknown as Blob; const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a'); a.href=url; a.download=`audit-logs.${opts.format}`; document.body.appendChild(a); a.click(); document.body.removeChild(a); window.URL.revokeObjectURL(url)
    MessagePlugin.success(t('audit.export.success')); showExportDialog.value=false
  } catch(err:unknown) { MessagePlugin.error(t('audit.export.failed')); handleError(err) }
}

onMounted(fetchLogs)
</script>

<style scoped>
.filter-card { margin-bottom: 16px; }
.filter-bar { display: flex; align-items: center; flex-wrap: wrap; gap: 12px; }
.filter-spacer { flex: 1; }
.log-table { margin-top: 16px; }
.detail-content { max-height: 600px; overflow-y: auto; }
.signature-pre { margin: 8px 0 0 0; padding: 8px; background: #f5f5f5; border-radius: 4px; max-height: 200px; overflow: auto; font-size: 12px; }
</style>
