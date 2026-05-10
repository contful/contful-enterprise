<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { DialogPlugin } from 'tdesign-vue-next'
import {
  getContentSchemas,
  createContentSchema,
  updateContentSchema,
  deleteContentSchema,
  type ContentSchema,
  type ContentSchemaCreate,
  type ContentSchemaUpdate,
} from '@/api/schema'
import { showError, showSuccess, getFriendlyError } from '@/utils/request'
import PageHeader from '@/components/PageHeader.vue'

const { t } = useI18n()
const router = useRouter()

// 状态
const loading = ref(false)
const dataList = ref<ContentSchema[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 创建/编辑对话框
const dialogVisible = ref(false)
const dialogTitle = computed(() => isEditing.value ? t('contentSchemas.formTitleEdit') : t('contentSchemas.formTitleCreate'))
const isEditing = ref(false)
const editingId = ref('')
const submitting = ref(false)

const formData = ref<ContentSchemaCreate>({
  name: '',
  slug: '',
  description: '',
  kind: 'collection',
  versioning_enabled: false,
})

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
    const res = await getContentSchemas({ page: currentPage.value, page_size: pageSize.value })
    dataList.value = res.data?.items || []
    total.value = res.data?.total || 0
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
const openEditDialog = (row: ContentSchema) => {
  isEditing.value = true
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
    formError.value = t('contentSchemas.namePlaceholder')
    return
  }
  if (!formData.value.slug) {
    formError.value = t('contentSchemas.slugPlaceholder')
    return
  }
  formError.value = ''
  submitting.value = true

  try {
    if (isEditing.value) {
      await updateContentSchema(editingId.value, formData.value as ContentSchemaUpdate)
      showSuccess(t('contentSchemas.updateSuccess'))
    } else {
      await createContentSchema(formData.value)
      showSuccess(t('contentSchemas.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } catch (e: any) {
    formError.value = getFriendlyError(e)
  } finally {
    submitting.value = false
  }
}

// 删除内容类型
const handleDelete = async (row: ContentSchema) => {
  const confirmDialog = DialogPlugin.confirm({
    header: t('contentSchemas.deleteConfirm'),
    body: t('contentSchemas.deleteConfirmMsg'),
    confirmBtn: { content: t('common.confirm'), theme: 'danger' },
    cancelBtn: t('common.cancel'),
    onConfirm: async () => {
      try {
        await deleteContentSchema(row.id)
        showSuccess(t('contentSchemas.deleteSuccess'))
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
const goToFields = (row: ContentSchema) => {
  router.push(`/content/schemas/${row.id}/fields`)
}

// 格式化时间
const formatDate = (date: string) => {
  return new Date(date).toLocaleString()
}

// 格式化 kind
const formatKind = (kind: string) => {
  return kind === 'collection' ? t('contentSchemas.kindCollection') : t('contentSchemas.kindSingle')
}

// 格式化状态
const formatStatus = (isActive: boolean) => {
  return isActive ? t('common.enabled') : t('common.disabled')
}

const kindOptions = [
  { value: 'collection', label: computed(() => t('contentSchemas.kindCollection')) },
  { value: 'single', label: computed(() => t('contentSchemas.kindSingle')) },
]

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="page page--padded">
    <PageHeader
      :title="t('contentSchemas.title')"
      :subtitle="t('contentSchemas.subtitle')"
      :show-refresh="true"
      @refresh="loadData"
    >
      <template #primary-action>
        <t-button theme="primary" @click="openCreateDialog">
          <template #icon><t-icon name="add" /></template>
          {{ t('contentSchemas.createTypeBtn') }}
        </t-button>
      </template>
    </PageHeader>

    <!-- 数据表格 -->
    <div class="card" style="padding: 0; overflow: hidden;">
      <table class="table">
        <thead>
          <tr>
            <th>{{ t('contentSchemas.tableName') }}</th>
            <th style="width: 100px;">{{ t('contentSchemas.tableType') }}</th>
            <th style="width: 100px;">{{ t('contentSchemas.tableStatus') }}</th>
            <th style="width: 100px;">{{ t('contentSchemas.tableVersioning') }}</th>
            <th>{{ t('common.description') }}</th>
            <th style="width: 180px;">{{ t('common.updatedAt') }}</th>
            <th style="width: 180px;">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" class="text-center">{{ t('common.loading') }}</td>
          </tr>
          <tr v-else-if="dataList.length === 0">
            <td colspan="7" class="empty-state">
              <div class="empty-content">
                <h3>{{ t('contentSchemas.noTypes') }}</h3>
                <p>{{ t('contentSchemas.noTypesHint') }}</p>
              </div>
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
                {{ formatStatus(row.is_active) }}
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
                <t-button variant="outline" size="small" @click="goToFields(row)" :title="t('contentSchemas.manageFields')">
                  <template #icon><t-icon name="setting" /></template>
                </t-button>
                <t-button variant="outline" size="small" @click="openEditDialog(row)" :title="t('common.edit')">
                  <template #icon><t-icon name="edit" /></template>
                </t-button>
                <t-button theme="danger" variant="outline" size="small" @click="handleDelete(row)" :title="t('common.delete')">
                  <template #icon><t-icon name="delete" /></template>
                </t-button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页 -->
      <div class="pagination">
        <span class="pagination-info">{{ t('common.total') }} {{ total }} {{ t('common.items') }}</span>
        <t-button
          variant="outline"
          size="small"
          :disabled="currentPage === 1"
          @click="onPageChange(currentPage - 1)"
        >
          {{ t('common.prevPage') }}
        </t-button>
        <span class="pagination-current">{{ currentPage }} / {{ Math.ceil(total / pageSize) || 1 }}</span>
        <t-button
          variant="outline"
          size="small"
          :disabled="currentPage >= Math.ceil(total / pageSize)"
          @click="onPageChange(currentPage + 1)"
        >
          {{ t('common.nextPage') }}
        </t-button>
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
            <label class="input-label">{{ t('common.name') }} <span class="required">*</span></label>
            <input
              v-model="formData.name"
              type="text"
              class="input"
              :placeholder="t('contentSchemas.namePlaceholder')"
              @blur="generateSlug"
            />
          </div>

          <div class="input-group">
            <label class="input-label">{{ t('contentSchemas.slug') }} <span class="required">*</span></label>
            <input
              v-model="formData.slug"
              type="text"
              class="input"
              :placeholder="t('contentSchemas.slugPlaceholder')"
              :disabled="isEditing"
            />
            <span class="input-hint">{{ t('contentSchemas.slugHint') }}</span>
          </div>

          <div class="input-group">
            <label class="input-label">{{ t('contentSchemas.kind') }}</label>
            <select v-model="formData.kind" class="input" :disabled="isEditing">
              <option value="collection">{{ t('contentSchemas.kindCollection') }}</option>
              <option value="single">{{ t('contentSchemas.kindSingle') }}</option>
            </select>
            <span class="input-hint">
              <strong>{{ t('contentSchemas.kindCollection') }}：</strong>{{ t('contentSchemas.kindCollectionHint') }}<br/>
              <strong>{{ t('contentSchemas.kindSingle') }}：</strong>{{ t('contentSchemas.kindSingleHint') }}
            </span>
          </div>

          <div class="input-group">
            <label class="input-label">{{ t('common.description') }}</label>
            <textarea
              v-model="formData.description"
              class="input"
              rows="3"
              :placeholder="t('common.description')"
            ></textarea>
          </div>

          <div class="input-group">
            <label class="input-label">
              <input v-model="formData.versioning_enabled" type="checkbox" />
              <span style="margin-left: 8px;">{{ t('contentSchemas.versioning') }}</span>
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <t-button variant="outline" @click="dialogVisible = false">{{ t('common.cancel') }}</t-button>
          <t-button theme="primary" :disabled="submitting" @click="submitForm">
            {{ submitting ? t('common.loading') : (isEditing ? t('common.save') : t('common.create')) }}
          </t-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 页面特有样式：内容类型列表 */

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

/* === Action buttons — 已提取到 common.css === */

.text-center {
  text-align: center;
  padding: 40px !important;
}

.empty-state {
  text-align: center;
  padding: 60px 20px !important;
}

.empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.empty-content h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text);
}

.empty-content p {
  margin: 0;
  font-size: 14px;
  color: var(--color-text-secondary);
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

/* === Form error — block 版本，已提取到 common.css === */

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
