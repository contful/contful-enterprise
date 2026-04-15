# Contful

开源 Headless CMS，为您提供更方便、快捷的内容管理。

## 特性

- **多站点隔离** - 通过 site_id 实现数据隔离
- **动态内容模型** - 支持灵活的内容类型定义
- **API 优先** - Admin API + Open API 分离设计
- **容器化部署** - Docker/Kubernetes 友好

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM v2 |
| 数据库 | PostgreSQL 18 |
| 缓存 | Redis 7 |
| 前端 | Vue 3 / TDesign Next / Pinia |
| 文档 | VitePress |

## 项目结构

```
contful/
├── admin/           # Admin API (端口 8080)
├── open/            # Open API (端口 8081)
├── console/         # 管理控制台前端
├── docker/          # Docker 配置
└── migrations/      # 数据库迁移
```

## 快速开始

### 前置要求

- Go 1.22+
- Node.js 18+
- PostgreSQL 18
- Redis 7
- Docker (可选)

### 本地开发

#### 1. 启动数据库和 Redis

```bash
# 使用已有 PostgreSQL 和 Redis，或通过 Docker 启动
docker run -d --name contful-postgres -e POSTGRES_USER=contful -e POSTGRES_PASSWORD=contful_dev -e POSTGRES_DB=contful -p 5432:5432 postgres:18
docker run -d --name contful-redis -p 6379:6379 redis:7
```

#### 2. 启动后端服务

```bash
# Admin API
cd admin
go mod tidy
go run cmd/server/main.go

# Open API (新终端)
cd open
go mod tidy
go run cmd/server/main.go
```

#### 3. 启动前端

```bash
cd console
npm install
npm run dev
```

#### 4. 启动文档站

```bash
cd ../website
npm install
npm run docs:dev
```

### Docker 部署

```bash
cd docker
cp .env.example .env
# 编辑 .env 配置数据库和 Redis 连接
docker compose up -d
```

## API 文档

- [Admin API](https://contful.dev/docs/api/admin-api/overview)
- [Open API](https://contful.dev/docs/api/open-api/overview)

## roadmap

详见 [Roadmap](https://contful.dev/guide/roadmap)

- M0 (Q2 2026) - MVP 基础功能
- M1 (Q2 2026) - 用户认证与权限
- M2 (Q3 2026) - 插件系统
- M3 (Q3 2026) - 国际化

## 许可证

MIT License
