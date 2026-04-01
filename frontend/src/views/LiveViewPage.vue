<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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

  // 連線 Mock WebSocket，使用當前直播的觀看人數作為初始值
  const initialCount = streamStore.currentStream?.viewer_count ?? 1234
  chatStore.connect(initialCount)
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
          <van-icon name="like-o" size="24" color="#fff" />
          <van-icon name="share-o" size="24" color="#fff" />
        </div>
      </div>
    </div>
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
</style>
