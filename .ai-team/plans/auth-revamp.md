# 註冊登入改版需求文件

**專案名稱**：streams-demo（類抖音直播平台）
**文件類型**：功能改版需求
**撰寫者**：PM Agent
**日期**：2026-04-01（社交登入需求更新）
**狀態**：✅ 老闆已確認 (2026-04-01)
**版本**：v2 — 新增 Apple / Google 社交登入

---

## 1. 變更概要

### 1.1 改版原因

需求規格書（requirements.md）明確要求使用「手機號碼 + SMS 驗證碼」方式進行註冊登入，但目前實作為「email + 密碼」方式，與規格不符。本次改版將認證機制全面切換至手機號碼 + SMS 驗證碼。

### 1.2 變更範圍

- **認證方式**：email + 密碼 → 手機號碼 + SMS 驗證碼 + Apple 登入 + Google 登入
- **影響層面**：前端 UI、後端 API、資料庫 Schema
- **MVP 階段**：SMS 驗證碼使用 mock 方式（不串接真實簡訊商），驗證碼固定為 `123456` 並同時在 server console 印出
- **社交登入 MVP 階段**：先做 Web 版本，iOS/Android 原生 SDK 整合留待 App 開發階段

### 1.2.1 登入方式平台支援矩陣

| 登入方式 | Web | iOS | Android |
|----------|-----|-----|---------|
| 手機號碼 SMS | ✅ | ✅ | ✅ |
| Apple 登入 | ✅ | ✅ | ❌ |
| Google 登入 | ✅ | ✅ | ✅ |

**裝置偵測規則**：程式自動偵測裝置類型（iOS / Android / Web），根據平台動態顯示對應的登入選項。Android 裝置不顯示 Apple 登入按鈕。

### 1.3 現況盤點

目前已實作的 email + 密碼認證相關檔案：

| 層級 | 檔案 | 現況 |
|------|------|------|
| 前端頁面 | `frontend/src/views/LoginPage.vue` | email + 密碼登入表單 |
| 前端頁面 | `frontend/src/views/RegisterPage.vue` | 暱稱 + email + 密碼 + 確認密碼表單 |
| 前端 Store | `frontend/src/stores/auth.ts` | LoginPayload(email, password)、RegisterPayload(email, password, nickname) |
| 後端 Handler | `backend/internal/handler/auth_handler.go` | Register、Login、RefreshToken、Logout |
| 後端 Service | `backend/internal/service/auth_service.go` | bcrypt 密碼加密、email 查詢登入 |
| 後端 Model | `backend/internal/model/user.go` | RegisterRequest(email, password, nickname)、LoginRequest(email, password) |
| DB Schema | users 表 | email(必填)、phone(選填)、password_hash(必填) |

---

## 2. 使用者流程

### 2.1 註冊流程

```
[步驟 1] 輸入手機號碼
         ↓
[步驟 2] 點擊「發送驗證碼」按鈕
         ↓ 按鈕變為 60 秒倒數計時，期間不可再次點擊
[步驟 3] 輸入 6 位數驗證碼
         ↓ MVP mock：固定 123456，同時 server console log 輸出
[步驟 4] 輸入暱稱
         ↓
[步驟 5] 點擊「註冊」
         ↓ 後端驗證驗證碼 → 建立帳號 → 回傳 Token
[完成]   註冊成功 → 自動登入 → 跳轉首頁
```

**UI 建議**：可做成兩步驟表單——
- 第一步：手機號碼 + 發送驗證碼 + 驗證碼輸入
- 第二步：輸入暱稱 + 完成註冊

### 2.2 登入流程

```
[步驟 1] 輸入手機號碼
         ↓
[步驟 2] 點擊「發送驗證碼」按鈕
         ↓ 按鈕變為 60 秒倒數計時
[步驟 3] 輸入 6 位數驗證碼
         ↓
[步驟 4] 點擊「登入」
         ↓ 後端驗證驗證碼 → 查詢用戶 → 回傳 Token
[完成]   登入成功 → 跳轉首頁（或原本要前往的頁面）
```

### 2.3 社交登入流程（Apple / Google）

