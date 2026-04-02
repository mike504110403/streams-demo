<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showDialog } from 'vant'
import { useAuthStore } from '../stores/auth'
import { useStreamStore } from '../stores/stream'

const router = useRouter()
const authStore = useAuthStore()
const streamStore = useStreamStore()

onMounted(() => {
  if (streamStore.streams.length === 0) {
    streamStore.loadStreams()
  }
})

function onRefresh() {
  streamStore.refreshStreams()
}

function onLoadMore() {
  streamStore.loadStreams()
}

function goToStream(id: string) {
  router.push(`/live/${id}`)
}

function showLoginPrompt(redirectPath: string) {
  showDialog({
    title: '尚未登入',
    message: '登入後才能使用此功能，要前往登入嗎？',
    confirmButtonText: '前往登入',
    cancelButtonText: '稍後再說',
    showCancelButton: true,
    confirmButtonColor: '#fe2c55',
  }).then(() => {
    router.push('/login?redirect=' + encodeURIComponent(redirectPath))
  }).catch(() => {
    // 取消
  })
}

function goToLive() {
  if (authStore.isAuthenticated) {
    router.push('/go-live')
  } else {
    showLoginPrompt('/go-live')
  }
}

function onTabChange(index: number) {
  if (index === 1) {
    goToLive()
  } else if (index === 2) {
    if (authStore.isAuthenticated) {
      router.push('/profile')
    } else {
      showLoginPrompt('/profile')
    }
  }
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

function getStatusLabel(status: string): string {
  switch (status) {
    case 'live':
      return '直播中'
    case 'replay':
      return '回放'
    case 'upcoming':
      return '預告'
    default:
      return ''
  }
}
</script>

<template>
  <div class="live-hall-page">
    <!-- 頂部標題 -->
    <div class="hall-header">
      <div class="header-logo">
        <van-icon name="video-o" size="24" color="#fe2c55" />
        <span class="header-title">Streams</span>
      </div>
    </div>

    <!-- 直播列表 -->
    <div class="hall-content">
      <van-pull-refresh
        v-model="streamStore.refreshing"
        @refresh="onRefresh"
        pulling-text="下拉刷新"
        loosing-text="釋放刷新"
        loading-text="刷新中..."
        success-text="刷新成功"
      >
        <van-list
          v-model:loading="streamStore.loading"
          :finished="!streamStore.hasMore"
          finished-text="沒有更多了"
          @load="onLoadMore"
        >
          <!-- 空狀態 -->
          <div
            v-if="!streamStore.loading && streamStore.streams.length === 0"
            class="empty-state"
          >
            <div class="empty-icon">
              <van-icon name="video-o" size="64" color="#444" />
            </div>
            <p class="empty-title">目前沒有直播</p>
            <p class="empty-subtitle">稍後再來看看，或自己開一場吧！</p>
            <van-button
              round
              plain
              size="small"
              color="#888"
              class="empty-refresh-btn"
              @click="onRefresh"
            >
              重新整理
            </van-button>
          </div>

          <!-- 直播卡片網格 -->
          <div v-else class="stream-grid">
            <div
              v-for="stream in streamStore.streams"
              :key="stream.id"
              class="stream-card"
              @click="goToStream(stream.id)"
            >
              <!-- 封面圖 -->
              <div class="card-cover">
                <!-- 有封面 URL 時顯示圖片 -->
                <van-image
                  v-if="stream.cover_url"
                  :src="stream.cover_url"
                  fit="cover"
                  width="100%"
                  height="100%"
                  class="cover-image"
                >
                  <template #loading>
                    <div class="cover-placeholder">
                      <van-loading type="spinner" size="20" />
                    </div>
                  </template>
                  <template #error>
                    <div class="cover-fallback">
                      <van-icon name="video-o" size="36" color="rgba(255,255,255,0.5)" />
                      <span class="fallback-title">{{ stream.title }}</span>
                    </div>
                  </template>
                </van-image>
                <!-- 沒有封面 URL 時顯示漸層背景 + 標題 -->
                <div v-else class="cover-fallback">
                  <van-icon name="video-o" size="36" color="rgba(255,255,255,0.5)" />
                  <span class="fallback-title">{{ stream.title }}</span>
                </div>

                <!-- 狀態標籤 -->
                <span
                  class="status-tag"
                  :class="{
                    'status-live': stream.status === 'live',
                    'status-replay': stream.status === 'replay',
                    'status-upcoming': stream.status === 'upcoming',
                  }"
                >
                  {{ getStatusLabel(stream.status) }}
                </span>

                <!-- 觀看人數 -->
                <span v-if="stream.viewer_count > 0" class="viewer-count">
                  <van-icon name="eye-o" size="12" />
                  {{ formatViewerCount(stream.viewer_count) }}
                </span>
              </div>

              <!-- 卡片資訊 -->
              <div class="card-info">
                <div class="card-title">{{ stream.title }}</div>
                <div class="card-host">
                  <van-image
                    round
                    :src="stream.host_avatar"
                    width="20"
                    height="20"
                    fit="cover"
                    class="host-avatar"
                  />
                  <span class="host-name">{{ stream.host_nickname }}</span>
                </div>
              </div>
            </div>
          </div>
        </van-list>
      </van-pull-refresh>
    </div>

    <!-- 右下角浮動開播按鈕 -->
    <div class="fab-go-live" @click="goToLive">
      <van-icon name="plus" size="28" color="#fff" />
    </div>

    <!-- 底部 TabBar -->
    <van-tabbar :model-value="0" @change="onTabChange">
      <van-tabbar-item icon="home-o">首頁</van-tabbar-item>
      <van-tabbar-item icon="video-o">開播</van-tabbar-item>
      <van-tabbar-item icon="user-o">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<style scoped>
