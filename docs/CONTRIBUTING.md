# 贡献指南

## 分支策略

```
main          ← 生产环境代码
  │
  ├── feature/*   ← 功能分支 (从 main 创建)
  └── fix/*       ← Bug 修复分支 (从 main 创建)
```

## 开发规范

### 代码规范

- Go: `gofmt` + `golangci-lint`
- JavaScript: ESLint + Prettier
- 提交前运行: `make lint`

### 提交规范 (Conventional Commits)

```
<type>(<scope>): <subject>

feat(site): add multi-site support
fix(auth): resolve token refresh issue
docs: update README
```

### 类型

| 类型 | 说明 |
|------|------|
| feat | 新功能 |
| fix | Bug 修复 |
| docs | 文档 |
| refactor | 重构 |
| test | 测试 |
| chore | 构建/工具 |

## 开发流程

```bash
# Fork 项目
# 克隆你的 Fork
git clone https://github.com/YOUR_NAME/contful.git

# 创建功能分支
git checkout -b feature/your-feature

# 开发 & 测试
make test

# 提交
git commit -m "feat(scope): description"

# 推送 & 创建 PR
git push origin feature/your-feature
```

## 代码审查

- 确保 `make test` 和 `make lint` 通过
- PR 需要至少 1 个 reviewer 批准
- 遵循项目代码风格

## 许可

贡献代码即表示同意以 MIT 许可证发布。
