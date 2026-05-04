#!/bin/bash

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful Docker 镜像构建脚本
# 自动检测当前平台架构进行构建
# =============================================================================
#
# 用法:
#   ./shell/build-image.sh              # 构建所有镜像（当前架构）
#   ./shell/build-image.sh console      # 构建 Console 镜像（当前架构）
#   ./shell/build-image.sh openapi      # 构建 Open API 镜像（当前架构）
#
# 环境变量:
#   DB_TYPE    数据库类型 (pg=PostgreSQL, dm=达梦 DM8)
#              支持逗号分隔，默认值: pg
#
# 示例:
#   ./shell/build-image.sh                       # 构建所有镜像（pg，当前架构）
#   DB_TYPE=pg,dm ./shell/build-image.sh         # 构建 pg + dm（当前架构）
#   DB_TYPE=dm ./shell/build-image.sh console    # 仅构建 dm 的 Console
#
# 镜像标签（每个平台构建时会同时打两个标签）:
#   contful/console:pg-amd64-latest     # 带架构后缀（amd64 平台构建）
#   contful/console:pg-latest           # 无后缀（amd64 平台构建的 native 镜像）
#   contful/console:pg-arm64-latest     # 带架构后缀（arm64 平台构建）
#   contful/console:pg-latest           # 无后缀（arm64 平台构建的 native 镜像）
#   contful/openapi:（同上）
#
# 说明:
#   - 无后缀标签（如 pg-latest）表示该平台上 native 架构的镜像
#   - 带架构后缀标签（如 pg-arm64-latest）可明确区分架构
#   - 不同平台构建的 pg-latest 对应各自平台的架构
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 路径配置
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
CONSOLE_DIR="$PROJECT_DIR/console"

# 数据库类型（默认 PostgreSQL）
DB_TYPES="${DB_TYPE:-pg}"

# 自动检测当前平台架构
get_current_platform() {
    case "$(uname -m)" in
        arm64|aarch64)
            echo "linux/arm64"
            ;;
        x86_64|amd64)
            echo "linux/amd64"
            ;;
        *)
            echo ""
            ;;
    esac
}

get_current_arch() {
    case "$(uname -m)" in
        arm64|aarch64)
            echo "arm64"
            ;;
        x86_64|amd64)
            echo "amd64"
            ;;
        *)
            echo ""
            ;;
    esac
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_docker() {
    if ! docker info &> /dev/null; then
        log_error "Docker 未运行，请先启动 Docker"
        exit 1
    fi
}

# 构建前端
build_frontend() {
    log_info "构建 Console 前端..."

    cd "$CONSOLE_DIR"

    if [ ! -d "node_modules" ]; then
        log_info "安装前端依赖..."
        npm install
    fi

    npm run build

    if [ -d "dist" ]; then
        log_success "前端构建完成: $CONSOLE_DIR/dist"
    else
        log_error "前端构建失败"
        exit 1
    fi
}

# 获取镜像标签（始终返回两个：带架构后缀 + 无后缀）
# 用法: get_tags <db_type> <arch>
# 输出: tag_with_arch tag_without_arch
get_tags() {
    local db_type="$1"
    local arch="$2"
    echo "${db_type}-${arch}-latest ${db_type}-latest"
}

# 构建 Console 镜像
# 用法: build_console_image <db_type> <arch> <platform>
build_console_image() {
    local db_type="$1"
    local arch="$2"
    local platform="$3"
    local tags
    local tag_arch
    local tag_latest

    read -r tag_arch tag_latest <<< "$(get_tags "$db_type" "$arch")"

    log_info "构建 Console 镜像: contful/console:${tag_arch} + contful/console:${tag_latest}"
    log_info "  数据库: ${db_type}, 平台: ${platform}"

    # 先构建前端
    build_frontend
    echo ""

    local dockerfile="$PROJECT_DIR/docker/Dockerfile.console"

    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile 不存在: $dockerfile"
        exit 1
    fi

    # 构建并打两个标签
    docker build \
        --no-cache \
        --platform "$platform" \
        --build-arg DB_TYPE="$db_type" \
        -t "contful/console:${tag_arch}" \
        -t "contful/console:${tag_latest}" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "Console 镜像构建完成:"
    log_success "  contful/console:${tag_arch}"
    log_success "  contful/console:${tag_latest}"
    echo ""
}