.live-hall-page {
  min-height: 100vh;
  max-width: 100%;
  background-color: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

.hall-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px 16px;
  background-color: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
}

.header-logo {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 20px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.5px;
}

.hall-content {
  flex: 1;
  padding: 12px 12px 60px;
  overflow-y: auto;
}

/* 直播卡片網格 - 響應式 */
.stream-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}

/* 平板：雙列 */
@media (min-width: 768px) {
  .stream-grid {
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
}

/* 桌面：三列 */
@media (min-width: 1024px) {
  .stream-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 14px;
  }
}

/* 大桌面：四列 */
@media (min-width: 1440px) {
  .stream-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
  }
}

.stream-card {
  border-radius: 10px;
  overflow: hidden;
  background-color: var(--bg-secondary);
  cursor: pointer;
  transition: transform 0.15s ease;
}

.stream-card:active {
  transform: scale(0.97);
}

/* 手機單列時卡片改為橫向佈局 */
@media (max-width: 767px) {
  .stream-card {
    display: flex;
    flex-direction: row;
  }

  .card-cover {
    width: 120px;
    min-height: 120px;
    aspect-ratio: auto;
    flex-shrink: 0;
  }

  .card-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 10px 12px;
  }

  .card-title {
    white-space: normal;
    -webkit-line-clamp: 2;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
}

/* 封面區域 */
.card-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 3 / 4;
  overflow: hidden;
}

.cover-image {
  display: block;
}

.cover-image :deep(.van-image__img) {
  display: block;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--bg-card);
}

/* 無封面時的漸層背景 */
.cover-fallback {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background: linear-gradient(135deg, #2a2a3e 0%, #1a1a2e 40%, #16213e 100%);
  padding: 12px;
}

.fallback-title {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  text-align: center;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* 狀態標籤 */
.status-tag {
  position: absolute;
  top: 8px;
  left: 8px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  color: #fff;
}

.status-live {
  background-color: var(--accent);
}

.status-replay {
  background-color: rgba(255, 255, 255, 0.3);
}

.status-upcoming {
  background-color: #ff9500;
}

/* 觀看人數 */
.viewer-count {
  position: absolute;
  bottom: 8px;
  left: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  background-color: rgba(0, 0, 0, 0.6);
  font-size: 11px;
  color: #fff;
}

/* 卡片資訊 */
.card-info {
  padding: 8px 10px 10px;
}

.card-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-host {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
}

.host-avatar {
  flex-shrink: 0;
}

.host-name {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 浮動開播按鈕 */
.fab-go-live {
  position: fixed;
  bottom: 80px;
  right: 20px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--accent, #fe2c55);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 14px rgba(254, 44, 85, 0.4);
  cursor: pointer;
  z-index: 20;
  transition: transform 0.15s ease;
}

.fab-go-live:active {
  transform: scale(0.92);
}

/* Vant 覆寫 */
:deep(.van-pull-refresh) {
  min-height: calc(100vh - 130px);
}

/* 自訂空狀態 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80px 32px 40px;
  text-align: center;
}

.empty-icon {
  margin-bottom: 20px;
  opacity: 0.6;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.empty-subtitle {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 24px;
}

.empty-refresh-btn {
  min-width: 100px;
}

:deep(.van-list__finished-text) {
  color: var(--text-muted);
  padding: 16px 0;
}

:deep(.van-list__loading) {
  padding: 16px 0;
}
</style>
