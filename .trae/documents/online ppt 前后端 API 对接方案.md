## 总览
- 前端目录：`ppt-framework/`（Vue 3 + Vite + Pinia）
- 后端目录：`online-ppt/`（Go + Gin），API 前缀：`/api/v1`（internal/http/router.go:21）
- 现状：前端已内置 Axios 客户端与接口封装，后端已实现认证与 PPT 记录管理；需完成统一配置、跨域与联调验证。

## 前端配置
- 基础地址：`VITE_API_BASE_URL` 已在 `.env.example` 与 `.env.local` 配置为 `http://localhost:8080/api/v1`（ppt-framework/.env.example:1，ppt-framework/.env.local:1）。
- Axios 基础配置：`src/api/client.js` 使用该地址并带超时与 `Content-Type`（ppt-framework/src/api/client.js:5-11）。
- 令牌注入：请求拦截器自动从 `localStorage.accessToken` 注入 `Authorization: Bearer <token>`（ppt-framework/src/api/client.js:13-21）。
- 刷新逻辑：响应拦截器在 401 时调用 `POST /auth/refresh`，用 `localStorage.refreshToken` 刷新并重试原请求（ppt-framework/src/api/client.js:26-60）。

## 接口映射
- 认证接口（后端）：
  - `GET /auth/captcha`（online-ppt/internal/http/handlers/auth_handlers.go:194-207）
  - `POST /auth/send-verification-code`（online-ppt/internal/http/handlers/auth_handlers.go:209-236）
  - `POST /auth/register`（验证码注册并自动登录，online-ppt/internal/http/handlers/auth_handlers.go:238-271）
  - `POST /auth/login`（online-ppt/internal/http/handlers/auth_handlers.go:52-72）
  - `POST /auth/refresh`（支持 body 和 Cookie，online-ppt/internal/http/handlers/auth_handlers.go:74-91, 161-177）
  - `POST /auth/logout`（清除 Cookie，204 无正文，online-ppt/internal/http/handlers/auth_handlers.go:93-108）
- 认证接口（前端封装）：`src/api/auth.js` 方法与后端字段一致（ppt-framework/src/api/auth.js:3-46）。
- PPT 记录接口（后端）：
  - `GET /ppts` 列表，查询参数 `q, tag, sort, limit, offset`（online-ppt/internal/http/handlers/records_manage_handler.go:16-54, 168-194）。
  - `POST /ppts` 创建（需鉴权，online-ppt/internal/http/handlers/records_create_handler.go:27-72）。
  - `GET /ppts/:id` 获取（需鉴权，online-ppt/internal/http/handlers/records_manage_handler.go:56-75）。
  - `PATCH /ppts/:id` 更新（需鉴权，online-ppt/internal/http/handlers/records_manage_handler.go:77-119）。
  - `DELETE /ppts/:id` 删除（需鉴权，online-ppt/internal/http/handlers/records_manage_handler.go:121-139）。
- PPT 记录接口（前端封装）：`src/api/ppts.js` 字段映射与后端一致（ppt-framework/src/api/ppts.js:3-37）。
- 前端状态管理：
  - 认证：`src/stores/auth.js` 统一保存 `user` 与 `accessToken`，并处理登录/注册/刷新/登出（ppt-framework/src/stores/auth.js:5-41, 43-116）。
  - 记录：`src/stores/ppts.js` 拉取/增删改查，列表响应读取 `response.items`（ppt-framework/src/stores/ppts.js:48-66）。

## 跨域与联调
- 方案 A（推荐，零后端改动）：在前端开发时新增 Vite 代理，将 `server.proxy['/api'] -> 'http://localhost:8080'`，并把 `VITE_API_BASE_URL` 设为 `'/api/v1'`，即可同源调用免 CORS；构建产物仍可用绝对地址或反向代理。
- 方案 B：在后端 Gin 增加 CORS 中间件，允许来自 `http://localhost:5173` 的跨域请求，`Access-Control-Allow-Credentials: true`，`Allow-Headers: Authorization, Content-Type`，并处理 `OPTIONS` 预检。保持前端 `VITE_API_BASE_URL=http://localhost:8080/api/v1` 不变。
- 两方案兼容性：若使用后端设置刷新 Cookie（`refresh_token`），需在前端请求开启 `withCredentials`；当前前端刷新使用 body `refreshToken`，可不启用 Cookie，也能协同工作。

## 运行配置
- 后端端口：`:8080`（online-ppt/configs/app.yaml:1-3）。
- 数据库：MySQL DSN（online-ppt/configs/app.yaml:9-12）。
- Redis：验证码与限流（online-ppt/configs/app.yaml:13-20）。
- SMTP：发送邮箱验证码（online-ppt/configs/app.yaml:21-30）。
- 演示文件根：`paths.presentationsRoot: ../ppt-framework/presentations`（online-ppt/configs/app.yaml:31-32）。

## 实施步骤
- 步骤 1：统一前端 API 基础地址
  - 开发用：选择方案 A，设置 `VITE_API_BASE_URL='/api/v1'` 并配置 Vite 代理；或选择方案 B 保持绝对地址并加后端 CORS。
- 步骤 2：网络层强化（可选）
  - 在 `src/api/client.js` 请求拦截器补充指纹头：`X-Client-Fingerprint`（后端用于会话绑定，online-ppt/internal/http/handlers/auth_handlers.go:153-159）。
  - 如需使用后端刷新 Cookie，开启 `axios.defaults.withCredentials = true` 并兼容现有 body 刷新。
- 步骤 3：接口联调验证
  - 打开 `AuthDemo.vue` 页面，走完整流程：获取验证码 → 发送邮箱验证码 → 注册（自动登录）/登录 → 拉取 `PPT Records`（ppt-framework/src/components/AuthDemo.vue:14-33, 35-55, 56-88, 95-127）。
  - 验证列表、创建、更新、删除在 `pptsStore` 中的行为与后端响应一致（ppt-framework/src/stores/ppts.js:89-149）。
- 步骤 4：错误与降级
  - 认证错误读取 `response.data.message` 已实现（ppt-framework/src/stores/auth.js:54, 72, 86）。
  - 若 SMTP 不可用，注册流程将失败；临时验证可使用已存在用户直接登录与列表接口。

## 验证清单
- 启后端：`online-ppt/cmd/server/main.go` 注册路由（online-ppt/cmd/server/main.go:103-107）并监听 8080。
- 启前端：Vite 开发服与代理，确认 `GET /auth/captcha`、`POST /auth/send-verification-code`、`POST /auth/register|login|refresh|logout` 正常；`GET /ppts` 返回 `items`，鉴权头存在。
- 浏览器网络面板检查：所有调用均指向 `VITE_API_BASE_URL` 路径，401 自动刷新重试成功。

## 交付与后续
- 首选交付为“前端代理配置 + 指纹头补充”，不改动后端；如需生产跨域，再补充后端 CORS 中间件。
- 后续可加：将刷新改为优先使用 HttpOnly Cookie，减少 `refreshToken` 在前端存储的风险。