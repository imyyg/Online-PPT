# Implementation Plan: 用户登录与PPT记录服务

**Branch**: `001-add-go-backend` | **Date**: 2025-11-05 | **Spec**: [specs/001-add-go-backend/spec.md](../001-add-go-backend/spec.md)
**Input**: Feature specification from `/specs/001-add-go-backend/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

实现一个与前端 `ppt-framework` 协同工作的后端服务，提供邮箱注册/登录、用户PPT记录管理与路径定位能力。后端部署为仓库同级的 `online-ppt` 目录，向前端暴露REST API，基于 MySQL 持久化数据，并为每位用户在 `presentations/<user-uuid>/` 下维护独立幻灯片目录以保持幻灯片独立性。

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.22 (WSL + 容器环境均可获得官方发行版)  
**Primary Dependencies**: Gin (web 框架), `github.com/go-sql-driver/mysql` (MySQL 驱动), `golang.org/x/crypto/argon2` (密码哈希), `github.com/stretchr/testify` (测试)  
**Storage**: MySQL 8.x（InnoDB，INT AUTO_INCREMENT 主键，用户表额外 UUID 字段）  
**Testing**: Go `testing` + `httptest` + Testify  
**Target Platform**: Linux server (WSL + container-friendly)
**Project Type**: Monorepo前端+独立后端API服务  
**Performance Goals**: 注册/登录在95%请求下≤700ms；记录查询≤500ms；批量列表分页≤3s/100条  
**Constraints**: 路径必须符合 `ppt-framework/presentations/<user-uuid>/<group>/slides`；Token 基于JWT并使用 HttpOnly Cookie  
**Scale/Scope**: 单用户≤500条记录的初期部署，支持未来扩展至小团队

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **P1 Slide Files**: 后端仅读取/验证路径，不修改 `presentations/<user-uuid>/<group>/slides` 下的HTML文件；所有创建记录任务需验证保持独立结构。
- **P2 Config Sync**: 若记录中引用分组信息，规划中必须确认 `slides.config.json` 为事实来源并在设计阶段明确读取方式，无需改动配置文件。
- **P3 Frontend Discipline**: 所有API输出需贴合现有Vue组件数据结构，不引入新的前端依赖；任何新增字段需在与前端协作后纳入文档。
- **P4 UX Baseline**: 计划将提供标题、标签、路径等元数据，支持前端继续保障快捷键导航与展示体验，不对幻灯片渲染逻辑造成破坏。

**Gate Status**: PASS（设计阶段未引入对幻灯片结构或前端技术栈的破坏性变更）。

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
online-ppt/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/
│   ├── records/
│   ├── storage/
│   ├── http/
│   └── config/
├── pkg/
│   └── validator/
├── configs/
│   └── app.yaml
├── migrations/            # 若选择数据库存储（待定）
└── tests/
  ├── integration/
  └── e2e/

ppt-framework/
├── src/
│   ├── components/
│   ├── stores/
│   └── utils/
└── tests/
```

**Structure Decision**: 新增 `online-ppt/` 作为Go后端服务根目录，采用 `cmd + internal + pkg` 分层；保留既有前端 `ppt-framework/` 目录并通过API对接后端。数据库相关目录（如 `migrations/`）将在研究阶段根据存储方案决定是否启用。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _None_ | — | — |
