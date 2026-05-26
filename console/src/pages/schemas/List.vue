<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, computed, onMounted, h } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next'
import {
  getContentSchemas,
  createContentSchema,
  updateContentSchema,
  deleteContentSchema,
  signSchema,
  verifySchema,
  type ContentSchema,
  type ContentSchemaCreate,
  type ContentSchemaUpdate,
} from '@/api/schema'
import { showError, showSuccess, getFriendlyError } from '@/utils/request'

function handleError(err: unknown) {
  if (err instanceof Error) {
    showError(err.message)
  } else {
    showError(String(err))
  }
}
import PageHeader from '@/components/PageHeader.vue'

const { t } = useI18n()
const router = useRouter()

// 状态
const loading = ref(false)
const dataList = ref<ContentSchema[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 创建/编辑对话框
const dialogVisible = ref(false)
const dialogTitle = computed(() => isEditing.value ? t('contentSchemas.formTitleEdit') : t('contentSchemas.formTitleCreate'))
const isEditing = ref(false)
const editingId = ref('')
const signingId = ref('')
const verifyingId = ref('')
const submitting = ref(false)

const formData = ref<ContentSchemaCreate>({
  name: '',
  slug: '',
  description: '',
  kind: 'collection',
  versioning_enabled: false,
})

// TDesign 表单验证规则
const formRules = computed(() => ({
  name: [{ required: true, message: t('contentSchemas.namePlaceholder'), trigger: 'blur' as const }],
  slug: [{ required: true, message: t('contentSchemas.slugPlaceholder'), trigger: 'blur' as const }],
}))

// 自动生成 slug
const generateSlug = () => {
  if (formData.value.name && !isEditing.value) {
    formData.value.slug = formData.value.name
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
  }
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const res = await getContentSchemas({ page: currentPage.value, page_size: pageSize.value })
    dataList.value = res.data?.items || []
    total.value = res.data?.total || 0
  } catch (e) {
    handleError(e)
  } finally {
    loading.value = false
  }
}

// 分页变化
const onPageChange = (pageInfo: { current: number; pageSize: number }) => {
  currentPage.value = pageInfo.current
  pageSize.value = pageInfo.pageSize
  loadData()
}

// 打开创建对话框
const openCreateDialog = () => {
  isEditing.value = false
  editingId.value = ''
  formData.value = {
    name: '',
    slug: '',
    description: '',
    kind: 'collection',
    versioning_enabled: false,
  }
  dialogVisible.value = true
}

// 打开编辑对话框
const openEditDialog = (row: ContentSchema) => {
  isEditing.value = true
  editingId.value = row.id
  formData.value = {
    name: row.name,
    slug: row.slug,
    description: row.description || '',
    kind: row.kind,
    versioning_enabled: row.versioning_enabled,
  }
  dialogVisible.value = true
}

// 提交表单（由 t-dialog confirm 事件触发）
const submitForm = async () => {
  // 前端校验
  if (!formData.value.name?.trim()) {
    MessagePlugin.warning(t('contentSchemas.namePlaceholder'))
    return
  }
  if (!formData.value.slug?.trim()) {
    MessagePlugin.warning(t('contentSchemas.slugPlaceholder'))
    return
  }
  submitting.value = true

  try {
    if (isEditing.value) {
      await updateContentSchema(editingId.value, formData.value as ContentSchemaUpdate)
      showSuccess(t('contentSchemas.updateSuccess'))
    } else {
      await createContentSchema(formData.value)
      showSuccess(t('contentSchemas.createSuccess'))
    }
    dialogVisible.value = false
    loadData()
  } catch (e: any) {
    MessagePlugin.error(getFriendlyError(e))
  } finally {
    submitting.value = false
  }
}

// 删除内容类型 — 显示具体名称
const handleDelete = (row: ContentSchema) => {
  const dialog = DialogPlugin.confirm({
    header: t('contentSchemas.deleteConfirm'),
    body: `${t('contentSchemas.deleteConfirmMsg')}「${row.name}」？`,
    theme: 'warning',
    confirmBtn: { content: t('common.confirm'), theme: 'danger' },
    cancelBtn: t('common.cancel'),
    onConfirm: async () => {
      try {
        await deleteContentSchema(row.id)
        MessagePlugin.success(t('contentSchemas.deleteSuccess'))
        loadData()
        dialog.destroy()
      } catch (e) {
        handleError(e)
      }
    },
  })
}

