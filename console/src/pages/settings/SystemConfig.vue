<template>
  <div class="system-config">
    <t-card :title="t('settings.systemConfig')">
      <t-loading :loading="loading" style="min-height: 200px">
        <t-table
          :data="configs"
          :columns="columns"
          row-key="config_key"
          :bordered="true"
          :hover="true"
        >
          <template #valueCell="{ row }">
            <t-tag v-if="row.value_type === 'boolean'" :theme="row.config_value === 'true' ? 'success' : 'danger'">
              {{ row.config_value }}
            </t-tag>
            <span v-else>{{ row.config_value }}</span>
          </template>

          <template #typeCell="{ row }">
            <t-tag>{{ row.value_type }}</t-tag>
          </template>

          <template #publicCell="{ row }">
            <t-switch v-model="row.is_public" @change="(val) => handleTogglePublic(row, val)" />
          </template>

          <template #operationCell="{ row }">
            <t-space>
              <t-button theme="primary" variant="text" @click="handleEdit(row)">
                {{ t('common.edit') }}
              </t-button>
            </t-space>
          </template>
        </t-table>
      </t-loading>
    </t-card>

    <!-- 编辑配置对话框 -->
    <t-dialog
      v-model:visible="editDialogVisible"
      :header="t('settings.editConfig')"
      :width="640"
      :confirm-on-enter="true"
      @confirm="handleSave"
    >
      <t-form :data="editForm" label-align="top" ref="editFormRef">
        <t-form-item :label="t('settings.configKey')">
          <t-input v-model="editForm.config_key" disabled />
        </t-form-item>

        <t-form-item :label="t('settings.configValue')">
          <t-input
            v-if="editForm.value_type === 'string' || editForm.value_type === 'number'"
            v-model="editForm.config_value"
            :placeholder="t('settings.enterValue')"
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
          <t-select v-model="editForm.value_type" :disabled="true">
            <t-option value="string" label="String" />
            <t-option value="number" label="Number" />
            <t-option value="boolean" label="Boolean" />
            <t-option value="json" label="JSON" />
          </t-select>
        </t-form-item>

        <t-form-item :label="t('settings.description')">
          <t-textarea v-model="editForm.description" :placeholder="t('settings.enterDescription')" />
        </t-form-item>

        <t-form-item :label="t('settings.isPublic')">
          <t-switch v-model="editForm.is_public" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next'
import { getSystemConfigs, updateSystemConfig } from '@/api/system-config'
import type { SystemConfig } from '@/types/system-config'

const { t } = useI18n()

const loading = ref(false)
const configs = ref<SystemConfig[]>([])
const editDialogVisible = ref(false)
const editFormRef = ref()

const editForm = reactive({
  config_key: '',
  config_value: '',
  config_value_bool: false,
  value_type: 'string',
  description: '',
  is_public: false,
})

const columns = [
  { colKey: 'config_key', title: t('settings.configKey'), width: 200 },
  { colKey: 'value', title: t('settings.configValue'), cell: 'valueCell' },
  { colKey: 'value_type', title: t('settings.valueType'), cell: 'typeCell', width: 100 },
  { colKey: 'description', title: t('settings.description'), ellipsis: true },
  { colKey: 'is_public', title: t('settings.isPublic'), cell: 'publicCell', width: 100 },
  { colKey: 'operation', title: t('common.operation'), cell: 'operationCell', width: 120 },
]

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
    row.is_public = !val // 回滚
    MessagePlugin.error(error?.response?.data?.msg || t('settings.updateFailed'))
  }
}

const handleSave = async () => {
  if (!editForm.config_value && editForm.value_type !== 'boolean') {
    MessagePlugin.error(t('settings.valueRequired'))
    return
  }

  const value = editForm.value_type === 'boolean'
    ? (editForm.config_value_bool ? 'true' : 'false')
    : editForm.config_value

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
</style>
