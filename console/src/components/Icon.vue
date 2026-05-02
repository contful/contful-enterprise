<script setup lang="ts">

// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

/**
 * 简化图标组件 - SVG 高级用法
 * 替代 <t-icon name="xxx" />，统一使用 <Icon name="xxx" size="xx" />
 *
 * 图标来源：https://tdesign.tencent.com/icons
 * 支持所有 TDesign 内置图标名称
 *
 * @example
 * <Icon name="add" />
 * <Icon name="search" size="20px" />
 * <Icon name="chevron-down" size="14" />
 */
defineProps<{
  /** 图标名称，对应 TDesign 图标库名称 */
  name: string
  /** 图标尺寸，默认 24px */
  size?: string | number
}>()
</script>

<template>
  <t-icon :name="name" :size="size" />
</template>
