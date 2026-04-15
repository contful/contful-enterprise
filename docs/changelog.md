# Changelog

> 本文件记录 Contful 各版本的更新内容。格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/) 规范。

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- 生成器会自动在下方插入版本内容，请勿手动修改 -->

<!-- INSERT_CHANGELOG_HERE -->

<!-- 版本模板
## [Unreleased]

### Added
- 新功能

### Changed
- 变更内容

### Deprecated
- 已废弃功能

### Fixed
- 问题修复

### Security
- 安全相关
-->

## [M0.1.0] - 2026-04-15

### Added
- 项目初始化完成
- Go + Gin 后端脚手架搭建
- Vue 3 + TDesign 前端脚手架
- CI/CD 流水线配置（GitHub Actions）
- VitePress 文档站初始化
- Admin API 基础架构（JWT 认证）
- Open API 基础架构（API Token）
- PostgreSQL 数据库 Schema 设计
- 多站点架构设计文档

### Planned
- [M1] MVP 功能开发（用户系统、内容类型、内容管理、媒体库、API Token）
- [M2] 插件系统
- [M3] 高级功能（i18n、内容关系、版本历史）

## 版本说明

| 版本 | 说明 |
|------|------|
| M0.x | 项目初始化与基础架构 |
| M1.x | MVP 功能集 |
| M2.x | 插件系统 |
| M3.x | 高级功能 |

## 提交规范

本项目使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

| 类型 | 说明 |
|------|------|
| feat | 新功能 |
| fix | 问题修复 |
| docs | 文档变更 |
| style | 代码格式（不影响功能） |
| refactor | 重构（不影响功能） |
| perf | 性能优化 |
| test | 测试相关 |
| chore | 构建/工具变更 |
