<template>
  <t-card class="advanced-filter">
    <div class="filter-row">
      <t-input v-model="filters.keyword" :placeholder="t('audit.advancedFilter.keywordPlaceholder')" clearable class="keyword-input" @enter="emitSearch">
        <template #prefix-icon><t-icon name="search" /></template>
      </t-input>
      <t-radio-group v-model="filters.combinator" variant="default-filled" size="small">
        <t-radio-button value="and">{{ t('audit.advancedFilter.and') }}</t-radio-button>
        <t-radio-button value="or">{{ t('audit.advancedFilter.or') }}</t-radio-button>
      </t-radio-group>
      <t-button-group variant="outline" size="small">
        <t-button :variant="filters.time_preset === '1h' ? 'base' : 'outline'" @click="setPreset('1h')">{{ t('audit.advancedFilter.preset1h') }}</t-button>
        <t-button :variant="filters.time_preset === '24h' ? 'base' : 'outline'" @click="setPreset('24h')">{{ t('audit.advancedFilter.preset24h') }}</t-button>
        <t-button :variant="filters.time_preset === '7d' ? 'base' : 'outline'" @click="setPreset('7d')">{{ t('audit.advancedFilter.preset7d') }}</t-button>
      </t-button-group>
      <t-button theme="primary" size="small" @click="emitSearch">{{ t('audit.filter.search') }}</t-button>
    </div>
  </t-card>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const emit = defineEmits<{
  search: [filters: { keyword: string; combinator: string; time_preset: string }]
}>()

const filters = reactive({
  keyword: '',
  combinator: 'and',
  time_preset: '',
})

function setPreset(val: string) {
  filters.time_preset = filters.time_preset === val ? '' : val
  emitSearch()
}

function emitSearch() {
  emit('search', { ...filters })
}
</script>

<style scoped>
.advanced-filter { margin-bottom: 12px; }
.filter-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.keyword-input { flex: 1; min-width: 200px; }
</style>
