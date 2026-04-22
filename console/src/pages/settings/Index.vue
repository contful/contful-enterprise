<script setup lang="ts">
import { ref } from 'vue'
import ApiTokens from './ApiTokens.vue'
import SiteSettings from './SiteSettings.vue'

const activeTab = ref('api-tokens')

const tabs = [
  { key: 'api-tokens', label: 'API Token' },
  { key: 'profile', label: '个人资料' },
  { key: 'site', label: '站点设置' },
]
</script>

<template>
  <div class="settings-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">设置</h1>
        <p class="page-subtitle">管理您的账户和系统设置</p>
      </div>
    </div>

    <div class="settings-layout">
      <!-- 侧边导航 -->
      <nav class="settings-nav">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="nav-item"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </nav>

      <!-- 内容区 -->
      <div class="settings-content">
        <ApiTokens v-if="activeTab === 'api-tokens'" />
        <div v-else-if="activeTab === 'profile'" class="card">
          <h3 class="card-title">个人资料</h3>
          <p class="text-secondary">个人资料设置功能开发中...</p>
        </div>
        <SiteSettings v-else-if="activeTab === 'site'" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-page {
  height: 100%;
}

.settings-layout {
  display: flex;
  gap: 20px;
}

.settings-nav {
  width: 200px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
}

.settings-nav .nav-item {
  padding: 10px 16px;
  text-align: left;
  background: transparent;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.settings-nav .nav-item:hover {
  background: var(--color-hover);
  color: var(--color-text);
}

.settings-nav .nav-item.active {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.settings-content {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 16px;
}

.text-secondary {
  color: var(--color-text-secondary);
  font-size: 14px;
}
</style>
