# Sprint 3 — 彈幕 UI + 後端直播 API 完整實作

**狀態**：已完成 ✅
**目標**：前端完成彈幕互動 UI；後端補齊直播 API 並串接 SRS

---

## 策略說明

前端繼續先行做彈幕 UI（Mock WebSocket），同時後端追上完成直播模組的完整 API。
Sprint 2 的直播 UI 在本 Sprint 與後端串接。

---

## 任務分配

### FRONTEND（繼續先行）
- [ ] 彈幕 UI（訊息列表 + 飄屏效果）
- [ ] 彈幕輸入框
- [ ] 「XX 進入直播間」系統訊息 UI
- [ ] 觀看人數即時顯示 UI
- [ ] Mock WebSocket 資料層（模擬彈幕收發）
- [ ] 跨裝置響應式適配調整

### DBA
- [ ] chat_messages 表 migration

### BACKEND（追上直播模組）
- [ ] POST /api/v1/streams — 建立直播間（產生 stream_key + 推流地址）
- [ ] GET /api/v1/streams — 直播列表（分頁、按觀看人數排序）
- [ ] GET /api/v1/streams/:id — 直播間詳情（含拉流地址）
- [ ] PUT /api/v1/streams/:id — 更新直播間資訊
- [ ] DELETE /api/v1/streams/:id — 結束直播
- [ ] SRS HTTP Callback 處理（on_publish / on_unpublish）
- [ ] Redis 快取：直播列表、線上人數
- [ ] **串接 Sprint 2 前端的直播 UI**（替換 Mock → 真實 API）

---

## 驗收標準
- 前端彈幕 UI：老闆可看到彈幕效果（Mock 資料驅動）
- 後端直播 API：可用 OBS 推流 → API 查到直播 → 前端大廳顯示真實直播
- Sprint 2 的直播 UI 已串接後端 API
- **彈幕尚未串接後端**（下一 Sprint）
