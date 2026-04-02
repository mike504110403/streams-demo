<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import type { UploaderFileListItem } from 'vant'
import api from '../services/api'

const router = useRouter()

const streamTitle = ref('')
const isStarting = ref(false)
const coverUrl = ref<string | null>(null)
const coverPreviewSrc = ref<string | null>(null)
const uploadingCover = ref(false)

// 攝影機相關
const videoRef = ref<HTMLVideoElement | null>(null)
const cameraStream = ref<MediaStream | null>(null)
const cameraError = ref<string | null>(null)
const cameraReady = ref(false)

onMounted(async () => {
  await startCamera()
})

onUnmounted(() => {
  stopCamera()
  // 釋放封面預覽 URL
  if (coverPreviewSrc.value) {
    URL.revokeObjectURL(coverPreviewSrc.value)
  }
})

async function startCamera() {
  cameraError.value = null
  try {
    const stream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: 'user', width: { ideal: 720 }, height: { ideal: 1280 } },
      audio: true,
    })
    cameraStream.value = stream
    if (videoRef.value) {
      videoRef.value.srcObject = stream
      videoRef.value.play()
    }
    cameraReady.value = true
  } catch (err: any) {
    console.error('[Camera] 存取失敗:', err)
    if (err.name === 'NotAllowedError') {
      cameraError.value = '請允許攝影機權限後重試'
    } else if (err.name === 'NotFoundError') {
      cameraError.value = '找不到攝影機裝置'
    } else {
      cameraError.value = '攝影機存取失敗'
    }
  }
}

function stopCamera() {
  if (cameraStream.value) {
    cameraStream.value.getTracks().forEach(track => track.stop())
    cameraStream.value = null
  }
  cameraReady.value = false
}

function goBack() {
  stopCamera()
  router.back()
}

/**
 * 封面圖上傳
 */
async function onCoverRead(file: UploaderFileListItem | UploaderFileListItem[]) {
  const item = Array.isArray(file) ? file[0] : file
  if (!item?.file) return

  const allowedTypes = ['image/jpeg', 'image/png']
  if (!allowedTypes.includes(item.file.type)) {
    showToast({ message: '只支援 JPG 和 PNG 格式', type: 'fail' })
    return
  }

  if (item.file.size > 2 * 1024 * 1024) {
    showToast({ message: '圖片大小不能超過 2MB', type: 'fail' })
    return
  }

  // 先顯示本地預覽
  coverPreviewSrc.value = URL.createObjectURL(item.file)

  uploadingCover.value = true
  try {
    const formData = new FormData()
    formData.append('file', item.file, item.file.name)
    formData.append('type', 'cover')

    const { data } = await api.post('/upload/image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })

    coverUrl.value = data.data.url
    showToast({ message: '封面已上傳', type: 'success' })
  } catch (err: any) {
    const message = err?.response?.data?.error || '封面上傳失敗'
    showToast({ message, type: 'fail' })
    coverUrl.value = null
    if (coverPreviewSrc.value) {
      URL.revokeObjectURL(coverPreviewSrc.value)
    }
    coverPreviewSrc.value = null
  } finally {
    uploadingCover.value = false
  }
}

function removeCover() {
  if (coverPreviewSrc.value) {
    URL.revokeObjectURL(coverPreviewSrc.value)
  }
  coverUrl.value = null
  coverPreviewSrc.value = null
}

async function handleStartLive() {
  if (!streamTitle.value.trim()) {
    showToast('請輸入直播標題')
    return
  }

  isStarting.value = true

  try {
    const payload: Record<string, string> = { title: streamTitle.value.trim() }
    if (coverUrl.value) {
      payload.cover_url = coverUrl.value
    }

    const { data } = await api.post('/streams', payload)
    const stream = data.data

    showToast({
      message: '直播間已建立！',
      type: 'success',
      duration: 1500,
    })

    // 停止攝影機（將在直播頁重新取得）
    stopCamera()

    // 導航到直播間
    router.push(`/live/${stream.id}`)
  } catch (err: any) {
    const message = err?.response?.data?.message || '建立直播間失敗，請稍後再試'
    showToast({ message, type: 'fail', duration: 2000 })
  } finally {
    isStarting.value = false
  }
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

    <!-- 攝影機預覽區域 -->
    <div class="camera-preview">
      <!-- 真實攝影機畫面 -->
      <video
        v-show="cameraReady"
        ref="videoRef"
        class="camera-video"
        autoplay
        muted
        playsinline
      />
      <!-- 攝影機錯誤/載入中 -->
      <div v-if="!cameraReady" class="camera-placeholder">
        <template v-if="cameraError">
          <van-icon name="warning-o" size="48" color="#ff4757" />
          <span class="camera-text error">{{ cameraError }}</span>
          <van-button size="small" round plain color="#fff" @click="startCamera">
            重新嘗試
          </van-button>
        </template>
        <template v-else>
          <van-loading size="36" color="#fff" />
          <span class="camera-text">啟動攝影機中...</span>
        </template>
      </div>
    </div>

    <!-- 直播設定 -->
    <div class="live-settings">
      <!-- 封面選擇 -->
      <div class="cover-selector">
        <!-- 已有封面：點擊移除 -->
        <div v-if="coverPreviewSrc" class="cover-preview has-cover" @click="removeCover">
          <van-image :src="coverPreviewSrc" width="72" height="96" fit="cover" radius="8" />
          <div class="cover-remove-badge">
            <van-icon name="cross" size="10" color="#fff" />
          </div>
          <van-loading v-if="uploadingCover" class="cover-loading" size="20" color="#fff" />
        </div>
        <!-- 無封面：上傳 -->
        <van-uploader
          v-else
          :after-read="onCoverRead"
          :max-count="1"
          :max-size="2 * 1024 * 1024"
          accept="image/jpeg,image/png"
          result-type="file"
          class="cover-uploader"
          @oversize="() => showToast({ message: '圖片不能超過 2MB', type: 'fail' })"
        >
          <div class="cover-preview">
            <van-icon name="photo-o" size="28" color="#666" />
            <span class="cover-text">選擇封面</span>
          </div>
        </van-uploader>
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
        :disabled="!cameraReady"
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
  position: relative;
  overflow: hidden;
}

.camera-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scaleX(-1); /* 鏡像翻轉，自拍模式 */
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

.camera-text.error {
  color: #ff4757;
  text-align: center;
  max-width: 200px;
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

.cover-uploader :deep(.van-uploader__wrapper) {
  margin: 0;
}

.cover-uploader :deep(.van-uploader__input-wrapper) {
  margin: 0;
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
  position: relative;
  overflow: hidden;
}

.cover-preview.has-cover {
  border: none;
}

.cover-preview:active {
  border-color: var(--accent);
}

.cover-text {
  font-size: 10px;
  color: #666;
}

.cover-remove-badge {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}

.cover-loading {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
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

.start-btn[disabled] {
  opacity: 0.5;
}
</style>
