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
import { useLicenseStore } from '@/stores/license'

function handleError(err: unknown) {
  if (err instanceof Error) {
    showError(err.message)
  } else {
    showError(String(err))
  }
}

const { t, locale } = useI18n()
const router = useRouter()
const licenseStore = useLicenseStore()

function formatDate(isoStr: string | null): string {
  if (!isoStr) return '-'
  try {
    const d = new Date(isoStr)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  } catch {
    return isoStr
  }
}

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
    licenseStore.fetchLicense()
  }
})

watch(layoutInitialized, async (ready) => {
  if (ready && loading.value) {
    await fetchDashboardData()
    licenseStore.fetchLicense()
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
          <t-icon
            :name="licenseStore.isExpired ? 'error-circle' : 'check-circle-filled'"
            :style="{ color: licenseStore.isExpired ? 'var(--td-warning-color)' : 'var(--td-success-color)' }"
            size="18px"
          />
          <span>{{ t('license.title') }}</span>
        </div>
      </template>

      <t-alert v-if="licenseStore.isUnlicensed" theme="warning" :message="t('license.unlicensed')" />

      <t-descriptions v-if="!licenseStore.isUnlicensed" :column="2" bordered size="small">
        <t-descriptions-item :label="t('license.productName')">
          {{ licenseStore.productName || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.productVersion')">
          {{ licenseStore.productVersion || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.type')">
          <t-tag :theme="licenseStore.isTrial ? 'warning' : 'success'" variant="light" size="small">
            {{ licenseStore.isTrial ? t('license.trial') : t('license.commercial') }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.productCode')">
          {{ licenseStore.productCode || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.customer')">
          {{ licenseStore.customer || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.issuedDate')">
          {{ formatDate(licenseStore.issuedDate) }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.expiryDate')" :span="2">
          <span :style="{ color: licenseStore.isExpired ? 'var(--td-error-color)' : undefined }">
            {{ formatDate(licenseStore.expiryDate) }}
          </span>
          <t-tag v-if="licenseStore.isExpired" theme="danger" size="small" style="margin-left: 8px">{{ t('license.expired') }}</t-tag>
        </t-descriptions-item>
      </t-descriptions>
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
</style>
