import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request, { setAccessToken, getAccessToken, setRefreshToken } from '@/utils/request'

interface User {
  id: string
  email: string
  nickname?: string
  avatar_url?: string
  status: string
  is_super_admin: boolean
  created_at: string
}

interface LoginResponse {
  user: User
  access_token: string
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
      const res = await request.post<LoginResponse>('/auth/login', {
        email,
        password,
      })
      if (res.data.code === 0) {
        // 解析 token (格式: accessToken.refreshToken)
        const tokenParts = res.data.data.access_token.split('.')
        const accessToken = tokenParts[0] + '.' + tokenParts[1]
        const refreshToken = tokenParts[2]

        setAccessToken(accessToken)
        setRefreshToken(refreshToken)
        setUser(res.data.data.user)
        return { success: true }
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
      if (res.data.code === 0) {
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
    setAccessToken(null)
    setRefreshToken(null)
    clearUser()
  }

  const fetchUser = async () => {
    if (!getAccessToken()) return

    try {
      const res = await request.get<{ user: User }>('/users/me')
      if (res.data.code === 0) {
        setUser(res.data.data.user)
      }
    } catch {
      // token 可能已过期
      logout()
    }
  }

  const listUsers = async (page = 1, pageSize = 20) => {
    const res = await request.get<{
      data: {
        data: User[]
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
  }
})
