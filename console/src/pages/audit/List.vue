<template>
    <PageHeader
      :title="t('audit.pageTitle')"
      :subtitle="t('audit.pageSubtitle')"
      :show-refresh="true"
      @refresh="handleSearch"
    />

    <!-- 筛选栏 -->
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
        <t-date-picker
          v-model="filterForm.startTime"
          enable-time-picker
          allow-input
          clearable
          style="width: 150px"
          :placeholder="t('audit.filter.startTimePlaceholder')"
        />
        <t-date-picker
          v-model="filterForm.endTime"
          enable-time-picker
          allow-input
          clearable
          style="width: 150px"
          :placeholder="t('audit.filter.endTimePlaceholder')"
        />
        <t-button theme="primary" @click="handleSearch">
          <template #icon><t-icon name="search" /></template>
        </t-button>
        <t-button theme="default" @click="handleReset">
          {{ t('audit.filter.reset') }}
        </t-button>
      </div>
    </t-card>

    <!-- 日志表格 -->
    <t-table
      :data="logs"
      :columns="columns"
      :loading="loading"
      :pagination="pagination"
      row-key="id"
      @page-change="onPageChange"
      hover
      stripe
      size="medium"
      class="log-table"
    >
      <template #category="{ row }">
        <t-tag variant="light">{{ t('audit.category.' + row.category) }}</t-tag>
      </template>
      <template #level="{ row }">
        <t-tag :theme="levelTagType(row.level)" variant="light">{{ t('audit.level.' + row.level) }}</t-tag>
      </template>
      <template #created_time="{ row }">
        {{ formatTime(row.created_time) }}
      </template>
      <template #operation="{ row }">
        <t-button variant="text" theme="primary" @click="handleViewDetail(row)">{{ t('audit.detailBtn') }}</t-button>
      </template>
    </t-table>

    <!-- 详情弹窗 -->
    <t-dialog
      v-model:visible="detailVisible"
      :header="t('audit.detailTitle')"
      :width="800"
      :footer="false"
    >
      <div class="detail-content" v-if="currentLog">
        <t-list :split="true">
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.id')" :description="currentLog.id" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.action')" :description="currentLog.action" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.category')" :description="t('audit.category.' + currentLog.category)" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.level')">
              <template #description>
                <t-tag :theme="levelTagType(currentLog.level)" variant="light">{{ t('audit.level.' + currentLog.level) }}</t-tag>
              </template>
            </t-list-item-meta>
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.resourceType')" :description="currentLog.resource_type || '-'" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.resourceId')" :description="currentLog.resource_id || '-'" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.userId')" :description="currentLog.user_id || '-'" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.siteId')" :description="currentLog.site_id || '-'" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.ipAddress')" :description="currentLog.ip_address || '-'" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.userAgent')" :description="currentLog.user_agent || '-'" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.details')" :description="currentLog.details || '-'" />
          </t-list-item>
          <t-list-item>
            <t-list-item-meta :title="t('audit.detail.time')" :description="formatTime(currentLog.created_time)" />
          </t-list-item>
          <t-list-item>
            <template #content>
              <div>
                <strong>{{ t('audit.dataSignature') }}：</strong>
                <pre class="signature-pre">{{ currentLog.data_signature || '-' }}</pre>
              </div>
            </template>
          </t-list-item>
        </t-list>
      </div>
    </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { getAuditLogs } from '@/api/audit'
import type { AuditLog, AuditLevel, AuditType } from '@/types/audit'
import PageHeader from '@/components/PageHeader.vue'

function handleError(err: unknown) {
  if (err instanceof Error) {
    MessagePlugin.error(err.message)
  } else {
    MessagePlugin.error(String(err))
  }
}

const { t, locale } = useI18n()
const logs = ref<AuditLog[]>([])
const loading = ref(false)
const detailVisible = ref(false)
const currentLog = ref<AuditLog | null>(null)

const filterForm = reactive({
  action: '',
  resource_type: '',
  category: '' as AuditType | '',
  level: '' as AuditLevel | '',
  startTime: '',
  endTime: '',
})

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
  showPageSize: true,
  pageSizeOptions: [10, 20, 50],
})

const columns = computed(() => [
  { colKey: 'action', title: t('audit.column.action'), width: 140 },
  { colKey: 'category', title: t('audit.column.category'), width: 100 },
  { colKey: 'level', title: t('audit.column.level'), width: 80 },
  { colKey: 'resource_type', title: t('audit.column.resourceType'), width: 120 },
  { colKey: 'ip_address', title: t('audit.column.ipAddress'), width: 140 },
  { colKey: 'created_time', title: t('audit.column.time'), width: 170 },
  { colKey: 'operation', title: t('audit.column.operation'), width: 80, fixed: 'right' as const },
])


function levelTagType(level: string) {
  const map: Record<string, string> = {
    debug: 'primary',
    info: 'success',
    warn: 'warning',
    error: 'danger',
  }
  return map[level] || 'primary'
}

function formatTime(time: string) {
  if (!time) return '-'
  return new Date(time).toLocaleString(locale.value, { hour12: false })
}

async function fetchLogs() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.current,
      page_size: pagination.pageSize,
    }
    if (filterForm.category) params.category = filterForm.category
    if (filterForm.level) params.level = filterForm.level
    if (filterForm.startTime) params.start_time = filterForm.startTime
    if (filterForm.endTime) params.end_time = filterForm.endTime
    const res = await getAuditLogs(params)
    logs.value = res.data?.items || []
    pagination.total = res.data?.total || 0
  } catch (err: unknown) {
    handleError(err)
  } finally {
    loading.value = false
  }
}

function onPageChange(pageInfo: { current: number; pageSize: number }) {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  fetchLogs()
}

function handleSearch() {
  pagination.current = 1
  fetchLogs()
}

function handleReset() {
  filterForm.category = ''
  filterForm.level = ''
  filterForm.startTime = ''
  filterForm.endTime = ''
  pagination.current = 1
  fetchLogs()
}

function handleViewDetail(row: AuditLog) {
  currentLog.value = row
  detailVisible.value = true
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.filter-card {
  margin-bottom: 16px;
}
.filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}
.log-table {
  margin-top: 16px;
}
.detail-content {
  max-height: 600px;
  overflow-y: auto;
}
.signature-pre {
  margin: 8px 0 0 0;
  padding: 8px;
  background: #f5f5f5;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
  font-size: 12px;
}
</style>
