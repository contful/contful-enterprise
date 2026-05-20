#!/bin/bash

# Copyright © 2026-present reepu.com
# SPDX-License-Identifier: Apache-2.0

# =============================================================================
# Contful Docker 镜像构建脚本
# 支持单架构和多架构镜像构建
# =============================================================================
#
# 用法:
#   ./shell/build-image.sh              # 构建所有镜像（当前架构）
#   ./shell/build-image.sh console     # 构建 Console 镜像（当前架构）
#   ./shell/build-image.sh openapi     # 构建 Open API 镜像（当前架构）
#   ./shell/build-image.sh --multi-arch # 构建多架构镜像（amd64 + arm64）
#
# 环境变量:
#   DB_TYPE         数据库类型 (postgresql=PostgreSQL)，支持逗号分隔，默认: postgresql
#   TARGET_ARCH    指定构建架构，多架构模式下可选 (amd64|arm64)，默认: amd64,arm64
#
# 镜像标签规则:
#   多架构构建: 使用 ${db_type}-latest 标签（如 postgresql-latest）
#   单架构构建: 使用 ${db_type}-latest + ${db_type}-${arch}-latest 标签
#              （如 postgresql-latest + postgresql-arm64-latest）
#
# 示例:
#   ./shell/build-image.sh                        # 单架构构建
#   ./shell/build-image.sh --multi-arch           # 多架构构建（amd64 + arm64）
#   ./shell/build-image.sh --multi-arch console  # 多架构构建 Console
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
COMMAND="all"

# 解析参数
for arg in "$@"; do
    case "$arg" in
        --multi-arch) MULTI_ARCH=true ;;
        --single-arch) MULTI_ARCH=false ;;
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
    docker buildx inspect contful-builder &> /dev/null || \
        docker buildx create --name contful-builder --use &> /dev/null
    docker buildx use contful-builder
}

# 构建单架构镜像
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
        -t "contful/enterprise-${image_name}:${tag_arch}" \
        -t "contful/enterprise-${image_name}:${tag_latest}" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "${image_name}:${tag_arch} + :${tag_latest} 构建完成"
}

# 构建多架构镜像
build_multi_arch() {
    local image_name="$1"
    local db_type="$2"
    local dockerfile="$3"

    local tag_latest="${db_type}-latest"

    log_info "构建 ${image_name} 多架构镜像..."
    log_info "  架构: amd64 + arm64"
    log_info "  标签: ${tag_latest}"

    [ ! -f "$dockerfile" ] && { log_error "Dockerfile 不存在: $dockerfile"; exit 1; }

    ensure_buildx_builder

    docker buildx build \
        --no-cache \
        --platform "linux/amd64,linux/arm64" \
        --build-arg DB_TYPE="$db_type" \
        -t "contful/enterprise-${image_name}:${tag_latest}" \
        -f "$dockerfile" \
        --load \
        "$PROJECT_DIR"

    log_success "${image_name}:${tag_latest} (amd64 + arm64) 构建完成"
}

# 构建 Console 镜像
build_console() {
    local console_dockerfile="$PROJECT_DIR/docker/Dockerfile.console"
    IFS=',' read -ra DB_LIST <<< "$DB_TYPES"

    for db_type in "${DB_LIST[@]}"; do
        db_type=$(echo "$db_type" | tr -d '[:space:]')
        echo ""

        if [ "$MULTI_ARCH" = true ]; then
            build_multi_arch "console" "$db_type" "$console_dockerfile"
        else
            local arch=$(get_native_arch)
            [ -z "$arch" ] && { log_error "不支持的架构: $(uname -m)"; exit 1; }
            build_single_arch "console" "$db_type" "$arch" "$console_dockerfile"
        fi
    done
}

# 构建 Open API 镜像
build_openapi() {
    local openapi_dockerfile="$PROJECT_DIR/docker/Dockerfile.openapi"
    IFS=',' read -ra DB_LIST <<< "$DB_TYPES"

    for db_type in "${DB_LIST[@]}"; do
        db_type=$(echo "$db_type" | tr -d '[:space:]')
        echo ""

        if [ "$MULTI_ARCH" = true ]; then
            build_multi_arch "openapi" "$db_type" "$openapi_dockerfile"
        else
            local arch=$(get_native_arch)
            [ -z "$arch" ] && { log_error "不支持的架构: $(uname -m)"; exit 1; }
            build_single_arch "openapi" "$db_type" "$arch" "$openapi_dockerfile"
        fi
    done
}

# 构建所有镜像
build_all() {
    log_info "开始构建 Contful Docker 镜像..."
    log_info "数据库类型: $DB_TYPES"
    log_info "构建模式: $([ "$MULTI_ARCH" = true ] && echo "多架构 (amd64 + arm64)" || echo "单架构 ($(get_native_arch))")"
    echo ""

    build_console
    echo ""
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
  --multi-arch   构建多架构镜像（amd64 + arm64）
  --single-arch  构建单架构镜像（默认）

命令:
  console        构建 Console 镜像
  openapi        构建 Open API 镜像
  all            构建所有镜像（默认）
  list           显示已构建的镜像

环境变量:
  DB_TYPE        数据库类型 (postgresql=PostgreSQL)
                 支持逗号分隔多个类型，默认值: postgresql

示例:
  # 单架构构建（自动检测当前架构）
  \$ $0                        # 构建所有镜像
  \$ $0 console                # 仅构建 Console 镜像

  # 多架构构建（amd64 + arm64）
  \$ $0 --multi-arch           # 构建所有镜像
  \$ $0 --multi-arch console   # 仅构建 Console 多架构镜像

镜像标签规则:
  多架构构建:
    contful/enterprise-console:postgresql-latest   (amd64 + arm64)
    contful/enterprise-openapi:postgresql-latest   (amd64 + arm64)

  单架构构建:
    contful/enterprise-console:postgresql-latest        + postgresql-arm64-latest  (arm64 机器)
    contful/enterprise-console:postgresql-latest        + postgresql-amd64-latest  (amd64 机器)
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
