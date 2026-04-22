<template>
  <div class="users-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">用户管理</h1>
        <p class="page-subtitle">管理系统用户与权限</p>
      </div>
      <button class="btn btn-primary" @click="openCreateDialog">
        <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
          <path d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"/>
        </svg>
        添加用户
      </button>
    </div>

    <!-- 用户列表 -->
    <div class="card" style="padding: 0; overflow: hidden;">
      <table class="table">
        <thead>
          <tr>
            <th>用户</th>
            <th>邮箱</th>
            <th>角色</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="6" class="empty-state">加载中...</td>
          </tr>
          <tr v-else-if="users.length === 0">
            <td colspan="6" class="empty-state">
              <svg width="48" height="48" viewBox="0 0 20 20" fill="currentColor" opacity="0.3">
                <path d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"/>
              </svg>
              <h3>暂无用户</h3>
              <p>点击「添加用户」创建第一个用户</p>
              <button class="btn btn-primary btn-sm" @click="openCreateDialog">添加用户</button>
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
              <span v-if="row.is_super_admin" class="badge badge-warning">超级管理员</span>
              <span v-else class="badge badge-default">普通用户</span>
            </td>
            <td>
              <span :class="['badge', getStatusBadge(row.status)]">{{ getStatusText(row.status) }}</span>
            </td>
            <td>{{ formatDate(row.created_time) }}</td>
            <td>
              <div style="display:flex;gap:8px;">
                <button class="btn btn-secondary btn-sm" @click="openEditDialog(row)">编辑</button>
                <button
                  class="btn btn-sm"
                  :class="row.is_super_admin ? 'btn-secondary' : 'btn-danger'"
                  :disabled="row.is_super_admin"
                  @click="handleDelete(row)"
                >删除</button>
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
          <h3>添加用户</h3>
        </div>
        <div class="modal-body">
          <div class="input-group">
            <label class="input-label">邮箱 <span class="required">*</span></label>
            <input v-model="createForm.email" class="input" type="email" placeholder="请输入邮箱" />
          </div>
          <div class="input-group">
            <label class="input-label">密码 <span class="required">*</span></label>
            <input v-model="createForm.password" class="input" type="password" placeholder="至少 8 位" />
          </div>
          <div class="input-group">
            <label class="input-label">昵称</label>
            <input v-model="createForm.nickname" class="input" type="text" placeholder="可选" />
          </div>
          <div class="input-group">
            <label class="input-label">超级管理员</label>
            <t-switch v-model="createForm.is_super_admin" />
          </div>
          <p v-if="createError" class="form-error">{{ createError }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="createVisible = false">取消</button>
          <button class="btn btn-primary" :disabled="creating" @click="handleCreate">
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑用户弹窗 -->
    <div v-if="editVisible" class="modal-overlay" @click.self="editVisible = false">
      <div class="modal">
        <div class="modal-header">
          <h3>编辑用户</h3>
        </div>
        <div class="modal-body">
          <div class="input-group">
            <label class="input-label">邮箱</label>
            <input v-model="editForm.email" class="input" disabled />
          </div>
          <div class="input-group">
            <label class="input-label">昵称</label>
            <input v-model="editForm.nickname" class="input" type="text" placeholder="可选" />
          </div>
          <div class="input-group">
            <label class="input-label">状态</label>
            <select v-model="editForm.status" class="input">
              <option value="active">正常</option>
              <option value="inactive">停用</option>
              <option value="suspended">封禁</option>
            </select>
          </div>
          <div class="input-group">
            <label class="input-label">超级管理员</label>
            <t-switch v-model="editForm.is_super_admin" :disabled="editForm.is_super_admin" />
          </div>
          <p v-if="editError" class="form-error">{{ editError }}</p>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="editVisible = false">取消</button>
          <button class="btn btn-primary" :disabled="updating" @click="handleUpdate">
            {{ updating ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { DialogPlugin } from 'tdesign-vue-next'
import { useUserStore } from '@/stores/user'
import { showError, showSuccess } from '@/utils/request'

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
  return new Date(dateStr).toLocaleString('zh-CN')
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
    active: '正常',
    inactive: '停用',
    suspended: '封禁',
  }
  return map[status] || status
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
  if (!createForm.email) { createError.value = '邮箱必填'; return }
  if (createForm.password.length < 8) { createError.value = '密码至少 8 位'; return }
  creating.value = true
  createError.value = ''
  try {
    await userStore.createUser({
      email: createForm.email,
      password: createForm.password,
      nickname: createForm.nickname || undefined,
      is_super_admin: createForm.is_super_admin,
    })
    showSuccess('用户创建成功')
    createVisible.value = false
    loadUsers()
  } catch (error: any) {
    const msg = error?.response?.data?.msg || '创建失败'
    createError.value = msg.includes('already exists') ? '该邮箱已被注册' : msg
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
    showSuccess('更新成功')
    editVisible.value = false
    loadUsers()
  } catch (error: any) {
    editError.value = error?.response?.data?.msg || '更新失败'
  } finally {
    updating.value = false
  }
}

const handleDelete = (user: User) => {
  DialogPlugin.confirm({
    header: '确认删除',
    body: `确定要删除用户 ${user.email} 吗？此操作不可恢复。`,
    theme: 'warning',
    onConfirm: async () => {
      try {
        await userStore.deleteUser(user.id)
        showSuccess('删除成功')
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
