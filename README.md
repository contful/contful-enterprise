# Contful

> 🌍 多语言文档 / Multilingual Documentation: [简体中文](doc/README_zh-CN.md) · [繁體中文](doc/README_zh-TW.md) · [English](doc/README_en.md) · [한국어](doc/README_ko.md) · [日本語](doc/README_ja.md)

开源 Headless CMS，支持多站点管理。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.25 / Gin / GORM |
| 前端 | Vue 3.5 / TDesign / Vite 8 |
| 数据库 | PostgreSQL 18 |
| 缓存 | Valkey 9 |

## 项目结构

```
contful/
├── admin/            # Admin API 服务（:9080）
├── openapi/          # Open API 服务（:8080）
├── console/          # Vue 3 控制台（:3000）
├── db/               # 数据库初始化脚本（init_pg.sql：DDL + 种子数据）
├── docker/           # Docker 配置（Dockerfile + docker-compose.yaml）
├── shell/            # 构建脚本
├── build/            # 编译产物（.gitignore）
├── logs/             # 日志文件（.gitignore）
└── uploads/           # 用户上传（.gitignore）
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
- Go 1.25+
- Node.js 24+

## 🐳 Docker 部署

### 方式一：直接拉取镜像（推荐）

无需从源码构建，直接使用 Docker Hub 上的预构建镜像：

**1. 启动数据库和缓存**

```bash
docker compose -f docker/docker-database.yaml up -d
```

**2. 初始化数据库**

```bash
# 创建数据库（如果尚未创建）
docker exec -i contful-postgres psql -U postgres -c "CREATE DATABASE contful;"

# 导入表结构和种子数据
docker exec -i contful-postgres psql -U postgres -d contful < db/init_pg.sql
```

**3. 配置环境变量**

```bash
cp .env.example .env
# 编辑 .env，填入数据库和缓存连接信息
```

**4. 拉取镜像并启动**

```bash
docker pull contful/console:postgresql-latest
docker pull contful/openapi:postgresql-latest
docker compose -f docker/docker-compose.yaml --env-file .env up -d
```

**5. 访问**

| 服务 | 地址 |
|------|------|
| 管理后台 | http://localhost |
| Open API | http://localhost:8080 |

> ⚠️ **注意**：Docker 镜像不包含数据库服务，需要单独启动 PostgreSQL + Valkey（步骤 1）。

### 方式二：从源码构建 Docker 镜像

前置条件：Docker + Docker Compose、PostgreSQL 14+ 实例、Valkey/Redis 7+ 实例。

步骤 1-3 同方式一（启动 DB、初始化、配置 .env），然后：

```bash
# 构建镜像（在 contful/ 目录执行）
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .

# 启动
docker compose -f docker/docker-compose.yaml --env-file .env up -d
```

> **提示**：构建命令在 `contful/` 目录执行，构建上下文为当前目录。

### 方式三：Docker Compose 一键启动（含自动初始化）

如果你的数据库为空，可以使用自动初始化模式：

**1. 配置环境变量**

```bash
cp .env.example .env
# 编辑 .env，填入你的数据库和缓存连接信息
```

**2. 准备挂载目录**

```bash
mkdir -p conf uploads logs
```

**3. 一键启动**

```bash
docker compose -f docker/docker-compose.yaml --env-file .env up -d
```

首次启动会自动完成：

- ✅ 数据库表创建（`init_pg.sql`）
- ✅ 种子数据导入（管理员账号等）
- ✅ 非对称密钥对生成（登录加密）

### 故障排查 FAQ

| 问题 | 解决方案 |
|------|----------|
| 启动后 DB 未初始化？ | 确认 `DB_HOST` / `DB_NAME` 正确，空数据库会自动初始化 |
| 容器反复重启？ | 检查 DB 连接参数，`docker logs contful-console` 查看详细日志 |
| 如何重置数据库？ | 删表后重启容器即可自动重建 |
| 如何更换密钥？ | `rm ./conf/keys/*.pem && docker restart contful-console` |

### 方式四：从源码安装（非 Docker）

适用于没有 Docker 环境、或希望从源码编译部署的场景。

**前置条件**

| 组件 | 版本要求 |
|------|----------|
| Go | 1.25+ |
| Node.js | 24+ |
| PostgreSQL | 14+ |
| Valkey / Redis | 7+ |

**1. 克隆项目**

```bash
git clone https://github.com/contful/contful.git
cd contful/contful
```

**2. 创建数据库**

```bash
# 连接 PostgreSQL
psql -U postgres

# 在 psql 中执行
CREATE DATABASE contful;
\q
```

**3. 导入数据库表结构和种子数据**

```bash
# init_pg.sql 包含完整 DDL + 种子数据，一条命令即可
psql -U postgres -d contful -f db/init_pg.sql
```

导入完成后，数据库包含默认管理员账号：`admin@contful.com` / `contful@com`。

**4. 配置环境变量**

```bash
cp .env.example .env
# 编辑 .env，填写你的数据库和缓存连接信息
```

关键配置项：

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<你的数据库密码>
DB_NAME=contful
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=<你的 Redis 密码>
SECRET=<运行 openssl rand -hex 32 生成>
CRYPTO_MODE=rsa
```

**5. 生成密钥对（用于登录加密）**

```bash
# RSA 模式（默认）
openssl genrsa -out conf/keys/private.pem 2048
openssl rsa -in conf/keys/private.pem -pubout -out conf/keys/public.pem

# SM2 模式（需 openssl 3.0+ 或 gmssl）
# openssl ecparam -genkey -name SM2 -out conf/keys/private.pem
# openssl ec -in conf/keys/private.pem -pubout -out conf/keys/public.pem
```

**6. 调整配置文件**

编辑 `conf/console.yaml`，确认密钥路径：

```yaml
security:
  pubkey_path: "conf/keys/public.pem"
  privkey_path: "conf/keys/private.pem"
  crypto_mode: "rsa"
```

**7. 编译并启动**

```bash
# 编译后端
./shell/build.sh

# 启动全部服务
./shell/dev.sh start
```

**8. 访问**

| 服务 | 地址 |
|------|------|
| 管理后台 | http://localhost:3000 |
| Admin API | http://localhost:9080 |
| Open API | http://localhost:8080 |

> 使用种子数据中的管理员账号登录：`admin@contful.com` / `contful@com`，首次登录后请立即修改密码。

**9. 验证**

```bash
# Admin API 健康检查
curl http://localhost:9080/health

# 查看服务状态
./shell/dev.sh status
```

### 方式五：服务管理

```bash
./shell/dev.sh logs admin   # 查看 Admin API 日志
./shell/dev.sh logs openapi # 查看 Open API 日志
./shell/dev.sh logs console # 查看 Console 日志
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

新站点创建时会自动写入以下默认配置（存储在 `contful_system_config` 表）：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `storage.driver` | `local` | 存储驱动：`local` / `oss` / `cos` / `obs` / `s3` |
| `storage.local.root` | `uploads` | 本地存储根目录 |
| `storage.local.base_url` | `/uploads` | 本地存储访问路径 |
| `integrity.enabled` | `false` | 是否启用数据签名（HMAC，算法取决于 `crypto_mode`） |
| `integrity.algorithm` | `HMAC-SHA256` | 签名算法（`rsa` 模式为 HMAC-SHA256，`sm` 模式为 SM3） |
| `integrity.signing_key` | _(空) | 签名密钥，AES-256-GCM 加密存储；`integrity.enabled=true` 时自动生成 |

> **提示**：敏感配置（`integrity.signing_key` 等）通过 `CONTFUL_CONFIG_MASTER_KEY` 环境变量加密存储。
> 生产环境请设置 32 字节随机字符串作为主密钥：
> ```bash
> openssl rand -hex 32
> ```

## 文档

- [快速开始](https://contful.com/guide/quickstart)
- [部署指南](https://contful.com/guide/deploy/)
- [系统架构](https://contful.com/guide/architecture/overview)
- [Admin API 文档](https://contful.com/api/admin-api/overview)
- [Open API 文档](https://contful.com/api/open-api/overview)
- [贡献指南](https://contful.com/about/developers)
- [更新日志](https://contful.com/guide/changelog)
