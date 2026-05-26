<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted, inject, type Ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/Icon.vue'
import StatCard from '@/components/common/StatCard.vue'
import { getDashboardStats } from '@/api/api'
import { showError } from '@/utils/request'
import PageHeader from '@/components/PageHeader.vue'

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

const stats = ref<{
  sites: number
  schemas: number
  entries: number
  assets: number
  users: number
  apiTokens: number
}>({
  sites: 0,
  schemas: 0,
  entries: 0,
  assets: 0,
  users: 0,
  apiTokens: 0,
})
const loading = ref(true)

async function fetchDashboardData() {
  try {
    const res = await getDashboardStats()
    const data = res.data as Record<string, number> | undefined
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

onMounted(async () => {
  if (layoutInitialized.value) {
    await fetchDashboardData()
  }
})

watch(layoutInitialized, async (ready) => {
  if (ready && loading.value) {
    await fetchDashboardData()
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
</script>

<template>
  <div class="dashboard">
    <PageHeader
      :title="t('dashboard.title')"
      :subtitle="t('dashboard.welcome')"
    />

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <StatCard
        variant="sites"
        :value="stats.sites"
        :label="t('dashboard.sites')"
        @click="router.push('/sites')"
      >
        <template #icon>
          <t-icon name="view-list" size="24px" />
        </template>
      </StatCard>

      <StatCard
        variant="entries"
        :value="stats.entries"
        :label="t('dashboard.contentEntries')"
        @click="router.push('/content/entries')"
      >
        <template #icon>
          <t-icon name="file" size="24px" />
        </template>
      </StatCard>

      <StatCard
        variant="schemas"
        :value="stats.schemas"
        :label="t('dashboard.schemas')"
        @click="router.push('/content/schemas')"
      >
        <template #icon>
          <t-icon name="chart-bar" size="24px" />
        </template>
      </StatCard>

      <StatCard
        variant="assets"
        :value="stats.assets"
        :label="t('dashboard.mediaFiles')"
        @click="router.push('/assets')"
      >
        <template #icon>
          <t-icon name="image" size="24px" />
        </template>
      </StatCard>

      <StatCard
        variant="users"
        :value="stats.users"
        :label="t('dashboard.users')"
        @click="router.push('/users')"
      >
        <template #icon>
          <t-icon name="user" size="24px" />
        </template>
      </StatCard>

      <StatCard
        variant="tokens"
        :value="stats.apiTokens"
        :label="t('dashboard.apiTokens')"
        @click="router.push('/tokens')"
      >
        <template #icon>
          <t-icon name="secured" size="24px" />
        </template>
      </StatCard>
    </div>

    <!-- 快速操作 -->
    <div class="card quick-actions">
      <h3 class="card-title">{{ t('dashboard.quickActions') }}</h3>
      <div class="actions-row">
        <button
          v-for="action in quickActions"
          :key="action.label"
          class="action-item"
          @click="router.push(action.path)"
        >
          <span class="action-icon" :style="{ background: action.color }">
            <t-icon name="add" size="16px" style="color:white" />
              <Icon v-else-if="action.icon === 'upload'" name="arrow-up" style="color: white" />
              <path v-else-if="action.icon === 'schema'" d="M4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm0 6a1 1 0 011-1h6a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1v-2z"/>
              <path v-else-if="action.icon === 'token'" d="M7 7a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zm-4 4a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zM4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5z"/>
            </svg>
          </span>
          <span class="action-label">{{ action.label }}</span>
        </button>
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
  gap: var(--space-6);
  margin-bottom: var(--space-6);
}

.card {
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: var(--space-4);
}

.actions-row {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.action-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  background: var(--color-hover);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: 14px;
  color: var(--color-text);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: inherit;
}

.action-item:hover {
  background: var(--color-primary-light);
  border-color: var(--color-primary);
  color: var(--color-primary);
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.action-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
}

.action-label {
  font-weight: 500;
  white-space: nowrap;
}
</style>
