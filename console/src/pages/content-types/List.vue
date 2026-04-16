<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next'
import {
  getContentTypes,
  createContentType,
  updateContentType,
  deleteContentType,
  type ContentType,
  type ContentTypeCreate,
  type ContentTypeUpdate,
} from '@/api/content-type'
import { showError, getFriendlyError } from '@/utils/request'

const router = useRouter()

// 状态
const loading = ref(false)
const dataList = ref<ContentType[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 创建/编辑对话框
const dialogVisible = ref(false)
const dialogTitle = ref('创建内容类型')
const isEditing = ref(false)
const editingId = ref('')
const submitting = ref(false) // MX-001: 表单提交 Loading

const formData = ref<ContentTypeCreate>({
  name: '',
  slug: '',
  description: '',
  kind: 'collection',
  versioning_enabled: false,
})

// 表单验证
const formError = ref('')

// 自动生成 slug
const generateSlug = () => {
  if (formData.value.name && !isEditing.value) {
    formData.value.slug = formData.value.name
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
  }
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const res = await getContentTypes({ page: currentPage.value, page_size: pageSize.value })
    if (res.data.code === 0) {
      dataList.value = res.data.data.items || []
      total.value = res.data.data.total || 0
    }
  } catch (e) {
    showError(e)
  } finally {
    loading.value = false
  }
}

// 分页变化
const onPageChange = (page: number) => {
  currentPage.value = page
  loadData()
}

