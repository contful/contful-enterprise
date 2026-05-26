#!/bin/bash

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful Docker 镜像构建脚本
# 支持单架构（本地）和多架构（推送 registry）构建
# =============================================================================
#
# 用法:
#   ./shell/build-image.sh              # 构建所有镜像（当前架构，本地）
#   ./shell/build-image.sh console     # 构建 Console 镜像（当前架构，本地）
#   ./shell/build-image.sh openapi     # 构建 Open API 镜像（当前架构，本地）
#   ./shell/build-image.sh --multi-arch # 构建多架构镜像并推送到 registry（amd64 + arm64）
#   PROXY=true ./shell/build-image.sh --multi-arch  # 国内网络启用阿里云镜像加速
#
# 环境变量:
#   DB_TYPE         数据库类型 (postgresql=PostgreSQL)，支持逗号分隔，默认: postgresql
#   PROXY           设为 true 使用阿里云镜像加速（国内网络），默认: false
#
# 镜像标签规则:
#   单架构构建: ${db_type}-latest + ${db_type}-${arch}-latest（本地）
#   多架构构建: ${db_type}-latest（推送到 registry，含 amd64 + arm64 清单）
#
# 示例:
#   ./shell/build-image.sh                   # 单架构构建（当前机器）
#   ./shell/build-image.sh --multi-arch      # 多架构推送 registry
#   ./shell/build-image.sh --multi-arch console  # 仅 Console 多架构推送
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

# 默认配置
DB_TYPES="${DB_TYPE:-postgresql}"
MULTI_ARCH=false
PROXY="${PROXY:-false}"
COMMAND="all"

# 解析参数
for arg in "$@"; do
    case "$arg" in
        --multi-arch) MULTI_ARCH=true ;;
        console|openapi|all|list|help|-h|--help) COMMAND="$arg" ;;
    esac
done

# 获取平台字符串
get_platform() {
    local arch="$1"
    case "$arch" in
        arm64) echo "linux/arm64" ;;
        amd64) echo "linux/amd64" ;;
        *) echo "" ;;
    esac
}

get_native_arch() {
    case "$(uname -m)" in
        arm64|aarch64) echo "arm64" ;;
        x86_64|amd64)  echo "amd64" ;;
        *) echo "" ;;
    esac
}

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn()   { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }

check_docker() {
    if ! docker info &> /dev/null; then
        log_error "Docker 未运行，请先启动 Docker"
        exit 1
    fi
}

check_buildx() {
    if ! docker buildx version &> /dev/null; then
        log_error "Docker BuildX 不可用"
        exit 1
    fi
}

ensure_buildx_builder() {
    # 优先使用 Docker Desktop 自带的 desktop-linux（已预配 QEMU + 多平台支持）
    if docker buildx inspect desktop-linux &> /dev/null; then
        docker buildx use desktop-linux
        return
    fi
    # 回退：创建 builder（CI 无 Docker Desktop 环境）
    docker buildx inspect contful-builder &> /dev/null || \
        docker buildx create --name contful-builder --driver docker-container --use &> /dev/null
    docker buildx use contful-builder
}

# 构建单架构镜像（本地 daemon）
build_single_arch() {
    local image_name="$1"
    local db_type="$2"
    local arch="$3"
    local platform=$(get_platform "$arch")
    local dockerfile="$4"

    local tag_arch="${db_type}-${arch}-latest"
    local tag_latest="${db_type}-latest"

    log_info "构建 ${image_name} 镜像 (${arch})..."
    log_info "  标签: ${tag_arch} + ${tag_latest}"

    [ ! -f "$dockerfile" ] && { log_error "Dockerfile 不存在: $dockerfile"; exit 1; }

    docker build \
        --no-cache \
        --platform "$platform" \
        --build-arg DB_TYPE="$db_type" \
        --build-arg PROXY="$PROXY" \
        -t "contful/enterprise-${image_name}:${tag_arch}" \
        -t "contful/enterprise-${image_name}:${tag_latest}" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "${image_name}:${tag_arch} + :${tag_latest} 构建完成"
}

# 构建多架构镜像（推送到 registry）
build_multi_arch() {
    local image_name="$1"
    local db_type="$2"
    local dockerfile="$3"

    local tag_latest="${db_type}-latest"

    log_info "构建 ${image_name} 多架构镜像 → registry"
    log_info "  架构: amd64 + arm64"
    log_info "  标签: contful/enterprise-${image_name}:${tag_latest}"

    [ ! -f "$dockerfile" ] && { log_error "Dockerfile 不存在: $dockerfile"; exit 1; }

    ensure_buildx_builder

    docker buildx build \
        --no-cache \
        --platform "linux/amd64,linux/arm64" \
        --build-arg DB_TYPE="$db_type" \
        --build-arg PROXY="$PROXY" \
        -t "contful/enterprise-${image_name}:${tag_latest}" \
        -f "$dockerfile" \
        --push \
        "$PROJECT_DIR"

    log_success "${image_name}:${tag_latest} (amd64 + arm64) 已推送"
}

