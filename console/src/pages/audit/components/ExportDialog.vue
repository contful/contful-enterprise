<template>
  <t-dialog :visible="visible" :header="t('audit.export.title')" width="500" @close="$emit('close')" @confirm="handleExport">
    <div class="export-body">
      <div class="export-section">
        <label>{{ t('audit.export.format') }}</label>
        <t-radio-group v-model="format">
          <t-radio value="csv">CSV</t-radio>
          <t-radio value="xlsx">XLSX</t-radio>
          <t-radio value="json">JSON</t-radio>
        </t-radio-group>
      </div>
      <div class="export-section">
        <label>{{ t('audit.export.fields') }}</label>
        <t-checkbox-group v-model="fields">
          <t-checkbox v-for="f in availableFields" :key="f.value" :value="f.value">{{ f.label }}</t-checkbox>
        </t-checkbox-group>
      </div>
    </div>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
defineProps<{ visible: boolean }>()
const emit = defineEmits<{ close: []; export: [{ format: string; fields: string[] }] }>()

const format = ref('csv')
const fields = ref<string[]>(['id','action','category','level','resource_type','ip_address','details','created_time'])
const availableFields = [
  { label: 'ID', value: 'id' }, { label: t('audit.column.action'), value: 'action' },
  { label: t('audit.column.category'), value: 'category' }, { label: t('audit.column.level'), value: 'level' },
  { label: t('audit.column.resourceType'), value: 'resource_type' }, { label: t('audit.column.ipAddress'), value: 'ip_address' },
  { label: t('audit.detail.details'), value: 'details' }, { label: t('audit.column.time'), value: 'created_time' },
]

function handleExport() {
  emit('export', { format: format.value, fields: fields.value })
}
</script>

<style scoped>
.export-body { display: flex; flex-direction: column; gap: 20px; }
.export-section { display: flex; flex-direction: column; gap: 10px; }
.export-section label { font-weight: 500; font-size: 14px; }
</style>