# 构建 Open API 镜像
# 用法: build_openapi_image <db_type> <arch> <platform>
build_openapi_image() {
    local db_type="$1"
    local arch="$2"
    local platform="$3"
    local tag_arch
    local tag_latest

    read -r tag_arch tag_latest <<< "$(get_tags "$db_type" "$arch")"

    log_info "构建 Open API 镜像: contful/openapi:${tag_arch} + contful/openapi:${tag_latest}"
    log_info "  数据库: ${db_type}, 平台: ${platform}"

    local dockerfile="$PROJECT_DIR/docker/Dockerfile.openapi"

    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile 不存在: $dockerfile"
        exit 1
    fi

    # 构建并打两个标签
    docker build \
        --no-cache \
        --platform "$platform" \
        --build-arg DB_TYPE="$db_type" \
        -t "contful/openapi:${tag_arch}" \
        -t "contful/openapi:${tag_latest}" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "Open API 镜像构建完成:"
    log_success "  contful/openapi:${tag_arch}"
    log_success "  contful/openapi:${tag_latest}"
    echo ""
}

# 构建所有 Console 镜像
build_console() {
    local arch
    local platform

    arch=$(get_current_arch)
    platform=$(get_current_platform)

    if [ -z "$platform" ]; then
        log_error "不支持的架构: $(uname -m)"
        exit 1
    fi

    IFS=',' read -ra DB_LIST <<< "$DB_TYPES"
    for db_type in "${DB_LIST[@]}"; do
        db_type=$(echo "$db_type" | tr -d '[:space:]')
        build_console_image "$db_type" "$arch" "$platform"
    done
}

# 构建所有 Open API 镜像
build_openapi() {
    local arch
    local platform

    arch=$(get_current_arch)
    platform=$(get_current_platform)

    if [ -z "$platform" ]; then
        log_error "不支持的架构: $(uname -m)"
        exit 1
    fi

    IFS=',' read -ra DB_LIST <<< "$DB_TYPES"
    for db_type in "${DB_LIST[@]}"; do
        db_type=$(echo "$db_type" | tr -d '[:space:]')
        build_openapi_image "$db_type" "$arch" "$platform"
    done
}

# 构建所有镜像
build_all() {
    local arch
    arch=$(get_current_arch)

    log_info "开始构建 Contful Docker 镜像..."
    log_info "数据库类型: $DB_TYPES"
    log_info "架构: $(uname -m) -> $(get_current_platform)"
    echo ""

    log_info "构建 Console 镜像..."
    build_console

    log_info "构建 Open API 镜像..."
    build_openapi

    log_success "所有镜像构建完成！"
    echo ""
    show_images
}

# 显示已构建的镜像
show_images() {
    log_info "当前 Contful 镜像:"
    echo ""
    echo "  Console:"
    docker images contful/console --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}" 2>/dev/null | grep -E "(pg|dm)" || echo "    (无)"
    echo ""
    echo "  Open API:"
    docker images contful/openapi --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}" 2>/dev/null | grep -E "(pg|dm)" || echo "    (无)"
    echo ""
}

show_help() {
    echo "Contful Docker 镜像构建脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  console   构建 Console 镜像（Admin API + Console 前端）"
    echo "  openapi   构建 Open API 镜像"
    echo "  all       构建所有镜像（默认）"
    echo "  list      显示已构建的镜像"
    echo "  help      显示此帮助信息"
    echo ""
    echo "环境变量:"
    echo "  DB_TYPE   数据库类型 (pg=PostgreSQL, dm=达梦 DM8)"
    echo "            支持逗号分隔多个类型，默认值: pg"
    echo ""
    echo "示例:"
    echo "  $0                              # 构建所有镜像（pg，当前架构）"
    echo "  $0 console                      # 仅构建 Console 镜像"
    echo "  DB_TYPE=pg,dm $0                # 构建 pg + dm（当前架构）"
    echo "  DB_TYPE=dm $0 openapi           # 仅构建 dm 的 Open API"
    echo ""
    echo "构建的镜像标签（每类打两个标签）:"
    echo "  contful/console:pg-amd64-latest     # + pg-latest（在 amd64 机器上）"
    echo "  contful/console:pg-arm64-latest     # + pg-latest（在 arm64 机器上）"
    echo "  contful/console:dm-amd64-latest     # + dm-latest（在 amd64 机器上）"
    echo "  contful/console:dm-arm64-latest     # + dm-latest（在 arm64 机器上）"
    echo "  contful/openapi:（同上）"
    echo ""
    echo "说明：无后缀标签（如 pg-latest）对应当前平台的 native 架构"
}

main() {
    cd "$PROJECT_DIR"

    check_docker

    case "${1:-}" in
        console)
            build_console
            ;;
        openapi)
            build_openapi
            ;;
        all|"")
            build_all
            ;;
        list)
            show_images
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
