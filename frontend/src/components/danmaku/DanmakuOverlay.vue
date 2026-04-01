<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import type { ChatMessage } from '../../services/mockWebSocket'

const props = defineProps<{
  messages: ChatMessage[]
}>()

interface DanmakuItem {
  id: string
  text: string
  top: number    // 百分比位置
  duration: number // 秒
}

const danmakuList = ref<DanmakuItem[]>([])
const TRACK_COUNT = 8  // 彈幕軌道數
let lastProcessed = 0
let trackNextAvailable: number[] = new Array(TRACK_COUNT).fill(0)

watch(() => props.messages.length, () => {
  const newMessages = props.messages.slice(lastProcessed)
  lastProcessed = props.messages.length

  for (const msg of newMessages) {
    if (msg.type !== 'chat') continue
    spawnDanmaku(msg)
  }
})

function spawnDanmaku(msg: ChatMessage) {
  const now = Date.now()
  // 找最早可用的軌道
  let bestTrack = 0
  let bestTime = trackNextAvailable[0]
  for (let i = 1; i < TRACK_COUNT; i++) {
    if (trackNextAvailable[i] < bestTime) {
      bestTime = trackNextAvailable[i]
      bestTrack = i
    }
  }

  const duration = 5 + Math.random() * 3 // 5~8 秒
  // 標記此軌道在一段時間後才可再用（避免重疊）
  trackNextAvailable[bestTrack] = now + 800

  const item: DanmakuItem = {
    id: msg.id,
    text: `${msg.nickname}: ${msg.content}`,
    top: (bestTrack / TRACK_COUNT) * 60 + 5, // 5%~65% 範圍
    duration,
  }

  danmakuList.value.push(item)

  // 動畫結束後移除
  setTimeout(() => {
    const idx = danmakuList.value.findIndex(d => d.id === item.id)
    if (idx !== -1) danmakuList.value.splice(idx, 1)
  }, duration * 1000 + 200)
}

onUnmounted(() => {
  danmakuList.value = []
})
</script>

<template>
  <div class="danmaku-overlay">
    <div
      v-for="item in danmakuList"
      :key="item.id"
      class="danmaku-item"
      :style="{
        top: item.top + '%',
        animationDuration: item.duration + 's',
      }"
    >
      {{ item.text }}
    </div>
  </div>
</template>

<style scoped>
.danmaku-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  overflow: hidden;
  z-index: 5;
}

.danmaku-item {
  position: absolute;
  white-space: nowrap;
  font-size: 15px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8), 0 0 4px rgba(0, 0, 0, 0.5);
  animation: danmaku-fly linear forwards;
  left: 100%;
  pointer-events: none;
}

@keyframes danmaku-fly {
  from {
    transform: translateX(0);
  }
  to {
    transform: translateX(calc(-100% - 100vw));
  }
}
</style>
