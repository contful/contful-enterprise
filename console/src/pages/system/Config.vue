<template>
  <!-- 页面标题 -->
  <PageHeader
    :title="t('settings.systemConfig')"
    :subtitle="t('settings.systemConfigDesc')"
  />

  <!-- 工具栏 -->
  <div class="toolbar">
    <t-input
      v-model="keyword"
      :placeholder="t('common.searchPlaceholder')"
      clearable
      style="width: 280px"
    >
      <template #prefix-icon>
        <t-icon name="search" />
      </template>
    </t-input>
    <t-space>
      <t-button variant="outline" :loading="clearingCache" theme="warning" @click="handleClearCache">
        {{ t('settings.clearCache') }}
      </t-button>
      <t-button variant="outline" @click="loadConfigs">
        <template #icon><t-icon name="refresh" /></template>
        {{ t('common.refresh') }}
      </t-button>
      <t-button theme="primary" @click="openCreateDialog">
        <template #icon><t-icon name="add" /></template>
        {{ t('settings.addConfig') }}
      </t-button>
    </t-space>
  </div>

  <!-- 配置列表 -->
  <t-table
    :data="filteredConfigs"
    :columns="columns"
    :loading="loading"
    row-key="config_key"
    hover
    stripe
    size="medium"
  >
    <template #keyCell="{ row }">
      <div style="display: flex; align-items: center; gap: 8px">
        <code class="config-key">{{ row.config_key }}</code>
        <t-tag v-if="row.is_system" theme="primary" variant="light" size="small">
          {{ t('settings.systemConfigTag') }}
        </t-tag>
        <t-tag v-else theme="default" variant="light" size="small">
          {{ t('settings.customConfigTag') }}
        </t-tag>
      </div>
    </template>

    <template #valueCell="{ row }">
      <template v-if="row.value_type === 'boolean'">
        <t-tag :theme="row.config_value === 'true' ? 'success' : 'danger'" variant="light" size="small">
          {{ row.config_value }}
        </t-tag>
      </template>
      <template v-else-if="row.config_value === ''">
        <span class="empty-value">—</span>
      </template>
      <template v-else>
        <span class="config-value">{{ row.config_value }}</span>
      </template>
    </template>

    <template #typeCell="{ row }">
      <t-tag variant="outline" size="small">{{ row.value_type }}</t-tag>
    </template>

    <template #publicCell="{ row }">
      <t-switch v-model="row.is_public" size="small" :disabled="saving" @change="(val: boolean) => handleTogglePublic(row, val)" />
    </template>

    <template #operations="{ row }">
      <t-space size="small">
        <t-button variant="text" theme="primary" size="small" @click="openEditDialog(row)">
          {{ t('common.edit') }}
        </t-button>
        <t-button
          v-if="!row.is_system"
          variant="text"
          theme="danger"
          size="small"
          @click="openDeleteDialog(row)"
        >
          {{ t('common.delete') }}
        </t-button>
      </t-space>
    </template>
  </t-table>

  <!-- 创建/编辑配置弹窗 -->
  <t-dialog
    v-model:visible="dialogVisible"
    :header="editingConfig ? t('settings.editConfig') : t('settings.createConfig')"
    :width="560"
    :confirm-btn="{ content: saving ? t('common.saving') : t('common.save'), theme: 'primary' as const, loading: saving }"
    :cancel-btn="{ content: t('common.cancel') }"
    @confirm="handleSave"
  >
    <t-form :data="form" label-align="top">
      <t-form-item :label="`${t('settings.configKey')} *`" v-if="!editingConfig">
        <t-input
          v-model="form.config_key"
          :placeholder="t('settings.configKeyPlaceholder')"
          clearable
        />
      </t-form-item>
      <t-form-item v-else :label="t('settings.configKey')">
        <t-input :value="form.config_key" disabled />
      </t-form-item>

      <t-form-item :label="t('settings.configValue')" :help="form.description">
        <t-input
          v-if="form.value_type === 'string'"
          v-model="form.config_value"
          :placeholder="t('settings.configValuePlaceholder')"
        />
        <t-input-number
          v-else-if="form.value_type === 'number'"
          v-model="form.config_value_number"
          :placeholder="t('settings.configValuePlaceholder')"
          style="width: 100%"
        />
        <t-switch
          v-else-if="form.value_type === 'boolean'"
          v-model="form.config_value_bool"
        />
        <t-textarea
          v-else-if="form.value_type === 'json'"
          v-model="form.config_value"
          :placeholder="t('settings.enterJson')"
          :autosize="{ minRows: 3, maxRows: 10 }"
        />
      </t-form-item>

      <t-form-item :label="t('settings.valueType')">
        <t-select v-model="form.value_type" :disabled="!!editingConfig">
          <t-option value="string" label="String" />
          <t-option value="number" label="Number" />
          <t-option value="boolean" label="Boolean" />
          <t-option value="json" label="JSON" />
        </t-select>
      </t-form-item>

      <t-form-item :label="t('settings.description')">
        <t-textarea
          v-model="form.description"
          :placeholder="t('settings.enterDescription')"
          :autosize="{ minRows: 2, maxRows: 4 }"
        />
      </t-form-item>

      <t-form-item :label="t('settings.isPublic')">
        <t-switch v-model="form.is_public" :disabled="!!(editingConfig?.is_system)" />
      </t-form-item>
    </t-form>
  </t-dialog>

  <!-- 删除确认弹窗 -->
  <t-dialog
    v-model:visible="deleteVisible"
    :header="t('common.confirmDelete')"
    theme="danger"
    :confirm-btn="{ content: deleting ? t('common.deleting') : t('common.delete'), theme: 'danger' as const, loading: deleting }"
    :cancel-btn="{ content: t('common.cancel') }"
    @confirm="handleDelete"
  >
    <p>{{ t('settings.deleteConfigConfirm', { key: deletingConfig?.config_key }) }}</p>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import PageHeader from '@/components/PageHeader.vue'
