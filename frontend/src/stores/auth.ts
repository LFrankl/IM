import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/user'
import { authApi } from '@/api/auth'
import type { ApiResponse } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string>(localStorage.getItem('token') || '')

  const isLoggedIn = computed(() => !!token.value && !!user.value)

  async function login(username: string, password: string) {
    const res = await authApi.login({ username, password })
    const body = res.data as unknown as ApiResponse<{ token: string; user: User }>
    token.value = body.data.token
    user.value = body.data.user
    localStorage.setItem('token', body.data.token)
  }

  async function register(username: string, password: string, nickname: string) {
    const res = await authApi.register({ username, password, nickname })
    const body = res.data as unknown as ApiResponse<{ token: string; user: User }>
    token.value = body.data.token
    user.value = body.data.user
    localStorage.setItem('token', body.data.token)
  }

  async function fetchMe() {
    const res = await authApi.me()
    const body = res.data as unknown as ApiResponse<User>
    user.value = body.data
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      user.value = null
      token.value = ''
      localStorage.removeItem('token')
    }
  }

  return { user, token, isLoggedIn, login, register, fetchMe, logout }
})