const handleSign = async (row: ContentSchema) => {
  signingId.value = row.id
  try {
    await signSchema(row.id)
    MessagePlugin.success(t('contentSchemas.signSuccess'))
  } catch (e) {
    handleError(e)
  } finally {
    signingId.value = ''
  }
}

const handleVerify = async (row: ContentSchema) => {
  verifyingId.value = row.id
  try {
    const res: any = await verifySchema(row.id)
    const result = res.data
    DialogPlugin.alert({
      header: t('contentSchemas.verifyResult'),
      body: h('div', { style: 'line-height:2' }, [
        h('p', [
          h('strong', t('contentSchemas.verifyStatus') + '：'),
          h('span', { style: `color:${result.valid ? 'var(--td-success-color)' : 'var(--td-error-color)'};font-weight:bold` },
            result.valid ? t('contentSchemas.verifyPass') : t('contentSchemas.verifyFail')),
        ]),
        h('p', [h('strong', t('contentSchemas.verifyAlgorithm') + '：'), result.algorithm || '-']),
        h('p', [h('strong', t('contentSchemas.verifySignature') + '：'), h('code', { style: 'font-size:12px;word-break:break-all' }, result.signature || '-')]),
        result.reason ? h('p', { style: 'color:var(--td-error-color)' }, result.reason) : null,
      ]),
      confirmBtn: t('common.close'),
    })
  } catch (e) {
    handleError(e)
  } finally {
    verifyingId.value = ''
  }
}

// 跳转到字段管理
const goToFields = (row: ContentSchema) => {
  router.push(`/content/schemas/${row.id}/fields`)
}

