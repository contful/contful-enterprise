<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { DialogPlugin } from 'tdesign-vue-next'
import { getConfigs, setConfig, deleteConfig, type SiteConfig } from '@/api/config'
import { showSuccess, showError } from '@/utils/request'

const { t } = useI18n()
const route = useRoute()

const siteId = computed(() => route.params.siteId as string)

// ============ 状态 ============
const loading = ref(false)
const saving = ref<string | null>(null)
const deleting = ref<string | null>(null)

const configs = ref<SiteConfig[]>([])

// 分组
const activeGroup = ref('all')
const groups = computed(() => {
  const gs = new Set(configs.value.map(c => c.config_group))
  return ['all', ...Array.from(gs)]
})

const filteredConfigs = computed(() => {
  if (activeGroup.value === 'all') return configs.value
  return configs.value.filter(c => c.config_group === activeGroup.value)
})

// ============ 编辑弹窗 (t-dialog) ============
const showEditModal = ref(false)
const editForm = ref({
  config_key: '',
  config_value: '',
  config_type: '',
  description: '',
  is_readonly: false,
})
const editError = ref('')

// 新增弹窗
const showAddModal = ref(false)
const addForm = ref({
  config_key: '',
  config_value: '',
  config_type: 'string',
  description: '',
})
const addError = ref('')

// ============ 加载数据 ============
async function loadConfigs() {
  if (!siteId.value) return
  loading.value = true
  try {
    const res = await getConfigs(siteId.value)
    configs.value = res || []
  } catch (e: any) {
    showError(e)
  } finally {
    loading.value = false
  }
}

// ============ 编辑 ============
function startEdit(cfg: SiteConfig) {
  if (cfg.is_readonly) return
  editForm.value = {
    config_key: cfg.config_key,
    config_value: cfg.config_value,
    config_type: cfg.config_type,
    description: cfg.description || '',
    is_readonly: cfg.is_readonly,
  }
  editError.value = ''
  showEditModal.value = true
}

async function saveEdit() {
  if (!siteId.value || !saving.value) return
  editError.value = ''
  try {
    await setConfig(siteId.value, editForm.value.config_key, {
      config_value: editForm.value.config_value,
      config_type: editForm.value.config_type,
    })
    const cfg = configs.value.find(c => c.config_key === editForm.value.config_key)
    if (cfg) cfg.config_value = editForm.value.config_value
    showEditModal.value = false
    showSuccess(t('common.saveSuccess'))
  } catch (e: any) {
    editError.value = e.message || String(e)
  } finally {
    saving.value = null
  }
}

// ============ 删除 — DialogPlugin.confirm ============
function handleDelete(cfg: SiteConfig) {
  if (!siteId.value || deleting.value || cfg.config_group === 'integrity') return
  DialogPlugin.confirm({
    header: t('common.confirmDelete'),
    body: t('settings.confirmDelete', { key: cfg.config_key }),
    theme: 'warning',
    onConfirm: async () => {
      deleting.value = cfg.config_key
      try {
        await deleteConfig(siteId.value, cfg.config_key)
        configs.value = configs.value.filter(c => c.id !== cfg.id)
        showSuccess(t('common.deleteSuccess'))
      } catch (e: any) {
        showError(e)
      } finally {
        deleting.value = null
      }
    },
  })
}

// ============ 新增 ============
function openAdd() {
  addForm.value = { config_key: '', config_value: '', config_type: 'string', description: '' }
  addError.value = ''
  showAddModal.value = true
}

async function handleAdd() {
  if (!siteId.value) return
  if (!addForm.value.config_key.trim()) {
    addError.value = t('settings.keyRequired')
    return
  }
  saving.value = 'add'
  addError.value = ''
  try {
    const res = await setConfig(siteId.value, addForm.value.config_key, {
      config_value: addForm.value.config_value,
      config_type: addForm.value.config_type,
      config_group: 'default',
      description: addForm.value.description,
    })
    if ((res as any).data) {
      configs.value.push((res as any).data)
    }
    showAddModal.value = false
    showSuccess(t('common.createSuccess'))
  } catch (e: any) {
    addError.value = e.message || String(e)
  } finally {
    saving.value = null
  }
}

