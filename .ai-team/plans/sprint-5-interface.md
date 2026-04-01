# Sprint 5 介面定義

**角色**：SA（系統架構師）
**日期**：2026-04-02
**狀態**：確認版 — BACKEND 與 FRONTEND 的共同合約

---

## 1. 新增 API

### 1.1 POST /api/v1/upload/image

通用圖片上傳端點。**需認證**。

**Request**

- Content-Type: `multipart/form-data`
- Headers: `Authorization: Bearer <access_token>`

| 欄位 | 類型 | 必填 | 說明 |
|------|------|------|------|
| file | file | 是 | 圖片檔案（image/jpeg, image/png） |
| type | string | 是 | `avatar` 或 `cover`，決定存放子目錄 |

**驗證規則**

- 檔案大小上限：2MB
- 允許 MIME type：image/jpeg, image/png（後端用 `net/http.DetectContentType` 二次驗證，不信任 Content-Type header）
- 檔名：後端產生隨機 UUID 檔名，忽略原始檔名（防路徑穿越）
- type 只接受 `avatar` | `cover`，其他值回 400

**存儲路徑**

```
backend/uploads/avatars/{uuid}.{ext}
backend/uploads/covers/{uuid}.{ext}
```

靜態檔案服務掛載在 `/static/uploads/`，最終可存取 URL 例如：
```
/static/uploads/avatars/550e8400-e29b-41d4-a716-446655440000.jpg
```

**Response 200**

```json
{
  "data": {
    "url": "/static/uploads/avatars/550e8400-e29b-41d4-a716-446655440000.jpg"
  }
}
```

**Error Responses**

| HTTP Status | 情境 | Body |
|-------------|------|------|
| 400 | 缺少 file 或 type 參數 | `{"error": "file and type are required"}` |
| 400 | type 不合法 | `{"error": "type must be avatar or cover"}` |
| 400 | 檔案超過 2MB | `{"error": "file size exceeds 2MB limit"}` |
| 400 | 非圖片檔 | `{"error": "only JPEG and PNG images are allowed"}` |
| 401 | 未認證 | `{"error": "unauthorized"}` |
| 500 | 儲存失敗 | `{"error": "internal server error"}` |

---

### 1.2 GET /api/v1/users/:id

取得指定用戶公開資訊。**不需認證**。

**Request**

- Path param: `id` — 用戶 UUID

**Response 200**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "nickname": "小明",
    "avatar_url": "/static/uploads/avatars/xxx.jpg",
    "bio": "大家好"
  }
}
```

注意：與 GET /users/me 的 `UserResponse` 不同，公開用戶資訊**不包含** phone、email、created_at。

**新增 Model**

```go
// PublicUserResponse 公開用戶資訊（不含敏感欄位）
type PublicUserResponse struct {
    ID        uuid.UUID `json:"id"`
    Nickname  string    `json:"nickname"`
    AvatarURL *string   `json:"avatar_url"`
    Bio       *string   `json:"bio"`
}
```

**Error Responses**

| HTTP Status | 情境 | Body |
|-------------|------|------|
| 400 | id 格式不合法 | `{"error": "invalid user id"}` |
| 404 | 用戶不存在 | `{"error": "user not found"}` |

---

## 2. 現有 WebSocket 事件格式確認

WebSocket 端點：`ws://{host}/ws/chat/:stream_id`

連線時需帶 Token（query param `?token=xxx` 或在第一則訊息認證，視現有實作）。

### 2.1 客戶端 → 伺服器（發送彈幕）

```json
{
  "type": "chat",
  "content": "666666"
}
```

### 2.2 伺服器 → 客戶端（廣播彈幕）

```json
{
  "type": "chat",
  "user_id": "550e8400-...",
  "nickname": "小明",
  "content": "666666",
  "timestamp": 1711929600
}
```

### 2.3 伺服器 → 客戶端（系統訊息 — join/leave）

已確認後端 `room.go` 在 join/leave 事件時會廣播系統訊息：

```json
{
  "type": "system",
  "nickname": "系統",
  "content": "小明 進入直播間",
  "timestamp": 1711929600,
  "viewer_count": 42
}
```

```json
{
  "type": "system",
  "nickname": "系統",
  "content": "小明 離開直播間",
  "timestamp": 1711929600,
  "viewer_count": 41
}
```

**前端注意**：`type === "system"` 的訊息用不同的 UI 樣式渲染（系統通知 bar），同時更新本地的 viewer_count 顯示。

### 2.4 伺服器 → 客戶端（直播結束通知）

當主播結束直播時（DELETE /streams/:id 或 SRS on_unpublish），後端需透過 WebSocket 廣播：

```json
{
  "type": "system",
  "nickname": "系統",
  "content": "直播已結束",
  "timestamp": 1711929600,
  "viewer_count": 0
}
```