# 构建 Console 镜像
build_console() {
    local dockerfile="$PROJECT_DIR/docker/Dockerfile.console"
    IFS=',' read -ra DB_LIST <<< "$DB_TYPES"

    for db_type in "${DB_LIST[@]}"; do
        db_type=$(echo "$db_type" | tr -d '[:space:]')
        echo ""

        if [ "$MULTI_ARCH" = true ]; then
            build_multi_arch "console" "$db_type" "$dockerfile"
        else
            local arch=$(get_native_arch)
            [ -z "$arch" ] && { log_error "不支持的架构: $(uname -m)"; exit 1; }
            build_single_arch "console" "$db_type" "$arch" "$dockerfile"
        fi
    done
}

# 构建 Open API 镜像
build_openapi() {
    local dockerfile="$PROJECT_DIR/docker/Dockerfile.openapi"
    IFS=',' read -ra DB_LIST <<< "$DB_TYPES"

    for db_type in "${DB_LIST[@]}"; do
        db_type=$(echo "$db_type" | tr -d '[:space:]')
        echo ""

        if [ "$MULTI_ARCH" = true ]; then
            build_multi_arch "openapi" "$db_type" "$dockerfile"
        else
            local arch=$(get_native_arch)
            [ -z "$arch" ] && { log_error "不支持的架构: $(uname -m)"; exit 1; }
            build_single_arch "openapi" "$db_type" "$arch" "$dockerfile"
        fi
    done
}

# 构建所有镜像
build_all() {
    local mode
    if [ "$MULTI_ARCH" = true ]; then
        mode="多架构推送 registry (amd64 + arm64)"
    else
        mode="单架构本地 $(get_native_arch)"
    fi

    log_info "开始构建 Contful Docker 镜像..."
    log_info "数据库类型: $DB_TYPES"
    log_info "构建模式: $mode"
    echo ""

    build_console
    echo ""
    build_openapi

    echo ""
    log_success "所有镜像构建完成！"

    if [ "$MULTI_ARCH" != true ]; then
        echo ""
        show_images
    fi
}

# 显示已构建的镜像
show_images() {
    log_info "本地 Contful 镜像:"
    echo ""
    echo "  Console:"
    docker images contful/enterprise-console --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}" 2>/dev/null | grep -E "${DB_TYPES}" | head -10 || echo "    (无)"
    echo ""
    echo "  Open API:"
    docker images contful/enterprise-openapi --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}" 2>/dev/null | grep -E "${DB_TYPES}" | head -10 || echo "    (无)"
    echo ""
}

show_help() {
    cat << EOF
Contful Docker 镜像构建脚本

用法: $0 [选项] [命令]

选项:
  --multi-arch   构建多架构镜像（amd64 + arm64）并推送到 registry

命令:
  console        构建 Console 镜像
  openapi        构建 Open API 镜像
  all            构建所有镜像（默认）
  list           显示本地已构建的镜像

环境变量:
  DB_TYPE        数据库类型，默认: postgresql
  PROXY          设为 true 使用阿里云镜像加速（国内网络），默认: false

示例:
  # 单架构构建（自动检测当前架构，本地 daemon）——最快
  \$ $0                        # 构建所有镜像
  \$ $0 console                # 仅构建 Console 镜像

  # 多架构构建（amd64 + arm64）→ registry
  \$ $0 --multi-arch           # 构建所有镜像并推送

  # 国内网络启用镜像加速
  PROXY=true \$ $0             # 单架构 + 阿里云源
  PROXY=true \$ $0 --multi-arch  # 多架构 + 阿里云源
  \$ $0 --multi-arch console   # 仅构建 Console 多架构并推送

镜像标签:
  单架构: contful/enterprise-console:postgresql-latest + :postgresql-amd64-latest
  多架构: contful/enterprise-console:postgresql-latest（registry 上的 multi-arch 清单）
EOF
}

main() {
    cd "$PROJECT_DIR"
    check_docker

    if [ "$MULTI_ARCH" = true ]; then
        check_buildx
    fi

    case "$COMMAND" in
        console)  build_console ;;
        openapi)  build_openapi ;;
        all|"")   build_all ;;
        list)     show_images ;;
        help|-h|--help) show_help ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
