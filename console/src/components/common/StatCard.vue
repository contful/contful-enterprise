<script setup lang="ts">
// Copyright (c) 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

defineProps<{
  value: number
  label: string
  variant: 'sites' | 'entries' | 'schemas' | 'assets' | 'users' | 'tokens'
}>()

defineEmits<{
  click: []
}>()
</script>

<template>
  <div
    class="stat-card"
    :class="variant"
    role="button"
    tabindex="0"
    @click="$emit('click')"
    @keydown.enter="$emit('click')"
    @keydown.space.prevent="$emit('click')"
  >
    <div class="stat-icon">
      <slot name="icon" />
    </div>
    <div class="stat-content">
      <div class="stat-value">{{ value }}</div>
      <div class="stat-label">{{ label }}</div>
    </div>
  </div>
</template>

<style scoped>
.stat-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  cursor: pointer;
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.stat-card:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

/* ── 变体色（CSS 变量驱动）── */
.stat-card.sites    { --accent: var(--color-warning); --accent-bg: var(--color-warning-light); }
.stat-card.entries  { --accent: var(--color-primary); --accent-bg: var(--color-primary-light); }
.stat-card.schemas  { --accent: #8b5cf6; --accent-bg: #f3e8ff; }
.stat-card.assets   { --accent: var(--color-success); --accent-bg: var(--color-success-light); }
.stat-card.users    { --accent: #f59e0b; --accent-bg: #fef3c7; }
.stat-card.tokens   { --accent: #ec4899; --accent-bg: #fce7f3; }

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  background: var(--accent-bg);
  color: var(--accent);
  flex-shrink: 0;
}

.stat-content {
  min-width: 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: var(--color-text-secondary);
  margin-top: 2px;
}
</style>
