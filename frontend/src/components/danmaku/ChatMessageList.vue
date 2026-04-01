<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { ChatMessage } from '../../services/mockWebSocket'

const props = defineProps<{
  messages: ChatMessage[]
}>()

const listRef = ref<HTMLElement | null>(null)

// 自動滾動到底部
watch(() => props.messages.length, async () => {
  await nextTick()
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight
  }
})
</script>

<template>
  <div class="chat-message-list" ref="listRef">
    <div
      v-for="msg in messages.slice(-50)"
      :key="msg.id"
      :class="['chat-msg', msg.type === 'system' ? 'system' : 'user']"
    >
      <template v-if="msg.type === 'system'">
        <span class="system-text">{{ msg.content }}</span>
      </template>
      <template v-else>
        <span class="msg-nickname">{{ msg.nickname }}</span>
        <span class="msg-content">{{ msg.content }}</span>
      </template>
    </div>
  </div>
</template>

<style scoped>
.chat-message-list {
  max-height: 180px;
  overflow-y: auto;
  padding: 8px 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  /* 隱藏捲軸 */
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.chat-message-list::-webkit-scrollbar {
  display: none;
}

.chat-msg {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.4;
  max-width: 80%;
  word-break: break-all;
}

.chat-msg.user {
  background-color: rgba(0, 0, 0, 0.45);
  color: #fff;
}

.chat-msg.system {
  background-color: transparent;
}

.system-text {
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
}

.msg-nickname {
  color: #5eb8ff;
  font-weight: 600;
  margin-right: 6px;
}

.msg-content {
  color: #fff;
}
</style>
