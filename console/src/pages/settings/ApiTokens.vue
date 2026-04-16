<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import {
  getApiTokens,
  createApiToken,
  updateApiToken,
  deleteApiToken,
  regenerateApiToken,
  revokeApiToken,
  type ApiToken,
} from '@/api/api-token'
import { showError } from '@/utils/request'

const loading = ref(false)
const tokens = ref<ApiToken[]>([])
const showModal = ref(false)
const showDeleteConfirm = ref(false)
const showRegenerateConfirm = ref(false)
const editingToken = ref<ApiToken | null>(null)
const tokenToDelete = ref<ApiToken | null>(null)
const tokenToRegenerate = ref<ApiToken | null>(null)
const newToken = ref('')
const submitting = ref(false) // MX-001: Loading 状态

// 表单数据
const formData = ref({
  name: '',
  description: '',
  expires_in_days: 365,
  permissions: [] as string[],
  rate_limit: 1000,
})

// 权限选项
const permissionOptions = [
  { value: 'content:read', label: '读取内容' },
  { value: 'content:write', label: '写入内容' },
  { value: 'assets:read', label: '读取媒体' },
  { value: 'assets:write', label: '写入媒体' },
]

// 加载 Token 列表
const loadTokens = async () => {
  loading.value = true
  try {
    const res = await getApiTokens({ page: 1, page_size: 100 })
    tokens.value = res.data.items || []
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

// 打开创建弹窗
const openCreateModal = () => {
  editingToken.value = null
  formData.value = {
    name: '',
    description: '',
    expires_in_days: 365,
    permissions: [],
    rate_limit: 1000,
  }
  showModal.value = true
}

// 打开编辑弹窗
const openEditModal = (token: ApiToken) => {
  editingToken.value = token
  formData.value = {
    name: token.name,
    description: token.description || '',
    expires_in_days: token.expires_in_days || 365,
    permissions: token.permissions || [],
    rate_limit: token.rate_limit || 1000,
  }
  showModal.value = true
}

// 提交表单
const handleSubmit = async () => {
  submitting.value = true
  try {
    if (editingToken.value) {
      await updateApiToken(editingToken.value.id, {
        name: formData.value.name,
        description: formData.value.description,
        permissions: formData.value.permissions,
        rate_limit: formData.value.rate_limit,
      })
      MessagePlugin.success('Token 已更新')
    } else {
      const res = await createApiToken({
        name: formData.value.name,
        description: formData.value.description,
        expires_in_days: formData.value.expires_in_days,
        permissions: formData.value.permissions,
        rate_limit: formData.value.rate_limit,
      })
      newToken.value = res.data.token || ''
      MessagePlugin.success('Token 创建成功')
    }
    showModal.value = false
    await loadTokens()
  } catch (error) {
    showError(error)
  } finally {
    submitting.value = false
  }
}

// 删除确认
const confirmDelete = (token: ApiToken) => {
  tokenToDelete.value = token
  showDeleteConfirm.value = true
}

// 执行删除
const handleDelete = async () => {
  if (!tokenToDelete.value) return

  try {
    await deleteApiToken(tokenToDelete.value.id)
    MessagePlugin.success('Token 已删除')
    showDeleteConfirm.value = false
    tokenToDelete.value = null
    await loadTokens()
  } catch (error) {
    showError(error)
  }
}

// 重新生成 Token
const confirmRegenerate = (token: ApiToken) => {
  tokenToRegenerate.value = token
  showRegenerateConfirm.value = true
}

const handleRegenerate = async () => {
  if (!tokenToRegenerate.value) return

  try {
    const res = await regenerateApiToken(tokenToRegenerate.value.id)
    newToken.value = res.data.token || ''
    showRegenerateConfirm.value = false
    tokenToRegenerate.value = null
    await loadTokens()
  } catch (error) {
    console.error('Failed to regenerate token:', error)
  }
}

// 撤销 Token
const handleRevoke = async (token: ApiToken) => {
  try {
    await revokeApiToken(token.id)
    await loadTokens()
  } catch (error) {
    console.error('Failed to revoke token:', error)
  }
}

// 复制 Token
const copyToken = (token: string) => {
  navigator.clipboard.writeText(token)
}

// 关闭弹窗
const closeModal = () => {
  showModal.value = false
  newToken.value = ''
}

// 格式化日期
const formatDate = (date: string | null) => {
  if (!date) return '永久有效'
  return new Date(date).toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

// 判断 Token 是否过期
const isExpired = (expiresAt: string | null) => {
  if (!expiresAt) return false
  return new Date(expiresAt) < new Date()
}

// 获取状态标签
const getStatusClass = (token: ApiToken) => {
  if (token.revoked) return 'badge-error'
  if (isExpired(token.expires_at)) return 'badge-warning'
  return 'badge-success'
}

const getStatusLabel = (token: ApiToken) => {
  if (token.revoked) return '已撤销'
  if (isExpired(token.expires_at)) return '已过期'
  return '活跃'
}

onMounted(() => {
  loadTokens()
})
</script>

<template>
  <div class="api-tokens">
    <div class="page-header">
      <div>
        <h1 class="page-title">API Token</h1>
        <p class="page-subtitle">管理 API 访问令牌</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
          <path d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"/>
        </svg>
        创建 Token
      </button>
    </div>

    <!-- Token 列表 -->
    <div class="card" style="padding: 0; overflow: hidden;">
      <table class="table">
        <thead>
          <tr>
            <th>名称</th>
            <th>权限</th>
            <th>限流</th>
            <th>有效期</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="7" class="text-center">加载中...</td>
          </tr>
          <tr v-else-if="tokens.length === 0">
            <td colspan="7" class="empty-state">
              <svg width="48" height="48" viewBox="0 0 20 20" fill="currentColor" opacity="0.3">
                <path d="M7 7a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zm-4 4a1 1 0 100-2 1 1 0 000 2zm4 0a1 1 0 100-2 1 1 0 000 2zM4 5a1 1 0 011-1h10a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5z"/>
              </svg>
              <h3>暂无 API Token</h3>
              <p>创建您的第一个 API Token 来开始使用 Open API</p>
              <button class="btn btn-primary btn-sm" @click="openCreateModal">
                创建 Token
              </button>
            </td>
          </tr>
          <tr v-else v-for="token in tokens" :key="token.id">
            <td>
              <div class="token-info">
                <span class="token-name">{{ token.name }}</span>
                <span v-if="token.description" class="token-desc">{{ token.description }}</span>
              </div>
            </td>
            <td>
              <div class="permissions">
                <span
                  v-for="perm in token.permissions?.slice(0, 2)"
                  :key="perm"
                  class="perm-badge"
                >
                  {{ perm }}
                </span>
                <span v-if="(token.permissions?.length || 0) > 2" class="perm-more">
                  +{{ token.permissions!.length - 2 }}
                </span>
              </div>
            </td>
            <td>{{ token.rate_limit }}/小时</td>
            <td>{{ formatDate(token.expires_at) }}</td>
            <td>
              <span :class="['badge', getStatusClass(token)]">
                {{ getStatusLabel(token) }}
              </span>
            </td>
            <td>{{ new Date(token.created_at).toLocaleDateString('zh-CN') }}</td>
            <td>
              <div class="action-btns">
                <button
                  v-if="!token.revoked && !isExpired(token.expires_at)"
                  class="btn btn-secondary btn-sm"
                  @click="confirmRegenerate(token)"
                  title="重新生成"
                >
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z"/>
                  </svg>
                </button>
                <button
                  v-if="!token.revoked"
                  class="btn btn-secondary btn-sm"
                  @click="handleRevoke(token)"
                  title="撤销"
                >
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"/>
                  </svg>
                </button>
                <button
                  class="btn btn-danger btn-sm"
                  @click="confirmDelete(token)"
                  title="删除"
                >
                  <svg width="14" height="14" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 创建/编辑弹窗 -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ editingToken ? '编辑 Token' : '创建 Token' }}</h3>
          <button class="modal-close" @click="closeModal">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <!-- 新 Token 显示 -->
          <div v-if="newToken" class="new-token-display">
            <div class="new-token-label">新 Token 已创建，请妥善保存：</div>
            <div class="new-token-value">
              <code>{{ newToken }}</code>
              <button class="btn btn-secondary btn-sm" @click="copyToken(newToken)">
                复制
              </button>
            </div>
            <p class="new-token-warning">
              <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"/>
              </svg>
              此 Token 只显示一次，请立即保存
            </p>
          </div>

          <template v-else>
            <div class="input-group">
              <label class="input-label">名称 <span class="required">*</span></label>
              <input
                v-model="formData.name"
                type="text"
                class="input"
                placeholder="例如：生产环境 API"
              />
            </div>

            <div class="input-group">
              <label class="input-label">描述</label>
              <input
                v-model="formData.description"
                type="text"
                class="input"
                placeholder="描述此 Token 的用途"
              />
            </div>

            <div class="input-group">
              <label class="input-label">有效期</label>
              <select v-model="formData.expires_in_days" class="input">
                <option :value="30">30 天</option>
                <option :value="90">90 天</option>
                <option :value="180">180 天</option>
                <option :value="365">1 年</option>
                <option :value="0">永久有效</option>
              </select>
            </div>

            <div class="input-group">
              <label class="input-label">权限</label>
              <div class="permissions-grid">
                <label
                  v-for="opt in permissionOptions"
                  :key="opt.value"
                  class="permission-item"
                >
                  <input
                    type="checkbox"
                    :value="opt.value"
                    v-model="formData.permissions"
                  />
                  <span>{{ opt.label }}</span>
                </label>
              </div>
            </div>

            <div class="input-group">
              <label class="input-label">速率限制 (请求/小时)</label>
              <input
                v-model="formData.rate_limit"
                type="number"
                class="input"
                min="100"
                max="100000"
              />
            </div>
          </template>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeModal">
            {{ newToken ? '关闭' : '取消' }}
          </button>
          <button v-if="!newToken" class="btn btn-primary" @click="handleSubmit">
            创建
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>确认删除</h3>
        </div>
        <div class="modal-body">
          <p>确定要删除此 Token 吗？使用此 Token 的应用将无法再访问 API。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDeleteConfirm = false">取消</button>
          <button class="btn btn-danger" @click="handleDelete">删除</button>
        </div>
      </div>
    </div>

    <!-- 重新生成确认弹窗 -->
    <div v-if="showRegenerateConfirm" class="modal-overlay" @click.self="showRegenerateConfirm = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3>重新生成 Token</h3>
        </div>
        <div class="modal-body">
          <p>重新生成将创建一个新的 Token，原 Token 将立即失效。</p>
          <p class="warning-text">此操作不可撤销，请确保没有应用仍在使用原 Token。</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showRegenerateConfirm = false">取消</button>
          <button class="btn btn-primary" @click="handleRegenerate">重新生成</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.api-tokens {
  height: 100%;
}

.token-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.token-name {
  font-weight: 500;
  color: var(--color-text);
}

.token-desc {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.permissions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.perm-badge {
  font-size: 11px;
  padding: 2px 6px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  border-radius: 4px;
}

.perm-more {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.action-btns {
  display: flex;
  gap: 6px;
}

.text-center {
  text-align: center;
  padding: 60px !important;
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

.required {
  color: var(--color-error);
}

.permissions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.permission-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: var(--color-hover);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.permission-item:hover {
  background: var(--color-primary-light);
}

.permission-item input {
  width: 16px;
  height: 16px;
}

.new-token-display {
  text-align: center;
}

.new-token-label {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin-bottom: 12px;
}

.new-token-value {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--color-hover);
  border-radius: 8px;
  margin-bottom: 12px;
}

.new-token-value code {
  flex: 1;
  font-size: 13px;
  font-family: monospace;
  word-break: break-all;
  color: var(--color-text);
}

.new-token-warning {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 13px;
  color: var(--color-warning);
}

.warning-text {
  font-size: 13px;
  color: var(--color-error);
  margin-top: 8px;
}
</style>
