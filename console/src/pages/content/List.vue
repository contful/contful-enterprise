<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSiteStore } from '@/stores/site'
import { showError, showSuccess } from '@/utils/request'
import {
  getContentTypes,
  type ContentType,
} from '@/api/content-type'
import {
  getEntries,
  getEntry,
  createEntry,
  updateEntry,
  deleteEntry,
  publishEntry,
  unpublishEntry,
  batchDeleteEntries,
  batchPublishEntries,
  batchUnpublishEntries,
  type Entry,
  type EntryCreate,
  type EntryUpdate,
} from '@/api/entry'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const siteStore = useSiteStore()

// 状态
const loading = ref(false)
const submitting = ref(false)
const contentTypes = ref<ContentType[]>([])
const entries = ref<Entry[]>([])
const selectedType = ref<ContentType | null>(null)
const showModal = ref(false)
const showDeleteConfirm = ref(false)
const editingEntry = ref<Entry | null>(null)
const entryToDelete = ref<Entry | null>(null)
const deleteLoading = ref(false)
const publishLoading = ref<string | null>(null)

// 分页
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// 表单数据
const formData = ref<Record<string, any>>({})

// 过滤器
const statusFilter = ref<string>('')

// 搜索与排序
const searchKeyword = ref<string>('')
const sortField = ref<string>('updated_time')
const sortOrder = ref<'asc' | 'desc'>('desc')

// 批量选择
const selectedIds = ref<Set<string>>(new Set())
const batchLoading = ref(false)
const showBatchConfirm = ref(false)
const batchAction = ref<'delete' | 'publish' | 'unpublish'>('delete')
const batchActionLabel = ref('')

// 计算属性
const isAllSelected = computed(() => {
  return entries.value.length > 0 && selectedIds.value.size === entries.value.length
})

const selectedCount = computed(() => selectedIds.value.size)
const hasSelected = computed(() => selectedIds.value.size > 0)

// 加载内容类型
const loadContentTypes = async () => {
  // 检查是否有当前站点
  if (!siteStore.currentSiteId) {
    contentTypes.value = []
    return
  }
  try {
    const res = await getContentTypes({ page: 1, page_size: 100 })
    contentTypes.value = res.data.data.items || []
  } catch (error) {
    showError(error)
  }
}

// 加载内容列表
const loadEntries = async () => {
  if (!selectedType.value) return
  loading.value = true
  selectedIds.value.clear()
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
      content_type_id: selectedType.value.id,
      sort_field: sortField.value,
      sort_order: sortOrder.value,
    }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }
    const res = await getEntries(params)
    entries.value = res.data.data.items || []
    total.value = res.data.data.total || 0
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

// 选择内容类型
const selectType = (type: ContentType) => {
  selectedType.value = type
  page.value = 1
  loadEntries()
}

// 打开创建弹窗
const openCreateModal = () => {
  editingEntry.value = null
  formData.value = {}
  showModal.value = true
}

// 打开编辑弹窗
const openEditModal = async (entry: Entry) => {
  try {
    const res = await getEntry(entry.id)
    editingEntry.value = res.data.data
    formData.value = { ...res.data.data.values || {} }
    showModal.value = true
  } catch (error) {
    showError(error)
  }
}

// 关闭弹窗
const closeModal = () => {
  showModal.value = false
  editingEntry.value = null
  formData.value = {}
}

// 提交表单
const handleSubmit = async () => {
  if (!selectedType.value) return
  submitting.value = true
  try {
    if (editingEntry.value) {
      await updateEntry(editingEntry.value.id, { values: formData.value } as any)
      showSuccess(t('content.updateSuccess'))
    } else {
      await createEntry({ content_type_id: selectedType.value.id, values: formData.value } as any)
      showSuccess(t('content.createSuccess'))
    }
    closeModal()
    loadEntries()
  } catch (error) {
    showError(error)
  } finally {
    submitting.value = false
  }
}

// 发布/取消发布
const handlePublish = async (entry: Entry) => {
  publishLoading.value = entry.id
  try {
    if (entry.status === 'published') {
      await unpublishEntry(entry.id)
      showSuccess(t('content.unpublishSuccess'))
    } else {
      await publishEntry(entry.id)
      showSuccess(t('content.publishSuccess'))
    }
    loadEntries()
  } catch (error) {
    showError(error)
  } finally {
    publishLoading.value = null
  }
}

// 删除确认
const confirmDelete = (entry: Entry) => {
  entryToDelete.value = entry
  showDeleteConfirm.value = true
}

