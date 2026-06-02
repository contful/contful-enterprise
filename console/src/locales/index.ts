// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

/**
 * Contful Console — i18n 配置
 * 技术栈: vue-i18n v9 (Composition API 模式)
 * 支持 10 种语言
 */

import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN.json'
import zhTW from './zh-TW.json'
import enUS from './en-US.json'
import koKR from './ko-KR.json'
import jaJP from './ja-JP.json'
import frFR from './fr-FR.json'
import ptBR from './pt-BR.json'
import ruRU from './ru-RU.json'
import esES from './es-ES.json'
import faIR from './fa-IR.json'

export type Locale = 'zh-CN' | 'zh-TW' | 'en-US' | 'ko-KR' | 'ja-JP' | 'fr-FR' | 'pt-BR' | 'ru-RU' | 'es-ES' | 'fa-IR'

const LOCALE_KEY = 'ct_console_locale'

const VALID_LOCALES: string[] = ['zh-CN', 'zh-TW', 'en-US', 'ko-KR', 'ja-JP', 'fr-FR', 'pt-BR', 'ru-RU', 'es-ES', 'fa-IR']

/**
 * 从浏览器语言推导 Locale
 * 优先级：localStorage > navigator.language > 默认 zh-CN
 */
function detectLocale(): Locale {
  // 1. 优先读取用户已保存的语言偏好
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved && VALID_LOCALES.includes(saved)) {
    return saved as Locale
  }

  // 2. 检测浏览器语言
  const navLang = navigator.language || 'zh-CN'

  if (navLang.startsWith('zh')) {
    const region = navLang.split('-')[1]?.toUpperCase()
    if (region && ['TW', 'HK', 'MO'].includes(region)) return 'zh-TW'
    return 'zh-CN'
  }
  if (navLang.startsWith('en')) return 'en-US'
  if (navLang.startsWith('ko')) return 'ko-KR'
  if (navLang.startsWith('ja')) return 'ja-JP'
  if (navLang.startsWith('fr')) return 'fr-FR'
  if (navLang.startsWith('pt')) return 'pt-BR'
  if (navLang.startsWith('ru')) return 'ru-RU'
  if (navLang.startsWith('es')) return 'es-ES'
  if (navLang.startsWith('fa')) return 'fa-IR'

  // 3. 默认简体中文
  return 'zh-CN'
}

export const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN, 'zh': zhCN,
    'zh-TW': zhTW,
    'en-US': enUS,
    'ko-KR': koKR,
    'ja-JP': jaJP,
    'fr-FR': frFR,
    'pt-BR': ptBR,
    'ru-RU': ruRU,
    'es-ES': esES,
    'fa-IR': faIR,
  },
})

export function setLocale(locale: Locale) {
  i18n.global.locale.value = locale
  localStorage.setItem(LOCALE_KEY, locale)
  document.documentElement.lang = locale
  document.documentElement.dir = locale === 'fa-IR' ? 'rtl' : 'ltr'
}

export function getLocale(): Locale {
  return i18n.global.locale.value as Locale
}

export const localeOptions = [
  { value: 'zh-CN' as Locale, label: '🇨🇳 简体中文' },
  { value: 'zh-TW' as Locale, label: '🇹🇼 繁體中文' },
  { value: 'en-US' as Locale, label: '🇺🇸 English' },
  { value: 'ja-JP' as Locale, label: '🇯🇵 日本語' },
  { value: 'ko-KR' as Locale, label: '🇰🇷 한국어' },
  { value: 'fr-FR' as Locale, label: '🇫🇷 Français' },
  { value: 'pt-BR' as Locale, label: '🇧🇷 Português' },
  { value: 'ru-RU' as Locale, label: '🇷🇺 Русский' },
  { value: 'es-ES' as Locale, label: '🇪🇸 Español' },
  { value: 'fa-IR' as Locale, label: '🇮🇷 فارسی' },
]
