# Sprint 4 — 全流程串接 + 整合測試

**狀態**：待啟動
**目標**：後端完成 WebSocket 彈幕服務，全功能串接，QA 驗收

---

## 任務分配

### BACKEND
- [ ] WebSocket 彈幕服務（ws/chat/:stream_id）
- [ ] 聊天訊息廣播（Room 機制）
- [ ] 進入/離開直播間通知
- [ ] 觀看人數即時更新（Redis INCR/DECR）
- [ ] 聊天訊息持久化（寫入 chat_messages 表）

### FRONTEND
- [ ] 彈幕 Mock 替換為真實 WebSocket 連線
- [ ] 觀看人數從 API 即時取得
- [ ] 全流程 UI 細節修正（根據前幾個 Sprint 老闆回饋）
- [ ] 跨裝置最終適配

### QA
- [ ] 完整測試計畫執行
- [ ] 用戶流程：註冊 → 登入 → 瀏覽大廳 → 觀看直播 → 發彈幕 → 開播
- [ ] 跨瀏覽器測試（Chrome、Safari）
- [ ] 效能基本驗證（WebSocket 連線穩定性）
- [ ] Bug 回報 + 修復循環

---

## 驗收標準
- 完整 MVP 流程可跑通：註冊 → 登入 → 大廳 → 觀看 → 彈幕 → 開播
- QA 通過所有 P0 測試案例
- 跨裝置基本可用