// 格式化时间
const formatDate = (date: string) => {
  if (!date) return '-'
  const d = new Date(date)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 格式化 kind
const formatKind = (kind: string) => {
  return kind === 'collection' ? t('contentSchemas.kindCollection') : t('contentSchemas.kindSingle')
}

// 格式化状态
const formatStatus = (isActive: boolean) => {
  return isActive ? t('common.enabled') : t('common.disabled')
}

// t-table 列定义
const columns = computed(() => [
  {
    colKey: 'name',
    title: t('contentSchemas.tableName'),
    cell: (h2: any, { row }: { row: ContentSchema }) => h2('span', { class: 'name-text' }, row.name),
  },
  {
    colKey: 'slug',
    title: t('contentSchemas.slug'),
    width: 140,
    cell: (h2: any, { row }: { row: ContentSchema }) => h2('code', { class: 'slug-code' }, row.slug),
  },
  { colKey: 'kind', title: t('contentSchemas.tableType'), width: 120 },
  { colKey: 'is_active', title: t('contentSchemas.tableStatus'), width: 100, cell: (_h: any, { row }: { row: ContentSchema }) => formatStatus(row.is_active) },
  {
    colKey: 'versioning_enabled',
    title: t('contentSchemas.tableVersioning'),
    width: 100,
    cell: (_h: any, { row }: { row: ContentSchema }) => (row.versioning_enabled ? '✓' : '✗'),
  },
  {
    colKey: 'description',
    title: t('common.description'),
    cell: (h2: any, { row }: { row: ContentSchema }) =>
      h2('span', { class: 'description' }, row.description || '-'),
  },
  { colKey: 'updated_time', title: t('common.updatedAt'), width: 180, cell: (_h: any, { row }: { row: ContentSchema }) => formatDate(row.updated_time) },
  { colKey: 'operations', title: t('common.actions'), width: 400 },
])

onMounted(() => {
  loadData()
})
</script>

<template>
    <PageHeader
      :title="t('contentSchemas.title')"
      :subtitle="t('contentSchemas.subtitle')"
      :show-refresh="true"
      @refresh="loadData"
    >
      <template #primary-action>
        <t-button theme="primary" @click="openCreateDialog">
          <template #icon><t-icon name="add" /></template>
          {{ t('contentSchemas.createTypeBtn') }}
        </t-button>
      </template>
    </PageHeader>

    <!-- 空状态 — t-empty 替换整张表 -->
    <div v-if="dataList.length === 0 && !loading" class="schema-empty-wrapper">
      <t-empty :description="t('contentSchemas.noTypes')">
        <template #action>
          <t-button theme="primary" @click="openCreateDialog">
            <template #icon><t-icon name="add" /></template>
            {{ t('contentSchemas.createTypeBtn') }}
          </t-button>
        </template>
      </t-empty>
    </div>

    <!-- 数据表格 — t-table（有数据时显示） -->
    <t-table
      v-else
      :data="dataList"
      :columns="columns"
      :loading="loading"
      :pagination="{ current: currentPage, total: total, pageSize: pageSize, showJumper: false, showPageSize: false }"
      row-key="id"
      @page-change="onPageChange"
      hover
      stripe
      size="medium"
    >
      <!-- kind 列：自定义渲染 -->
      <template #kind-cell="{ row }">
        <t-tag :theme="row.kind === 'collection' ? 'primary' as const : 'warning' as const" variant="light" size="small">
          {{ formatKind(row.kind) }}
        </t-tag>
      </template>
      <template #operations="{ row }">
        <t-space size="small">
          <t-button variant="outline" size="small" @click="goToFields(row)">{{ t('contentSchemas.manageFields') }}</t-button>
          <t-button variant="outline" size="small" @click="openEditDialog(row)">{{ t('common.edit') }}</t-button>
          <t-button variant="outline" size="small" theme="warning" :loading="signingId === row.id" @click="handleSign(row)">{{ t('contentSchemas.sign') }}</t-button>
          <t-button variant="outline" size="small" theme="success" :loading="verifyingId === row.id" @click="handleVerify(row)">{{ t('contentSchemas.verify') }}</t-button>
          <t-button theme="danger" variant="outline" size="small" @click="handleDelete(row)">{{ t('common.delete') }}</t-button>
        </t-space>
      </template>
    </t-table>

    <!-- 创建/编辑对话框 — t-dialog（带动画、ESC 关闭、遮罩关闭） -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      :width="520"
      :confirm-btn="{ content: isEditing ? t('common.save') : t('common.create'), theme: 'primary' as const, loading: submitting }"
      :cancel-btn="{ content: t('common.cancel') }"
      :close-on-overlay-click="true"
      :close-on-esc-keydown="true"
      @confirm="submitForm"
    >
      <t-form :data="formData" :rules="formRules" label-align="left" :label-width="100">
        <t-form-item :label="t('common.name')" name="name">
          <t-input
            v-model="formData.name"
            :placeholder="t('contentSchemas.namePlaceholder')"
            clearable
            @blur="generateSlug"
          />
        </t-form-item>

        <t-form-item :label="t('contentSchemas.slug')" name="slug">
          <t-input
            v-model="formData.slug"
            :placeholder="t('contentSchemas.slugPlaceholder')"
            :disabled="isEditing"
          />
          <template #tips>
            <span class="input-hint">{{ t('contentSchemas.slugHint') }}</span>
          </template>
        </t-form-item>

        <t-form-item :label="t('contentSchemas.kind')">
          <t-select v-model="formData.kind" :disabled="isEditing">
            <t-option value="collection" :label="t('contentSchemas.kindCollection')" />
            <t-option value="single" :label="t('contentSchemas.kindSingle')" />
          </t-select>
          <template #tips>
            <span class="input-hint">
              <strong>{{ t('contentSchemas.kindCollection') }}：</strong>{{ t('contentSchemas.kindCollectionHint') }}<br/>
              <strong>{{ t('contentSchemas.kindSingle') }}：</strong>{{ t('contentSchemas.kindSingleHint') }}
            </span>
          </template>
        </t-form-item>

        <t-form-item :label="t('common.description')">
          <t-textarea
            v-model="formData.description"
            :placeholder="t('common.description')"
            :autosize="{ minRows: 3, maxRows: 6 }"
          />
        </t-form-item>

        <t-form-item :label="t('contentSchemas.versioning')">
          <t-switch v-model="formData.versioning_enabled" />
        </t-form-item>
      </t-form>
    </t-dialog>
</template>

<style scoped>
/* 页面特有样式：内容类型列表 — page-header/header-actions 已提取到 common.css */

.name-text {
  font-weight: 500;
}

.slug-code {
  font-size: 12px;
  color: var(--color-text-secondary);
  background: var(--color-hover);
  padding: 2px 6px;
  border-radius: 4px;
}

.description {
  color: var(--color-text-secondary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* === Action buttons — 已提取到 common.css === */

/* === Empty state === */
.schema-empty-wrapper {
  padding: 80px 20px;
}

/* Form hints */
.input-hint {
  display: block;
  font-size: 12px;
  color: var(--color-text-secondary);
  margin-top: 4px;
  line-height: 1.5;
}
</style>
