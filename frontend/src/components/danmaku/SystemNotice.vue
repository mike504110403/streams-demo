<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import type { ChatMessage } from '../../services/mockWebSocket'

const props = defineProps<{
  messages: ChatMessage[]
}>()

interface NoticeItem {
  id: string
  text: string
}

const notices = ref<NoticeItem[]>([])
let lastProcessed = 0

watch(() => props.messages.length, () => {
  const newMessages = props.messages.slice(lastProcessed)
  lastProcessed = props.messages.length

  for (const msg of newMessages) {
    if (msg.type !== 'system') continue
    showNotice(msg)
  }
})

function showNotice(msg: ChatMessage) {
  const item: NoticeItem = {
    id: msg.id,
    text: msg.content,
  }
  notices.value.push(item)

  // 3 秒後移除
  setTimeout(() => {
    const idx = notices.value.findIndex(n => n.id === item.id)
    if (idx !== -1) notices.value.splice(idx, 1)
  }, 3000)
}

onUnmounted(() => {
  notices.value = []
})
</script>

<template>
  <div class="system-notice-container">
    <TransitionGroup name="notice">
      <div
        v-for="notice in notices.slice(-3)"
        :key="notice.id"
        class="notice-item"
      >
        <van-icon name="friends-o" size="14" color="#ffd700" />
        <span>{{ notice.text }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.system-notice-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
  pointer-events: none;
}

.notice-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 14px;
  background: linear-gradient(90deg, rgba(255, 215, 0, 0.2) 0%, rgba(255, 215, 0, 0.05) 100%);
  font-size: 12px;
  color: rgba(255, 255, 255, 0.8);
  width: fit-content;
}

/* Transition */
.notice-enter-active {
  transition: all 0.3s ease-out;
}

.notice-leave-active {
  transition: all 0.3s ease-in;
}

.notice-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.notice-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
</style>
