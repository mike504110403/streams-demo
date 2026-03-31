<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const activeTab = ref(0)

function onTabChange(index: number) {
  if (index === 2) {
    if (authStore.isAuthenticated) {
      router.push('/profile')
    } else {
      router.push('/login')
    }
  }
}
</script>

<template>
  <div class="home-page">
    <div class="home-content">
      <div class="placeholder-section">
        <div class="logo-icon">
          <van-icon name="video-o" size="64" color="#fe2c55" />
        </div>
        <h1 class="welcome-title">Streams</h1>
        <p class="welcome-subtitle">Live your moment</p>
        <p class="coming-soon">直播大廳即將上線</p>
        <p class="sprint-info">Sprint 2 開發中...</p>
      </div>
    </div>

    <van-tabbar v-model="activeTab" @change="onTabChange">
      <van-tabbar-item icon="home-o">首頁</van-tabbar-item>
      <van-tabbar-item icon="video-o">開播</van-tabbar-item>
      <van-tabbar-item icon="user-o">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  background-color: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

.home-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-bottom: 50px;
}

.placeholder-section {
  text-align: center;
  padding: 0 32px;
}

.logo-icon {
  margin-bottom: 24px;
}

.welcome-title {
  font-size: 48px;
  font-weight: 800;
  color: var(--text-primary);
  margin-bottom: 8px;
  letter-spacing: -1.5px;
}

.welcome-subtitle {
  font-size: 18px;
  color: var(--text-secondary);
  margin-bottom: 48px;
}

.coming-soon {
  font-size: 16px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.sprint-info {
  font-size: 13px;
  color: var(--text-muted);
  opacity: 0.6;
}
</style>
