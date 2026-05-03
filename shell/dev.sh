#!/bin/bash

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful 本地开发环境启动脚本
# =============================================================================
#
# 用法:
#   ./shell/dev.sh start        # 启动全部服务（使用远程数据库）
#   ./shell/dev.sh start local   # 启动全部服务（使用 Docker 本地数据库）
#   ./shell/dev.sh stop         # 停止所有服务
#   ./shell/dev.sh status       # 查看服务状态
#   ./shell/dev.sh logs [svc]   # 查看日志（可选：admin/openapi/console）
#   ./shell/dev.sh restart      # 重启服务
#
# 环境变量:
#   MODE=local      # 使用 Docker 本地数据库（默认使用远程数据库）
#   DB_HOST         # PostgreSQL 主机（默认 localhost 或 139.198.171.102）
#   DB_PASSWORD     # PostgreSQL 密码
#   REDIS_PASSWORD  # Redis 密码
#
# 示例:
#   # 使用远程数据库（默认）
#   ./shell/dev.sh start
#
#   # 使用 Docker 本地数据库
#   MODE=local ./shell/dev.sh start
#
#   # 查看 Admin API 日志
#   ./shell/dev.sh logs admin
#
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 路径配置
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ADMIN_DIR="$PROJECT_DIR/admin"
OPENAPI_DIR="$PROJECT_DIR/openapi"
CONSOLE_DIR="$PROJECT_DIR/console"
DOCKER_DIR="$PROJECT_DIR/docker"
BUILD_DIR="$PROJECT_DIR/build"
LOG_DIR="$PROJECT_DIR/logs"
UPLOAD_DIR="$PROJECT_DIR/uploads"

# 默认配置
MODE="${MODE:-remote}"  # remote 或 local
ADMIN_PORT="${ADMIN_PORT:-9080}"
OPENAPI_PORT="${OPENAPI_PORT:-8080}"
CONSOLE_PORT="${CONSOLE_PORT:-3000}"

# =============================================================================
# 辅助函数
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${CYAN}==== $1 ====${NC}"
}

# 创建必要的目录
ensure_dirs() {
    mkdir -p "$BUILD_DIR" "$LOG_DIR" "$UPLOAD_DIR"
}

