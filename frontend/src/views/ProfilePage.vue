<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast, showConfirmDialog } from 'vant'
import type { UploaderFileListItem } from 'vant'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 判斷是否為公開模式（/user/:id 路由）
const isPublicMode = computed(() => !!route.params.id)
const isOwnProfile = computed(() => {
  if (!isPublicMode.value) return true
  return route.params.id === authStore.user?.id
})

// 公開用戶資料
interface PublicUser {
  id: string
  nickname: string
  avatar_url: string | null
  bio: string | null
}

const publicUser = ref<PublicUser | null>(null)
const loadingPublic = ref(false)

// 顯示用的用戶資料
const displayUser = computed(() => {
  if (isPublicMode.value && !isOwnProfile.value) {
    return publicUser.value
  }
  return authStore.user
})

const isEditing = ref(false)
const editNickname = ref('')
const editBio = ref('')
const saving = ref(false)
const uploadingAvatar = ref(false)

onMounted(async () => {
  if (isPublicMode.value && !isOwnProfile.value) {
    await loadPublicUser(route.params.id as string)
  } else {
    try {
      await authStore.fetchProfile()
    } catch {
      showToast({ message: '載入個人資料失敗', type: 'fail' })
    }
  }
})

// 監聽路由參數變化
watch(() => route.params.id, async (newId) => {
  if (newId && newId !== authStore.user?.id) {
    await loadPublicUser(newId as string)
  }
})

async function loadPublicUser(userId: string) {
  loadingPublic.value = true
  try {
    const { data } = await api.get(`/users/${userId}`)
    publicUser.value = data.data
  } catch {
    showToast({ message: '用戶不存在', type: 'fail' })
    router.back()
  } finally {
    loadingPublic.value = false
  }
}

function startEdit() {
  editNickname.value = authStore.user?.nickname || ''
  editBio.value = authStore.user?.bio || ''
  isEditing.value = true
}

function cancelEdit() {
  isEditing.value = false
}

async function saveProfile() {
  if (!editNickname.value.trim()) {
    showToast('暱稱不能為空')
    return
  }

  saving.value = true
  showLoadingToast({ message: '儲存中...', forbidClick: true })

  try {
    await authStore.updateProfile({
      nickname: editNickname.value.trim(),
      bio: editBio.value.trim(),
    })
    closeToast()
    showToast({ message: '更新成功', type: 'success' })
    isEditing.value = false
  } catch {
    closeToast()
    showToast({ message: '更新失敗', type: 'fail' })
  } finally {
    saving.value = false
  }
}

async function handleLogout() {
  try {
    await showConfirmDialog({
      title: '確認登出',
      message: '確定要登出嗎？',
      confirmButtonColor: '#fe2c55',
    })
    await authStore.logout()
  } catch {
    // 使用者取消
  }
}

function getAvatarUrl(): string {
  const user = displayUser.value
  return user?.avatar_url || 'https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg'
}

/**
 * 頭像上傳 — 使用 Vant Uploader 的 afterRead 回呼
 * 流程：選擇圖片 → 前端裁切為正方形 → POST /api/v1/upload/image → PUT /users/me 更新 avatar_url
 */
async function onAvatarRead(file: UploaderFileListItem | UploaderFileListItem[]) {
  const item = Array.isArray(file) ? file[0] : file
  if (!item?.file) return

  // 驗證檔案類型
  const allowedTypes = ['image/jpeg', 'image/png']
  if (!allowedTypes.includes(item.file.type)) {
    showToast({ message: '只支援 JPG 和 PNG 格式', type: 'fail' })
    return
  }

  // 驗證檔案大小（2MB）
  if (item.file.size > 2 * 1024 * 1024) {
    showToast({ message: '圖片大小不能超過 2MB', type: 'fail' })
    return
  }

  uploadingAvatar.value = true
  showLoadingToast({ message: '上傳中...', forbidClick: true })

  try {
    // 前端裁切為正方形
    const croppedBlob = await cropToSquare(item.file)

    // 上傳圖片
    const formData = new FormData()
    formData.append('file', croppedBlob, item.file.name)
    formData.append('type', 'avatar')

    const { data: uploadRes } = await api.post('/upload/image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })

    const avatarUrl = uploadRes.data.url

    // 更新用戶資料
    await authStore.updateProfile({ avatar_url: avatarUrl })

    closeToast()
    showToast({ message: '頭像更新成功', type: 'success' })
  } catch (err: any) {
    closeToast()
    const message = err?.response?.data?.error || '頭像上傳失敗'
    showToast({ message, type: 'fail' })
  } finally {
    uploadingAvatar.value = false
  }
}

/**
 * 前端裁切圖片為正方形（取中心區域）
 */
function cropToSquare(file: File): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    const url = URL.createObjectURL(file)

    img.onload = () => {
      URL.revokeObjectURL(url)

      const size = Math.min(img.width, img.height)
      const offsetX = (img.width - size) / 2
      const offsetY = (img.height - size) / 2

      // 最終輸出尺寸：最大 400x400
      const outputSize = Math.min(size, 400)

      const canvas = document.createElement('canvas')
      canvas.width = outputSize
      canvas.height = outputSize

      const ctx = canvas.getContext('2d')
      if (!ctx) {
        reject(new Error('Canvas context not available'))
        return
      }

      ctx.drawImage(img, offsetX, offsetY, size, size, 0, 0, outputSize, outputSize)

      canvas.toBlob(
        (blob) => {
          if (blob) {
            resolve(blob)
          } else {
            reject(new Error('Canvas toBlob failed'))
          }
        },
        file.type === 'image/png' ? 'image/png' : 'image/jpeg',
        0.85
      )
    }

    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('Image load failed'))
    }

    img.src = url
  })
}