import { getSystemConfigs, updateSystemConfig, createSystemConfig, deleteSystemConfig, clearSystemConfigCache } from '@/api/system-config'
import type { SystemConfig } from '@/types/system/config'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const clearingCache = ref(false)
const configs = ref<SystemConfig[]>([])
const keyword = ref('')
const dialogVisible = ref(false)
const deleteVisible = ref(false)
const editingConfig = ref<SystemConfig | null>(null)
const deletingConfig = ref<SystemConfig | null>(null)

const form = reactive({
  config_key: '',
  config_value: '',
  config_value_bool: false,
  config_value_number: 0,
  value_type: 'string' as string,
  description: '',
  is_public: false,
})

const columns = [
  { colKey: 'config_key', title: t('settings.configKey'), cell: 'keyCell', width: 280 },
  { colKey: 'value', title: t('settings.configValue'), cell: 'valueCell', width: 160 },
  { colKey: 'value_type', title: t('settings.valueType'), cell: 'typeCell', width: 90 },
  { colKey: 'description', title: t('settings.description'), ellipsis: true },
  { colKey: 'is_public', title: t('settings.isPublic'), cell: 'publicCell', width: 90, align: 'center' as const },
  { colKey: 'operations', title: t('common.actions'), cell: 'operations', width: 140 },
]

const filteredConfigs = computed(() => {
  if (!keyword.value) return configs.value
  const kw = keyword.value.toLowerCase()
  return configs.value.filter(c => c.config_key.toLowerCase().includes(kw))
})

const resetForm = () => {
  form.config_key = ''
  form.config_value = ''
  form.config_value_bool = false
  form.config_value_number = 0
  form.value_type = 'string'
  form.description = ''
  form.is_public = false
}

const loadConfigs = async () => {
  loading.value = true
  try {
    configs.value = await getSystemConfigs()
  } catch (error: any) {
    MessagePlugin.error(error?.response?.data?.msg || t('settings.loadFailed'))
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  editingConfig.value = null
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (row: SystemConfig) => {
  editingConfig.value = row
  form.config_key = row.config_key
  form.config_value = row.config_value
  form.config_value_bool = row.config_value === 'true'
  form.config_value_number = row.value_type === 'number' ? Number(row.config_value) : 0
  form.value_type = row.value_type
  form.description = row.description
  form.is_public = row.is_public
  dialogVisible.value = true
}

const openDeleteDialog = (row: SystemConfig) => {
  deletingConfig.value = row
  deleteVisible.value = true
}

const handleTogglePublic = async (row: SystemConfig, val: boolean) => {
  try {
    await updateSystemConfig(row.config_key, { is_public: val })
    MessagePlugin.success(t('settings.updateSuccess'))
  } catch (error: any) {
    row.is_public = !val
    MessagePlugin.error(error?.response?.data?.msg || t('settings.updateFailed'))
  }
}

const handleSave = async () => {
  if (!editingConfig.value && !form.config_key) {
    MessagePlugin.error(t('settings.valueRequired'))
    return
  }

  const value = form.value_type === 'boolean'
    ? (form.config_value_bool ? 'true' : 'false')
    : form.value_type === 'number'
      ? String(form.config_value_number)
      : form.config_value

  saving.value = true
  try {
    if (editingConfig.value) {
      await updateSystemConfig(form.config_key, {
        config_value: value,
        value_type: form.value_type,
        description: form.description,
        is_public: form.is_public,
      })
    } else {
      await createSystemConfig({
        config_key: form.config_key,
        config_value: value,
        value_type: form.value_type,
        description: form.description,
        is_public: form.is_public,
      })
    }
    MessagePlugin.success(t('settings.updateSuccess'))
    dialogVisible.value = false
    await loadConfigs()
  } catch (error: any) {
    MessagePlugin.error(error?.response?.data?.msg || t('settings.updateFailed'))
  } finally {
    saving.value = false
  }
}

const handleDelete = async () => {
  if (!deletingConfig.value) return
  deleting.value = true
  try {
    await deleteSystemConfig(deletingConfig.value.config_key)
    MessagePlugin.success(t('common.success'))
    deleteVisible.value = false
    await loadConfigs()
  } catch (error: any) {
    MessagePlugin.error(error?.response?.data?.msg || t('settings.cannotDeleteSystem'))
  } finally {
    deleting.value = false
  }
}

const handleClearCache = async () => {
  clearingCache.value = true
  try {
    await clearSystemConfigCache()
    MessagePlugin.success(t('common.success'))
  } catch (error: any) {
    MessagePlugin.error(error?.response?.data?.msg || t('common.error'))
  } finally {
    clearingCache.value = false
  }
}

onMounted(() => {
  loadConfigs()
})
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.config-key {
  font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
  font-size: 13px;
  background: var(--td-bg-color-component);
  padding: 2px 6px;
  border-radius: 4px;
}

.config-value {
  word-break: break-all;
}

.empty-value {
  color: var(--td-text-color-placeholder);
}
</style>
