# Implementation Tasks

## 1. Database Migration
- [x] 1.1 创建迁移脚本 `migrations/003_add_ppt_record_title.sql`
- [x] 1.2 添加 `ALTER TABLE ppt_records ADD COLUMN title VARCHAR(255) NULL AFTER name;`
- [x] 1.3 测试迁移脚本在开发环境的执行

## 2. 数据模型更新
- [x] 2.1 更新 `specs/001-add-go-backend/data-model.md` 中 PptRecord 实体定义，添加 title 字段说明
- [x] 2.2 在 `internal/records/models.go` 中为 PptRecord 结构体添加 `Title` 字段（如果存在该文件）

## 3. Repository 层
- [x] 3.1 更新 `internal/records/repository.go` 中的 CREATE 方法，支持 title 字段插入
- [x] 3.2 更新 `internal/records/repository.go` 中的 UPDATE 方法，支持 title 字段更新
- [x] 3.3 更新 `internal/records/repository.go` 中的 QUERY 方法，查询结果包含 title 字段
- [x] 3.4 编写或更新 repository 单元测试，验证 title 字段操作

## 4. Service 层
- [x] 4.1 更新 `internal/records/service.go` 中的业务逻辑，处理 title 字段验证（长度不超过 255 字符）
- [x] 4.2 确保创建和更新 PPT 记录时 title 为可选参数
- [x] 4.3 编写或更新 service 单元测试

## 5. HTTP Handlers
- [x] 5.1 更新 `internal/http/handlers/records_create_handler.go`，支持接收 title 参数
- [x] 5.2 更新 `internal/http/handlers/records_manage_handler.go`，支持 title 字段的更新和查询
- [x] 5.3 确保 API 响应中包含 title 字段
- [x] 5.4 添加 title 字段的参数验证（最大长度 255 字符）

## 6. 集成测试
- [x] 6.1 在 `tests/integration/` 中添加测试用例，验证创建带 title 的 PPT 记录
- [x] 6.2 添加测试用例验证创建不带 title 的 PPT 记录（应允许 NULL）
- [x] 6.3 添加测试用例验证更新 PPT 记录的 title
- [x] 6.4 添加测试用例验证 title 超长时的错误处理

## 7. 文档更新
- [x] 7.1 更新 API 文档（如果有 `specs/001-add-go-backend/contracts/api.yaml`），添加 title 字段说明
- [x] 7.2 更新 README.md 中的相关说明（如果需要）
