<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showDialog } from 'vant'
import { useAuthStore } from '../../stores/auth'

const emit = defineEmits<{
  (e: 'send', content: string): void
}>()

const router = useRouter()
const authStore = useAuthStore()

const inputValue = ref('')
const showInput = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)

function openInput() {
  // 未登入時彈出登入提示
  if (!authStore.isAuthenticated) {
    showLoginPrompt()
    return
  }
  showInput.value = true
  // 等 DOM 更新後 focus
  setTimeout(() => {
    inputRef.value?.focus()
  }, 100)
}

function showLoginPrompt() {
  showDialog({
    title: '尚未登入',
    message: '登入後才能發送彈幕，要前往登入嗎？',
    confirmButtonText: '前往登入',
    cancelButtonText: '稍後再說',
    showCancelButton: true,
    confirmButtonColor: '#fe2c55',
  }).then(() => {
    router.push('/login?redirect=' + encodeURIComponent(window.location.pathname))
  }).catch(() => {
    // 取消
  })
}

function handleSend() {
  const text = inputValue.value.trim()
  if (!text) return
  emit('send', text)
  inputValue.value = ''
}

function handleBlur() {
  // 延遲關閉以便 send 按鈕有時間觸發
  setTimeout(() => {
    if (!inputValue.value.trim()) {
      showInput.value = false
    }
  }, 200)
}
</script>

<template>
  <div class="chat-input-wrapper">
    <!-- 預設狀態：假輸入框 -->
    <div
      v-if="!showInput"
      class="input-trigger"
      @click="openInput"
    >
      <van-icon name="edit" size="16" color="#999" />
      <span>說點什麼...</span>
    </div>

    <!-- 輸入狀態：真實輸入框 -->
    <div v-else class="input-active">
      <input
        ref="inputRef"
        v-model="inputValue"
        class="real-input"
        placeholder="說點什麼..."
        maxlength="100"
        enterkeyhint="send"
        @keyup.enter="handleSend"
        @blur="handleBlur"
      />
      <button
        class="send-btn"
        :class="{ active: inputValue.trim().length > 0 }"
        @mousedown.prevent="handleSend"
      >
        發送
      </button>
    </div>
  </div>
</template>

<style scoped>
.chat-input-wrapper {
  flex: 1;
}

.input-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 20px;
  background-color: rgba(255, 255, 255, 0.12);
  font-size: 14px;
  color: #999;
  cursor: pointer;
}

.input-trigger:active {
  background-color: rgba(255, 255, 255, 0.18);
}

.input-active {
  display: flex;
  align-items: center;
  gap: 8px;
}

.real-input {
  flex: 1;
  padding: 10px 14px;
  border-radius: 20px;
  background-color: rgba(255, 255, 255, 0.15);
  border: none;
  outline: none;
  font-size: 14px;
  color: #fff;
}

.real-input::placeholder {
  color: #999;
}

.send-btn {
  padding: 8px 16px;
  border-radius: 20px;
  border: none;
  background-color: rgba(255, 255, 255, 0.2);
  color: rgba(255, 255, 255, 0.5);
  font-size: 14px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.send-btn.active {
  background-color: #ff4d6a;
  color: #fff;
}
</style>
