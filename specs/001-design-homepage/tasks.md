# Tasks: 网站首页概念设计

**Input**: Design documents from `/specs/001-design-homepage/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/
**Tests**: 包含 Vitest 集成测试与可访问性自检
**Organization**: 任务按用户故事划分，确保每个故事可独立实现与验收

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立公共资源与分析基础，供各用户故事复用

- [X] T001 创建首页资源目录占位文件 `ppt-framework/src/assets/homepage/.gitkeep`
- [X] T002 [P] 新建内容数据源骨架 `ppt-framework/src/utils/homepageContent.js` 并导出区块占位结构
- [X] T003 [P] 新建 CTA 分析工具 `ppt-framework/src/utils/analytics.js` 并定义 `trackHomepageCta(payload)` 占位实现

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 搭建首页容器与路由判定逻辑，确保不影响现有幻灯片播放体验

**⚠️ CRITICAL**: 未完成前不得开始任何用户故事工作

- [X] T004 创建聚合组件骨架 `ppt-framework/src/components/homepage/HomepageShell.vue`（预留 hero/feature/证明/页脚插槽）
- [X] T005 更新 `ppt-framework/src/App.vue` 以在营销模式下渲染 `HomepageShell`，并维持幻灯片查看模式的既有逻辑
- [X] T006 调整 `ppt-framework/src/main.js` 以检测根路径触发营销模式且跳过自动播放行为

**Checkpoint**: 首页容器与路由切换准备就绪

---

## Phase 3: User Story 1 - 首屏快速了解价值 (Priority: P1) 🎯 MVP

**Goal**: 访客首屏即可理解产品价值并注意到左上角注册入口

**Independent Test**: 运行 `vitest run ppt-framework/tests/integration/homepage-hero.spec.ts` 验证英雄区文案与注册 CTA 呈现

### Tests for User Story 1

- [X] T007 [P] [US1] 编写英雄区渲染测试 `ppt-framework/tests/integration/homepage-hero.spec.ts`

### Implementation for User Story 1

- [X] T008 [P] [US1] 创建导航组件 `ppt-framework/src/components/homepage/HomepageTopNav.vue` 并渲染左上角注册 CTA
- [X] T009 [P] [US1] 实现英雄区组件 `ppt-framework/src/components/homepage/HomepageHero.vue`，加载一句话价值主张与主召唤按钮
- [X] T010 [US1] 组合导航与英雄区到 `ppt-framework/src/components/homepage/HomepageShell.vue` 并将注册 CTA 链接指向现有注册路径
- [X] T011 [US1] 填充英雄区内容数据 `ppt-framework/src/utils/homepageContent.js`（标题、副文案、按钮文案、图像引用）
- [X] T012 [US1] 在 `ppt-framework/src/components/homepage/HomepageHero.vue` 调用 `trackHomepageCta` 记录注册 CTA 点击

**Checkpoint**: 英雄区与注册入口可独立演示

---

## Phase 4: User Story 2 - 回访用户快速登录 (Priority: P2)

**Goal**: 回访用户能在左上角立即找到登录入口并在移动端获得良好响应式体验

**Independent Test**: 运行 `vitest run ppt-framework/tests/integration/homepage-auth.spec.ts` 验证登录 CTA 可见且具备可访问标签

### Tests for User Story 2

- [X] T013 [P] [US2] 编写登录 CTA 可视性测试 `ppt-framework/tests/integration/homepage-auth.spec.ts`

### Implementation for User Story 2

- [X] T014 [US2] 扩展 `ppt-framework/src/components/homepage/HomepageTopNav.vue` 以渲染登录 CTA 并在 <768px 时改为纵向排列
- [X] T015 [US2] 更新 `ppt-framework/src/utils/homepageContent.js` 补充登录 CTA 文案与目标链接
- [X] T016 [US2] 在 `ppt-framework/src/components/homepage/HomepageTopNav.vue` 调用 `trackHomepageCta` 记录登录 CTA 点击

**Checkpoint**: 登录入口在桌面与移动端均可顺畅访问

---

## Phase 5: User Story 3 - 深入了解产品亮点 (Priority: P3)

**Goal**: 访客可浏览核心功能、模板示例与社会证明，并找到次级 CTA

**Independent Test**: 运行 `vitest run ppt-framework/tests/integration/homepage-content.spec.ts` 验证功能亮点与社会证明区块渲染

### Tests for User Story 3

- [X] T017 [P] [US3] 编写内容区块测试 `ppt-framework/tests/integration/homepage-content.spec.ts`

### Implementation for User Story 3

- [X] T018 [P] [US3] 创建功能亮点组件 `ppt-framework/src/components/homepage/HomepageFeatures.vue` 并渲染 v-for 列表
- [X] T019 [P] [US3] 创建社会证明组件 `ppt-framework/src/components/homepage/HomepageSocialProof.vue` 显示客户 Logo 或证言
- [X] T020 [P] [US3] 创建页脚 CTA 组件 `ppt-framework/src/components/homepage/HomepageFooterCta.vue` 并支持新标签打开示例
- [X] T021 [US3] 扩充 `ppt-framework/src/utils/homepageContent.js` 添加功能亮点、模板示例与社会证明数据
- [X] T022 [US3] 在 `ppt-framework/src/components/homepage/HomepageShell.vue` 组合功能亮点、社会证明与页脚 CTA
- [X] T023 [US3] 为所有次级 CTA 调用 `trackHomepageCta` 并传入区块 `placement`

**Checkpoint**: 首页整体内容栈可完整展示

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 对齐宪章要求、完成文档与质量核查

- [ ] T024 更新 `specs/001-design-homepage/quickstart.md` 使步骤与最终实现一致
- [ ] T025 记录无障碍巡检结果至 `specs/001-design-homepage/notes/accessibility.md`
- [ ] T026 保存性能与 Lighthouse 报告至 `specs/001-design-homepage/notes/performance.md`
- [ ] T027 运行 `npm run test -- --runInBand` 并在 `ppt-framework/tests/README.md` 记录新测试入口

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 → Phase 2**: Setup 完成后方可搭建首页容器
- **Phase 2 → Phases 3-5**: 基础容器与路由完成后，三个用户故事可并行推进
- **Phase 6**: 依赖所有计划交付的用户故事完成

### User Story Dependencies

- **US1**: 仅依赖 Phase 2；可独立演示（MVP）
- **US2**: 依赖导航组件已存在（T008），其余可并行
- **US3**: 依赖 `HomepageShell` 与内容数据文件（T004、T011）；不依赖 US2

---

## Parallel Opportunities

- Phase 1 中 T002 与 T003 可并行
- Phase 2 完成后，US1/US2/US3 可由不同成员并行推进
- 各用户故事中的测试任务（T007、T013、T017）在实现前可并行草拟
- US3 内的组件实现（T018、T019、T020）可并行进行

---

## Parallel Examples

### User Story 1

```bash
# 并行起步：先写测试再填充组件
vitest run ppt-framework/tests/integration/homepage-hero.spec.ts --watch
# 同时开发导航与英雄区组件
code ppt-framework/src/components/homepage/HomepageTopNav.vue \
     ppt-framework/src/components/homepage/HomepageHero.vue