社交登入採用 OAuth 流程，登入與註冊合一：新用戶自動建立帳號，舊用戶直接登入，無需額外註冊步驟。

#### 2.3.1 Apple 登入流程

```
[步驟 1] 使用者點擊「透過 Apple 登入」按鈕
         ↓
[步驟 2] 跳轉至 Apple OAuth 授權頁面（Sign in with Apple JS / REST API）
         ↓ 使用者授權（首次登入時 Apple 提供 email、姓名）
[步驟 3] Apple 回傳 authorization code + id_token 至前端
         ↓
[步驟 4] 前端將 authorization code + id_token 發送至後端
         ↓ 後端驗證 token → 取得 Apple user ID (sub)
[判斷]   查詢 user_oauth_providers 表：
         ├─ 已有帳號 → 直接登入 → 回傳 Token
         └─ 無帳號 → 自動建立帳號（暱稱預設為 Apple 提供的姓名或「Apple 用戶」）→ 回傳 Token
[完成]   登入成功 → 跳轉首頁
```

**Apple 登入注意事項**：
- Apple 僅在用戶**首次授權**時提供 email 和姓名，後續登入不再提供，後端需在首次時存好
- MVP 階段使用 Sign in with Apple JS（Web），不需要 iOS 原生 SDK

#### 2.3.2 Google 登入流程

```
[步驟 1] 使用者點擊「透過 Google 登入」按鈕
         ↓
[步驟 2] 跳轉至 Google OAuth 授權頁面（Google Identity Services）
         ↓ 使用者授權
[步驟 3] Google 回傳 authorization code / credential 至前端
         ↓
[步驟 4] 前端將 credential 發送至後端
         ↓ 後端驗證 token → 取得 Google user ID + email + name
[判斷]   查詢 user_oauth_providers 表：
         ├─ 已有帳號 → 直接登入 → 回傳 Token
         └─ 無帳號 → 自動建立帳號（暱稱預設為 Google 帳號名稱）→ 回傳 Token
[完成]   登入成功 → 跳轉首頁
```

#### 2.3.3 帳號關聯機制

同一用戶可以綁定多種登入方式（手機 + Apple + Google）。關聯規則：

- 社交登入建立的新帳號，後續可在「帳號設定」中綁定手機號碼
- 手機號碼登入的既有用戶，後續可在「帳號設定」中綁定 Apple / Google
- MVP 階段：帳號綁定功能暫不實作 UI，僅預留後端 API 和 DB 結構。若社交登入的 email 與現有帳號 email 相同，**不自動合併**，避免安全問題

### 2.4 登出流程

不變，維持現有邏輯。

### 2.5 錯誤場景（SMS 相關）

| 場景 | 處理方式 |
|------|---------|
| 手機號碼格式不正確 | 前端驗證，提示「請輸入正確的手機號碼」 |
| 60 秒內重複發送驗證碼 | 按鈕 disabled + 倒數計時顯示 |
| 驗證碼錯誤 | 顯示「驗證碼錯誤，請重新輸入」 |
| 驗證碼過期（超過 5 分鐘） | 顯示「驗證碼已過期，請重新發送」 |
| 登入時手機號碼未註冊 | 顯示「此手機號碼尚未註冊」 |
| 註冊時手機號碼已存在 | 顯示「此手機號碼已被註冊」 |

### 2.6 錯誤場景（社交登入相關）

| 場景 | 處理方式 |
|------|---------|
| 使用者取消 OAuth 授權 | 返回登入頁，顯示「登入已取消」 |
| Apple/Google token 驗證失敗 | 顯示「登入失敗，請重試」 |
| OAuth 服務不可用 | 顯示「服務暫時不可用，請使用手機號碼登入」 |
| Apple 登入未回傳 email（非首次） | 正常流程，透過 Apple sub ID 比對即可 |

---

## 3. API 變更

### 3.1 新增端點

#### POST /api/v1/auth/send-code

發送 SMS 驗證碼。

**Request Body：**
```json
{
  "phone": "+886912345678"
}
```

**Response 200：**
```json
{
  "data": {
    "message": "驗證碼已發送",
    "expires_in": 300
  }
}
```

**Response 429（60 秒內重複發送）：**
```json
{
  "error": "請等待 60 秒後再試",
  "retry_after": 45
}
```

