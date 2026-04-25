<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  Space,
  Button,
  Card,
  Form,
  FormItem,
  Input,
  Select,
  Option,
  Switch,
  MessagePlugin,
  Tooltip,
  Popconfirm,
  Alert,
  Dialog,
} from 'tdesign-vue-next'
import {
  getContentType,
  getFields,
  createField,
  updateField,
  deleteField,
  reorderFields,
  type ContentType,
  type Field,
  type FieldCreate,
  type FieldUpdate,
  FIELD_TYPES,
} from '@/api/content-type'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

// 状态
const loading = ref(false)
const contentType = ref<ContentType | null>(null)
const fields = ref<Field[]>([])
const total = ref(0)

// 编辑对话框
const dialogVisible = ref(false)
const dialogTitle = computed(() => isEditing.value ? t('fields.editField') : t('fields.addField'))
const isEditing = ref(false)
const editingId = ref('')

const formData = ref<FieldCreate>({
  name: '',
  label: '',
  description: '',
  field_type: 'text',
  validation: {},
  display: {},
})

// 表单规则
const formRules = {
  name: [
    { required: true, message: t('fields.nameRequired'), trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, message: t('fields.nameFormat'), trigger: 'blur' },
  ],
  label: [{ required: true, message: t('fields.displayNameRequired'), trigger: 'blur' }],
  field_type: [{ required: true, message: t('fields.typeRequired'), trigger: 'change' }],
}

// 字段类型选项（带 i18n label）
const fieldTypeOptions = computed(() =>
  Object.entries(FIELD_TYPES).map(([key, info]) => ({
    value: key,
    label: t(info.labelKey),
  }))
)

// 自动生成 label
const generateLabel = () => {
  if (formData.value.name && !isEditing.value) {
    formData.value.label = formData.value.name
      .replace(/_/g, ' ')
      .replace(/([A-Z])/g, ' $1')
      .trim()
  }
}

// 加载内容类型详情
const loadContentType = async () => {
  try {
    const id = route.params.id as string
    const res = await getContentType(id)
    if (res.data.code === 200) {
      contentType.value = res.data.data
    }
  } catch {
    MessagePlugin.error(t('contentTypes.loadFailed'))
  }
}

