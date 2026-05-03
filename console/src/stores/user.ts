// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request, { setAccessToken, getAccessToken, logout as clearAuthSession, initializeSession } from '@/utils/request'
import { useSiteStore } from '@/stores/site'

interface User {
  id: string
  email: string
  nickname?: string
  avatar_url?: string
  status: string
  is_super_admin: boolean
  mfa_enabled: boolean
  created_time: string
}

interface LoginResponse {
  user: User
  access_token: string
  refresh_token: string
}

interface MFARequiredResponse {
  mfa_required: true
  mfa_token: string
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const isLoading = ref(false)

  const isLoggedIn = computed(() => !!getAccessToken())

  const setUser = (newUser: User) => {
    user.value = newUser
  }

  const clearUser = () => {
    user.value = null
  }

  const login = async (email: string, password: string) => {
    isLoading.value = true
    try {
      const res = await request.post<LoginResponse | MFARequiredResponse>('/auth/login', {
        email,
        password,
      })
      if (res.data.code === 200) {
        const data = res.data.data as any

        // MFA 两步验证：返回 mfa_required
        if (data?.mfa_required === true) {
          return {
            success: true,
            mfa_required: true,
            mfa_token: data.mfa_token,
            email,
          }
        }

        // 正常登录
        const loginData = data as LoginResponse
        // 只存储 AccessToken（内存），RefreshToken 已由后端写入 HttpOnly Cookie
        setAccessToken(loginData.access_token)
        setUser(loginData.user)

        // 登录成功后自动加载站点列表
        const siteStore = useSiteStore()
        await siteStore.fetchSites()

        return { success: true, mfa_required: false }
      }
      return { success: false, message: res.data.msg }
    } catch (error: any) {
      const msg = error.response?.data?.msg || '登录失败'
      return { success: false, message: msg }
    } finally {
      isLoading.value = false
    }
  }

  const register = async (email: string, password: string, nickname?: string) => {
    isLoading.value = true
    try {
      const res = await request.post<User>('/auth/register', {
        email,
        password,
        nickname,
      })
      if (res.data.code === 200) {
        return { success: true, data: res.data.data }
      }
      return { success: false, message: res.data.msg }
    } catch (error: any) {
      const msg = error.response?.data?.msg || '注册失败'
      return { success: false, message: msg }
    } finally {
      isLoading.value = false
    }
  }

  const logout = async () => {
    try {
      await request.post('/auth/logout')
    } catch {
      // ignore
    }
    // 清除内存中的 AccessToken 和 Cookie 中的 RefreshToken
    clearAuthSession()
    clearUser()

    // 登出时清除站点状态
    const siteStore = useSiteStore()
    siteStore.clearSites()
  }

  const fetchUser = async () => {
    // 如果没有 AccessToken，尝试从 Cookie 刷新恢复会话
    if (!getAccessToken()) {
      const restored = await initializeSession()
      if (!restored) {
        // 无法恢复，跳转登录
        logout()
        return
      }
    }

    try {
      const res = await request.get<{ user: User }>('/users/me')
      if (res.data.code === 200) {
        setUser(res.data.data as any)
      }
    } catch {
      // token 可能已过期
      logout()
    }
  }

  const listUsers = async (page = 1, pageSize = 20) => {
    const res = await request.get<{
      data: {
        items: User[]
        total: number
        page: number
        page_size: number
      }
    }>('/users', { params: { page, page_size: pageSize } })
    return res.data.data
  }

  const deleteUser = async (userId: string) => {
    await request.delete(`/users/${userId}`)
  }

  const createUser = async (data: { email: string; password: string; nickname?: string; is_super_admin?: boolean }) => {
    const res = await request.post<{ data: User }>('/users', data)
    return res.data
  }

  const updateUser = async (id: string, data: { nickname?: string; status?: string; is_super_admin?: boolean }) => {
    const res = await request.put<{ data: User }>(`/users/${id}`, data)
    return res.data
  }

  return {
    user,
    isLoading,
    isLoggedIn,
    setUser,
    clearUser,
    login,
    register,
    logout,
    fetchUser,
    listUsers,
    deleteUser,
    createUser,
    updateUser,
  }
})
