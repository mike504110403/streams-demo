<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { useAuthStore } from '../stores/auth'
import { shouldShowAppleLogin, shouldShowGoogleLogin } from '../utils/device'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const phone = ref('')
const code = ref('')
const loading = ref(false)
const oauthLoading = ref(false)
const codeSent = ref(false)
const countdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const DEBUG_MODE = import.meta.env.VITE_DEBUG_MODE !== 'false'
const showApple = shouldShowAppleLogin()
const showGoogle = shouldShowGoogleLogin()

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
    codeSent.value = true
    startCountdown()
    showToast({ message: '驗證碼已發送', type: 'success' })
  } catch (error: unknown) {
    const err = error as { response?: { data?: { error?: string } } }
    showToast({ message: err.response?.data?.error || '發送失敗，請稍後再試', type: 'fail' })
  }
}

async function handleLogin() {
  if (!phone.value || !code.value) {
    showToast('請填寫手機號碼和驗證碼')
    return
  }

  if (!DEBUG_MODE && code.value.length !== 6) {
    showToast('請輸入 6 位數驗證碼')
    return
  }

  loading.value = true
  showLoadingToast({ message: '登入中...', forbidClick: true })

  try {
    const formatted = '+886' + phone.value.slice(1)
    await authStore.login({
      phone: formatted,
      code: code.value,
    })
    closeToast()
    showToast({ message: '登入成功', type: 'success' })

    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (error: unknown) {
    closeToast()
    const err = error as { response?: { data?: { message?: string } } }
    const msg = err.response?.data?.message || '登入失敗，請檢查驗證碼'
    showToast({ message: msg, type: 'fail' })
  } finally {
    loading.value = false
  }
}

async function handleAppleLogin() {
  oauthLoading.value = true
  showLoadingToast({ message: '透過 Apple 登入中...', forbidClick: true })

  try {
    // MVP Mock：模擬 Apple OAuth 授權流程
    const result = await authStore.oauthApple({
      code: 'mock_apple_auth_code',
      id_token: 'mock_apple_id_token_' + Date.now(),
      user: { name: 'Apple 用戶', email: 'user@icloud.com' },
    })
    closeToast()
    const msg = result.is_new_user ? '帳號已建立，歡迎加入！' : '登入成功'
    showToast({ message: msg, type: 'success' })
    router.push((route.query.redirect as string) || '/')
  } catch (error: unknown) {
    closeToast()
    const err = error as { response?: { data?: { message?: string } } }
    showToast({ message: err.response?.data?.message || 'Apple 登入失敗，請重試', type: 'fail' })
  } finally {
    oauthLoading.value = false
  }
}

async function handleGoogleLogin() {
  oauthLoading.value = true
  showLoadingToast({ message: '透過 Google 登入中...', forbidClick: true })

  try {
    // MVP Mock：模擬 Google OAuth 授權流程
    const result = await authStore.oauthGoogle({
      credential: 'mock_google_credential_' + Date.now(),
    })
    closeToast()
    const msg = result.is_new_user ? '帳號已建立，歡迎加入！' : '登入成功'
    showToast({ message: msg, type: 'success' })
    router.push((route.query.redirect as string) || '/')
  } catch (error: unknown) {
    closeToast()
    const err = error as { response?: { data?: { message?: string } } }
    showToast({ message: err.response?.data?.message || 'Google 登入失敗，請重試', type: 'fail' })
  } finally {
    oauthLoading.value = false
  }
}

function goRegister() {
  router.push('/register')
}
</script>

<template>
  <div class="login-page">
    <div class="login-header">
      <h1 class="app-title">Streams</h1>
      <p class="app-subtitle">Live your moment</p>
      <span v-if="DEBUG_MODE" class="debug-badge">DEBUG MODE — 驗證碼可隨意輸入</span>
    </div>

    <div class="login-form">
      <van-field
        v-model="phone"
        type="tel"
        placeholder="手機號碼（09 開頭）"
        :border="false"
        autocomplete="tel"
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
          @keyup.enter="handleLogin"
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
        :loading="loading"
        loading-text="登入中..."
        :disabled="!DEBUG_MODE && !codeSent"
        @click="handleLogin"
      >
        登入
      </van-button>

      <!-- 分隔線 -->
      <div class="divider">
        <span class="divider-line"></span>
        <span class="divider-text">或</span>
        <span class="divider-line"></span>
      </div>

      <!-- 社交登入 -->
      <div class="oauth-buttons">
        <van-button
          v-if="showApple"
          class="oauth-btn apple-btn"
          :loading="oauthLoading"
          @click="handleAppleLogin"
        >
          <span class="oauth-icon">&#xF8FF;</span>
          透過 Apple 登入
        </van-button>

        <van-button
          v-if="showGoogle"
          class="oauth-btn google-btn"
          :loading="oauthLoading"
          @click="handleGoogleLogin"
        >
          <span class="oauth-icon">G</span>
          透過 Google 登入
        </van-button>
      </div>

      <div class="login-footer">
        <span class="footer-text">還沒有帳號？</span>
        <a class="footer-link" @click="goRegister">去註冊</a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 0 32px;
  background-color: var(--bg-primary);
}

.login-header {
  text-align: center;
  margin-bottom: 48px;
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

.login-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.login-form .van-field {
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

.login-form .btn-primary {
  margin-top: 8px;
}

.login-footer {
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

/* 分隔線 */
.divider {
  display: flex;
  align-items: center;
  margin: 24px 0 20px;
}

.divider-line {
  flex: 1;
  height: 1px;
  background-color: var(--border-color, #333);
}

.divider-text {
  padding: 0 16px;
  font-size: 13px;
  color: var(--text-secondary);
}

/* OAuth 按鈕 */
.oauth-buttons {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 8px;
}

.oauth-btn {
  width: 100%;
  height: 48px;
  border-radius: 24px;
  font-size: 15px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border: 1px solid var(--border-color, #333) !important;
}

.apple-btn {
  background-color: #000 !important;
  color: #fff !important;
  border-color: #000 !important;
}

.google-btn {
  background-color: #fff !important;
  color: #333 !important;
  border-color: #ddd !important;
}

.oauth-icon {
  font-size: 18px;
  font-weight: 700;
}
</style>
