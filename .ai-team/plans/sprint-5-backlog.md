# Sprint 5 — 功能補齊 Backlog

**撰寫者**：PM Agent
**日期**：2026-04-02
**依據**：requirements.md 功能清單 vs Sprint 1~4 實作現況

---

## 盤點方法

逐一比對 requirements.md 中的功能清單，與 Sprint 1~4 的 commit 紀錄和程式碼現況，找出尚未實作或僅有 UI 殼但缺真實功能的項目。

---

## 待補齊功能

### P1（重要）

- [ ] **頭像上傳功能** — 後端缺少檔案上傳 API（目前 UpdateMe 僅接受 JSON 的 avatar_url），前端 ProfilePage 也沒有頭像上傳 UI（僅能編輯暱稱和簡介）。需要：後端新增 POST /api/v1/users/me/avatar（multipart）+ 前端頭像選擇/裁切/上傳元件 — **BACKEND + FRONTEND**

- [ ] **開播推流功能** — GoLivePage 目前只是靜態 UI 殼（攝影機預覽區是假的 placeholder），缺少真實的 getUserMedia 攝影機存取、WebRTC/RTMP 推流到 SRS 的實作。這是直播平台核心功能 — **FRONTEND + SA（確認推流方案）**

- [ ] **設定直播封面圖** — requirements 要求「設定直播標題與封面圖」，目前只實作了標題輸入，缺少封面圖上傳。後端 streams 表需確認是否有 cover_image 欄位 — **FRONTEND + BACKEND**

- [ ] **直播間進入/離開通知（真實版）** — 後端 WebSocket 的 join/leave 事件需確認是否有廣播「XX 進入直播間」給所有觀眾。前端 SystemNotice 元件已有 UI，但 realWebSocket.ts 需確認是否正確處理系統訊息類型 — **BACKEND + FRONTEND**

- [ ] **觀看直播（真實拉流播放）** — LiveViewPage 需串接真實的 HLS/HTTP-FLV 播放器（如 flv.js 或 hls.js），目前可能仍為 placeholder。需要確認 SRS 輸出格式並整合播放器 — **FRONTEND**

- [ ] **Sprint 4 QA 驗收** — Sprint 4 狀態為「開發完成，待 QA」，完整用戶流程測試、WebSocket 連線穩定性測試尚未執行 — **QA**

### P2（次要）

- [ ] **直播結束通知** — 觀眾端應顯示「直播已結束」提示（requirements 驗收標準 3.2），需確認後端 on_unpublish 回調是否通知觀眾、前端是否有對應 UI — **BACKEND + FRONTEND**

- [ ] **直播列表依觀看人數排序** — 後端 GET /streams 的排序邏輯需確認是否已實作 viewer_count 降序。前端大廳是否傳入排序參數 — **BACKEND + FRONTEND**

- [ ] **未登入用戶限制發彈幕** — requirements 要求未登入只能觀看不能發彈幕，需確認前端是否隱藏輸入框、後端 WebSocket 是否驗證 Token — **FRONTEND + BACKEND**

- [ ] **跨裝置響應式適配** — requirements 1.5 要求支援 Mobile 裝置，目前未發現任何 @media query 或響應式處理。Vant 本身是 mobile-first，但頁面佈局可能在不同螢幕尺寸下有問題 — **FRONTEND**

- [ ] **封面圖顯示在直播列表** — 大廳卡片需顯示封面、標題、主播暱稱、觀看人數（requirements 驗收標準 3.3），需確認封面是否正確顯示 — **FRONTEND**

- [ ] **空狀態 UI** — 直播大廳無直播時應顯示「目前沒有直播，稍後再來看看」（requirements 驗收標準 3.3） — **FRONTEND**

---

## 建議 Sprint 拆分

功能較多，建議拆為 Sprint 5 + Sprint 6：

### Sprint 5 — 核心缺口補齊（推流/拉流 + QA）
重點：讓直播的「推」和「拉」真正能用，跑完 QA
1. SA 確認推流方案（瀏覽器端 WebRTC → SRS，或其他方案）
2. 前端實作真實攝影機存取 + 推流
3. 前端整合 HLS/FLV 播放器（真實拉流）
4. 直播結束通知（觀眾端）
5. Sprint 4 遺留的 QA 驗收一併完成

### Sprint 6 — 細節打磨 + 個人資料完善
重點：補齊 UX 細節，完成 MVP 所有驗收標準
1. 頭像上傳（後端 API + 前端 UI）
2. 直播封面圖上傳 + 顯示
3. 未登入用戶限制
4. 跨裝置響應式適配
5. 直播列表排序確認
6. 空狀態 UI
7. 最終 MVP 全流程 QA