// 打开创建对话框
const openCreateDialog = () => {
  isEditing.value = false
  dialogTitle.value = '创建内容类型'
  formError.value = ''
  formData.value = {
    name: '',
    slug: '',
    description: '',
    kind: 'collection',
    versioning_enabled: false,
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const openEditDialog = (row: ContentType) => {
  isEditing.value = true
  dialogTitle.value = '编辑内容类型'
  formError.value = ''
  editingId.value = row.id
  formData.value = {
    name: row.name,
    slug: row.slug,
    description: row.description || '',
    kind: row.kind,
    versioning_enabled: row.versioning_enabled,
  }
  dialogVisible.value = true
}

// 提交表单
const submitForm = async () => {
  if (!formData.value.name) {
    formError.value = '请输入名称'
    return
  }
  if (!formData.value.slug) {
    formError.value = '请输入标识符'
    return
  }
  formError.value = ''
  submitting.value = true

  try {
    if (isEditing.value) {
      await updateContentType(editingId.value, formData.value as ContentTypeUpdate)
      MessagePlugin.success('内容类型已更新')
    } else {
      await createContentType(formData.value)
      MessagePlugin.success('内容类型创建成功')
    }
    dialogVisible.value = false
    loadData()
  } catch (e: any) {
    formError.value = getFriendlyError(e)
  } finally {
    submitting.value = false
  }
}

// 删除内容类型 - MX-001: 二次确认弹窗
const handleDelete = async (row: ContentType) => {
  const confirmDialog = DialogPlugin.confirm({
    header: '确认删除',
    body: `确定删除「${row.name}」吗？此操作不可恢复。`,
    confirmBtn: { content: '确认删除', theme: 'danger' },
    cancelBtn: '取消',
    onConfirm: async () => {
      try {
        await deleteContentType(row.id)
        MessagePlugin.success('删除成功')
        loadData()
      } catch (e) {
        showError(e)
      } finally {
        confirmDialog.hide()
      }
    },
    onClose: () => confirmDialog.hide(),
  })
}

// 跳转到字段管理
const goToFields = (row: ContentType) => {
  router.push(`/content-types/${row.id}/fields`)
}

// 格式化时间
const formatDate = (date: string) => {
  return new Date(date).toLocaleString('zh-CN')
}

// 格式化 kind
const formatKind = (kind: string) => {
  return kind === 'collection' ? '集合' : '单条'
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="content-types-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div class="title-section">
        <h1>内容类型</h1>
        <p class="subtitle">定义和管理内容的数据结构</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="loadData">刷新</button>
        <button class="btn btn-primary" @click="openCreateDialog">
          <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
            <path d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"/>
          </svg>
          创建内容类型
        </button>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="card" style="padding: 0; overflow: hidden;">
      <table class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th style="width: 100px;">类型</th>
            <th style="width: 100px;">状态</th>
            <th style="width: 100px;">版本控制</th>
            <th>描述</th>
            <th style="width: 180px;">更新时间</th>
            <th style="width: 180px;">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" class="text-center">加载中...</td>
          </tr>
          <tr v-else-if="dataList.length === 0">
            <td colspan="7" class="empty-state">
              <h3>暂无内容类型</h3>
              <p>创建您的第一个内容类型来开始</p>
              <button class="btn btn-primary btn-sm" @click="openCreateDialog">
                创建内容类型
              </button>
            </td>
          </tr>
          <tr v-else v-for="row in dataList" :key="row.id">
            <td>
              <div class="name-cell">
                <span class="name">{{ row.name }}</span>
                <span class="slug">{{ row.slug }}</span>
              </div>
            </td>
            <td>
              <span :class="['badge', row.kind === 'collection' ? 'badge-primary' : 'badge-warning']">
                {{ formatKind(row.kind) }}
              </span>
            </td>
            <td>
              <span :class="['badge', row.is_active ? 'badge-success' : 'badge-default']">
                {{ row.is_active ? '启用' : '禁用' }}
              </span>
            </td>
            <td>
              <svg v-if="row.versioning_enabled" width="20" height="20" viewBox="0 0 20 20" fill="#10b981">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"/>
              </svg>
              <svg v-else width="20" height="20" viewBox="0 0 20 20" fill="#94a3b8">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"/>
              </svg>
            </td>
            <td class="description">{{ row.description || '-' }}</td>
            <td class="time">{{ formatDate(row.updated_time) }}</td>
            <td>
              <div class="action-btns">
                <button class="btn btn-secondary btn-sm" @click="goToFields(row)" title="管理字段">
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z"/>
                  </svg>
                </button>
                <button class="btn btn-secondary btn-sm" @click="openEditDialog(row)" title="编辑">
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z"/>
                  </svg>
                </button>
                <button class="btn btn-danger btn-sm" @click="handleDelete(row)" title="删除">
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页 -->
      <div class="pagination">
        <span class="pagination-info">共 {{ total }} 条</span>
        <button
          class="btn btn-secondary btn-sm"
          :disabled="currentPage === 1"
          @click="onPageChange(currentPage - 1)"
        >
          上一页
        </button>
        <span class="pagination-current">{{ currentPage }} / {{ Math.ceil(total / pageSize) || 1 }}</span>
        <button
          class="btn btn-secondary btn-sm"
          :disabled="currentPage >= Math.ceil(total / pageSize)"
          @click="onPageChange(currentPage + 1)"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <div v-if="dialogVisible" class="modal-overlay" @click.self="dialogVisible = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ dialogTitle }}</h3>
          <button class="modal-close" @click="dialogVisible = false">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div v-if="formError" class="form-error">{{ formError }}</div>

          <div class="input-group">
            <label class="input-label">名称 <span class="required">*</span></label>
            <input
              v-model="formData.name"
              type="text"
              class="input"
              placeholder="请输入内容类型名称"
              @blur="generateSlug"
            />
          </div>

          <div class="input-group">
            <label class="input-label">标识符 <span class="required">*</span></label>
            <input
              v-model="formData.slug"
              type="text"
              class="input"
              placeholder="如：article、product"
              :disabled="isEditing"
            />
            <span class="input-hint">只能包含小写字母、数字和连字符，必须以字母开头</span>
          </div>

          <div class="input-group">
            <label class="input-label">类型</label>
            <select v-model="formData.kind" class="input" :disabled="isEditing">
              <option value="collection">集合类型</option>
              <option value="single">单条类型</option>
            </select>
          </div>

          <div class="input-group">
            <label class="input-label">描述</label>
            <textarea
              v-model="formData.description"
              class="input"
              rows="3"
              placeholder="可选，简要描述这个内容类型的用途"
            ></textarea>
          </div>

          <div class="input-group">
            <label class="input-label">
              <input v-model="formData.versioning_enabled" type="checkbox" />
              <span style="margin-left: 8px;">启用版本控制</span>
            </label>
            <span class="input-hint">启用后可保留内容的历史版本</span>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="dialogVisible = false">取消</button>
          <button class="btn btn-primary" :disabled="submitting" @click="submitForm">
            {{ submitting ? '处理中...' : (isEditing ? '保存' : '创建') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.content-types-page {
  height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.title-section h1 {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 4px 0;
}

.subtitle {
  color: var(--color-text-secondary);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.name-cell .name {
  font-weight: 500;
}

.name-cell .slug {
  font-size: 12px;
  color: var(--color-text-secondary);
  font-family: monospace;
}

.description {
  color: var(--color-text-secondary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  color: var(--color-text-secondary);
  font-size: 13px;
}

.action-btns {
  display: flex;
  gap: 6px;
}

.text-center {
  text-align: center;
  padding: 40px !important;
}

.badge-primary {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.badge-warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--color-border);
}

.pagination-info {
  font-size: 14px;
  color: var(--color-text-secondary);
}

.pagination-current {
  font-size: 14px;
  color: var(--color-text);
}

/* Modal */
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
  width: 500px;
  max-height: 80vh;
  background: var(--color-card);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border);
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.modal-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--color-text-secondary);
  cursor: pointer;
}

.modal-close:hover {
  background: var(--color-hover);
}

.modal-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--color-border);
}

.form-error {
  padding: 10px 12px;
  background: var(--color-error-light);
  color: var(--color-error);
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 14px;
}

.required {
  color: var(--color-error);
}

.input-hint {
  display: block;
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-top: 4px;
}
</style>
