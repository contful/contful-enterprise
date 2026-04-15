import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request, { setAccessToken, getAccessToken } from '@/utils/request'

interface User {
  id: string
  email: string
  name: string
  role: string
  avatar?: string
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const isLoggedIn = computed(() => !!getAccessToken())

  const setUser = (newUser: User) => {
    user.value = newUser
  }

  const login = async (email: string, password: string) => {
    const res = await request.post<{ user: User; access_token: string }>('/auth/login', {
      email,
      password,
    })
    if (res.data.code === 0) {
      // Token 存储在内存中，不持久化
      setAccessToken(res.data.data.access_token)
      setUser(res.data.data.user)
      return true
    }
    return false
  }

  const logout = async () => {
    await request.post('/auth/logout')
    setAccessToken(null)
    user.value = null
  }

  const fetchUser = async () => {
    const res = await request.get<{ user: User }>('/users/me')
    if (res.data.code === 0) {
      setUser(res.data.data.user)
    }
  }

  return {
    user,
    isLoggedIn,
    setUser,
    login,
    logout,
    fetchUser,
  }
})
