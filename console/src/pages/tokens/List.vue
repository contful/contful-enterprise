<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted, reactive, h, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next'
import { Dropdown as TDropdown } from 'tdesign-vue-next'
import {
  getApiTokens,
  createApiToken,
  updateApiToken,
  deleteApiToken,
  regenerateApiToken,
  revokeApiToken,
  exportApiToken,
  type ApiToken,
  type TokenStatus,
} from '@/api/api-token'
import { showError } from '@/utils/request'

function handleError(err: unknown) {
  if (err instanceof Error) {
    showError(err.message)
  } else {
    showError(String(err))
  }
}
import PageHeader from '@/components/PageHeader.vue'

const { t } = useI18n()

const loading = ref(false)
const tokens = ref<ApiToken[]>([])
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
})
// 创建/编辑/显示新 Token 弹窗
const showModal = ref(false)
const editingToken = ref<ApiToken | null>(null)
const newToken = ref('')
const submitting = ref(false)

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
  { value: 'content:read', labelKey: 'apiTokens.permissionContentRead' },
  { value: 'content:write', labelKey: 'apiTokens.permissionContentWrite' },
  { value: 'assets:read', labelKey: 'apiTokens.permissionAssetsRead' },
  { value: 'assets:write', labelKey: 'apiTokens.permissionAssetsWrite' },
]

const permissionLabels = computed(() =>
  permissionOptions.map(opt => ({ ...opt, label: t(opt.labelKey) }))
)

// 搜索和筛选
const searchKeyword = ref('')
const statusFilter = ref<TokenStatus | ''>('')
const debouncedKeyword = ref('')

// 搜索防抖：watch searchKeyword，300ms 无输入后触发实际搜索
let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch(searchKeyword, (val) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    debouncedKeyword.value = val
    pagination.current = 1
    loadTokens()
  }, 300)
})

const handleSearch = () => {
  // 立即执行（点击搜索按钮时清除防抖）
  if (debounceTimer) clearTimeout(debounceTimer)
  debouncedKeyword.value = searchKeyword.value
  pagination.current = 1
  loadTokens()
}

// 加载 Token 列表
const loadTokens = async () => {
  loading.value = true
  try {
    const res = await getApiTokens({
      page: pagination.current,
      page_size: pagination.pageSize,
      name: debouncedKeyword.value || undefined,
      status: statusFilter.value || undefined,
    })
    tokens.value = res?.items || []
    pagination.total = res?.total || 0
  } catch (error) {
    handleError(error)
  } finally {
    loading.value = false
  }
}

// 分页变化
const onPageChange = ({ current, pageSize }: { current: number; pageSize: number }) => {
  pagination.current = current
  pagination.pageSize = pageSize
  loadTokens()
}

// 打开创建弹窗
const openCreateModal = () => {
  editingToken.value = null
  formData.value = { name: '', description: '', expires_in_days: 365, permissions: [], rate_limit: 1000 }
  newToken.value = ''
  showModal.value = true
}

// 打开编辑弹窗
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const openEditModal = (token: ApiToken) => {
  const tok = token as unknown as Record<string, any>
  editingToken.value = token
  formData.value = {
    name: token.name,
    description: token.description || '',
    expires_in_days: tok.expires_in_days || 365,
    permissions: (tok.permissions as string[]) || [],
    rate_limit: tok.rate_limit || tok.rate_limits?.requests_per_day || 1000,
  }
  newToken.value = ''
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
      MessagePlugin.success(t('apiTokens.updateSuccess') || 'Token updated')
      showModal.value = false
    } else {
      const res = await createApiToken({
        name: formData.value.name,
        description: formData.value.description,
        expires_in_days: formData.value.expires_in_days,
        permissions: formData.value.permissions,
        rate_limit: formData.value.rate_limit,
      })
      newToken.value = res.token || ''
      MessagePlugin.success(t('apiTokens.createSuccess'))
    }
    await loadTokens()
  } catch (error) {
    handleError(error)
  } finally {
    submitting.value = false
  }
}

// 关闭弹窗
const closeModal = () => {
  showModal.value = false
  newToken.value = ''
}

// 格式化日期
const formatDate = (date: string | null | undefined) => {
  if (!date) return t('apiTokens.permanent')
  return new Date(date).toLocaleDateString()
}

// 判断 Token 是否过期
const isExpired = (expiresAt: string | null | undefined) => {
  if (!expiresAt) return false
  return new Date(expiresAt) < new Date()
}