```

### User Story 2

```bash
# 移动端布局调试 + 测试同步进行
vitest run ppt-framework/tests/integration/homepage-auth.spec.ts --watch
code ppt-framework/src/components/homepage/HomepageTopNav.vue
```

### User Story 3

```bash
# 功能亮点、社会证明、页脚 CTA 可拆分给不同成员
code ppt-framework/src/components/homepage/HomepageFeatures.vue \
     ppt-framework/src/components/homepage/HomepageSocialProof.vue \
     ppt-framework/src/components/homepage/HomepageFooterCta.vue
```

---

## Implementation Strategy

### MVP（仅交付 User Story 1）
1. 完成 Phase 1 与 Phase 2，确保首页容器与路由切换正常
2. 按顺序完成 T007-T012，验证注册 CTA 与英雄区展示
3. 运行 `vitest run ppt-framework/tests/integration/homepage-hero.spec.ts` 并收集转化指标基线

### 增量迭代
1. 在 MVP 基础上完成 US2（T013-T016），提升回访用户体验
2. 追加 US3（T017-T023），提供完整内容与社会证明
3. 每完成一故事即更新 Phase 6 文档与质量检查项

### 团队并行策略
- 成员 A：负责 Phase 2 + US1 组件与测试
- 成员 B：在 Phase 2 后立即开启 US2 响应式与登录 CTA
- 成员 C：专注 US3 内容区块与数据填充
- 全员共享 Phase 6 质量验收，轮流执行可访问性与性能检查
