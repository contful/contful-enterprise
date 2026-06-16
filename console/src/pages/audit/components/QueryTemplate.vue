<template>
  <div class="query-template">
    <t-select v-model="selectedId" :placeholder="t('audit.queryTemplate.placeholder')" clearable size="small" style="width:200px" @change="handleSelect">
      <t-option v-for="tmpl in templates" :key="tmpl.id" :label="tmpl.name" :value="tmpl.id" />
    </t-select>
    <t-button variant="outline" size="small" @click="showSave=true"><t-icon name="save" /></t-button>
    <t-popconfirm v-if="selectedId" :content="t('audit.queryTemplate.deleteConfirm')" @confirm="handleDelete">
      <t-button variant="outline" size="small" theme="danger"><t-icon name="delete" /></t-button>
    </t-popconfirm>

    <t-dialog v-model:visible="showSave" :header="t('audit.queryTemplate.saveTitle')" @confirm="handleSave">
      <t-input v-model="templateName" :placeholder="t('audit.queryTemplate.namePlaceholder')" />
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { MessagePlugin } from 'tdesign-vue-next'
import { getQueryTemplates, createQueryTemplate, deleteQueryTemplate } from '@/api/audit-query-template'
import type { AuditQueryTemplate } from '@/types/audit-query-template'

const { t } = useI18n()
const emit = defineEmits<{ apply: [conditions: Record<string, any>] }>()

const templates = ref<AuditQueryTemplate[]>([])
const selectedId = ref('')
const showSave = ref(false)
const templateName = ref('')

async function fetchTemplates() {
  try { const res = await getQueryTemplates(); templates.value = res.data || [] } catch {}
}
function handleSelect(id: string) {
  const tmpl = templates.value.find(t => t.id === id)
  if (tmpl) emit('apply', tmpl.conditions as Record<string, any>)
}
async function handleSave() {
  if (!templateName.value) return
  try {
    await createQueryTemplate({ name: templateName.value, conditions: {} })
    MessagePlugin.success(t('audit.queryTemplate.saveSuccess'))
    showSave.value = false; templateName.value = ''; fetchTemplates()
  } catch {}
}
async function handleDelete() {
  try {
    await deleteQueryTemplate(selectedId.value)
    MessagePlugin.success(t('audit.queryTemplate.deleteSuccess'))
    selectedId.value = ''; fetchTemplates()
  } catch {}
}
onMounted(fetchTemplates)
</script>

<style scoped>
.query-template { display: flex; align-items: center; gap: 8px; }
</style>
