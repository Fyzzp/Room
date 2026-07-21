#!/bin/bash
# ==============================================
#  Room — Xray 多机管理面板 安装脚本
#  安装:   curl -sL https://raw.githubusercontent.com/Fyzzp/Room/main/scripts/install.sh | sudo bash
#  更新:   curl -sL https://raw.githubusercontent.com/Fyzzp/Room/main/scripts/install.sh | sudo bash -s update
#  卸载:   curl -sL https://raw.githubusercontent.com/Fyzzp/Room/main/scripts/install.sh | sudo bash -s uninstall
# ==============================================
set -e

GITHUB_REPO="Fyzzp/Room"
VERSION="" ; BINARY_NAME="" ; WEB_TAR_NAME="room-web-dist.tar.gz"
INSTALL_DIR="/usr/local/bin" ; SERVICE_NAME="room"
DATA_DIR="/etc/room" ; WEB_DIR="$DATA_DIR/web"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

check_root() {
    if [ "$EUID" -ne 0 ]; then error "请使用 root: sudo bash install.sh"; exit 1; fi
}

check_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) BINARY_NAME="room-linux-amd64" ;;
        aarch64|arm64) BINARY_NAME="room-linux-arm64" ;;
        *) error "不支持的架构: $ARCH"; exit 1 ;;
    esac
    info "架构: $ARCH → $BINARY_NAME"
}

get_latest_version() {
    if [ -z "$VERSION" ]; then
        info "获取最新版本..."
        VERSION=$(curl -sL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null | grep -o '"tag_name": *"[^"]*"' | head -1 | grep -o '"[^"]*"$' | tr -d '"')
        [ -z "$VERSION" ] && VERSION="v0.0.1"
    fi
    info "版本: $VERSION"
}

install_deps() {
    info "安装系统依赖 (PostgreSQL, Redis)..."
    apt-get update -qq
    apt-get install -y curl postgresql postgresql-client redis-server >/dev/null 2>&1
    # 启动 PG 和 Redis
    systemctl enable postgresql redis-server 2>/dev/null || true
    systemctl start postgresql redis-server 2>/dev/null || true
    info "PostgreSQL + Redis 已安装并启动"
}

setup_database() {
    info "配置数据库..."
    mkdir -p "$DATA_DIR"

    # 生成随机密码
    DB_PASS=$(cat /dev/urandom 2>/dev/null | tr -dc 'a-zA-Z0-9' | fold -w 20 | head -n 1)
    [ -z "$DB_PASS" ] && DB_PASS="room-$(date +%s)-$(od -A n -t u4 /dev/urandom 2>/dev/null | tr -d ' ' || echo $RANDOM)"
    (umask 077; echo "$DB_PASS" > "$DATA_DIR/.db_password")

    # 创建用户和数据库（幂等）— 用外层 Shell 展开 ${DB_PASS}
    su - postgres -c "psql -tc \"SELECT 1 FROM pg_roles WHERE rolname='room'\"" 2>/dev/null | grep -q 1 || \
        su - postgres -c "psql -c \"CREATE USER room WITH PASSWORD '${DB_PASS}';\"" 2>/dev/null || true
    su - postgres -c "psql -tc \"SELECT 1 FROM pg_database WHERE datname='room'\"" 2>/dev/null | grep -q 1 || \
        su - postgres -c "psql -c \"CREATE DATABASE room OWNER room;\"" 2>/dev/null || true
    su - postgres -c "psql -c \"ALTER USER room WITH PASSWORD '${DB_PASS}';\"" 2>/dev/null || true
    info "数据库就绪（密码: $DATA_DIR/.db_password）"
}

download_release() {
    info "下载 Room $VERSION..."
    BIN_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}"
    WEB_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${WEB_TAR_NAME}"

    for url in "$BIN_URL" "https://gh-proxy.com/$BIN_URL" "https://mirror.ghproxy.com/$BIN_URL"; do
        if curl -fsSL --connect-timeout 10 --max-time 300 -o "/tmp/${BINARY_NAME}" "$url" 2>/dev/null; then
            break
        fi
        warn "二进制镜像失败，尝试下一个..."
    done

    # 下载前端文件
    curl -fsSL --connect-timeout 10 --max-time 120 -o "/tmp/${WEB_TAR_NAME}" "$WEB_URL" 2>/dev/null || \
        warn "前端文件下载失败，将使用嵌入式前端"

    info "下载完成"
}

install_files() {
    info "安装文件..."
    chmod +x "/tmp/${BINARY_NAME}"
    mv "/tmp/${BINARY_NAME}" "$INSTALL_DIR/$SERVICE_NAME"

    mkdir -p "$WEB_DIR"
    if [ -f "/tmp/${WEB_TAR_NAME}" ]; then
        tar xzf "/tmp/${WEB_TAR_NAME}" -C "$WEB_DIR" 2>/dev/null || true
        rm -f "/tmp/${WEB_TAR_NAME}"
    fi
    rm -f "/tmp/${BINARY_NAME}"
}

