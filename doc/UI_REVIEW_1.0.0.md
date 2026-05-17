# Contful 1.0.0 UI 设计评审报告

> 日期: 2026-05-17 | 评审范围: 全局布局、登录页、仪表盘、列表页、表单页、侧栏

---

## 一、设计系统评分

| 维度 | 评分 | 说明 |
|------|------|------|
| 视觉一致性 | ⭐⭐⭐ | 有 token 体系但未全面贯彻，内联 style 多 |
| 信息架构 | ⭐⭐⭐⭐ | 二级侧栏分组合理，页面间导航清晰 |
| 响应式 | ⭐⭐ | 无响应式断点，移动端不可用 |
| 可访问性 | ⭐⭐ | 缺少 focus 样式、ARIA 标签、skip-link |
| 交互反馈 | ⭐⭐⭐ | 加载/空/错误状态基本覆盖，动效缺失 |
| 代码质量 | ⭐⭐⭐ | 内联样式过多，CSS 变量使用不彻底 |

**总评分: 2.7/5** — 功能完整，视觉设计有待提升。

---

## 二、具体问题清单

### P0 — 严重影响 UX

#### 1. 无响应式设计
**文件**: `Layout.vue:549-556`, 全局
- 侧栏固定 220px，主内容区未设置 `min-width: 0` 防止溢出
- 无 `@media` 断点，平板以下无法正常使用
- 仪表盘网格 `grid-template-columns: repeat(auto-fill, minmax(240px, 1fr))` 在窄屏仍可排列但内容区未适配

**建议**: 增加 768px / 1024px 两个响应式断点，侧栏在小屏折叠为底部 TabBar。

#### 2. 可访问性缺失
**文件**: `Layout.vue:332-348`, `Dashboard.vue:89-100`
- 侧栏 `nav-item` 无 `aria-current="page"` 标记当前页
- stat-card 无 `role="button"` 和 `tabindex="0"`，键盘无法操作
- 缺少 skip-link 跳过导航区
- TDesign t-input 无 `aria-label` 关联

**建议**: 补充核心交互元素的 ARIA 标记，增加 skip-link 组件。

#### 3. 无暗色模式完整覆盖
**文件**: `App.vue:73-81`
- dark theme 仅定义 6 个变量，大量硬编码颜色未切换
- 侧栏 `background: #1e293b` (Layout.vue:551) 硬编码，不跟随主题
- Dashboard stat-card 内联 `background: #fef2f2` 等亮色固定

**建议**: 将所有硬编码颜色迁移至 `var(--color-*)` 变量，dark 模式补全。

---

### P1 — 视觉品质

#### 4. 内联样式过多
**文件**: `Dashboard.vue:91-140`, 多处
```html
<!-- 当前写法 -->
<div class="stat-icon" style="background: #fef2f2; color: #ef4444;">
```
**建议**: 改用 CSS 类 + 语义变量，如 `.stat-icon.sites { --stat-bg: var(--color-warning-light); --stat-color: var(--color-warning); }`

#### 5. 页面缺少统一 Header 组件
**文件**: `Dashboard.vue:81-86`, `users/List.vue:3-15`
- Dashboard 使用自定义 page-header，User List 使用 `PageHeader` 组件
- 样式不统一（字体大小、间距、面包屑缺失）

**建议**: 所有页面统一使用 `PageHeader` 组件，组件内补充面包屑路径。

#### 6. 表格页面操作区布局不统一
**文件**: `entries/List.vue:446-491`, `users/List.vue`, `tokens/List.vue`
- entries 用 `toolbar-left` / `toolbar-right` 左右分区
- users 用单行 flex 布局
- tokens 已修复为右对齐

**建议**: 提取 `<TableToolbar>` 公共组件，统一左筛选右操作的布局模式。

#### 7. 空状态设计单调
**文件**: 各列表页 `t-table` 空状态
- 默认 TDesign 空状态不带图标和引导文案
- 缺少"创建第一个 X"的引导按钮

**建议**: 定制空状态 slot，加入图标 + 描述 + 创建引导按钮。

