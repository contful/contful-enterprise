<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next'
import PageHeader from '@/components/PageHeader.vue'
import {
  listWebhooks, createWebhook, updateWebhook, deleteWebhook,
  listDeliveries, testWebhook, EVENT_OPTIONS,
  type Webhook, type WebhookDelivery,
} from '@/api/webhook'
import { showError } from '@/utils/request'

const { t } = useI18n()
const loading = ref(false)
const webhooks = ref<Webhook[]>([])
const showDialog = ref(false)
const isEdit = ref(false)
const saving = ref(false)

const form = ref<Partial<Webhook>>({
  name: '', url: '', events: [], secret: '', is_active: true,
})

const formRules = {
  name: [{ required: true, message: '请输入名称' }],
  url: [{ required: true, message: '请输入 URL' }],
  events: [{ required: true, message: '至少选择一个事件' }],
}

// 投递记录
const deliveryDialog = ref(false)
const deliveries = ref<WebhookDelivery[]>([])
const deliveryLoading = ref(false)
const currentWebhook = ref<Webhook | null>(null)

onMounted(() => fetchList())

async function fetchList() {
  loading.value = true
  try {
    const res = await listWebhooks()
    webhooks.value = res.data || []
  } catch (e) { handleError(e) }
  finally { loading.value = false }
}

function openCreate() {
  isEdit.value = false
  form.value = { name: '', url: '', events: [], secret: '', is_active: true }
  showDialog.value = true
}

function openEdit(w: Webhook) {
  isEdit.value = true
  form.value = { ...w }
  showDialog.value = true
}

async function handleSave() {
  saving.value = true
  try {
    if (isEdit.value && form.value.id) {
      await updateWebhook(form.value.id, form.value)
      MessagePlugin.success('已更新')
    } else {
      await createWebhook(form.value)
      MessagePlugin.success('已创建')
    }
    showDialog.value = false
    fetchList()
  } catch (e) { handleError(e) }
  finally { saving.value = false }
}

async function handleDelete(w: Webhook) {
  const dlg = DialogPlugin.confirm({
    header: t('common.confirm'),
    body: `确定删除 Webhook「${w.name}」？`,
    confirmBtn: t('common.delete'),
    cancelBtn: t('common.cancel'),
    theme: 'danger',
    onConfirm: async () => {
      try {
        await deleteWebhook(w.id)
        MessagePlugin.success('已删除')
        dlg.destroy()
        fetchList()
      } catch (e) { handleError(e) }
    },
  })
}

async function handleTest(w: Webhook) {
  try {
    await testWebhook(w.id)
    MessagePlugin.success('测试请求已发送')
  } catch (e) { handleError(e) }
}

async function openDeliveries(w: Webhook) {
  currentWebhook.value = w
  deliveryLoading.value = true
  deliveryDialog.value = true
  try {
    const res = await listDeliveries(w.id)
    deliveries.value = res.data || []
  } catch (e) { handleError(e) }
  finally { deliveryLoading.value = false }
}

function eventLabel(ev: string) {
  const opt = EVENT_OPTIONS.find(o => o.value === ev)
  return opt ? opt.label : ev
}

function deliveryStatusTag(status: string) {
  const map: Record<string, [string, string]> = {
    success: ['success', '成功'],
    failed: ['danger', '失败'],
    pending: ['warning', '处理中'],
  }
  const [theme, label] = map[status] || ['default', status]
  return { theme, label }
}

function formatTime(t: string) { return t?.replace('T', ' ').replace('Z', '') || '-' }

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  MessagePlugin.success('已复制')
}

function handleError(e: unknown) {
  showError(String((e as any)?.response?.data?.msg || (e as any)?.message || e))
}
</script>

