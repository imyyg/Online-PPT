---

description: "Task list for feature implementation"
---

# Tasks: 用户登录与PPT记录服务

**Input**: 设计文档位于 `/specs/001-add-go-backend/`
**Prerequisites**: plan.md（必读）、spec.md（必读）、research.md、data-model.md、contracts/

**Tests**: 仅在文档或需求明确要求时添加测试任务；下方列出的测试均用于满足验收标准。

**Organization**: 任务按用户故事拆分，保证每个故事都能独立完成并验证。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行执行（不同文件且无前置依赖）
- **[Story]**: 任务所属用户故事（US1、US2、US3）
- 描述中必须包含精确文件路径

## Constitution Guardrails

- 添加路径校验与记录写入任务，确保不修改 `presentations/<user-uuid>/<group>/slides` 下的 HTML 内容。
- 同步规划 MySQL 结构与前端所依赖的 `slides.config.json` 信息，避免记录失真。
- 审查新增 API 响应，保持与现有 Vue 组件兼容。
- 最终验证体验，包括目录定位与快捷操作表现。

---

## Phase 1: Setup（共享基础设施）

**Purpose**: 初始化后端工程结构与必需依赖。

- [X] T001 在 online-ppt/ 下按照计划创建目录骨架（cmd/server/, internal/{auth,records,storage,http,config}, pkg/validator, configs, migrations, tests/{integration,e2e})
- [X] T002 初始化 Go 模块与核心依赖，写入 online-ppt/go.mod 并执行 `go mod tidy`
- [X] T003 创建服务器启动入口，在 online-ppt/cmd/server/main.go 中加载配置占位并启动 Gin 引擎骨架

---

## Phase 2: Foundational（阻塞性前置条件）

**Purpose**: 构建所有用户故事共享的基础能力；完成前不得进入任何用户故事开发。

- [X] T004 实现配置加载器，解析 configs/app.yaml 到结构体（online-ppt/internal/config/config.go）
- [X] T005 [P] 创建 MySQL 连接工厂与连接池管理（online-ppt/internal/storage/mysql.go）
- [X] T006 建立迁移执行入口并准备首个空迁移文件（online-ppt/internal/storage/migrate.go 与 online-ppt/migrations/000_bootstrap.sql）
- [X] T007 [P] 定义基础路由装配与版本前缀，写入 online-ppt/internal/http/router.go
- [X] T008 [P] 添加统一日志与错误处理中间件（online-ppt/internal/http/middleware/logger.go）
- [X] T009 配置 JWT 签名管理与刷新策略（online-ppt/internal/auth/token_manager.go）

**Checkpoint**: 完成后可开始实现用户故事。

---

## Phase 3: 用户故事 1 - 邮箱注册登录（Priority: P1）🎯 MVP

**Goal**: 提供邮箱注册、登录、刷新与注销 API，完成基础账户体系。

**Independent Test**: 通过集成测试验证注册 → 登录 → 刷新 → 注销全流程，并确认会话状态保存与吊销。

### 实现任务

- [X] T010 [US1] 编写用户与会话表迁移，创建 online-ppt/migrations/001_create_user_tables.sql（含索引与约束）
- [X] T011 [P] [US1] 定义用户与会话模型及扫描器，写入 online-ppt/internal/auth/models.go
- [X] T012 [P] [US1] 实现密码哈希与验证工具（Argon2id）于 online-ppt/internal/auth/password.go
- [X] T013 [US1] 开发用户仓储层，处理注册、查找与会话持久化（online-ppt/internal/auth/repository.go）
- [X] T014 [US1] 构建认证服务，封装注册/登录/刷新逻辑并记录审计事件（online-ppt/internal/auth/service.go）
- [X] T015 [US1] 实现 `/auth/register|login|logout|refresh` 处理器并设置 HttpOnly Cookie（online-ppt/internal/http/handlers/auth_handlers.go）
- [X] T016 [US1] 更新路由装配连接认证端点并确保主函数启用中间件（online-ppt/internal/http/router.go）
- [X] T017 [US1] 编写端到端集成测试验证邮箱注册与登录流程（online-ppt/tests/integration/auth_flow_test.go）

