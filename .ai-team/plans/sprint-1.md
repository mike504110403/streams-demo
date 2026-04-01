# Sprint 1 — 基礎建設 + 用戶系統

**狀態**：✅ 已完成
**開始日期**：2026-04-01
**目標**：跑通註冊登入流程，建立前後端骨架

---

## 任務分配

### DBA
- [x] users 表 migration（UUID PK、email、phone、password_hash、nickname、avatar_url、bio）
- [x] 建立索引（email UNIQUE、phone UNIQUE）
- [x] 認證改版 migration — phone 必填 + verification_codes + oauth_providers

### BACKEND
- [x] 專案初始化（Go module、Gin、目錄結構）
- [x] Docker Compose（PostgreSQL + Redis）
- [x] DB 連線 + sqlc 設定
- [x] POST /api/v1/auth/send-code — 發送驗證碼
- [x] POST /api/v1/auth/register — 用戶註冊（手機+驗證碼）
- [x] POST /api/v1/auth/login — 用戶登入（手機+驗證碼）
- [x] POST /api/v1/auth/refresh — 刷新 Token
- [x] POST /api/v1/auth/logout — 登出（Token 黑名單）
- [x] POST /api/v1/auth/oauth/apple — Apple OAuth 登入
- [x] POST /api/v1/auth/oauth/google — Google OAuth 登入
- [x] GET /api/v1/users/me — 取得個人資訊
- [x] PUT /api/v1/users/me — 更新個人資料
- [x] JWT 中間件
- [x] CORS 中間件

### FRONTEND
- [x] 專案初始化（Vue 3 + Vite + Vant 4 + Pinia + Vue Router）
- [x] 登入頁面（手機+驗證碼+Apple/Google OAuth）
- [x] 註冊頁面（手機+驗證碼）
- [x] Token 管理（Axios interceptor + auto refresh）
- [x] 個人頁面（查看 + 編輯）
- [x] 路由守衛（未登入跳轉登入頁）

---

## 驗收標準
- 用戶可完成：註冊 → 登入 → 查看個人資料 → 編輯個人資料 → 登出
- JWT Token 過期後自動刷新，用戶無感
- 前後端 API 串接正常
