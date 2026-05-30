<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, onMounted, inject, type Ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/Icon.vue'
import StatCard from '@/components/common/StatCard.vue'
import { getDashboardStats } from '@/api/api'
import { showError } from '@/utils/request'
import PageHeader from '@/components/PageHeader.vue'
import { version } from '../../package.json'

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

    <!-- 授权信息 -->
    <t-card :bordered="true" class="license-card">
      <template #title>
        <div class="license-title">
          <t-icon name="error-circle" style="color: var(--td-warning-color)" size="18px" />
          <span>{{ t('license.title') }}</span>
        </div>
      </template>
      <t-descriptions :column="2" bordered size="small">
        <t-descriptions-item :label="t('license.productName')">Contful</t-descriptions-item>
        <t-descriptions-item :label="t('license.productVersion')">{{ version }}</t-descriptions-item>
        <t-descriptions-item :label="t('license.type')">
          <t-tag theme="success" variant="light" size="small">{{ t('license.community') }}</t-tag>
        </t-descriptions-item>
      </t-descriptions>
      <div class="license-upgrade">
        <t-alert theme="info">
          <template #message>
            <div class="license-upgrade-msg">
              <span>{{ t('license.upgrade') }}</span>
              <a href="https://contful.com" target="_blank" rel="noopener noreferrer" class="license-btn">
                {{ t('license.subscribe') }}
              </a>
            </div>
          </template>
        </t-alert>
      </div>
    </t-card>

    <!-- 开源社区 -->
    <t-card :bordered="true" class="community-card">
      <template #title>
        <div class="community-title">
          <t-icon name="logo-github" size="18px" />
          <span>{{ t('community.title') }}</span>
        </div>
      </template>
      <div class="community-grid">
        <div class="community-section">
          <h4 class="community-section-title">{{ t('community.links') }}</h4>
          <a href="https://contful.com" target="_blank" rel="noopener noreferrer" class="community-link">
            <t-icon name="internet" />
            contful.com
            <span class="community-link-hint">{{ t('community.docs') }}</span>
          </a>
          <a href="https://github.com/contful/contful" target="_blank" rel="noopener noreferrer" class="community-link">
            <t-icon name="logo-github" />
            GitHub
            <span class="community-link-hint">{{ t('community.star') }}</span>
          </a>
          <a href="https://gitee.com/contful/contful" target="_blank" rel="noopener noreferrer" class="community-link">
            <t-icon name="logo-git" />
            Gitee
            <span class="community-link-hint">{{ t('community.mirror') }}</span>
          </a>
          <a href="https://github.com/contful/contful/issues" target="_blank" rel="noopener noreferrer" class="community-link">
            <t-icon name="bug-report" />
            {{ t('community.issues') }}
            <span class="community-link-hint">{{ t('community.issuesDesc') }}</span>
          </a>
          <a href="mailto:hi@reepu.com" class="community-link">
            <t-icon name="mail" />
            hi@reepu.com
            <span class="community-link-hint">{{ t('community.contact') }}</span>
          </a>
        </div>
        <div class="community-section">
          <h4 class="community-section-title">{{ t('community.thanks') }}</h4>
          <div class="community-thanks">
            <a href="https://go.dev" target="_blank" rel="noopener" class="thanks-item">Go</a>
            <a href="https://gin-gonic.com" target="_blank" rel="noopener" class="thanks-item">Gin</a>
            <a href="https://gorm.io" target="_blank" rel="noopener" class="thanks-item">GORM</a>
            <a href="https://www.postgresql.org" target="_blank" rel="noopener" class="thanks-item">PostgreSQL</a>
            <a href="https://valkey.io" target="_blank" rel="noopener" class="thanks-item">Valkey</a>
            <a href="https://vuejs.org" target="_blank" rel="noopener" class="thanks-item">Vue 3</a>
            <a href="https://tdesign.tencent.com" target="_blank" rel="noopener" class="thanks-item">TDesign</a>
            <a href="https://vitejs.dev" target="_blank" rel="noopener" class="thanks-item">Vite</a>
            <a href="https://www.tiny.cloud" target="_blank" rel="noopener" class="thanks-item">TinyMCE</a>
            <a href="https://vitepress.dev" target="_blank" rel="noopener" class="thanks-item">VitePress</a>
            <span class="thanks-item thanks-divider">{{ t('community.andMore') }}</span>
          </div>
        </div>
      </div>
    </t-card>
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

.license-card {
  margin-top: 0;
}

.license-card :deep(.t-card__header) {
  background: var(--td-bg-color-secondarycontainer);
  border-bottom: 1px solid var(--td-component-border);
}

.license-card :deep(.t-card__title) {
  font-size: 16px;
  font-weight: 600;
}

.license-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.license-upgrade {
  margin-top: 16px;
}

.license-upgrade-msg {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.license-btn {
  display: inline-flex;
  align-items: center;
  padding: 4px 14px;
  border-radius: var(--td-radius-small);
  background: var(--td-brand-color);
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  white-space: nowrap;
  transition: background 0.2s;
}

.license-btn:hover {
  background: var(--td-brand-color-hover);
  text-decoration: none;
}

.community-card {
  margin-top: 16px;
}

.community-card :deep(.t-card__title) {
  font-size: 16px;
  font-weight: 600;
}

.community-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.community-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.community-section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--td-text-color-secondary);
  margin: 0 0 12px 0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.community-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  text-decoration: none;
  color: var(--td-text-color-primary);
  font-size: 14px;
  transition: color 0.15s;
}

.community-link:hover {
  color: var(--td-brand-color);
}

.community-link-hint {
  margin-left: auto;
  color: var(--td-text-color-placeholder);
  font-size: 12px;
}

.community-thanks {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.thanks-item {
  display: inline-block;
  padding: 2px 10px;
  background: var(--td-bg-color-secondarycontainer);
  border-radius: 4px;
  font-size: 13px;
  color: var(--td-text-color-secondary);
  text-decoration: none;
  transition: color 0.15s, background 0.15s;
}

.thanks-item:hover {
  color: var(--td-brand-color);
  background: var(--td-brand-color-light);
}

.thanks-divider {
  background: none;
  padding: 2px 0;
  font-style: italic;
  color: var(--td-text-color-disabled);
}
</style>
