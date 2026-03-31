<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const nickname = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)

async function handleRegister() {
  if (!nickname.value || !email.value || !password.value || !confirmPassword.value) {
    showToast('請填寫所有欄位')
    return
  }

  if (password.value !== confirmPassword.value) {
    showToast({ message: '兩次密碼不一致', type: 'fail' })
    return
  }

  if (password.value.length < 6) {
    showToast({ message: '密碼至少需要 6 個字元', type: 'fail' })
    return
  }

  loading.value = true
  showLoadingToast({ message: '註冊中...', forbidClick: true })

  try {
    await authStore.register({
      nickname: nickname.value,
      email: email.value,
      password: password.value,
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

function goLogin() {
  router.push('/login')
}
</script>

<template>
  <div class="register-page">
    <div class="register-header">
      <h1 class="app-title">Streams</h1>
      <p class="app-subtitle">建立你的帳號</p>
    </div>

    <div class="register-form">
      <van-field
        v-model="nickname"
        placeholder="暱稱"
        :border="false"
        size="large"
        left-icon="user-o"
        autocomplete="nickname"
      />

      <van-field
        v-model="email"
        type="email"
        placeholder="Email"
        :border="false"
        size="large"
        left-icon="envelop-o"
        autocomplete="email"
      />

      <van-field
        v-model="password"
        type="password"
        placeholder="密碼（至少 6 個字元）"
        :border="false"
        size="large"
        left-icon="lock"
        autocomplete="new-password"
      />

      <van-field
        v-model="confirmPassword"
        type="password"
        placeholder="確認密碼"
        :border="false"
        size="large"
        left-icon="lock"
        autocomplete="new-password"
        @keyup.enter="handleRegister"
      />

      <van-button
        type="primary"
        class="btn-primary"
        :loading="loading"
        loading-text="註冊中..."
        @click="handleRegister"
      >
        註冊
      </van-button>

      <div class="register-footer">
        <span class="footer-text">已有帳號？</span>
        <a class="footer-link" @click="goLogin">去登入</a>
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
  margin-bottom: 40px;
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

.register-form {
  display: flex;
  flex-direction: column;
}

.register-form .van-field {
  margin-bottom: 16px;
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