**MVP Mock 行為**：
- 不實際發送簡訊
- 驗證碼固定為 `123456`
- 同時在 server console 印出：`[MOCK SMS] Phone: +886912345678, Code: 123456`
- 驗證碼存入 verification_codes 表或 Redis，5 分鐘過期

### 3.2 新增端點：社交登入

#### POST /api/v1/auth/oauth/apple

Apple OAuth 登入/自動註冊。

**Request Body：**
```json
{
  "code": "apple_authorization_code",
  "id_token": "apple_id_token",
  "user": {
    "name": "小明",
    "email": "user@icloud.com"
  }
}
```

> `user` 欄位僅在 Apple 首次授權時由前端傳送，後續登入可省略。

**Response 200（既有用戶登入）：**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "is_new_user": false,
    "user": {
      "id": "uuid",
      "phone": "+886912345678",
      "nickname": "小明",
      "avatar_url": null,
      "bio": null,
      "created_at": "..."
    }
  }
}
```

**Response 201（新用戶自動建立帳號）：**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "is_new_user": true,
    "user": {
      "id": "uuid",
      "phone": null,
      "nickname": "Apple 用戶",
      "avatar_url": null,
      "bio": null,
      "created_at": "..."
    }
  }
}
```

**錯誤回應：**
- 400：authorization code 或 id_token 無效
- 502：Apple 驗證服務不可用

**後端驗證流程：**
1. 使用 Apple 的公鑰驗證 `id_token`（從 `https://appleid.apple.com/auth/keys` 取得 JWKS）
2. 從 id_token 解出 `sub`（Apple user ID）
3. 查詢 `user_oauth_providers` 表是否有對應記錄
4. 有 → 取得關聯的 user → 產生 Token
5. 沒有 → 建立新 user + 建立 oauth provider 記錄 → 產生 Token

#### POST /api/v1/auth/oauth/google

Google OAuth 登入/自動註冊。

**Request Body：**
```json
{
  "credential": "google_id_token_or_authorization_code"
}
```

**Response 200（既有用戶登入）：**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "is_new_user": false,
    "user": {
      "id": "uuid",
      "phone": "+886912345678",
      "nickname": "小明",
      "avatar_url": "https://lh3.googleusercontent.com/...",
      "bio": null,
      "created_at": "..."
    }
  }
}
```

**Response 201（新用戶自動建立帳號）：**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "is_new_user": true,
    "user": {
      "id": "uuid",
      "phone": null,
      "nickname": "Google 用戶名稱",
      "avatar_url": "https://lh3.googleusercontent.com/...",
      "bio": null,
      "created_at": "..."
    }
  }
}
```

**錯誤回應：**
- 400：credential 無效
- 502：Google 驗證服務不可用

**後端驗證流程：**
1. 使用 Google 的公鑰驗證 `credential`（呼叫 Google token info endpoint 或本地 JWKS 驗證）
2. 從 token 解出 `sub`（Google user ID）、`email`、`name`、`picture`
3. 查詢 `user_oauth_providers` 表是否有對應記錄
4. 有 → 取得關聯的 user → 產生 Token
5. 沒有 → 建立新 user（暱稱取 Google name，頭像取 picture）+ 建立 oauth provider 記錄 → 產生 Token

### 3.3 修改端點

#### POST /api/v1/auth/register

**現有 Request Body（移除）：**
```json
{
  "email": "user@example.com",
  "password": "mypassword",
  "nickname": "小明"
}
```

**新 Request Body：**
```json
{
  "phone": "+886912345678",
  "code": "123456",
  "nickname": "小明"
}
```

