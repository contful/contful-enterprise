# Contful

开源 Headless CMS，支持多站点管理。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM |
| 前端 | Vue 3.4+ / TDesign |
| 数据库 | PostgreSQL 16+ |
| 缓存 | Redis 7+ |

## 项目结构

```
contful/
├── admin-api/        # Admin API 服务 (:8080)
├── open-api/         # Open API 服务 (:8081)
├── console/          # Vue 3 控制台
├── migrations/        # 数据库迁移
├── docs/             # 本文档目录
└── deploy/           # 部署配置
```

## 快速开始

```bash
# 启动依赖
docker-compose -f deploy/docker-compose.yml up -d postgres redis

# 运行迁移
cd migrations
 migrate -database DATABASE_URL up

# 启动 Admin API
cd admin-api
go run cmd/server/main.go

# 启动 Open API
cd open-api
go run cmd/server/main.go

# 启动控制台
cd console
npm install && npm run dev
```

## 访问

- 控制台: http://localhost:3000
- Admin API: http://localhost:8080
- Open API: http://localhost:8081

## 相关链接

- [完整文档](../website/docs/)
- [贡献指南](./CONTRIBUTING.md)
- [更新日志](./changelog.md)
