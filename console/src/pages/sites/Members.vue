<template>
  <div class="page page--padded">
    <!-- 页面标题 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('members.title') }}</h1>
        <p class="page-subtitle">{{ t('members.subtitle') }}</p>
      </div>
      <t-button theme="primary" @click="openInviteDialog">
        <template #icon><t-icon name="user-add" /></template>
        {{ t('members.invite') }}
      </t-button>
    </div>

    <!-- 成员列表 -->
    <t-table
      :data="members"
      :columns="columns"
      :loading="loading"
      :pagination="{
        current: pagination.current,
        total: pagination.total,
        pageSize: pagination.pageSize,
        showPageSize: false,
      }"
      row-key="id"
      hover
      stripe
      size="medium"
      @page-change="onPageChange"
    />

    <!-- 邀请成员弹窗 -->
    <t-dialog
      v-model:visible="inviteVisible"
      :header="t('members.inviteTitle')"
      :width="480"
      :confirm-btn="{ content: inviting ? t('common.processing') : t('members.sendInvite'), theme: 'primary' as const, loading: inviting }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleInvite"
      @close="resetInviteForm"
    >
      <t-form :data="inviteForm" label-align="top">
        <t-form-item :label="`${t('users.email')} *`">
          <t-input v-model="inviteForm.email" type="email" :placeholder="t('users.enterEmail')" clearable />
        </t-form-item>
        <t-form-item :label="`${t('roles.role')} *`">
          <t-select v-model="inviteForm.role_id" :placeholder="t('members.selectRole')">
            <t-option v-for="role in siteRoles" :key="role.id" :value="role.id" :label="role.name" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 更换角色弹窗 -->
    <t-dialog
      v-model:visible="changeRoleVisible"
      :header="t('members.changeRole')"
      :width="400"
      :confirm-btn="{ content: changingRole ? t('common.saving') : t('common.save'), theme: 'primary' as const, loading: changingRole }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleChangeRole"
    >
      <t-form label-align="top">
        <t-form-item :label="t('roles.role')">
          <t-select v-model="changeRoleForm.role_id" :placeholder="t('members.selectRole')">
            <t-option v-for="role in siteRoles" :key="role.id" :value="role.id" :label="role.name" />
          </t-select>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 移除成员确认 -->
    <t-dialog
      v-model:visible="removeVisible"
      :header="t('common.confirmDelete')"
      theme="danger"
      :confirm-btn="{ content: removing ? t('common.deleting') : t('common.remove'), theme: 'danger' as const, loading: removing }"
      :cancel-btn="{ content: t('common.cancel') }"
      @confirm="handleRemove"
    >
      <p>{{ t('members.removeConfirm', { email: removingMember?.email }) }}</p>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, h, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import {
  listSiteMembers,
  addSiteMember,
  updateSiteMemberRole,
  updateSiteMemberStatus,
  removeSiteMember,
  listSiteRoles,
  type SiteMember,
  type SiteRole,
} from '@/api/rbac'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const route = useRoute()
const userStore = useUserStore()

const siteId = computed(() => route.params.siteId as string)

// ─── 状态 ─────────────────────────────────────────────────────
const loading = ref(false)
const inviting = ref(false)
const changingRole = ref(false)
const removing = ref(false)
const members = ref<SiteMember[]>([])
const siteRoles = ref<SiteRole[]>([])

const inviteVisible = ref(false)
const changeRoleVisible = ref(false)
const removeVisible = ref(false)

const changingMember = ref<SiteMember | null>(null)
const removingMember = ref<SiteMember | null>(null)

const pagination = reactive({ current: 1, total: 0, pageSize: 20 })

const inviteForm = reactive({ email: '', role_id: '' })
const changeRoleForm = reactive({ role_id: '' })