<template>
  <div class="webhook-page">
    <PageHeader :title="t('webhook.title')" :subtitle="t('webhook.subtitle')">
      <template #actions>
        <t-button theme="primary" @click="openCreate">
          <template #icon><t-icon name="add" /></template>
          {{ t('webhook.create') }}
        </t-button>
      </template>
    </PageHeader>

    <t-table
      :data="webhooks"
      :loading="loading"
      row-key="id"
      :columns="[
        { colKey: 'name', title: t('webhook.name'), width: 180 },
        { colKey: 'url', title: 'URL', ellipsis: true },
        { colKey: 'events', title: t('webhook.events'), width: 250 },
        { colKey: 'status', title: t('webhook.status'), width: 80 },
        { colKey: 'actions', title: t('webhook.actions'), width: 220 },
      ]"
      table-layout="auto"
      hover
      stripe
    >
      <template #url="{ row }">
        <span class="mono-cell">{{ row.url }}</span>
      </template>
      <template #events="{ row }">
        <t-space :size="4">
          <t-tag v-for="ev in row.events" :key="ev" size="small" variant="light">{{ eventLabel(ev) }}</t-tag>
        </t-space>
      </template>
      <template #status="{ row }">
        <t-tag :theme="row.is_active ? 'success' : 'default'" size="small" variant="light">
          {{ row.is_active ? t('webhook.active') : t('webhook.inactive') }}
        </t-tag>
      </template>
      <template #actions="{ row }">
        <t-space :size="4">
          <t-button size="small" variant="text" @click="openEdit(row)">{{ t('common.edit') }}</t-button>
          <t-button size="small" variant="text" theme="warning" @click="handleTest(row)">{{ t('webhook.test') }}</t-button>
          <t-button size="small" variant="text" @click="openDeliveries(row)">{{ t('webhook.deliveries') }}</t-button>
          <t-button size="small" variant="text" theme="danger" @click="handleDelete(row)">{{ t('common.delete') }}</t-button>
        </t-space>
      </template>
    </t-table>

    <!-- 创建/编辑弹窗 -->
    <t-dialog
      v-model:visible="showDialog"
      :header="isEdit ? t('webhook.editTitle') : t('webhook.createTitle')"
      :confirm-btn="saving ? undefined : { content: t('common.save'), theme: 'primary' }"
      :cancel-btn="t('common.cancel')"
      @confirm="handleSave"
      :close-on-overlay-click="false"
      width="560px"
    >
      <t-form :data="form" :rules="formRules" label-width="80px">
        <t-form-item :label="t('webhook.name')" name="name">
          <t-input v-model="form.name" placeholder="例如：Contentful 同步" />
        </t-form-item>
        <t-form-item label="URL" name="url">
          <t-input v-model="form.url" placeholder="https://example.com/webhook" />
        </t-form-item>
        <t-form-item :label="t('webhook.events')" name="events">
          <t-checkbox-group v-model="form.events">
            <t-checkbox v-for="opt in EVENT_OPTIONS" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </t-checkbox>
          </t-checkbox-group>
        </t-form-item>
        <t-form-item :label="t('webhook.secret')" name="secret">
          <t-input v-model="form.secret" placeholder="可选：HMAC-SHA256 签名密钥" />
        </t-form-item>
        <t-form-item :label="t('webhook.status')" name="is_active">
          <t-switch v-model="form.is_active" :label="[t('webhook.active'), t('webhook.inactive')]" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 投递记录弹窗 -->
    <t-dialog
      v-model:visible="deliveryDialog"
      :header="t('webhook.deliveries') + ' — ' + (currentWebhook?.name || '')"
      :footer="false"
      width="800px"
    >
      <t-table
        :data="deliveries"
        :loading="deliveryLoading"
        row-key="id"
        :columns="[
          { colKey: 'event', title: t('webhook.event'), width: 130 },
          { colKey: 'status', title: t('webhook.status'), width: 80 },
          { colKey: 'response_status', title: 'HTTP', width: 70 },
          { colKey: 'attempt', title: t('webhook.attempt'), width: 60 },
          { colKey: 'error_message', title: t('webhook.error'), ellipsis: true },
          { colKey: 'created_time', title: t('webhook.time'), width: 160 },
        ]"
        size="small"
        hover
        max-height="400px"
      >
        <template #event="{ row }">
          <t-tag size="small" variant="light">{{ eventLabel(row.event) }}</t-tag>
        </template>
        <template #status="{ row }">
          <t-tag size="small" variant="light" :theme="deliveryStatusTag(row.status).theme">
            {{ deliveryStatusTag(row.status).label }}
          </t-tag>
        </template>
        <template #created_time="{ row }">{{ formatTime(row.created_time) }}</template>
        <template #error_message="{ row }">{{ row.error_message || '-' }}</template>
      </t-table>
    </t-dialog>
  </div>
</template>

<style scoped>
.webhook-page {
  width: 100%;
}
.mono-cell {
  font-family: 'SF Mono', Monaco, Menlo, monospace;
  font-size: 13px;
}
</style>
