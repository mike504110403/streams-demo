<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { showToast, showLoadingToast, closeToast, showConfirmDialog } from 'vant'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()

const isEditing = ref(false)
const editNickname = ref('')
const editBio = ref('')
const saving = ref(false)

onMounted(async () => {
  try {
    await authStore.fetchProfile()
  } catch {
    showToast({ message: '載入個人資料失敗', type: 'fail' })
  }
})

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
  return authStore.user?.avatar_url || 'https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg'
}
</script>

<template>
  <div class="profile-page">
    <van-nav-bar title="個人資料" />

    <div class="profile-content">
      <!-- 頭像區域 -->
      <div class="avatar-section">
        <van-image
          round
          width="80"
          height="80"
          :src="getAvatarUrl()"
          fit="cover"
          class="avatar"
        />
        <h2 class="nickname">{{ authStore.user?.nickname || '未設定暱稱' }}</h2>
        <p class="phone">{{ authStore.user?.phone || '' }}</p>
      </div>

      <!-- 檢視模式 -->
      <div v-if="!isEditing" class="info-section">
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

      <!-- 編輯模式 -->
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

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 0 24px;
}

.avatar {
  margin-bottom: 16px;
  border: 3px solid var(--accent);
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
