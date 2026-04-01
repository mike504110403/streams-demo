<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
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

onMounted(async () => {
  const id = route.params.id as string
  await streamStore.loadStreamById(id)

  // 連線 WebSocket（Mock 或真實），使用當前直播的觀看人數作為初始值
  const initialCount = streamStore.currentStream?.viewer_count ?? 1234
  chatStore.connect(id, initialCount)
})

onUnmounted(() => {
  streamStore.clearCurrentStream()
  chatStore.disconnect()
})

function goBack() {
  router.back()
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
  left: number   // 隨機水平偏移 (px)
  delay: number  // 隨機延遲 (ms)
  size: number   // 隨機大小 (px)
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
    left: Math.random() * 40 - 20,   // -20px ~ +20px
    delay: Math.random() * 100,       // 0 ~ 100ms
    size: 24 + Math.random() * 14,    // 24 ~ 38px
  }

  floatingHearts.value.push(heart)

  // 動畫結束後移除 DOM 節點
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

// 檢查 Web Share API 是否可用，不支援就只顯示複製連結
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
      // fallback
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
      // 使用者取消分享，不做任何事
    }
  }

  showShareSheet.value = false
}
</script>

<template>
  <div class="live-view-page">
    <!-- 播放器區域（Mock 黑色背景） -->
    <div class="player-area">
      <div class="player-placeholder">
        <van-icon name="video-o" size="48" color="#444" />
        <span class="player-text" v-if="streamStore.currentStream">
          {{ streamStore.currentStream.title }}
        </span>
        <span class="player-text" v-else>載入中...</span>
      </div>
    </div>

    <!-- 飄屏彈幕 -->
    <DanmakuOverlay :messages="chatStore.messages" />

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
    <div class="overlay-bottom">
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
}

.player-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  background: linear-gradient(180deg, #1a1a1a 0%, #0a0a0a 100%);
}

.player-text {
  font-size: 16px;
  color: #555;
  max-width: 80%;
  text-align: center;
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
