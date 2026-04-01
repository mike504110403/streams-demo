<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

const router = useRouter()

const streamTitle = ref('')
const isStarting = ref(false)

function goBack() {
  router.back()
}

function handleStartLive() {
  if (!streamTitle.value.trim()) {
    showToast('請輸入直播標題')
    return
  }

  isStarting.value = true

  // 模擬開播（Sprint 2 只做 UI）
  setTimeout(() => {
    isStarting.value = false
    showToast({
      message: '開播功能開發中，敬請期待！',
      type: 'text',
      duration: 2000,
    })
  }, 1000)
}
</script>

<template>
  <div class="go-live-page">
    <!-- 頂部導航 -->
    <div class="go-live-header">
      <div class="back-btn" @click="goBack">
        <van-icon name="arrow-left" size="22" color="#fff" />
      </div>
      <span class="header-title">開始直播</span>
      <div class="header-placeholder" />
    </div>

    <!-- 攝影機預覽區域（Mock） -->
    <div class="camera-preview">
      <div class="camera-placeholder">
        <van-icon name="photograph" size="56" color="#444" />
        <span class="camera-text">攝影機預覽</span>
      </div>
    </div>

    <!-- 直播設定 -->
    <div class="live-settings">
      <!-- 封面選擇 -->
      <div class="cover-selector">
        <div class="cover-preview">
          <van-icon name="photo-o" size="28" color="#666" />
          <span class="cover-text">選擇封面</span>
        </div>
      </div>

      <!-- 標題輸入 -->
      <div class="title-input-wrapper">
        <van-field
          v-model="streamTitle"
          placeholder="輸入你的直播標題..."
          :border="false"
          maxlength="50"
          show-word-limit
          class="title-field"
        />
      </div>

      <!-- 開始直播按鈕 -->
      <van-button
        type="primary"
        round
        block
        size="large"
        class="start-btn"
        :loading="isStarting"
        loading-text="準備中..."
        @click="handleStartLive"
      >
        開始直播
      </van-button>
    </div>
  </div>
</template>

<style scoped>
.go-live-page {
  min-height: 100vh;
  background-color: #000;
  display: flex;
  flex-direction: column;
}

/* 頂部導航 */
.go-live-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 48px 16px 12px;
  background: transparent;
  position: relative;
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
}

.back-btn:active {
  background-color: rgba(255, 255, 255, 0.3);
}

.header-title {
  font-size: 17px;
  font-weight: 600;
  color: #fff;
}

.header-placeholder {
  width: 36px;
}

/* 攝影機預覽 */
.camera-preview {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.camera-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.camera-text {
  font-size: 14px;
  color: #555;
}

/* 直播設定 */
.live-settings {
  padding: 20px 20px calc(32px + env(safe-area-inset-bottom, 0px));
  background: linear-gradient(0deg, rgba(0, 0, 0, 0.9) 0%, transparent 100%);
}

/* 封面選擇 */
.cover-selector {
  margin-bottom: 16px;
}

.cover-preview {
  width: 72px;
  height: 96px;
  border-radius: 8px;
  border: 1.5px dashed rgba(255, 255, 255, 0.3);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  cursor: pointer;
}

.cover-preview:active {
  border-color: var(--accent);
}

.cover-text {
  font-size: 10px;
  color: #666;
}

/* 標題輸入 */
.title-input-wrapper {
  margin-bottom: 20px;
}

.title-field {
  background-color: rgba(255, 255, 255, 0.1) !important;
  border-radius: 10px;
}

.title-field :deep(.van-field__control) {
  color: #fff !important;
}

.title-field :deep(.van-field__control::placeholder) {
  color: #777 !important;
}

.title-field :deep(.van-field__word-limit) {
  color: #555;
}

/* 開始直播按鈕 */
.start-btn {
  height: 50px;
  font-size: 17px;
  font-weight: 700;
  background: var(--accent) !important;
  border-color: var(--accent) !important;
}
</style>
