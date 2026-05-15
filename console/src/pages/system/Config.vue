<template>
  <div class="system-config">
    <t-card>
      <template #title>
        <span>{{ t('settings.systemConfig') }}</span>
      </template>
      <template #subtitle>
        <span class="subtitle">{{ t('settings.systemConfigDesc') }}</span>
      </template>

      <!-- 搜索栏 -->
      <t-input
        v-model="keyword"
        :placeholder="t('common.searchPlaceholder')"
        clearable
        style="width: 320px; margin-bottom: 16px"
        @change="handleSearch"
      >
        <template #prefix-icon>
          <t-icon name="search" />
        </template>
      </t-input>

      <t-loading :loading="loading" style="min-height: 200px">
        <t-table
          :data="filteredConfigs"
          :columns="columns"
          row-key="config_key"
          :bordered="true"
          :hover="true"
          :stripe="true"
          size="medium"
          :empty="t('common.noData')"
        >
          <template #keyCell="{ row }">
            <code class="config-key">{{ row.config_key }}</code>
          </template>

          <template #valueCell="{ row }">
            <template v-if="row.value_type === 'boolean'">
              <t-tag :theme="row.config_value === 'true' ? 'success' : 'danger'" variant="light">
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
            <t-switch v-model="row.is_public" size="small" @change="(val: boolean) => handleTogglePublic(row, val)" />
          </template>

          <template #operationCell="{ row }">
            <t-button theme="primary" variant="text" size="small" @click="handleEdit(row)">
              {{ t('common.edit') }}
            </t-button>
          </template>
        </t-table>
      </t-loading>
    </t-card>

    <!-- 编辑配置对话框 -->
    <t-dialog
      v-model:visible="editDialogVisible"
      :header="t('settings.editConfig')"
      :width="560"
      :confirm-on-enter="true"
      @confirm="handleSave"
    >
      <t-form :data="editForm" label-align="top">
        <t-form-item :label="t('settings.configKey')">
          <t-input v-model="editForm.config_key" disabled />
        </t-form-item>

        <t-form-item :label="t('settings.configValue')" :help="editForm.description">
          <t-input
            v-if="editForm.value_type === 'string'"
            v-model="editForm.config_value"
            :placeholder="t('settings.enterValue')"
          />
          <t-input-number
            v-else-if="editForm.value_type === 'number'"
            v-model="editForm.config_value_number"
            :placeholder="t('settings.enterValue')"
            style="width: 100%"
          />
          <t-switch
            v-else-if="editForm.value_type === 'boolean'"
            v-model="editForm.config_value_bool"
          />
          <t-textarea
            v-else-if="editForm.value_type === 'json'"
            v-model="editForm.config_value"
            :placeholder="t('settings.enterJson')"
            :autosize="{ minRows: 3, maxRows: 10 }"
          />
        </t-form-item>

        <t-form-item :label="t('settings.valueType')">
          <t-select v-model="editForm.value_type" disabled>
            <t-option value="string" label="String" />
            <t-option value="number" label="Number" />
            <t-option value="boolean" label="Boolean" />
            <t-option value="json" label="JSON" />
          </t-select>
        </t-form-item>

        <t-form-item :label="t('settings.description')">
          <t-textarea v-model="editForm.description" :placeholder="t('settings.enterDescription')" :autosize="{ minRows: 2, maxRows: 4 }" />
        </t-form-item>

        <t-form-item :label="t('settings.isPublic')">
          <t-switch v-model="editForm.is_public" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { getSystemConfigs, updateSystemConfig } from '@/api/system-config'
import type { SystemConfig } from '@/types/system/config'

const { t } = useI18n()

const loading = ref(false)
const configs = ref<SystemConfig[]>([])
const keyword = ref('')
const editDialogVisible = ref(false)

const editForm = reactive({
  config_key: '',
  config_value: '',
  config_value_bool: false,
  config_value_number: 0,
  value_type: 'string',
  description: '',
  is_public: false,
})

const columns = [
  { colKey: 'config_key', title: t('settings.configKey'), cell: 'keyCell', width: 220 },
  { colKey: 'value', title: t('settings.configValue'), cell: 'valueCell', width: 180 },
  { colKey: 'value_type', title: t('settings.valueType'), cell: 'typeCell', width: 90 },
  { colKey: 'description', title: t('settings.description'), ellipsis: true },
  { colKey: 'is_public', title: t('settings.isPublic'), cell: 'publicCell', width: 90, align: 'center' as const },
  { colKey: 'operation', title: t('common.operation'), cell: 'operationCell', width: 80 },
]

const filteredConfigs = computed(() => {
  if (!keyword.value) return configs.value
  const kw = keyword.value.toLowerCase()
  return configs.value.filter(
    c => c.config_key.toLowerCase().includes(kw) || c.description.toLowerCase().includes(kw)
  )
})

const handleSearch = () => {
  // computed 自动响应 keyword 变化
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

const handleEdit = (row: SystemConfig) => {
  editForm.config_key = row.config_key
  editForm.config_value = row.config_value
  editForm.config_value_bool = row.config_value === 'true'
  editForm.config_value_number = row.value_type === 'number' ? Number(row.config_value) : 0
  editForm.value_type = row.value_type
  editForm.description = row.description
  editForm.is_public = row.is_public
  editDialogVisible.value = true
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
  let value: string
  if (editForm.value_type === 'boolean') {
    value = editForm.config_value_bool ? 'true' : 'false'
  } else if (editForm.value_type === 'number') {
    value = String(editForm.config_value_number)
  } else {
    if (!editForm.config_value) {
      MessagePlugin.error(t('settings.valueRequired'))
      return
    }
    value = editForm.config_value
  }

  try {
    await updateSystemConfig(editForm.config_key, {
      config_value: value,
      value_type: editForm.value_type,
      description: editForm.description,
      is_public: editForm.is_public,
    })
    MessagePlugin.success(t('settings.updateSuccess'))
    editDialogVisible.value = false
    await loadConfigs()
  } catch (error: any) {
    MessagePlugin.error(error?.response?.data?.msg || t('settings.updateFailed'))
  }
}

onMounted(() => {
  loadConfigs()
})
</script>

<style scoped>
.system-config {
  padding: 24px;
}

.subtitle {
  color: var(--td-text-color-secondary);
  font-size: 14px;
  font-weight: 400;
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
