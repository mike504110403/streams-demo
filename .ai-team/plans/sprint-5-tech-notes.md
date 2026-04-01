# Sprint 5 — 技術補充筆記

**SA 撰寫**｜2026-04-02
**目的**：對照 requirements.md 中尚未完成的功能，列出架構層面需補充的項目

---

## 現況摘要

Sprint 1-4 已完成：
- 認證系統（手機驗證碼 + Apple/Google OAuth + JWT + refresh token）
- 用戶模組（GET/PUT /users/me）
- 直播模組（CRUD + SRS 回調 + 直播列表分頁）
- 彈幕模組（WebSocket + 聊天歷史 API + 觀看人數 API）
- 前端全部頁面 UI + Mock/Real 切換機制

Sprint 4 待完成：QA 驗收（4.5 整合測試 + 4.6 QA）

---

## 未完成功能對照

| requirements.md 項目 | 狀態 | Sprint 5 需處理 |
|---|---|---|
| 頭像上傳（JPG/PNG < 2MB，自動裁切正方形） | 缺 — 目前只接受 URL 字串，無檔案上傳 | 是 |
| 封面圖上傳（直播開播時設定封面） | 缺 — GoLivePage 有 UI 但無上傳邏輯 | 是 |
| GET /users/:id 公開用戶資訊 | 缺 — architecture.md 有定義但未實作 | 是 |
| 直播列表依觀看人數排序 | 需確認 — 後端 ListStreams 是否真的用 viewer_count 排序 | 驗證 |
| Sprint 4 QA 驗收 | 未完成 | 是 |

---

## 新增/修改的 API

- `POST /api/v1/upload/image` — POST — 通用圖片上傳端點，接受 multipart/form-data，回傳圖片 URL。用於頭像和封面圖上傳。參數：file（圖片檔）、type（avatar / cover，決定裁切邏輯和存放路徑）
- `GET /api/v1/users/:id` — GET — 取得指定用戶公開資訊（nickname、avatar_url、bio），不含 phone/email 等敏感欄位。不需認證

## DB Schema 變更

- 不需要新增表或欄位。現有 users.avatar_url 和 streams.cover_url 已是 VARCHAR 型別，上傳後存 URL 即可
- 上傳的檔案 MVP 階段存在本地磁碟（backend/uploads/ 目錄），後端提供靜態檔案服務。後續可遷移到 S3/OSS

## 前端變更

- **ProfilePage.vue** — 頭像區塊新增點擊上傳功能，呼叫 upload API 後拿到 URL 再 PUT /users/me 更新 avatar_url。需整合圖片裁切（正方形）邏輯，可用 Vant 的 Uploader 元件
- **GoLivePage.vue** — 封面選擇器串接上傳 API，選取圖片後上傳拿到 URL，建立直播間時帶入 cover_url
- **路由** — 新增 `/user/:id` 路由，顯示其他用戶的公開個人頁（可複用 ProfilePage 元件，區分「自己」和「他人」模式）
- **LiveViewPage.vue / LiveHallPage.vue** — 主播頭像/暱稱處新增點擊跳轉到 `/user/:id` 的邏輯

## 技術風險

- **圖片上傳安全**：需驗證檔案 MIME type（只允許 image/jpeg、image/png）、限制檔案大小（2MB）、產生隨機檔名避免路徑穿越攻擊。後端用 Go 的 `net/http.DetectContentType` 做二次驗證，不能只信任 Content-Type header
- **本地儲存路徑暴露**：靜態檔案服務需確保只能存取 uploads/ 目錄，不能讀到其他路徑。建議掛載在 `/static/uploads/` 前綴下
- **圖片裁切效能**：MVP 階段頭像裁切建議在前端完成（canvas API 或 Vant Cropper），避免後端引入圖片處理依賴。後端只做儲存
- **Sprint 4 QA 尚未執行**：Sprint 5 開發前應先完成 Sprint 4 的 QA 驗收，確保基礎流程穩定。建議 Sprint 5 前半段先跑 QA，修完 Bug 再開始新功能開發
