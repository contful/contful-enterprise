// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

/**
 * Contful Console — i18n 配置
 * 技术栈: vue-i18n v9 (Composition API 模式)
 */

import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN.json'
import zhTW from './zh-TW.json'
import enUS from './en-US.json'
import koKR from './ko-KR.json'
import jaJP from './ja-JP.json'
import frFR from './fr-FR.json'

export type Locale = 'zh-CN' | 'zh-TW' | 'en-US' | 'ko-KR' | 'ja-JP' | 'fr-FR'

const LOCALE_KEY = 'ct_console_locale'

/**
 * 从浏览器语言推导 Locale
 * 优先级：localStorage > navigator.language > 默认 zh-CN
 */
function detectLocale(): Locale {
  // 1. 优先读取用户已保存的语言偏好
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved && ['zh-CN', 'zh-TW', 'en-US', 'ko-KR', 'ja-JP', 'fr-FR'].includes(saved)) {
    return saved as Locale
  }

  // 2. 检测浏览器语言
  const navLang = navigator.language || 'zh-CN'

  if (navLang.startsWith('zh')) {
    const region = navLang.split('-')[1]?.toUpperCase()
    if (region && ['TW', 'HK', 'MO'].includes(region)) {
      return 'zh-TW'
    }
    return 'zh-CN'
  }

  if (navLang.startsWith('en')) {
    return 'en-US'
  }

  if (navLang.startsWith('ko')) {
    return 'ko-KR'
  }

  if (navLang.startsWith('ja')) {
    return 'ja-JP'
  }

  if (navLang.startsWith('fr')) {
    return 'fr-FR'
  }

  // 3. 默认简体中文
  return 'zh-CN'
}

export const i18n = createI18n({
  legacy: false, // 必须为 false，启用 Composition API 模式
  locale: detectLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'zh': zhCN,      // 兼容浏览器返回 zh 无地区后缀
    'zh-TW': zhTW,
    'en-US': enUS,
    'ko-KR': koKR,
    'ja-JP': jaJP,
    'fr-FR': frFR,
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
  { value: 'zh-TW' as Locale, label: '🇭🇰 繁體中文' },
  { value: 'en-US' as Locale, label: '🇺🇸 English' },
  { value: 'ja-JP' as Locale, label: '🇯🇵 日本語' },
  { value: 'ko-KR' as Locale, label: '🇰🇷 한국어' },
  { value: 'fr-FR' as Locale, label: '🇫🇷 Français' },
]
