/**
 * 真實 WebSocket 服務
 * 連接後端 ws/chat/:stream_id，處理即時彈幕與觀看人數
 */

import type { ChatMessage } from './mockWebSocket'

type MessageHandler = (msg: ChatMessage) => void
type ViewerCountHandler = (count: number) => void

/** 後端下行訊息格式 */
interface ServerMessage {
  type: 'chat' | 'system'
  user_id?: string
  nickname: string
  content: string
  timestamp: number      // unix seconds
  viewer_count?: number
}

let messageIdCounter = 0

export class RealWebSocket {
  private socket: WebSocket | null = null
  private messageHandlers: MessageHandler[] = []
  private viewerCountHandlers: ViewerCountHandler[] = []
  private connected = false
  private _viewerCount = 0
  private streamId = ''
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private manualClose = false

  get viewerCount() {
    return this._viewerCount
  }

  connect(streamId: string) {
    if (this.connected) return
    this.streamId = streamId
    this.manualClose = false
    this.doConnect()
  }

  disconnect() {
    this.manualClose = true
    this.connected = false
    this.reconnectAttempts = 0
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.socket) {
      this.socket.close()
      this.socket = null
    }
  }

  onMessage(handler: MessageHandler) {
    this.messageHandlers.push(handler)
  }

  onViewerCountChange(handler: ViewerCountHandler) {
    this.viewerCountHandlers.push(handler)
  }

  /**
   * 發送聊天訊息到後端
   */
  sendMessage(content: string) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return
    this.socket.send(JSON.stringify({ type: 'chat', content }))
  }

  private doConnect() {
    const token = localStorage.getItem('access_token')
    if (!token) {
      console.warn('[WS] 無 token，無法連線')
      return
    }

    const wsBase = import.meta.env.VITE_WS_BASE || 'ws://localhost:8081'
    const url = `${wsBase}/ws/chat/${this.streamId}?token=${encodeURIComponent(token)}`

    try {
      this.socket = new WebSocket(url)
    } catch (e) {
      console.error('[WS] 建立連線失敗:', e)
      this.scheduleReconnect()
      return
    }

    this.socket.onopen = () => {
      console.log('[WS] 已連線:', this.streamId)
      this.connected = true
      this.reconnectAttempts = 0
    }

    this.socket.onmessage = (event) => {
      try {
        const raw: ServerMessage = JSON.parse(event.data)
        const msg: ChatMessage = {
          id: `msg_${Date.now()}_${++messageIdCounter}`,
          type: raw.type,
          userId: raw.user_id,
          nickname: raw.nickname,
          content: raw.content,
          timestamp: raw.timestamp * 1000, // 後端用秒，前端用毫秒
        }
        this.messageHandlers.forEach(h => h(msg))

        // 系統訊息帶有 viewer_count
        if (raw.viewer_count != null) {
          this._viewerCount = raw.viewer_count
          this.viewerCountHandlers.forEach(h => h(this._viewerCount))
        }
      } catch (e) {
        console.warn('[WS] 訊息解析失敗:', e)
      }
    }

    this.socket.onclose = () => {
      console.log('[WS] 連線關閉')
      this.connected = false
      this.socket = null
      if (!this.manualClose) {
        this.scheduleReconnect()
      }
    }

    this.socket.onerror = (e) => {
      console.error('[WS] 錯誤:', e)
    }
  }

  private scheduleReconnect() {
    if (this.manualClose) return
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn('[WS] 已達最大重連次數，停止重連')
      return
    }
    // 指數退避：1s, 2s, 4s, 8s... 最大 30s
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000)
    this.reconnectAttempts++
    console.log(`[WS] ${delay / 1000}s 後嘗試重連 (第 ${this.reconnectAttempts} 次)`)
    this.reconnectTimer = setTimeout(() => {
      this.doConnect()
    }, delay)
  }
}
