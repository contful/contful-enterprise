# Contful

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
├── admin/        # Admin API 服务（:9080）
├── openapi/      # Open API 服务（:8080）
├── console/      # Vue 3 控制台（:3000）
├── sql/          # 数据库初始化 SQL
├── docker/       # Docker 配置
└── shell/        # 构建脚本
```

## 快速开始

### 前置条件
- PostgreSQL 18
- Valkey 9+

### 1. 初始化数据库

```bash
# 创建数据库（PostgreSQL）
psql -h <host> -U <user> -c "CREATE DATABASE contful;"

# 导入初始化 SQL
psql -h <host> -U <user> -d contful -f sql/init_pg.sql

# 或达梦 DM8 版本
# sql init_dm.sql
```

### 2. Docker 启动

```bash
cd docker
cp .env.example .env
# 编辑 .env，填入 DB_HOST、DB_PASSWORD、SECRET

# 启动全部服务（Admin + Open API）
docker-compose --env-file .env up -d

# 访问
#   管理后台:  http://localhost         (Console + Admin API)
#   Open API: http://localhost:8080/   (直连)
```

### 3. 水平扩展 Open API

```bash
# 扩展 Open API 到 3 个实例
docker-compose --env-file .env up -d --scale openapi=3

# 生产环境建议在 Open API 前加 Nginx/HAProxy 负载均衡
```

### 4. 本地开发启动

```bash
# PostgreSQL 编译版本（推荐）
cd admin && go build -tags=pg -o admin_pg . && ./admin_pg    # Admin API (:9080)

# 达梦 DM8 编译版本（可选）
cd admin && go build -tags=dm -o admin_dm . && ./admin_dm

cd openapi && go build -o openapi . && ./openapi               # Open API (:8080)
cd console && npm run dev                                      # Console (:3000)
```

## 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| contful-admin | 80 | 管理后台（OpenResty → Console SPA + /admin/ 代理） |
| contful-api | 8080 | Open API，可水平扩展 |

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
