#!/bin/sh

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful Entrypoint - 双模式支持
# 模式1: 环境变量模式 - 通过环境变量配置所有参数
# 模式2: 配置文件模式 - 挂载 /app/config.yml
#
# 支持服务: console (admin-api) / openapi (openapi-server)
# =============================================================================

set -e

# 确保 /app/logs 和 /app/uploads 有写权限（非 root 用户支持）
mkdir -p /app/logs /app/uploads
if [ "$(id -u)" != "0" ]; then
    # 非 root 用户：确保目录可写
    if [ -w /app/logs ]; then
        : # 已经有写权限
    else
        echo "[Entrypoint] WARNING: /app/logs not writable, logs may fail"
    fi
fi

# 检测服务类型
if [ -f "/app/admin-api" ]; then
    SERVICE_TYPE="console"
    SERVICE_PORT="${CONTFUL_SERVER_PORT:-9080}"
elif [ -f "/app/openapi-server" ]; then
    SERVICE_TYPE="openapi"
    SERVICE_PORT="${CONTFUL_SERVER_PORT:-8080}"
else
    echo "[Entrypoint] ERROR: No service binary found"
    exit 1
fi

# MODE 环境变量: console (full) / api (api-only)
MODE="${CONTFUL_MODE:-console}"

echo "[Entrypoint] Contful starting (service: $SERVICE_TYPE, mode: $MODE)"

# =========================================================================
# 环境变量校验（确保必填项存在且有效）
# =========================================================================
validate_environment() {
    local errors=0

    # 检查 SECRET（JWT 签名密钥）
    if [ -z "$SECRET" ]; then
        echo "[Entrypoint] ERROR: SECRET environment variable is required"
        errors=$((errors + 1))
    elif [ ${#SECRET} -lt 32 ]; then
        echo "[Entrypoint] ERROR: SECRET must be at least 32 characters for security"
        errors=$((errors + 1))
    fi

    # 检查数据库连接（环境变量模式必需）
    if [ -n "$DB_HOST" ] || [ -n "$DB_PASSWORD" ] || [ -n "$SECRET" ]; then
        # 使用环境变量模式
        if [ -z "$DB_HOST" ]; then
            echo "[Entrypoint] ERROR: DB_HOST is required in environment variable mode"
            errors=$((errors + 1))
        fi
        if [ -z "$DB_PASSWORD" ]; then
            echo "[Entrypoint] ERROR: DB_PASSWORD is required in environment variable mode"
            errors=$((errors + 1))
        fi
    fi

    if [ $errors -gt 0 ]; then
        echo "[Entrypoint] ERROR: $errors validation error(s), exiting"
        exit 1
    fi

    echo "[Entrypoint] Environment validation passed"
}

validate_environment

# 生成 YAML 配置文件（环境变量模式）
generate_config() {
    cat > /app/config-generated.yml << 'YAMLEOF'
server:
  port: "${CONTFUL_SERVER_PORT:-8080}"
  mode: "${CONTFUL_SERVER_MODE:-release}"
  read_timeout: 60
  write_timeout: 60
  shutdown_timeout: 30

database:
  type: "${DB_TYPE:-postgres}"
  host: "${DB_HOST:-postgres}"
  port: "${DB_PORT:-5432}"
  user: "${DB_USER:-postgres}"
  password: "${DB_PASSWORD}"
  name: "${DB_NAME:-contful}"
  ssl_mode: "${DB_SSL_MODE:-disable}"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600

redis:
  host: "${REDIS_HOST:-redis}"
  port: "${REDIS_PORT:-6379}"
  password: "${REDIS_PASSWORD}"
  db: "${REDIS_DB:-0}"
  pool_size: 100

security:
  secret: "${SECRET}"
  algorithm: "${SECRET_ALGORITHM:-aes-256-gcm}"

jwt:
  access_token_expire_minutes: 15
  refresh_token_expire_days: 7

storage:
  driver: "${STORAGE_DRIVER:-local}"
  upload_dir: "${STORAGE_UPLOAD_DIR:-./uploads}"
  max_upload_size_mb: 10
  base_url: "${STORAGE_BASE_URL:-/uploads}"

cors:
  allowed_origins:
    - "*"
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allowed_headers:
    - "*"
  allow_credentials: true
  max_age: 86400

logging:
  level: "${LOG_LEVEL:-info}"
  format: json
  output: stdout

audit:
  enabled: true
  log_all_requests: false

multi_site:
  enabled: true
  default_site_id: ""

features:
  version_history: true
  api_tokens: true
  media_library: true

rate_limit:
  enabled: true
  requests_per_minute: 100
YAMLEOF

    # 替换环境变量占位符
    envsubst < /app/config-generated.yml > /app/config.yml
}

# 检查是否使用环境变量模式
if [ -n "$DB_HOST" ] || [ -n "$DB_PASSWORD" ] || [ -n "$SECRET" ]; then
    echo "[Entrypoint] Environment variable mode detected"
    generate_config
    echo "[Entrypoint] Generated config from environment variables"
elif [ -f "/app/config.yml" ]; then
    echo "[Entrypoint] Using mounted config: /app/config.yml"
else
    echo "[Entrypoint] Using default config: /app/config.yml (built-in)"
fi

echo "[Entrypoint] Config location: /app/config.yml"
echo "[Entrypoint] Database: ${DB_HOST:-postgres}:${DB_PORT:-5432}/${DB_NAME:-contful}"

# 根据服务类型启动（使用 su-exec 以非 root 用户运行 Go 二进制）
start_service() {
    local binary="$1"
    local port="$2"
    local log_file="$3"

    echo "[Entrypoint] Starting $binary on :$port..."
    # 使用 su-exec 以非 root 用户运行（如果可用）
    if command -v su-exec >/dev/null 2>&1 && [ "$(id -u)" = "0" ]; then
        # 已配置 su-exec 且当前以 root 运行：以 contful 用户启动
        # 日志由 entrypoint 以 root 写入，二进制以 contful 用户运行
        su-exec contful /app/$binary > /app/logs/$log_file 2>&1 &
    else
        # 回退：以当前用户运行
        /app/$binary > /app/logs/$log_file 2>&1 &
    fi
}

case "$SERVICE_TYPE" in
    "console")
        # Go 二进制使用非 root 用户运行（安全）
        start_service "admin-api" "$SERVICE_PORT" "admin-api.log"

        # 仅在 console 模式启动 nginx（nginx 需 root 绑定 80 端口，启动后会降权）
        if [ "$MODE" = "console" ]; then
            echo "[Entrypoint] Starting Console SPA on :80..."
            exec /usr/local/openresty/bin/openresty -g "daemon off;"
        else
            echo "[Entrypoint] API mode: nginx skipped"
            wait
        fi
        ;;

    "openapi")
        start_service "openapi-server" "$SERVICE_PORT" "openapi.log"
        wait
        ;;

    *)
        echo "[Entrypoint] ERROR: Unknown service type: $SERVICE_TYPE"
        exit 1
        ;;
esac
