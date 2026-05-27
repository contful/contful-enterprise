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
const editingData = ref<Record<string, unknown>>({})
const mode = ref<'tree' | 'raw'>('tree')

watch(() => props.modelValue, (val) => {
  rawText.value = val || ''
  try {
    editingData.value = val ? JSON.parse(val) : {}
  } catch {
    editingData.value = {}
  }
})

const parsedJson = computed(() => {
  if (!rawText.value) return {}
  try {
    return JSON.parse(rawText.value)
  } catch {
    return null
  }
})

watch(editingData, (val) => {
  rawText.value = JSON.stringify(val, null, 2)
  emit('update:modelValue', rawText.value)
}, { deep: true })

const handleRawInput = (e: Event) => {
  const target = e.target as HTMLTextAreaElement
  rawText.value = target.value
  emit('update:modelValue', target.value)
}

const toggleMode = () => {
  mode.value = mode.value === 'tree' ? 'raw' : 'tree'
}
</script>

<template>
  <div class="json-editor">
    <div class="json-editor-toolbar">
      <span class="json-editor-mode">
        {{ mode === 'tree' ? '树形编辑' : '源码编辑' }}
      </span>
      <button type="button" class="json-editor-toggle" @click="toggleMode">
        {{ mode === 'tree' ? '切换源码' : '切换树形' }}
      </button>
    </div>
    <VueJsonPretty
      v-if="mode === 'tree' && parsedJson"
      v-model:data="editingData"
      :deep="3"
      :editable="true"
      :editableTrigger="'dblclick'"
      :showLength="true"
      :placeholder="placeholder || '请输入 JSON'"
    />
    <textarea
      v-else-if="mode === 'raw'"
      class="json-editor-raw"
      :value="rawText"
      :placeholder="placeholder || '请输入 JSON'"
      @input="handleRawInput"
    ></textarea>
    <div v-else class="json-editor-empty">
      {{ placeholder || '请输入 JSON' }}
    </div>
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
}
.json-editor-empty {
  padding: 24px;
  text-align: center;
  color: var(--td-text-color-placeholder);
}
</style>
