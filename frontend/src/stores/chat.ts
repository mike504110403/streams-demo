import { defineStore } from 'pinia'
import { ref } from 'vue'
import { MockWebSocket, type ChatMessage } from '../services/mockWebSocket'
import { RealWebSocket } from '../services/realWebSocket'

const MAX_MESSAGES = 200
const MOCK_MODE = import.meta.env.VITE_MOCK_API !== 'false'

export type { ChatMessage }

export const useChatStore = defineStore('chat', () => {
  // === State ===
  const messages = ref<ChatMessage[]>([])
  const viewerCount = ref(0)
  const connected = ref(false)

  // === Internal ===
  let mockWs: MockWebSocket | null = null
  let realWs: RealWebSocket | null = null

  // === Actions ===

  function connect(streamId: string, initialViewerCount?: number) {
    if (connected.value) return

    if (MOCK_MODE) {
      // Mock 模式
      mockWs = new MockWebSocket()
      mockWs.onMessage(handleMessage)
      mockWs.onViewerCountChange((count) => { viewerCount.value = count })
      mockWs.connect(initialViewerCount)
      viewerCount.value = mockWs.viewerCount
    } else {
      // 真實模式
      realWs = new RealWebSocket()
      realWs.onMessage(handleMessage)
      realWs.onViewerCountChange((count) => { viewerCount.value = count })
      realWs.connect(streamId)
    }

    connected.value = true
  }

  function disconnect() {
    if (mockWs) {
      mockWs.disconnect()
      mockWs = null
    }
    if (realWs) {
      realWs.disconnect()
      realWs = null
    }
    connected.value = false
    messages.value = []
  }

  function sendMessage(content: string) {
    if (!content.trim()) return

    if (MOCK_MODE && mockWs) {
      const userStr = localStorage.getItem('user')
      let nickname = '我'
      let avatar: string | undefined
      if (userStr) {
        try {
          const user = JSON.parse(userStr)
          nickname = user.nickname || '我'
          avatar = user.avatar_url
        } catch { /* ignore */ }
      }
      mockWs.sendMessage(nickname, content.trim(), avatar)
    } else if (realWs) {
      realWs.sendMessage(content.trim())
    }
  }

  function handleMessage(msg: ChatMessage) {
    messages.value.push(msg)
    if (messages.value.length > MAX_MESSAGES) {
      messages.value = messages.value.slice(-100)
    }
  }

  return {
    messages,
    viewerCount,
    connected,
    connect,
    disconnect,
    sendMessage,
  }
})
