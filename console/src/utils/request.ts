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

// ── 清除 refresh_token Cookie（前端侧兜底清理，HttpOnly Cookie 清除由后端负责）─────
function clearRefreshTokenCookie(): void {
  document.cookie = `refresh_token=; Max-Age=0; Path=/`
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
  // refresh_token 存储在 HttpOnly Cookie 中（前端 JS 无法通过 document.cookie 读取）。
  // 直接向后端 /auth/refresh 发送请求，后端自动从 Cookie 读取 refresh_token。
  // 使用独立的 axios 实例避免触发请求拦截器（拦截器会尝试再加 Authorization header）
  // 同时设置 withCredentials: true 确保 Cookie 被发送。
  console.log('[doRefresh] 尝试刷新 Token, URL:', `${import.meta.env.VITE_API_BASE_URL || '/admin/api/v1'}/auth/refresh`)
  try {
    const res = await axios.post(
      `${import.meta.env.VITE_API_BASE_URL || '/admin/api/v1'}/auth/refresh`,
      {},
      {
        headers: { 'Content-Type': 'application/json' },
        withCredentials: true,
      }
    )
    const data = res.data as API.Response<{ access_token: string; refresh_token: string }>
    console.log('[doRefresh] 响应:', JSON.stringify(data))
    if (data.code === 200 && data.data?.access_token) {
      accessToken = data.data.access_token
      console.log('[doRefresh] AccessToken 已更新')
      return data.data.access_token
    }
    throw new Error('refresh failed: code=' + data.code)
  } catch (err: any) {
    console.error('[doRefresh] 刷新失败:', err?.response?.status, err?.response?.data)
    throw err
  }
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
  // refresh_token 存在 HttpOnly Cookie 中，前端无法通过 document.cookie 读取。
  // 直接尝试调用 refresh 接口，让后端从 Cookie 获取 token。
  try {
    await doRefresh()
    return true
  } catch {
    return false
  }
}

// ── Axios 实例 ─────────────────────────────────────────────────────────────
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/admin/api/v1',
  timeout: 30_000,
  withCredentials: true, // 确保所有请求携带 HttpOnly Cookie（refresh_token）
})

// ── 请求拦截器：附加 Authorization + X-Site-ID ────────────────────────────────
request.interceptors.request.use(
  async (config) => {
    const url = (config.url || '').replace(config.baseURL || '', '')
    console.log(`[Interceptor] 请求: ${config.method?.toUpperCase()} ${url}`)
    // 跳过不需要 Token 的请求（登录、注册、刷新、RSA 公钥等公开端点）
    if (url.startsWith('/auth/')) {
      console.log('[Interceptor] 跳过公开路径 /auth/*，不附加 Token')
      return config
    }
    // 精确匹配公开配置端点（site + public），其余 /system/config/* 需要鉴权
    if (url === '/system/config/site' || url === '/system/config/public') {
      console.log('[Interceptor] 跳过公开路径，不附加 Token')
      return config
    }

    // 获取有效 Token（自动刷新过期 Token）
    const token = await getValidAccessToken()
    console.log('[Interceptor] getValidAccessToken 返回:', token ? '有Token(' + token.substring(0, 10) + '...)' : 'null')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
      console.log('[Interceptor] 已附加 Authorization header')
    } else {
      console.warn('[Interceptor] 无有效 Token，请求将不带 Authorization')
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
    console.error("[Response Interceptor] 401 响应体:", error.response?.data)
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

export function patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  return request.patch<API.Response<T>>(url, data, config).then((res) => res.data)
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
