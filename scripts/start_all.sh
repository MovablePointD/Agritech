#!/usr/bin/env bash
# =============================================
#  rxtcloud 全部服务启动脚本 (Linux/macOS)
#  后台启动所有微服务
# =============================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$ROOT_DIR/logs"

mkdir -p "$LOG_DIR"

# 颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  rxtcloud 微服务启动器${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

declare -A SERVICES
SERVICES=(
    ["common-service"]="$ROOT_DIR/common:8080"
    ["knowledge-service"]="$ROOT_DIR/knowledge-service:8081"
    ["message-service"]="$ROOT_DIR/message-service:8082"
    ["post-service"]="$ROOT_DIR/post-service:8083"
    ["product-service"]="$ROOT_DIR/product-service:8084"
)

i=1
for svc in "${!SERVICES[@]}"; do
    IFS=':' read -r dir port <<< "${SERVICES[$svc]}"
    printf "  [%d] %-22s :%s\n" "$i" "$svc" "$port"
    ((i++))
done

echo ""
echo -e "${YELLOW}正在启动所有服务...${NC}"
echo ""

i=1
for svc in "${!SERVICES[@]}"; do
    IFS=':' read -r dir port <<< "${SERVICES[$svc]}"
    LOGFILE="$LOG_DIR/${svc}.log"
    
    echo -e "  -> ${CYAN}启动 ${svc}...${NC}"
    cd "$dir" || exit 1
    
    # 检查和清理旧端口占用
    if command -v lsof &> /dev/null; then
        fuser -k "${port}/tcp" 2>/dev/null
    fi
    
    nohup go run main.go > "$LOGFILE" 2>&1 &
    echo "    PID: $! 日志: $LOGFILE"
    cd "$ROOT_DIR"
    
    sleep 2
    ((i++))
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  所有服务已启动!${NC}"
echo -e "${GREEN}  日志目录: $LOG_DIR${NC}"
echo -e "${GREEN}========================================${NC}"
