# Contful

> 🏠 [返回首页](../README.md) &nbsp;|&nbsp; 🇨🇳 [简体中文](README_zh-CN.md) &nbsp;|&nbsp; 🇭🇰 [繁體中文](README_zh-TW.md) &nbsp;|&nbsp; 🇺🇸 [English](README_en.md) &nbsp;|&nbsp; 🇰🇷 [한국어](README_ko.md) &nbsp;|&nbsp; 🇯🇵 [日本語](README_ja.md)

开源 Headless CMS，支持多站点管理。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM |
| 前端 | Vue 3.4+ / TDesign |
| 数据库 | PostgreSQL 16+ |
| 缓存 | Valkey 9+ |

## 项目结构

```
contful/
├── admin/            # Admin API 服务（:9080）
├── openapi/          # Open API 服务（:8080）
├── console/          # Vue 3 控制台（:3000）
├── sql/              # 数据库初始化 SQL
├── docker/           # Docker 配置（Dockerfile + docker-compose.yaml）
├── shell/            # 构建脚本
├── build/            # 编译产物（.gitignore）
├── logs/             # 日志文件（.gitignore）
└── uploads/          # 用户上传（.gitignore）
```

## 快速开始

### 默认账号

首次部署后使用以下账号登录管理后台：

| 字段 | 值 |
|------|-----|
| 邮箱 | `admin@contful.com` |
| 密码 | `contful@com` |

> ⚠️ **安全提示**：首次登录后请立即修改密码。

### 前置条件

- PostgreSQL 18
- Valkey 9+
- Go 1.22+
- Node.js 18+

### 方式一：Docker 部署

```bash
# 1. 构建镜像（在 contful/ 目录执行）
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .

# 2. 编辑配置文件
#    - conf/console.yaml   # Console 服务配置
#    - conf/openapi.yaml   # Open API 服务配置
#    配置文件中已预置默认值，只需修改数据库密码等敏感信息

# 3. 启动服务
docker-compose -f docker/docker-compose.yaml up -d

# 访问
#   管理后台:  http://localhost         (Console + Admin API)
#   Open API: http://localhost:8080/   (直连)
```

> **提示**：构建命令在 `contful/` 目录执行，构建上下文为当前目录。

### 方式二：本地开发

```bash
# 1. 复制环境变量配置
cp conf/.env.example .env

# 2. 启动数据库和缓存（使用远程或 Docker 本地）
docker run -d --name contful-postgres -p 5432:5432 -e POSTGRES_PASSWORD=xxx postgres:18-alpine
docker run -d --name contful-redis -p 6379:6379 redis:7-alpine

# 2. 初始化数据库
psql -h <host> -U <user> -d contful -f sql/init_pg.sql


# 3. 构建
./shell/build.sh

# 4. 启动服务
./shell/dev.sh start

# 访问
#   管理后台:  http://localhost:3000   (Console + Admin API :9080)
#   Open API: http://localhost:8080/
```

### 方式三：分别启动

```bash
# 构建
./shell/build.sh console    # 构建 Console（Admin API + 前端）
./shell/build.sh openapi    # 构建 Open API

# 单独启动某个服务
./shell/dev.sh logs admin   # 查看 Admin API 日志
./shell/dev.sh status       # 查看服务状态
./shell/dev.sh stop         # 停止所有服务
```

## 脚本说明

| 脚本 | 用途 |
|------|------|
| `./shell/build-image.sh` | 构建 Docker 镜像（用于部署） |
| `./shell/build.sh` | 本地编译（生成 build/ 目录产物） |
| `./shell/dev.sh` | 本地开发启动（编译 + 运行） |

### 构建参数

```bash
# PostgreSQL 版本（默认）
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .


# ARM64 平台
docker build -f docker/Dockerfile.console -t contful/console:pg-arm64 --platform linux/arm64 .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-arm64 --platform linux/arm64 .
```

> **注意**：构建命令在 `contful/` 目录执行。使用 `--platform` 参数交叉编译时，`TARGETOS` 和 `TARGETARCH` 会自动适配。

## 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| Console | 3000 | Vue 管理后台（开发模式） / 80（Docker） |
| Admin API | 9080 | 管理后台 API |
| Open API | 8080 | 内容 API，可水平扩展 |

## 站点默认配置

新站点创建时会自动写入以下默认配置（存储在 `site_configs` 表）：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `storage.driver` | `local` | 存储驱动：`local` / `oss` / `cos` / `obs` / `s3` |
| `storage.local.root` | `uploads` | 本地存储根目录 |
| `storage.local.base_url` | `/uploads` | 本地存储访问路径 |
| `integrity.enabled` | `false` | 是否启用数据签名（HMAC-SHA256） |
| `integrity.algorithm` | `HMAC-SHA256` | 签名算法 |
| `integrity.signing_key` | _(空) | 签名密钥，AES-256-GCM 加密存储；`integrity.enabled=true` 时自动生成 |

> **提示**：敏感配置（`integrity.signing_key` 等）通过 `CONTFUL_CONFIG_MASTER_KEY` 环境变量加密存储。
> 生产环境请设置 32 字节随机字符串作为主密钥：
> ```bash
> openssl rand -hex 32
> ```

## 文档

- [快速开始](https://contful.com/docs/getting-started)
- [部署指南](https://contful.com/docs/deployment)
- [系统架构](https://contful.com/docs/architecture/overview)
- [Admin API 文档](https://contful.com/docs/api/admin-api/overview)
- [Open API 文档](https://contful.com/docs/api/open-api/overview)
- [数据库 Schema](https://contful.com/docs/database/schema)
- [贡献指南](https://contful.com/docs/community/contributing)
- [更新日志](https://contful.com/guide/release)
