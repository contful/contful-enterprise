<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'

const props = defineProps<{
  modelValue?: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const rawText = ref(props.modelValue || '')
const mode = ref<'tree' | 'raw'>('tree')

watch(() => props.modelValue, (val) => {
  rawText.value = val || ''
})

const parsedJson = computed(() => {
  if (!rawText.value) return {}
  try {
    return JSON.parse(rawText.value)
  } catch {
    return null
  }
})

const handleRawInput = (e: Event) => {
  const target = e.target as HTMLTextAreaElement
  rawText.value = target.value
  emit('update:modelValue', target.value)
}

const toggleMode = () => {
  if (mode.value === 'tree') {
    // 切换到源码前格式化
    try {
      const obj = JSON.parse(rawText.value || '{}')
      rawText.value = JSON.stringify(obj, null, 2)
    } catch { /* 保持原样 */ }
  }
  mode.value = mode.value === 'tree' ? 'raw' : 'tree'
}
</script>

<template>
  <div class="json-editor">
    <div class="json-editor-toolbar">
      <span class="json-editor-mode">
        {{ mode === 'tree' ? '树形查看' : '源码编辑' }}
      </span>
      <button type="button" class="json-editor-toggle" @click="toggleMode">
        {{ mode === 'tree' ? '编辑源码' : '查看树形' }}
      </button>
    </div>
    <VueJsonPretty
      v-if="mode === 'tree' && parsedJson"
      :data="parsedJson"
      :deep="3"
      :showLength="true"
    />
    <div v-else-if="mode === 'tree'" class="json-editor-empty">
      {{ placeholder || '请输入 JSON' }}
    </div>
    <textarea
      v-else
      class="json-editor-raw"
      :value="rawText"
      :placeholder="placeholder || '请输入 JSON'"
      @input="handleRawInput"
    ></textarea>
  </div>
</template>

<style scoped>
.json-editor {
  width: 100%;
  border: 1px solid var(--td-component-border);
  border-radius: 6px;
  overflow: hidden;
}
.json-editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 12px;
  background: var(--td-bg-color-secondarycontainer);
  border-bottom: 1px solid var(--td-component-border);
  font-size: 12px;
}
.json-editor-mode {
  color: var(--td-text-color-secondary);
}
.json-editor-toggle {
  background: none;
  border: 1px solid var(--td-component-border);
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 12px;
  color: var(--td-brand-color);
  cursor: pointer;
}
.json-editor-toggle:hover {
  background: var(--td-brand-color-light);
}
.json-editor :deep(.vjs-tree) {
  padding: 8px 12px;
}
.json-editor-raw {
  width: 100%;
  min-height: 200px;
  padding: 12px;
  border: none;
  font-family: 'SF Mono', 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  outline: none;
  box-sizing: border-box;
}
.json-editor-empty {
  padding: 24px;
  text-align: center;
  color: var(--td-text-color-placeholder);
}
</style>
