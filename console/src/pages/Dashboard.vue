<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted, inject, type Ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/Icon.vue'
import { getDashboardStats } from '@/api/api'
import { getContentEntries } from '@/api/api'
import { showError } from '@/utils/request'

function handleError(err: unknown) {
  if (err instanceof Error) {
    showError(err.message)
  } else {
    showError(String(err))
  }
}

const { t, locale } = useI18n()
const router = useRouter()

// 等待 Layout 初始化完成（确保 currentSiteId 已设置）
const layoutInitialized = inject<Ref<boolean>>('layoutInitialized', ref(true))

const stats = ref({
  sites: 0,
  schemas: 0,
  entries: 0,
  assets: 0,
  users: 0,
  apiTokens: 0,
})
const recentEntries = ref<any[]>([])
const loading = ref(true)

async function fetchDashboardData() {
  try {
    // 单个接口获取全部统计
    const res = await getDashboardStats()
    const data = res.data
    stats.value = {
      sites: data?.sites || 0,
      schemas: data?.schemas || 0,
      entries: data?.entries || 0,
      assets: data?.assets || 0,
      users: data?.users || 0,
      apiTokens: data?.api_tokens || 0,
    }
  } catch (error) {
    handleError(error)
  } finally {
    loading.value = false
  }
}

// 最近内容单独获取（可选 site_id，无 site 时不阻塞）
async function fetchRecentEntries() {
  try {
    const res = await getContentEntries({ page: 1, page_size: 5 })
    recentEntries.value = res.data?.items || []
  } catch {
    recentEntries.value = []
  }
}

onMounted(async () => {
  if (layoutInitialized.value) {
    await fetchDashboardData()
    await fetchRecentEntries()
  }
})

watch(layoutInitialized, async (ready) => {
  if (ready && loading.value) {
    await fetchDashboardData()
    await fetchRecentEntries()
  }
})

const quickActions = computed(() => {
  void locale.value
  return [
    { icon: 'add', label: t('dashboard.createContent'), path: '/content/entries', color: '#3b82f6' },
    { icon: 'upload', label: t('dashboard.uploadMedia'), path: '/assets', color: '#10b981' },
    { icon: 'schema', label: t('dashboard.manageTypes'), path: '/content/schemas', color: '#8b5cf6' },
    { icon: 'token', label: t('menu.apiTokens'), path: '/tokens', color: '#f59e0b' },
  ]
})

const getStatusClass = (status: string) => {
  const map: Record<string, string> = {
    published: 'badge-success',
    draft: 'badge-warning',
    archived: 'badge-default',
  }
  return map[status] || 'badge-default'
}

// t-table 列定义（最近内容）
const getStatusTheme = (status: string) => {
  const map: Record<string, string> = {
    published: 'success',
    draft: 'warning',
    archived: 'default',
  }
  return map[status] || 'default'
}

const recentColumns = computed(() => [
  {
    colKey: 'title',
    title: t('dashboard.titleCol'),
    cell: (_h: any, { row }: { row: any }) =>
      row.values?.find((v: any) => v.field?.name === 'title')?.text_value || row.id.slice(0, 8),
  },
  {
    colKey: 'type',
    title: t('dashboard.typeCol'),
    cell: (_h: any, { row }: { row: any }) =>
      row.content_schema?.name || row.schema_id?.slice(0, 8) || '-',
  },
  { colKey: 'status', title: t('dashboard.statusCol'), width: 100 },
  {
    colKey: 'updated_time',
    title: t('dashboard.updatedCol'),
    width: 120,
    cell: (_h: any, { row }: { row: any }) => new Date(row.updated_time).toLocaleDateString(),
  },
])

const getStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    published: t('content.published'),
    draft: t('content.draft'),
    archived: t('content.archived'),
  }
  return map[status] || status
}
</script>

