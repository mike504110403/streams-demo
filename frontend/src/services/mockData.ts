export interface StreamItem {
  id: string
  title: string
  cover_url: string
  host_nickname: string
  host_avatar: string
  viewer_count: number
  status: 'live' | 'replay' | 'upcoming'
}

const coverImages = [
  'https://picsum.photos/seed/stream1/400/600',
  'https://picsum.photos/seed/stream2/400/600',
  'https://picsum.photos/seed/stream3/400/600',
  'https://picsum.photos/seed/stream4/400/600',
  'https://picsum.photos/seed/stream5/400/600',
  'https://picsum.photos/seed/stream6/400/600',
  'https://picsum.photos/seed/stream7/400/600',
  'https://picsum.photos/seed/stream8/400/600',
  'https://picsum.photos/seed/stream9/400/600',
  'https://picsum.photos/seed/stream10/400/600',
  'https://picsum.photos/seed/stream11/400/600',
  'https://picsum.photos/seed/stream12/400/600',
  'https://picsum.photos/seed/stream13/400/600',
  'https://picsum.photos/seed/stream14/400/600',
  'https://picsum.photos/seed/stream15/400/600',
]

const avatarImages = [
  'https://picsum.photos/seed/avatar1/100/100',
  'https://picsum.photos/seed/avatar2/100/100',
  'https://picsum.photos/seed/avatar3/100/100',
  'https://picsum.photos/seed/avatar4/100/100',
  'https://picsum.photos/seed/avatar5/100/100',
  'https://picsum.photos/seed/avatar6/100/100',
  'https://picsum.photos/seed/avatar7/100/100',
  'https://picsum.photos/seed/avatar8/100/100',
]

export const mockStreams: StreamItem[] = [
  {
    id: '1',
    title: '深夜唱歌給你聽 🎤',
    cover_url: coverImages[0],
    host_nickname: '小美',
    host_avatar: avatarImages[0],
    viewer_count: 1283,
    status: 'live',
  },
  {
    id: '2',
    title: '吃播！今天挑戰超大碗拉麵',
    cover_url: coverImages[1],
    host_nickname: '大胃王阿傑',
    host_avatar: avatarImages[1],
    viewer_count: 856,
    status: 'live',
  },
  {
    id: '3',
    title: '教你畫水彩風景畫',
    cover_url: coverImages[2],
    host_nickname: '畫畫老師',
    host_avatar: avatarImages[2],
    viewer_count: 432,
    status: 'live',
  },
  {
    id: '4',
    title: '打遊戲聊天室～來陪我',
    cover_url: coverImages[3],
    host_nickname: 'GameMaster',
    host_avatar: avatarImages[3],
    viewer_count: 2105,
    status: 'live',
  },
  {
    id: '5',
    title: '瑜伽晨間課 - 一起伸展',
    cover_url: coverImages[4],
    host_nickname: '瑜伽女孩',
    host_avatar: avatarImages[4],
    viewer_count: 678,
    status: 'live',
  },
  {
    id: '6',
    title: '街頭表演 LIVE',
    cover_url: coverImages[5],
    host_nickname: '街頭藝人小王',
    host_avatar: avatarImages[5],
    viewer_count: 345,
    status: 'live',
  },
  {
    id: '7',
    title: '下午茶聊天～今天心情好',
    cover_url: coverImages[6],
    host_nickname: '甜甜',
    host_avatar: avatarImages[6],
    viewer_count: 189,
    status: 'live',
  },
  {
    id: '8',
    title: '程式教學：Vue 3 入門',
    cover_url: coverImages[7],
    host_nickname: '碼農大叔',
    host_avatar: avatarImages[7],
    viewer_count: 523,
    status: 'live',
  },
  {
    id: '9',
    title: '昨天的演唱會回放',
    cover_url: coverImages[8],
    host_nickname: '小美',
    host_avatar: avatarImages[0],
    viewer_count: 3201,
    status: 'replay',
  },
  {
    id: '10',
    title: '料理教室：家常菜篇',
    cover_url: coverImages[9],
    host_nickname: '主廚阿明',
    host_avatar: avatarImages[1],
    viewer_count: 912,
    status: 'live',
  },
  {
    id: '11',
    title: '戶外露營 VLOG',
    cover_url: coverImages[10],
    host_nickname: '露營達人',
    host_avatar: avatarImages[2],
    viewer_count: 267,
    status: 'live',
  },
  {
    id: '12',
    title: '明天下午 3 點開播預告',
    cover_url: coverImages[11],
    host_nickname: 'GameMaster',
    host_avatar: avatarImages[3],
    viewer_count: 0,
    status: 'upcoming',
  },
  {
    id: '13',
    title: '手作飾品教學',
    cover_url: coverImages[12],
    host_nickname: '手作小姐',
    host_avatar: avatarImages[4],
    viewer_count: 156,
    status: 'live',
  },
  {
    id: '14',
    title: '健身重訓直播',
    cover_url: coverImages[13],
    host_nickname: '肌肉男',
    host_avatar: avatarImages[5],
    viewer_count: 489,
    status: 'live',
  },
  {
    id: '15',
    title: '睡前故事時間',
    cover_url: coverImages[14],
    host_nickname: '故事姐姐',
    host_avatar: avatarImages[6],
    viewer_count: 734,
    status: 'live',
  },
]

/**
 * 模擬 API 取得直播列表（分頁）
 */
export function fetchMockStreams(page: number, pageSize: number = 6): Promise<{
  list: StreamItem[]
  hasMore: boolean
  total: number
}> {
  return new Promise((resolve) => {
    setTimeout(() => {
      const start = (page - 1) * pageSize
      const end = start + pageSize
      const list = mockStreams.slice(start, end)
      resolve({
        list,
        hasMore: end < mockStreams.length,
        total: mockStreams.length,
      })
    }, 600) // 模擬網路延遲
  })
}

/**
 * 模擬 API 取得單一直播詳情
 */
export function fetchMockStreamById(id: string): Promise<StreamItem | null> {
  return new Promise((resolve) => {
    setTimeout(() => {
      const stream = mockStreams.find((s) => s.id === id) || null
      resolve(stream)
    }, 300)
  })
}