**Response 201（不變）：**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "user": {
      "id": "uuid",
      "phone": "+886912345678",
      "nickname": "小明",
      "avatar_url": null,
      "bio": null,
      "created_at": "..."
    }
  }
}
```

**錯誤回應：**
- 400：驗證碼錯誤或已過期
- 409：手機號碼已被註冊

#### POST /api/v1/auth/login

**現有 Request Body（移除）：**
```json
{
  "email": "user@example.com",
  "password": "mypassword"
}
```

**新 Request Body：**
```json
{
  "phone": "+886912345678",
  "code": "123456"
}
```

**Response 200（不變，但 user 物件中 email 改為 phone）：**
```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "user": {
      "id": "uuid",
      "phone": "+886912345678",
      "nickname": "小明",
      "avatar_url": null,
      "bio": null,
      "created_at": "..."
    }
  }
}
```

**錯誤回應：**
- 400：驗證碼錯誤或已過期
- 404：手機號碼未註冊

### 3.4 不變端點

| 端點 | 說明 |
|------|------|
| POST /api/v1/auth/refresh | Token 刷新，不變 |
| POST /api/v1/auth/logout | 登出，不變 |

---

## 4. DB 變更

### 4.1 users 表修改

| 欄位 | 變更 | 說明 |
|------|------|------|
| `phone` | `*string` → `string`（必填，UNIQUE） | 改為主要認證欄位 |
| `email` | `string`（必填）→ `*string`（選填） | 降級為選填，未來可讓用戶自行補填 |
| `password_hash` | 移除 | 不再需要密碼認證 |

**Migration 要點：**
- 將 `phone` 設為 NOT NULL + UNIQUE（需先處理現有資料）
- 將 `email` 改為 NULLABLE，移除 UNIQUE 約束（或保留 UNIQUE 但允許 NULL）
- 移除 `password_hash` 欄位
- 注意：如有現有測試資料，需提供資料遷移腳本

### 4.2 新增 verification_codes 表

```sql
CREATE TABLE verification_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone      VARCHAR(20) NOT NULL,
    code       VARCHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used       BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_verification_codes_phone ON verification_codes(phone);
CREATE INDEX idx_verification_codes_expires_at ON verification_codes(expires_at);
```

**欄位說明：**
- `phone`：手機號碼
- `code`：6 位數驗證碼
- `expires_at`：過期時間（建立時間 + 5 分鐘）
- `used`：是否已使用（驗證成功後標記為 true，防止重複使用）

**替代方案**：也可使用 Redis 存放驗證碼（key: `sms_code:{phone}`, value: code, TTL: 300s），由 SA/BACKEND 決定。

---

## 5. 前端變更

### 5.1 LoginPage.vue

**現有**：email 輸入框 + 密碼輸入框 + 登入按鈕

**改為**：
- 手機號碼輸入框（type: tel，左側圖示改為手機圖示）
- 發送驗證碼按鈕（手機號碼輸入框右側或下方，60 秒倒數）
- 6 位數驗證碼輸入框（可考慮用 van-password-input 或 van-field）
- 登入按鈕

### 5.2 RegisterPage.vue

**現有**：暱稱 + email + 密碼 + 確認密碼 + 註冊按鈕

**改為兩步驟**：
- 步驟一：手機號碼 + 發送驗證碼 + 驗證碼輸入 + 「下一步」按鈕
- 步驟二：暱稱輸入 + 「完成註冊」按鈕

### 5.3 auth store（stores/auth.ts）

**介面變更：**

```typescript
// 移除
interface LoginPayload {
  email: string
  password: string
}
interface RegisterPayload {
  email: string
  password: string
  nickname: string
}

