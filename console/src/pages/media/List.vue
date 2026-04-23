<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  getAssets,
  getAssetFolders,
  createAsset,
  createFolder,
  deleteAsset,
  type Asset,
  type AssetFolder,
} from '@/api/asset'
import { showError, showSuccess } from '@/utils/request'

const { t } = useI18n()

const loading = ref(false)
const assets = ref<Asset[]>([])
const folders = ref<AssetFolder[]>([])
const selectedFolder = ref<string | null>(null)
const viewMode = ref<'grid' | 'list'>('grid')
const showUploadModal = ref(false)
const showNewFolderModal = ref(false)
const showDeleteConfirm = ref(false)
const assetToDelete = ref<Asset | null>(null)
const selectedAssets = ref<Set<string>>(new Set())
const uploading = ref(false)
const uploadProgress = ref(0)

// 新建文件夹
const newFolderName = ref('')

// 拖拽状态
const isDragging = ref(false)

// 分页
const page = ref(1)
const pageSize = ref(24)
const total = ref(0)

// 过滤器
const typeFilter = ref<string>('')
const searchKeyword = ref('')

// 加载媒体列表
const loadAssets = async () => {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (selectedFolder.value) {
      params.folder_id = selectedFolder.value
    }
    if (typeFilter.value) {
      params.type = typeFilter.value
    }
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }
    const res = await getAssets(params)
    assets.value = res.data.items || []
    total.value = res.data.total || 0
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

// 加载文件夹
const loadFolders = async () => {
  try {
    const res = await getAssetFolders()
    folders.value = res.data || []
  } catch (error) {
    showError(error)
  }
}

// 上传文件
const handleUpload = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return

  uploading.value = true
  uploadProgress.value = 0

  const file = input.files[0]
  try {
    await createAsset({
      file,
      folder_id: selectedFolder.value,
    })
    showSuccess(t('media.uploadSuccess'))
    await loadAssets()
    showUploadModal.value = false
  } catch (error) {
    showError(error)
  } finally {
    uploading.value = false
    input.value = ''
  }
}

// 拖拽上传
const handleDrop = async (event: DragEvent) => {
  isDragging.value = false
  const files = event.dataTransfer?.files
  if (!files?.length) return

  uploading.value = true
  try {
    for (let i = 0; i < files.length; i++) {
      await createAsset({
        file: files[i],
        folder_id: selectedFolder.value,
      })
    }
    await loadAssets()
    showSuccess(t('media.uploadSuccess'))
  } catch (error) {
    showError(error)
  } finally {
    showUploadModal.value = false
  }
}

const handleCreateFolder = async () => {
  if (!newFolderName.value.trim()) return

  try {
    await createFolder({
      name: newFolderName.value,
      parent_id: selectedFolder.value,
    })
    await loadFolders()
    showNewFolderModal.value = false
    newFolderName.value = ''
  } catch (error) {
    showError(error)
  }
}

// 删除文件
const confirmDelete = (asset: Asset) => {
  assetToDelete.value = asset
  showDeleteConfirm.value = true
}

const handleDelete = async () => {
  if (!assetToDelete.value) return

  try {
    await deleteAsset(assetToDelete.value.id)
    showSuccess(t('media.deleteSuccess'))
    showDeleteConfirm.value = false
    assetToDelete.value = null
    await loadAssets()
  } catch (error) {
    showError(error)
  }
}

// 选择文件
const toggleSelect = (asset: Asset) => {
  if (selectedAssets.value.has(asset.id)) {
    selectedAssets.value.delete(asset.id)
  } else {
    selectedAssets.value.add(asset.id)
  }
}

const selectAll = () => {
  if (selectedAssets.value.size === assets.value.length) {
    selectedAssets.value.clear()
  } else {
    assets.value.forEach(a => selectedAssets.value.add(a.id))
  }
}

// 格式化文件大小
const formatSize = (bytes: number) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

// 获取文件类型图标
const getFileIcon = (type: string) => {
  if (type.startsWith('image/')) return 'image'
  if (type.startsWith('video/')) return 'video'
  if (type.startsWith('audio/')) return 'audio'
  if (type.includes('pdf')) return 'pdf'
  if (type.includes('word') || type.includes('document')) return 'doc'
  if (type.includes('sheet') || type.includes('excel')) return 'sheet'
  return 'file'
}

// 判断是否为图片
const isImage = (asset: Asset) => {
  return asset.mime_type?.startsWith('image/')
}

onMounted(() => {
  loadAssets()
  loadFolders()
})
</script>