// ─── 表格列 ───────────────────────────────────────────────────
const columns = computed(() => [
  {
    colKey: 'email',
    title: t('users.email'),
    ellipsis: true,
    cell: (_: unknown, { row }: { row: SiteMember }) =>
      h('div', { class: 'member-cell' }, [
        row.avatar_url
          ? h('t-avatar', { image: row.avatar_url, size: '28px', style: 'margin-right:8px' })
          : h('t-avatar', { size: '28px', style: 'margin-right:8px' }, () => row.nickname?.[0]?.toUpperCase() || row.email[0].toUpperCase()),
        h('div', {}, [
          h('div', { class: 'member-email' }, row.email),
          row.nickname ? h('div', { class: 'member-nickname' }, row.nickname) : null,
        ]),
      ]),
  },
  {
    colKey: 'role_name',
    title: t('roles.role'),
    width: 160,
    cell: (_: unknown, { row }: { row: SiteMember }) =>
      h('t-tag', { theme: 'default', variant: 'light' }, () => row.role_name),
  },
  {
    colKey: 'status',
    title: t('common.status'),
    width: 100,
    cell: (_: unknown, { row }: { row: SiteMember }) =>
      h('t-tag', { theme: row.status === 'active' ? 'success' : 'default', variant: 'light' }, () =>
        row.status === 'active' ? t('common.active') : t('common.inactive'),
      ),
  },
  {
    colKey: 'joined_at',
    title: t('common.createdAt'),
    width: 160,
    cell: (_: unknown, { row }: { row: SiteMember }) =>
      new Date(row.joined_at).toLocaleString(),
  },
  {
    colKey: 'actions',
    title: t('common.actions'),
    width: 200,
    fixed: 'right',
    cell: (_: unknown, { row }: { row: SiteMember }) => {
      const isCurrentUser = row.user_id === userStore.currentUser?.id
      return h('div', { class: 'action-cell' }, [
        h(
          't-button',
          { variant: 'text', theme: 'primary', size: 'small', onClick: () => openChangeRoleDialog(row) },
          () => t('members.changeRole'),
        ),
        h(
          't-button',
          {
            variant: 'text',
            theme: row.status === 'active' ? 'default' : 'primary',
            size: 'small',
            onClick: () => handleToggleStatus(row),
          },
          () => row.status === 'active' ? t('common.inactive') : t('common.active'),
        ),
        !isCurrentUser
          ? h(
              't-button',
              { variant: 'text', theme: 'danger', size: 'small', onClick: () => openRemoveDialog(row) },
              () => t('common.remove'),
            )
          : null,
      ])
    },
  },
])

// ─── 方法 ─────────────────────────────────────────────────────

async function loadMembers() {
  loading.value = true
  try {
    const res = await listSiteMembers(siteId.value, {
      page: pagination.current,
      page_size: pagination.pageSize,
    })
    members.value = res.data?.items || []
    pagination.total = res.data?.total || 0
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  } finally {
    loading.value = false
  }
}

async function loadSiteRoles() {
  try {
    const res = await listSiteRoles(siteId.value)
    siteRoles.value = res.data || []
  } catch {
    // 静默失败
  }
}

function onPageChange({ current }: { current: number }) {
  pagination.current = current
  loadMembers()
}

function openInviteDialog() {
  inviteForm.email = ''
  inviteForm.role_id = ''
  inviteVisible.value = true
}

function resetInviteForm() {
  inviteForm.email = ''
  inviteForm.role_id = ''
}

function openChangeRoleDialog(member: SiteMember) {
  changingMember.value = member
  changeRoleForm.role_id = member.role_id
  changeRoleVisible.value = true
}

function openRemoveDialog(member: SiteMember) {
  removingMember.value = member
  removeVisible.value = true
}

async function handleInvite() {
  if (!inviteForm.email || !inviteForm.role_id) {
    MessagePlugin.warning(t('members.fillRequired'))
    return
  }

  inviting.value = true
  try {
    await addSiteMember(siteId.value, {
      email: inviteForm.email,
      role_id: inviteForm.role_id,
    })
    MessagePlugin.success(t('common.success'))
    inviteVisible.value = false
    await loadMembers()
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  } finally {
    inviting.value = false
  }
}

async function handleChangeRole() {
  if (!changingMember.value || !changeRoleForm.role_id) return
  changingRole.value = true
  try {
    await updateSiteMemberRole(siteId.value, changingMember.value.user_id, changeRoleForm.role_id)
    MessagePlugin.success(t('common.success'))
    changeRoleVisible.value = false
    await loadMembers()
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  } finally {
    changingRole.value = false
  }
}

async function handleToggleStatus(member: SiteMember) {
  const newStatus = member.status === 'active' ? 'inactive' : 'active'
  try {
    await updateSiteMemberStatus(siteId.value, member.user_id, newStatus)
    MessagePlugin.success(t('common.success'))
    await loadMembers()
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  }
}

async function handleRemove() {
  if (!removingMember.value) return
  removing.value = true
  try {
    await removeSiteMember(siteId.value, removingMember.value.user_id)
    MessagePlugin.success(t('common.success'))
    removeVisible.value = false
    await loadMembers()
  } catch (e: any) {
    MessagePlugin.error(e?.data?.msg || t('common.error'))
  } finally {
    removing.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadMembers(), loadSiteRoles()])
})
</script>

<style scoped>
.member-cell {
  display: flex;
  align-items: center;
}
.member-email {
  font-size: 13px;
  color: var(--td-text-color-primary);
}
.member-nickname {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
.action-cell {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
</style>
