/**
 * Contful Console - HTTP Client
 * 基于 Axios 封装，统一处理 Admin API 请求
 *
 * Token 存储策略（简化版，无 Refresh Token 轮换）：
 * - AccessToken：JWT，存 localStorage + 内存变量，15min 有效
 * - RefreshToken：hex token，存 localStorage（无 HttpOnly），7d 有效
 *
 * 所有 token 均通过 localStorage 持久化，页面刷新后仍可用。
 */

import axios, { AxiosError, type AxiosRequestConfig, type InternalAxiosRequestConfig } from 'axios'
import { MessagePlugin } from 'tdesign-vue-next'
import router from '@/router'

// ── Token 内存缓存 ──────────────────────────────────────────────────────────
let accessToken: string | null = null

// ── localStorage Key ────────────────────────────────────────────────────────
const ACCESS_TOKEN_KEY = 'ct_access_token'
const REFRESH_TOKEN_KEY = 'ct_refresh_token'

// ── Access Token 读写 ────────────────────────────────────────────────────────
export function getAccessToken(): string | null {
  return accessToken
}

export function setAccessToken(token: string | null): void {
  accessToken = token
  if (token) {
    localStorage.setItem(ACCESS_TOKEN_KEY, token)
  } else {
    localStorage.removeItem(ACCESS_TOKEN_KEY)
  }
}

// ── Refresh Token 读写（localStorage 明文存储）───────────────────────────────
export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setRefreshToken(token: string | null): void {
  if (token) {
    localStorage.setItem(REFRESH_TOKEN_KEY, token)
  } else {
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  }
}

export function clearRefreshToken(): void {
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

// ── 初始化：从 localStorage 恢复 Access Token ──────────────────────────────
setAccessToken(localStorage.getItem(ACCESS_TOKEN_KEY))

// ── Axios 实例 ─────────────────────────────────────────────────────────────
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/admin/api/v1',
  timeout: 30_000,
})

// ── 请求拦截器：附加 Authorization + X-Site-ID ───────────────────────────────
request.interceptors.request.use(
  (config) => {
    // 优先用内存中的 token（最新）
    const token = accessToken || localStorage.getItem(ACCESS_TOKEN_KEY)
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

// ── 响应拦截器：401 自动 Refresh ───────────────────────────────────────────
request.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // 非 401 或已重试过，不再拦截
    if (!error.response || error.response.status !== 401 || originalRequest._retry) {
      return Promise.reject(error)
    }
    originalRequest._retry = true

    const refreshToken = getRefreshToken()
    if (!refreshToken) {
      redirectToLogin()
      return Promise.reject(error)
    }

    try {
      // 调用 Refresh 接口，返回新的 access_token + refresh_token
      const res = await axios.post(
        `${import.meta.env.VITE_API_BASE_URL || '/admin/api/v1'}/auth/refresh`,
        {},
        {
          headers: {
            Authorization: `Bearer ${refreshToken}`,
            'Content-Type': 'application/json',
          },
        }
      )

      const data = res.data as API.Response<{ access_token: string; refresh_token: string }>

      if (data.code === 200 && data.data) {
        // 更新 AccessToken（内存 + localStorage）
        setAccessToken(data.data.access_token)
        // 更新 RefreshToken（localStorage）
        setRefreshToken(data.data.refresh_token)

        // 重发被 401 拦截的原请求
        originalRequest.headers.Authorization = `Bearer ${data.data.access_token}`
        return request(originalRequest)
      } else {
        // Refresh 也失败，说明 Refresh Token 失效
        redirectToLogin()
      }
    } catch {
      // Refresh 请求本身出错（网络问题等），跳登录
      redirectToLogin()
    }

    return Promise.reject(error)
  }
)

// ── 跳转登录 ────────────────────────────────────────────────────────────────
function redirectToLogin() {
  setAccessToken(null)
  clearRefreshToken()
  router.push('/login')
}

// ── 通用 GET/POST/PUT/DELETE ────────────────────────────────────────────────
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