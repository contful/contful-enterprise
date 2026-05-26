#!/bin/sh

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful Entrypoint — 自动数据库初始化 + 密钥生成 + 服务启动
#
# 流程:
#   1. wait_for_db    — 轮询 /app/admin check-db，最多 10×3s
#   2. auto_init      — 空库 + AUTO_INIT=true 时导入 SQL
#   3. keygen         — 缺失密钥对时自动生成
#   4. start_service  — 启动 Go 服务 + nginx
# =============================================================================

set -e

# ─── 全局变量 ────────────────────────────────────────────────────────────────

BINARY="/app/admin"
PG_HOST="${DB_HOST:-postgres}"
PG_PORT="${DB_PORT:-5432}"
PG_USER="${DB_USER:-contful}"
PG_NAME="${DB_NAME:-contful}"
PG_PASS="${DB_PASSWORD:-contful}"

export PGPASSFILE="/tmp/.pgpass"
PSQL_CONN="-h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_NAME"

# ─── Step 1: 等待数据库就绪 ──────────────────────────────────────────────────

wait_for_db() {
    local max_retries=10
    local retry=0

    while [ $retry -lt $max_retries ]; do
        rc=0
        "$BINARY" check-db 2>/dev/null || rc=$?
        case $rc in
            0) return 0 ;;                     # 有表，数据库就绪
            1) return 1 ;;                     # 可达但无表（空库）
        esac
        retry=$((retry + 1))
        if [ $retry -lt $max_retries ]; then
            sleep 3
        fi
    done

    echo "[Entrypoint] ERROR: Cannot connect to database after ${max_retries} attempts" >&2
    exit 1
}

# ─── Step 2: 自动初始化 ──────────────────────────────────────────────────────

setup_pgpass() {
    echo "$PG_HOST:$PG_PORT:$PG_NAME:$PG_USER:$PG_PASS" > "$PGPASSFILE"
    chmod 600 "$PGPASSFILE"
}

cleanup_pgpass() {
    rm -f "$PGPASSFILE"
}

do_import() {
    echo "[Init] Acquiring advisory lock..."
    setup_pgpass
    trap cleanup_pgpass EXIT

    psql $PSQL_CONN -c "SELECT pg_advisory_lock(12345)" >/dev/null

    echo "[Init] Importing /app/db/init_pg.sql (DDL + seed data)..."
    psql $PSQL_CONN -v ON_ERROR_STOP=1 -f /app/db/init_pg.sql

    psql $PSQL_CONN -c "SELECT pg_advisory_unlock(12345)" >/dev/null
    echo "[Init] Database initialized successfully."
}

# ─── Step 3: 密钥检查 ────────────────────────────────────────────────────────

ensure_keys() {
    if [ ! -f "/app/conf/public.pem" ] || [ ! -f "/app/conf/private.pem" ]; then
        echo "[Entrypoint] Generating key pair..."
        "$BINARY" gen-key
    fi
}

# ─── Step 4: 服务启动（保留现有逻辑）─────────────────────────────────────────

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

# ─── 主流程 ──────────────────────────────────────────────────────────────────

main() {
    # 确保运行时目录存在
    mkdir -p /app/logs /app/uploads /app/conf

    # 检测服务类型
    if [ -f "/app/admin" ]; then
        SERVICE_TYPE="console"
        SERVICE_PORT="${CONTFUL_SERVER_PORT:-9080}"
    elif [ -f "/app/openapi" ]; then
        SERVICE_TYPE="openapi"
        SERVICE_PORT="${CONTFUL_SERVER_PORT:-8080}"
    else
        echo "[Entrypoint] ERROR: No service binary found" >&2
        exit 1
    fi

    MODE="${CONTFUL_MODE:-console}"

    # Step 1: 等待数据库
    wait_for_db
    result=$?

    # Step 2: 初始化检查
    AUTO_INIT="${CONTFUL_AUTO_INIT:-true}"
    if [ $result -eq 1 ]; then
        # 空库
        if [ "$AUTO_INIT" = "true" ]; then
            do_import
        else
            echo "[Entrypoint] ERROR: No tables found and CONTFUL_AUTO_INIT=false" >&2
            exit 1
        fi
    fi
    # result=0: 有表，静默跳过

    # Step 3: 密钥检查
    ensure_keys

    # Step 4: 启动服务
    case "$SERVICE_TYPE" in
        "console")
            start_service "admin" "$SERVICE_PORT" "admin.log"

            if [ "$MODE" = "console" ]; then
                if ! wait_for_port "$SERVICE_PORT"; then
                    echo "[Entrypoint] ERROR: admin failed to start. Check /app/logs/admin.log" >&2
                    exit 1
                fi
                echo "[Entrypoint] Starting Console SPA on :80..."
                echo ""
                echo "  Contful is ready: http://localhost (admin / admin123)"
                echo ""
                exec /usr/local/openresty/bin/openresty -g "daemon off;"
            else
                echo "[Entrypoint] API mode: nginx skipped"
                if wait_for_port "$SERVICE_PORT"; then
                    echo ""
                    echo "  Contful API is ready: http://localhost:$SERVICE_PORT"
                    echo ""
                fi
                wait
            fi
            ;;
        "openapi")
            start_service "openapi" "$SERVICE_PORT" "openapi.log"
            if ! wait_for_port "$SERVICE_PORT"; then
                echo "[Entrypoint] ERROR: openapi failed to start. Check /app/logs/openapi.log" >&2
                exit 1
            fi
            echo "[Entrypoint] Open API ready on :$SERVICE_PORT"
            wait
            ;;
    esac
}

main
