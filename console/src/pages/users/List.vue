<template>
  <div class="users-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('users.title') }}</h1>
        <p class="page-subtitle">{{ t('users.subtitle') }}</p>
      </div>
      <button class="btn btn-primary" @click="openCreateDialog">
        <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
          <path d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"/>
        </svg>
        {{ t('users.addUser') }}
      </button>
    </div>

    <!-- 用户列表 -->
    <div class="card" style="padding: 0; overflow: hidden;">
      <table class="table">
        <thead>
          <tr>
            <th>{{ t('users.user') }}</th>
            <th>{{ t('users.email') }}</th>
            <th>{{ t('users.role') }}</th>
            <th>{{ t('users.status') }}</th>
            <th>{{ t('users.createdTime') }}</th>
            <th>{{ t('users.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="6" class="text-center">{{ t('users.loading') }}</td>
          </tr>
          <tr v-else-if="users.length === 0">
            <td colspan="6" class="empty-state">
              <h3>{{ t('users.noUsers') }}</h3>
              <p>{{ t('users.noUsersHint') }}</p>
            </td>
          </tr>
          <tr v-else v-for="row in users" :key="row.id">
            <td>
              <div class="user-info">
                <div class="user-avatar">{{ row.email?.charAt(0).toUpperCase() }}</div>
                <span class="user-name">{{ row.nickname || '—' }}</span>
              </div>
            </td>
            <td>{{ row.email }}</td>
            <td>
              <span v-if="row.is_super_admin" class="badge badge-warning">{{ t('users.superAdmin') }}</span>
              <span v-else class="badge badge-default">{{ t('users.normalUser') }}</span>
            </td>
            <td>
              <span :class="['badge', getStatusBadge(row.status)]">{{ getStatusText(row.status) }}</span>
            </td>
            <td>{{ formatDate(row.created_time) }}</td>
            <td>
              <div style="display:flex;gap:8px;">
                <button class="btn btn-secondary btn-sm" @click="openEditDialog(row)">{{ t('common.edit') }}</button>
                <button
                  class="btn btn-sm"
                  :class="row.is_super_admin ? 'btn-secondary' : 'btn-danger'"
                  :disabled="row.is_super_admin"
                  @click="handleDelete(row)"
                >{{ t('common.delete') }}</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div v-if="pagination.total > pagination.pageSize" class="pagination-bar">
      <t-pagination
        v-model="pagination.current"
        :total="pagination.total"
        :page-size="pagination.pageSize"
        @change="onPageChange"
      />
    </div>

    <!-- 创建用户弹窗 -->
    <div v-if="createVisible" class="modal-overlay" @click.self="createVisible = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ t('users.createTitle') }}</h3>
        </div>
        <div class="modal-body">
          <div class="input-group">
            <label class="input-label">{{ t('users.email') }} <span class="required">*</span></label>
            <input v-model="createForm.email" class="input" type="email" :placeholder="t('users.enterEmail')" />
          </div>
          <div class="input-group">
            <label class="input-label">{{ t('users.password') }} <span class="required">*</span></label>
            <input v-model="createForm.password" class="input" type="password" :placeholder="t('users.enterPassword')" />
            <div class="password-strength">
              <div class="strength-bar">
                <div class="strength-fill" :class="passwordStrength.level" :style="{ width: passwordStrength.width }"></div>
              </div>
              <span class="strength-text" :class="passwordStrength.level">{{ passwordStrength.label }}</span>
            </div>
            <div class="password-hint">{{ t('users.passwordHint') }}</div>
          </div>
          <div class="input-group">
            <label class="input-label">{{ t('users.nickname') }}</label>
            <input v-model="createForm.nickname" class="input" type="text" :placeholder="t('users.enterNickname')" />
          </div>
          <div class="input-group">
            <label class="input-label">{{ t('users.superAdminSwitch') }}</label>
            <t-switch v-model="createForm.is_super_admin" />
          </div>
          <p v-if="createError" class="form-error">{{ createError }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="createVisible = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="creating" @click="handleCreate">
            {{ creating ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑用户弹窗 -->
    <div v-if="editVisible" class="modal-overlay" @click.self="editVisible = false">
      <div class="modal">
        <div class="modal-header">
          <h3>{{ t('users.editTitle') }}</h3>
        </div>
        <div class="modal-body">
          <div class="input-group">
            <label class="input-label">{{ t('users.email') }}</label>
            <input v-model="editForm.email" class="input" disabled />
          </div>
          <div class="input-group">
            <label class="input-label">{{ t('users.nickname') }}</label>
            <input v-model="editForm.nickname" class="input" type="text" :placeholder="t('users.enterNickname')" />
          </div>
          <div class="input-group">
            <label class="input-label">{{ t('users.status') }}</label>
            <select v-model="editForm.status" class="input">
              <option value="active">{{ t('users.statusActive') }}</option>
              <option value="inactive">{{ t('users.statusInactive') }}</option>
              <option value="suspended">{{ t('users.statusBanned') }}</option>
            </select>
          </div>
          <div class="input-group">
            <label class="input-label">{{ t('users.superAdminSwitch') }}</label>
            <t-switch v-model="editForm.is_super_admin" :disabled="editForm.is_super_admin" />
          </div>
          <p v-if="editError" class="form-error">{{ editError }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="editVisible = false">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="updating" @click="handleUpdate">
            {{ updating ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { DialogPlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/stores/user'
import { showError, showSuccess } from '@/utils/request'

const { t } = useI18n()

const userStore = useUserStore()

interface User {
  id: string
  email: string
  nickname?: string
  avatar_url?: string
  status: string
  is_super_admin: boolean
  created_time: string
}

const users = ref<User[]>([])
const loading = ref(false)
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

// 创建弹窗
const createVisible = ref(false)
const creating = ref(false)
const createError = ref('')
const createForm = reactive({
  email: '',
  password: '',
  nickname: '',
  is_super_admin: false,
})

// 密码强度计算
const passwordStrength = computed(() => {
  const pwd = createForm.password
  if (!pwd) return { level: '', width: '0%', label: '' }

  let score = 0
  if (pwd.length >= 8) score++
  if (pwd.length >= 12) score++
  if (/[a-z]/.test(pwd)) score++
  if (/[A-Z]/.test(pwd)) score++
  if (/[0-9]/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++

  if (score < 3) return { level: 'weak', width: '33%', label: t('users.passwordWeak') }
  if (score < 5) return { level: 'medium', width: '66%', label: t('users.passwordMedium') }
  return { level: 'strong', width: '100%', label: t('users.passwordStrong') }
})

// 编辑弹窗
const editVisible = ref(false)
const updating = ref(false)
const editError = ref('')
const editingUserId = ref('')
const editForm = reactive({
  email: '',
  nickname: '',
  status: 'active' as string,
  is_super_admin: false,
})

const loadUsers = async () => {
  loading.value = true
  try {
    const result = await userStore.listUsers(pagination.current, pagination.pageSize)
    if (result) {
      users.value = result.items
      pagination.total = result.total
    }
  } catch (error) {
    showError(error)
  } finally {
    loading.value = false
  }
}

const onPageChange = ({ current, pageSize }: { current: number; pageSize: number }) => {
  pagination.current = current
  pagination.pageSize = pageSize
  loadUsers()
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

const getStatusBadge = (status: string): string => {
  const map: Record<string, string> = {
    active: 'badge-success',
    inactive: 'badge-warning',
    suspended: 'badge-error',
  }
  return map[status] || 'badge-default'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    active: t('users.statusActive'),
    inactive: t('users.statusInactive'),
    suspended: t('users.statusBanned'),
  }
  return map[status] || status
}

/**
 * 密码强度检查：至少8位，包含大小写字母与数字
 */
const checkPasswordStrength = (pwd: string): boolean => {
  if (pwd.length < 8) return false
  if (!/[a-z]/.test(pwd)) return false
  if (!/[A-Z]/.test(pwd)) return false
  if (!/[0-9]/.test(pwd)) return false
  return true
}

const openCreateDialog = () => {
  createForm.email = ''
  createForm.password = ''
  createForm.nickname = ''
  createForm.is_super_admin = false
  createError.value = ''
  createVisible.value = true
}

const handleCreate = async () => {
  if (!createForm.email) { createError.value = t('users.emailRequired'); return }
  if (!checkPasswordStrength(createForm.password)) { createError.value = t('users.passwordWeak'); return }
  creating.value = true
  createError.value = ''
  try {
    await userStore.createUser({
      email: createForm.email,
      password: createForm.password,
      nickname: createForm.nickname || undefined,
      is_super_admin: createForm.is_super_admin,
    })
    showSuccess(t('users.createSuccess'))
    createVisible.value = false
    loadUsers()
  } catch (error: any) {
    const msg = error?.response?.data?.msg || t('users.createFailed')
    createError.value = msg.includes('already exists') ? t('users.emailTaken') : msg
  } finally {
    creating.value = false
  }
}

const openEditDialog = (user: User) => {
  editingUserId.value = user.id
  editForm.email = user.email
  editForm.nickname = user.nickname || ''
  editForm.status = user.status
  editForm.is_super_admin = user.is_super_admin
  editError.value = ''
  editVisible.value = true
}

const handleUpdate = async () => {
  updating.value = true
  editError.value = ''
  try {
    await userStore.updateUser(editingUserId.value, {
      nickname: editForm.nickname || undefined,
      status: editForm.status,
      is_super_admin: editForm.is_super_admin,
    })
    showSuccess(t('users.updateSuccess'))
    editVisible.value = false
    loadUsers()
  } catch (error: any) {
    editError.value = error?.response?.data?.msg || t('users.updateFailed')
  } finally {
    updating.value = false
  }
}

const handleDelete = (user: User) => {
  DialogPlugin.confirm({
    header: t('users.confirmDeleteUser'),
    body: t('users.deleteConfirmMsg', { email: user.email }),
    theme: 'warning',
    onConfirm: async () => {
      try {
        await userStore.deleteUser(user.id)
        showSuccess(t('users.deleteSuccess'))
        loadUsers()
      } catch (error) {
        showError(error)
      }
    },
  })
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.users-page {
  padding: 24px;
}

.text-center {
  text-align: center;
  padding: 40px !important;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.user-name {
  font-weight: 500;
  color: var(--color-text);
}

.required {
  color: var(--color-error);
}

.form-error {
  margin-top: 8px;
  font-size: 13px;
  color: var(--color-error);
}

.password-strength {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.strength-bar {
  flex: 1;
  height: 4px;
  background: var(--color-border);
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s, background-color 0.3s;
}

.strength-fill.weak { background: #f85149; }
.strength-fill.medium { background: #e3b341; }
.strength-fill.strong { background: #3fb950; }

.strength-text {
  font-size: 12px;
  min-width: 36px;
}

.strength-text.weak { color: #f85149; }
.strength-text.medium { color: #e3b341; }
.strength-text.strong { color: #3fb950; }

.password-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-secondary);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--color-card);
  border-radius: 12px;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.modal-header {
  padding: 20px 24px 0;
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.modal-body {
  padding: 20px 24px;
}

.modal-footer {
  padding: 0 24px 20px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