// 获取状态标签
const getStatusTag = (token: ApiToken): { theme: 'success' | 'warning' | 'danger' | 'default'; label: string } => {
  const tok = token as unknown as Record<string, any>
  if (tok.revoked || token.status === 'revoked') return { theme: 'danger', label: t('apiTokens.revoked') }
  if (isExpired(token.expires_time) || token.status === 'expired') return { theme: 'warning', label: t('apiTokens.expired') }
  return { theme: 'success', label: t('apiTokens.active') || 'Active' }
}

// 删除确认 → DialogPlugin.confirm
const handleDeleteConfirm = (token: ApiToken) => {
  DialogPlugin.confirm({
    header: t('common.confirmDelete'),
    body: t('apiTokens.deleteMsg'),
    theme: 'warning',
    onConfirm: async () => {
      try {
        await deleteApiToken(token.id)
        MessagePlugin.success(t('apiTokens.deleteSuccess'))
        await loadTokens()
      } catch (error) {
        handleError(error)
      }
    },
  })
}

// 重新生成确认 → DialogPlugin.confirm
const handleRegenerateConfirm = (token: ApiToken) => {
  DialogPlugin.confirm({
    header: t('apiTokens.regenerateConfirm'),
    body: `${t('apiTokens.regenerateMsg')}<p style="color:var(--color-error);margin-top:8px">${t('apiTokens.regenerateWarning')}</p>`,
    theme: 'warning',
    onConfirm: async () => {
      try {
        const res = await regenerateApiToken(token.id)
        newToken.value = res.token || ''
        editingToken.value = null // 以"显示新token"模式打开主弹窗
        formData.value = { name: '', description: '', expires_in_days: 365, permissions: [], rate_limit: 1000 }
        showModal.value = true
        await loadTokens()
      } catch (error) {
        handleError(error)
      }
    },
  })
}

// 撤销 Token
const handleRevoke = (token: ApiToken) => {
  DialogPlugin.confirm({
    header: t('apiTokens.revoke'),
    body: t('apiTokens.revokeWarning') || `${t('apiTokens.revoke')} "${token.name}"?`,
    theme: 'warning',
    onConfirm: async () => {
      try {
        await revokeApiToken(token.id)
        MessagePlugin.success(t('apiTokens.revokeSuccess'))
        await loadTokens()
      } catch (error) {
        handleError(error)
      }
    },
  })
}

// 导出/查看详情 → DialogPlugin.confirm
const handleExportConfirm = (token: ApiToken) => {
  DialogPlugin.confirm({
    header: t('apiTokens.exportConfirm'),
    body: `${t('apiTokens.exportMsg')}<p style="color:var(--color-error);margin-top:8px">${t('apiTokens.exportWarningDetail')}</p>`,
    theme: 'info',
    confirmBtn: t('common.confirm') || 'Confirm',
    onConfirm: async () => {
      try {
        const res = await exportApiToken(token.id)
        newToken.value = res.token || ''
        editingToken.value = null
        formData.value = { name: '', description: '', expires_in_days: 365, permissions: [], rate_limit: 1000 }
        showModal.value = true
      } catch (error) {
        handleError(error)
      }
    },
  })
}

// 复制 Token（带反馈）
const copyToken = async (token: string) => {
  try {
    await navigator.clipboard.writeText(token)
    MessagePlugin.success(t('common.copied') || 'Copied to clipboard')
  } catch {
    // fallback: 旧浏览器不支持 clipboard API 时静默失败
  }
}

