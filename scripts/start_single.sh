#!/usr/bin/env bash
# =============================================
#  rxtcloud 单个服务启动 (Linux/macOS)
#  用法: ./start_single.sh <服务名>
#  例如: ./start_single.sh common
# =============================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

declare -A SERVICES
SERVICES=(
    ["common"]="公共服务|common|8080"
    ["knowledge"]="知识服务|knowledge-service|8081"
    ["message"]="消息服务|message-service|8082"
    ["post"]="动态服务|post-service|8083"
    ["product"]="商品服务|product-service|8084"
)

show_menu() {
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  rxtcloud 微服务快捷启动${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "  可用服务:"
    for key in "${!SERVICES[@]}"; do
        IFS='|' read -r desc dir port <<< "${SERVICES[$key]}"
        printf "    %-14s %-22s :%s  %s\n" "$key" "$dir" "$port" "$desc"
    done
    echo ""
}

SERVICE_NAME="$1"

if [ -z "$SERVICE_NAME" ]; then
    show_menu
    read -r -p "请输入服务名 (common/knowledge/message/post/product): " SERVICE_NAME
    echo ""
fi

if [ -z "${SERVICES[$SERVICE_NAME]}" ]; then
    echo -e "${RED}错误: 未知服务 '$SERVICE_NAME'${NC}"
    echo -e "${YELLOW}可用: common, knowledge, message, post, product${NC}"
    exit 1
fi

IFS='|' read -r desc dir port <<< "${SERVICES[$SERVICE_NAME]}"
TARGET_DIR="$ROOT_DIR/$dir"

if [ ! -d "$TARGET_DIR" ]; then
    echo -e "${RED}错误: 目录不存在 '$TARGET_DIR'${NC}"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  启动 $dir${NC}"
echo -e "${GREEN}  端口: $port  描述: $desc${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 清理旧端口
if command -v fuser &> /dev/null; then
    fuser -k "${port}/tcp" 2>/dev/null
fi

cd "$TARGET_DIR" || exit 1
go run main.go
