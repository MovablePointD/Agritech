#!/usr/bin/env bash
# =============================================
#  rxtcloud 全部服务停止脚本 (Linux/macOS)
#  按端口号终止所有微服务进程
# =============================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  正在停止所有服务...${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 服务端口列表
PORTS=(
    "8080:common-service"
    "8081:knowledge-service"
    "8082:message-service"
    "8083:post-service"
    "8084:product-service"
)

i=1
for entry in "${PORTS[@]}"; do
    IFS=':' read -r port name <<< "$entry"
    echo -e "[$i/5] 停止 ${YELLOW}$name${NC} (:$port)..."
    
    if command -v lsof &> /dev/null; then
        # macOS / Linux with lsof
        pids=$(lsof -ti ":$port" 2>/dev/null)
        if [ -n "$pids" ]; then
            for pid in $pids; do
                kill -9 "$pid" 2>/dev/null
                echo "       进程 $pid 已终止"
            done
        fi
    else
        # Linux with fuser
        fuser -k "${port}/tcp" 2>/dev/null
    fi
    
    ((i++))
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  所有服务已停止!${NC}"
echo -e "${GREEN}========================================${NC}"
