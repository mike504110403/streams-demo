<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useStreamStore } from '../stores/stream'

const route = useRoute()
const router = useRouter()
const streamStore = useStreamStore()

onMounted(async () => {
  const id = route.params.id as string
  await streamStore.loadStreamById(id)
})

onUnmounted(() => {
  streamStore.clearCurrentStream()
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

    <!-- 頂部浮層：返回按鈕 + 主播資訊 -->
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
            {{ formatViewerCount(streamStore.currentStream.viewer_count) }} 觀看
          </span>
        </div>
      </div>
    </div>

    <!-- 底部預留彈幕區域 -->
    <div class="overlay-bottom">
      <div class="danmu-placeholder">
        <span class="danmu-hint">彈幕功能開發中（Sprint 3）</span>
      </div>
      <div class="bottom-actions">
        <div class="input-mock">
          <van-icon name="edit" size="16" color="#999" />
          <span>說點什麼...</span>
        </div>
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
}

.danmu-placeholder {
  min-height: 120px;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding-bottom: 12px;
}

.danmu-hint {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.3);
  text-align: center;
}

.bottom-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.input-mock {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 20px;
  background-color: rgba(255, 255, 255, 0.12);
  font-size: 14px;
  color: #999;
}

.action-icons {
  display: flex;
  align-items: center;
  gap: 16px;
}
</style>
