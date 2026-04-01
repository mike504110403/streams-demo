import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'
import type { StreamItem } from '../services/mockData'

export const useStreamStore = defineStore('stream', () => {
  // === State ===
  const streams = ref<StreamItem[]>([])
  const currentStream = ref<StreamItem | null>(null)
  const page = ref(1)
  const hasMore = ref(true)
  const loading = ref(false)
  const refreshing = ref(false)

  const PAGE_SIZE = 6

  // === Actions ===

  /**
   * 載入直播列表（下一頁）
   */
  async function loadStreams() {
    if (loading.value || !hasMore.value) return

    loading.value = true
    try {
      const { data } = await api.get('/streams', {
        params: { page: page.value, limit: PAGE_SIZE },
      })
      const result = data.data
      streams.value.push(...result.streams)
      hasMore.value = page.value * PAGE_SIZE < result.total
      page.value++
    } finally {
      loading.value = false
    }
  }

  /**
   * 刷新直播列表（重頭開始）
   */
  async function refreshStreams() {
    refreshing.value = true
    try {
      page.value = 1
      hasMore.value = true
      const { data } = await api.get('/streams', {
        params: { page: 1, limit: PAGE_SIZE },
      })
      const result = data.data
      streams.value = result.streams
      hasMore.value = PAGE_SIZE < result.total
      page.value = 2
    } finally {
      refreshing.value = false
    }
  }

  /**
   * 載入單一直播詳情
   */
  async function loadStreamById(id: string) {
    const { data } = await api.get(`/streams/${id}`)
    currentStream.value = data.data
  }

  /**
   * 清除當前直播
   */
  function clearCurrentStream() {
    currentStream.value = null
  }

  return {
    // State
    streams,
    currentStream,
    page,
    hasMore,
    loading,
    refreshing,
    // Actions
    loadStreams,
    refreshStreams,
    loadStreamById,
    clearCurrentStream,
  }
})
