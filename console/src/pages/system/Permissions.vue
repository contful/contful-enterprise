<template>
  <PageHeader :title="t('permissions.title')" :subtitle="t('permissions.subtitle')" />

  <div class="toolbar">
    <span></span>
    <t-space>
      <t-button variant="outline" :loading="clearingCache" @click="clearCache">
        {{ t('permissions.clearCache') }}
      </t-button>
      <t-button variant="outline" @click="loadData">
        <template #icon><t-icon name="refresh" /></template>
        {{ t('common.refresh') }}
      </t-button>
      <t-button theme="primary" @click="openGroupDialog()">
        <template #icon><t-icon name="add" /></template>
        {{ t('permissions.addGroup') }}
      </t-button>
    </t-space>
  </div>

  <t-loading :loading="loading">
    <t-collapse v-if="groups.length > 0" expand-icon-placement="left">
      <t-collapse-panel v-for="group in groups" :key="group.id" :value="group.id">
        <template #header>
          <div class="group-header">
            <span class="group-label">{{ group.label }}</span>
            <t-tag variant="outline" size="small">{{ group.group_key }}</t-tag>
            <span class="group-perm-count">{{ group.permissions?.length ?? 0 }} {{ t('permissions.perms') }}</span>
          </div>
        </template>
        <template #headerRightContent>
          <t-space size="small" @click.stop>
            <t-button variant="text" size="small" @click="openPermDialog(group)">{{ t('permissions.addPerm') }}</t-button>
            <t-button variant="text" size="small" @click="openGroupDialog(group)">{{ t('common.edit') }}</t-button>
            <t-popconfirm :content="t('permissions.deleteGroupConfirm', { name: group.label })" @confirm="handleDeleteGroup(group.id)">
              <t-button variant="text" theme="danger" size="small">{{ t('common.delete') }}</t-button>
            </t-popconfirm>
          </t-space>
        </template>
        <t-table :data="group.permissions || []" :columns="permColumns" row-key="id" size="small" hover>
          <template #operations="{ row }">
            <t-space size="small">
              <t-button variant="text" size="small" @click="openPermDialog(group, row)">{{ t('common.edit') }}</t-button>
              <t-popconfirm :content="t('permissions.deletePermConfirm', { action: row.action })" @confirm="handleDeletePerm(row.id)">
                <t-button variant="text" theme="danger" size="small">{{ t('common.delete') }}</t-button>
              </t-popconfirm>
            </t-space>
          </template>
        </t-table>
      </t-collapse-panel>
    </t-collapse>
    <t-empty v-else :description="t('permissions.noData')" />
  </t-loading>

  <!-- 分组弹窗 -->
  <t-dialog
    v-model:visible="groupDialogVisible"
    :header="editingGroupId ? t('permissions.editGroup') : t('permissions.addGroup')"
    :width="480"
    :confirm-btn="{ content: saving ? t('common.saving') : t('common.save'), theme: 'primary' as const, loading: saving }"
    @confirm="handleSaveGroup"
  >
    <t-form :data="groupForm" label-align="top">
      <t-form-item :label="t('permissions.groupKey')" required>
        <t-input v-model="groupForm.group_key" :disabled="!!editingGroupId" :placeholder="t('permissions.groupKeyHint')" />
      </t-form-item>
      <t-form-item :label="t('permissions.groupLabel')" required>
        <t-input v-model="groupForm.label" :placeholder="t('permissions.groupLabelHint')" />
      </t-form-item>
      <t-form-item :label="t('permissions.groupLabelEn')">
        <t-input v-model="groupForm.label_en" placeholder="English label" />
      </t-form-item>
      <t-form-item :label="t('permissions.sortOrder')">
        <t-input-number v-model="groupForm.sort_order" :min="0" style="width:100%" />
      </t-form-item>
    </t-form>
  </t-dialog>

  <!-- 权限项弹窗 -->
  <t-dialog
    v-model:visible="permDialogVisible"
    :header="editingPermId ? t('permissions.editPerm') : t('permissions.addPerm')"
    :width="480"
    :confirm-btn="{ content: saving ? t('common.saving') : t('common.save'), theme: 'primary' as const, loading: saving }"
    @confirm="handleSavePerm"
  >
    <t-form :data="permForm" label-align="top">
      <t-form-item :label="`${t('permissions.permAction')} *`">
        <t-input v-model="permForm.action" :disabled="!!editingPermId" :placeholder="t('permissions.permActionHint')" />
      </t-form-item>
      <t-form-item :label="t('permissions.permLabel')" required>
        <t-input v-model="permForm.label" :placeholder="t('permissions.permLabelHint')" />
      </t-form-item>
      <t-form-item :label="t('permissions.permLabelEn')">
        <t-input v-model="permForm.label_en" placeholder="English label" />
      </t-form-item>
      <t-form-item :label="t('permissions.sortOrder')">
        <t-input-number v-model="permForm.sort_order" :min="0" style="width:100%" />
      </t-form-item>
    </t-form>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import PageHeader from '@/components/PageHeader.vue'
