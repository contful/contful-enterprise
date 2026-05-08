<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { getSite, updateSite, type SiteConfig } from '@/api/site'
import { showError } from '@/utils/request'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

// ============ 状态 ============
const loading = ref(false)
const saving = ref(false)
const siteId = computed(() => route.params.siteId as string)

// 站点元数据（用于页头展示）
const siteMeta = ref<{ name: string; slug: string; id: string } | null>(null)

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

// 原始 slug（不可修改提示）
const originalSlug = ref('')

// 表单校验
const formRef = ref()
const rules = {
  name: [{ required: true, message: () => t('settings.siteNameRequired'), trigger: 'blur' }],
  slug: [{ required: true, message: () => t('settings.slugRequired'), trigger: 'blur' }],
}

// 时区选项
const timezoneOptions = computed(() => [
  { value: 'Asia/Shanghai', label: t('settings.tzShanghai') },
  { value: 'Asia/Tokyo', label: t('settings.tzTokyo') },
  { value: 'Asia/Singapore', label: t('settings.tzSingapore') },
  { value: 'Asia/Hong_Kong', label: t('settings.tzHongKong') },
  { value: 'America/New_York', label: t('settings.tzNewYork') },
  { value: 'America/Los_Angeles', label: t('settings.tzLosAngeles') },
  { value: 'Europe/London', label: t('settings.tzLondon') },
  { value: 'Europe/Paris', label: t('settings.tzParis') },
  { value: 'UTC', label: t('settings.tzUTC') },
])

// 语言选项
const localeOptions = computed(() => [
  { value: 'zh-CN', label: t('settings.langZhCN') },
  { value: 'zh-TW', label: t('settings.langZhTW') },
  { value: 'en-US', label: t('settings.langEn') },
  { value: 'ja-JP', label: t('settings.langJa') },
  { value: 'ko-KR', label: t('settings.langKo') },
])

// 是否有未保存变更
const hasChanges = computed(() =>
  form.value.name !== (siteMeta.value?.name ?? '') ||
  form.value.slug !== originalSlug.value ||
  form.value.description !== '' ||
  form.value.site_url !== ''
)

// Slug 格式校验状态
const slugError = computed(() => {
  const val = form.value.slug.trim()
  if (!val) return ''
  // 只允许小写字母、数字、连字符，且必须以字母开头
  if (!/^[a-z][a-z0-9-]*$/.test(val)) {
    return t('settings.slugFormatInvalid')
  }
  return ''
})

