#!/bin/bash
# =============================================================================
# Contful Docker 镜像构建脚本
# =============================================================================
#
# 用法:
#   ./shell/build-image.sh              # 构建所有镜像
#   ./shell/build-image.sh console      # 构建 Console 镜像（Admin API + 前端）
#   ./shell/build-image.sh openapi      # 构建 Open API 镜像
#
# 构建参数:
#   DB_TYPE=pg  # PostgreSQL（默认）
#   DB_TYPE=dm  # 达梦 DM8
#
# 示例:
#   DB_TYPE=dm ./shell/build-image.sh console  # 构建达梦版本的 Console 镜像
#   ./shell/build-image.sh                     # 构建所有（PostgreSQL 版本）
#
# 镜像说明:
#   contful/console:pg   - Admin API + Console 前端（打包在同一镜像中）
#   contful/console:dm    - Admin API + Console 前端（达梦版）
#   contful/openapi:pg   - Open API 服务（PostgreSQL）
#   contful/openapi:dm   - Open API 服务（达梦）
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

# 数据库类型（默认 PostgreSQL）
DB_TYPE="${DB_TYPE:-pg}"

# 镜像版本
ADMIN_IMAGE="contful/console"
OPENAPI_IMAGE="contful/openapi"

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

# =============================================================================
# 构建函数
# =============================================================================

# 构建 Console 镜像（Admin API + Console 前端，同一镜像）
build_console() {
    log_info "构建 Console 镜像 (DB_TYPE=$DB_TYPE)..."

    # 先构建前端（镜像内含）
    build_frontend
    echo ""

    local tag="${ADMIN_IMAGE}:${DB_TYPE}"
    local dockerfile="$PROJECT_DIR/docker/Dockerfile.console"

    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile 不存在: $dockerfile"
        exit 1
    fi

    docker build \
        --build-arg DB_TYPE="$DB_TYPE" \
        -t "$tag" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "Console 镜像构建完成: $tag"
}

# 构建 Open API 镜像
build_openapi() {
    log_info "构建 Open API 镜像 (DB_TYPE=$DB_TYPE)..."

    local tag="${OPENAPI_IMAGE}:${DB_TYPE}"
    local dockerfile="$PROJECT_DIR/docker/Dockerfile.open"

    if [ ! -f "$dockerfile" ]; then
        log_error "Dockerfile 不存在: $dockerfile"
        exit 1
    fi

    docker build \
        --build-arg DB_TYPE="$DB_TYPE" \
        -t "$tag" \
        -f "$dockerfile" \
        "$PROJECT_DIR"

    log_success "Open API 镜像构建完成: $tag"
}

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

# 构建所有镜像
build_all() {
    log_info "开始构建 Contful Docker 镜像..."
    log_info "数据库类型: $DB_TYPE"
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
    echo "  Admin (Console):"
    docker images "$ADMIN_IMAGE" --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}"
    echo ""
    echo "  Open API:"
    docker images "$OPENAPI_IMAGE" --format "    {{.Repository}}:{{.Tag}} ({{.Size}}) - {{.CreatedSince}}"
    echo ""
    echo "  latest 标签:"
    docker images "$ADMIN_IMAGE" --format "    {{.Repository}}:{{.Tag}}"
    docker images "$OPENAPI_IMAGE" --format "    {{.Repository}}:{{.Tag}}"
    echo ""
}

# =============================================================================
# 辅助函数
# =============================================================================

show_help() {
    echo "Contful Docker 镜像构建脚本"
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
    echo "            默认值: pg"
    echo ""
    echo "示例:"
    echo "  $0                        # 构建所有镜像（PostgreSQL 版本）"
    echo "  $0 console                # 仅构建 Console 镜像"
    echo "  DB_TYPE=dm $0 console     # 构建达梦版本的 Console 镜像"
    echo "  DB_TYPE=dm $0             # 构建所有镜像（达梦版本）"
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