// 执行删除
const handleDelete = async () => {
  if (!entryToDelete.value) return
  deleteLoading.value = true
  try {
    await deleteEntry(entryToDelete.value.id)
    showDeleteConfirm.value = false
    entryToDelete.value = null
    showSuccess(t('content.deleteSuccess'))
    loadEntries()
  } catch (error) {
    showError(error)
  } finally {
    deleteLoading.value = false
  }
}

// 切换单选
const toggleSelect = (id: string) => {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id)
  } else {
    selectedIds.value.add(id)
  }
  selectedIds.value = new Set(selectedIds.value)
}

// 切换全选
const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value.clear()
  } else {
    entries.value.forEach(entry => {
      selectedIds.value.add(entry.id)
    })
  }
  selectedIds.value = new Set(selectedIds.value)
}

// 批量操作确认
const confirmBatchAction = (action: 'delete' | 'publish' | 'unpublish') => {
  batchAction.value = action
  const labels: Record<string, string> = {
    delete: t('content.batchDelete'),
    publish: t('content.batchPublish'),
    unpublish: t('content.batchUnpublish'),
  }
  batchActionLabel.value = labels[action]
  showBatchConfirm.value = true
}

// 执行批量操作
const executeBatchAction = async () => {
  const ids = Array.from(selectedIds.value)
  if (ids.length === 0) return
  batchLoading.value = true
  try {
    switch (batchAction.value) {
      case 'delete':
        await batchDeleteEntries(ids)
        showSuccess(t('content.deleted', { count: ids.length }))
        break
      case 'publish':
        await batchPublishEntries(ids)
        showSuccess(t('content.publishedCount', { count: ids.length }))
        break
      case 'unpublish':
        await batchUnpublishEntries(ids)
        showSuccess(t('content.unpublishedCount', { count: ids.length }))
        break
    }
    showBatchConfirm.value = false
    selectedIds.value.clear()
    selectedIds.value = new Set()
    loadEntries()
  } catch (error) {
    showError(error)
  } finally {
    batchLoading.value = false
  }
}

// 获取批量操作确认提示
const getBatchConfirmText = () => {
  const count = selectedIds.value.size
  const action = batchActionLabel.value
  if (batchAction.value === 'delete') {
    return t('content.batchDeleteMsg', { count })
  }
  return t('content.batchActionMsg', { action, count })
}

// 清除搜索
const clearSearch = () => {
  searchKeyword.value = ''
  page.value = 1
  loadEntries()
}

// 状态样式
const getStatusClass = (status: string) => {
  const map: Record<string, string> = {
    published: 'badge-success',
    draft: 'badge-warning',
    archived: 'badge-default',
  }
  return map[status] || 'badge-default'
}

const getStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    published: t('content.published'),
    draft: t('content.draft'),
    archived: t('content.archived'),
  }
  return map[status] || status
}

// 格式化日期
const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 监听内容类型变化
watch(() => route.query.type, (newType) => {
  if (newType) {
    const type = contentTypes.value.find(t => t.id === newType)
    if (type) selectType(type)
  }
}, { immediate: true })

onMounted(() => {
  loadContentTypes()
})
</script>

