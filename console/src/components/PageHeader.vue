<script setup lang="ts">
/**
 * PageHeader — 统一页面头部组件
 * 
 * 功能：
 * - 显示页面标题和副标题
 * - 可选刷新按钮（右上角）
 * - 提供 actions 插槽（左侧操作区）和 primary-action 插槽（右侧主操作按钮）
 */
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

withDefaults(defineProps<{
  title?: string
  subtitle?: string
  showRefresh?: boolean
}>(), {
  title: '',
  subtitle: '',
  showRefresh: false,
})

const emit = defineEmits<{
  refresh: []
}>()
</script>

<template>
  <div class="page-header">
    <div>
      <h1 v-if="title" class="page-title">{{ title }}</h1>
      <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
      <slot name="subtitle" />
    </div>
    <div class="header-actions">
      <slot name="actions" />
      <t-button
        v-if="showRefresh"
        variant="outline"
        @click="emit('refresh')"
      >
        <template #icon><t-icon name="refresh" /></template>
        {{ t('common.refresh') }}
      </t-button>
      <slot name="primary-action" />
    </div>
  </div>
</template>