<template>
  <div class="media-library">
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('media.title') }}</h1>
        <p class="page-subtitle">{{ t('media.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="showNewFolderModal = true">
          <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
            <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"/>
          </svg>
          {{ t('media.newFolder') }}
        </button>
        <button class="btn btn-primary" @click="showUploadModal = true">
          <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
            <path d="M10 3a1 1 0 011 1v5.586l1.707-1.707a1 1 0 011.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 011.414-1.414L9 9.586V4a1 1 0 011-1z"/>
          </svg>
          {{ t('media.upload') }}
        </button>
      </div>
    </div>

    <div class="media-layout">
      <!-- 工具栏 -->
      <div class="media-toolbar">
        <div class="toolbar-left">
          <input
            v-model="searchKeyword"
            type="text"
            class="input"
            :placeholder="t('media.searchFiles')"
            style="width: 240px;"
            @keyup.enter="loadAssets"
          />
          <select v-model="typeFilter" class="input" style="width: 120px;" @change="loadAssets">
            <option value="">{{ t('media.allTypes') }}</option>
            <option value="image">{{ t('media.image') }}</option>
            <option value="video">{{ t('media.video') }}</option>
            <option value="audio">{{ t('media.audio') }}</option>
            <option value="document">{{ t('media.document') }}</option>
          </select>
          <button class="btn btn-secondary btn-sm" @click="loadAssets">{{ t('media.searchBtn') }}</button>
        </div>
        <div class="toolbar-right">
          <span class="selection-info" v-if="selectedAssets.size > 0">
            {{ t('media.selectedFiles', { count: selectedAssets.size }) }}
          </span>
          <button
            class="btn btn-secondary btn-sm"
            :class="{ active: viewMode === 'grid' }"
            @click="viewMode = 'grid'"
          >
            <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
              <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/>
            </svg>
          </button>
          <button
            class="btn btn-secondary btn-sm"
            :class="{ active: viewMode === 'list' }"
            @click="viewMode = 'list'"
          >
            <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- 拖拽区域 -->
      <div
        class="drop-zone"
        :class="{ dragging: isDragging }"
        @dragover.prevent="isDragging = true"
        @dragleave="isDragging = false"
        @drop.prevent="handleDrop"
      >
        <!-- 网格视图 -->
        <div v-if="viewMode === 'grid'" class="asset-grid">
          <div v-if="loading" class="loading">{{ t('common.loading') }}</div>
          <template v-else>
            <div
              v-for="asset in assets"
              :key="asset.id"
              class="asset-card"
              :class="{ selected: selectedAssets.has(asset.id) }"
              @click="toggleSelect(asset)"
            >
              <div class="asset-preview">
                <img v-if="isImage(asset)" :src="asset.url" :alt="asset.name" />
                <div v-else class="file-icon" :class="getFileIcon(asset.mime_type || '')">
                  <svg width="32" height="32" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M4 4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm0 2h12v7l-4-3-2 1.5L6 12V5z"/>
                  </svg>
                </div>
                <div class="asset-overlay">
                  <button class="btn btn-danger btn-sm" @click.stop="confirmDelete(asset)">
                    {{ t('common.delete') }}
                  </button>
                </div>
              </div>
              <div class="asset-info">
                <div class="asset-name" :title="asset.name">{{ asset.name }}</div>
                <div class="asset-meta">
                  <span>{{ formatSize(asset.size) }}</span>
                  <span>{{ new Date(asset.created_time).toLocaleDateString() }}</span>
                </div>
              </div>
              <div v-if="selectedAssets.has(asset.id)" class="selected-badge">
                <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"/>
                </svg>
              </div>
            </div>
          </template>

          <div v-if="!loading && assets.length === 0" class="empty-state">
            <svg width="64" height="64" viewBox="0 0 20 20" fill="currentColor" opacity="0.3">
              <path d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm0 2h12v7l-4-3-2 1.5L6 12V5z"/>
            </svg>
            <h3>{{ t('media.noMedia') }}</h3>
            <p>{{ t('media.noMediaHint') }}</p>
            <button class="btn btn-primary" @click="showUploadModal = true">{{ t('media.uploadFirstFile') }}</button>
          </div>
        </div>

        <!-- 列表视图 -->
        <div v-else class="asset-list">
          <table class="table">
            <thead>
              <tr>
                <th style="width: 40px;">
                  <input
                    type="checkbox"
                    :checked="selectedAssets.size === assets.length && assets.length > 0"
                    @change="selectAll"
                  />
                </th>
                <th>{{ t('media.fileName') }}</th>
                <th style="width: 100px;">{{ t('media.fileType') }}</th>
                <th style="width: 100px;">{{ t('media.fileSize') }}</th>
                <th style="width: 150px;">{{ t('media.uploadTime') }}</th>
                <th style="width: 120px;">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading">
                <td colspan="6" class="text-center">{{ t('common.loading') }}</td>
              </tr>
              <tr v-else-if="assets.length === 0">
                <td colspan="6" class="empty-state">
                  <p>{{ t('media.noMedia') }}</p>
                </td>
              </tr>
              <tr v-else v-for="asset in assets" :key="asset.id">
                <td>
                  <input
                    type="checkbox"
                    :checked="selectedAssets.has(asset.id)"
                    @change="toggleSelect(asset)"
                  />
                </td>
                <td>
                  <div class="file-cell">
                    <img v-if="isImage(asset)" :src="asset.url" class="file-thumb" />
                    <span class="file-name" :title="asset.name">{{ asset.name }}</span>
                  </div>
                </td>
                <td>
                  <span class="type-badge">{{ asset.mime_type?.split('/')[1] || '-' }}</span>
                </td>
                <td>{{ formatSize(asset.size) }}</td>
                <td>{{ new Date(asset.created_time).toLocaleDateString() }}</td>
                <td>
                  <button class="btn btn-secondary btn-sm" @click="confirmDelete(asset)">{{ t('common.delete') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination" v-if="total > pageSize">
        <span class="pagination-info">{{ t('media.totalFiles', { total }) }}</span>
        <button
          class="btn btn-secondary btn-sm"
          :disabled="page === 1"
          @click="page--; loadAssets()"
        >
          {{ t('media.prevPage') }}
        </button>
        <span class="pagination-current">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button
          class="btn btn-secondary btn-sm"
          :disabled="page >= Math.ceil(total / pageSize)"
          @click="page++; loadAssets()"
        >
          {{ t('media.nextPage') }}
        </button>
      </div>
    </div>

    <!-- 上传弹窗 -->
    <div v-if="showUploadModal" class="modal-overlay" @click.self="showUploadModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ t('media.uploadFile') }}</h3>
          <button class="modal-close" @click="showUploadModal = false">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div
            class="upload-zone"
            :class="{ uploading }"
            @click="($refs.fileInput as HTMLInputElement).click()"
          >
            <input
              ref="fileInput"
              type="file"
              hidden
              @change="handleUpload"
            />
            <div v-if="uploading" class="upload-progress">
              <div class="spinner"></div>
              <p>{{ t('media.uploading') }}</p>
            </div>
            <template v-else>
              <svg width="48" height="48" viewBox="0 0 20 20" fill="currentColor" opacity="0.3">
                <path d="M10 3a1 1 0 011 1v5.586l1.707-1.707a1 1 0 011.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 011.414-1.414L9 9.586V4a1 1 0 011-1z"/>
                <path d="M4 16a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1z"/>
              </svg>
              <p>{{ t('media.dragTip') }}</p>
              <span class="upload-hint">{{ t('media.fileTypeTip') }}</span>
            </template>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建文件夹弹窗 -->
    <div v-if="showNewFolderModal" class="modal-overlay" @click.self="showNewFolderModal = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ t('media.createFolder') }}</h3>
        </div>
        <div class="modal-body">
          <div class="input-group">
            <label class="input-label">{{ t('media.folderName') }}</label>
            <input
              v-model="newFolderName"
              type="text"
              class="input"
              :placeholder="t('media.enterFolderName')"
              @keyup.enter="handleCreateFolder"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showNewFolderModal = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" @click="handleCreateFolder">{{ t('common.create') }}</button>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>{{ t('media.confirmDeleteFile') }}</h3>
        </div>
        <div class="modal-body">
          <p>{{ t('media.deleteFileMsg') }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteConfirm = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" @click="handleDelete">{{ t('common.delete') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.media-library {
  height: 100%;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.media-layout {
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 20px;
}

.media-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selection-info {
  font-size: 13px;
  color: var(--color-primary);
  margin-right: 12px;
}

.btn.active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.drop-zone {
  min-height: 400px;
}

.drop-zone.dragging {
  background: var(--color-primary-light);
  border: 2px dashed var(--color-primary);
}

.asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 16px;
}

.asset-card {
  background: var(--color-hover);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.asset-card:hover {
  border-color: var(--color-primary);
}

.asset-card.selected {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px var(--color-primary-light);
}

.asset-preview {
  height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg);
  overflow: hidden;
}

.asset-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.file-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--color-text-secondary);
}

.file-icon.image { color: #10b981; }
.file-icon.video { color: #8b5cf6; }
.file-icon.audio { color: #f59e0b; }
.file-icon.pdf { color: #ef4444; }
.file-icon.doc { color: #3b82f6; }
.file-icon.sheet { color: #10b981; }

.asset-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.asset-card:hover .asset-overlay {
  opacity: 1;
}

.selected-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  background: var(--color-primary);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.asset-info {
  padding: 12px;
}

.asset-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.asset-meta {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.asset-list {
  overflow-x: auto;
}

.file-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-thumb {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 4px;
}

.file-name {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.type-badge {
  font-size: 12px;
  padding: 2px 8px;
  background: var(--color-hover);
  border-radius: 4px;
  color: var(--color-text-secondary);
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

.loading {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px;
  color: var(--color-text-secondary);
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
  background: var(--color-card);
  border-radius: 12px;
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
}

.modal-body {
  padding: 20px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--color-border);
}

.upload-zone {
  border: 2px dashed var(--color-border);
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-zone:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.upload-zone.uploading {
  pointer-events: none;
}

.upload-progress {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.upload-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
}
</style>
