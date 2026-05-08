// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

/**
 * Contful Console - HTTP Client
 * 基于 Axios 封装，统一处理 Admin API 请求
 *
 * Token 存储策略（安全增强版）：
 * - AccessToken：JWT，存内存（页面刷新丢失），15min 有效
 * - RefreshToken：HttpOnly Cookie（JS 无法访问，XSS 无法窃取），7d 有效
 *
 * 安全改进：
 * - RefreshToken 不再存 localStorage，移到 HttpOnly Cookie
 * - 实现刷新竞态处理，防止多个并发请求同时触发刷新
 * - 页面初始化时自动从 Cookie 刷新获取 AccessToken
 */

import axios, { AxiosError, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'
import { MessagePlugin } from 'tdesign-vue-next'
import router from '@/router'

// ── Token 内存缓存（仅内存，刷新页面丢失）─────────────────────────────────────
let accessToken: string | null = null

// ── 刷新竞态控制 ─────────────────────────────────────────────────────────────
let isRefreshing = false
let refreshPromise: Promise<string> | null = null

// ── Cookie Key ─────────────────────────────────────────────────────────────────
const REFRESH_TOKEN_COOKIE = 'refresh_token'

// ── 从 HttpOnly Cookie 读取 refresh_token ─────────────────────────────────────
function getRefreshTokenFromCookie(): string | null {
  const match = document.cookie.match(new RegExp(`${REFRESH_TOKEN_COOKIE}=([^;]+)`))
  return match ? match[1] : null
}

// ── 清除 refresh_token Cookie ─────────────────────────────────────────────────
function clearRefreshTokenCookie(): void {
  document.cookie = `${REFRESH_TOKEN_COOKIE}=; Max-Age=0; Path=/`
}

// ── Access Token 读写（仅内存，无 localStorage）────────────────────────────────
export function getAccessToken(): string | null {
  return accessToken
}

export function setAccessToken(token: string | null): void {
  accessToken = token
}

// ── JWT 解码（提取 exp 字段）──────────────────────────────────────────────────
function decodeJWT(token: string): { exp: number } | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const payload = JSON.parse(atob(parts[1]))
    return payload
  } catch {
    return null
  }
}

// ── 检查 AccessToken 是否过期 ─────────────────────────────────────────────────
function isAccessTokenExpired(token: string | null): boolean {
  if (!token) return true
  const payload = decodeJWT(token)
  if (!payload || !payload.exp) return true
  // 提前 30 秒认为过期，避免时间误差
  return Date.now() >= (payload.exp * 1000 - 30000)
}

// ── 执行刷新（内部使用）────────────────────────────────────────────────────────
async function doRefresh(): Promise<string> {
  const cookieToken = getRefreshTokenFromCookie()
  if (!cookieToken) {
    throw new Error('no refresh token')
  }

  const res = await axios.post(
    `${import.meta.env.VITE_API_BASE_URL || '/admin/api/v1'}/auth/refresh`,
    {},
    {
      headers: {
        'Content-Type': 'application/json',
      },
      // 重要：不设置 Authorization Header，让后端从 Cookie 读取
    }
  )

  const data = res.data as API.Response<{ access_token: string; refresh_token: string }>

  if (data.code === 200 && data.data?.access_token) {
    // AccessToken 存内存
    accessToken = data.data.access_token
    return data.data.access_token
  }

  throw new Error('refresh failed')
}

// ── 获取有效的 AccessToken（带竞态处理）────────────────────────────────────────
async function getValidAccessToken(): Promise<string | null> {
  const token = accessToken

  // Token 有效且未过期，直接返回
  if (token && !isAccessTokenExpired(token)) {
    return token
  }

  // 无 Token 或已过期，触发刷新
  // 注意：必须先设置 isRefreshing=true，再设置 refreshPromise，防止竞态
  if (!isRefreshing) {
    isRefreshing = true
    try {
      refreshPromise = doRefresh()
      const newToken = await refreshPromise
      return newToken
    } catch {
      // 刷新失败
      return null
    } finally {
      isRefreshing = false
      refreshPromise = null
    }
  }

  // 已有刷新在进行中，等待结果
  if (refreshPromise) {
    try {
      return await refreshPromise
    } catch {
      return null
    }
  }

  return null
}

