<template>
  <div class="users-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('users.title') }}</h1>
        <p class="page-subtitle">{{ t('users.subtitle') }}</p>
      </div>
      <t-button theme="primary" @click="openCreateDialog">
        <template #icon><t-icon name="add" /></template>
        {{ t('users.addUser') }}
      </t-button>
    </div>

    <!-- 用户列表 — t-table -->
    <t-table
      :data="users"
      :columns="columns"
      :loading="loading"
      :pagination="{ current: pagination.current, total: pagination.total, pageSize: pagination.pageSize, showPageSize: false }"
      row-key="id"
      @page-change="onPageChange"
      hover
      stripe
      size="medium"
    />

    <!-- 创建用户弹窗 — t-dialog + t-form -->
    <t-dialog
      v-model:visible="createVisible"
      :header="t('users.createTitle')"
      :width="480"
      :confirm-btn="{ content: creating ? t('common.creating') : t('common.create'), theme: 'primary' as const, loading: creating }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleCreate"
    >
      <t-form :data="createForm" label-align="top">
        <t-form-item :label="`${t('users.email')} *`">
          <t-input v-model="createForm.email" type="email" :placeholder="t('users.enterEmail')" clearable />
          <template v-if="createError && createError.includes('@')" #help>
            <span class="form-error">{{ createError }}</span>
          </template>
        </t-form-item>
        <t-form-item :label="`${t('users.password')} *`">
          <t-input v-model="createForm.password" type="password" :placeholder="t('users.enterPassword')" clearable />
          <!-- 密码强度条 -->
          <div class="password-strength">
            <div class="strength-bar">
              <div class="strength-fill" :class="passwordStrength.level" :style="{ width: passwordStrength.width }"></div>
            </div>
            <span class="strength-text" :class="passwordStrength.level">{{ passwordStrength.label }}</span>
          </div>
          <p class="password-hint">{{ t('users.passwordHint') }}</p>
        </t-form-item>
        <t-form-item :label="t('users.nickname')">
          <t-input v-model="createForm.nickname" :placeholder="t('users.enterNickname')" clearable />
        </t-form-item>
        <t-form-item :label="t('users.superAdminSwitch')">
          <t-switch v-model="createForm.is_super_admin" />
        </t-form-item>
        <t-alert v-if="createError && !createError.includes('@')" theme="error" :message="createError" closable @close="createError = ''" />
      </t-form>
    </t-dialog>

    <!-- 编辑用户弹窗 — t-dialog + t-form -->
    <t-dialog
      v-model:visible="editVisible"
      :header="t('users.editTitle')"
      :width="480"
      :confirm-btn="{ content: updating ? t('common.saving') : t('common.save'), theme: 'primary' as const, loading: updating }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleUpdate"
    >
      <t-form :data="editForm" label-align="top">
        <t-form-item :label="t('users.email')">
          <t-input v-model="editForm.email" disabled />
        </t-form-item>
        <t-form-item :label="t('users.nickname')">
          <t-input v-model="editForm.nickname" :placeholder="t('users.enterNickname')" clearable />
        </t-form-item>
        <t-form-item :label="t('users.status')">
          <t-select v-model="editForm.status" :options="[
            { label: t('users.statusActive'), value: 'active' },
            { label: t('users.statusInactive'), value: 'inactive' },
            { label: t('users.statusBanned'), value: 'suspended' },
          ]" />
        </t-form-item>
        <t-form-item :label="t('users.superAdminSwitch')">
          <t-switch v-model="editForm.is_super_admin" :disabled="editForm.is_super_admin" />
        </t-form-item>
        <t-alert v-if="editError" theme="error" :message="editError" closable @close="editError = ''" />
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, reactive, computed, onMounted, h } from 'vue'
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

const getStatusBadge = (status: string): 'success' | 'warning' | 'danger' | 'default' => {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'default'> = {
    active: 'success',
    inactive: 'warning',
    suspended: 'danger',
  }
  return map[status] || 'default'
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

// t-table columns
const columns = computed(() => [
  {
    colKey: 'user',
    title: t('users.user'),
    cell: (_h: any, { row }: { row: User }) => h('div', { class: 'user-info' }, [
      h('div', { class: 'user-avatar' }, row.email?.charAt(0).toUpperCase()),
      h('span', { class: 'user-name' }, row.nickname || '\u2014'),
    ]),
  },
  { colKey: 'email', title: t('users.email') },
  {
    colKey: 'role',
    title: t('users.role'),
    cell: (h: any, { row }: { row: User }) => {
      const isAdmin = row.is_super_admin || (row as unknown as Record<string, any>).role === 'super_admin'
      return isAdmin
        ? h('t-tag', { props: { theme: 'warning', variant: 'light', size: 'small' } }, () => t('users.superAdmin'))
        : h('t-tag', { props: { variant: 'light', size: 'small' } }, () => t('users.normalUser'))
    },
  },
  {
    colKey: 'status',
    title: t('users.status'),
    cell: (h: any, { row }: { row: User }) => {
      const status = row.status || 'unknown'
      const theme = getStatusBadge(status)
      const map: Record<string, string> = {
        active: t('users.statusActive'),
        inactive: t('users.statusInactive'),
        suspended: t('users.statusBanned'),
      }
      return h('t-tag', { props: { theme, variant: 'light', size: 'small' } }, () => map[status] || status)
    },
  },
  { colKey: 'created_time', title: t('users.createdTime'), cell: (_h: any, { row }: { row: User }) => formatDate(row.created_time) },
  {
    colKey: 'operations',
    title: t('users.actions'),
    cell: (h: any, { row }: { row: User }) => h('div', { class: 'action-btns' }, [
      h('t-tooltip', { props: { content: t('common.edit') } }, () =>
        h('t-button', {
          props: { variant: 'outline', size: 'small', shape: 'circle' },
          on: { click: () => openEditDialog(row) },
        }, () => h('t-icon', { props: { name: 'edit' } }))
      ),
      h('t-tooltip', { props: { content: t('common.delete') } }, () =>
        h('t-button', {
          props: { theme: 'danger', variant: 'outline', size: 'small', shape: 'circle', disabled: row.is_super_admin },
          on: { click: () => handleDelete(row) },
        }, () => h('t-icon', { props: { name: 'delete' } }))
      ),
    ]),
  },
])

onMounted(() => {
  loadUsers()
})
</script>

<!-- 非 scoped：供 h() 渲染的表格单元格使用 -->
<style>
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
</style>

<style scoped>
.users-page {
  height: 100%;
  padding: 24px;
}

/* === Form error === */
.form-error {
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-error);
}

/* === Password strength === */
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

/* === Action buttons in table === */
.action-btns {
  display: flex;
  gap: 6px;
}
</style>
