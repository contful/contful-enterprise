/**
 * Contful Console — i18n 配置
 * 技术栈: vue-i18n v9 (Composition API 模式)
 */

import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN.json'
import enUS from './en-US.json'

export type Locale = 'zh-CN' | 'en-US'

const LOCALE_KEY = 'ct_console_locale'

export const i18n = createI18n({
  legacy: false, // 必须为 false，启用 Composition API 模式
  locale: localStorage.getItem(LOCALE_KEY) || 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})

/**
 * 切换 Console UI 语言
 * 会同步更新 localStorage 存储，下次访问自动恢复
 */
export function setLocale(locale: Locale) {
  i18n.global.locale.value = locale
  localStorage.setItem(LOCALE_KEY, locale)
  document.documentElement.lang = locale
}

/**
 * 获取当前语言
 */
export function getLocale(): Locale {
  return i18n.global.locale.value as Locale
}

/**
 * 语言切换选项（用于下拉菜单）
 */
export const localeOptions = [
  { value: 'zh-CN' as Locale, label: '🇨🇳 简体中文' },
  { value: 'en-US' as Locale, label: '🇺🇸 English' },
]
