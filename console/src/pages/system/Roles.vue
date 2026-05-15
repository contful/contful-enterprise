<template>
    <!-- 页面标题 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('roles.title') }}</h1>
        <p class="page-subtitle">{{ t('roles.subtitle') }}</p>
      </div>
      <t-button theme="primary" @click="openCreateDialog">
        <template #icon><t-icon name="add" /></template>
        {{ t('roles.addRole') }}
      </t-button>
    </div>

    <!-- 角色列表 -->
    <t-table
      :data="roles"
      :columns="columns"
      :loading="loading"
      row-key="id"
      hover
      stripe
      size="medium"
    >
      <template #operations="{ row }">
        <t-space size="small">
          <t-button variant="text" theme="primary" size="small" @click="openEditDialog(row)">{{ t('common.edit') }}</t-button>
          <t-button v-if="!row.is_system" variant="text" theme="danger" size="small" @click="openDeleteDialog(row)">{{ t('common.delete') }}</t-button>
        </t-space>
      </template>
    </t-table>

    <!-- 创建/编辑角色弹窗 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="editingRole ? t('roles.editTitle') : t('roles.createTitle')"
      :width="720"
      :confirm-btn="{ content: saving ? t('common.saving') : t('common.save'), theme: 'primary' as const, loading: saving }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleSave"
      @close="resetForm"
    >
      <t-form :data="form" label-align="top" class="role-form">
        <t-form-item :label="`${t('common.name')} *`">
          <t-input
            v-model="form.name"
            :placeholder="t('roles.namePlaceholder')"
            :disabled="!!(editingRole?.is_system)"
            clearable
          />
        </t-form-item>
        <t-form-item :label="t('common.description')">
          <t-textarea
            v-model="form.description"
            :placeholder="t('roles.descPlaceholder')"
            :disabled="!!(editingRole?.is_system)"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
        </t-form-item>
        <t-form-item :label="t('roles.permissions')">
          <div class="permission-tree">
            <div v-for="(group, groupKey) in permTree" :key="groupKey" class="perm-group">
              <div class="perm-group-title">{{ permGroupLabel(groupKey) }}</div>
              <div class="perm-actions">
                <t-checkbox
                  v-for="(desc, action) in group"
                  :key="`${groupKey}:${action}`"
                  :value="`${groupKey}:${action}`"
                  v-model:checked="permChecked[`${groupKey}:${action}`]"
                  @change="onPermChange"
                >
                  {{ desc }}
                </t-checkbox>
              </div>
            </div>
          </div>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 删除确认弹窗 -->
    <t-dialog
      v-model:visible="deleteVisible"
      :header="t('common.confirmDelete')"
      theme="danger"
      :confirm-btn="{ content: deleting ? t('common.deleting') : t('common.delete'), theme: 'danger' as const, loading: deleting }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleDelete"
    >
      <p>{{ t('roles.deleteConfirm', { name: deletingRole?.name }) }}</p>
    </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, h, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import {
  listSystemRoles,
  createSystemRole,
  updateSystemRole,
  deleteSystemRole,
  getSystemPermissions,
  type SystemRole,
} from '@/api/rbac'

const { t } = useI18n()

// ─── 状态 ─────────────────────────────────────────────────────
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const roles = ref<SystemRole[]>([])
const dialogVisible = ref(false)
const deleteVisible = ref(false)
const editingRole = ref<SystemRole | null>(null)
const deletingRole = ref<SystemRole | null>(null)

const form = reactive({
  name: '',
  description: '',
})

// 权限树（平铺 key → checked）
const permChecked = reactive<Record<string, boolean>>({})
// 权限元数据
const permMeta = ref<Record<string, any>>({})

// ─── 计算属性 ──────────────────────────────────────────────────

// 权限树（从后端平铺 map 构建）
const permTree = computed(() => {
  const tree: Record<string, Record<string, string>> = {}
  const meta = permMeta.value
  if (!meta) return tree

  for (const [group, actions] of Object.entries(meta)) {
    if (typeof actions === 'object' && actions !== null) {
      tree[group] = actions as Record<string, string>
    }
  }
  return tree
})

// 当前选中的权限列表
const selectedPerms = computed(() => {
  return Object.entries(permChecked)
    .filter(([, v]) => v)
    .map(([k]) => k)
})