function goBack() {
  router.back()
}
</script>

<template>
  <div class="profile-page">
    <van-nav-bar
      :title="isPublicMode && !isOwnProfile ? '用戶資料' : '個人資料'"
      :left-arrow="isPublicMode"
      @click-left="goBack"
    />

    <!-- 載入中（公開模式） -->
    <div v-if="loadingPublic" class="loading-wrapper">
      <van-loading size="36" color="var(--accent)" />
    </div>

    <div v-else class="profile-content">
      <!-- 頭像區域 -->
      <div class="avatar-section">
        <div class="avatar-wrapper">
          <van-image
            round
            width="80"
            height="80"
            :src="getAvatarUrl()"
            fit="cover"
            class="avatar"
          />
          <!-- 自己的頁面才顯示上傳按鈕 -->
          <van-uploader
            v-if="isOwnProfile"
            :after-read="onAvatarRead"
            :max-count="1"
            :max-size="2 * 1024 * 1024"
            accept="image/jpeg,image/png"
            result-type="file"
            class="avatar-uploader"
            @oversize="() => showToast({ message: '圖片大小不能超過 2MB', type: 'fail' })"
          >
            <div class="avatar-edit-icon">
              <van-icon name="photograph" size="14" color="#fff" />
            </div>
          </van-uploader>
        </div>
        <h2 class="nickname">{{ displayUser?.nickname || '未設定暱稱' }}</h2>
        <p v-if="isOwnProfile && authStore.user?.phone" class="phone">{{ authStore.user.phone }}</p>
      </div>

      <!-- 公開模式（他人）：只顯示基本資訊 -->
      <div v-if="isPublicMode && !isOwnProfile" class="info-section">
        <van-cell-group :border="false" class="info-group">
          <van-cell title="暱稱" :value="displayUser?.nickname || '未設定'" />
          <van-cell title="簡介" :value="displayUser?.bio || '這個人很懶，什麼都沒寫'" />
        </van-cell-group>
      </div>

      <!-- 自己的頁面：檢視模式 -->
      <div v-else-if="!isEditing" class="info-section">
        <van-cell-group :border="false" class="info-group">
          <van-cell title="暱稱" :value="authStore.user?.nickname || '未設定'" />
          <van-cell title="手機號碼" :value="authStore.user?.phone || '未設定'" />
          <van-cell title="簡介" :value="authStore.user?.bio || '這個人很懶，什麼都沒寫'" />
        </van-cell-group>

        <div class="action-buttons">
          <van-button
            type="primary"
            class="btn-primary"
            @click="startEdit"
          >
            編輯資料
          </van-button>

          <van-button
            plain
            class="btn-logout"
            @click="handleLogout"
          >
            登出
          </van-button>
        </div>
      </div>

      <!-- 自己的頁面：編輯模式 -->
      <div v-else class="edit-section">
        <van-field
          v-model="editNickname"
          label="暱稱"
          placeholder="請輸入暱稱"
          :border="false"
          size="large"
        />

        <van-field
          v-model="editBio"
          label="簡介"
          type="textarea"
          placeholder="介紹一下自己"
          :border="false"
          rows="3"
          autosize
          maxlength="200"
          show-word-limit
        />

        <div class="action-buttons">
          <van-button
            type="primary"
            class="btn-primary"
            :loading="saving"
            loading-text="儲存中..."
            @click="saveProfile"
          >
            儲存
          </van-button>

          <van-button
            plain
            class="btn-cancel"
            @click="cancelEdit"
          >
            取消
          </van-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  min-height: 100vh;
  background-color: var(--bg-primary);
}

.profile-content {
  padding: 0 20px;
}

.loading-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 50vh;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 0 24px;
}

.avatar-wrapper {
  position: relative;
  margin-bottom: 16px;
}

.avatar {
  border: 3px solid var(--accent);
}

.avatar-uploader {
  position: absolute;
  bottom: 0;
  right: 0;
}

.avatar-uploader :deep(.van-uploader__wrapper) {
  margin: 0;
}

.avatar-uploader :deep(.van-uploader__input-wrapper) {
  margin: 0;
}

.avatar-edit-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background-color: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--bg-primary);
  cursor: pointer;
}

.nickname {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.phone {
  font-size: 14px;
  color: var(--text-secondary);
}

.info-group {
  margin-bottom: 24px;
  border-radius: 12px;
  overflow: hidden;
}

.info-group .van-cell {
  background-color: var(--bg-secondary);
}

.info-group .van-cell__title {
  color: var(--text-secondary);
}

.info-group .van-cell__value {
  color: var(--text-primary);
}

.edit-section .van-field {
  margin-bottom: 16px;
  background-color: var(--input-bg) !important;
  border-radius: 8px;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 24px;
  padding-bottom: 32px;
}

.btn-logout {
  width: 100%;
  height: 48px;
  border-radius: 24px;
  font-size: 16px;
  color: var(--accent) !important;
  border-color: var(--accent) !important;
  background: transparent !important;
}

.btn-cancel {
  width: 100%;
  height: 48px;
  border-radius: 24px;
  font-size: 16px;
  color: var(--text-secondary) !important;
  border-color: var(--border-color) !important;
  background: transparent !important;
}
</style>
