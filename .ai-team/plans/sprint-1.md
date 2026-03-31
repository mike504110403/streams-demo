# Sprint 1 — 基礎建設 + 用戶系統

**狀態**：進行中
**開始日期**：2026-04-01
**目標**：跑通註冊登入流程，建立前後端骨架

---

## 任務分配

### DBA
- [ ] users 表 migration（UUID PK、email、phone、password_hash、nickname、avatar_url、bio）
- [ ] 建立索引（email UNIQUE、phone UNIQUE）

### BACKEND
- [ ] 專案初始化（Go module、Gin、目錄結構）
- [ ] Docker Compose（PostgreSQL + Redis）
- [ ] DB 連線 + sqlc 設定
- [ ] POST /api/v1/auth/register — 用戶註冊
- [ ] POST /api/v1/auth/login — 用戶登入（回傳 JWT）
- [ ] POST /api/v1/auth/refresh — 刷新 Token
- [ ] POST /api/v1/auth/logout — 登出（Token 黑名單）
- [ ] GET /api/v1/users/me — 取得個人資訊
- [ ] PUT /api/v1/users/me — 更新個人資料
- [ ] JWT 中間件
- [ ] CORS 中間件

### FRONTEND
- [ ] 專案初始化（Vue 3 + Vite + Vant 4 + Pinia + Vue Router）
- [ ] 登入頁面
- [ ] 註冊頁面
- [ ] Token 管理（Axios interceptor + auto refresh）
- [ ] 個人頁面（查看 + 編輯）
- [ ] 路由守衛（未登入跳轉登入頁）

---

## 驗收標準
- 用戶可完成：註冊 → 登入 → 查看個人資料 → 編輯個人資料 → 登出
- JWT Token 過期後自動刷新，用戶無感
- 前後端 API 串接正常