// ─── 表格列 ───────────────────────────────────────────────────
const columns = computed(() => [
  {
    colKey: 'name',
    title: t('common.name'),
    width: 200,
    cell: (_: unknown, { row }: { row: SystemRole }) =>
      h('div', { class: 'role-name-cell' }, [
        h('span', { class: 'role-name' }, row.name),
        row.is_system ? h('t-tag', { theme: 'primary', variant: 'light', size: 'small', style: 'margin-left:6px' }, () => t('roles.builtIn')) : null,
      ]),
  },
  {
    colKey: 'description',
    title: t('common.description'),
    ellipsis: true,
    cell: (_: unknown, { row }: { row: SystemRole }) => row.description || '-',
  },
  {
    colKey: 'permissions',
    title: t('roles.permCount'),
    width: 120,
    cell: (_: unknown, { row }: { row: SystemRole }) =>
      h('span', {}, `${row.permissions?.length ?? 0} ${t('roles.perms')}`),
  },
  {
    colKey: 'operations',
    title: t('common.actions'),
    width: 160,
  },
])

// ─── 方法 ─────────────────────────────────────────────────────

function permGroupLabel(key: string) {
  const labels: Record<string, string> = {
    users: t('roles.perm.users'),
    sites: t('roles.perm.sites'),
    tokens: t('roles.perm.tokens'),
    settings: t('roles.perm.settings'),
    audit: t('roles.perm.audit'),
    roles: t('roles.perm.roles'),
    dashboard: t('roles.perm.dashboard'),
    content_schema: t('roles.perm.contentSchema'),
    entry: t('roles.perm.entry'),
    asset: t('roles.perm.asset'),
  }
  return labels[key] || key
}

function onPermChange() {
  // checkbox 双向绑定，无需额外处理
}

function openCreateDialog() {
  editingRole.value = null
  form.name = ''
  form.description = ''
  // 清空所有权限选中
  Object.keys(permChecked).forEach((k) => (permChecked[k] = false))
  dialogVisible.value = true
}

function openEditDialog(role: SystemRole) {
  editingRole.value = role
  form.name = role.name
  form.description = role.description
  // 根据角色的 permissions 初始化 checked
  Object.keys(permChecked).forEach((k) => (permChecked[k] = false))
  role.permissions?.forEach((p) => {
    permChecked[p] = true
  })
  dialogVisible.value = true
}

function openDeleteDialog(role: SystemRole) {
  deletingRole.value = role
  deleteVisible.value = true
}

function resetForm() {
  editingRole.value = null
  form.name = ''
  form.description = ''
}

async function handleSave() {
  if (!form.name.trim()) {
    MessagePlugin.warning(t('roles.nameRequired'))
    return
  }

  saving.value = true
  try {
    if (editingRole.value) {
      await updateSystemRole(editingRole.value.id, {
        name: editingRole.value.is_system ? undefined : form.name,
        description: editingRole.value.is_system ? undefined : form.description,
        permissions: selectedPerms.value,
      })
      MessagePlugin.success(t('common.success'))
    } else {
      await createSystemRole({
        name: form.name,
        description: form.description,
        permissions: selectedPerms.value,
      })
      MessagePlugin.success(t('common.success'))
    }
    dialogVisible.value = false
    await loadRoles()
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (!deletingRole.value) return
  deleting.value = true
  try {
    await deleteSystemRole(deletingRole.value.id)
    MessagePlugin.success(t('common.success'))
    deleteVisible.value = false
    await loadRoles()
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  } finally {
    deleting.value = false
  }
}

async function loadRoles() {
  loading.value = true
  try {
    const res = await listSystemRoles()
    roles.value = res.data || []
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  } finally {
    loading.value = false
  }
}

async function loadPermMeta() {
  try {
    const res = await getSystemPermissions()
    permMeta.value = res.data || {}
    // 初始化 permChecked 键
    for (const [group, actions] of Object.entries(permMeta.value)) {
      if (typeof actions === 'object' && actions !== null) {
        for (const action of Object.keys(actions as object)) {
          permChecked[`${group}:${action}`] = false
        }
      }
    }
  } catch {
    // 静默失败，权限树为空
  }
}

onMounted(async () => {
  await Promise.all([loadRoles(), loadPermMeta()])
})
</script>

<style scoped>
.role-name-cell {
  display: flex;
  align-items: center;
  gap: 4px;
}
.role-name {
  font-weight: 500;
}
.action-cell {
  display: flex;
  gap: 4px;
}
.role-form :deep(.t-form__item) {
  margin-bottom: 16px;
}
.permission-tree {
  border: 1px solid var(--td-component-border);
  border-radius: 6px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
  width: 100%;
  box-sizing: border-box;
}

.perm-group {
  margin-bottom: 16px;
}

.perm-group:last-child {
  margin-bottom: 0;
}

.perm-group-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--td-text-color-secondary);
  margin-bottom: 10px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--td-component-border);
}

.perm-actions {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 10px;
  padding-left: 4px;
}

.perm-actions :deep(.t-checkbox) {
  margin: 0;
  min-width: 0;
}
</style>