// t-table columns 定义
const columns = computed(() => [
  {
    colKey: 'name',
    title: t('apiTokens.tableName'),
    cell: (h: any, { row }: { row: ApiToken }) => h('div', { class: 'token-info' }, [
      h('span', { class: 'token-name' }, row.name),
      row.description ? h('span', { class: 'token-desc' }, row.description) : null,
    ].filter(Boolean)),
  },
  {
    colKey: 'permissions',
    title: t('apiTokens.tablePermissions'),
    cell: (_h2: any, { row }: { row: ApiToken }) => {
      const tok = row as unknown as Record<string, any>
      const schemas = tok.permissions?.schemas || []
      const displayItems = Array.isArray(schemas) ? schemas.slice(0, 2) : []
      const moreCount = Array.isArray(schemas) ? Math.max(0, schemas.length - 2) : 0
      return h('div', { class: 'permissions' }, [
        ...displayItems.map((perm: string) =>
          h('span', { class: 'perm-badge', key: perm }, perm)
        ),
        moreCount > 0 ? h('span', { class: 'perm-more', key: 'more' }, `+${moreCount}`) : null,
      ].filter(Boolean))
    },
  },
  {
    colKey: 'rate_limits',
    title: t('apiTokens.tableRateLimit'),
    cell: (_h: any, { row }: { row: ApiToken }) =>
      `${row.rate_limits?.requests_per_day || 0}/${t('apiTokens.rateLimitUnit')}`,
  },
  {
    colKey: 'expires_time',
    title: t('apiTokens.tableExpires'),
    cell: (_h: any, { row }: { row: ApiToken }) => formatDate(row.expires_time),
  },
  {
    colKey: 'status',
    title: t('apiTokens.tableStatus'),
    cell: (h: any, { row }: { row: ApiToken }) => {
      const st = getStatusTag(row)
      return h('t-tag', { theme: st.theme, variant: 'light', size: 'small' }, () => st.label)
    },
  },
  {
    colKey: 'last_used_time',
    title: t('apiTokens.lastUsedTime'),
    cell: (_h: any, { row }: { row: ApiToken }) => {
      const tok = row as unknown as Record<string, any>
      const time = tok.last_used_time
      if (!time) return h('span', { style: 'color: var(--color-text-placeholder)' }, '—')
      return new Date(time).toLocaleDateString()
    },
  },
  {
    colKey: 'created_time',
    title: t('common.createdAt'),
    cell: (_h: any, { row }: { row: ApiToken }) => new Date(row.created_time).toLocaleDateString(),
  },
  {
    colKey: 'operations',
    title: t('common.actions'),
    cell: (_h: any, { row }: { row: ApiToken }) => {
      const tok = row as unknown as Record<string, any>
      const active = (!tok.revoked && row.status !== 'revoked') && !isExpired(row.expires_time)
      return h('div', { class: 'action-btns' }, [
        h(TDropdown, {
          options: [
            ...(active ? [
              { content: t('common.edit'), value: 'edit', prefixIcon: () => h('t-icon', { name: 'edit' }) },
              { content: t('apiTokens.viewDetail'), value: 'view', prefixIcon: () => h('t-icon', { name: 'browse' }) },
              { content: t('apiTokens.regenerate'), value: 'regenerate', prefixIcon: () => h('t-icon', { name: 'refresh' }) },
            ] : []),
            ...((!tok.revoked && row.status !== 'revoked') ? [
              { content: t('apiTokens.revoke'), value: 'revoke', prefixIcon: () => h('t-icon', { name: 'close-circle' }) },
            ] : []),
            { content: t('common.delete'), value: 'delete', theme: 'error' as const, prefixIcon: () => h('t-icon', { name: 'delete' }) },
          ],
          onClick: (data: any) => {
            switch (data.value as string) {
              case 'edit': openEditModal(row); break
              case 'view': handleExportConfirm(row); break
              case 'regenerate': handleRegenerateConfirm(row); break
              case 'revoke': handleRevoke(row); break
              case 'delete': handleDeleteConfirm(row); break
            }
          },
          trigger: 'click',
        }, () =>
          h('t-button', { variant: 'outline', size: 'small', shape: 'circle' }, () => h('t-icon', { name: 'ellipsis' }))
        ),
      ])
    },
  },
])

onMounted(() => {
  loadTokens()
})
</script>