**Checkpoint**: 注册登录能力可独立演示，MVP 达成。

---

## Phase 4: 用户故事 2 - 记录用户PPT（Priority: P2）

**Goal**: 登录用户可创建 PPT 记录，系统生成合法目录并返回完整路径。

**Independent Test**: 已登录用户创建记录后，测试检查目录生成、路径校验与记录返回字段。

### 实现任务

- [X] T018 [US2] 新增 PPT 记录表与索引迁移（online-ppt/migrations/002_create_ppt_records.sql）
- [X] T019 [P] [US2] 实现路径清洗与校验工具，确保符合 presentations/<user-uuid>/<group>/slides（online-ppt/internal/records/path_validator.go）
- [X] T020 [P] [US2] 实现记录实体与仓储访问层，写入 online-ppt/internal/records/repository.go（含事务处理）
- [X] T021 [US2] 构建记录创建服务，负责目录创建与路径缓存（online-ppt/internal/records/service.go）
- [X] T022 [US2] 实现 POST /ppts 处理器，返回 relativePath 与 canonicalPath（online-ppt/internal/http/handlers/records_create_handler.go）
- [X] T023 [US2] 添加集成测试覆盖记录创建与路径限制（online-ppt/tests/integration/ppts_create_test.go）

**Checkpoint**: 登录用户可以创建记录并定位目录。

---

## Phase 5: 用户故事 3 - 管理个人列表（Priority: P3）

**Goal**: 支持记录查询、搜索、更新、删除，保障仅限本人操作。

**Independent Test**: 登录后执行列表、搜索、更新、删除并断言权限与路径校验。

### 实现任务

- [X] T024 [US3] 扩展查询构建与筛选逻辑，新增排序与分页支持（online-ppt/internal/records/query_builder.go）
- [X] T025 [US3] 更新记录服务以支持搜索、更新描述与删除（online-ppt/internal/records/service.go）
- [X] T026 [US3] 实现 GET/PATCH/DELETE /ppts/{id} 处理器并限制访问范围（online-ppt/internal/http/handlers/records_manage_handler.go）
- [X] T027 [US3] 添加集成测试验证查询、更新与删除场景（online-ppt/tests/integration/ppts_manage_test.go）

**Checkpoint**: 用户可维护个人 PPT 记录列表且具备权限控制。

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: 收尾工作与跨故事改进。

- [X] T028 [P] 将审计事件写入结构化日志或数据库，封装在 online-ppt/internal/storage/audit_logger.go
- [X] T029 [P] 更新部署与配置文档，补充 online-ppt/README.md 与 specs/001-add-go-backend/quickstart.md 的 MySQL 指引
- [X] T030 运行 Quickstart 全流程验证并记录结论（specs/001-add-go-backend/quickstart.md）

---

## Dependencies & Execution Order

1. **Setup → Foundational**：Phase 1 完成后才能构建共享基础；Phase 2 任务互有依赖，建议按 T004 → T009 顺序执行。
2. **User Stories**：三个用户故事均依赖 Foundational 完成；推荐顺序为 US1 → US2 → US3，确保身份体系先于记录能力再到管理能力。
3. **Polish**：所有核心故事结束后执行，用于统一日志、文档与终检。

---

## Parallel Opportunities

- Foundational 中 T005、T007、T008 可在配置加载完成后并行推进。
- 用户故事 1 中 T011 与 T012 可在迁移设计完成后并行，T013 需等待它们结束。
- 用户故事 2 中 T019 与 T020 可在迁移 T018 完成后并行。
- 用户故事 3 中列表处理与处理器实现（T025、T026）可与查询构建（T024）并行评审。
- Polish 阶段 T028 与 T029 可并行，T030 需依赖前述任务完成。

---

## Implementation Strategy

1. **MVP 交付**：完成 Phase 1 → Phase 2 → Phase 3（US1），即可提供注册登录的最小可用版本。
2. **增量扩展**：在 MVP 验收后，顺序实现 US2、US3，分别带来记录能力与记录管理功能，每个阶段完成后都可单独发布。
3. **质量收尾**：全部故事完成后执行 Polish，统一日志、文档与快速体验验证，确保部署指引与前端协同一致。
