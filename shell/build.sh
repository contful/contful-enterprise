#!/bin/bash
# =============================================================================
# Contful 本地构建脚本
# =============================================================================
#
# 用法:
#   ./shell/build.sh              # 构建所有（前端 + Console + Open API）
#   ./shell/build.sh console      # 构建 Console 镜像（Admin API + 前端）
#   ./shell/build.sh openapi      # 构建 Open API 镜像
#
# 构建参数:
#   DB_TYPE=pg  # PostgreSQL（默认）
#   DB_TYPE=dm  # 达梦 DM8
#
# 示例:
#   ./shell/build.sh                        # 构建所有（PostgreSQL 版本）
#   DB_TYPE=dm ./shell/build.sh console     # 构建达梦版本的 Console
#
# 构建产物:
#   build/admin-server    - Admin API 二进制
#   build/openapi-server  - Open API 二进制
#   console/dist/         - 前端静态文件
#
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
ADMIN_DIR="$PROJECT_DIR/admin"
OPENAPI_DIR="$PROJECT_DIR/openapi"
CONSOLE_DIR="$PROJECT_DIR/console"
BUILD_DIR="$PROJECT_DIR/build"

# 数据库类型（默认 PostgreSQL）
DB_TYPE="${DB_TYPE:-pg}"

# =============================================================================
# 辅助函数
# =============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装"
        exit 1
    fi
}

# =============================================================================
# 构建函数
# =============================================================================

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

# 构建 Console（Admin API + 前端）
build_console() {
    log_info "构建 Console (DB_TYPE=$DB_TYPE)..."

    # 先构建前端
    build_frontend
    echo ""

    # 创建构建目录
    mkdir -p "$BUILD_DIR"

    # 构建 Admin API
    cd "$ADMIN_DIR"
    go build -tags="${DB_TYPE}" -ldflags="-s -w" -o "$BUILD_DIR/admin-server" .

    if [ -f "$BUILD_DIR/admin-server" ]; then
        log_success "Console 构建完成"
    else
        log_error "Console 构建失败"
        exit 1
    fi
}

# 构建 Open API
build_openapi() {
    log_info "构建 Open API (DB_TYPE=$DB_TYPE)..."

    mkdir -p "$BUILD_DIR"

    cd "$OPENAPI_DIR"
    go build -tags="${DB_TYPE}" -ldflags="-s -w" -o "$BUILD_DIR/openapi-server" .

    if [ -f "$BUILD_DIR/openapi-server" ]; then
        log_success "Open API 构建完成: $BUILD_DIR/openapi-server"
    else
        log_error "Open API 构建失败"
        exit 1
    fi
}

# 构建所有
build_all() {
    log_info "开始构建 Contful..."
    log_info "数据库类型: $DB_TYPE"
    echo ""

    build_console
    echo ""

    build_openapi
    echo ""

    log_success "所有构建完成！"
    echo ""
    echo "产物位置:"
    echo "  - Console 前端: $CONSOLE_DIR/dist"
    echo "  - Admin API:     $BUILD_DIR/admin-server"
    echo "  - Open API:      $BUILD_DIR/openapi-server"
}

# =============================================================================
# 主逻辑
# =============================================================================

show_help() {
    echo "Contful 本地构建脚本"
    echo ""
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  console   构建 Console（Admin API + 前端）"
    echo "  openapi   构建 Open API"
    echo "  all       构建所有（默认）"
    echo ""
    echo "环境变量:"
    echo "  DB_TYPE   数据库类型 (pg=PostgreSQL, dm=达梦 DM8)"
    echo "            默认值: pg"
    echo ""
    echo "示例:"
    echo "  $0                        # 构建所有（PostgreSQL 版本）"
    echo "  $0 console                # 仅构建 Console"
    echo "  DB_TYPE=dm $0 openapi     # 构建达梦版本的 Open API"
    echo ""
}

main() {
    cd "$PROJECT_DIR"

    # 创建构建目录
    mkdir -p "$BUILD_DIR"

    case "${1:-}" in
        console)
            check_command go
            build_console
            ;;
        openapi)
            check_command go
            build_openapi
            ;;
        all|"")
            check_command go
            check_command node
            check_command npm
            build_all
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
