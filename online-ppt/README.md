# Online-PPT 后端服务

## 项目简介
Online-PPT 后端基于 Go 1.22 与 Gin 实现，提供邮箱注册登录、PPT 记录管理以及结构化审计日志能力，配合前端 `ppt-framework/` 项目构成完整的演示内容管理方案。

## 运行前准备
- Go 1.22+
- MySQL 8.x（推荐 Docker 运行）
- 已克隆的仓库根目录，并确保 `ppt-framework/presentations/` 可读写

## 安装依赖
```bash
cd online-ppt
go mod tidy
```

## 配置说明
在 `configs/app.yaml` 创建配置文件，可参考以下示例：
```yaml
server:
  addr: ":8080"
security:
  jwtSecret: "change-me"
  accessTokenTTL: "15m"
  refreshTokenTTL: "720h"
storage:
  driver: "mysql"
  dsn: "user:pass@tcp(127.0.0.1:3306)/online_ppt?parseTime=true&loc=UTC"
paths:
  presentationsRoot: "../ppt-framework/presentations"
```
按需调整以下字段：
- `server.addr`：服务监听地址，默认 `:8080`
- `security.jwtSecret`：替换为自定义密钥
- `security.accessTokenTTL`、`security.refreshTokenTTL`：控制访问令牌与刷新令牌有效期
- `storage.dsn`：设置 MySQL 连接串
- `paths.presentationsRoot`：指向前端演示目录，如 `../ppt-framework/presentations`

## 启动服务
```bash
go run cmd/server/main.go
```
服务默认暴露在 `http://localhost:8080`，API 前缀为 `/api/v1`。首次启动会自动执行 `migrations/` 目录下的全部 SQL 脚本。

## 运行测试
```bash
go test ./...
```
集成测试位于 `tests/integration/`，覆盖注册登录及记录操作流程。

## 审计日志
所有认证与记录相关操作会写入 JSON Line 格式的结构化日志，默认输出到标准输出，例如：
```text
{"event":"auth.login","status":"success","userId":1,"sessionId":42,"timestamp":"2025-11-06T02:15:04.123456Z"}
```
可通过配置自定义日志输出或收集策略，便于后续接入集中式日志平台。

## 目录结构
- `cmd/server/`：应用入口与依赖注入
- `internal/auth/`：账号、会话与令牌逻辑
- `internal/records/`：PPT 记录业务逻辑与路径校验
- `internal/storage/`：数据库访问、审计日志工具
- `internal/http/`：路由、处理器与中间件
- `migrations/`：SQL 迁移脚本
- `tests/`：集成与端到端测试
