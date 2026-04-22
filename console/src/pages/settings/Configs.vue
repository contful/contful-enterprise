<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { getConfigs, setConfig, deleteConfig, type SiteConfig } from '@/api/config'
import { showSuccess, showError } from '@/utils/request'

const { t } = useI18n()
const siteStore = useSiteStore()

// ============ 状态 ============
const loading = ref(false)
const saving = ref<string | null>(null)
const deleting = ref<string | null>(null)

// 原始配置列表
const configs = ref<SiteConfig[]>([])

// 分组
const activeGroup = ref('all')
const groups = computed(() => {
  const gs = new Set(configs.value.map(c => c.config_group))
  return ['all', ...Array.from(gs)]
})

// 过滤后的配置
const filteredConfigs = computed(() => {
  if (activeGroup.value === 'all') return configs.value
  return configs.value.filter(c => c.config_group === activeGroup.value)
})

// 按 key 分组展示
const grouped = computed(() => {
  const map = new Map<string, SiteConfig[]>()
  for (const c of filteredConfigs.value) {
    const list = map.get(c.config_key) || []
    list.push(c)
    map.set(c.config_key, list)
  }
  return map
})

// ============ 编辑状态 ============
const editingKey = ref<string | null>(null)
const editValue = ref('')
const editError = ref('')

// 新增弹窗
const showAddModal = ref(false)
const addForm = ref({
  config_key: '',
  config_value: '',
  config_type: 'string',
  config_group: 'default',
  description: '',
})
const addError = ref('')

// ============ 加载数据 ============
async function loadConfigs() {
  if (!siteStore.currentSiteId) return
  loading.value = true
  try {
    const res = await getConfigs(siteStore.currentSiteId)
    configs.value = res.items || []
  } catch (e: any) {
    showError(e)
  } finally {
    loading.value = false
  }
}

// ============ 编辑 ============
function startEdit(cfg: SiteConfig) {
  if (cfg.is_readonly) return
  editingKey.value = cfg.config_key
  editValue.value = cfg.config_value
  editError.value = ''
}

function cancelEdit() {
  editingKey.value = null
  editValue.value = ''
  editError.value = ''
}

async function saveEdit(cfg: SiteConfig) {
  if (!siteStore.currentSiteId || saving.value) return
  saving.value = cfg.config_key
  editError.value = ''
  try {
    await setConfig(siteStore.currentSiteId, cfg.config_key, editValue.value, cfg.config_type)
    cfg.config_value = editValue.value
    editingKey.value = null
    showSuccess(t('common.saveSuccess'))
  } catch (e: any) {
    editError.value = e.message || String(e)
  } finally {
    saving.value = null
  }
}

// ============ 删除 ============
async function handleDelete(cfg: SiteConfig) {
  if (!siteStore.currentSiteId || deleting.value || cfg.is_readonly) return
  if (!confirm(t('settings.confirmDelete', { key: cfg.config_key }))) return
  deleting.value = cfg.config_key
  try {
    await deleteConfig(siteStore.currentSiteId, cfg.config_key)
    configs.value = configs.value.filter(c => c.id !== cfg.id)
    showSuccess(t('common.deleteSuccess'))
  } catch (e: any) {
    showError(e)
  } finally {
    deleting.value = null
  }
}

// ============ 新增 ============
function openAdd() {
  addForm.value = {
    config_key: '',
    config_value: '',
    config_type: 'string',
    config_group: activeGroup.value === 'all' ? 'default' : activeGroup.value,
    description: '',
  }
  addError.value = ''
  showAddModal.value = true
}

