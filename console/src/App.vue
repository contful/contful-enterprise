<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import Layout from '@/components/Layout.vue'

const { t } = useI18n()
const route = useRoute()
// 不需要 Layout 的认证类页面
const isAuthPage = computed(() => ['/login', '/mfa'].includes(route.path))


// =============================================================================
// MX-002-2: 全局网络状态监听
// =============================================================================

const handleOffline = () => {
  MessagePlugin.warning({
    content: t('app.offline'),
    duration: 0,
    closeBtn: true,
  })
}

const handleOnline = () => {
  MessagePlugin.success(t('app.online'))
}

onMounted(() => {
  window.addEventListener('offline', handleOffline)
  window.addEventListener('online', handleOnline)
})

onUnmounted(() => {
  window.removeEventListener('offline', handleOffline)
  window.removeEventListener('online', handleOnline)
})
</script>

<template>
  <a href="#main-content" class="skip-link">跳到主要内容</a>
  <Layout v-if="!isAuthPage">
    <router-view />
  </Layout>
  <router-view v-else />
</template>

<style>
/* ==========================================================================
   Utility Classes (Tokens 已迁移至 src/styles/index.css)
   ========================================================================== */

a {
  color: inherit;
  text-decoration: none;
}

button {
  font-family: inherit;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--color-text);
}

.page-subtitle {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin-top: 4px;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
}

.btn-secondary {
  background: var(--color-card);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.btn-secondary:hover {
  background: var(--color-hover);
}

.btn-danger {
  background: var(--color-error);
  color: white;
}

.btn-danger:hover {
  background: #dc2626;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
}

.btn-icon {
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 6px;
}

.card {
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th,
.table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.table th {
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text-secondary);
  background: var(--color-hover);
}

.table td {
  font-size: 14px;
  color: var(--color-text);
}

.table tr:hover td {
  background: var(--color-hover);
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 9999px;
}

.badge-success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.badge-warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.badge-error {
  background: var(--color-error-light);
  color: var(--color-error);
}

.badge-default {
  background: var(--color-hover);
  color: var(--color-text-secondary);
}

.input {
  width: 100%;
  padding: 10px 14px;
  font-size: 14px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-card);
  color: var(--color-text);
  transition: all 0.2s;
}

.input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-light);
}

.input-group {
  margin-bottom: 16px;
}

.input-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  margin-bottom: 6px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--color-text-secondary);
}

.empty-state svg {
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-state h3 {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--color-text);
}

.empty-state p {
  font-size: 14px;
  margin-bottom: 16px;
}
</style>
