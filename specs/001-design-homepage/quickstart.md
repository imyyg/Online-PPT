# Quickstart: 网站首页概念设计

## Prerequisites
- Node.js 18+，已在 `ppt-framework` 目录执行 `npm install`
- Vite 开发服务器：`npm run dev`
- 分析埋点 SDK 已在全局初始化（参考现有 analytics 模块）

## Step 1: 创建首页组件骨架
1. 在 `ppt-framework/src/components` 新增 `HomepageHero.vue`, `HomepageFeatures.vue`, `HomepageSocialProof.vue`, `HomepageFooterCta.vue`。
2. 在 `App.vue`（或现有路由容器）中引入上述组件，保持首屏包裹结构，确保左上角导航区域可复用现有 `LeftMenu` 或新增 `TopBar` 组件。
3. 为每个组件准备本地化数据结构，可直接定义在 `src/utils/homepageContent.js`。

## Step 2: 实现左上角注册/登录导航
1. 复用现有样式变量，确保 Tailwind 类 `flex`, `gap-x-3`, `md:flex-row`，并在移动端（`sm:` 以下）改为 `flex-col`。
2. 添加 `aria-label` 与 `aria-describedby`，确保屏幕阅读器可区分两个按钮。
3. 为按钮绑定埋点 `emitCtaClick({ placement: 'header-register' })`。

## Step 3: 填充核心内容模块
1. Hero：绑定一句话价值主张与主 CTA；旁边展示静态预览图（放置于 `src/assets/homepage/hero-preview.png`）。
2. 功能亮点：创建数组 `features`，包含图标和文案，使用 `v-for` 渲染三列布局。
3. 模板示例：在 `HomepageSocialProof` 中展示模板卡片，CTA 以 `target="_blank"` 打开示例页面。
4. 社会证明：引入客户 Logo 轮播或证言列表，保持无障碍 `alt` 文案。

## Step 4: 添加分析与可访问性保障
1. 在 `src/utils/analytics.js` 定义 `trackHomepageCta(payload)`，统一事件名称 `cta_click`。
2. 在各组件调用时传入 `placement` 与 `content`，确保所有 CTA 被追踪。
3. 使用 `tabindex`, `role="button"`（如需），并测试键盘导航顺序。

## Step 5: 测试与性能验证
1. 编写 Vitest 测试 `tests/integration/homepage.spec.ts`，覆盖按钮渲染、点击回调与内容展示。
2. 使用 Lighthouse 或 WebPageTest 验证首屏 LCP ≤ 2.5s，移动端交互热区 ≥ 44px。
3. 手动检查移动端断点（375px/768px）确保按钮布局正确。

## Step 6: 发布准备
1. 运行 `npm run build` 确认构建通过且无 lint 错误。
2. 更新 README 或产品公告所需的截图与文案（如有）。
3. 在 PR 描述中引用本规格及主要成功指标，确保评审关注转化与无障碍要求。
