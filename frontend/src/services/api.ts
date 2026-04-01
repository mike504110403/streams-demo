import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import { setupMockApi } from './mock'

const MOCK_MODE = import.meta.env.VITE_MOCK_API !== 'false'
const DEBUG_MODE = import.meta.env.VITE_DEBUG_MODE !== 'false'

const api: AxiosInstance = axios.create({
  baseURL: 'http://localhost:8081/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Mock 模式啟用（預設開啟，直到後端就緒後設定 VITE_MOCK_API=false）
if (MOCK_MODE) {
  console.log('[API] Mock 模式已啟用 — 所有請求由前端模擬回應')
  setupMockApi(api)
}
if (DEBUG_MODE) {
  console.log('[API] Debug 模式已啟用 — 驗證碼可隨意輸入')
}

// 是否正在刷新 Token
let isRefreshing = false
// 等待刷新的請求佇列
let failedQueue: Array<{
  resolve: (value: unknown) => void
  reject: (reason: unknown) => void
}> = []

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

// Request Interceptor：自動加上 Authorization header
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response Interceptor：401 時自動用 refresh token 刷新
api.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error) => {
    const originalRequest = error.config

    // 如果不是 401 錯誤，直接拋出
    if (!error.response || error.response.status !== 401) {
      return Promise.reject(error)
    }

    // 如果是刷新 Token 的請求本身失敗了，清除登入狀態
    if (originalRequest.url?.includes('/auth/refresh')) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('user')
      window.location.href = '/login'
      return Promise.reject(error)
    }

    // 防止重複刷新
    if (originalRequest._retry) {
      return Promise.reject(error)
    }

    if (isRefreshing) {
      // 如果正在刷新，把請求放入佇列等待
      return new Promise((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      }).then((token) => {
        originalRequest.headers.Authorization = `Bearer ${token}`
        return api(originalRequest)
      })
    }

    originalRequest._retry = true
    isRefreshing = true

    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('user')
      window.location.href = '/login'
      return Promise.reject(error)
    }

    try {
      const { data } = await axios.post(
        'http://localhost:8081/api/v1/auth/refresh',
        { refresh_token: refreshToken }
      )

      const newAccessToken = data.data.access_token
      const newRefreshToken = data.data.refresh_token

      localStorage.setItem('access_token', newAccessToken)
      if (newRefreshToken) {
        localStorage.setItem('refresh_token', newRefreshToken)
      }

      originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
      processQueue(null, newAccessToken)

      return api(originalRequest)
    } catch (refreshError) {
      processQueue(refreshError, null)
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('user')
      window.location.href = '/login'
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)

export default api