<template>
  <div class="dashboard">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('dashboard.title') }}</h1>
        <p class="page-subtitle">{{ t('dashboard.welcome') }}</p>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" @click="router.push('/sites')">
        <div class="stat-icon" style="background: #fef2f2; color: #ef4444;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 6a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1v-2zm0 6a1 1 0 011-1h6a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1v-2z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.sites }}</div>
          <div class="stat-label">{{ t('dashboard.sites') }}</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/content/entries')">
        <div class="stat-icon" style="background: #eff6ff; color: #3b82f6;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M4 4a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 0v10h8V4H6z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.entries }}</div>
          <div class="stat-label">{{ t('dashboard.contentEntries') }}</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/content/schemas')">
        <div class="stat-icon" style="background: #f3e8ff; color: #8b5cf6;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1v-2zm0 6a1 1 0 011-1h6a1 1 0 011 1v2a1 1 0 011 1H4a1 1 0 01-1-1v-2z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.schemas }}</div>
          <div class="stat-label">{{ t('dashboard.schemas') }}</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/assets')">
        <div class="stat-icon" style="background: #ecfdf5; color: #10b981;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm0 2h12v7l-4-3-2 1.5L6 12V5z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.assets }}</div>
          <div class="stat-label">{{ t('dashboard.mediaFiles') }}</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/users')">
        <div class="stat-icon" style="background: #fef3c7; color: #f59e0b;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.users }}</div>
          <div class="stat-label">{{ t('dashboard.users') }}</div>
        </div>
      </div>

      <div class="stat-card" @click="router.push('/tokens')">
        <div class="stat-icon" style="background: #fce7f3; color: #ec4899;">
          <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
            <path d="M7 7a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zm-4 4a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zM4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5z"/>
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.apiTokens }}</div>
          <div class="stat-label">{{ t('dashboard.apiTokens') }}</div>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <!-- 快速操作 -->
      <div class="card quick-actions">
        <h3 class="card-title">{{ t('dashboard.quickActions') }}</h3>
        <div class="actions-list">
          <button
            v-for="action in quickActions"
            :key="action.label"
            class="action-item"
            @click="router.push(action.path)"
          >
            <span class="action-icon" :style="{ background: action.color }">
              <svg width="16" height="16" viewBox="0 0 20 20" fill="white">
                <path v-if="action.icon === 'add'" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"/>
                <Icon v-else-if="action.icon === 'upload'" name="arrow-up" style="color: white" />
                <path v-else-if="action.icon === 'schema'" d="M4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h6a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1v-2z"/>
                <path v-else-if="action.icon === 'token'" d="M7 7a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zm-4 4a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zM4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5z"/>
              </svg>
            </span>
            <span class="action-label">{{ action.label }}</span>
            <svg class="action-arrow" width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- 最近内容 -->
      <div class="card recent-entries">
        <h3 class="card-title">{{ t('dashboard.recentContent') }}</h3>
        <div v-if="loading" class="loading">{{ t('common.loading') }}</div>
        <div v-else-if="recentEntries.length === 0" class="empty-state">
          <p>{{ t('dashboard.noContent') }}</p>
          <t-button theme="primary" @click="router.push('/content/entries')">
            {{ t('dashboard.createFirstContent') }}
          </t-button>
        </div>
        <t-table
          v-else
          :data="recentEntries"
          :columns="recentColumns"
          :pagination="false"
          hover
          size="small"
          row-key="id"
          @row-click="(row) => router.push(`/content/entries?type=${row.schema_id}&id=${row.id}`)"
          class="recent-table"
        >
          <template #status-cell="{ row }">
            <t-tag :theme="(getStatusTheme(row.status) as any)" variant="light" size="small">
              {{ getStatusLabel(row.status) }}
            </t-tag>
          </template>
        </t-table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  width: 100%;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.stat-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.1);
}

.stat-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text);
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: var(--color-text-secondary);
}

.dashboard-grid {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 16px;
}

.quick-actions .actions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
  width: 100%;
}

.action-item:hover {
  background: var(--color-hover);
  border-color: var(--color-primary);
}

.action-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.action-label {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
}

.action-arrow {
  color: var(--color-text-secondary);
}

.loading {
  text-align: center;
  padding: 40px;
  color: var(--color-text-secondary);
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