create_service() {
    PORT="${PORT:-12889}"
    # 获取密码：优先 .db_password，否则从现有 systemd unit 提取（旧版迁移），最后回退 room
    if [ -f "$DATA_DIR/.db_password" ]; then
        DB_PASS=$(cat "$DATA_DIR/.db_password")
    elif [ -f /etc/systemd/system/${SERVICE_NAME}.service ]; then
        DB_PASS=$(grep -oP 'DB_PASSWORD=\K[^"]*' /etc/systemd/system/${SERVICE_NAME}.service 2>/dev/null || true)
        [ -z "$DB_PASS" ] && DB_PASS="room"
        (umask 077; echo "$DB_PASS" > "$DATA_DIR/.db_password")
        info "从旧部署迁移密码到 $DATA_DIR/.db_password"
    else
        DB_PASS="room"
        echo "$DB_PASS" > "$DATA_DIR/.db_password"
        chmod 600 "$DATA_DIR/.db_password"
        warn "未找到密码文件，使用默认密码（请登录后修改）"
    fi
    info "创建 systemd 服务（端口: $PORT）"

    cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Room - Xray 多机管理面板
After=network.target postgresql.service redis-server.service
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=$DATA_DIR
ExecStart=$INSTALL_DIR/$SERVICE_NAME
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

Environment="PORT=$PORT"
Environment="DB_HOST=localhost"
Environment="DB_PORT=5432"
Environment="DB_USER=room"
Environment="DB_PASSWORD=${DB_PASS}"
Environment="DB_NAME=room"
Environment="REDIS_ADDR=localhost:6379"
Environment="LOG_LEVEL=info"

NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    chmod 600 /etc/systemd/system/${SERVICE_NAME}.service
}

start_service() {
    PORT=$(grep 'PORT=' /etc/systemd/system/${SERVICE_NAME}.service | sed 's/.*PORT=\([0-9]*\).*/\1/')
    PORT=${PORT:-12889}

    # 清理占用端口的旧进程（如旧 Docker 部署残留）
    if ss -tlnp 2>/dev/null | grep -q ":${PORT} "; then
        warn "端口 $PORT 被占用，尝试清理..."
        fuser -k ${PORT}/tcp 2>/dev/null || true
        sleep 1
    fi

    info "启动服务..."
    systemctl enable ${SERVICE_NAME}.service 2>/dev/null || true
    systemctl start ${SERVICE_NAME}.service
    sleep 3

    if systemctl is-active --quiet ${SERVICE_NAME}.service; then
        info "服务启动成功"
        return 0
    else
        error "服务启动失败（端口 $PORT 可能仍被占用）"
        return 1
    fi
}

show_done() {
    PORT=$(grep 'PORT=' /etc/systemd/system/${SERVICE_NAME}.service | sed 's/.*PORT=\([0-9]*\).*/\1/')
    PORT=${PORT:-12889}
    IP=$(hostname -I 2>/dev/null | awk '{print $1}')
    [ -z "$IP" ] && IP="YOUR_SERVER_IP"

    echo ""
    echo "======================================"
    info "Room 安装完成！"
    echo "======================================"
    echo ""
    echo "  访问:     http://${IP}:${PORT}"
    echo "  目录:     $DATA_DIR"
    echo "  二进制:   $INSTALL_DIR/$SERVICE_NAME"
    echo "  数据库:   PostgreSQL (room/room@localhost:5432/room)"
    echo "  缓存:     Redis (localhost:6379)"
    echo ""
    echo "管理命令:"
    echo "  状态:     systemctl status $SERVICE_NAME"
    echo "  日志:     journalctl -u $SERVICE_NAME -f"
    echo "  更新:     curl -sL https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/install.sh | sudo bash -s update"
    echo "  卸载:     curl -sL https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/install.sh | sudo bash -s uninstall"
    echo ""
}

# === 更新 ===
update_service() {
    if [ ! -f "$INSTALL_DIR/$SERVICE_NAME" ]; then error "未安装"; exit 1; fi
    info "更新 Room..."
    systemctl stop ${SERVICE_NAME}.service || true
    cp "$INSTALL_DIR/$SERVICE_NAME" "$INSTALL_DIR/${SERVICE_NAME}.bak" 2>/dev/null || true
    download_release
    install_files
    create_service   # 重建 systemd unit
    if start_service; then
        rm -f "$INSTALL_DIR/${SERVICE_NAME}.bak"
        show_done
    else
        error "更新后启动失败，回滚..."
        mv "$INSTALL_DIR/${SERVICE_NAME}.bak" "$INSTALL_DIR/$SERVICE_NAME" 2>/dev/null || true
        systemctl start ${SERVICE_NAME}.service || true
        exit 1
    fi
}

# === 卸载 ===
uninstall_service() {
    [ ! -f "$INSTALL_DIR/$SERVICE_NAME" ] && { error "未安装"; exit 1; }
    info "卸载 Room..."
    systemctl stop ${SERVICE_NAME}.service 2>/dev/null || true
    systemctl disable ${SERVICE_NAME}.service 2>/dev/null || true
    rm -f /etc/systemd/system/${SERVICE_NAME}.service
    systemctl daemon-reload
    rm -f "$INSTALL_DIR/$SERVICE_NAME" "$INSTALL_DIR/${SERVICE_NAME}.bak"
    rm -rf "$DATA_DIR"
    info "卸载完成（PostgreSQL/Redis 未删除）"
}

# === 主流程 ===
main() {
    case "$1" in
        update)    check_root; check_arch; get_latest_version; update_service ;;
        uninstall) check_root; uninstall_service ;;
        *)
            check_root; check_arch; install_deps; get_latest_version
            setup_database; download_release; install_files; create_service
            if start_service; then show_done
            else error "安装失败: journalctl -u $SERVICE_NAME -n 50"; exit 1; fi
            ;;
    esac
}

main "$@"
