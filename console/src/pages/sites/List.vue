<template>
  <div class="sites-page">
    <!-- 页头 -->
    <div class="page-header">
      <div class="title-section">
        <h1 class="page-title">{{ t('sites.title') }}</h1>
        <p class="page-subtitle">{{ t('sites.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <t-button variant="outline" @click="fetchSites">
          <template #icon><t-icon name="refresh" /></template>
          {{ t('common.refresh') }}
        </t-button>
        <t-button theme="primary" @click="openCreate">
          <template #icon><t-icon name="add" /></template>
          {{ t('sites.createSite') }}
        </t-button>
      </div>
    </div>

    <!-- 站点卡片列表 -->
    <div v-if="loading" class="loading-wrap">
      <t-loading size="large" />
    </div>

    <div v-else-if="sites.length === 0" class="empty-wrap">
      <t-empty :description="t('sites.noSites')" />
      <t-button theme="primary" style="margin-top: 16px;" @click="openCreate">
        {{ t('sites.createFirstSite') }}
      </t-button>
    </div>

    <div v-else class="sites-grid">
      <div
        v-for="site in sites"
        :key="site.id"
        class="site-card"
        :class="{ 'site-card--current': site.id === siteStore.currentSiteId }"
      >
        <!-- 当前站点标记 -->
        <div v-if="site.id === siteStore.currentSiteId" class="current-badge">
          {{ t('sites.current') }}
        </div>

        <!-- 卡片头部 -->
        <div class="site-card__header">
          <div class="site-avatar" :style="{ background: getSiteColor(site.name) }">
            {{ site.name.charAt(0).toUpperCase() }}
          </div>
          <div class="site-info">
            <h3 class="site-name">{{ site.name }}</h3>
            <span class="site-slug">{{ site.slug }}</span>
          </div>
          <div class="site-actions">
            <t-dropdown trigger="click" @click.stop>
              <t-button shape="square" variant="text" size="small">
                <template #icon><t-icon name="ellipsis" /></template>
              </t-button>
              <template #dropdown>
                <t-dropdown-menu>
                  <t-dropdown-item v-if="site.id !== siteStore.currentSiteId" @click="switchTo(site)">
                    <template #prefix-icon><t-icon name="swap" /></template>
                    {{ t('sites.switchTo') }}
                  </t-dropdown-item>
                  <t-dropdown-item @click="openEdit(site)">
                    <template #prefix-icon><t-icon name="edit" /></template>
                    {{ t('common.edit') }}
                  </t-dropdown-item>
                  <t-dropdown-item @click="router.push(`/sites/${site.id}/setting`)">
                    <template #prefix-icon><t-icon name="setting" /></template>
                    {{ t('menu.settings') }}
                  </t-dropdown-item>
                  <t-dropdown-item @click="router.push(`/sites/${site.id}/config`)">
                    <template #prefix-icon><t-icon name="tools" /></template>
                    {{ t('menu.configs') }}
                  </t-dropdown-item>
                  <t-dropdown-item @click="copySiteId(site)">
                    <template #prefix-icon><t-icon name="file-copy" /></template>
                    {{ t('sites.copySiteId') }}
                  </t-dropdown-item>
                  <t-divider />
                  <t-dropdown-item theme="error" :disabled="site.id === siteStore.currentSiteId" @click="confirmDelete(site)">
                    <template #prefix-icon><t-icon name="delete" /></template>
                    {{ t('common.delete') }}
                  </t-dropdown-item>
                </t-dropdown-menu>
              </template>
            </t-dropdown>
          </div>
        </div>

        <!-- 描述 -->
        <p class="site-desc">{{ site.description || t('sites.noDescription') }}</p>

        <!-- Site URL -->
        <div v-if="site.site_url" class="site-url">
          <t-icon name="link" size="14px" />
          <a :href="site.site_url" target="_blank" rel="noopener">{{ site.site_url }}</a>
        </div>

        <!-- Site ID 展示 -->
        <div class="site-id-row">
          <span class="site-id-label">Site ID</span>
          <code class="site-id-value">{{ site.id }}</code>
          <t-tooltip :content="t('sites.copySiteId')" placement="top">
            <t-button shape="square" variant="text" size="small" @click.stop="copySiteId(site)">
              <template #icon><t-icon name="file-copy" /></template>
            </t-button>
          </t-tooltip>
        </div>

        <!-- 底部元信息 -->
        <div class="site-meta">
          <span class="site-status" :class="site.is_active ? 'status--active' : 'status--inactive'">
            {{ site.is_active ? t('common.active') : t('common.inactive') }}
          </span>
          <span class="site-updated">
            {{ t('common.updatedAt') }}: {{ formatDate(site.updated_time) }}
          </span>
        </div>
      </div>
    </div>

    <!-- 创建/编辑弹窗 -->
    <t-dialog
      v-model:visible="showDialog"
      :header="isEditing ? t('sites.editSite') : t('sites.createSite')"
      :close-on-overlay-click="true"
      :close-on-esc-keydown="true"
      :destroy-on-close="true"
      width="520px"
      @close="closeDialog"
    >
      <t-form
        ref="formRef"
        :data="form"
        :rules="formRules"
        layout="vertical"
        label-align="top"
        :required-mark="false"
      >
        <t-form-item :label="t('sites.form.name')" name="name">
          <t-input
            v-model="form.name"
            :placeholder="t('sites.form.namePlaceholder')"
            :maxlength="200"
            @change="onNameChange"
          />
        </t-form-item>

        <t-form-item :label="t('sites.form.slug')" name="slug">
          <t-input
            v-model="form.slug"
            :placeholder="t('sites.form.slugPlaceholder')"
            :maxlength="100"
          >
            <template #tips>
              <span class="form-tip">{{ t('sites.form.slugTip') }}</span>
            </template>
          </t-input>
        </t-form-item>

        <t-form-item :label="t('sites.form.description')" name="description">
          <t-textarea
            v-model="form.description"
            :placeholder="t('sites.form.descPlaceholder')"
            :maxlength="2000"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
        </t-form-item>

        <t-form-item :label="t('sites.form.siteUrl')" name="site_url">
          <t-input
            v-model="form.site_url"
            placeholder="https://example.com"
          >
            <template #prefix-icon><t-icon name="link" /></template>
          </t-input>
        </t-form-item>

        <t-form-item v-if="isEditing" :label="t('sites.form.status')" name="is_active">
          <t-switch v-model="form.is_active" />
          <span style="margin-left: 8px; font-size: 13px; color: var(--td-text-color-secondary)">
            {{ form.is_active ? t('common.active') : t('common.inactive') }}
          </span>
        </t-form-item>
      </t-form>

      <template #footer>
        <div class="dialog-footer">
          <t-button variant="outline" @click="closeDialog">{{ t('common.cancel') }}</t-button>
          <t-button theme="primary" :loading="submitting" @click="onSubmit">
            {{ isEditing ? t('common.save') : t('common.create') }}
          </t-button>
        </div>
      </template>
    </t-dialog>

    <!-- 删除确认弹窗 -->
    <t-dialog
      v-model:visible="showDeleteDialog"
      :header="t('sites.deleteConfirm')"
      :confirm-btn="{ content: t('common.delete'), theme: 'danger', loading: deleting }"
      :cancel-btn="t('common.cancel')"
      width="420px"
      @confirm="doDelete"
      @close="showDeleteDialog = false"
    >
      <p>{{ t('sites.deleteMsg', { name: deletingTarget?.name }) }}</p>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { MessagePlugin } from 'tdesign-vue-next'
import {
  getMySites,
  createSite,
  updateSite,
  deleteSite,
  type Site,
} from '@/api/site'
import { useSiteStore } from '@/stores/site'
import { showError } from '@/utils/request'

const { t } = useI18n()
const router = useRouter()
const siteStore = useSiteStore()

// ── 列表 ──────────────────────────────────────────────────────────────────────
const sites = ref<Site[]>([])
const loading = ref(false)
const total = ref(0)

async function fetchSites() {
  loading.value = true
  try {
    const res = await getMySites({ page: 1, page_size: 100 })
    if (res.code === 200) {
      sites.value = res.data?.items || []
      total.value = res.data?.total || 0
    }
  } catch (e) {
    showError(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchSites)

// ── 切换当前站点 ────────────────────────────────────────────────────────────
function switchTo(site: Site) {
  siteStore.setCurrentSite(site.id)
  MessagePlugin.success(t('sites.switchedTo', { name: site.name }))
}

// ── 复制 Site ID ───────────────────────────────────────────────────────────
async function copySiteId(site: Site) {
  try {
    await navigator.clipboard.writeText(site.id)
    MessagePlugin.success(t('sites.siteIdCopied'))
  } catch {
    MessagePlugin.error(t('sites.copyFailed'))
  }
}

// ── 表单弹窗 ───────────────────────────────────────────────────────────────
const showDialog = ref(false)
const isEditing = ref(false)
const submitting = ref(false)
const editingId = ref('')
const formRef = ref()

const form = reactive({
  name: '',
  slug: '',
  description: '',
  site_url: '',
  is_active: true,
})

const formRules = {
  name: [{ required: true, message: t('sites.form.nameRequired'), trigger: 'blur' as const }],
  slug: [
    { required: true, message: t('sites.form.slugRequired'), trigger: 'blur' as const },
    {
      validator: (val: string) => /^[a-z][a-z0-9-]{0,98}[a-z0-9]$/.test(val) || val.length >= 2,
      message: t('sites.form.slugFormat'),
      trigger: 'blur' as const,
    },
  ],
}

function generateSlug(name: string) {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fff]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 100) || ''
}

function onNameChange(val: string) {
  if (!isEditing.value || !form.slug) {
    form.slug = generateSlug(String(val || ''))
  }
}

function openCreate() {
  isEditing.value = false
  editingId.value = ''
  form.name = ''
  form.slug = ''
  form.description = ''
  form.site_url = ''
  form.is_active = true
  showDialog.value = true
}

function openEdit(site: Site) {
  isEditing.value = true
  editingId.value = site.id
  form.name = site.name
  form.slug = site.slug
  form.description = site.description || ''
  form.site_url = site.site_url || ''
  form.is_active = site.is_active
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
}

async function onSubmit() {
  const valid = await formRef.value?.validate()
  if (valid !== true) return

  submitting.value = true
  try {
    if (isEditing.value) {
      const res = await updateSite(editingId.value, {
        name: form.name.trim(),
        slug: form.slug.trim(),
        description: form.description.trim() || undefined,
        site_url: form.site_url.trim() || undefined,
        is_active: form.is_active,
      })
      if (res.code === 200) {
        MessagePlugin.success(t('sites.updateSuccess'))
        closeDialog()
        await fetchSites()
        // 同步 store 中的站点列表
        await siteStore.fetchSites()
      } else {
        MessagePlugin.error(res.msg || t('sites.updateFailed'))
      }
    } else {
      const res = await createSite({
        name: form.name.trim(),
        slug: form.slug.trim(),
        description: form.description.trim() || undefined,
      })
      if (res.code === 200) {
        MessagePlugin.success(t('sites.createSuccess'))
        // 自动切换到新建站点
        if (res.data?.id) {
          siteStore.setCurrentSite(res.data.id)
        }
        closeDialog()
        await fetchSites()
        await siteStore.fetchSites()
      } else {
        MessagePlugin.error(res.msg || t('sites.createFailed'))
      }
    }
  } catch (e) {
    showError(e)
  } finally {
    submitting.value = false
  }
}

// ── 删除 ──────────────────────────────────────────────────────────────────────
const showDeleteDialog = ref(false)
const deleting = ref(false)
const deletingTarget = ref<Site | null>(null)

function confirmDelete(site: Site) {
  deletingTarget.value = site
  showDeleteDialog.value = true
}

async function doDelete() {
  if (!deletingTarget.value) return
  deleting.value = true
  try {
    await deleteSite(deletingTarget.value.id)
    MessagePlugin.success(t('sites.deleteSuccess'))
    showDeleteDialog.value = false
    deletingTarget.value = null
    await fetchSites()
    await siteStore.fetchSites()
  } catch (e) {
    showError(e)
  } finally {
    deleting.value = false
  }
}

// ── 辅助 ──────────────────────────────────────────────────────────────────────
const PALETTE = [
  '#3b82f6', '#10b981', '#8b5cf6', '#f59e0b',
  '#ef4444', '#ec4899', '#06b6d4', '#84cc16',
]
function getSiteColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return PALETTE[Math.abs(hash) % PALETTE.length]
}

function formatDate(dt: string) {
  return new Date(dt).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}
</script>

<style scoped>
.sites-page {
  width: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.title-section .page-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 4px 0;
}

.page-subtitle {
  color: var(--color-text-secondary);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 80px 0;
}

.empty-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80px 0;
}

/* 卡片网格 */
.sites-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

/* 单张卡片 */
.site-card {
  position: relative;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.site-card:hover {
  border-color: var(--td-brand-color, #3b82f6);
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.1);
}

.site-card--current {
  border-color: var(--td-brand-color, #3b82f6);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.12);
}

.current-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  background: var(--td-brand-color, #3b82f6);
  color: #fff;
  border-radius: 999px;
}

/* 卡片头部 */
.site-card__header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 12px;
}

.site-avatar {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.site-info {
  flex: 1;
  min-width: 0;
}

.site-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin: 0 0 2px;
}

.site-slug {
  font-size: 12px;
  color: var(--color-text-secondary);
  font-family: monospace;
}

.site-actions {
  flex-shrink: 0;
  margin-left: auto;
}

/* 描述 */
.site-desc {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: 0 0 10px;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 40px;
}

/* URL */
.site-url {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--td-brand-color, #3b82f6);
  margin-bottom: 10px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.site-url a {
  color: inherit;
  text-decoration: none;
}

.site-url a:hover {
  text-decoration: underline;
}

/* Site ID 行 */
.site-id-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  background: var(--color-hover);
  border-radius: 6px;
  margin-bottom: 12px;
}

.site-id-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.site-id-value {
  flex: 1;
  font-family: monospace;
  font-size: 11px;
  color: var(--color-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 底部元信息 */
.site-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.site-status {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 999px;
}

.status--active {
  background: var(--color-success-light);
  color: var(--color-success);
}

.status--inactive {
  background: var(--color-hover);
  color: var(--color-text-secondary);
}

.site-updated {
  font-size: 12px;
  color: var(--color-text-secondary);
}

/* 弹窗底部 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.form-tip {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
</style>
