<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { useSiteStore } from '@/stores/site'
import { getSite, updateSite, type Site, type SiteConfig } from '@/api/site'
import { showError } from '@/utils/request'

const { t } = useI18n()

// ============ 状态 ============
const siteStore = useSiteStore()
const loading = ref(false)
const saving = ref(false)

// 站点数据
const site = ref<Site | null>(null)

// 表单数据
const form = ref({
  name: '',
  slug: '',
  description: '',
  site_url: '',
  config: {
    timezone: 'Asia/Shanghai',
    locale: 'zh-CN',
  } as SiteConfig,
})

// 原始 slug（不可修改）
const originalSlug = ref('')

// 时区选项（labelKey 模式）
const timezoneOptions = [
  { value: 'Asia/Shanghai', labelKey: 'settings.tzShanghai' },
  { value: 'Asia/Tokyo', labelKey: 'settings.tzTokyo' },
  { value: 'Asia/Singapore', labelKey: 'settings.tzSingapore' },
  { value: 'Asia/Hong_Kong', labelKey: 'settings.tzHongKong' },
  { value: 'America/New_York', labelKey: 'settings.tzNewYork' },
  { value: 'America/Los_Angeles', labelKey: 'settings.tzLosAngeles' },
  { value: 'Europe/London', labelKey: 'settings.tzLondon' },
  { value: 'Europe/Paris', labelKey: 'settings.tzParis' },
  { value: 'UTC', labelKey: 'settings.tzUTC' },
]

const timezoneLabels = computed(() =>
  timezoneOptions.map(opt => ({ ...opt, label: t(opt.labelKey) }))
)

// 语言选项（labelKey 模式）
const localeOptions = [
  { value: 'zh-CN', labelKey: 'settings.langZhCN' },
  { value: 'zh-TW', labelKey: 'settings.langZhTW' },
  { value: 'en-US', labelKey: 'settings.langEn' },
  { value: 'ja-JP', labelKey: 'settings.langJa' },
  { value: 'ko-KR', labelKey: 'settings.langKo' },
]

const localeLabels = computed(() =>
  localeOptions.map(opt => ({ ...opt, label: t(opt.labelKey) }))
)

// ============ 加载数据 ============
const loadSite = async () => {
  // 如果没有当前站点，先加载站点列表
  if (!siteStore.currentSiteId) {
    await siteStore.fetchSites()
    // 加载后仍没有站点，无需继续
    if (!siteStore.currentSiteId) {
      loading.value = false
      return
    }
  }

  loading.value = true
  try {
    const res = await getSite(siteStore.currentSiteId)
    if (res.code === 200) {
      site.value = res.data

      // 填充表单
      const data = res.data
      form.value = {
        name: data.name || '',
        slug: data.slug || '',
        description: data.description || '',
        site_url: data.site_url || '',
        config: (data.config as SiteConfig) || { timezone: 'Asia/Shanghai', locale: 'zh-CN' },
      }
      originalSlug.value = data.slug || ''
    }
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

// ============ 保存 ============
const handleSave = async () => {
  if (!siteStore.currentSiteId) return

  saving.value = true
  try {
    const res = await updateSite(siteStore.currentSiteId, {
      name: form.value.name,
      slug: form.value.slug,
      description: form.value.description,
      site_url: form.value.site_url || undefined,
      config: form.value.config,
    })

    if (res.code === 200) {
      site.value = res.data
      originalSlug.value = res.data.slug
      MessagePlugin.success(t('settings.saved'))
    }
  } catch (error) {
    showError(error)
  } finally {
    saving.value = false
  }
}

// ============ 挂载 ============
onMounted(() => {
  loadSite()
})
</script>

<template>
  <div class="site-settings">
    <div v-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>{{ t('settings.loading') }}</p>
    </div>

    <div v-else-if="!siteStore.currentSiteId" class="empty-state">
      <p class="text-secondary">{{ t('settings.noSite') }}</p>
    </div>

    <div v-else class="settings-form">
      <!-- 基础信息 -->
      <section class="settings-section">
        <h3 class="section-title">{{ t('settings.basicInfo') }}</h3>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">{{ t('settings.siteName') }} <span class="required">*</span></label>
            <input
              v-model="form.name"
              type="text"
              class="input"
              :placeholder="t('settings.siteNamePlaceholder')"
              maxlength="200"
            />
          </div>
          <div class="form-item">
            <label class="form-label">
              {{ t('settings.siteSlug') }} <span class="required">*</span>
            </label>
            <input
              v-model="form.slug"
              type="text"
              class="input"
              placeholder="site-identifier"
              maxlength="100"
            />
            <p class="form-hint">{{ t('settings.slugFormat') }}</p>
            <p v-if="form.slug !== originalSlug" class="form-warning">
              {{ t('settings.slugWarning') }}
            </p>
          </div>
          <div class="form-item form-item-full">
            <label class="form-label">{{ t('settings.siteDescription') }}</label>
            <textarea
              v-model="form.description"
              class="input textarea"
              :placeholder="t('settings.siteDescPlaceholder')"
              rows="3"
              maxlength="2000"
            ></textarea>
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('settings.siteUrl') }}</label>
            <input
              v-model="form.site_url"
              type="url"
              class="input"
              :placeholder="t('settings.siteUrlPlaceholder')"
            />
            <p class="form-hint">{{ t('settings.siteUrlTip') }}</p>
          </div>
        </div>
      </section>

      <!-- 区域设置 -->
      <section class="settings-section">
        <h3 class="section-title">{{ t('settings.region') }}</h3>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">{{ t('settings.timezone') }}</label>
            <select v-model="form.config.timezone" class="input select">
              <option v-for="opt in timezoneLabels" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
          <div class="form-item">
            <label class="form-label">{{ t('settings.language') }}</label>
            <select v-model="form.config.locale" class="input select">
              <option v-for="opt in localeLabels" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
        </div>
      </section>

      <!-- 提示信息 -->
      <section class="settings-section">
        <h3 class="section-title">{{ t('settings.deploymentNote') }}</h3>
        <p class="section-desc">{{ t('settings.deploymentNoteDesc') }}</p>
      </section>

      <!-- 提交按钮 -->
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="saving || !form.name || !form.slug" @click="handleSave">
          {{ saving ? t('settings.savingBtn') : t('settings.saveBtn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.site-settings {
  min-height: 200px;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  color: var(--color-text-secondary);
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.settings-section {
  padding: 24px 0;
  border-bottom: 1px solid var(--color-border);
}

.settings-section:last-of-type {
  border-bottom: none;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 16px;
}

.section-desc {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: -8px 0 16px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-item-full {
  grid-column: 1 / -1;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}

.required {
  color: var(--color-error);
}

.form-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
  margin: 0;
}

.form-warning {
  font-size: 12px;
  color: var(--color-warning, #e6a23c);
  margin: 0;
}

.input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  color: var(--color-text);
  background: var(--color-bg);
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.input:disabled {
  background: var(--color-bg-secondary, #f5f5f5);
  color: var(--color-text-secondary);
  cursor: not-allowed;
}

.textarea {
  resize: vertical;
  min-height: 80px;
  font-family: inherit;
  line-height: 1.5;
}

.select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23666' d='M2 4l4 4 4-4'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
  cursor: pointer;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

/* Form Actions */
.form-actions {
  padding: 24px 0 0;
  display: flex;
  justify-content: flex-end;
}

.text-secondary {
  color: var(--color-text-secondary);
  font-size: 14px;
}
</style>
