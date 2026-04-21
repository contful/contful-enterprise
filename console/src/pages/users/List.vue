<template>
  <div class="users-page">
    <t-card>
      <template #header>
        <div class="users-header">
          <h2>用户管理</h2>
          <t-button theme="primary" @click="openCreateDialog">
            <template #icon>
              <AddIcon />
            </template>
            添加用户
          </t-button>
        </div>
      </template>

      <t-table
        :data="users"
        :columns="columns"
        :pagination="pagination"
        :loading="loading"
        row-key="id"
        hover
        stripe
        @page-change="onPageChange"
      >
        <template #avatar="{ row }">
          <t-avatar v-if="row.avatar_url" :src="row.avatar_url" />
          <t-avatar v-else>
            {{ row.email?.charAt(0).toUpperCase() }}
          </t-avatar>
        </template>

        <template #status="{ row }">
          <t-tag :theme="getStatusTheme(row.status)" variant="light">
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>

        <template #is_super_admin="{ row }">
          <t-tag v-if="row.is_super_admin" theme="warning" variant="light">
            超级管理员
          </t-tag>
          <span v-else class="text-gray">普通用户</span>
        </template>

        <template #created_time="{ row }">
          {{ formatDate(row.created_time) }}
        </template>

        <template #operations="{ row }">
          <t-space>
            <t-button size="small" variant="text" @click="openEditDialog(row)">
              编辑
            </t-button>
            <t-button
              size="small"
              variant="text"
              theme="danger"
              :disabled="row.is_super_admin"
              @click="handleDelete(row)"
            >
              删除
            </t-button>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 创建用户弹窗 -->
    <t-dialog
      v-model:visible="createVisible"
      header="添加用户"
      :close-on-overlay-click="false"
      :on-confirm="handleCreate"
      :on-cancel="createVisible = false"
    >
      <t-form ref="createFormRef" :data="createForm" :rules="createRules" layout="vertical">
        <t-form-item label="邮箱" name="email">
          <t-input v-model="createForm.email" placeholder="请输入邮箱" />
        </t-form-item>
        <t-form-item label="密码" name="password">
          <t-input v-model="createForm.password" type="password" placeholder="至少 8 位" />
        </t-form-item>
        <t-form-item label="昵称" name="nickname">
          <t-input v-model="createForm.nickname" placeholder="可选" />
        </t-form-item>
        <t-form-item label="超级管理员">
          <t-switch v-model="createForm.is_super_admin" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 编辑用户弹窗 -->
    <t-dialog
      v-model:visible="editVisible"
      header="编辑用户"
      :close-on-overlay-click="false"
      :on-confirm="handleUpdate"
      :on-cancel="editVisible = false"
    >
      <t-form ref="editFormRef" :data="editForm" :rules="editRules" layout="vertical">
        <t-form-item label="邮箱">
          <t-input v-model="editForm.email" disabled />
        </t-form-item>
        <t-form-item label="昵称" name="nickname">
          <t-input v-model="editForm.nickname" placeholder="可选" />
        </t-form-item>
        <t-form-item label="状态" name="status">
          <t-select v-model="editForm.status">
            <t-option value="active" label="正常" />
            <t-option value="inactive" label="停用" />
            <t-option value="suspended" label="封禁" />
          </t-select>
        </t-form-item>
        <t-form-item label="超级管理员">
          <t-switch v-model="editForm.is_super_admin" :disabled="editForm.is_super_admin" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next'
import { AddIcon } from 'tdesign-icons-vue-next'
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
const createFormRef = ref()
const createForm = reactive({
  email: '',
  password: '',
  nickname: '',
  is_super_admin: false,
})
const createRules = {
  email: [{ required: true, message: '邮箱必填', type: 'error' }],
  password: [
    { required: true, message: '密码必填', type: 'error' },
    { min: 8, message: '密码至少 8 位', type: 'warning' },
  ],
}

// 编辑弹窗
const editVisible = ref(false)
const editFormRef = ref()
const editingUserId = ref('')
const editForm = reactive({
  email: '',
  nickname: '',
  status: 'active' as string,
  is_super_admin: false,
})
const editRules = {
  status: [{ required: true }],
}

const columns = [
  { colKey: 'avatar', title: '头像', width: 80 },
  { colKey: 'email', title: '邮箱', minWidth: 200 },
  { colKey: 'nickname', title: '昵称', minWidth: 120 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'is_super_admin', title: '角色', width: 120 },
  { colKey: 'created_time', title: '创建时间', width: 180 },
  { colKey: 'operations', title: '操作', width: 120 },
]

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

const onPageChange = (pageInfo: { current: number; pageSize: number }) => {
  pagination.current = pageInfo.current
  pagination.pageSize = pageInfo.pageSize
  loadUsers()
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const getStatusTheme = (status: string): 'success' | 'warning' | 'danger' => {
  const map: Record<string, 'success' | 'warning' | 'danger'> = {
    active: 'success',
    inactive: 'warning',
    suspended: 'danger',
  }
  return map[status] || 'default'
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
  createVisible.value = true
}

const handleCreate = async () => {
  const valid = await (createFormRef.value as any).validate()
  if (!valid) return

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
    if (msg.includes('already exists')) {
      showError('该邮箱已被注册')
    } else {
      showError(msg)
    }
  }
}

const openEditDialog = (user: User) => {
  editingUserId.value = user.id
  editForm.email = user.email
  editForm.nickname = user.nickname || ''
  editForm.status = user.status
  editForm.is_super_admin = user.is_super_admin
  editVisible.value = true
}

const handleUpdate = async () => {
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
    showError(error?.response?.data?.msg || '更新失败')
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

.users-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.users-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  flex: 1;
}

.text-gray {
  color: var(--td-text-color-secondary);
}
</style>
