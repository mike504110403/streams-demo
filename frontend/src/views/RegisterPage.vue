<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const DEBUG_MODE = import.meta.env.VITE_DEBUG_MODE !== 'false'
const step = ref(1)
const phone = ref('')
const code = ref('')
const nickname = ref('')
const loading = ref(false)
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const canSendCode = computed(() => {
  return phone.value.length >= 10 && countdown.value === 0
})

function startCountdown() {
  countdown.value = 60
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      if (countdownTimer) clearInterval(countdownTimer)
      countdownTimer = null
    }
  }, 1000)
}

onUnmounted(() => {
  if (countdownTimer) clearInterval(countdownTimer)
})

async function handleSendCode() {
  if (!phone.value) {
    showToast('請輸入手機號碼')
    return
  }

  const phoneRegex = /^09\d{8}$/
  if (!phoneRegex.test(phone.value)) {
    showToast('請輸入正確的手機號碼（09 開頭，10 碼）')
    return
  }

  try {
    const formatted = '+886' + phone.value.slice(1)
    await authStore.sendCode({ phone: formatted })
    startCountdown()
    showToast({ message: '驗證碼已發送', type: 'success' })
  } catch (error: unknown) {
    const err = error as { response?: { data?: { error?: string } } }
    showToast({ message: err.response?.data?.error || '發送失敗，請稍後再試', type: 'fail' })
  }
}

function handleNextStep() {
  if (!phone.value || !code.value) {
    showToast('請填寫手機號碼和驗證碼')
    return
  }

  if (!DEBUG_MODE && code.value.length !== 6) {
    showToast('請輸入 6 位數驗證碼')
    return
  }

  step.value = 2
}

async function handleRegister() {
  if (!nickname.value) {
    showToast('請輸入暱稱')
    return
  }

  loading.value = true
  showLoadingToast({ message: '註冊中...', forbidClick: true })

  try {
    const formatted = '+886' + phone.value.slice(1)
    await authStore.register({
      phone: formatted,
      code: code.value,
      nickname: nickname.value,
    })
    closeToast()
    showToast({ message: '註冊成功', type: 'success' })
    router.push('/')
  } catch (error: unknown) {
    closeToast()
    const err = error as { response?: { data?: { message?: string } } }
    const msg = err.response?.data?.message || '註冊失敗，請稍後再試'
    showToast({ message: msg, type: 'fail' })
  } finally {
    loading.value = false
  }
}

function goBack() {
  step.value = 1
}

function goLogin() {
  router.push('/login')
}
</script>

<template>
  <div class="register-page">
    <div class="register-header">
      <h1 class="app-title">Streams</h1>
      <p class="app-subtitle">建立你的帳號</p>
      <span v-if="DEBUG_MODE" class="debug-badge">DEBUG MODE — 驗證碼可隨意輸入</span>
    </div>

    <!-- 步驟指示 -->
    <div class="steps">
      <div class="step" :class="{ active: step >= 1 }">1</div>
      <div class="step-line" :class="{ active: step >= 2 }"></div>
      <div class="step" :class="{ active: step >= 2 }">2</div>
    </div>

    <!-- 步驟一：手機號碼 + 驗證碼 -->
    <div v-if="step === 1" class="register-form">
      <van-field
        v-model="phone"
        type="tel"
        placeholder="手機號碼（09 開頭）"
        :border="false"
        size="large"
        left-icon="phone-o"
        maxlength="10"
      />

      <div class="code-row">
        <van-field
          v-model="code"
          type="digit"
          placeholder="6 位數驗證碼"
          :border="false"
          size="large"
          left-icon="shield-o"
          maxlength="6"
          class="code-input"
        />
        <van-button
          size="small"
          type="primary"
          class="send-code-btn"
          :disabled="!canSendCode"
          @click="handleSendCode"
        >
          {{ countdown > 0 ? `${countdown}s` : '發送驗證碼' }}
        </van-button>
      </div>

      <van-button
        type="primary"
        class="btn-primary"
        @click="handleNextStep"
      >
        下一步
      </van-button>

      <div class="register-footer">
        <span class="footer-text">已有帳號？</span>
        <a class="footer-link" @click="goLogin">去登入</a>
      </div>
    </div>

    <!-- 步驟二：暱稱 -->
    <div v-if="step === 2" class="register-form">
      <van-field
        v-model="nickname"
        placeholder="取一個暱稱"
        :border="false"
        size="large"
        left-icon="user-o"
        @keyup.enter="handleRegister"
      />

      <van-button
        type="primary"
        class="btn-primary"
        :loading="loading"
        loading-text="註冊中..."
        @click="handleRegister"
      >
        完成註冊
      </van-button>

      <div class="register-footer">
        <a class="footer-link" @click="goBack">← 上一步</a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.register-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 0 32px;
  background-color: var(--bg-primary);
}

.register-header {
  text-align: center;
  margin-bottom: 24px;
}

.app-title {
  font-size: 42px;
  font-weight: 700;
  color: var(--accent);
  margin-bottom: 8px;
  letter-spacing: -1px;
}

.app-subtitle {
  font-size: 16px;
  color: var(--text-secondary);
}

.debug-badge {
  display: inline-block;
  margin-top: 8px;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  color: #fff;
  background-color: #ff9500;
}

.steps {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
  margin-bottom: 32px;
}

.step {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
  background-color: var(--bg-secondary, #2a2a2a);
  color: var(--text-secondary);
  transition: all 0.3s;
}

.step.active {
  background-color: var(--accent);
  color: #fff;
}

.step-line {
  width: 60px;
  height: 2px;
  background-color: var(--bg-secondary, #2a2a2a);
  transition: all 0.3s;
}

.step-line.active {
  background-color: var(--accent);
}

.register-form {
  display: flex;
  flex-direction: column;
}

.register-form .van-field {
  margin-bottom: 16px;
}

.code-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.code-row .code-input {
  flex: 1;
}

.send-code-btn {
  flex-shrink: 0;
  height: 40px;
  padding: 0 16px;
  font-size: 13px;
  white-space: nowrap;
}

.register-form .btn-primary {
  margin-top: 8px;
}

.register-footer {
  text-align: center;
  margin-top: 24px;
}

.footer-text {
  color: var(--text-secondary);
  font-size: 14px;
}

.footer-link {
  color: var(--accent);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  margin-left: 4px;
}
</style>
