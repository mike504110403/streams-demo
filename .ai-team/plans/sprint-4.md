# Sprint 4 — 全流程串接

**狀態**：開發完成，待 QA 🔄
**目標**：前端從 Mock 模式切換到真實後端，完成 MVP 全流程

---

## 策略說明

後端 WebSocket、聊天持久化、SRS 回調已在 Sprint 3 完成。
前端 UI + Mock 資料層也完成。
本 Sprint 純粹做「接線」+ 補缺口，不新增大功能。

切成 6 個獨立子任務，避免超時：

---

## 子任務

### 4.1 Backend: 補齊 REST API 缺口 ✅（Sprint 3 已完成）
- [x] GET /api/v1/streams/:id/messages — 聊天歷史（分頁）
- [x] GET /api/v1/streams/:id/viewers — 即時觀看人數（從 Hub 取）
- [x] 註冊路由 + 測試

### 4.2 Frontend: 實作真實 WebSocket Service ✅
- [x] 新增 realWebSocket.ts（ws://host/ws/chat/:stream_id）
- [x] Token 認證（query param）
- [x] 斷線重連機制（指數退避，最大 30s，10 次上限）
- [x] 與 MockWebSocket 相同介面

### 4.3 Frontend: Chat Store 串接 ✅
- [x] chat.ts 根據 VITE_MOCK_API 環境變數選擇 Mock/Real WebSocket
- [x] 訊息格式對齊後端（user_id → userId, timestamp 秒→毫秒）
- [x] LiveViewPage.vue 呼叫端更新（傳入 streamId）

### 4.4 Frontend: 關閉 Mock 模式 + 環境設定 ✅
- [x] .env.development（VITE_MOCK_API=false, VITE_API_BASE, VITE_WS_BASE）
- [x] .env.mock（VITE_MOCK_API=true，供純前端開發用）
- [x] api.ts baseURL 改用 VITE_API_BASE 環境變數
- [x] api.ts refresh token URL 也改用環境變數
- [x] 新增 npm script: `dev:mock`（vite --mode mock）

### 4.5 整合驗證 + Bug 修復 ✅
- [x] 前端 TypeScript 編譯通過（零錯誤）
- [x] 後端 Go build 通過
- [ ] 啟動後端 + 前端，跑完整流程（待 QA）

### 4.6 QA 驗收
- [ ] 完整用戶流程測試
- [ ] WebSocket 連線穩定性
- [ ] Bug 回報 + 修復

---

## 變更檔案清單
- `frontend/src/services/realWebSocket.ts` — 新增，真實 WebSocket 客戶端
- `frontend/src/stores/chat.ts` — 重構，支援 Mock/Real 切換
- `frontend/src/services/api.ts` — baseURL 改用環境變數
- `frontend/src/views/LiveViewPage.vue` — connect() 傳入 streamId
- `frontend/.env.development` — 新增，開發環境設定
- `frontend/.env.mock` — 新增，Mock 模式設定
- `frontend/package.json` — 新增 dev:mock script

---

## 驗收標準
- 註冊 → 登入 → 大廳 → 點入直播間 → 即時彈幕 → 發送彈幕 — 全流程跑通
- VITE_MOCK_API=false 時所有功能正常
- WebSocket 斷線可重連
