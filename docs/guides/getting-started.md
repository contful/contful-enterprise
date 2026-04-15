# 快速开始

> 版本: v1.0.0 | 更新日期: 2026-04-15

---

## 1. 环境要求

| 组件 | 最低版本 | 推荐版本 |
|------|----------|----------|
| Go | 1.22 | 1.22+ |
| Node.js | 18 | 20+ |
| PostgreSQL | 15 | 18 |
| Redis | 7 | 7 |
| Docker | 24 | Latest |

---

## 2. 快速启动 (Docker)

### 2.1 克隆项目

```bash
git clone https://github.com/contful/contful.git
cd contful
```

### 2.2 启动服务

```bash
# 启动所有服务
docker-compose -f docker-compose.dev.yml up -d

# 查看服务状态
docker-compose -f docker-compose.dev.yml ps
```

### 2.3 访问服务

| 服务 | 地址 |
|------|------|
| Console | http://localhost:3000 |
| Admin API | http://localhost:8080 |
| Open API | http://localhost:8081 |
| API 文档 | http://localhost:8080/docs |

---

## 3. 本地开发

### 3.1 数据库准备

```bash
# 创建数据库
psql -U postgres -c "CREATE DATABASE contful;"

# 运行迁移
cd contful/contful/migrations
migrate -path . -database "postgres://postgres:postgres@localhost:5432/contful?sslmode=disable" up
```

### 3.2 配置环境变量

```bash
# 复制配置模板
cp .env.example .env

# 编辑配置
vim .env
```

关键配置项：
```env
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_NAME=contful
DB_USER=contful
DB_PASSWORD=your_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your_256bit_secret

# Admin API
ADMIN_API_PORT=8080

# Open API
OPEN_API_PORT=8081
```

### 3.3 启动后端服务

```bash
# 进入 API 目录
cd contful/admin-api

# 安装依赖
go mod download

# 运行服务
go run cmd/server/main.go
```

### 3.4 启动前端 Console

```bash
# 新开终端
cd contful/console

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

---

## 4. 初始化配置

### 4.1 创建超级管理员

首次启动后，通过以下命令创建管理员：

```bash
curl -X POST http://localhost:8080/admin/v1/auth/init \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "YourPassword123!",
    "nickname": "Administrator"
  }'
```

### 4.2 登录 Console

1. 打开 http://localhost:3000
2. 使用创建的账号登录
3. 开始管理站点

---

## 5. 基本使用

### 5.1 创建第一个站点

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/admin/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "YourPassword123!"}'

# 创建站点
curl -X POST http://localhost:8080/admin/v1/sites \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "我的网站", "slug": "my-site"}'
```

### 5.2 创建内容类型

```bash
# 创建文章类型
curl -X POST http://localhost:8080/admin/v1/content-types \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "文章",
    "slug": "articles",
    "description": "博客文章",
    "kind": "collection"
  }'
```

### 5.3 添加字段

```bash
# 添加标题字段
curl -X POST http://localhost:8080/admin/v1/content-types/$TYPE_ID/fields \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "title",
    "label": "标题",
    "field_type": "text",
    "config": {"max_length": 200}
  }'
```

---

## 6. 下一步

- [部署指南](deployment.md) - 生产环境部署
- [系统架构](../architecture/overview.md) - 了解更多架构细节
- [API 文档](../api/admin-api/overview.md) - API 使用详解
