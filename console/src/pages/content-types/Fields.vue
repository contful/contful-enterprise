<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
  Icon,
  Tooltip,
  Popconfirm,
  Alert,
} from 'tdesign-vue-next'
import {
  ArrowLeftIcon,
  AddIcon,
  DeleteIcon,
  EditIcon,
  DragIcon,
  SaveIcon,
} from 'tdesign-icons-vue-next'
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

const route = useRoute()
const router = useRouter()

// 状态
const loading = ref(false)
const contentType = ref<ContentType | null>(null)
const fields = ref<Field[]>([])
const total = ref(0)

// 编辑对话框
const dialogVisible = ref(false)
const dialogTitle = ref('添加字段')
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
    { required: true, message: '请输入字段名', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, message: '只能包含字母、数字和下划线，必须以字母开头', trigger: 'blur' },
  ],
  label: [{ required: true, message: '请输入显示名称', trigger: 'blur' }],
  field_type: [{ required: true, message: '请选择字段类型', trigger: 'change' }],
}

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
    if (res.data.code === 0) {
      contentType.value = res.data.data
    }
  } catch (e) {
    MessagePlugin.error('加载内容类型失败')
  }
}

// 加载字段列表
const loadFields = async () => {
  loading.value = true
  try {
    const id = route.params.id as string
    const res = await getFields(id)
    if (res.data.code === 0) {
      fields.value = res.data.data.items
      total.value = res.data.data.items.length
    }
  } catch (e) {
    MessagePlugin.error('加载字段失败')
  } finally {
    loading.value = false
  }
}

// 打开添加对话框
const openCreateDialog = () => {
  isEditing.value = false
  dialogTitle.value = '添加字段'
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
  dialogTitle.value = '编辑字段'
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
      await updateField(editingId.value, formData.value as FieldUpdate)
      MessagePlugin.success('更新成功')
    } else {
      await createField(contentTypeId, formData.value)
      MessagePlugin.success('添加成功')
    }
    dialogVisible.value = false
    loadFields()
  } catch (e: any) {
    MessagePlugin.error(e?.response?.data?.msg || '操作失败')
  }
}

// 删除字段
const handleDelete = async (field: Field) => {
  try {
    await deleteField(field.id)
    MessagePlugin.success('删除成功')
    loadFields()
  } catch (e: any) {
    MessagePlugin.error(e?.response?.data?.msg || '删除失败')
  }
}

// 获取字段类型标签
const getFieldTypeLabel = (type: string) => {
  return FIELD_TYPES[type as keyof typeof FIELD_TYPES]?.label || type
}

// 返回列表
const goBack = () => {
  router.push('/content-types')
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
          <template #icon><ArrowLeftIcon /></template>
          返回
        </Button>
        <div class="title-info">
          <h1>{{ contentType?.name || '加载中...' }}</h1>
          <p class="subtitle">
            <span class="slug">{{ contentType?.slug }}</span>
            <span class="kind">{{ contentType?.kind === 'collection' ? '集合' : '单条' }}</span>
          </p>
        </div>
      </div>
      <Button theme="primary" @click="openCreateDialog">
        <template #icon><AddIcon /></template>
        添加字段
      </Button>
    </div>

    <!-- 字段列表 -->
    <Card>
      <template #header>
        <div class="card-header">
          <span>字段列表</span>
          <span class="field-count">共 {{ total }} 个字段</span>
        </div>
      </template>

      <div v-if="loading" class="loading">加载中...</div>

      <div v-else-if="fields.length === 0" class="empty">
        <Alert theme="info" message="还没有添加任何字段" />
        <Button theme="primary" style="margin-top: 16px" @click="openCreateDialog">
          <template #icon><AddIcon /></template>
          添加第一个字段
        </Button>
      </div>

      <div v-else class="fields-list">
        <div
          v-for="(field, index) in fields"
          :key="field.id"
          class="field-item"
        >
          <div class="field-drag">
            <Tooltip content="拖动排序">
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
            <Tooltip content="编辑">
              <Button size="small" variant="text" @click="openEditDialog(field)">
                <template #icon><EditIcon /></template>
              </Button>
            </Tooltip>
            <Popconfirm @confirm="handleDelete(field)">
              <template #content>
                <p>确定删除字段「{{ field.name }}」吗？</p>
              </template>
              <Tooltip content="删除">
                <Button size="small" variant="text" theme="danger">
                  <template #icon><DeleteIcon /></template>
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
        <FormItem label="字段名" name="name">
          <Input
            v-model="formData.name"
            placeholder="如：title、content"
            :disabled="isEditing"
            @blur="generateLabel"
          />
          <template #help>
            <span style="color: var(--td-text-color-secondary)">
              只能包含字母、数字和下划线，将作为 API 字段名
            </span>
          </template>
        </FormItem>

        <FormItem label="显示名称" name="label">
          <Input
            v-model="formData.label"
            placeholder="如：标题、正文"
          />
        </FormItem>

        <FormItem label="字段类型" name="field_type">
          <Select v-model="formData.field_type" :disabled="isEditing">
            <Option
              v-for="(info, key) in FIELD_TYPES"
              :key="key"
              :value="key"
              :label="info.label"
            />
          </Select>
        </FormItem>

        <FormItem label="描述">
          <Input
            v-model="formData.description"
            type="textarea"
            placeholder="可选，对字段用途的说明"
            :rows="2"
          />
        </FormItem>
      </Form>

      <template #footer>
        <Space>
          <Button @click="dialogVisible = false">取消</Button>
          <Button theme="primary" @click="submitForm">
            {{ isEditing ? '保存' : '添加' }}
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
