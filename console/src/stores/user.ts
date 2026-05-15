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
  permissions?: string[]
}

interface LoginResponse {
  user: User
  access_token: string
  refresh_token: string
  password_expired?: boolean
  password_expire_days?: number
  mfa_setup_required?: boolean
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const permissions = ref<string[]>([])
  const isLoading = ref(false)

  const isLoggedIn = computed(() => !!getAccessToken())
  const isSuperAdmin = computed(() => user.value?.is_super_admin === true)

  const setUser = (newUser: User) => {
    user.value = newUser
  }

  const clearUser = () => {
    user.value = null
    permissions.value = []
  }

  /** 获取当前用户的权限列表 */
  const fetchPermissions = async () => {
    // super_admin 拥有所有权限
    if (user.value?.is_super_admin) {
      permissions.value = []
      return
    }
    try {
      const res = await request.get<{ data: string[] }>('/users/me/permissions')
      const body = res.data as any
      if (body.code === 200 && body.data) {
        permissions.value = body.data
      } else {
        permissions.value = []
      }
    } catch {
      permissions.value = []
    }
  }

  /** 检查用户是否拥有指定权限 */
  const hasPermission = (permission: string): boolean => {
    // super_admin 跳过所有权限检查
    if (isSuperAdmin.value) return true
    return permissions.value.includes(permission)
  }

  /** 检查用户是否拥有任一指定权限 */
  const hasAnyPermission = (perms: string[]): boolean => {
    if (isSuperAdmin.value) return true
    return perms.some(p => permissions.value.includes(p))
  }

  const login = async (email: string, password: string, encryptedPassword?: string, tokenId?: string, rsaToken?: string) => {
    isLoading.value = true
    try {
      const body: any = { email }
      if (encryptedPassword && tokenId && rsaToken) {
        body.encrypted_password = encryptedPassword
        body.token_id = tokenId
        body.rsa_token = rsaToken
      } else {
        body.password = password
      }
      const res = await request.post<any>('/auth/login', body)
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

        // 正常登录（即使密码过期也完成登录）
        const loginData = data as LoginResponse
        // 只存储 AccessToken（内存），RefreshToken 已由后端写入 HttpOnly Cookie
        setAccessToken(loginData.access_token)
        setUser(loginData.user)

        // 登录成功后自动加载站点列表
        const siteStore = useSiteStore()
        await siteStore.fetchSites()

        // 返回密码过期状态
        return {
          success: true,
          mfa_required: false,
          mfa_setup_required: loginData.mfa_setup_required || false,
          password_expired: loginData.password_expired || false,
          password_expire_days: loginData.password_expire_days,
        }
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
      const res = await request.post<any>('/auth/register', {
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
        clearUser()
        clearAuthSession()
        return false
      }
    }

    try {
      const res = await request.get('/users/me')
      const body = res.data as any
      if (body.code === 200 && body.data) {
        setUser(body.data)
        return true
      }
      return false
    } catch {
      return false
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
    permissions,
    isLoading,
    isLoggedIn,
    isSuperAdmin,
    setUser,
    clearUser,
    login,
    register,
    logout,
    fetchUser,
    fetchPermissions,
    hasPermission,
    hasAnyPermission,
    listUsers,
    deleteUser,
    createUser,
    updateUser,
  }
})
