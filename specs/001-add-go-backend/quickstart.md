# Quickstart: 用户登录与 PPT 记录服务

## 前置条件
- Go 1.22 及以上版本（`go version` 验证）
- MySQL 8.x（可通过 Docker 快速启动）
- 仓库已克隆并包含前端 `ppt-framework/`

### MySQL Docker 示例
```bash
docker run --name online-ppt-mysql \
  -e MYSQL_ROOT_PASSWORD=devpass \
  -e MYSQL_DATABASE=online_ppt \
  -p 3306:3306 -d mysql:8
```
启动后可使用 `mysql` 客户端确认数据库已创建：
```sql
CREATE DATABASE IF NOT EXISTS online_ppt CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## 步骤 1：安装依赖
```bash
cd online-ppt
go mod tidy
```

## 步骤 2：配置应用
在 `configs/app.yaml` 写入以下示例，并按需修改：
```yaml
server:
  addr: ":8080"
security:
  jwtSecret: "change-me"
  accessTokenTTL: "15m"
  refreshTokenTTL: "720h"
storage:
  driver: "mysql"
  dsn: "root:devpass@tcp(127.0.0.1:3306)/online_ppt?parseTime=true&loc=UTC"
paths:
  presentationsRoot: "../ppt-framework/presentations"
```
确保 `paths.presentationsRoot` 指向前端演示目录。

## 步骤 3：启动后端服务
```bash
go run cmd/server/main.go
```
首次启动会自动执行所有 SQL 迁移，创建用户、会话与 PPT 记录表。服务将监听 `http://localhost:8080`，提供 `/api/v1` 前缀的 REST API，终端会输出结构化审计日志（JSON Line 格式）。

## 步骤 4：验证核心 API
1. 注册新用户：
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/register \
        -H "Content-Type: application/json" \
        -d '{"email":"user@example.com","password":"PptDemo123!"}'
   ```
2. 登录获取令牌：
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"user@example.com","password":"PptDemo123!"}'
   ```
   响应包含访问令牌，刷新令牌通过 HttpOnly Cookie 返回。
3. 创建 PPT 记录：
   ```bash
   curl -X POST http://localhost:8080/api/v1/ppts \
        -H "Authorization: Bearer <access-token>" \
        -H "Content-Type: application/json" \
        -d '{"name":"DemoDeck","description":"团队周会"}'
   ```

## 步骤 5：运行测试
```bash
go test ./...
```
集成测试位于 `tests/integration/`，涵盖注册登录、记录创建与管理。

## 与前端协同
1. 在 `ppt-framework/` 执行 `npm install && npm run dev`
2. 配置 `VITE_API_BASE=http://localhost:8080/api/v1`
3. 若需要使用 Cookie 刷新令牌，确保前端请求携带凭证（如 Axios `withCredentials: true`）

## 审计日志
后端已集成结构化审计日志，默认输出到标准输出。示例：
```text
{"event":"records.create","status":"success","userId":1,"recordId":42,"timestamp":"2025-11-06T02:30:55.123456Z"}
```
可根据部署环境将日志重定向到文件或集中式日志系统。

## 常见问题排查
- **连接数据库失败**：确认 `storage.dsn` 中的主机、端口与凭证是否正确；Docker 场景下可使用 `127.0.0.1` 或容器网络名称。
- **路径校验失败**：演示名称需匹配 `^[A-Za-z0-9_-]+$`，并确保 `presentations/<user-uuid>/<group>/slides` 可创建。
- **令牌失效**：调 `/api/v1/auth/refresh` 获取新令牌，若刷新会话过期需重新登录。

## 验证记录
- 2025-11-06：在 Linux 本地环境完成依赖安装、配置、服务启动、核心 API 调用与 `go test ./...`，全部步骤通过。
