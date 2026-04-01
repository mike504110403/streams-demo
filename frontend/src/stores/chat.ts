import { defineStore } from 'pinia'
import { ref } from 'vue'
import { MockWebSocket, type ChatMessage } from '../services/mockWebSocket'

const MAX_MESSAGES = 200

export const useChatStore = defineStore('chat', () => {
  // === State ===
  const messages = ref<ChatMessage[]>([])
  const viewerCount = ref(0)
  const connected = ref(false)

  // === Internal ===
  let ws: MockWebSocket | null = null

  // === Actions ===

  function connect(initialViewerCount?: number) {
    if (connected.value) return

    ws = new MockWebSocket()

    ws.onMessage((msg) => {
      messages.value.push(msg)
      // 限制訊息數量，避免記憶體膨脹
      if (messages.value.length > MAX_MESSAGES) {
        messages.value = messages.value.slice(-100)
      }
    })

    ws.onViewerCountChange((count) => {
      viewerCount.value = count
    })

    ws.connect(initialViewerCount)
    viewerCount.value = ws.viewerCount
    connected.value = true
  }

  function disconnect() {
    if (ws) {
      ws.disconnect()
      ws = null
    }
    connected.value = false
    messages.value = []
  }

  function sendMessage(content: string) {
    if (!ws || !content.trim()) return

    // 從 localStorage 取得目前使用者資訊
    const userStr = localStorage.getItem('user')
    let nickname = '我'
    let avatar: string | undefined
    if (userStr) {
      try {
        const user = JSON.parse(userStr)
        nickname = user.nickname || '我'
        avatar = user.avatar_url
      } catch {
        // ignore
      }
    }

    ws.sendMessage(nickname, content.trim(), avatar)
  }

  return {
    // State
    messages,
    viewerCount,
    connected,
    // Actions
    connect,
    disconnect,
    sendMessage,
  }
})
