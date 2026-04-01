import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchMockStreams, fetchMockStreamById, type StreamItem } from '../services/mockData'

export const useStreamStore = defineStore('stream', () => {
  // === State ===
  const streams = ref<StreamItem[]>([])
  const currentStream = ref<StreamItem | null>(null)
  const page = ref(1)
  const hasMore = ref(true)
  const loading = ref(false)
  const refreshing = ref(false)

  // === Actions ===

  /**
   * 載入直播列表（下一頁）
   */
  async function loadStreams() {
    if (loading.value || !hasMore.value) return

    loading.value = true
    try {
      const result = await fetchMockStreams(page.value)
      streams.value.push(...result.list)
      hasMore.value = result.hasMore
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
      const result = await fetchMockStreams(1)
      streams.value = result.list
      hasMore.value = result.hasMore
      page.value = 2
    } finally {
      refreshing.value = false
    }
  }

  /**
   * 載入單一直播詳情
   */
  async function loadStreamById(id: string) {
    currentStream.value = await fetchMockStreamById(id)
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
