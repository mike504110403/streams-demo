/**
 * Mock API 層 — 前端 Demo 用，不依賴後端即可展示完整流程
 *
 * 啟用方式：VITE_MOCK_API=true（預設啟用，直到後端就緒）
 */

import type { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import { mockStreams, type StreamItem } from './mockData'

const MOCK_DELAY = 600

interface MockUser {
  id: string
  phone: string | null
  nickname: string
  avatar_url: string | null
  bio: string | null
  created_at: string
}

// 模擬資料庫
const mockDB = {
  users: new Map<string, MockUser>(),
  codes: new Map<string, { code: string; expiresAt: number }>(),
  oauthProviders: new Map<string, string>(), // provider:providerId → userId
}

function generateId(): string {
  return crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2)
}

function generateTokens() {
  return {
    access_token: 'mock_access_' + Date.now(),
    refresh_token: 'mock_refresh_' + Date.now(),
  }
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function mockResponse(status: number, data: unknown): AxiosResponse {
  return {
    data,
    status,
    statusText: status === 200 ? 'OK' : status === 201 ? 'Created' : 'Error',
    headers: {},
    config: {} as InternalAxiosRequestConfig,
  }
}

// 路由處理器
type Handler = (config: InternalAxiosRequestConfig) => Promise<AxiosResponse>

const handlers: Array<{ method: string; pattern: RegExp; handler: Handler }> = [
  // POST /auth/send-code
  {
    method: 'post',
    pattern: /\/auth\/send-code$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const body = JSON.parse(config.data || '{}')
      const phone = body.phone

      // 檢查 60 秒限制
      const existing = mockDB.codes.get(phone)
      if (existing && existing.expiresAt > Date.now() && (existing.expiresAt - Date.now()) > 240000) {
        const retryAfter = Math.ceil((existing.expiresAt - 240000 - Date.now()) / 1000)
        throw { response: mockResponse(429, { error: `請等待 ${retryAfter} 秒後再試`, retry_after: retryAfter }) }
      }

      // 儲存驗證碼
      mockDB.codes.set(phone, { code: '123456', expiresAt: Date.now() + 300000 })
      console.log(`[MOCK SMS] Phone: ${phone}, Code: 123456`)

      return mockResponse(200, { data: { message: '驗證碼已發送', expires_in: 300 } })
    },
  },
  // POST /auth/login
  {
    method: 'post',
    pattern: /\/auth\/login$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const body = JSON.parse(config.data || '{}')
      const { phone, code } = body

      // 驗證碼檢查
      const stored = mockDB.codes.get(phone)
      if (!stored || stored.expiresAt < Date.now()) {
        throw { response: mockResponse(400, { message: '驗證碼已過期，請重新發送' }) }
      }
      if (stored.code !== code) {
        throw { response: mockResponse(400, { message: '驗證碼錯誤，請重新輸入' }) }
      }

      // 查詢用戶
      let user: MockUser | undefined
      for (const u of mockDB.users.values()) {
        if (u.phone === phone) { user = u; break }
      }
      if (!user) {
        throw { response: mockResponse(404, { message: '此手機號碼尚未註冊' }) }
      }

      // 標記已使用
      mockDB.codes.delete(phone)

      return mockResponse(200, {
        data: { ...generateTokens(), user },
      })
    },
  },
  // POST /auth/register
  {
    method: 'post',
    pattern: /\/auth\/register$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const body = JSON.parse(config.data || '{}')
      const { phone, code, nickname } = body

      // 驗證碼檢查
      const stored = mockDB.codes.get(phone)
      if (!stored || stored.expiresAt < Date.now()) {
        throw { response: mockResponse(400, { message: '驗證碼已過期，請重新發送' }) }
      }
      if (stored.code !== code) {
        throw { response: mockResponse(400, { message: '驗證碼錯誤，請重新輸入' }) }
      }

      // 檢查手機號碼是否已註冊
      for (const u of mockDB.users.values()) {
        if (u.phone === phone) {
          throw { response: mockResponse(409, { message: '此手機號碼已被註冊' }) }
        }
      }

      mockDB.codes.delete(phone)

      const user: MockUser = {
        id: generateId(),
        phone,
        nickname,
        avatar_url: null,
        bio: null,
        created_at: new Date().toISOString(),
      }
      mockDB.users.set(user.id, user)

      return mockResponse(201, {
        data: { ...generateTokens(), user },
      })
    },
  },
  // POST /auth/oauth/apple
  {
    method: 'post',
    pattern: /\/auth\/oauth\/apple$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const body = JSON.parse(config.data || '{}')
      const providerKey = `apple:mock_apple_sub_${body.id_token?.slice(0, 8) || 'default'}`

      const existingUserId = mockDB.oauthProviders.get(providerKey)
      if (existingUserId) {
        const user = mockDB.users.get(existingUserId)
        return mockResponse(200, {
          data: { ...generateTokens(), is_new_user: false, user },
        })
      }

      const user: MockUser = {
        id: generateId(),
        phone: null,
        nickname: body.user?.name || 'Apple 用戶',
        avatar_url: null,
        bio: null,
        created_at: new Date().toISOString(),
      }
      mockDB.users.set(user.id, user)
      mockDB.oauthProviders.set(providerKey, user.id)

      return mockResponse(201, {
        data: { ...generateTokens(), is_new_user: true, user },
      })
    },
  },
  // POST /auth/oauth/google
  {
    method: 'post',
    pattern: /\/auth\/oauth\/google$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const body = JSON.parse(config.data || '{}')
      const providerKey = `google:mock_google_sub_${body.credential?.slice(0, 8) || 'default'}`

      const existingUserId = mockDB.oauthProviders.get(providerKey)
      if (existingUserId) {
        const user = mockDB.users.get(existingUserId)
        return mockResponse(200, {
          data: { ...generateTokens(), is_new_user: false, user },
        })
      }

      const user: MockUser = {
        id: generateId(),
        phone: null,
        nickname: 'Google 用戶',
        avatar_url: 'https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg',
        bio: null,
        created_at: new Date().toISOString(),
      }
      mockDB.users.set(user.id, user)
      mockDB.oauthProviders.set(providerKey, user.id)

      return mockResponse(201, {
        data: { ...generateTokens(), is_new_user: true, user },
      })
    },
  },
  // POST /auth/refresh
  {
    method: 'post',
    pattern: /\/auth\/refresh$/,
    handler: async () => {
      await delay(200)
      return mockResponse(200, {
        data: generateTokens(),
      })
    },
  },
  // POST /auth/logout
  {
    method: 'post',
    pattern: /\/auth\/logout$/,
    handler: async () => {
      await delay(200)
      return mockResponse(200, { data: { message: 'ok' } })
    },
  },
  // GET /users/me
  {
    method: 'get',
    pattern: /\/users\/me$/,
    handler: async () => {
      await delay(300)
      const storedUser = localStorage.getItem('user')
      if (storedUser) {
        const user = JSON.parse(storedUser)
        // 確保 mockDB 也有這個用戶
        if (!mockDB.users.has(user.id)) {
          mockDB.users.set(user.id, user)
        }
        return mockResponse(200, { data: user })
      }
      throw { response: mockResponse(401, { message: 'Unauthorized' }) }
    },
  },
  // PUT /users/me
  {
    method: 'put',
    pattern: /\/users\/me$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const body = JSON.parse(config.data || '{}')
      const storedUser = localStorage.getItem('user')
      if (!storedUser) {
        throw { response: mockResponse(401, { message: 'Unauthorized' }) }
      }

      const user = JSON.parse(storedUser)
      if (body.nickname !== undefined) user.nickname = body.nickname
      if (body.bio !== undefined) user.bio = body.bio
      if (body.avatar_url !== undefined) user.avatar_url = body.avatar_url

      mockDB.users.set(user.id, user)
      return mockResponse(200, { data: user })
    },
  },
  // ========== 直播相關 ==========
  // GET /streams — 直播列表
  {
    method: 'get',
    pattern: /\/streams(\?|$)/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const url = config.url || ''
      const params = new URLSearchParams(url.split('?')[1] || '')
      const page = parseInt(params.get('page') || '1', 10)
      const limit = parseInt(params.get('limit') || '20', 10)

      // 按 viewer_count DESC 排序
      const sorted = [...mockStreams].sort((a, b) => b.viewer_count - a.viewer_count)
      const start = (page - 1) * limit
      const end = start + limit
      const streams = sorted.slice(start, end)

      return mockResponse(200, {
        data: {
          streams,
          total: mockStreams.length,
          page,
          limit,
        },
      })
    },
  },
  // POST /streams — 建立直播間
  {
    method: 'post',
    pattern: /\/streams$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const storedUser = localStorage.getItem('user')
      if (!storedUser) {
        throw { response: mockResponse(401, { message: 'Unauthorized' }) }
      }
      const user = JSON.parse(storedUser)
      const body = JSON.parse(config.data || '{}')
      const streamId = generateId()
      const streamKey = 'sk_' + generateId()

      const newStream = {
        id: streamId,
        user_id: user.id,
        title: body.title || '未命名直播',
        cover_url: body.cover_url || null,
        stream_key: streamKey,
        status: 'pending' as const,
        viewer_count: 0,
        host_nickname: user.nickname,
        host_avatar: user.avatar_url,
        created_at: new Date().toISOString(),
        rtmp_url: `rtmp://localhost:1935/live/${streamKey}`,
        flv_url: `http://localhost:8080/live/${streamKey}.flv`,
        hls_url: `http://localhost:8080/live/${streamKey}.m3u8`,
      }

      return mockResponse(201, { data: newStream })
    },
  },
  // GET /streams/:id — 直播詳情
  {
    method: 'get',
    pattern: /\/streams\/([^/?]+)$/,
    handler: async (config) => {
      await delay(300)
      const url = config.url || ''
      const match = url.match(/\/streams\/([^/?]+)$/)
      const id = match ? match[1] : ''

      const stream = mockStreams.find((s) => s.id === id)
      if (!stream) {
        throw { response: mockResponse(404, { message: '直播間不存在' }) }
      }

      const streamKey = 'sk_mock_' + stream.id
      return mockResponse(200, {
        data: {
          ...stream,
          user_id: 'mock_user_' + stream.id,
          created_at: new Date().toISOString(),
          flv_url: `http://localhost:8080/live/${streamKey}.flv`,
          hls_url: `http://localhost:8080/live/${streamKey}.m3u8`,
        },
      })
    },
  },
  // PUT /streams/:id — 更新直播
  {
    method: 'put',
    pattern: /\/streams\/([^/?]+)$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const url = config.url || ''
      const match = url.match(/\/streams\/([^/?]+)$/)
      const id = match ? match[1] : ''
      const body = JSON.parse(config.data || '{}')

      const stream = mockStreams.find((s) => s.id === id)
      if (!stream) {
        throw { response: mockResponse(404, { message: '直播間不存在' }) }
      }

      const updated = {
        ...stream,
        title: body.title !== undefined ? body.title : stream.title,
        cover_url: body.cover_url !== undefined ? body.cover_url : stream.cover_url,
        user_id: 'mock_user_' + stream.id,
        created_at: new Date().toISOString(),
      }

      return mockResponse(200, { data: updated })
    },
  },
  // DELETE /streams/:id — 結束直播
  {
    method: 'delete',
    pattern: /\/streams\/([^/?]+)$/,
    handler: async (config) => {
      await delay(MOCK_DELAY)
      const url = config.url || ''
      const match = url.match(/\/streams\/([^/?]+)$/)
      const id = match ? match[1] : ''

      const stream = mockStreams.find((s) => s.id === id)
      if (!stream) {
        throw { response: mockResponse(404, { message: '直播間不存在' }) }
      }

      return mockResponse(200, { data: { message: '直播已結束' } })
    },
  },
]

export function setupMockApi(api: AxiosInstance): void {
  api.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
    const method = (config.method || 'get').toLowerCase()
    const url = config.url || ''

    for (const route of handlers) {
      if (route.method === method && route.pattern.test(url)) {
        // 找到匹配的 mock handler
        try {
          const response = await route.handler(config)
          // 透過 adapter 返回 mock 結果
          config.adapter = () => Promise.resolve(response)
          return config
        } catch (err) {
          config.adapter = () => Promise.reject(err)
          return config
        }
      }
    }

    // 沒有匹配到的請求，正常發送（或在純 mock 模式下返回 404）
    config.adapter = () => Promise.reject({
      response: mockResponse(404, { message: `[Mock] 未實作的端點: ${method.toUpperCase()} ${url}` }),
    })
    return config
  })
}