// 加载字段列表
const loadFields = async () => {
  loading.value = true
  try {
    const id = route.params.id as string
    const res = await getFields(id)
    if (res.data.code === 200) {
      fields.value = res.data.data || []
      total.value = (res.data.data || []).length
    }
  } catch {
    MessagePlugin.error(t('fields.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 打开添加对话框
const openCreateDialog = () => {
  isEditing.value = false
  editingId.value = ''
  formData.value = {
    name: '',
    label: '',
    description: '',
    field_type: 'text',
    validation: {},
    display: {},
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const openEditDialog = (field: Field) => {
  isEditing.value = true
  editingId.value = field.id
  formData.value = {
    name: field.name,
    label: field.label,
    description: field.description,
    field_type: field.field_type,
    validation: field.validation,
    display: field.display,
  }
  dialogVisible.value = true
}

// 提交表单
const submitForm = async () => {
  try {
    const contentTypeId = route.params.id as string
    if (isEditing.value) {
      await updateField(contentTypeId, editingId.value, formData.value as FieldUpdate)
      MessagePlugin.success(t('fields.updateSuccess'))
    } else {
      await createField(contentTypeId, formData.value)
      MessagePlugin.success(t('fields.createSuccess'))
    }
    dialogVisible.value = false
    loadFields()
  } catch (e: any) {
    MessagePlugin.error(e?.response?.data?.msg || t('fields.deleteFailed'))
  }
}

// 删除字段
const handleDelete = async (field: Field) => {
  try {
    const contentTypeId = route.params.id as string
    await deleteField(contentTypeId, field.id)
    MessagePlugin.success(t('fields.deleteSuccess'))
    loadFields()
  } catch (e: any) {
    MessagePlugin.error(e?.response?.data?.msg || t('fields.deleteFailed'))
  }
}

// 获取字段类型标签（通过 i18n key）
const getFieldTypeLabel = (type: string) => {
  const info = FIELD_TYPES[type as keyof typeof FIELD_TYPES]
  return info ? t(info.labelKey) : type
}

// 返回列表
const goBack = () => {
  router.push('/content/types')
}

onMounted(() => {
  loadContentType()
  loadFields()
})

// 监听路由变化
watch(() => route.params.id, () => {
  if (route.params.id) {
    loadContentType()
    loadFields()
  }
})
</script>

<template>
  <div class="fields-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="title-section">
        <Button theme="default" variant="text" @click="goBack">
          <template #icon><Icon name="arrow-left" /></template>
          {{ t('common.back') }}
        </Button>
        <div class="title-info">
          <h1>{{ contentType?.name || t('common.loading') }}</h1>
          <p class="subtitle">
            <span class="slug">{{ contentType?.slug }}</span>
            <span class="kind">{{ contentType?.kind === 'collection' ? t('contentTypes.kindCollection') : t('contentTypes.kindSingle') }}</span>
          </p>
        </div>
      </div>
      <Button theme="primary" @click="openCreateDialog">
        <template #icon><Icon name="add" /></template>
        {{ t('fields.addField') }}
      </Button>
    </div>

    <!-- 字段列表 -->
    <Card>
      <template #header>
        <div class="card-header">
          <span>{{ t('fields.fieldList') }}</span>
          <span class="field-count">{{ t('fields.totalFields', { count: total }) }}</span>
        </div>
      </template>

      <div v-if="loading" class="loading">{{ t('common.loading') }}</div>

      <div v-else-if="fields.length === 0" class="empty">
        <Alert theme="info" :message="t('fields.noFields')" />
        <Button theme="primary" style="margin-top: 16px" @click="openCreateDialog">
          <template #icon><Icon name="add" /></template>
          {{ t('fields.addFirstField') }}
        </Button>
      </div>

      <div v-else class="fields-list">
        <div
          v-for="(field, index) in fields"
          :key="field.id"
          class="field-item"
        >
          <div class="field-drag">
            <Tooltip :content="t('fields.dragToSort')">
              <Icon name="drag" />
            </Tooltip>
          </div>

          <div class="field-order">{{ index + 1 }}</div>

          <div class="field-main">
            <div class="field-name">
              <span class="name">{{ field.name }}</span>
              <span class="type-badge">{{ getFieldTypeLabel(field.field_type) }}</span>
            </div>
            <div class="field-label">{{ field.label }}</div>
            <div v-if="field.description" class="field-desc">{{ field.description }}</div>
          </div>

          <div class="field-actions">
            <Tooltip :content="t('common.edit')">
              <Button size="small" variant="text" @click="openEditDialog(field)">
                <template #icon><Icon name="edit" /></template>
              </Button>
            </Tooltip>
            <Popconfirm @confirm="handleDelete(field)">
              <template #content>
                <p>{{ t('fields.deleteConfirm') }}</p>
              </template>
              <Tooltip :content="t('common.delete')">
                <Button size="small" variant="text" theme="danger">
                  <template #icon><Icon name="delete" /></template>
                </Button>
              </Tooltip>
            </Popconfirm>
          </div>
        </div>
      </div>
    </Card>

    <!-- 添加/编辑对话框 -->
    <Dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      width="600px"
      :close-on-overlay-click="false"
    >
      <Form :data="formData" :rules="formRules" label-width="100px" colon>
        <FormItem :label="t('fields.fieldName')" name="name">
          <Input
            v-model="formData.name"
            :placeholder="t('fields.fieldNamePlaceholder')"
            :disabled="isEditing"
            @blur="generateLabel"
          />
          <template #help>
            <span style="color: var(--td-text-color-secondary)">
              {{ t('fields.fieldNameHint') }}
            </span>
          </template>
        </FormItem>

        <FormItem :label="t('fields.displayName')" name="label">
          <Input
            v-model="formData.label"
            :placeholder="t('fields.displayNamePlaceholder')"
          />
        </FormItem>

        <FormItem :label="t('fields.fieldType')" name="field_type">
          <Select v-model="formData.field_type" :disabled="isEditing">
            <Option
              v-for="opt in fieldTypeOptions"
              :key="opt.value"
              :value="opt.value"
              :label="opt.label"
            />
          </Select>
        </FormItem>

        <FormItem :label="t('fields.fieldDescription')">
          <Input
            v-model="formData.description"
            type="textarea"
            :placeholder="t('fields.fieldDescPlaceholder')"
            :rows="2"
          />
        </FormItem>
      </Form>

      <template #footer>
        <Space>
          <Button @click="dialogVisible = false">{{ t('common.cancel') }}</Button>
          <Button theme="primary" @click="submitForm">
            {{ isEditing ? t('common.save') : t('fields.addField') }}
          </Button>
        </Space>
      </template>
    </Dialog>
  </div>
</template>

<style scoped>
.fields-page {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.title-section {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.title-info h1 {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 4px 0;
}

.subtitle {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0;
}

.slug {
  font-family: monospace;
  color: var(--td-text-color-secondary);
}

.kind {
  padding: 2px 8px;
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
  border-radius: 4px;
  font-size: 12px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.field-count {
  color: var(--td-text-color-secondary);
  font-size: 14px;
}

.loading,
.empty {
  padding: 48px;
  text-align: center;
}

.fields-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: var(--td-bg-content-container);
  border-radius: 8px;
  transition: background 0.2s;
}

.field-item:hover {
  background: var(--td-bg-container);
}

.field-drag {
  cursor: grab;
  color: var(--td-text-color-secondary);
}

.field-order {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--td-bg-container);
  border-radius: 50%;
  font-size: 12px;
  color: var(--td-text-color-secondary);
}

.field-main {
  flex: 1;
  min-width: 0;
}

.field-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.field-name .name {
  font-family: monospace;
  font-weight: 500;
}

.type-badge {
  padding: 2px 6px;
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
  border-radius: 4px;
  font-size: 11px;
}

.field-label {
  color: var(--td-text-color-secondary);
  font-size: 13px;
  margin-top: 2px;
}

.field-desc {
  color: var(--td-text-color-disabled);
  font-size: 12px;
  margin-top: 4px;
}

.field-actions {
  display: flex;
  gap: 4px;
}
</style>
