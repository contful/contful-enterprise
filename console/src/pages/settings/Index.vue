<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ApiTokens from './ApiTokens.vue'
import SiteSettings from './SiteSettings.vue'
import Configs from './Configs.vue'

const { t } = useI18n()

const activeTab = ref('api-tokens')
</script>

<template>
  <div class="settings-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('settings.title') }}</h1>
        <p class="page-subtitle">{{ t('settings.subtitle') }}</p>
      </div>
    </div>

    <div class="settings-layout">
      <nav class="settings-nav">
        <button
          class="nav-item"
          :class="{ active: activeTab === 'api-tokens' }"
          @click="activeTab = 'api-tokens'"
        >
          {{ t('menu.apiTokens') }}
        </button>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'site' }"
          @click="activeTab = 'site'"
        >
          {{ t('menu.siteSettings') }}
        </button>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'configs' }"
          @click="activeTab = 'configs'"
        >
          {{ t('settings.configs') }}
        </button>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'profile' }"
          @click="activeTab = 'profile'"
        >
          {{ t('settings.personalProfile') }}
        </button>
      </nav>

      <div class="settings-content">
        <ApiTokens v-if="activeTab === 'api-tokens'" />
        <SiteSettings v-else-if="activeTab === 'site'" />
        <Configs v-else-if="activeTab === 'configs'" />
        <div v-else-if="activeTab === 'profile'" class="card">
          <h3 class="card-title">{{ t('settings.personalProfile') }}</h3>
          <p class="text-secondary">{{ t('settings.profileComingSoon') }}</p>
        </div>
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
