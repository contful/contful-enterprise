<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next'
import Icon from '@/components/Icon.vue'
import {
  getAssets,
  getAssetFolders,
  createAsset,
  createFolder,
  deleteAsset,
  type Asset,
  type FolderResponse,
} from '@/api/asset'
import { showError, showSuccess } from '@/utils/request'
import PageHeader from '@/components/PageHeader.vue'

const { t } = useI18n()

const loading = ref(false)
const assets = ref<Asset[]>([])
const folders = ref<FolderResponse[]>([])
const selectedFolder = ref<string | null>(null)
const viewMode = ref<'grid' | 'list'>('grid')
const showUploadModal = ref(false)
const showNewFolderModal = ref(false)
const selectedAssets = ref<Set<string>>(new Set())
const uploading = ref(false)
const uploadFolderId = ref<string | null>(null)

// 新建文件夹
const newFolderName = ref('')
const creatingFolder = ref(false)

// 打开上传弹窗
function openUpload() {
  uploadFolderId.value = selectedFolder.value
  showUploadModal.value = true
}

// 切换文件夹
function selectFolder(folderId: string | null) {
  selectedFolder.value = folderId
  page.value = 1
  loadAssets()
}

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
    const res = await getAssets(params) as any
    assets.value = res.items || []
    total.value = res.total || 0
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

// 加载文件夹
const loadFolders = async () => {
  try {
    const data = await getAssetFolders()
    folders.value = data || []
  } catch (error) {
    showError(error)
  }
}

// 上传文件
const handleUpload = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return

  uploading.value = true
  const file = input.files[0]
  try {
    await createAsset({
      file,
      folder_id: uploadFolderId.value || undefined,
    })
    MessagePlugin.success(t('media.uploadSuccess'))
    await loadAssets()
    showUploadModal.value = false
    uploadFolderId.value = null
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
        folder_id: uploadFolderId.value || undefined,
      })
    }
    await loadAssets()
    MessagePlugin.success(t('media.uploadSuccess'))
    uploadFolderId.value = null
  } catch (error) {
    showError(error)
  } finally {
    showUploadModal.value = false
  }
}

// 新建文件夹 — t-dialog 内提交
const handleCreateFolder = async () => {
  if (!newFolderName.value.trim()) return
  creatingFolder.value = true
  try {
    await createFolder({
      name: newFolderName.value,
      parent_id: selectedFolder.value || undefined,
    })
    await loadFolders()
    showNewFolderModal.value = false
    newFolderName.value = ''
    MessagePlugin.success(t('common.createSuccess'))
  } catch (error) {
    showError(error)
  } finally {
    creatingFolder.value = false
  }
}

