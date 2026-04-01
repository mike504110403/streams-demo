import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../services/api'
import router from '../router'

export interface User {
  id: string
  phone: string
  nickname: string
  avatar_url: string
  bio: string
  created_at: string
}

export interface SendCodePayload {
  phone: string
}

export interface LoginPayload {
  phone: string
  code: string
}

export interface RegisterPayload {
  phone: string
  code: string
  nickname: string
}

export interface AppleOAuthPayload {
  code: string
  id_token: string
  user?: { name?: string; email?: string }
}

export interface GoogleOAuthPayload {
  credential: string
}

export interface UpdateProfilePayload {
  nickname?: string
  bio?: string
  avatar_url?: string
}

export const useAuthStore = defineStore('auth', () => {
  // === State ===
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const refreshTokenValue = ref<string | null>(null)

  // === Getters ===
  const isAuthenticated = computed(() => !!accessToken.value)

  // === 初始化：從 localStorage 恢復 ===
  function initialize() {
    const storedToken = localStorage.getItem('access_token')
    const storedRefresh = localStorage.getItem('refresh_token')
    const storedUser = localStorage.getItem('user')

    if (storedToken) {
      accessToken.value = storedToken
    }
    if (storedRefresh) {
      refreshTokenValue.value = storedRefresh
    }
    if (storedUser) {
      try {
        user.value = JSON.parse(storedUser)
      } catch {
        localStorage.removeItem('user')
      }
    }
  }

  // === Actions ===

  async function sendCode(payload: SendCodePayload) {
    const { data } = await api.post('/auth/send-code', payload)
    return data.data
  }

  async function login(payload: LoginPayload) {
    const { data } = await api.post('/auth/login', payload)
    const result = data.data

    accessToken.value = result.access_token
    refreshTokenValue.value = result.refresh_token
    user.value = result.user

    localStorage.setItem('access_token', result.access_token)
    localStorage.setItem('refresh_token', result.refresh_token)
    localStorage.setItem('user', JSON.stringify(result.user))
  }

  async function register(payload: RegisterPayload) {
    const { data } = await api.post('/auth/register', payload)
    const result = data.data

    accessToken.value = result.access_token
    refreshTokenValue.value = result.refresh_token
    user.value = result.user

    localStorage.setItem('access_token', result.access_token)
    localStorage.setItem('refresh_token', result.refresh_token)
    localStorage.setItem('user', JSON.stringify(result.user))
  }

  async function oauthApple(payload: AppleOAuthPayload) {
    const { data } = await api.post('/auth/oauth/apple', payload)
    const result = data.data

    accessToken.value = result.access_token
    refreshTokenValue.value = result.refresh_token
    user.value = result.user

    localStorage.setItem('access_token', result.access_token)
    localStorage.setItem('refresh_token', result.refresh_token)
    localStorage.setItem('user', JSON.stringify(result.user))

    return result
  }

  async function oauthGoogle(payload: GoogleOAuthPayload) {
    const { data } = await api.post('/auth/oauth/google', payload)
    const result = data.data

    accessToken.value = result.access_token
    refreshTokenValue.value = result.refresh_token
    user.value = result.user

    localStorage.setItem('access_token', result.access_token)
    localStorage.setItem('refresh_token', result.refresh_token)
    localStorage.setItem('user', JSON.stringify(result.user))

    return result
  }

  async function logout() {
    try {
      await api.post('/auth/logout')
    } catch {
      // 即使 API 失敗也要清除本地狀態
    }

    accessToken.value = null
    refreshTokenValue.value = null
    user.value = null

    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')

    router.push('/login')
  }

  async function refreshToken() {
    if (!refreshTokenValue.value) {
      throw new Error('No refresh token available')
    }

    const { data } = await api.post('/auth/refresh', {
      refresh_token: refreshTokenValue.value,
    })
    const result = data.data

    accessToken.value = result.access_token
    if (result.refresh_token) {
      refreshTokenValue.value = result.refresh_token
      localStorage.setItem('refresh_token', result.refresh_token)
    }
    localStorage.setItem('access_token', result.access_token)
  }

  async function fetchProfile() {
    const { data } = await api.get('/users/me')
    user.value = data.data
    localStorage.setItem('user', JSON.stringify(data.data))
  }

  async function updateProfile(payload: UpdateProfilePayload) {
    const { data } = await api.put('/users/me', payload)
    user.value = data.data
    localStorage.setItem('user', JSON.stringify(data.data))
  }

  // 初始化
  initialize()

  return {
    // State
    user,
    accessToken,
    refreshTokenValue,
    // Getters
    isAuthenticated,
    // Actions
    sendCode,
    login,
    register,
    oauthApple,
    oauthGoogle,
    logout,
    refreshToken,
    fetchProfile,
    updateProfile,
    initialize,
  }
})
