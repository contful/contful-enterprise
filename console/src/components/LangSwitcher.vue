<template>
  <t-dropdown>
    <t-button variant="text" theme="default" class="lang-switcher-btn">
      <span class="lang-icon">{{ currentLocale === 'zh-CN' ? '🇨🇳' : '🇺🇸' }}</span>
      <span class="lang-label">{{ currentLocaleLabel }}</span>
      <Icon name="chevron-down" size="14px" />
    </t-button>
    <template #dropdown>
      <t-dropdown-menu>
        <t-dropdown-item
          v-for="opt in localeOptions"
          :key="opt.value"
          :active="opt.value === currentLocale"
          @click="switchLocale(opt.value)"
        >
          {{ opt.label }}
        </t-dropdown-item>
      </t-dropdown-menu>
    </template>
  </t-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { i18n, localeOptions, setLocale, type Locale } from '@/locales'

const { t } = useI18n()
const currentLocale = computed(() => i18n.global.locale.value as Locale)

const currentLocaleLabel = computed(() => {
  const opt = localeOptions.find(o => o.value === currentLocale.value)
  return opt?.label || (currentLocale.value === 'zh-CN' ? t('app.langZhCN') : t('app.langEn'))
})

function switchLocale(locale: Locale) {
  if (locale === currentLocale.value) return
  setLocale(locale)
  location.reload()
}
</script>

<style scoped>
.lang-switcher-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  padding: 4px 8px;
  border-radius: 6px;
}
.lang-switcher-btn:hover {
  background: var(--color-hover);
}
.lang-icon {
  font-size: 16px;
}
.lang-label {
  font-size: 13px;
}
</style>
