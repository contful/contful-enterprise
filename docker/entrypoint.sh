#!/bin/sh

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful Entrypoint - 环境变量覆盖配置文件方案
#
# 设计理念:
#   1. 镜像内置 config.yml，包含合理的默认值
#   2. Go 程序使用 viper 加载配置
#   3. 程序自动读取环境变量覆盖配置文件中的值
#   4. 未设置的环境变量使用 config.yml 中的默认值
#
# 程序配置加载顺序（viper 优先级）:
#   1. 环境变量（最高）
#   2. config.yml
#   3. 内置默认值（最低）
#
# 支持服务: console (admin-api) / openapi (openapi-server)
# =============================================================================

set -e

# 确保 /app/logs 和 /app/uploads 有写权限
mkdir -p /app/logs /app/uploads

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

# 显示配置信息（隐藏敏感值）
echo "[Entrypoint] Config: /app/config.yml"
echo "[Entrypoint] Database: ${DB_HOST:-postgres}:${DB_PORT:-5432}/${DB_NAME:-contful}"
echo "[Entrypoint] Redis: ${REDIS_HOST:-redis}:${REDIS_PORT:-6379}/${REDIS_DB:-0}"
if [ -n "$SECRET" ]; then
    echo "[Entrypoint] SECRET: [已设置] (长度: ${#SECRET})"
else
    echo "[Entrypoint] SECRET: [使用配置默认值]"
fi

# =========================================================================
# 启动服务
# =========================================================================
start_service() {
    local binary="$1"
    local port="$2"
    local log_file="$3"

    echo "[Entrypoint] Starting $binary on :$port..."
    if command -v su-exec >/dev/null 2>&1 && [ "$(id -u)" = "0" ]; then
        su-exec contful /app/$binary > /app/logs/$log_file 2>&1 &
    else
        /app/$binary > /app/logs/$log_file 2>&1 &
    fi
}

# 等待端口就绪（最多 30s）
wait_for_port() {
    local port="$1"
    local max_wait=30
    local waited=0
    echo -n "[Entrypoint] Waiting for :$port "
    while [ $waited -lt $max_wait ]; do
        if wget -q -O- http://127.0.0.1:$port/health >/dev/null 2>&1; then
            echo " OK ($waited s)"
            return 0
        fi
        sleep 1
        waited=$((waited + 1))
        echo -n "."
    done
    echo " TIMEOUT"
    return 1
}

case "$SERVICE_TYPE" in
    "console")
        start_service "admin-api" "$SERVICE_PORT" "admin-api.log"

        if [ "$MODE" = "console" ]; then
            if ! wait_for_port "$SERVICE_PORT"; then
                echo "[Entrypoint] ERROR: admin-api failed to start. Check /app/logs/admin-api.log"
                exit 1
            fi
            echo "[Entrypoint] Starting Console SPA on :80..."
            exec /usr/local/openresty/bin/openresty -g "daemon off;"
        else
            echo "[Entrypoint] API mode: nginx skipped"
            wait
        fi
        ;;

    "openapi")
        start_service "openapi-server" "$SERVICE_PORT" "openapi.log"
        if ! wait_for_port "$SERVICE_PORT"; then
            echo "[Entrypoint] ERROR: openapi-server failed to start. Check /app/logs/openapi.log"
            exit 1
        fi
        echo "[Entrypoint] Open API ready on :$SERVICE_PORT"
        wait
        ;;

    *)
        echo "[Entrypoint] ERROR: Unknown service type: $SERVICE_TYPE"
        exit 1
        ;;
esac