// ── 页面初始化：自动从 Cookie 刷新获取 AccessToken ────────────────────────────
export async function initializeSession(): Promise<boolean> {
  const cookieToken = getRefreshTokenFromCookie()
  if (!cookieToken) {
    return false
  }

  try {
    await doRefresh() // doRefresh 内部会设置 accessToken
    return true
  } catch {
    return false
  }
}

// ── Axios 实例 ─────────────────────────────────────────────────────────────
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/admin/api/v1',
  timeout: 30_000,
})

// ── 请求拦截器：附加 Authorization + X-Site-ID ────────────────────────────────
request.interceptors.request.use(
  async (config) => {
    // 获取有效 Token（自动刷新过期 Token）
    const token = await getValidAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    // 注入 X-Site-ID header
    const siteId = localStorage.getItem('currentSiteId')
    if (siteId) {
      config.headers['X-Site-ID'] = siteId
    }
    return config
  },
  (error) => Promise.reject(error)
)

// ── 响应拦截器：401 自动处理 ────────────────────────────────────────────────
request.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // 非 401 或已重试过，不再拦截
    if (!error.response || error.response.status !== 401 || originalRequest._retry) {
      return Promise.reject(error)
    }

    // 标记已重试，防止死循环
    originalRequest._retry = true

    // 尝试刷新获取新 Token
    try {
      const newToken = await getValidAccessToken()
      if (newToken) {
        // 重发被 401 拦截的原请求
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return request(originalRequest)
      }
    } catch {
      // 刷新失败
    }

    // 刷新失败，跳转登录
    redirectToLogin()
    return Promise.reject(error)
  }
)

// ── 跳转登录 ────────────────────────────────────────────────────────────────
function redirectToLogin() {
  accessToken = null
  clearRefreshTokenCookie()
  router.push('/login')
}

// ── 登出（供外部调用）────────────────────────────────────────────────────────
export function logout() {
  accessToken = null
  clearRefreshTokenCookie()
}

// ── 通用 GET/POST/PUT/DELETE ───────────────────────────────────────────────
export function get<T = unknown>(url: string, config?: AxiosRequestConfig) {
  return request.get<API.Response<T>>(url, config).then((res) => res.data)
}

export function post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  return request.post<API.Response<T>>(url, data, config).then((res) => res.data)
}

export function put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  return request.put<API.Response<T>>(url, data, config).then((res) => res.data)
}

export function del<T = unknown>(url: string, config?: AxiosRequestConfig) {
  return request.delete<API.Response<T>>(url, config).then((res) => res.data)
}

// ── 用户反馈：showSuccess / showError ──────────────────────────────────────

type ErrorLike = Error | AxiosError | { response?: { data?: { msg?: string; message?: string } }; message?: string } | string

/**
 * 成功提示
 */
export function showSuccess(message: string | { msg?: string; message?: string }): void {
  const msg = typeof message === 'string' ? message : (message.msg || message.message || '操作成功')
  MessagePlugin.success(msg)
}

/**
 * 错误提示（自动从多种格式提取错误信息）
 */
export function showError(error: ErrorLike): void {
  let msg = '操作失败'

  if (typeof error === 'string') {
    msg = error
  } else if (error && typeof error === 'object') {
    // AxiosError 优先取 response.data.msg
    const axiosError = error as AxiosError<{ msg?: string; message?: string }>
    msg = axiosError.response?.data?.msg
      || axiosError.response?.data?.message
      || axiosError.message
      || msg
  }

  MessagePlugin.error(msg)
}

/**
 * 提取友好错误信息（用于表单字段级错误展示）
 */
export function getFriendlyError(error: unknown): string {
  if (!error) return ''

  const err = error as { response?: { data?: { msg?: string; message?: string; errors?: Record<string, string[]> } }; message?: string }

  // 优先取字段级错误
  if (err.response?.data?.errors) {
    const errors = err.response.data.errors
    const firstKey = Object.keys(errors)[0]
    if (firstKey && Array.isArray(errors[firstKey])) {
      return errors[firstKey][0] || ''
    }
  }

  return err.response?.data?.msg || err.response?.data?.message || err.message || ''
}

export default request