// 新增
interface SendCodePayload {
  phone: string
}
interface LoginPayload {
  phone: string
  code: string
}
interface RegisterPayload {
  phone: string
  code: string
  nickname: string
}
```

**新增 action：**
- `sendCode(payload: SendCodePayload)` — 呼叫 POST /api/v1/auth/send-code

**User 介面變更：**
```typescript
interface User {
  id: string
  phone: string       // 原本是 email
  nickname: string
  avatar_url: string
  bio: string
  created_at: string
}
```

### 5.4 驗證碼倒數計時邏輯

```
- 點擊「發送驗證碼」後啟動 60 秒倒數
- 倒數期間按鈕 disabled，顯示「重新發送 (XXs)」
- 倒數結束後恢復按鈕可點擊狀態
- 倒數狀態存在元件內即可，不需要持久化
```

### 5.5 手機號碼格式驗證（前端）

- 支援台灣手機號碼格式：09xx-xxx-xxx
- 前端自動加上國碼前綴：`09xxxxxxxx` → `+886912345678`
- 或直接接受 `+886` 開頭的完整格式
- 具體驗證規則由 SA 定義，MVP 先做基本長度檢查即可

---

## 6. 驗收標準

### 6.1 註冊流程

- [ ] 輸入手機號碼 → 點擊發送驗證碼 → 收到驗證碼（mock：123456）
- [ ] 輸入正確驗證碼 + 暱稱 → 註冊成功 → 自動登入 → 跳轉首頁
- [ ] 註冊成功後，users 表中有對應的 phone 記錄
- [ ] 已註冊的手機號碼再次註冊 → 顯示「此手機號碼已被註冊」

### 6.2 登入流程

- [ ] 輸入已註冊手機號碼 → 發送驗證碼 → 輸入驗證碼 → 登入成功
- [ ] 登入成功後取得 access_token 和 refresh_token
- [ ] 未註冊手機號碼嘗試登入 → 顯示「此手機號碼尚未註冊」

### 6.3 驗證碼機制

- [ ] Mock SMS：驗證碼固定 123456 或在 server console 印出
- [ ] 驗證碼 5 分鐘後過期 → 使用過期驗證碼顯示「驗證碼已過期」
- [ ] 60 秒倒數期間按鈕 disabled，不可重複發送
- [ ] 驗證碼使用後標記為已用，不可重複使用
- [ ] 輸入錯誤驗證碼 → 顯示「驗證碼錯誤」

### 6.4 不受影響的功能

- [ ] Token 刷新機制正常運作
- [ ] 登出功能正常運作
- [ ] 個人頁面查看/編輯正常運作（頭像、暱稱、簡介）

---

## 7. 開發順序

依照專案規則（前端先行 → 老闆確認 → 後端實作 → 串接）：

### 階段一：前端 UI 改版

**負責**：FRONTEND Agent
**Branch**：`feature/frontend-auth-revamp`

1. 修改 `LoginPage.vue` — 手機號碼 + 驗證碼兩步驟 UI
2. 修改 `RegisterPage.vue` — 手機號碼 + 驗證碼 + 暱稱
3. 修改 `stores/auth.ts` — 新介面 + sendCode action
4. 前端 mock API response（不依賴後端即可展示完整流程）
5. 倒數計時元件邏輯

**交付物**：可操作的前端 UI，mock 資料驅動

### 階段二：老闆確認畫面

- ngrok 暴露前端服務
- 發送連結給老闆確認 UI/UX
- 根據回饋調整

### 階段三：後端 API + DB 改版

**負責**：BACKEND Agent + DBA Agent
**Branch**：`feature/backend-auth-revamp`、`feature/dba-auth-revamp`

1. DB Migration：修改 users 表 + 新增 verification_codes 表
2. 新增 POST /api/v1/auth/send-code（含 mock SMS 邏輯）
3. 修改 Register — 驗證碼驗證取代密碼
4. 修改 Login — 驗證碼驗證取代密碼
5. 移除 bcrypt 相關邏輯
6. 更新 model（RegisterRequest、LoginRequest、User）

### 階段四：前後端串接 + QA

1. 前端移除 mock，串接真實 API
2. QA 執行完整測試（依據驗收標準）
3. Bug 修復 → 重測 → 通過

---

## 8. 技術注意事項

1. **手機號碼格式**：統一存儲為 E.164 格式（如 `+886912345678`），SA 需定義具體驗證規則
2. **驗證碼安全性**：MVP 階段暫不實作頻率限制（rate limiting），但建議後端預留介面，未來可加上每小時/每日發送上限
3. **SMS 服務切換**：mock 邏輯應封裝為獨立 service/interface，未來切換至真實簡訊商（如 Twilio、Nexmo）時只需替換實作
4. **向下相容**：此為 breaking change，不需考慮向下相容（MVP 階段無正式用戶資料）
5. **測試資料**：migration 後現有測試帳號將失效，需重新建立

---

## 附註

- 本次改版為 P0 優先級，對應 requirements.md 中 US-01（註冊）與 US-02（登入）
- SMS 驗證碼服務未來切換至真實簡訊商時，將作為獨立 Sprint 任務處理
- 手機號碼格式驗證規則的最終決定由 SA 負責
