# 贡献指南

感谢你对 Contful 的关注！我们欢迎任何形式的贡献。

## 贡献流程

### 1. Fork 项目

在 GitHub 上 Fork [contful/contful](https://github.com/contful/contful) 仓库到你自己的账号下。

### 2. Clone 到本地

```bash
git clone https://github.com/<你的用户名>/contful.git
cd contful
```

### 3. 创建功能分支

```bash
git checkout -b feature/<功能名称>
# 或
git checkout -b fix/<问题名称>
```

分支命名建议使用英文小写，单词之间用 `-` 连接。

### 4. 开发

在本地进行代码开发和测试，确保现有功能不受影响。

### 5. 提交代码

请遵循以下 Commit 规范：

- `feat:` — 新功能
- `fix:` — 修复 Bug
- `docs:` — 文档更新
- `refactor:` — 代码重构
- `test:` — 测试相关
- `chore:` — 构建/工具/依赖等杂项

示例：

```bash
git commit -m "feat: 添加定时发布排期功能"
git commit -m "fix: 修复 Header 站点下拉刷新后消失的问题"
git commit -m "docs: 更新 README 部署指南链接"
```

### 6. 推送并提交 PR

```bash
git push origin feature/<功能名称>
```

然后在 GitHub 上创建 Pull Request，请确保：

- 关联相关 Issue（如有）
- 描述变更内容和原因
- 说明测试验证方式
- 确保 CI 流水线通过

## 代码规范

### Go

- 遵循 [Google Go Style Guide](https://google.github.io/styleguide/go/)
- 使用 `gofmt` 格式化代码
- 使用有意义的变量和函数名
- 为公开函数和结构体添加注释

### Vue / TypeScript

- 遵循项目中的 ESLint 配置
- 使用 `<script setup lang="ts">` 语法
- 组件命名使用 PascalCase
- 文件命名使用 kebab-case

## 行为准则

请参阅 [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)。

## 联系方式

如有任何问题，请发送邮件至 [hi@reepu.com](mailto:hi@reepu.com)。
