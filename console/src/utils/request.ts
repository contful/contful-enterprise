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
        window.location.href = '/login'
      }
    } else if (status === 403) {
      MessagePlugin.error('权限不足')
    } else if (status === 500) {
      MessagePlugin.error('服务器错误')
    }
    return Promise.reject(error)
  }
)

export default request
