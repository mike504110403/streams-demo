<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import mpegts from 'mpegts.js'
import { useStreamStore } from '../stores/stream'
import { useChatStore } from '../stores/chat'
import DanmakuOverlay from '../components/danmaku/DanmakuOverlay.vue'
import ChatMessageList from '../components/danmaku/ChatMessageList.vue'
import ChatInput from '../components/danmaku/ChatInput.vue'
import SystemNotice from '../components/danmaku/SystemNotice.vue'

const route = useRoute()
const router = useRouter()
const streamStore = useStreamStore()
const chatStore = useChatStore()

// 播放器相關
const videoRef = ref<HTMLVideoElement | null>(null)
let player: mpegts.Player | null = null
const playerReady = ref(false)
const playerError = ref<string | null>(null)
const playerLoading = ref(true)

// 直播結束覆蓋層
const streamEnded = ref(false)

onMounted(async () => {
  const id = route.params.id as string
  await streamStore.loadStreamById(id)

  // 連線 WebSocket（Mock 或真實），使用當前直播的觀看人數作為初始值
  const initialCount = streamStore.currentStream?.viewer_count ?? 1234
  chatStore.connect(id, initialCount)

  // 初始化播放器
  await nextTick()
  initPlayer()
})

onUnmounted(() => {
  destroyPlayer()
  streamStore.clearCurrentStream()
  chatStore.disconnect()
})

// 監聽系統訊息：偵測「直播已結束」
watch(() => chatStore.messages.length, () => {
  const msgs = chatStore.messages
  if (msgs.length === 0) return

  const lastMsg = msgs[msgs.length - 1]
  if (lastMsg.type === 'system' && lastMsg.content === '直播已結束') {
    handleStreamEnded()
  }
})

/**
 * 初始化串流播放器
 * 策略：支援 MSE → mpegts.js (HTTP-FLV)，不支援 → HLS fallback
 */
function initPlayer() {
  const stream = streamStore.currentStream as any
  if (!stream) {
    playerError.value = '直播資訊載入失敗'
    playerLoading.value = false
    return
  }

  const flvUrl = stream.flv_url
  const hlsUrl = stream.hls_url

  if (!flvUrl && !hlsUrl) {
    playerError.value = '無可用的串流地址'
    playerLoading.value = false
    return
  }

  // 判斷是否支援 FLV（MSE）
  const canUseFLV = mpegts.isSupported() && typeof MediaSource !== 'undefined'

  if (canUseFLV && flvUrl) {
    initFLVPlayer(flvUrl)
  } else if (hlsUrl) {
    initHLSPlayer(hlsUrl)
  } else {
    playerError.value = '您的瀏覽器不支援此串流格式'
    playerLoading.value = false
  }
}

function initFLVPlayer(url: string) {
  if (!videoRef.value) return

  try {
    player = mpegts.createPlayer({
      type: 'flv',
      url: url,
      isLive: true,
    }, {
      enableWorker: true,
      lazyLoadMaxDuration: 3 * 60,
      seekType: 'range',
      liveBufferLatencyChasing: true,
      liveBufferLatencyMaxLatency: 3,
      liveBufferLatencyMinRemain: 0.5,
    })

    player.attachMediaElement(videoRef.value)
    player.load()
    player.play()

    player.on(mpegts.Events.ERROR, (errorType, errorDetail, errorInfo) => {
      console.error('[Player] FLV 錯誤:', errorType, errorDetail, errorInfo)
      playerError.value = '播放失敗，請重試'
      playerLoading.value = false
    })

    player.on(mpegts.Events.LOADING_COMPLETE, () => {
      console.log('[Player] 載入完成')
    })

    // 監聽 video 元素事件
    videoRef.value.addEventListener('playing', () => {
      playerReady.value = true
      playerLoading.value = false
      playerError.value = null
    })

    videoRef.value.addEventListener('waiting', () => {
      playerLoading.value = true
    })

    videoRef.value.addEventListener('error', () => {
      playerError.value = '播放失敗'
      playerLoading.value = false
    })
  } catch (e) {
    console.error('[Player] 初始化失敗:', e)
    playerError.value = '播放器初始化失敗'
    playerLoading.value = false
  }
}

function initHLSPlayer(url: string) {
  if (!videoRef.value) return

  // iOS Safari 原生支援 HLS
  videoRef.value.src = url
  videoRef.value.play().catch(() => {
    // 自動播放可能被瀏覽器阻擋
    playerError.value = '點擊播放'
    playerLoading.value = false
  })

  videoRef.value.addEventListener('playing', () => {
    playerReady.value = true
    playerLoading.value = false
    playerError.value = null
  })

  videoRef.value.addEventListener('error', () => {
    playerError.value = '播放失敗'
    playerLoading.value = false
  })
}

function destroyPlayer() {
  if (player) {
    try {
      player.pause()
      player.unload()
      player.detachMediaElement()
      player.destroy()
    } catch { /* ignore */ }
    player = null
  }
}

function retryPlayer() {
  playerError.value = null
  playerLoading.value = true
  destroyPlayer()
  initPlayer()
}

