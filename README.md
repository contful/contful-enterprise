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
├── admin/        # Admin API 服务
├── open/         # Open API 服务
├── console/      # Vue 3 控制台
├── sql/          # 数据库初始化 SQL
├── docker/       # Docker 配置
└── website/      # 官网 & 文档 (VitePress)
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
psql -h <host> -U <user> -d contful -f sql/init.sql
```

### 2. Docker 启动

```bash
cd docker
cp .env.example .env
# 编辑 .env，填入 DB_HOST、DB_PASSWORD、JWT_SECRET

# 启动全部服务（Admin + Open API）
docker-compose --env-file .env up -d

# 访问
#   管理后台:  http://localhost         (Console + Admin API)
#   Open API: http://localhost:8080/   (直连)
```

### 3. 水平扩展 Open API

```bash
# 扩展 Open API 到 3 个实例
docker-compose --env-file .env up -d --scale api=3

# 生产环境建议在 Open API 前加 Nginx/HAProxy 负载均衡
```

### 4. 本地开发启动

```bash
cd admin && go run cmd/server/main.go   # Admin API (:8080)
cd open && go run cmd/server/main.go    # Open API (:8081)
cd console && yarn dev                   # Console (:3000)
```

## 服务说明

| 服务 | 端口 | 说明 |
|------|------|------|
| contful-admin | 80 | 管理后台（OpenResty → Console SPA + /admin/ 代理） |
| contful-api | 8080 | Open API，可水平扩展 |

## 文档

- [快速开始](https://contful.com/docs/getting-started)
- [部署指南](https://contful.com/docs/deployment)
- [系统架构](https://contful.com/docs/architecture/overview)
- [Admin API 文档](https://contful.com/docs/api/admin-api/overview)
- [Open API 文档](https://contful.com/docs/api/open-api/overview)
- [数据库 Schema](https://contful.com/docs/database/schema)
- [贡献指南](https://contful.com/docs/community/contributing)
- [更新日志](https://contful.com/guide/release)
