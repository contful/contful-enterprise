# ADR-0002: 配置文件管理策略

## 状态

**已接受** — 2026-04-15

---

## 上下文

当前 Contful 项目使用 `.env` 文件管理配置。在 MVP 阶段这是简单直接的选择，但随着项目复杂度增加，特别是进入 Docker 部署和多环境管理阶段，需要重新评估配置方案。

### 当前配置现状

- `admin/.env.example` — Admin API 配置（17 个变量）
- `open/.env.example` — Open API 配置（15 个变量）
- `docker/.env.example` — Docker 环境配置（20+ 个变量）

### 痛点

1. **Docker 部署繁琐** — 需逐个映射环境变量，或维护多个 .env 文件
2. **配置结构扁平** — 所有变量平铺，分类不清晰
3. **缺乏类型校验** — .env 纯字符串，启动时无法验证配置合法性
4. **扩展性差** — 未来增加组件（如消息队列、缓存集群）配置会膨胀

---

## 决策

我们决定使用 **config.yaml** 作为配置文件，并采用以下策略：

### 1. 配置文件结构

```
config/
├── config.yaml              # 默认配置（可提交到 Git）
├── config.prod.yaml          # 生产环境配置（不提交）
├── config.dev.yaml           # 开发环境配置（不提交）
└── config.local.yaml         # 本地覆盖配置（不提交 .gitignore）
```

### 2. config.yaml 结构示例

```yaml
# Contful Admin API 配置
app:
  name: "Contful Admin API"
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # debug, release, test

# JWT 认证配置
jwt:
  admin:
    secret: "${ADMIN_JWT_SECRET}"  # 从环境变量读取敏感信息
    access_expire_minutes: 15
    refresh_expire_days: 7
  open:
    secret: "${OPEN_JWT_SECRET}"
    access_expire_minutes: 60
    refresh_expire_days: 30

# 数据库配置
database:
  host: "${DB_HOST}"
  port: 5432
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  name: "${DB_NAME}"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime_minutes: 30

# Redis 配置
redis:
  host: "${REDIS_HOST}"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0
  pool_size: 10

# 存储配置
storage:
  type: "local"  # local, s3, r2
  upload_dir: "./uploads"
  max_upload_size_mb: 10
  allowed_types:
    - "image/jpeg"
    - "image/png"
    - "image/gif"
    - "image/webp"
    - "application/pdf"

# 日志配置
log:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"  # stdout, file
```

### 3. 环境变量覆盖

敏感信息仍通过环境变量提供，YAML 中使用 `${VAR}` 语法引用：

```yaml
database:
  password: "${DB_PASSWORD}"  # 从环境变量读取
jwt:
  admin:
    secret: "${ADMIN_JWT_SECRET}"
```

### 4. Docker 部署方式

```yaml
# docker-compose.yml
services:
  admin-api:
    image: contful/admin-api:latest
    volumes:
      - ./config/config.prod.yaml:/app/config/config.yaml:ro
    environment:
      - ADMIN_JWT_SECRET=${ADMIN_JWT_SECRET}
      - DB_PASSWORD=${DB_PASSWORD}
    ports:
      - "8080:8080"
```

### 5. 配置加载优先级

```
命令行参数 > 环境变量 > config.local.yaml > config.{ENV}.yaml > config.yaml
```

---

## 后果

### 正面

1. **配置结构清晰** — 分层组织，易于阅读和维护
2. **Docker 友好** — 直接挂载配置文件，无需逐个环境变量映射
3. **类型安全** — 配置库支持类型校验和默认值
4. **可扩展** — 新增组件配置更自然
5. **文档化** — YAML 注释可内嵌配置说明

### 负面

1. **迁移成本** — 需将现有 .env 迁移到 YAML
2. **学习曲线** — 团队需适应新配置格式
3. **配置库依赖** — 引入 Viper 或类似库

### 中性

1. **敏感信息** — 仍需环境变量或 Vault 管理，但这是最佳实践

---

## 实现计划

| 阶段 | 任务 | 负责 |
|------|------|------|
| 1 | 创建 `config/config.yaml` 结构定义 | 架构师 |
| 2 | 引入 Viper/G_cfg 配置库 | 后端工程师 |
| 3 | 重构 Admin API 配置加载 | 后端工程师 |
| 4 | 重构 Open API 配置加载 | 后端工程师 |
| 5 | 更新 Docker 配置示例 | 后端工程师 |
| 6 | 更新文档（README、部署指南） | 文档工程师 |

---

## 被否决的方案

### 方案 A: 继续使用 .env

**否决理由**: 虽然 .env 在简单场景下足够，但随着配置复杂度增加和维护需求提升，其缺点会越来越明显。特别是在 Docker 和多环境部署场景下，不如 YAML 灵活。

### 方案 B: 纯环境变量（无配置文件）

**否决理由**: 缺乏默认值、类型校验和文档化，对于需要数十个配置项的中型项目不够友好。

---

## 参考

- [Viper - Go 配置库](https://github.com/spf13/viper)
- [12-Factor App - 配置](https://12factor.net/zh_cn/config)