import {
  listPermissions,
  createPermissionGroup,
  updatePermissionGroup,
  deletePermissionGroup,
  createPermission,
  updatePermission,
  deletePermission,
  clearPermissionCache,
  type PermissionGroup,
} from '@/api/rbac'

const { t } = useI18n()

const loading = ref(false)
const saving = ref(false)
const clearingCache = ref(false)
const groups = ref<PermissionGroup[]>([])

const groupDialogVisible = ref(false)
const editingGroupId = ref('')
const groupForm = reactive({ group_key: '', label: '', label_en: '', sort_order: 0 })

const permDialogVisible = ref(false)
const editingPermId = ref('')
const editingPermGroupId = ref('')
const permForm = reactive({ action: '', label: '', label_en: '', sort_order: 0 })

const permColumns = computed(() => [
  { colKey: 'action', title: t('permissions.permAction'), width: 120 },
  { colKey: 'label', title: t('permissions.permLabel'), width: 180 },
  { colKey: 'sort_order', title: t('permissions.sortOrder'), width: 80 },
  { colKey: 'operations', title: t('common.actions'), width: 140 },
])

async function loadData() {
  loading.value = true
  try {
    const res = await listPermissions()
    groups.value = res.data || []
  } catch (e: any) {
    MessagePlugin.error(e?.msg || t('common.error'))
  } finally {
    loading.value = false
  }
}

// ── 分组操作 ──

function openGroupDialog(group?: PermissionGroup) {
  if (group) {
    editingGroupId.value = group.id
    groupForm.group_key = group.group_key
    groupForm.label = group.label
    groupForm.label_en = group.label_en || ''
    groupForm.sort_order = group.sort_order
  } else {
    editingGroupId.value = ''
    groupForm.group_key = ''
    groupForm.label = ''
    groupForm.label_en = ''
    groupForm.sort_order = 0
  }
  groupDialogVisible.value = true
}

async function handleSaveGroup() {
  if (!groupForm.group_key || !groupForm.label) {
    MessagePlugin.warning(t('permissions.fillRequired'))
    return
  }
  saving.value = true
  try {
    if (editingGroupId.value) {
      await updatePermissionGroup(editingGroupId.value, {
        label: groupForm.label,
        label_en: groupForm.label_en,
        sort_order: groupForm.sort_order,
      })
    } else {
      await createPermissionGroup({
        group_key: groupForm.group_key,
        label: groupForm.label,
        label_en: groupForm.label_en,
        sort_order: groupForm.sort_order,
      })
    }
    MessagePlugin.success(t('common.success'))
    groupDialogVisible.value = false
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.msg || t('common.error'))
  } finally {
    saving.value = false
  }
}

async function handleDeleteGroup(id: string) {
  try {
    await deletePermissionGroup(id)
    MessagePlugin.success(t('common.success'))
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.msg || t('common.error'))
  }
}

// ── 权限项操作 ──

function openPermDialog(group: PermissionGroup, perm?: PermissionItem) {
  editingPermGroupId.value = group.id
  if (perm) {
    editingPermId.value = perm.id
    permForm.action = perm.action
    permForm.label = perm.label
    permForm.label_en = perm.label_en || ''
    permForm.sort_order = perm.sort_order
  } else {
    editingPermId.value = ''
    permForm.action = ''
    permForm.label = ''
    permForm.label_en = ''
    permForm.sort_order = 0
  }
  permDialogVisible.value = true
}

async function handleSavePerm() {
  if (!permForm.action || !permForm.label) {
    MessagePlugin.warning(t('permissions.fillRequired'))
    return
  }
  saving.value = true
  try {
    if (editingPermId.value) {
      await updatePermission(editingPermId.value, {
        label: permForm.label,
        label_en: permForm.label_en,
        sort_order: permForm.sort_order,
      })
    } else {
      await createPermission({
        group_id: editingPermGroupId.value,
        action: permForm.action,
        label: permForm.label,
        label_en: permForm.label_en,
        sort_order: permForm.sort_order,
      })
    }
    MessagePlugin.success(t('common.success'))
    permDialogVisible.value = false
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.msg || t('common.error'))
  } finally {
    saving.value = false
  }
}

async function handleDeletePerm(id: string) {
  try {
    await deletePermission(id)
    MessagePlugin.success(t('common.success'))
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.msg || t('common.error'))
  }
}

async function clearCache() {
  clearingCache.value = true
  try {
    await clearPermissionCache()
    MessagePlugin.success(t('common.success'))
    loadData()
  } catch (e: any) {
    MessagePlugin.error(e?.msg || t('common.error'))
  } finally {
    clearingCache.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.group-header {
  display: flex;
  align-items: center;
  gap: 10px;
}
.group-label {
  font-weight: 600;
  font-size: 14px;
}
.group-perm-count {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}
</style>
