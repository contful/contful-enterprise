<template>
  <PageHeader
    :title="t('license.title')"
    :subtitle="t('license.subtitle')"
  />

  <div class="license-container">
    <!-- 未授权状态 -->
    <t-alert v-if="licenseStore.isUnlicensed" theme="warning" :message="t('license.unlicensed')" />

    <!-- 已加载 License 信息 -->
    <t-card v-if="!licenseStore.isUnlicensed" :bordered="true" class="license-card">
      <template #header>
        <div class="card-header">
          <t-icon :name="licenseStore.isExpired ? 'error-circle' : 'check-circle'"
            :style="{ color: licenseStore.isExpired ? 'var(--td-warning-color)' : 'var(--td-success-color)' }" size="24px" />
          <span class="card-title">{{ licenseStore.isExpired ? t('license.expired') : t('license.active') }}</span>
        </div>
      </template>

      <t-descriptions :column="2" bordered>
        <t-descriptions-item :label="t('license.productName')">
          {{ licenseStore.productName || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.productVersion')">
          {{ licenseStore.productVersion || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.customer')">
          {{ licenseStore.customer || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.type')">
          <t-tag :theme="licenseStore.isTrial ? 'warning' : 'success'" variant="light">
            {{ licenseStore.isTrial ? t('license.trial') : t('license.commercial') }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.issuedDate')">
          {{ licenseStore.issuedDate || '-' }}
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.expiryDate')">
          <span :style="{ color: licenseStore.isExpired ? 'var(--td-error-color)' : undefined }">
            {{ licenseStore.expiryDate || '-' }}
          </span>
        </t-descriptions-item>
        <t-descriptions-item :label="t('license.productCode')" :span="2">
          {{ licenseStore.productCode || '-' }}
        </t-descriptions-item>
      </t-descriptions>

      <!-- 过期提示 -->
      <t-alert v-if="licenseStore.isExpired" theme="warning" :message="t('license.expiryNotice')" style="margin-top: 16px" />
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useLicenseStore } from '@/stores/license'
import PageHeader from '@/components/PageHeader.vue'
import { onMounted } from 'vue'

const { t } = useI18n()
const licenseStore = useLicenseStore()

onMounted(() => {
  licenseStore.fetchLicense()
})
</script>

<style scoped>
.license-container {
  max-width: 720px;
}

.license-card {
  margin-top: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
}
</style>