# 加载 .env 文件（如果存在）
load_env() {
    if [ -f "$PROJECT_DIR/.env" ]; then
        log_info "加载 .env 配置..."
        set -a
        source "$PROJECT_DIR/.env"
        set +a
    fi
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# =============================================================================
# Docker 本地数据库
# =============================================================================

start_local_db() {
    log_section "启动 Docker 本地数据库"

    # 检查 Docker
    if ! docker info &> /dev/null; then
        log_error "Docker 未运行，请先启动 Docker"
        exit 1
    fi

    # 启动 PostgreSQL
    log_info "启动 PostgreSQL 18..."
    docker run -d \
        --name contful-postgres \
        -e POSTGRES_USER=postgres \
        -e POSTGRES_PASSWORD="${DB_PASSWORD:-contful-secret}" \
        -e POSTGRES_DB=contful \
        -p 5432:5432 \
        -v postgres_data:/var/lib/postgresql/data \
        postgres:18-alpine

    # 等待 PostgreSQL 就绪
    log_info "等待 PostgreSQL 启动..."
    for i in {1..30}; do
        if docker exec contful-postgres pg_isready -U postgres &>/dev/null; then
            log_success "PostgreSQL 已就绪"
            break
        fi
        sleep 1
    done

    # 启动 Redis
    log_info "启动 Redis 7..."
    docker run -d \
        --name contful-redis \
        -p 6379:6379 \
        -v redis_data:/data \
        redis:7-alpine redis-server --appendonly yes

    # 等待 Redis 就绪
    log_info "等待 Redis 启动..."
    for i in {1..10}; do
        if docker exec contful-redis redis-cli ping &>/dev/null; then
            log_success "Redis 已就绪"
            break
        fi
        sleep 1
    done

    log_success "本地数据库启动完成"
}

stop_local_db() {
    log_info "停止 Docker 本地数据库..."
    docker stop contful-postgres contful-redis 2>/dev/null || true
    docker rm contful-postgres contful-redis 2>/dev/null || true
    docker volume rm postgres_data redis_data 2>/dev/null || true
    log_success "本地数据库已停止"
}

# =============================================================================
# Admin API 服务
# =============================================================================

start_admin() {
    log_section "启动 Admin API"

    if check_port $ADMIN_PORT; then
        log_warn "Admin API 端口 $ADMIN_PORT 已被占用，跳过启动"
        return
    fi

    # 编译到 build 目录
    log_info "编译 Admin API..."
    cd "$ADMIN_DIR"
    go build -tags=pg -ldflags="-s -w" -o "$BUILD_DIR/admin-server" .

    # 设置环境变量
    export CONTFUL_SERVER_PORT="$ADMIN_PORT"
    export CONTFUL_DB_HOST="${DB_HOST:-localhost}"
    export CONTFUL_DB_PORT="${DB_PORT:-5432}"
    export CONTFUL_DB_USER="${DB_USER:-postgres}"
    export CONTFUL_DB_PASSWORD="${DB_PASSWORD}"
    export CONTFUL_DB_NAME="${DB_NAME:-contful}"
    export CONTFUL_REDIS_HOST="${REDIS_HOST:-localhost}"
    export CONTFUL_REDIS_PORT="${REDIS_PORT:-6379}"
    export CONTFUL_REDIS_PASSWORD="${REDIS_PASSWORD}"
    export CONTFUL_SECRET="${SECRET:-dev-secret-change-in-production}"
    export CONTFUL_LOG_LEVEL="${LOG_LEVEL:-info}"
    export CONTFUL_STORAGE_UPLOAD_DIR="$UPLOAD_DIR"

    # 后台运行
    nohup "$BUILD_DIR/admin-server" > "$LOG_DIR/admin.log" 2>&1 &
    echo $! > "$LOG_DIR/admin.pid"

    # 等待启动
    sleep 2

    if check_port $ADMIN_PORT; then
        log_success "Admin API 已启动 (http://localhost:$ADMIN_PORT)"
    else
        log_error "Admin API 启动失败，查看日志: tail -f $LOG_DIR/admin.log"
    fi
}

stop_admin() {
    if [ -f "$LOG_DIR/admin.pid" ]; then
        PID=$(cat "$LOG_DIR/admin.pid")
        if kill -0 "$PID" 2>/dev/null; then
            log_info "停止 Admin API (PID: $PID)..."
            kill "$PID"
            rm -f "$LOG_DIR/admin.pid"
        fi
    fi
}

# =============================================================================
# Open API 服务
# =============================================================================

start_openapi() {
    log_section "启动 Open API"

    if check_port $OPENAPI_PORT; then
        log_warn "Open API 端口 $OPENAPI_PORT 已被占用，跳过启动"
        return
    fi

    # 编译到 build 目录
    log_info "编译 Open API..."
    cd "$OPENAPI_DIR"
    go build -tags=pg -ldflags="-s -w" -o "$BUILD_DIR/openapi-server" .

    # 后台运行（直接传入环境变量，避免被 .env 的 SERVER_PORT 污染）
    nohup env \
        SERVER_PORT="$OPENAPI_PORT" \
        DB_HOST="${DB_HOST:-localhost}" \
        DB_PORT="${DB_PORT:-5432}" \
        DB_USER="${DB_USER:-postgres}" \
        DB_PASSWORD="${DB_PASSWORD}" \
        DB_NAME="${DB_NAME:-contful}" \
        REDIS_HOST="${REDIS_HOST:-localhost}" \
        REDIS_PORT="${REDIS_PORT:-6379}" \
        REDIS_PASSWORD="${REDIS_PASSWORD}" \
        STORAGE_UPLOAD_DIR="$UPLOAD_DIR" \
        "$BUILD_DIR/openapi-server" > "$LOG_DIR/openapi.log" 2>&1 &
    echo $! > "$LOG_DIR/openapi.pid"

    # 等待启动
    sleep 2

    if check_port $OPENAPI_PORT; then
        log_success "Open API 已启动 (http://localhost:$OPENAPI_PORT)"
    else
        log_error "Open API 启动失败，查看日志: tail -f $LOG_DIR/openapi.log"
    fi
}

stop_openapi() {
    if [ -f "$LOG_DIR/openapi.pid" ]; then
        PID=$(cat "$LOG_DIR/openapi.pid")
        if kill -0 "$PID" 2>/dev/null; then
            log_info "停止 Open API (PID: $PID)..."
            kill "$PID"
            rm -f "$LOG_DIR/openapi.pid"
        fi
    fi
}

# =============================================================================
# Console 前端
# =============================================================================

start_console() {
    log_section "启动 Console 前端"

    if check_port $CONSOLE_PORT; then
        log_warn "Console 端口 $CONSOLE_PORT 已被占用，跳过启动"
        return
    fi

    cd "$CONSOLE_DIR"

    # 检查依赖
    if [ ! -d "node_modules" ]; then
        log_info "安装前端依赖..."
        npm install
    fi

    # 设置 API 地址
    export VITE_API_BASE_URL="http://localhost:$ADMIN_PORT/admin/api/v1"

    # 后台运行
    log_info "启动 Vite 开发服务器..."
    nohup npm run dev > "$LOG_DIR/console.log" 2>&1 &
    echo $! > "$LOG_DIR/console.pid"

    # 等待启动
    sleep 5

    if check_port $CONSOLE_PORT; then
        log_success "Console 已启动 (http://localhost:$CONSOLE_PORT)"
    else
        log_error "Console 启动失败，查看日志: tail -f $LOG_DIR/console.log"
    fi
}

stop_console() {
    if [ -f "$LOG_DIR/console.pid" ]; then
        PID=$(cat "$LOG_DIR/console.pid")
        if kill -0 "$PID" 2>/dev/null; then
            log_info "停止 Console (PID: $PID)..."
            kill "$PID"
            rm -f "$LOG_DIR/console.pid"
        fi
    fi
}

# =============================================================================
# 服务状态
# =============================================================================

show_status() {
    log_section "服务状态"

    echo ""
    echo -e "  ${CYAN}Admin API${NC}   (9080): $(check_port 9080 && echo -e "${GREEN}运行中${NC}" || echo -e "${RED}已停止${NC}")"
    echo -e "  ${CYAN}Open API${NC}    (8080): $(check_port 8080 && echo -e "${GREEN}运行中${NC}" || echo -e "${RED}已停止${NC}")"
    echo -e "  ${CYAN}Console${NC}     (3000): $(check_port 3000 && echo -e "${GREEN}运行中${NC}" || echo -e "${RED}已停止${NC}")"
    echo ""

    # Docker 状态
    if docker info &>/dev/null; then
        echo -e "  ${CYAN}PostgreSQL${NC}  (5432): $(docker ps --filter name=contful-postgres --format '{{.Names}}' 2>/dev/null | grep -q contful-postgres && echo -e "${GREEN}运行中${NC}" || echo -e "${GRAY}未启动${NC}")"
        echo -e "  ${CYAN}Redis${NC}       (6379): $(docker ps --filter name=contful-redis --format '{{.Names}}' 2>/dev/null | grep -q contful-redis && echo -e "${GREEN}运行中${NC}" || echo -e "${GRAY}未启动${NC}")"
    fi
    echo ""
}

show_logs() {
    local service="${1:-}"

    case "$service" in
        admin)
            tail -n 50 -f "$LOG_DIR/admin.log"
            ;;
        openapi|api)
            tail -n 50 -f "$LOG_DIR/openapi.log"
            ;;
        console|frontend)
            tail -n 50 -f "$LOG_DIR/console.log"
            ;;
        "")
            echo "用法: $0 logs [admin|openapi|console]"
            ;;
        *)
            log_error "未知服务: $service"
            ;;
    esac
}

