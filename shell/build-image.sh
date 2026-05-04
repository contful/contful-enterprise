#!/bin/bash

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful Docker 镜像构建脚本（支持多架构）
# =============================================================================
#
# 用法:
#   ./shell/build-image.sh              # 构建所有镜像（所有架构、所有数据库）
#   ./shell/build-image.sh console      # 构建 Console 镜像（所有架构、所有数据库）
#   ./shell/build-image.sh openapi      # 构建 Open API 镜像（所有架构、所有数据库）
#
# 构建参数:
#   DB_TYPE=pg          # PostgreSQL（默认，可逗号分隔 pg,dm）
#   DB_TYPE=dm          # 达梦 DM8
#   ARCH=amd64          # 架构：amd64, arm64, all（默认 all）
#
# 示例:
#   DB_TYPE=pg,dm ./shell/build-image.sh console  # 构建所有架构的 Console 镜像
#   ARCH=arm64 ./shell/build-image.sh              # 仅构建 arm64 架构
#   DB_TYPE=pg ARCH=amd64 ./shell/build-image.sh  # 仅构建 pg + amd64
#
# 镜像标签格式:
#   contful/console:pg-latest          # PostgreSQL + amd64
#   contful/console:pg-arm64-latest    # PostgreSQL + arm64
#   contful/console:dm-latest         # 达梦 DM8 + amd64
#   contful/console:dm-arm64-latest   # 达梦 DM8 + arm64
#   contful/openapi:pg-latest         # （同上）
#   contful/openapi:pg-arm64-latest
#   contful/openapi:dm-latest
#   contful/openapi:dm-arm64-latest
#
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 路径配置
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_DIR="$PROJECT_DIR/docker"
CONSOLE_DIR="$PROJECT_DIR/console"

# 数据库类型（默认 PostgreSQL，支持逗号分隔）
DB_TYPES="${DB_TYPE:-pg}"

# 架构（默认所有架构）
ARCH="${ARCH:-all}"

# 镜像名称
ADMIN_IMAGE="contful/console"
OPENAPI_IMAGE="contful/openapi"

# 支持的平台映射
declare -A PLATFORM_MAP
PLATFORM_MAP[amd64]="linux/amd64"
PLATFORM_MAP[arm64]="linux/arm64"

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

# 检查 Docker 是否运行
check_docker() {
    if ! docker info &> /dev/null; then
        log_error "Docker 未运行，请先启动 Docker"
        exit 1
    fi
}

# 解析数据库类型列表
get_db_types() {
    IFS=',' read -ra DB_LIST <<< "$DB_TYPES"
    echo "${DB_LIST[@]}"
}

# 获取需要构建的架构列表
get_arch_list() {
    case "$ARCH" in
        all)
            echo "amd64 arm64"
            ;;
        amd64|arm64)
            echo "$ARCH"
            ;;
        *)
            log_error "不支持的架构: $ARCH（支持: amd64, arm64, all）"
            exit 1
            ;;
    esac
}

# 获取 Docker 平台参数
get_platform() {
    local arch="$1"
    echo "${PLATFORM_MAP[$arch]}"
}

# 获取镜像标签
get_tag() {
    local db_type="$1"
    local arch="$2"
    
    if [ "$arch" = "amd64" ]; then
        echo "${db_type}-latest"
    else
        echo "${db_type}-${arch}-latest"
    fi
}

# 获取简写标签（无架构后缀）
get_short_tag() {
    local db_type="$1"
    echo "${db_type}-latest"
}

# 构建完成后复制标签（arm64 构建后同时打上简写标签）
tag_also_short() {
    local image="$1"
    local tag="$2"
    local short_tag="$3"
    
    # 仅当标签包含 arm64 时才复制
    if [[ "$tag" == *"arm64"* ]]; then
        log_info "复制标签: $image:$short_tag -> $image:$tag"
        docker tag "$image:$tag" "$image:$short_tag"
        log_success "标签复制完成: $image:$short_tag"
    fi
}

# =============================================================================
# 构建函数
# =============================================================================

# 构建前端（内部函数，供 Console 镜像构建时调用）
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

# 构建单个 Console 镜像
build_console_image() {
    local db_type="$1"
    local arch="$2"
    local tag
    local platform
    
    tag=$(get_tag "$db_type" "$arch")
    platform=$(get_platform "$arch")
    
    log_info "构建 Console 镜像: $ADMIN_IMAGE:$tag (DB_TYPE=$db_type, PLATFORM=$platform)..."
    
    # 先构建前端（镜像内含）
    build_frontend
    echo ""
    
    local dockerfile="$PROJECT_DIR/docker/Dockerfile.console"
    
    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile 不存在: $dockerfile"
        exit 1
    fi
    
    docker build \
        --no-cache \
        --platform "$platform" \
        --build-arg DB_TYPE="$db_type" \
        -t "$ADMIN_IMAGE:$tag" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "Console 镜像构建完成: $ADMIN_IMAGE:$tag"
    
    # 如果是 arm64 构建，同时打上简写标签
    local short_tag
    short_tag=$(get_short_tag "$db_type")
    tag_also_short "$ADMIN_IMAGE" "$tag" "$short_tag"
    
    echo ""
}

