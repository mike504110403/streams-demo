/**
 * Mock WebSocket 服務
 * 模擬直播間的即時訊息（彈幕、系統通知、觀看人數）
 */

export interface ChatMessage {
  id: string
  type: 'chat' | 'system'
  userId?: string
  nickname: string
  avatar?: string
  content: string
  timestamp: number
}

type MessageHandler = (msg: ChatMessage) => void
type ViewerCountHandler = (count: number) => void

// 假使用者列表
const MOCK_USERS = [
  { nickname: '小明', avatar: 'https://i.pravatar.cc/100?u=1' },
  { nickname: '花花', avatar: 'https://i.pravatar.cc/100?u=2' },
  { nickname: '阿傑', avatar: 'https://i.pravatar.cc/100?u=3' },
  { nickname: '小美', avatar: 'https://i.pravatar.cc/100?u=4' },
  { nickname: '大雄', avatar: 'https://i.pravatar.cc/100?u=5' },
  { nickname: '靜香', avatar: 'https://i.pravatar.cc/100?u=6' },
  { nickname: '胖虎', avatar: 'https://i.pravatar.cc/100?u=7' },
  { nickname: '小夫', avatar: 'https://i.pravatar.cc/100?u=8' },
  { nickname: '天天開心', avatar: 'https://i.pravatar.cc/100?u=9' },
  { nickname: '夜貓子', avatar: 'https://i.pravatar.cc/100?u=10' },
  { nickname: '追劇達人', avatar: 'https://i.pravatar.cc/100?u=11' },
  { nickname: '吃貨一號', avatar: 'https://i.pravatar.cc/100?u=12' },
]

const MOCK_MESSAGES = [
  '好厲害！',
  '太強了吧',
  '666',
  '主播加油！',
  '哈哈哈哈',
  '來了來了',
  '第一次來',
  '主播唱首歌',
  '好好看',
  '支持主播',
  '❤️❤️❤️',
  '好帥',
  '太有趣了',
  '笑死',
  '主播在哪個城市？',
  '晚安',
  '明天還直播嗎？',
  '已關注',
  '送禮物！',
  '這也太強了',
]

let messageIdCounter = 0

function generateId(): string {
  return `msg_${Date.now()}_${++messageIdCounter}`
}

function randomItem<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)]
}

function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

export class MockWebSocket {
  private messageHandlers: MessageHandler[] = []
  private viewerCountHandlers: ViewerCountHandler[] = []
  private chatTimer: ReturnType<typeof setTimeout> | null = null
  private systemTimer: ReturnType<typeof setTimeout> | null = null
  private viewerTimer: ReturnType<typeof setInterval> | null = null
  private connected = false
  private _viewerCount = 0

  get viewerCount() {
    return this._viewerCount
  }

  connect(initialViewerCount = 1234) {
    if (this.connected) return
    this.connected = true
    this._viewerCount = initialViewerCount

    // 開始產生隨機彈幕
    this.scheduleNextChat()
    // 開始產生系統通知
    this.scheduleNextSystemNotice()
    // 開始觀看人數波動
    this.viewerTimer = setInterval(() => {
      const delta = randomInt(-5, 8)
      this._viewerCount = Math.max(1, this._viewerCount + delta)
      this.viewerCountHandlers.forEach(h => h(this._viewerCount))
    }, 3000)
  }

  disconnect() {
    this.connected = false
    if (this.chatTimer) {
      clearTimeout(this.chatTimer)
      this.chatTimer = null
    }
    if (this.systemTimer) {
      clearTimeout(this.systemTimer)
      this.systemTimer = null
    }
    if (this.viewerTimer) {
      clearInterval(this.viewerTimer)
      this.viewerTimer = null
    }
  }

  onMessage(handler: MessageHandler) {
    this.messageHandlers.push(handler)
  }

  onViewerCountChange(handler: ViewerCountHandler) {
    this.viewerCountHandlers.push(handler)
  }

  /**
   * 發送訊息（本地回顯）
   */
  sendMessage(nickname: string, content: string, avatar?: string): ChatMessage {
    const msg: ChatMessage = {
      id: generateId(),
      type: 'chat',
      nickname,
      avatar,
      content,
      timestamp: Date.now(),
    }
    this.messageHandlers.forEach(h => h(msg))
    return msg
  }

  private scheduleNextChat() {
    if (!this.connected) return
    const delay = randomInt(800, 2500)
    this.chatTimer = setTimeout(() => {
      if (!this.connected) return
      const user = randomItem(MOCK_USERS)
      const msg: ChatMessage = {
        id: generateId(),
        type: 'chat',
        nickname: user.nickname,
        avatar: user.avatar,
        content: randomItem(MOCK_MESSAGES),
        timestamp: Date.now(),
      }
      this.messageHandlers.forEach(h => h(msg))
      this.scheduleNextChat()
    }, delay)
  }

  private scheduleNextSystemNotice() {
    if (!this.connected) return
    const delay = randomInt(3000, 7000)
    this.systemTimer = setTimeout(() => {
      if (!this.connected) return
      const user = randomItem(MOCK_USERS)
      const msg: ChatMessage = {
        id: generateId(),
        type: 'system',
        nickname: user.nickname,
        content: `${user.nickname} 進入直播間`,
        timestamp: Date.now(),
      }
      this.messageHandlers.forEach(h => h(msg))
      this.scheduleNextSystemNotice()
    }, delay)
  }
}