function handleStreamEnded() {
  streamEnded.value = true
  destroyPlayer()
}

function goBack() {
  router.back()
}

function goHome() {
  router.push('/')
}

function formatViewerCount(count: number): string {
  if (count >= 10000) {
    return (count / 10000).toFixed(1) + '萬'
  }
  if (count >= 1000) {
    return (count / 1000).toFixed(1) + 'k'
  }
  return String(count)
}

function handleSendMessage(content: string) {
  chatStore.sendMessage(content)
}

// ==================== 愛心動畫 ====================
interface FloatingHeart {
  id: number
  color: string
  left: number
  delay: number
  size: number
}

const HEART_COLORS = ['#FF2D55', '#FF6B81', '#FF6348', '#A55EEA', '#FF69B4', '#FF4757', '#FFA502']
let heartIdCounter = 0

const floatingHearts = ref<FloatingHeart[]>([])
const likeCount = ref(0)

function handleLike() {
  likeCount.value++

  const heart: FloatingHeart = {
    id: heartIdCounter++,
    color: HEART_COLORS[Math.floor(Math.random() * HEART_COLORS.length)],
    left: Math.random() * 40 - 20,
    delay: Math.random() * 100,
    size: 24 + Math.random() * 14,
  }

  floatingHearts.value.push(heart)

  setTimeout(() => {
    floatingHearts.value = floatingHearts.value.filter(h => h.id !== heart.id)
  }, 1600)
}

// ==================== 分享面板 ====================
const showShareSheet = ref(false)

const shareActions = [
  { name: '複製連結', icon: 'link-o' },
  { name: '更多分享', icon: 'share-o' },
]

const supportsWebShare = typeof navigator !== 'undefined' && !!navigator.share
const filteredShareActions = supportsWebShare
  ? shareActions
  : shareActions.filter(a => a.name !== '更多分享')

function handleShare() {
  showShareSheet.value = true
}

async function onShareSelect(action: { name: string }) {
  const shareUrl = window.location.href
  const shareTitle = streamStore.currentStream?.title ?? '來看直播'

  if (action.name === '複製連結') {
    try {
      await navigator.clipboard.writeText(shareUrl)
      showToast('已複製連結')
    } catch {
      const input = document.createElement('input')
      input.value = shareUrl
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
      showToast('已複製連結')
    }
  } else if (action.name === '更多分享') {
    try {
      await navigator.share({ title: shareTitle, url: shareUrl })
    } catch {
      // 使用者取消
    }
  }

  showShareSheet.value = false
}
</script>

<template>
  <div class="live-view-page">
    <!-- 播放器區域 -->
    <div class="player-area">
      <!-- 真實串流播放 -->
      <video
        ref="videoRef"
        class="player-video"
        autoplay
        playsinline
        :style="{ display: playerReady && !streamEnded ? 'block' : 'none' }"
      />

      <!-- 載入中 -->
      <div v-if="playerLoading && !streamEnded" class="player-overlay">
        <van-loading size="40" color="#fff" />
        <span class="player-overlay-text">正在連線...</span>
      </div>

      <!-- 播放錯誤 -->
      <div v-if="playerError && !streamEnded" class="player-overlay">
        <van-icon name="warning-o" size="48" color="#ff4757" />
        <span class="player-overlay-text error">{{ playerError }}</span>
        <van-button size="small" round plain color="#fff" @click="retryPlayer">
          重新播放
        </van-button>
      </div>

      <!-- 無串流時的 fallback -->
      <div v-if="!playerLoading && !playerError && !playerReady && !streamEnded" class="player-overlay">
        <van-icon name="video-o" size="48" color="#444" />
        <span class="player-overlay-text" v-if="streamStore.currentStream">
          {{ streamStore.currentStream.title }}
        </span>
        <span class="player-overlay-text" v-else>載入中...</span>
      </div>
    </div>

    <!-- 直播結束覆蓋層 -->
    <Transition name="fade">
      <div v-if="streamEnded" class="stream-ended-overlay">
        <div class="ended-content">
          <van-icon name="video-o" size="64" color="#666" />
          <h2 class="ended-title">直播已結束</h2>
          <p class="ended-subtitle">感謝觀看</p>
          <van-button
            type="primary"
            round
            class="ended-btn"
            @click="goHome"
          >
            回到大廳
          </van-button>
        </div>
      </div>
    </Transition>

    <!-- 飄屏彈幕 -->
    <DanmakuOverlay v-if="!streamEnded" :messages="chatStore.messages" />

    <!-- 頂部浮層：返回按鈕 + 主播資訊 + 即時觀看人數 -->
    <div class="overlay-top">
      <div class="back-btn" @click="goBack">
        <van-icon name="arrow-left" size="22" color="#fff" />
      </div>

      <div
        v-if="streamStore.currentStream"
        class="host-info"
      >
        <van-image
          round
          :src="streamStore.currentStream.host_avatar"
          width="36"
          height="36"
          fit="cover"
          class="host-avatar"
        />
        <div class="host-detail">
          <span class="host-name">{{ streamStore.currentStream.host_nickname }}</span>
          <span class="viewer-badge">
            <van-icon name="eye-o" size="12" />
            {{ formatViewerCount(chatStore.connected ? chatStore.viewerCount : streamStore.currentStream.viewer_count) }} 觀看
          </span>
        </div>
      </div>
    </div>

    <!-- 底部浮層：系統通知 + 聊天列表 + 輸入框 -->
    <div v-if="!streamEnded" class="overlay-bottom">
      <!-- 系統通知（XX 進入直播間） -->
      <SystemNotice :messages="chatStore.messages" />

      <!-- 聊天訊息列表 -->
      <ChatMessageList :messages="chatStore.messages" />

      <!-- 底部操作列 -->
      <div class="bottom-actions">
        <ChatInput @send="handleSendMessage" />
        <div class="action-icons">
          <!-- 愛心按鈕 -->
          <div class="like-btn-wrapper" @click="handleLike">
            <van-icon name="like-o" size="24" color="#fff" />
            <span v-if="likeCount > 0" class="like-count">{{ likeCount }}</span>
            <!-- 飄浮愛心 -->
            <div
              v-for="heart in floatingHearts"
              :key="heart.id"
              class="floating-heart"
              :style="{
                color: heart.color,
                left: heart.left + 'px',
                fontSize: heart.size + 'px',
                animationDelay: heart.delay + 'ms',
              }"
            >&#x2764;</div>
          </div>
          <!-- 分享按鈕 -->
          <van-icon name="share-o" size="24" color="#fff" @click="handleShare" />
        </div>
      </div>
    </div>

    <!-- 分享面板 -->
    <van-action-sheet
      v-model:show="showShareSheet"
      :actions="filteredShareActions"
      cancel-text="取消"
      @select="onShareSelect"
    />
  </div>
