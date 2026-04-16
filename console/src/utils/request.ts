import axios, { type AxiosInstance, type AxiosError, type AxiosRequestConfig } from 'axios'
import { MessagePlugin } from 'tdesign-vue-next'

// Token 存储在内存中，不持久化（安全要求）
let accessToken: string | null = null
let refreshToken: string | null = null

export const setAccessToken = (token: string | null) => {
  accessToken = token
}

export const getAccessToken = () => accessToken

export const setRefreshToken = (token: string | null) => {
  refreshToken = token
}

export const getRefreshToken = () => refreshToken

// =============================================================================
// MX-002: 统一错误处理 - 友好提示映射
// =============================================================================

// HTTP 状态码 → 用户友好提示
const HTTP_ERROR_MESSAGES: Record<number, string> = {
  400: '请求参数有误，请检查输入',
  401: '登录已过期，请重新登录',
  403: '没有权限执行此操作',
  404: '请求的资源不存在',
  409: '操作冲突，可能已被其他用户修改',
  422: '数据验证失败，请检查输入',
  429: '请求过于频繁，请稍后重试',
  500: '服务器异常，请稍后重试',
  502: '网关错误，请稍后重试',
  503: '服务暂不可用，请稍后重试',
  504: '请求超时，请稍后重试',
}

// 业务错误码 → 用户友好提示（根据后端 model/response.go）
const BIZ_ERROR_MESSAGES: Record<number, string> = {
  40101: 'Token 无效，请重新登录',
  40102: 'Token 已过期，请重新登录',
  40103: 'Token 已被撤销',
  40301: '没有权限访问此资源',
  40302: 'Token 权限不足',
  40401: '资源不存在',
  40901: '资源已存在，请勿重复创建',
  42201: '数据验证失败',
  50001: '服务器内部错误',
}

// 根据 HTTP 状态码返回友好提示
export function getFriendlyError(error: any): string {
  // 优先检查业务错误码
  const bizCode = error?.response?.data?.code
  if (bizCode && BIZ_ERROR_MESSAGES[bizCode]) {
    return BIZ_ERROR_MESSAGES[bizCode]
  }

  // HTTP 状态码
  const status = error?.response?.status
  if (status && HTTP_ERROR_MESSAGES[status]) {
    return HTTP_ERROR_MESSAGES[status]
  }

  // 网络错误
  if (error?.code === 'ERR_NETWORK' || error?.message?.includes('Network Error')) {
    return '网络连接失败，请检查网络'
  }

  // 超时
  if (error?.code === 'ECONNABORTED') {
    return '请求超时，请稍后重试'
  }

  // 后端返回的错误消息
  if (error?.response?.data?.msg) {
    return error.response.data.msg
  }

  return error?.message || '操作失败，请稍后重试'
}

// 全局错误 Toast
export function showError(error: any) {
  MessagePlugin.error(getFriendlyError(error))
}

// 全局成功 Toast
export function showSuccess(message: string) {
  MessagePlugin.success(message)
}

// =============================================================================
// Axios 实例配置
// =============================================================================

const request: AxiosInstance = axios.create({
  baseURL: '/admin/v1',
  timeout: 30000,
  withCredentials: true, // 携带 HttpOnly Cookie (refresh_token)
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}.${refreshToken}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
request.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const status = error.response?.status
    const config = error.config as AxiosRequestConfig & { _retry?: boolean }

    if (status === 401 && !config._retry) {
      config._retry = true
      // Token 过期，尝试刷新
      try {
        const refreshRes = await axios.post(
          '/admin/v1/auth/refresh',
          {},
          {
            withCredentials: true,
            headers: accessToken ? { Authorization: `Bearer ${accessToken}.${refreshToken}` } : {}
          }
        )
        if (refreshRes.data.code === 0) {
          // 刷新成功，更新内存中的 token
          const newToken = refreshRes.data.data.access_token
          const tokenParts = newToken.split('.')
          accessToken = tokenParts[0] + '.' + tokenParts[1]
          refreshToken = tokenParts[2]

          // 重试原请求
          config.headers!.Authorization = `Bearer ${accessToken}.${refreshToken}`
          return request(config)
        }
      } catch {
        // 刷新失败，清除 token，跳转登录
        accessToken = null
        refreshToken = null
        MessagePlugin.warning('登录已过期，请重新登录')
        window.location.href = '/login'
      }
    } else if (status === 403) {
      MessagePlugin.error('权限不足，无法执行此操作')
    } else if (status === 404) {
      MessagePlugin.error('请求的资源不存在')
    } else if (status === 422) {
      const msg = error.response?.data?.msg || '数据验证失败，请检查输入'
      MessagePlugin.error(msg)
    } else if (status === 429) {
      MessagePlugin.warning('请求过于频繁，请稍后重试')
    } else if (status && status >= 500) {
      MessagePlugin.error('服务器异常，请稍后重试')
    } else if (error.code === 'ERR_NETWORK') {
      MessagePlugin.error('网络连接失败，请检查网络')
    }
    return Promise.reject(error)
  }
)

export default request