#### 8. 间距系统不统一
**文件**: `Layout.vue:563`, `Dashboard.vue:89`
- 侧栏 padding: 16px 12px，主内容区 padding 20px/24px 混用
- 卡片间距 16px/20px/24px 三种混用

**建议**: 建立 4px 基准间距体系（8/12/16/20/24/32），全局统一。

---

### P2 — 微交互与细节

#### 9. 缺少过渡动画
**文件**: 全局
- 侧栏展开/折叠仅 `transition: width 0.3s`，内容无渐变
- 页面切换无过渡效果
- stat-card hover 无微动效（仅有 cursor: pointer）

**建议**: 侧栏折叠时给 nav-label 加 opacity/fade 动画；card hover 加 `translateY(-2px)` 微动效。

#### 10. Dashboard 图表缺失
**文件**: `Dashboard.vue`
- 仪表盘仅有 6 个 stat card，无趋势图/分布图
- quickActions 区为纯列表，无直观的快捷入口卡片

**建议**: 增加近期内容创建趋势图（折线图）、内容类型分布（饼图），使用 chart.js 或项目已有依赖。

#### 11. 表单错误提示位置不统一
**文件**: `users/List.vue:64`
- 自定义 span.form-error vs TDesign t-form 的 help slot
- 密码强度条与表单验证分离

**建议**: 统一使用 TDesign t-form-item 的 `#help` slot 展示错误信息。

#### 12. 登录页背景图加载无降级
**文件**: `Login.vue:2`
- `backgroundImage` 通过后端配置动态提供
- 加载失败时无 fallback 纯色背景

**建议**: 添加 CSS fallback `background-color: var(--color-bg)` 兜底。

---

## 三、架构优化建议

### 1. 提取设计 Token 文件
```
src/styles/tokens.css  ← 集中所有 --color-* --space-* --font-* 变量
```
当前散布在 `App.vue`、内联 style、组件 scoped CSS 三处。

### 2. 统一组件目录结构
```
src/components/
  ├── layout/          ← 分出 Layout.vue 的子组件
  │   ├── AppHeader.vue
  │   ├── AppSidebar.vue
  │   └── UserDropdown.vue
  ├── common/           ← 通用组件
  │   ├── PageHeader.vue       ← 已存在
  │   ├── TableToolbar.vue     ← 新增
  │   ├── EmptyState.vue       ← 新增
  │   └── StatCard.vue         ← 新增
  └── ...
```

### 3. 暗色模式完整方案
- 所有硬编码颜色 → CSS 变量
- `[data-theme="dark"]` 补全 20+ 变量
- 切换按钮放在 Header 右侧（当前无入口）

### 4. 响应式策略
| 断点 | 宽度 | 布局变化 |
|------|------|---------|
| Mobile | < 768px | 侧栏折叠为底部 Tab，表格横向滚动 |
| Tablet | 768-1023px | 侧栏可折叠，表格自适应 |
| Desktop | ≥ 1024px | 当前固定侧栏布局 |

---

## 四、快速胜率清单（推荐优先修复）

1. ✅ **统一所有页面的 PageHeader** — 改动 4 个文件 30 分钟
2. ✅ **提取 stat-card 内联样式为 CSS 类** — 改动 Dashboard.vue 15 分钟
3. ✅ **TableToolbar 公共组件** — 新增组件，改 3 个页面 1 小时
4. ✅ **focus 样式 + aria 标签** — 改动 Layout.vue / Dashboard.vue 30 分钟
5. ✅ **暗色模式变量补全** — 改动 App.vue + 全局 CSS 1 小时
6. ✅ **空状态定制** — 改动各列表页 1 小时
7. ⬜ 响应式断点 — 3 小时
8. ⬜ Dashboard 图表 — 4 小时
9. ⬜ 过渡动画 — 2 小时

---

## 五、参考规范

- [WCAG 2.1 AA](https://www.w3.org/TR/WCAG21/) — 可访问性标准
- [TDesign Vue Next](https://tdesign.tencent.com/vue-next/) — 组件库约定
- 项目代码规范: `ai/soul.md` — 路由/命名/结构约定