<template>
  <div class="content-management">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('content.title') }}</h1>
        <p class="page-subtitle">{{ t('content.subtitle') }}</p>
      </div>
    </div>

    <div class="content-layout">
      <!-- 无站点提示 -->
      <div v-if="!siteStore.currentSiteId" class="no-site-container">
        <div class="no-site-card">
          <svg width="64" height="64" viewBox="0 0 20 20" fill="currentColor" opacity="0.3">
            <path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"/>
          </svg>
          <h3>{{ t('site.noSiteTitle') || '暂无站点' }}</h3>
          <p>{{ t('site.noSiteHint') || '请先创建一个站点，才能管理内容' }}</p>
          <button class="btn btn-primary" @click="router.push('/')">{{ t('site.goToCreate') || '返回首页创建站点' }}</button>
        </div>
      </div>

      <!-- 有站点时正常显示 -->
      <template v-else>
      <!-- 侧边：内容类型列表 -->
      <aside class="type-sidebar">
        <div class="sidebar-header">
          <h3>{{ t('contentTypes.title') }}</h3>
        </div>
        <div class="type-list">
          <button
            v-for="type in contentTypes"
            :key="type.id"
            class="type-item"
            :class="{ active: selectedType?.id === type.id }"
            @click="selectType(type)"
          >
            <span class="type-icon">
              <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                <path d="M4 4a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"/>
              </svg>
            </span>
            <span class="type-name">{{ type.name }}</span>
            <span class="type-count">{{ type.field_count || 0 }}</span>
          </button>
          <div v-if="contentTypes.length === 0" class="empty-tip">
            {{ t('content.noContentTypes') }}，<router-link to="/content/types">{{ t('content.goToCreate') }}</router-link>
          </div>
        </div>
      </aside>

      <!-- 主内容区 -->
      <main class="content-main">
        <template v-if="selectedType">
          <div class="content-toolbar">
            <div class="toolbar-left">
              <h2>{{ selectedType.name }}</h2>
              <select v-model="statusFilter" class="input" style="width: 120px;" @change="loadEntries">
                <option value="">{{ t('content.allStatus') }}</option>
                <option value="draft">{{ t('content.draft') }}</option>
                <option value="published">{{ t('content.published') }}</option>
                <option value="archived">{{ t('content.archived') }}</option>
              </select>

              <!-- 搜索框 -->
              <div class="search-box">
                <svg class="search-icon" width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z"/>
                </svg>
                <input
                  v-model="searchKeyword"
                  type="text"
                  class="input search-input"
                  :placeholder="t('content.searchContent')"
                  @keyup.enter="loadEntries"
                />
                <button v-if="searchKeyword" class="search-clear" @click="clearSearch">
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"/>
                  </svg>
                </button>
              </div>

              <!-- 排序 -->
              <select v-model="sortField" class="input" style="width: 130px;" @change="loadEntries">
                <option value="updated_time">{{ t('content.sortByUpdated') }}</option>
                <option value="created_time">{{ t('content.sortByCreated') }}</option>
                <option value="published_time">{{ t('content.sortByPublished') }}</option>
                <option value="sort_weight">{{ t('content.sortByWeight') }}</option>
              </select>
              <button class="btn btn-secondary btn-sm" @click="sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'; loadEntries()">
                {{ sortOrder === 'asc' ? t('content.asc') : t('content.desc') }}
              </button>
            </div>

            <div class="toolbar-right">
              <!-- 批量操作 -->
              <div v-if="hasSelected" class="batch-actions">
                <span class="selected-count">{{ t('common.selectedCount', { count: selectedCount }) }}</span>
                <button class="btn btn-secondary btn-sm" :disabled="batchLoading" @click="confirmBatchAction('publish')">
                  {{ t('content.batchPublish') }}
                </button>
                <button class="btn btn-secondary btn-sm" :disabled="batchLoading" @click="confirmBatchAction('unpublish')">
                  {{ t('content.batchUnpublish') }}
                </button>
                <button class="btn btn-danger btn-sm" :disabled="batchLoading" @click="confirmBatchAction('delete')">
                  {{ t('content.batchDelete') }}
                </button>
              </div>
              <button class="btn btn-primary" @click="openCreateModal">
                <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"/>
                </svg>
                {{ t('content.createEntry') }}
              </button>
            </div>
          </div>

          <!-- 表格 -->
          <div class="card" style="padding: 0; overflow: hidden;">
            <table class="table">
              <thead>
                <tr>
                  <th class="checkbox-col">
                    <input
                      type="checkbox"
                      :checked="isAllSelected"
                      :indeterminate="hasSelected && !isAllSelected"
                      @change="toggleSelectAll"
                    />
                  </th>
                  <th>{{ t('common.id') }}</th>
                  <th v-for="field in selectedType.fields?.slice(0, 3)" :key="field.id">
                    {{ field.name }}
                  </th>
                  <th>{{ t('common.status') }}</th>
                  <th>{{ t('common.updatedAt') }}</th>
                  <th>{{ t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="loading">
                  <td colspan="7" class="loading-state">
                    <div class="spinner"></div>
                    <span>{{ t('common.loading') }}</span>
                  </td>
                </tr>
                <tr v-else-if="entries.length === 0">
                  <td colspan="7" class="empty-state">
                    <div class="empty-icon">
                      <svg width="48" height="48" viewBox="0 0 20 20" fill="currentColor" opacity="0.3">
                        <path d="M4 4a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"/>
                      </svg>
                    </div>
                    <p class="empty-title">{{ t('content.noContent') }}</p>
                    <p class="empty-desc">{{ t('content.createFirst') }}</p>
                    <button class="btn btn-primary" @click="openCreateModal">
                      {{ t('content.createFirstEntry') }}
                    </button>
                  </td>
                </tr>
                <tr v-else v-for="entry in entries" :key="entry.id" :class="{ selected: selectedIds.has(entry.id) }">
                  <td class="checkbox-col">
                    <input
                      type="checkbox"
                      :checked="selectedIds.has(entry.id)"
                      @change="toggleSelect(entry.id)"
                    />
                  </td>
                  <td class="id-cell">{{ entry.id.slice(0, 8) }}</td>
                  <td v-for="field in selectedType.fields?.slice(0, 3)" :key="field.id">
                    {{ entry.values?.[field.name] || '-' }}
                  </td>
                  <td>
                    <span :class="['badge', getStatusClass(entry.status)]">
                      {{ getStatusLabel(entry.status) }}
                    </span>
                  </td>
                  <td>{{ formatDate(entry.updated_time) }}</td>
                  <td class="actions-cell">
                    <button class="btn btn-secondary btn-sm" :disabled="publishLoading === entry.id" @click="openEditModal(entry)">{{ t('common.edit') }}</button>
                    <button
                      class="btn btn-sm"
                      :class="entry.status === 'published' ? 'btn-secondary' : 'btn-primary'"
                      :disabled="publishLoading === entry.id"
                      @click="handlePublish(entry)"
                    >
                      <span v-if="publishLoading === entry.id" class="btn-spinner"></span>
                      {{ entry.status === 'published' ? t('content.unpublish') : t('content.publish') }}
                    </button>
                    <button class="btn btn-danger btn-sm" :disabled="deleteLoading" @click="confirmDelete(entry)">{{ t('common.delete') }}</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- 分页 -->
          <div class="pagination" v-if="total > pageSize">
            <span class="pagination-info">{{ t('common.total') }} {{ total }} {{ t('common.items') }}</span>
            <button
              class="btn btn-secondary btn-sm"
              :disabled="page === 1"
              @click="page--; loadEntries()"
            >
              {{ t('common.prevPage') }}
            </button>
            <span class="pagination-current">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
            <button
              class="btn btn-secondary btn-sm"
              :disabled="page >= Math.ceil(total / pageSize)"
              @click="page++; loadEntries()"
            >
              {{ t('common.nextPage') }}
            </button>
          </div>
        </template>

        <div v-else class="empty-state">
          <svg width="64" height="64" viewBox="0 0 20 20" fill="currentColor" opacity="0.3">
            <path d="M4 4a2 2 0 012-2h8a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V4z"/>
          </svg>
          <h3>{{ t('content.selectType') }}</h3>
          <p>{{ t('content.selectTypeHint') }}</p>
        </div>
      </main>
      </template>
    </div>

    <!-- 创建/编辑弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ editingEntry ? t('content.createEntry') : t('content.createEntry') }}</h3>
          <button class="modal-close" @click="closeModal">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div v-for="field in selectedType?.fields" :key="field.id" class="input-group">
            <label class="input-label">
              {{ field.name }}
              <span v-if="field.required" class="required">*</span>
            </label>
            <input
              v-if="field.field_type === 'text' || field.field_type === 'email' || field.field_type === 'url'"
              v-model="formData[field.name]"
              type="text"
              class="input"
              :placeholder="t('content.enterField', { fieldName: field.name })"
            />
            <textarea
              v-else-if="field.field_type === 'rich_text' || field.field_type === 'json'"
              v-model="formData[field.name]"
              class="input"
              rows="4"
              :placeholder="t('content.enterField', { fieldName: field.name })"
            ></textarea>
            <input
              v-else-if="field.field_type === 'number'"
              v-model="formData[field.name]"
              type="number"
              class="input"
              :placeholder="t('content.enterNumber')"
            />
            <input
              v-else-if="field.field_type === 'date'"
              v-model="formData[field.name]"
              type="date"
              class="input"
            />
            <input
              v-else-if="field.field_type === 'datetime'"
              v-model="formData[field.name]"
              type="datetime-local"
              class="input"
            />
            <label v-else-if="field.field_type === 'boolean'" class="checkbox-label">
              <input v-model="formData[field.name]" type="checkbox" />
              <span>{{ formData[field.name] ? t('content.yes') : t('content.no') }}</span>
            </label>
            <select
              v-else-if="field.field_type === 'enum' && (field.options || field.config?.options)"
              v-model="formData[field.name]"
              class="input"
            >
              <option value="">{{ t('content.select') }}</option>
              <option v-for="opt in (field.options || field.config?.options)" :key="opt" :value="opt">{{ opt }}</option>
            </select>
            <input
              v-else
              v-model="formData[field.name]"
              type="text"
              class="input"
              :placeholder="t('content.enterField', { fieldName: field.name })"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal" :disabled="submitting">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="submitting" @click="handleSubmit">
            <span v-if="submitting" class="btn-spinner"></span>
            {{ submitting ? t('common.processing') : (editingEntry ? t('common.save') : t('common.create')) }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ t('common.confirmDelete') }}</h3>
        </div>
        <div class="modal-body">
          <p>{{ t('content.confirmDelete') }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteConfirm = false" :disabled="deleteLoading">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="deleteLoading" @click="handleDelete">
            <span v-if="deleteLoading" class="btn-spinner"></span>
            {{ deleteLoading ? t('common.deleting') : t('common.delete') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量操作确认弹窗 -->
    <div v-if="showBatchConfirm" class="modal-overlay" @click.self="showBatchConfirm = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ t('common.confirmAction', { action: batchActionLabel }) }}</h3>
        </div>
        <div class="modal-body">
          <p>{{ getBatchConfirmText() }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showBatchConfirm = false" :disabled="batchLoading">{{ t('common.cancel') }}</button>
          <button
            class="btn"
            :class="batchAction === 'delete' ? 'btn-danger' : 'btn-primary'"
            :disabled="batchLoading"
            @click="executeBatchAction"
          >
            <span v-if="batchLoading" class="btn-spinner"></span>
            {{ batchLoading ? t('common.processing') : batchActionLabel }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.content-management {
  height: 100%;
}

.content-layout {
  display: flex;
  gap: 20px;
  height: calc(100vh - 160px);
}

.type-sidebar {
  width: 240px;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
}

.sidebar-header h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.type-list {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}

.type-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s;
  color: var(--color-text);
}

.type-item:hover {
  background: var(--color-hover);
}

.type-item.active {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.type-icon {
  display: flex;
  color: var(--color-text-secondary);
}

.type-item.active .type-icon {
  color: var(--color-primary);
}

.type-name {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
}

.type-count {
  font-size: 12px;
  color: var(--color-text-secondary);
  background: var(--color-hover);
  padding: 2px 8px;
  border-radius: 10px;
}

.empty-tip {
  padding: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.empty-tip a {
  color: var(--color-primary);
}

.content-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.content-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.toolbar-left h2 {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.batch-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--color-primary-light);
  border-radius: 8px;
}

.selected-count {
  font-size: 13px;
  color: var(--color-primary);
  font-weight: 500;
  margin-right: 4px;
}

/* 搜索框 */
.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 10px;
  color: var(--color-text-secondary);
  pointer-events: none;
}

.search-input {
  padding-left: 34px !important;
  padding-right: 34px !important;
  width: 200px;
}

.search-clear {
  position: absolute;
  right: 8px;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: 4px;
}

.search-clear:hover {
  background: var(--color-hover);
  color: var(--color-text);
}

/* 多选列 */
.checkbox-col {
  width: 40px;
  text-align: center;
}

.checkbox-col input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--color-primary);
}

/* 选中行样式 */
tr.selected {
  background: var(--color-primary-light) !important;
}

.id-cell {
  font-family: monospace;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.actions-cell {
  display: flex;
  gap: 8px;
}

.text-center {
  text-align: center;
  padding: 40px !important;
  color: var(--color-text-secondary);
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 20px;
}

.pagination-info {
  font-size: 14px;
  color: var(--color-text-secondary);
}

.pagination-current {
  font-size: 14px;
  color: var(--color-text);
}

.required {
  color: var(--color-error);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
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
  width: 600px;
  max-height: 80vh;
  background: var(--color-card);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-sm {
  width: 400px;
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
  color: var(--color-text);
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

/* Loading 状态 */
.loading-state {
  text-align: center;
  padding: 40px !important;
  color: var(--color-text-secondary);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 按钮内 loading */
.btn-spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-right: 6px;
  vertical-align: middle;
}

/* 空态优化 */
.empty-state {
  text-align: center;
  padding: 48px 24px !important;
}

.empty-icon {
  margin-bottom: 16px;
}

.empty-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text);
  margin: 0 0 8px;
}

.empty-desc {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0 0 20px;
}

/* 按钮禁用状态 */
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn:disabled .btn-spinner {
  border-color: currentColor;
  border-top-color: transparent;
}

/* 无站点提示 */
.no-site-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
}

.no-site-card {
  text-align: center;
  max-width: 400px;
  padding: 40px;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
}

.no-site-card h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
  margin: 16px 0 8px;
}

.no-site-card p {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0 0 24px;
}
</style>