# 构建单个 Open API 镜像
build_openapi_image() {
    local db_type="$1"
    local arch="$2"
    local tag
    local platform
    
    tag=$(get_tag "$db_type" "$arch")
    platform=$(get_platform "$arch")
    
    log_info "构建 Open API 镜像: $OPENAPI_IMAGE:$tag (DB_TYPE=$db_type, PLATFORM=$platform)..."
    
    local dockerfile="$PROJECT_DIR/docker/Dockerfile.openapi"
    
    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile 不存在: $dockerfile"
        exit 1
    fi
    
    docker build \
        --no-cache \
        --platform "$platform" \
        --build-arg DB_TYPE="$db_type" \
        -t "$OPENAPI_IMAGE:$tag" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "Open API 镜像构建完成: $OPENAPI_IMAGE:$tag"
    
    # 如果是 arm64 构建，同时打上简写标签
    local short_tag
    short_tag=$(get_short_tag "$db_type")
    tag_also_short "$OPENAPI_IMAGE" "$tag" "$short_tag"
    
    echo ""
}

# 构建所有 Console 镜像（所有数据库类型 + 所有架构）
build_console() {
    local db_types
    local arch_list
    
    db_types=$(get_db_types)
    arch_list=$(get_arch_list)
    
    for db_type in $db_types; do
        for arch in $arch_list; do
            build_console_image "$db_type" "$arch"
        done
    done
}

# 构建所有 Open API 镜像（所有数据库类型 + 所有架构）
build_openapi() {
    local db_types
    local arch_list
    
    db_types=$(get_db_types)
    arch_list=$(get_arch_list)
    
    for db_type in $db_types; do
        for arch in $arch_list; do
            build_openapi_image "$db_type" "$arch"
        done
    done
}

# 构建所有镜像
build_all() {
    log_info "开始构建所有 Contful Docker 镜像..."
    log_info "数据库类型: $DB_TYPES"
    log_info "架构: $ARCH"
    echo ""
    
    # 构建 Console 镜像（包含前端）
    build_console
    echo ""
    
    # 构建 Open API 镜像
    build_openapi
    echo ""
    
    log_success "所有镜像构建完成！"
    echo ""
    show_images
}

# 显示已构建的镜像
show_images() {
    log_info "当前 Contful 镜像:"
    echo ""
    echo "  Console:"
    docker images "$ADMIN_IMAGE" --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}" | grep -E "(pg|dm)" || true
    echo ""
    echo "  Open API:"
    docker images "$OPENAPI_IMAGE" --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}" | grep -E "(pg|dm)" || true
    echo ""
}

# =============================================================================
# 帮助信息
# =============================================================================

show_help() {
    echo "Contful Docker 镜像构建脚本（支持多架构）"
    echo ""
    echo "用法: $0 [命令] [选项]"
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
    echo "  ARCH      目标架构 (amd64, arm64, all)"
    echo "            默认值: all"
    echo ""
    echo "示例:"
    echo "  $0                              # 构建所有镜像（所有架构、pg）"
    echo "  $0 console                      # 仅构建 Console 镜像"
    echo "  DB_TYPE=pg,dm $0                # 构建 pg + dm 所有架构"
    echo "  ARCH=arm64 $0                   # 仅构建 arm64 架构"
    echo "  DB_TYPE=pg ARCH=amd64 $0        # 仅构建 pg + amd64"
    echo "  DB_TYPE=dm ARCH=arm64 $0 console  # 仅构建 dm + arm64 的 Console"
    echo ""
    echo "镜像标签格式:"
    echo "  contful/console:pg-latest           # PostgreSQL + amd64"
    echo "  contful/console:pg-arm64-latest     # PostgreSQL + arm64"
    echo "  contful/console:dm-latest          # 达梦 DM8 + amd64"
    echo "  contful/console:dm-arm64-latest    # 达梦 DM8 + arm64"
    echo "  contful/openapi:pg-latest          # （同上）"
    echo "  contful/openapi:pg-arm64-latest"
    echo "  contful/openapi:dm-latest"
    echo "  contful/openapi:dm-arm64-latest"
    echo ""
    echo "镜像说明:"
    echo "  Console 镜像 = Admin API + Console 前端，打包在同一镜像中"
    echo ""
}

# =============================================================================
# 主逻辑
# =============================================================================

main() {
    cd "$PROJECT_DIR"
    
    # 检查 Docker
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
