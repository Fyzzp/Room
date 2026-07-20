#!/bin/bash
# ==============================================
#  Room — 主控面板一键安装脚本 v0.0.1
#  用法: curl -fsSL https://raw.githubusercontent.com/Fyzzp/Room/main/install.sh | sudo bash
# ==============================================
set -e

GITHUB_REPO="Fyzzp/Room"
VERSION="${VERSION:-v0.0.1}"
INSTALL_DIR="${INSTALL_DIR:-/opt/room}"
DATA_DIR="${DATA_DIR:-/var/lib/room}"
PANEL_PORT="${PANEL_PORT:-12889}"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

if [ "$EUID" -ne 0 ]; then echo -e "${RED}请使用 root 权限${NC}"; exit 1; fi

ARCH=$(uname -m)
case $ARCH in
    x86_64)  ARCH_NAME="amd64" ;;
    aarch64|arm64) ARCH_NAME="arm64" ;;
    *) echo -e "${RED}不支持的架构: $ARCH${NC}"; exit 1 ;;
esac

echo -e "${GREEN}==========================================${NC}"
echo -e "${GREEN}  Room 主控面板 v${VERSION} 一键安装${NC}"
echo -e "${GREEN}==========================================${NC}"

# === 1. 检测依赖 ===
echo -e "${YELLOW}[1/6]${NC} 检查依赖..."
if ! command -v docker >/dev/null 2>&1; then
    echo "安装 Docker..."
    curl -fsSL https://get.docker.com | bash
fi

# === 2. 创建目录 ===
echo -e "${YELLOW}[2/6]${NC} 创建目录..."
mkdir -p "$INSTALL_DIR" "$DATA_DIR"/{postgres,redis,master}

# === 3. 下载二进制（只下载需要的主控二进制文件，不拉源码）===
echo -e "${YELLOW}[3/6]${NC} 下载 Room 二进制..."
BIN_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/room-linux-${ARCH_NAME}"

if command -v curl >/dev/null 2>&1; then
    curl -fsSL --connect-timeout 10 --max-time 300 -o "$INSTALL_DIR/room" "$BIN_URL"
else
    wget -q --connect-timeout=10 --read-timeout=300 -O "$INSTALL_DIR/room" "$BIN_URL"
fi
chmod +x "$INSTALL_DIR/room"

# 下载前端静态文件
echo -e "${YELLOW}      ${NC} 下载前端文件..."
WEB_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/room-web-dist.tar.gz"
if command -v curl >/dev/null 2>&1; then
    curl -fsSL --connect-timeout 10 --max-time 120 -o /tmp/room-web.tar.gz "$WEB_URL"
else
    wget -q --connect-timeout=10 --read-timeout=120 -O /tmp/room-web.tar.gz "$WEB_URL"
fi
mkdir -p "$INSTALL_DIR/web"
tar xzf /tmp/room-web.tar.gz -C "$INSTALL_DIR/web" 2>/dev/null || echo "前端文件解压失败，将使用嵌入式前端"
rm -f /tmp/room-web.tar.gz

# === 4. 生成配置 ===
echo -e "${YELLOW}[4/6]${NC} 生成配置..."
JWT_SECRET=$(cat /dev/urandom 2>/dev/null | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1 || echo "change-me-$(date +%s)")

cat > "$INSTALL_DIR/docker-compose.yml" << COMPOSE_EOF
version: '3.8'
services:
  postgres:
    image: postgres:17-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: room
      POSTGRES_PASSWORD: room
      POSTGRES_DB: room
    volumes:
      - ${DATA_DIR}/postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U room"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    volumes:
      - ${DATA_DIR}/redis:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  master:
    image: alpine:3.21  # 轻量运行时容器
    restart: unless-stopped
    ports:
      - "${PANEL_PORT}:12889"
    environment:
      PORT: "12889"
      DB_HOST: postgres
      DB_PORT: "5432"
      DB_USER: room
      DB_PASSWORD: room
      DB_NAME: room
      REDIS_ADDR: redis:6379
      JWT_SECRET: "${JWT_SECRET}"
      LOG_LEVEL: info
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ${INSTALL_DIR}/room:/app/room:ro
      - ${INSTALL_DIR}/web:/app/web:ro
      - ${DATA_DIR}/master:/app/data
    entrypoint: ["/app/room"]
COMPOSE_EOF

# === 5. 启动服务 ===
echo -e "${YELLOW}[5/6]${NC} 启动服务..."
cd "$INSTALL_DIR"
docker compose up -d

# === 6. 等待就绪 ===
echo -e "${YELLOW}[6/6]${NC} 等待面板启动..."
for i in $(seq 1 30); do
    if curl -sf "http://127.0.0.1:${PANEL_PORT}/api/health" >/dev/null 2>&1; then
        break
    fi
    sleep 2
done

SERVER_IP=$(curl -s4 ifconfig.me 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}')

echo ""
echo -e "${GREEN}==========================================${NC}"
echo -e "${GREEN}  Room 安装完成！${NC}"
echo -e "${GREEN}==========================================${NC}"
echo -e "访问:     ${GREEN}http://${SERVER_IP}:${PANEL_PORT}${NC}"
echo -e "目录:     ${INSTALL_DIR}"
echo -e "数据:     ${DATA_DIR}"
echo ""
echo "管理命令:"
echo "  启动:   cd ${INSTALL_DIR} && docker compose up -d"
echo "  停止:   cd ${INSTALL_DIR} && docker compose down"
echo "  日志:   cd ${INSTALL_DIR} && docker compose logs -f"
echo "  更新:   cd ${INSTALL_DIR} && 重新下载二进制后重启"
echo "  卸载:   cd ${INSTALL_DIR} && docker compose down -v && rm -rf ${INSTALL_DIR} ${DATA_DIR}"