</template>

<style scoped>
.live-view-page {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #000;
  max-width: 480px;
  margin: 0 auto;
}

/* 播放器 */
.player-area {
  width: 100%;
  height: 100%;
  position: relative;
}

.player-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  background-color: #000;
}

.player-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  background: linear-gradient(180deg, #1a1a1a 0%, #0a0a0a 100%);
}

.player-overlay-text {
  font-size: 16px;
  color: #555;
  max-width: 80%;
  text-align: center;
}

.player-overlay-text.error {
  color: #ff4757;
}

/* 直播結束覆蓋層 */
.stream-ended-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}

.ended-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  text-align: center;
  padding: 32px;
}

.ended-title {
  font-size: 24px;
  font-weight: 700;
  color: #fff;
  margin: 0;
}

.ended-subtitle {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.5);
  margin: 0;
}

.ended-btn {
  margin-top: 16px;
  min-width: 160px;
  height: 44px;
  font-size: 16px;
  background: var(--accent) !important;
  border-color: var(--accent) !important;
}

/* Fade 過渡 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 頂部浮層 */
.overlay-top {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  padding: 48px 16px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.6) 0%, transparent 100%);
  z-index: 10;
}

.back-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}

.back-btn:active {
  background-color: rgba(255, 255, 255, 0.3);
}

.host-info {
  display: flex;
  align-items: center;
  gap: 10px;
  background-color: rgba(0, 0, 0, 0.4);
  border-radius: 24px;
  padding: 4px 12px 4px 4px;
}

.host-avatar {
  flex-shrink: 0;
}

.host-detail {
  display: flex;
  flex-direction: column;
}

.host-name {
  font-size: 13px;
  font-weight: 600;
  color: #fff;
}

.viewer-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: rgba(255, 255, 255, 0.7);
}

/* 底部浮層 */
.overlay-bottom {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 16px;
  padding-bottom: calc(16px + env(safe-area-inset-bottom, 0px));
  background: linear-gradient(0deg, rgba(0, 0, 0, 0.7) 0%, transparent 100%);
  z-index: 10;
}

.bottom-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}

.action-icons {
  display: flex;
  align-items: center;
  gap: 16px;
}

/* 愛心按鈕 */
.like-btn-wrapper {
  position: relative;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.like-btn-wrapper:active {
  transform: scale(1.2);
  transition: transform 0.1s;
}

.like-count {
  position: absolute;
  top: -10px;
  right: -10px;
  font-size: 10px;
  color: #fff;
  background-color: #FF2D55;
  border-radius: 8px;
  padding: 1px 5px;
  min-width: 16px;
  text-align: center;
  line-height: 14px;
}

/* 飄浮愛心動畫 */
.floating-heart {
  position: absolute;
  bottom: 20px;
  pointer-events: none;
  animation: heart-float 1.5s ease-out forwards;
  opacity: 0;
}

@keyframes heart-float {
  0% {
    opacity: 1;
    transform: translateY(0) scale(1) rotate(0deg);
  }
  25% {
    opacity: 1;
    transform: translateY(-40px) scale(1.1) rotate(-8deg);
  }
  50% {
    opacity: 0.8;
    transform: translateY(-90px) scale(1) rotate(6deg);
  }
  75% {
    opacity: 0.4;
    transform: translateY(-140px) scale(0.9) rotate(-4deg);
  }
  100% {
    opacity: 0;
    transform: translateY(-200px) scale(0.7) rotate(8deg);
  }
}
</style>