<template>
  <div class="page page--padded">
    <PageHeader
      :title="t('apiTokens.title')"
      :subtitle="t('apiTokens.subtitle')"
      :show-refresh="true"
      @refresh="loadTokens"
    >
      <template #primary-action>
        <t-button theme="primary" @click="openCreateModal">
          <template #icon><t-icon name="add" /></template>
          {{ t('apiTokens.createToken') }}
        </t-button>
      </template>
    </PageHeader>

    <!-- 搜索栏 -->
    <div class="tokens-toolbar">
      <t-input
        v-model="searchKeyword"
        :placeholder="t('common.searchPlaceholder')"
        clearable
        style="width: 280px"
      >
        <template #prefix-icon>
          <t-icon name="search" />
        </template>
      </t-input>
      <t-select
        v-model="statusFilter"
        :placeholder="t('apiTokens.tableStatus')"
        clearable
        style="width: 140px"
        :options="[
          { label: t('apiTokens.active'), value: 'active' },
          { label: t('apiTokens.expired'), value: 'expired' },
          { label: t('apiTokens.revoked'), value: 'revoked' },
        ]"
        @change="handleSearch"
      />
    </div>

    <!-- Token 列表 — t-table -->
    <t-table
      :data="tokens"
      :columns="(columns as any)"
      :loading="loading"
      :pagination="{ current: pagination.current, total: pagination.total, pageSize: pagination.pageSize, showPageSize: false }"
      row-key="id"
      @page-change="onPageChange"
      :empty="() => h('div', { class: 'token-empty' }, [
        h('div', { class: 'empty-icon' }, [
          h('t-icon', { name: 'lock-on', size: '48px', style: { color: 'var(--color-text-placeholder)', opacity: 0.5 } }),
        ]),
        h('p', { class: 'empty-title' }, t('apiTokens.noTokens')),
        h('p', { class: 'empty-desc' }, t('apiTokens.noTokensHint')),
        h('t-button', {
          theme: 'primary',
          onClick: openCreateModal,
        }, () => t('apiTokens.createToken')),
      ])"
      hover
      stripe
      size="medium"
    />

    <!-- 创建/编辑/显示新Token — 统一使用 t-dialog -->
    <t-dialog
      v-model:visible="showModal"
      :header="newToken ? t('apiTokens.tokenShownOnce') : (editingToken ? t('apiTokens.editTitle') : t('apiTokens.createTitle'))"
      :width="520"
      :confirm-btn="newToken ? { content: t('common.close'), theme: 'default', variant: 'outline' as const } : { content: editingToken ? t('common.save') : t('common.create'), theme: 'primary' as const, loading: submitting }"
      :cancel-btn="!!newToken"
      :on-confirm="newToken ? closeModal : handleSubmit"
      :on-close="closeModal"
    >
      <!-- 新 Token 显示区 -->
      <div v-if="newToken" class="new-token-display">
        <div class="new-token-value">
          <code>{{ newToken }}</code>
          <t-button variant="outline" size="small" @click="copyToken(newToken)">
            <template #icon><t-icon name="copy" /></template>
            {{ t('common.copy') || 'Copy' }}
          </t-button>
        </div>
        <t-alert theme="warning" :message="t('apiTokens.onlyShowOnce')" />
      </div>

      <!-- 表单区 -->
      <t-form v-else :data="formData" label-align="left" :label-width="120">
        <t-form-item :label="t('apiTokens.tableName')" required>
          <t-input
            v-model="formData.name"
            :placeholder="t('apiTokens.tokenNamePlaceholder')"
            clearable
          />
        </t-form-item>

        <t-form-item :label="t('apiTokens.description')">
          <t-input
            v-model="formData.description"
            :placeholder="t('apiTokens.descriptionPlaceholder')"
            clearable
          />
        </t-form-item>

        <t-form-item :label="t('apiTokens.expiresAt')">
          <t-select v-model="formData.expires_in_days" :options="[
            { label: t('apiTokens.expiredDays', { days: 30 }), value: 30 },
            { label: t('apiTokens.expiredDays', { days: 90 }), value: 90 },
            { label: t('apiTokens.expiredDays', { days: 180 }), value: 180 },
            { label: `1 ${t('settings.year')}`, value: 365 },
            { label: t('apiTokens.permanent'), value: 0 },
          ]" />
        </t-form-item>

        <t-form-item :label="t('apiTokens.permissions')">
          <t-checkbox-group v-model="formData.permissions" :options="permissionLabels.map(o => ({ label: o.label, value: o.value }))" />
        </t-form-item>

        <t-form-item :label="`${t('apiTokens.rateLimit')} (${t('apiTokens.rateLimitUnit')})`">
          <t-input-number
            v-model="formData.rate_limit"
            :min="100"
            :max="100000"
            :step="100"
            theme="normal"
          />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<style scoped>
/* 页面特有样式：Token 列表 */
.tokens-toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
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

/* === Permissions badge === */
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

/* === Action buttons — 已提取到 common.css === */

/* === Empty state === */
.token-empty {
  text-align: center;
  padding: 48px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.empty-icon {
  margin-bottom: 8px;
}

.empty-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text);
  margin: 0;
}

.empty-desc {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0;
}

/* === New token display === */
.new-token-display {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.new-token-value {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--color-hover);
  border-radius: 8px;
}

.new-token-value code {
  flex: 1;
  font-size: 13px;
  font-family: monospace;
  word-break: break-all;
  color: var(--color-text);
}
</style>