// ============ 辅助 ============
function formatValue(cfg: SiteConfig): string {
  if (cfg.is_encrypted) return '\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022'
  if (cfg.config_type === 'boolean') return cfg.config_value
  if (cfg.config_type === 'json') {
    try { return JSON.stringify(JSON.parse(cfg.config_value), null, 2) } catch { return cfg.config_value }
  }
  return cfg.config_value
}

function typeLabel(type_: string): string {
  const map: Record<string, string> = {
    string: t('settings.typeString'),
    number: t('settings.typeNumber'),
    boolean: t('settings.typeBoolean'),
    json: t('settings.typeJson'),
    encrypted: t('settings.typeEncrypted'),
  }
  return map[type_] || type_
}

onMounted(loadConfigs)
</script>

<template>
  <div class="configs-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('settings.configs') }}</h1>
        <p class="page-subtitle">{{ t('settings.configsSubtitle') }}</p>
      </div>
      <t-button theme="primary" @click="openAdd">
        {{ t('settings.addConfig') }}
      </t-button>
    </div>

    <!-- 分组切换 -->
    <div class="group-tabs" v-if="groups.length > 2">
      <t-radio-group v-model="activeGroup" variant="default-filled" size="small">
        <t-radio-button
          v-for="g in groups"
          :key="g"
          :value="g"
        >
          {{ g === 'all' ? t('settings.allGroups') : g }}
        </t-radio-button>
      </t-radio-group>
    </div>

    <!-- 配置列表 — t-table -->
    <div v-loading="loading" class="config-table-wrap">
      <table v-if="!loading && filteredConfigs.length > 0" class="table">
        <thead>
          <tr>
            <th>{{ t('settings.key') }}</th>
            <th>{{ t('settings.value') }}</th>
            <th>{{ t('settings.type') }}</th>
            <th>{{ t('settings.description') }}</th>
            <th style="width: 120px">{{ t('settings.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="cfg in filteredConfigs" :key="cfg.id">
            <tr :class="{ 'row-readonly': cfg.is_readonly }">
              <td class="cell-key">
                <span class="key-name">{{ cfg.config_key }}</span>
                <t-tag v-if="cfg.is_encrypted" theme="success" variant="light" size="small">\u{1F512} {{ t('settings.encrypted') }}</t-tag>
                <t-tag v-if="cfg.is_readonly" theme="warning" variant="light" size="small">{{ t('settings.readonly') }}</t-tag>
              </td>
              <td class="cell-value">
                <span class="value-text" :class="{ 'value-masked': cfg.is_encrypted }">{{ formatValue(cfg) }}</span>
              </td>
              <td><t-tag theme="default" variant="light" size="small">{{ typeLabel(cfg.config_type) }}</t-tag></td>
              <td class="cell-desc">{{ cfg.description || '\u2014' }}</td>
              <td>
                <div class="row-actions">
                  <t-button
                    v-if="!cfg.is_readonly"
                    variant="outline"
                    size="small"
                    :disabled="!!saving"
                    @click="startEdit(cfg)"
                  >
                    {{ t('common.edit') }}
                  </t-button>
                  <t-button
                    v-if="cfg.config_group !== 'integrity'"
                    theme="danger"
                    variant="outline"
                    size="small"
                    :disabled="deleting === cfg.config_key"
                    :loading="deleting === cfg.config_key"
                    @click="handleDelete(cfg)"
                  >
                    {{ t('common.delete') }}
                  </t-button>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>

      <!-- 空状态 -->
      <t-empty v-else-if="!loading && filteredConfigs.length === 0" :description="t('settings.noConfigs')">
        <template #action>
          <t-button variant="outline" @click="openAdd">{{ t('settings.addFirstConfig') }}</t-button>
        </template>
      </t-empty>
    </div>

    <!-- 新增弹窗 — t-dialog + t-form -->
    <t-dialog
      v-model:visible="showAddModal"
      :header="t('settings.addConfig')"
      :width="480"
      :confirm-btn="{ content: t('common.create'), theme: 'primary' as const, loading: saving === 'add' }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleAdd"
    >
      <t-form :data="addForm" label-align="top">
        <t-form-item :label="`${t('settings.key')} *`">
          <t-input v-model="addForm.config_key" :placeholder="t('settings.keyPlaceholder')" clearable />
        </t-form-item>
        <t-form-item :label="t('settings.value')">
          <textarea v-model="addForm.config_value" class="config-textarea" rows="3" :placeholder="t('settings.valuePlaceholder')" />
        </t-form-item>
        <t-form-item :label="t('settings.type')">
          <t-select v-model="addForm.config_type" :options="[
            { label: t('settings.typeString'), value: 'string' },
            { label: t('settings.typeNumber'), value: 'number' },
            { label: t('settings.typeBoolean'), value: 'boolean' },
            { label: t('settings.typeJson'), value: 'json' },
          ]" />
        </t-form-item>
        <t-form-item :label="t('settings.description')">
          <t-input v-model="addForm.description" :placeholder="t('settings.descPlaceholder')" clearable />
        </t-form-item>
        <t-alert v-if="addError" theme="error" :message="addError" closable @close="addError = ''" />
      </t-form>
    </t-dialog>

    <!-- 编辑弹窗 — t-dialog + t-form -->
    <t-dialog
      v-model:visible="showEditModal"
      :header="`${t('common.edit')}: ${editForm.config_key}`"
      :width="480"
      :confirm-btn="{ content: t('common.save'), theme: 'primary' as const, loading: !!saving }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="saveEdit"
    >
      <t-form :data="editForm" label-align="top">
        <t-form-item :label="t('settings.type')">
          <t-tag theme="default" variant="light" size="small">{{ typeLabel(editForm.config_type) }}</t-tag>
        </t-form-item>
        <t-form-item :label="t('settings.value')">
          <textarea v-model="editForm.config_value" class="config-textarea" :rows="editForm.config_type === 'json' ? 6 : 3" :placeholder="t('settings.valuePlaceholder')" />
        </t-form-item>
        <t-form-item :label="t('settings.description')">
          <p class="desc-text">{{ editForm.description || '\u2014' }}</p>
        </t-form-item>
        <t-alert v-if="editError" theme="error" :message="editError" closable @close="editError = ''" />
      </t-form>
    </t-dialog>
  </div>
</template>

<style scoped>
.configs-page {
  height: 100%;
}

/* 表格行高统一 */
:deep(.table) td,
:deep(.table) th {
  vertical-align: top !important;
  padding: 10px 12px !important;
}

:deep(.table) td {
  min-height: 48px !important;
}

:deep(.table .row-actions) {
  white-space: nowrap;
}

.group-tabs {
  margin-bottom: 16px;
}

.config-table-wrap {
  position: relative;
  min-height: 200px;
}

.cell-key {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 160px;
}

.key-name {
  font-family: monospace;
  font-size: 13px;
  color: var(--color-text);
  font-weight: 500;
}

.cell-value {
  min-width: 200px;
}

.value-text {
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
  color: var(--color-text-secondary);
}

.value-masked {
  color: var(--color-text-hint);
  font-style: italic;
}

.desc-text {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin: 0;
}

.cell-desc {
  font-size: 12px;
  color: var(--color-text-secondary);
  max-width: 200px;
}

.row-readonly {
  background: var(--color-bg-secondary);
  opacity: 0.7;
}

.row-actions {
  display: flex;
  gap: 4px;
}

/* textarea 样式 */
.config-textarea {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-size: 14px;
  line-height: 1.5;
  color: var(--color-text);
  background: var(--color-bg-white);
  resize: vertical;
  box-sizing: border-box;
  font-family: inherit;
}

.config-textarea:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px rgba(var(--color-primary-rgb, 22, 119, 255), 0.15);
}
</style>
