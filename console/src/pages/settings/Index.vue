<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Configs from './Configs.vue'
import Security from './Security.vue'
import Profile from './Profile.vue'

const { t } = useI18n()

const activeTab = ref('profile')
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
          :class="{ active: activeTab === 'profile' }"
          @click="activeTab = 'profile'"
        >
          {{ t('settings.personalProfile') }}
        </button>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'security' }"
          @click="activeTab = 'security'"
        >
          {{ t('settings.security') }}
        </button>
        <button
          class="nav-item"
          :class="{ active: activeTab === 'configs' }"
          @click="activeTab = 'configs'"
        >
          {{ t('settings.configs') }}
        </button>
      </nav>

      <div class="settings-content">
        <Profile v-if="activeTab === 'profile'" />
        <Security v-else-if="activeTab === 'security'" />
        <Configs v-else-if="activeTab === 'configs'" />
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
