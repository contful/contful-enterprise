<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import AppLayout from '@/components/AppLayout.vue'

const { t } = useI18n()
const route = useRoute()
const isLoginPage = computed(() => route.path === '/login')

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
  <AppLayout v-if="!isLoginPage">
    <router-view />
  </AppLayout>
  <router-view v-else />
</template>

<style>
:root {
  /* 亮色主题 */
  --color-primary: #3b82f6;
  --color-primary-light: #eff6ff;
  --color-success: #10b981;
  --color-success-light: #ecfdf5;
  --color-warning: #f59e0b;
  --color-warning-light: #fffbeb;
  --color-error: #ef4444;
  --color-error-light: #fef2f2;

  --color-bg: #f8fafc;
  --color-sidebar: #1e293b;
  --color-card: #ffffff;
  --color-border: #e2e8f0;
  --color-hover: #f1f5f9;
  --color-text: #1e293b;
  --color-text-secondary: #64748b;
}

[data-theme="dark"] {
  --color-bg: #0f172a;
  --color-sidebar: #1e293b;
  --color-card: #1e293b;
  --color-border: #334155;
  --color-hover: #334155;
  --color-text: #f1f5f9;
  --color-text-secondary: #94a3b8;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

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
  background: #2563eb;
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