# =============================================================================
# 主逻辑
# =============================================================================

stop_all() {
    log_section "停止所有服务"

    stop_admin
    stop_openapi
    stop_console

    if [ "$MODE" = "local" ]; then
        stop_local_db
    fi

    log_success "所有服务已停止"
}

start_all() {
    ensure_dirs
    load_env

    if [ "$MODE" = "local" ]; then
        start_local_db
        # 本地模式使用 localhost
        export DB_HOST="${DB_HOST:-localhost}"
        export REDIS_HOST="${REDIS_HOST:-localhost}"
    else
        # 远程模式，使用配置的远程地址
        export DB_HOST="${DB_HOST:-139.198.171.102}"
        export REDIS_HOST="${REDIS_HOST:-139.198.171.102}"
    fi

    log_section "启动 Contful 开发服务"
    log_info "模式: $MODE"
    log_info "数据库: $DB_HOST:$DB_PORT"
    log_info "Redis: $REDIS_HOST:$REDIS_PORT"
    echo ""

    start_admin
    start_openapi
    start_console

    echo ""
    log_success "所有服务启动完成！"
    echo ""
    show_status
}

restart_all() {
    stop_all
    echo ""
    sleep 2
    start_all
}

show_help() {
    cat << EOF
Contful 本地开发环境启动脚本

用法: $0 [命令] [选项]

命令:
  start [local]  启动全部服务（可选：local 使用 Docker 本地数据库）
  stop           停止所有服务
  restart        重启所有服务
  status         查看服务状态
  logs [服务]    查看日志（admin|openapi|console）
  help           显示此帮助信息

环境变量:
  MODE=local     使用 Docker 本地数据库（默认使用远程数据库）
  DB_HOST        PostgreSQL 主机（默认 localhost 或 139.198.171.102）
  DB_PASSWORD    PostgreSQL 密码
  REDIS_PASSWORD Redis 密码
  SECRET         JWT 签名密钥

示例:
  # 使用远程数据库（默认）
  ./shell/dev.sh start

  # 使用 Docker 本地数据库
  MODE=local ./shell/dev.sh start

  # 查看 Admin API 日志
  ./shell/dev.sh logs admin

EOF
}

main() {
    cd "$PROJECT_DIR"

    case "${1:-}" in
        start)
            if [ "${2:-}" = "local" ]; then
                MODE="local"
            fi
            start_all
            ;;
        stop)
            stop_all
            ;;
        restart)
            restart_all
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
