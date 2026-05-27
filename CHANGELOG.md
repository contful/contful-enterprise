# Changelog

所有重要变更记录。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [1.3.0] - 2026-05-27

### Added

- 定时发布排期功能（Cron 调度 + 日历视图）
- 仪表盘授权信息卡片
- Docker 多架构镜像构建（amd64 + arm64）
- PROXY 国内镜像加速

### Changed

- conf/ 目录重组（keys/ 与 ssl/ 分离）
- entrypoint.sh 完整重写
- 文件上传改用 StorageProvider 统一抽象

### Fixed

- Header 站点下拉刷新后消失（竞态修复）
- Layout 版本号动态读取 package.json

## [1.2.0] - 2026-05-20

### Added

- Docker 一键启动（auto-init）
- crypto_mode 可插拔加密架构（RSA/SM2）
- SM2+SM3+SM4 国密支持

## [1.1.0] - 2026-05-15

### Added

- API Token 管理
- 审计日志（HMAC 防篡改）

## [1.0.0] - 2026-05-01

### Added

- 多站点管理
- 内容模型 + 内容条目
- 媒体库（本地/OSS/S3/COS/OBS）
- RBAC 权限体系
- JWT 认证 + MFA 双因子
- Admin API + Open API
- Vue 3 管理后台（TDesign）
