<!-- فارسی (Brasil) -->
# Contful

> 🏠 [Back to root](../README.md) &nbsp;|&nbsp; 🇨🇳 [简体中文](README_zh-CN.md) &nbsp;|&nbsp; 🇭🇰 [繁體中文](README_zh-TW.md) &nbsp;|&nbsp; 🇺🇸 [English](README_en.md) &nbsp;|&nbsp; 🇰🇷 [한국어](README_ko.md) &nbsp;|&nbsp; 🇯🇵 [日本語](README_ja.md)

Open source Headless CMS with multi-site management support.

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.25 / Gin / GORM |
| Frontend | Vue 3.5 / TDesign / Vite 8 |
| Database | PostgreSQL 18 |
| Cache | Valkey 9 |

## Project Structure

```
contful/
├── admin/            # Admin API service (:9080)
├── openapi/          # Open API service (:8080)
├── console/          # Vue 3 console (:3000)
├── db/               # Database initialization scripts (init_pg.sql: DDL + seed data)
├── docker/           # Docker configuration (Dockerfile + docker-compose.yaml)
├── shell/            # Build scripts
├── build/            # Build artifacts (.gitignore)
├── logs/             # Log files (.gitignore)
└── uploads/          # User uploads (.gitignore)
```

## Quick Start

### Default Account

Log in to the admin dashboard with the following credentials after first deployment:

| Field | Value |
|-------|-------|
| Email | `admin@contful.com` |
| Password | `contful@com` |

> ⚠️ **Security Notice**: Please change your password immediately after first login.

### Prerequisites

- PostgreSQL 18
- Valkey 9+
- Go 1.25+
- Node.js 24+

### Method 1: Docker Deployment

```bash
# 1. Build images (execute in contful/ directory)
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .

# 2. Edit configuration files
#    - conf/console.yaml   # Console service config
#    - conf/openapi.yaml   # Open API service config
#    Default values are pre-configured; only update sensitive info like DB passwords

# 3. Start services
docker-compose -f docker/docker-compose.yaml up -d

# Access
#   Admin Dashboard:  http://localhost         (Console + Admin API)
#   Open API:        http://localhost:8080/   (direct)
```

> **Note**: Build commands are executed in the `contful/` directory with the current directory as the build context.

### Method 2: Local Development

```bash
# 1. Copy environment variable configuration
cp .env.example .env

# 2. Start database and cache (use remote or Docker local)
docker run -d --name contful-postgres -p 5432:5432 -e POSTGRES_PASSWORD=xxx postgres:18-alpine
docker run -d --name contful-redis -p 6379:6379 redis:7-alpine

# 2. Initialize database
psql -h <host> -U <user> -d contful -f db/init_pg.sql


# 3. Build
./shell/build.sh

# 4. Start services
./shell/dev.sh start

# Access
#   Admin Dashboard:  http://localhost:3000   (Console + Admin API :9080)
#   Open API:        http://localhost:8080/
```

### Method 3: Start Individually

```bash
# Build
./shell/build.sh console    # Build Console (Admin API + frontend)
./shell/build.sh openapi    # Build Open API

# Start a specific service
./shell/dev.sh logs admin   # View Admin API logs
./shell/dev.sh status       # Check service status
./shell/dev.sh stop         # Stop all services
```

## Script Reference

| Script | Purpose |
|--------|---------|
| `./shell/build-image.sh` | Build Docker images (for deployment) |
| `./shell/build.sh` | Local compilation (outputs to build/) |
| `./shell/dev.sh` | Local dev startup (compile + run) |

### Build Arguments

```bash
# PostgreSQL version (default)
docker build -f docker/Dockerfile.console -t contful/console:pg-amd64-latest .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-amd64-latest .


# ARM64 platform
docker build -f docker/Dockerfile.console -t contful/console:pg-arm64 --platform linux/arm64 .
docker build -f docker/Dockerfile.openapi -t contful/openapi:pg-arm64 --platform linux/arm64 .
```

> **Note**: Build commands are executed in the `contful/` directory. When cross-compiling with `--platform`, `TARGETOS` and `TARGETARCH` are automatically configured.

## Service Reference

| Service | Port | Description |
|---------|------|-------------|
| Console | 3000 | Vue admin dashboard (dev mode) / 80 (Docker) |
| Admin API | 9080 | Admin dashboard API |
| Open API | 8080 | Content API, horizontally scalable |

## Site Default Configuration

When a new site is created, the following default configuration is automatically written (stored in `contful_system_config` table):

| Config Key | Default Value | Description |
|------------|---------------|-------------|
| `storage.driver` | `local` | Storage driver: `local` / `oss` / `cos` / `obs` / `s3` |
| `storage.local.root` | `uploads` | Local storage root directory |
| `storage.local.base_url` | `/uploads` | Local storage access path |
| `integrity.enabled` | `false` | Enable data integrity signing (HMAC-SHA256) |
| `integrity.algorithm` | `HMAC-SHA256` | Signing algorithm |
| `integrity.signing_key` | _(empty) | Signing key, AES-256-GCM encrypted; auto-generated when `integrity.enabled=true` |

> **Tip**: Sensitive configs (`integrity.signing_key`, etc.) are encrypted via the `CONTFUL_CONFIG_MASTER_KEY` environment variable.
> For production, set a 32-byte random string as the master key:
> ```bash
> openssl rand -hex 32
> ```

## Documentation

- [Getting Started](https://contful.com/guide/quickstart)
- [Deployment Guide](https://contful.com/guide/deploy/)
- [Architecture Overview](https://contful.com/guide/architecture/overview)
- [Admin API Reference](https://contful.com/api/admin-api/overview)
- [Open API Reference](https://contful.com/api/open-api/overview)
- [Contributing Guide](https://contful.com/about/developers)
- [Release Notes](https://contful.com/guide/changelog)