// 删除文件 — DialogPlugin.confirm
const confirmDelete = (asset: Asset) => {
  DialogPlugin.confirm({
    header: t('media.confirmDeleteFile'),
    body: t('media.deleteFileMsg', { name: asset.name }),
    theme: 'warning',
    onConfirm: async () => {
      try {
        await deleteAsset(asset.id)
        showSuccess(t('media.deleteSuccess'))
        await loadAssets()
      } catch (error) {
        showError(error)
      }
    },
  })
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

// 分页变化
const onPageChange = ({ current, pageSize: ps }: { current: number; pageSize: number }) => {
  page.value = current
  pageSize.value = ps
  loadAssets()
}

onMounted(() => {
  loadAssets()
  loadFolders()
})
</script>

<template>
  <div class="page page--padded media-library">
    <PageHeader
      :title="t('media.title')"
      :subtitle="t('media.subtitle')"
      :show-refresh="true"
      @refresh="loadMedia(currentFolder, currentPage)"
    >
      <template #actions>
        <t-button variant="outline" @click="showNewFolderModal = true">
          <template #icon><t-icon name="folder-add" /></template>
          {{ t('media.newFolder') }}
        </t-button>
      </template>
      <template #primary-action>
        <t-button theme="primary" @click="openUpload">
          <Icon name="arrow-up" />
          {{ t('media.upload') }}
        </t-button>
      </template>
    </PageHeader>

    <div class="media-layout">
      <!-- 文件夹侧边栏 -->
      <div class="folder-sidebar">
        <div class="folder-header">
          <span>{{ t('media.folders') }}</span>
        </div>
        <div class="folder-list">
          <div
            class="folder-item"
            :class="{ active: selectedFolder === null }"
            @click="selectFolder(null)"
          >
            <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
              <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"/>
            </svg>
            <span>{{ t('media.allFiles') }}</span>
          </div>
          <template v-for="folder in folders" :key="folder.id">
            <div
              class="folder-item"
              :class="{ active: selectedFolder === folder.id }"
              @click="selectFolder(folder.id)"
            >
              <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"/>
              </svg>
              <span>{{ folder.name }}</span>
            </div>
          </template>
        </div>
      </div>

      <!-- 主内容区 -->
      <div class="media-main">
        <!-- 工具栏 -->
        <div class="media-toolbar">
          <div class="toolbar-left">
            <!-- 搜索框 → t-input -->
            <t-input
              v-model="searchKeyword"
              :placeholder="t('media.searchFiles')"
              clearable
              size="small"
              style="width: 240px"
              @enter="loadAssets"
              @clear="loadAssets"
            >
              <template #prefixIcon><t-icon name="search" /></template>
            </t-input>
            <!-- 类型过滤 → t-select -->
            <t-select
              v-model="typeFilter"
              size="small"
              style="width: 120px"
              clearable
              @change="loadAssets"
            >
              <t-option value="image" :label="t('media.image')" />
              <t-option value="video" :label="t('media.video')" />
              <t-option value="audio" :label="t('media.audio')" />
              <t-option value="document" :label="t('media.document')" />
            </t-select>
            <t-button size="small" variant="outline" @click="loadAssets">{{ t('media.searchBtn') }}</t-button>
          </div>
          <div class="toolbar-right">
            <span v-if="selectedAssets.size > 0" class="selection-info">
              {{ t('media.selectedFiles', { count: selectedAssets.size }) }}
            </span>
            <t-button
              variant="outline"
              :theme="viewMode === 'grid' ? 'primary' : 'default'"
              @click="viewMode = 'grid'"
            >
              <template #icon><t-icon name="layout-grid" /></template>
            </t-button>
            <t-button
              variant="outline"
              :theme="viewMode === 'list' ? 'primary' : 'default'"
              @click="viewMode = 'list'"
            >
              <template #icon><t-icon name="view-list" /></template>
            </t-button>
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
        <!-- 加载中 → t-loading -->
        <div v-if="loading" class="loading-state">
          <t-loading size="medium" />
        </div>

        <!-- 无数据 → t-empty -->
        <t-empty v-else-if="assets.length === 0" :description="t('media.noMediaHint')" />

        <!-- 网格视图 -->
        <div v-else-if="viewMode === 'grid'" class="asset-grid">
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
                <t-button theme="danger" size="small" @click.stop="confirmDelete(asset)">
                  {{ t('common.delete') }}
                </t-button>
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
              <tr v-for="asset in assets" :key="asset.id">
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
                <td><t-tag variant="light" size="small">{{ asset.mime_type?.split('/')[1] || '-' }}</t-tag></td>
                <td>{{ formatSize(asset.size) }}</td>
                <td>{{ new Date(asset.created_time).toLocaleDateString() }}</td>
                <td>
                  <t-button variant="outline" size="small" @click="confirmDelete(asset)">{{ t('common.delete') }}</t-button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      </div><!-- /media-main -->

      <!-- 分页 → t-pagination -->
      <div v-if="total > pageSize" class="pagination-bar">
        <t-pagination
          v-model:current="page"
          v-model:pageSize="pageSize"
          :total="total"
          :show-page-size="true"
          :page-size-options="[12, 24, 48]"
          size="small"
          :total-content="(totalVal: number) => `${t('media.totalFiles', { total: totalVal })}`"
          @change="onPageChange"
        />
      </div>
    </div>

    <!-- 上传弹窗 — t-dialog -->
    <t-dialog
      v-model:visible="showUploadModal"
      :header="t('media.uploadFile')"
      :width="520"
      :cancel-btn="{ content: t('common.cancel') as string }"
      :confirm-btn="null"
      @close="showUploadModal = false"
    >
      <!-- 文件夹选择 -->
      <t-form label-align="top" style="margin-top: 8px">
        <t-form-item :label="t('media.selectFolder') || '\u4e0a\u4f20\u5230'">
          <t-select v-model="uploadFolderId" :options="[
            { label: t('media.noFolder') || '\u6839\u76ee\u5f55', value: '' },
            ...folders.map((f: FolderResponse) => ({ label: f.name, value: f.id })),
          ]" clearable allow-input />
        </t-form-item>
      </t-form>
      <!-- 上传区域 -->
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
          <t-loading size="small" />
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
    </t-dialog>

    <!-- 新建文件夹弹窗 — t-dialog + t-form -->
    <t-dialog
      v-model:visible="showNewFolderModal"
      :header="t('media.createFolder')"
      :width="400"
      :confirm-btn="{ content: t('common.create'), theme: 'primary' as const, loading: creatingFolder }"
      :cancel-btn="{ content: t('common.cancel') }"
      :confirm-on-enter="true"
      @confirm="handleCreateFolder"
    >
      <t-form label-align="top" style="margin-top: 8px">
        <t-form-item :label="`${t('media.folderName')} *`">
          <t-input
            v-model="newFolderName"
            :placeholder="t('media.enterFolderName')"
            clearable
            autofocus
          />
        </t-form-item>
      </t-form>
    </t-dialog>
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
  display: flex;
  gap: 20px;
}

/* === Folder sidebar === */
.folder-sidebar {
  width: 200px;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border);
  padding-right: 16px;
}

.folder-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  margin-bottom: 8px;
  padding: 0 8px;
}

.folder-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.folder-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text);
  transition: background 0.15s;
}

.folder-item:hover {
  background: var(--color-bg-secondary);
}

.folder-item.active {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.folder-item svg {
  flex-shrink: 0;
}

/* === Main content === */
.media-main {
  flex: 1;
  min-width: 0;
}

.media-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
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

/* === Drop zone === */
.drop-zone {
  min-height: 400px;
}

.drop-zone.dragging {
  background: var(--color-primary-light);
  border: 2px dashed var(--color-primary);
  border-radius: 8px;
}

.loading-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

/* === Asset grid === */
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

/* === List view === */
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

/* === Pagination === */
.pagination-bar {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

/* === Upload zone === */
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

.upload-hint {
  font-size: 12px;
  color: var(--color-text-secondary);
}
</style>
