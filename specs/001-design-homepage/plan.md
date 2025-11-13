# Implementation Plan: 网站首页概念设计

**Branch**: `001-design-homepage` | **Date**: 2025-11-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-design-homepage/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

围绕首页首屏价值陈述、左上角注册/登录入口、功能亮点与社会证明四大模块展开内容规划，确保访客在 10 秒内理解产品价值并完成注册/登录入口点击。技术实施将基于现有 `ppt-framework` Vue 3 单页应用，复用既有布局组件与 Tailwind 工具类，仅新增所需的内容组件与跟踪事件。

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: JavaScript (Vue 3)  
**Primary Dependencies**: Vue 3, Vite, Tailwind CSS, Pinia  
**Storage**: N/A（前端静态内容）  
**Testing**: Vitest + @vue/test-utils（需为关键交互编写单元测试）  
**Target Platform**: 现代桌面与移动浏览器  
**Project Type**: Web 单页应用  
**Performance Goals**: 首屏可交互时间 ≤ 2s（桌面宽带），LCP ≤ 2.5s  
**Constraints**: 需满足无障碍导航、移动端响应式布局、按钮点击热区 ≥ 44px  
**Scale/Scope**: 单页首页与关联组件，覆盖注册/登录入口与 4 个内容区块

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **P1 Slide Files**: 本次仅改造首页（`ppt-framework/src`），不新增或修改任何 `presentations/<group>/slides` 文件；如需新增演示示例，将通过链接指向现有静态文件并保持其独立 HTML 结构。
- **P2 Config Sync**: 首页改造不涉及 `slides.config.json`。如未来新增“Homepage Demo”示例，将在规划阶段同步更新配置并记录在计划文档。
- **P3 Frontend Discipline**: 全程在既有 Vue 3 + Vite + Pinia + Tailwind 栈内实施；若评审后需引入额外组件库，将在评审会上提交依赖审查。
- **P4 UX Baseline**: 首页交互遵循无障碍与响应式规范，重点确保键盘可访问、移动端左上角导航可点击，并对登录/注册跳转进行无障碍标签标注。

**Phase 1 Re-check**: 依据 `data-model.md` 与 `contracts/cta-events.yaml`，上述四项原则保持满足状态，未引入需额外审批的结构或依赖。

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
ppt-framework/
├── src/
│   ├── App.vue
│   ├── main.js
│   ├── style.css
│   ├── tw.css
│   ├── api/
│   ├── components/
│   ├── composables/
│   ├── stores/
│   └── utils/
├── public/
│   └── templates/
├── presentations/
│   └── example/
│       ├── slides.config.json
│       └── slides/
└── tests/
    ├── e2e/
    └── integration/
```

**Structure Decision**: 在 `ppt-framework/src/components` 内新增首页专属组件，并在 `App.vue` 或对应路由容器加载；保持现有 `presentations` 与 `public/templates` 目录不变。

## Complexity Tracking

当前方案未触发宪章限制，无需额外复杂度登记。
