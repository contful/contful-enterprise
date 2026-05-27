<script setup lang="ts">
// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { ref, watch } from 'vue'
import Editor from '@tinymce/tinymce-vue'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const editorValue = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  editorValue.value = val
})

const initConfig: Record<string, unknown> = {
  height: 400,
  menubar: false,
  plugins: [
    'advlist', 'autolink', 'lists', 'link', 'image',
    'charmap', 'preview', 'anchor', 'searchreplace', 'visualblocks',
    'code', 'fullscreen', 'insertdatetime', 'media', 'table',
    'help', 'wordcount',
  ],
  toolbar:
    'undo redo | blocks | ' +
    'bold italic forecolor backcolor | alignleft aligncenter ' +
    'alignright alignjustify | bullist numlist outdent indent | ' +
    'removeformat | help',
  content_style:
    'body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; font-size: 14px }',
  placeholder: props.placeholder,
  language: 'zh_CN',
  promotion: false,
  branding: false,
  license_key: 'o5lwq9uqku3wewduzmssmqkuqvfr0bfjdm6w7cxcx9j9hvrs',
}
</script>

<template>
  <Editor
    :model-value="editorValue"
    :init="initConfig"
    @update:model-value="(val: string) => emit('update:modelValue', val)"
  />
</template>