// ============ 加载数据 ============
const loadSite = async () => {
  if (!siteId.value) return

  loading.value = true
  try {
    const res = await getSite(siteId.value)
    if (res.code === 200) {
      siteMeta.value = { id: res.data.id, name: res.data.name, slug: res.data.slug }
      form.value = {
        name: res.data.name || '',
        slug: res.data.slug || '',
        description: res.data.description || '',
        site_url: res.data.site_url || '',
        config: (res.data.config as SiteConfig) || { timezone: 'Asia/Shanghai', locale: 'zh-CN' },
      }
      originalSlug.value = res.data.slug || ''
    }
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

// ============ 保存 ============
const handleSave = async () => {
  if (!siteId.value) return

  // 先执行表单校验
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  // Slug 格式校验
  if (slugError.value) return

  saving.value = true
  try {
    const res = await updateSite(siteId.value, {
      name: form.value.name,
      slug: form.value.slug,
      description: form.value.description,
      site_url: form.value.site_url || undefined,
      config: form.value.config,
    })

    if (res.code === 200) {
      siteMeta.value = { id: res.data.id, name: res.data.name, slug: res.data.slug }
      originalSlug.value = res.data.slug
      MessagePlugin.success(t('settings.saved'))
    }
  } catch (error) {
    showError(error)
  } finally {
    saving.value = false
  }
}

// 返回站点列表
const goBack = () => {
  router.push('/sites')
}

// 复制 Site ID
const copySiteId = async () => {
  if (!siteMeta.value?.id) return
  try {
    await navigator.clipboard.writeText(siteMeta.value.id)
    MessagePlugin.success(t('sites.siteIdCopied'))
  } catch {
    // fallback
    const input = document.createElement('input')
    input.value = siteMeta.value.id
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    MessagePlugin.success(t('sites.siteIdCopied'))
  }
}

// ============ 挂载 ============
onMounted(() => {
  loadSite()
})
</script>

<template>
  <div class="site-config-page">
    <!-- 页头：面包屑 + 站点信息 -->
    <div class="page-header">
      <div class="breadcrumb">
        <router-link to="/sites" class="breadcrumb-link">{{ t('menu.sites') }}</router-link>
        <t-icon name="chevron-right" class="breadcrumb-icon" />
        <span class="breadcrumb-current">{{ siteMeta?.name || siteId }}</span>
        <t-icon name="chevron-right" class="breadcrumb-icon" />
        <span class="breadcrumb-current">{{ t('settings.siteSettings') }}</span>
      </div>

      <div v-if="siteMeta" class="header-info">
        <div class="site-avatar-sm" :style="{ background: '#3b82f6' }">
          {{ siteMeta.name.charAt(0).toUpperCase() }}
        </div>
        <div class="header-detail">
          <h1 class="site-title">{{ t('settings.siteSettings') }}</h1>
          <div class="header-meta">
            <code class="site-slug">{{ siteMeta.slug }}</code>
            <span class="meta-sep">·</span>
            <button class="copy-id-btn" :title="t('sites.copySiteId')" @click="copySiteId">
              <t-icon name="file-copy" size="14px" />
              <span>{{ siteMeta.id.slice(0, 8) }}…</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 主内容区：TDesign Loading 包裹 -->
    <t-loading :loading="loading">
      <t-form
        v-if="!loading"
        ref="formRef"
        :data="form"
        :rules="rules"
        label-align="top"
        :label-width="0"
        class="settings-form"
        @submit.prevent="handleSave"
      >
        <!-- 基础信息 -->
        <section class="settings-section">
          <div class="section-header">
            <t-icon name="root-list" class="section-icon" />
            <h3 class="section-title">{{ t('settings.basicInfo') }}</h3>
          </div>
          <div class="form-grid">
            <t-form-item name="name" :label="t('settings.siteName')">
              <t-input
                v-model="form.name"
                :placeholder="t('settings.siteNamePlaceholder')"
                :maxlength="200"
                clearable
              >
                <template #prefix-icon><t-icon name="edit" /></template>
              </t-input>
            </t-form-item>
            <t-form-item name="slug" :label="t('settings.siteSlug')">
              <t-input
                v-model="form.slug"
                placeholder="site-identifier"
                :maxlength="100"
                :status="slugError ? 'error' : 'default'"
                :tips="slugError || t('settings.slugFormat')"
              >
                <template #prefix-icon><t-icon name="link" /></template>
                <template #suffix v-if="form.slug && form.slug !== originalSlug">
                  <span class="slug-changed-badge">
                    {{ t('settings.slugModified') }}
                  </span>
                </template>
              </t-input>
            </t-form-item>
            <t-form-item :label="t('settings.siteDescription')" class="form-item-full">
              <t-textarea
                v-model="form.description"
                :placeholder="t('settings.siteDescPlaceholder')"
                :maxlength="2000"
                :autosize="{ minRows: 3, maxRows: 6 }"
              />
            </t-form-item>
            <t-form-item :label="t('settings.siteUrl')">
              <t-input
                v-model="form.site_url"
                :placeholder="t('settings.siteUrlPlaceholder')"
                clearable
              >
                <template #prefix-icon><t-icon name="browse" /></template>
              </t-input>
              <template #help>{{ t('settings.siteUrlTip') }}</template>
            </t-form-item>
          </div>
        </section>

        <!-- 区域设置 -->
        <section class="settings-section">
          <div class="section-header">
            <t-icon name="globe" class="section-icon" />
            <h3 class="section-title">{{ t('settings.region') }}</h3>
          </div>
          <div class="form-grid">
            <t-form-item :label="t('settings.timezone')">
              <t-select
                v-model="form.config.timezone"
                :options="timezoneOptions"
                filterable
                :placeholder="t('common.selectPlaceholder', '请选择')"
              >
                <template #prefix-icon><t-icon name="time" /></template>
              </t-select>
            </t-form-item>
            <t-form-item :label="t('settings.language')">
              <t-select
                v-model="form.config.locale"
                :options="localeOptions"
                :placeholder="t('common.selectPlaceholder', '请选择')"
              >
                <template #prefix-icon><t-icon name="translate" /></template>
              </t-select>
            </t-form-item>
          </div>
        </section>

        <!-- 部署提示 -->
        <section class="settings-section note-section">
          <div class="section-header">
            <t-icon name="info-circle" class="section-icon note-icon" />
            <h3 class="section-title">{{ t('settings.deploymentNote') }}</h3>
          </div>
          <p class="section-desc">{{ t('settings.deploymentNoteDesc') }}</p>
        </section>

        <!-- 底部操作栏（粘性定位） -->
        <div class="form-actions-bar">
          <div class="actions-inner">
            <div class="actions-left">
              <span v-if="hasChanges" class="unsaved-hint">
                <t-icon name="info-circle" size="14px" />
                {{ t('settings.unsavedChanges') }}
              </span>
            </div>
            <div class="actions-right">
              <t-button variant="outline" @click="goBack">
                <template #icon><t-icon name="chevron-left" /></template>
                {{ t('common.back') }}
              </t-button>
              <t-button theme="primary" :loading="saving" @click="handleSave">
                <template #icon v-if="!saving"><t-icon name="save" /></template>
                {{ saving ? t('settings.savingBtn') : t('settings.saveBtn') }}
              </t-button>
            </div>
          </div>
        </div>
      </t-form>
    </t-loading>
  </div>
</template>

<style scoped>
.site-config-page {
  max-width: 800px;
}

/* ====== 页头 ====== */
.page-header {
  margin-bottom: 24px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-bottom: 12px;
}

.breadcrumb-link {
  color: var(--td-brand-color, #3b82f6);
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.breadcrumb-icon {
  font-size: 12px;
  opacity: 0.4;
}

.breadcrumb-current {
  color: var(--color-text-secondary);
}

.header-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.site-avatar-sm {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.header-detail {
  min-width: 0;
}

.site-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0 0 4px;
}

.header-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.site-slug {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  color: var(--color-text-secondary);
  background: var(--td-bg-color-container-hover, rgba(0, 0, 0, 0.04));
  padding: 2px 8px;
  border-radius: 4px;
}

.meta-sep {
  color: var(--color-border);
  font-size: 12px;
}

.copy-id-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
  color: var(--color-text-secondary);
  background: none;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 2px 6px;
  cursor: pointer;
  transition: all 0.15s;
}

.copy-id-btn:hover {
  border-color: var(--td-brand-color, #3b82f6);
  color: var(--td-brand-color, #3b82f6);
}

/* ====== 表单区域 ====== */
.settings-form {
  display: flex;
  flex-direction: column;
}

.settings-section {
  padding: 24px 0;
  border-bottom: 1px solid var(--color-border);
}

.settings-section:last-of-type {
  border-bottom: none;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.section-icon {
  font-size: 18px;
  color: var(--td-brand-color, #3b82f6);
  opacity: 0.7;
}

.note-icon {
  color: var(--color-warning, #e6a23c);
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
}

.form-item-full {
  grid-column: 1 / -1;
}

/* Slug 变更标记 */
.slug-changed-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  color: var(--color-warning, #e6a23c);
  white-space: nowrap;
  font-weight: 500;
}

/* 提示信息卡片 */
.note-section .section-desc {
  padding: 14px 16px;
  font-size: 13px;
  color: var(--color-text-secondary);
  line-height: 1.65;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.04), rgba(59, 130, 246, 0.02));
  border-radius: 8px;
  border-left: 3px solid var(--td-brand-color, #3b82f6);
  margin: 0;
}

/* ====== 粘性操作栏 ====== */
.form-actions-bar {
  position: sticky;
  bottom: 0;
  z-index: 10;
  margin-top: 28px;
  padding: 16px 0 8px;
  background: var(--td-bg-color-container, #fff);
  border-top: 1px solid var(--component-stroke, rgba(0, 0, 0, 0.06));
}

.actions-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 800px;
}

.actions-left {
  min-width: 0;
}

.unsaved-hint {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--color-warning, #e6a23c);
}

.actions-right {
  display: flex;
  gap: 10px;
}
</style>