**BACKEND 任務**：確認 EndStream / on_unpublish 流程中有對 Room 廣播結束訊息後再 Close Room。
**FRONTEND 任務**：收到 content 為「直播已結束」的系統訊息時，顯示全螢幕結束覆蓋層。

---

## 3. 資料庫異動

**無**。SA 已確認：

- `users.avatar_url` (VARCHAR) — 已存在，上傳後存相對 URL
- `streams.cover_url` (VARCHAR) — 已存在，上傳後存相對 URL
- 不需要新增表或欄位

---

## 4. 前後端共識的資料格式

### 4.1 上傳回應格式

所有上傳 API 統一回傳：

```json
{
  "data": {
    "url": "/static/uploads/{type_dir}/{uuid}.{ext}"
  }
}
```

前端拿到 url 後：
- 頭像：呼叫 `PUT /api/v1/users/me` 帶 `{"avatar_url": url}`
- 封面：呼叫 `POST /api/v1/streams` 帶 `{"title": "...", "cover_url": url}`

### 4.2 公開用戶資訊格式

```typescript
// 前端 TypeScript 型別
interface PublicUser {
  id: string
  nickname: string
  avatar_url: string | null
  bio: string | null
}
```

### 4.3 統一錯誤回應格式

所有 API 錯誤回應維持現有格式：

```json
{
  "error": "錯誤訊息字串"
}
```

### 4.4 統一成功回應格式

所有 API 成功回應維持現有格式：

```json
{
  "data": { ... }
}
```

---

## 5. 技術決策

### 5.1 圖片存儲方式

- MVP 階段：存本地磁碟 `backend/uploads/` 目錄
- 後端新增 Gin 靜態檔案中間件，掛載 `/static/uploads/` → `./uploads/`
- 後續可遷移到 S3/OSS，只需改存儲邏輯和 URL 前綴

### 5.2 前端推流方案

**方案：getUserMedia → MediaRecorder → WHIP → SRS**

由於瀏覽器不支援直接 RTMP 推流，且 MVP 需要在 Web 端實現推流，採用以下路線：

1. **攝影機存取**：`navigator.mediaDevices.getUserMedia({ video: true, audio: true })`
2. **推流到 SRS**：使用 SRS 內建的 WebRTC/WHIP 支援
   - SRS 6 支援 WHIP (WebRTC-HTTP Ingestion Protocol)
   - 前端用 RTCPeerConnection 建立 WebRTC 連線，透過 WHIP endpoint 推流
   - WHIP URL：`http://{srs_host}:1985/rtc/v1/whip/?app=live&stream={stream_key}`
3. **備選方案**：若 WHIP 整合有困難，可退回「提示用戶用 OBS 推 RTMP」的方案

**前端實作重點**：
- GoLivePage 用 getUserMedia 取得真實攝影機預覽（替換現有 placeholder）
- 按下「開播」後，先呼叫 `POST /api/v1/streams` 取得 stream_key
- 用 stream_key 組合 WHIP URL，建立 RTCPeerConnection 推流
- 推流狀態顯示：連線中 / 直播中 / 連線失敗

### 5.3 前端拉流方案

**方案：mpegts.js (HTTP-FLV) 為主，HLS 為備援**

1. **桌面瀏覽器 + Android**：使用 mpegts.js 播放 HTTP-FLV
   - FLV URL：從 `GET /api/v1/streams/:id` 回應的 `flv_url` 取得
   - 延遲 1-3 秒，體驗接近即時
2. **iOS Safari**：使用原生 `<video>` 標籤播放 HLS
   - HLS URL：從 `GET /api/v1/streams/:id` 回應的 `hls_url` 取得
   - iOS Safari 不支援 MSE，故無法用 mpegts.js
3. **判斷邏輯**：前端根據 UA 或 MSE 支援度自動切換
   ```typescript
   const canUseFLV = typeof MediaSource !== 'undefined'
     && MediaSource.isTypeSupported('video/x-flv')
   ```
   若 `canUseFLV` 為 false，fallback 到 HLS。

**前端實作重點**：
- LiveViewPage 引入 mpegts.js（npm 安裝）
- 建立 player helper：根據環境選擇 FLV 或 HLS
- 播放失敗時顯示重試按鈕
- 直播結束時銷毀 player 實例，顯示結束畫面

---

## 6. 路由變更摘要

### 後端新增路由

```
POST   /api/v1/upload/image     → UploadHandler.UploadImage  [需認證]
GET    /api/v1/users/:id        → UserHandler.GetUser        [不需認證]
GET    /static/uploads/*path    → gin.Static                 [不需認證]
```

### 前端新增路由

```
/user/:id    → PublicProfilePage（或 ProfilePage 的公開模式）
```

---

## 7. 依賴清單

### 後端

- 無新依賴（圖片上傳用 Go 標準庫即可）

### 前端

- `mpegts.js` — HTTP-FLV 播放器（拉流用）
- 其他皆為現有依賴（Vant Uploader 已內含於 Vant 4）
