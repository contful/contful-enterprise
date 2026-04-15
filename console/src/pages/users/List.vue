<template>
  <div class="users-page">
    <t-card>
      <template #header>
        <div class="users-header">
          <h2>用户管理</h2>
          <t-button theme="primary" @click="showCreateDialog = true">
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

        <template #created_at="{ row }">
          {{ formatDate(row.created_at) }}
        </template>

        <template #operations="{ row }">
          <t-space>
            <t-button size="small" variant="text" @click="viewUser(row)">
              查看
            </t-button>
            <t-button size="small" variant="text" @click="editUser(row)">
              编辑
            </t-button>
            <t-button 
              size="small" 
              variant="text" 
              theme="danger"
              :disabled="row.is_super_admin"
              @click="deleteUser(row)"
            >
              删除
            </t-button>
          </t-space>
        </template>
      </t-table>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next'
import { AddIcon } from 'tdesign-icons-vue-next'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

interface User {
  id: string
  email: string
  nickname?: string
  avatar_url?: string
  status: string
  is_super_admin: boolean
  created_at: string
}

const users = ref<User[]>([])
const showCreateDialog = ref(false)
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})

const columns = [
  { colKey: 'avatar', title: '头像', width: 80 },
  { colKey: 'email', title: '邮箱', minWidth: 200 },
  { colKey: 'nickname', title: '昵称', minWidth: 120 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'is_super_admin', title: '角色', width: 120 },
  { colKey: 'created_at', title: '创建时间', width: 180 },
  { colKey: 'operations', title: '操作', width: 180 },
]

const loadUsers = async () => {
  try {
    const result = await userStore.listUsers(pagination.current, pagination.pageSize)
    if (result) {
      users.value = result.data
      pagination.total = result.total
    }
  } catch (error) {
    MessagePlugin.error('加载用户列表失败')
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

const getStatusTheme = (status: string) => {
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

const viewUser = (user: User) => {
  MessagePlugin.info(`查看用户: ${user.email}`)
}

const editUser = (user: User) => {
  MessagePlugin.info(`编辑用户: ${user.email}`)
}

const deleteUser = (user: User) => {
  const dialog = DialogPlugin.confirm({
    header: '确认删除',
    body: `确定要删除用户 ${user.email} 吗？此操作不可恢复。`,
    theme: 'warning',
    onConfirm: () => {
      MessagePlugin.success('删除成功')
      dialog.hide()
      loadUsers()
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
}

.users-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.text-gray {
  color: var(--td-text-color-secondary);
}
</style>