async function handleAdd() {
  if (!siteStore.currentSiteId) return
  if (!addForm.value.config_key.trim()) {
    addError.value = t('settings.keyRequired')
    return
  }
  saving.value = 'add'
  addError.value = ''
  try {
    const res = await setConfig(
      siteStore.currentSiteId,
      addForm.value.config_key,
      addForm.value.config_value,
      addForm.value.config_type,
    )
    // 添加到列表
    if (res.data) {
      configs.value.push(res.data)
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
  if (cfg.is_encrypted) return '••••••••'
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
      <button class="btn btn-primary" @click="openAdd">
        {{ t('settings.addConfig') }}
      </button>
    </div>

    <!-- 分组切换 -->
    <div class="group-tabs" v-if="groups.length > 2">
      <button
        v-for="g in groups"
        :key="g"
        class="group-tab"
        :class="{ active: activeGroup === g }"
        @click="activeGroup = g"
      >
        {{ g === 'all' ? t('settings.allGroups') : g }}
      </button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="state-loading">
      {{ t('common.loading') }}
    </div>

    <!-- 空状态 -->
    <div v-else-if="filteredConfigs.length === 0" class="state-empty">
      <p>{{ t('settings.noConfigs') }}</p>
      <button class="btn btn-default" @click="openAdd">{{ t('settings.addFirstConfig') }}</button>
    </div>

    <!-- 配置列表 -->
    <div v-else class="card">
      <table class="table">
        <thead>
          <tr>
            <th>{{ t('settings.key') }}</th>
            <th>{{ t('settings.value') }}</th>
            <th>{{ t('settings.type') }}</th>
            <th>{{ t('settings.description') }}</th>
            <th style="width: 100px">{{ t('settings.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="cfg in filteredConfigs" :key="cfg.id">
            <tr :class="{ 'row-readonly': cfg.is_readonly }">
              <td class="cell-key">
                <span class="key-name">{{ cfg.config_key }}</span>
                <span v-if="cfg.is_encrypted" class="badge badge-info">🔒 {{ t('settings.encrypted') }}</span>
                <span v-if="cfg.is_readonly" class="badge badge-warning">{{ t('settings.readonly') }}</span>
              </td>
              <td class="cell-value">
                <!-- 编辑状态 -->
                <template v-if="editingKey === cfg.config_key">
                  <div class="edit-row">
                    <textarea
                      v-model="editValue"
                      class="input"
                      :rows="cfg.config_type === 'json' ? 4 : 1"
                      style="width: 300px"
                    />
                    <div class="edit-actions">
                      <button class="btn btn-sm btn-primary" :disabled="saving === cfg.config_key" @click="saveEdit(cfg)">
                        {{ t('common.save') }}
                      </button>
                      <button class="btn btn-sm btn-default" @click="cancelEdit">{{ t('common.cancel') }}</button>
                    </div>
                    <p v-if="editError" class="field-error">{{ editError }}</p>
                  </div>
                </template>
                <!-- 只读/加密 -->
                <template v-else>
                  <span class="value-text" :class="{ 'value-masked': cfg.is_encrypted }">{{ formatValue(cfg) }}</span>
                </template>
              </td>
              <td><span class="type-label">{{ typeLabel(cfg.config_type) }}</span></td>
              <td class="cell-desc">{{ cfg.description || '—' }}</td>
              <td>
                <div class="row-actions">
                  <button
                    v-if="!cfg.is_readonly"
                    class="btn btn-sm btn-default"
                    :disabled="!!saving"
                    @click="startEdit(cfg)"
                  >
                    {{ t('common.edit') }}
                  </button>
                  <button
                    v-if="!cfg.is_readonly"
                    class="btn btn-sm btn-danger"
                    :disabled="deleting === cfg.config_key"
                    @click="handleDelete(cfg)"
                  >
                    {{ t('common.delete') }}
                  </button>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <!-- 新增弹窗 -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ t('settings.addConfig') }}</h3>
          <button class="modal-close" @click="showAddModal = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>{{ t('settings.key') }} *</label>
            <input v-model="addForm.config_key" class="input" :placeholder="t('settings.keyPlaceholder')" />
          </div>
          <div class="form-group">
            <label>{{ t('settings.value') }}</label>
            <textarea v-model="addForm.config_value" class="input" rows="3" :placeholder="t('settings.valuePlaceholder')" />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>{{ t('settings.type') }}</label>
              <select v-model="addForm.config_type" class="input">
                <option value="string">{{ t('settings.typeString') }}</option>
                <option value="number">{{ t('settings.typeNumber') }}</option>
                <option value="boolean">{{ t('settings.typeBoolean') }}</option>
                <option value="json">{{ t('settings.typeJson') }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>{{ t('settings.group') }}</label>
              <input v-model="addForm.config_group" class="input" :placeholder="t('settings.groupPlaceholder')" />
            </div>
          </div>
          <div class="form-group">
            <label>{{ t('settings.description') }}</label>
            <input v-model="addForm.description" class="input" :placeholder="t('settings.descPlaceholder')" />
          </div>
          <p v-if="addError" class="field-error">{{ addError }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-default" @click="showAddModal = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="!!saving" @click="handleAdd">
            {{ t('common.create') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.configs-page {
  height: 100%;
}

.group-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.group-tab {
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid var(--color-border);
  background: transparent;
  font-size: 13px;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.group-tab:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.group-tab.active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: #fff;
}

.state-loading,
.state-empty {
  padding: 48px;
  text-align: center;
  color: var(--color-text-secondary);
}

.state-empty p {
  margin-bottom: 16px;
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

.edit-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.edit-actions {
  display: flex;
  gap: 4px;
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

.type-label {
  font-size: 12px;
  padding: 2px 8px;
  background: var(--color-bg-secondary);
  border-radius: 4px;
  color: var(--color-text-secondary);
  font-family: monospace;
}

.badge {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.badge-info {
  background: #e8f5e9;
  color: #2e7d32;
}

.badge-warning {
  background: #fff3e0;
  color: #e65100;
}

/* 弹窗 */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: #fff;
  border-radius: 12px;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--color-border);
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--color-text-secondary);
  cursor: pointer;
  line-height: 1;
}

.modal-body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 24px;
  border-top: 1px solid var(--color-border);
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-row .form-group {
  flex: 1;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
}

.field-error {
  color: var(--color-danger, #e53935);
  font-size: 12px;
  margin: 0;
}
</style>
