<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast, showLoadingToast, closeToast } from 'vant'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!email.value || !password.value) {
    showToast('請填寫所有欄位')
    return
  }

  loading.value = true
  showLoadingToast({ message: '登入中...', forbidClick: true })

  try {
    await authStore.login({
      email: email.value,
      password: password.value,
    })
    closeToast()
    showToast({ message: '登入成功', type: 'success' })

    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (error: unknown) {
    closeToast()
    const err = error as { response?: { data?: { message?: string } } }
    const msg = err.response?.data?.message || '登入失敗，請檢查帳號密碼'
    showToast({ message: msg, type: 'fail' })
  } finally {
    loading.value = false
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
    </div>

    <div class="login-form">
      <van-field
        v-model="email"
        type="email"
        placeholder="Email"
        :border="false"
        autocomplete="email"
        size="large"
        left-icon="envelop-o"
      />

      <van-field
        v-model="password"
        type="password"
        placeholder="密碼"
        :border="false"
        autocomplete="current-password"
        size="large"
        left-icon="lock"
        @keyup.enter="handleLogin"
      />

      <van-button
        type="primary"
        class="btn-primary"
        :loading="loading"
        loading-text="登入中..."
        @click="handleLogin"
      >
        登入
      </van-button>

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

.login-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.login-form .van-field {
  margin-bottom: 16px;
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
</style>
