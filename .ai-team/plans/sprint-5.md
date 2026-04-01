# Sprint 5 — 核心缺口補齊（圖片上傳 + 推拉流 + 通知）

**狀態**：開發中 🔄
**開始日期**：2026-04-02
**目標**：讓直播推拉流真正可用，補齊頭像/封面上傳，完成 Sprint 4 QA

---

## 任務清單

### BACKEND（圖片上傳 API + 公開用戶資訊）
1. `POST /api/v1/upload/image` — multipart/form-data，支援 avatar/cover 類型，存本地 uploads/
2. `GET /api/v1/users/:id` — 公開用戶資訊（nickname, avatar_url, bio）
3. 靜態檔案服務 `/static/uploads/`
4. 圖片安全驗證（MIME type, 2MB 限制, 隨機檔名）
5. 確認 WebSocket join/leave 事件有廣播「XX 進入直播間」

### FRONTEND（推拉流 + 上傳元件 + 通知）
1. ProfilePage — 頭像上傳（Vant Uploader + 前端裁切）
2. GoLivePage — 封面圖選擇上傳 + 真實攝影機存取（getUserMedia）
3. LiveViewPage — 整合 flv.js 或 hls.js 真實拉流播放
4. 直播結束通知 UI（觀眾端「直播已結束」）
5. 進入/離開通知串接真實 WebSocket 事件
6. 新增 `/user/:id` 路由（公開個人頁）

### DBA
- 不需要 Schema 變更（SA 確認現有欄位足夠）
- 本 Sprint 無 DBA 任務

---

## 驗收標準
- 可以上傳頭像並在個人頁看到
- 開播時可選封面、攝影機預覽為真實畫面
- 觀眾可看到真實串流（非 placeholder）
- 進入直播間時所有觀眾收到「XX 進入直播間」通知
- 直播結束時觀眾收到提示
