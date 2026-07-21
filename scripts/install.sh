#!/bin/bash
# ==============================================
#  Room — Xray 多机管理面板 安装脚本
#  安装:   curl -sL https://raw.githubusercontent.com/Fyzzp/Room/main/scripts/install.sh | sudo bash
#  非交互: DB_USER=xx DB_PASS=xx ... curl ... | sudo bash
#  更新:   curl -sL https://raw.githubusercontent.com/Fyzzp/Room/main/scripts/install.sh | sudo bash -s update
#  卸载:   curl -sL https://raw.githubusercontent.com/Fyzzp/Room/main/scripts/install.sh | sudo bash -s uninstall
# ==============================================
set -e

GITHUB_REPO="Fyzzp/Room"
VERSION="" ; BINARY_NAME="" ; WEB_TAR_NAME="room-web-dist.tar.gz"
INSTALL_DIR="/opt/room"
SERVICE_NAME="room"
CONFIG_FILE="$INSTALL_DIR/config.json"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
ask()   { echo -ne "${BLUE}[?]${NC} $1: "; }

check_root() {
    if [ "$EUID" -ne 0 ]; then error "请使用 root: sudo bash install.sh"; exit 1; fi
}

check_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in x86_64|amd64) BINARY_NAME="room-linux-amd64" ;; aarch64|arm64) BINARY_NAME="room-linux-arm64" ;; *) error "不支持: $ARCH"; exit 1 ;; esac
    info "架构: $ARCH → $BINARY_NAME"
}

get_latest_version() {
    if [ -z "$VERSION" ]; then
        VERSION=$(curl -sL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null | grep -o '"tag_name": *"[^"]*"' | head -1 | grep -o '"[^"]*"$' | tr -d '"')
        [ -z "$VERSION" ] && VERSION="v0.0.1"
    fi
    info "版本: $VERSION"
}

# 交互式收集数据库和 Redis 配置
collect_config() {
    echo ""
    echo -e "${GREEN}==========================================${NC}"
    echo -e "${GREEN}  数据库 & 缓存 配置${NC}"
    echo -e "${GREEN}==========================================${NC}"
    echo -e "按回车使用默认值（方括号内）"
    echo ""

    # 数据库
    if [ -z "$DB_HOST" ]; then ask "数据库主机"; read DB_HOST; [ -z "$DB_HOST" ] && DB_HOST="localhost"; fi
    if [ -z "$DB_PORT" ]; then ask "数据库端口"; read DB_PORT; [ -z "$DB_PORT" ] && DB_PORT="5432"; fi
    if [ -z "$DB_NAME" ]; then ask "数据库名称"; read DB_NAME; [ -z "$DB_NAME" ] && DB_NAME="room"; fi
    if [ -z "$DB_USER" ]; then ask "数据库用户"; read DB_USER; [ -z "$DB_USER" ] && DB_USER="room"; fi
    if [ -z "$DB_PASS" ]; then ask "数据库密码 (留空自动生成)"; read DB_PASS; fi
    if [ -z "$DB_PASS" ]; then
        DB_PASS=$(cat /dev/urandom 2>/dev/null | tr -dc 'a-zA-Z0-9' | fold -w 20 | head -n 1)
        [ -z "$DB_PASS" ] && DB_PASS="room-$(date +%s)"
        info "已生成随机密码: $DB_PASS"
    fi

    echo ""
    # Redis
    if [ -z "$REDIS_HOST" ]; then ask "Redis 主机"; read REDIS_HOST; [ -z "$REDIS_HOST" ] && REDIS_HOST="localhost"; fi
    if [ -z "$REDIS_PORT" ]; then ask "Redis 端口"; read REDIS_PORT; [ -z "$REDIS_PORT" ] && REDIS_PORT="6379"; fi
    if [ -z "$REDIS_PREFIX" ]; then ask "Redis Key 前缀"; read REDIS_PREFIX; [ -z "$REDIS_PREFIX" ] && REDIS_PREFIX="room:"; fi
    if [ -z "$REDIS_USER" ]; then ask "Redis 用户名 (留空=无)"; read REDIS_USER; fi
    if [ -z "$REDIS_PASS" ]; then ask "Redis 密码 (留空=无)"; read REDIS_PASS; fi

    # 面板端口
    if [ -z "$PANEL_PORT" ]; then ask "面板端口"; read PANEL_PORT; [ -z "$PANEL_PORT" ] && PANEL_PORT="12889"; fi

    echo ""
    info "配置收集完成"
}

# 保存配置到 config.json
save_config() {
    mkdir -p "$INSTALL_DIR"
    cat > "$CONFIG_FILE" <<EOF
{
  "port": "$PANEL_PORT",
  "db_host": "$DB_HOST",
  "db_port": "$DB_PORT",
  "db_name": "$DB_NAME",
  "db_user": "$DB_USER",
  "db_password": "$DB_PASS",
  "redis_host": "$REDIS_HOST",
  "redis_port": "$REDIS_PORT",
  "redis_prefix": "$REDIS_PREFIX",
  "redis_user": "$REDIS_USER",
  "redis_password": "$REDIS_PASS",
  "jwt_secret": "$(cat /dev/urandom 2>/dev/null | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)",
  "log_level": "info"
}
EOF
    chmod 600 "$CONFIG_FILE"
    info "配置已保存到 $CONFIG_FILE"
}

install_deps() {
    info "安装系统依赖..."
    if command -v apt-get >/dev/null 2>&1; then
        apt-get update -qq && apt-get install -y curl postgresql postgresql-client redis-server
    elif command -v dnf >/dev/null 2>&1; then
        dnf install -y curl postgresql postgresql-server redis
    elif command -v yum >/dev/null 2>&1; then
        yum install -y curl postgresql postgresql-server redis
    elif command -v apk >/dev/null 2>&1; then
        apk add --no-cache curl postgresql postgresql-client redis
    else
        error "未检测到支持的包管理器（apt/dnf/yum/apk），请手动安装 postgresql 和 redis"; exit 1
    fi
    systemctl enable postgresql redis-server 2>/dev/null || systemctl enable redis 2>/dev/null || true
    systemctl start postgresql redis-server 2>/dev/null || systemctl start redis 2>/dev/null || true
    info "PostgreSQL + Redis 已启动"
}

setup_database() {
    info "配置数据库..."

    # 校验输入仅含安全字符（防 SQL 注入）
    if ! [[ "$DB_USER" =~ ^[a-zA-Z0-9_]+$ ]]; then error "DB_USER 仅允许字母数字下划线"; exit 1; fi
    if ! [[ "$DB_NAME" =~ ^[a-zA-Z0-9_]+$ ]]; then error "DB_NAME 仅允许字母数字下划线"; exit 1; fi

    # 通过 heredoc 传 SQL 到 psql，避免 su -c 的 shell 命令注入
    run_psql() {
        su - postgres -c "psql" <<SQL_EOF
$1
SQL_EOF
    }

    run_psql "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER';" 2>/dev/null | grep -q 1 || \
        run_psql "CREATE USER \"$DB_USER\" WITH PASSWORD '${DB_PASS//\'/\'\'}';"
    run_psql "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';" 2>/dev/null | grep -q 1 || \
        run_psql "CREATE DATABASE \"$DB_NAME\" OWNER \"$DB_USER\";"
    run_psql "ALTER USER \"$DB_USER\" WITH PASSWORD '${DB_PASS//\'/\'\'}';"
    info "数据库就绪"
}

# Redis 密码配置（如果有）
setup_redis() {
    if [ -n "$REDIS_PASS" ]; then
        # 使用 | 作为 sed 分隔符，防止密码中的 / 破坏表达式
        if grep -q "^requirepass" /etc/redis/redis.conf 2>/dev/null; then
            sed -i "s|^requirepass.*|requirepass $REDIS_PASS|" /etc/redis/redis.conf
        else
            printf 'requirepass %s\n' "$REDIS_PASS" >> /etc/redis/redis.conf
        fi
        systemctl restart redis-server 2>/dev/null || systemctl restart redis 2>/dev/null || true
        info "Redis 密码已设置"
    fi
}

download_release() {
    info "下载 Room $VERSION..."
    BIN_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}"
    WEB_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${WEB_TAR_NAME}"

    for url in "$BIN_URL" "https://gh-proxy.com/$BIN_URL" "https://mirror.ghproxy.com/$BIN_URL"; do
        if curl -fsSL --connect-timeout 10 --max-time 300 -o "/tmp/${BINARY_NAME}" "$url" 2>/dev/null; then break; fi
        warn "镜像失败，尝试下一个..."
    done

    curl -fsSL --connect-timeout 10 --max-time 120 -o "/tmp/${WEB_TAR_NAME}" "$WEB_URL" 2>/dev/null || \
        warn "前端文件下载失败"
    info "下载完成"
}

install_files() {
    info "安装文件..."
    mkdir -p "$INSTALL_DIR"
    chmod +x "/tmp/${BINARY_NAME}"
    mv "/tmp/${BINARY_NAME}" "$INSTALL_DIR/room"

    if [ -f "/tmp/${WEB_TAR_NAME}" ]; then
        mkdir -p "$INSTALL_DIR/web"
        tar xzf "/tmp/${WEB_TAR_NAME}" -C "$INSTALL_DIR/web" 2>/dev/null || true
        rm -f "/tmp/${WEB_TAR_NAME}"
    fi
    rm -f "/tmp/${BINARY_NAME}"
}

create_service() {
    info "创建 systemd 服务（端口: $PANEL_PORT）"

    # 构建 Redis 连接串
    REDIS_ADDR="${REDIS_HOST}:${REDIS_PORT}"
    [ -n "$REDIS_USER" ] && REDIS_ADDR="${REDIS_USER}@${REDIS_ADDR}"

    cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Room - Xray 多机管理面板
After=network.target postgresql.service redis-server.service
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/room -c $CONFIG_FILE
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

Environment="PORT=$PANEL_PORT"
Environment="DB_HOST=$DB_HOST"
Environment="DB_PORT=$DB_PORT"
Environment="DB_USER=$DB_USER"
Environment="DB_PASSWORD=$DB_PASS"
Environment="DB_NAME=$DB_NAME"
Environment="REDIS_ADDR=$REDIS_ADDR"
Environment="REDIS_PREFIX=$REDIS_PREFIX"
Environment="REDIS_PASSWORD=$REDIS_PASS"
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
    PORT="${PANEL_PORT:-12889}"
    if ss -tlnp 2>/dev/null | grep -q ":${PORT} "; then
        warn "端口 $PORT 被占用，尝试清理..."
        fuser -k ${PORT}/tcp 2>/dev/null || true; sleep 1
    fi

    info "启动服务..."
    systemctl enable ${SERVICE_NAME}.service 2>/dev/null || true
    systemctl start ${SERVICE_NAME}.service; sleep 3

    if systemctl is-active --quiet ${SERVICE_NAME}.service; then
        info "服务启动成功"; return 0
    else
        error "服务启动失败"; return 1
    fi
}

show_done() {
    PORT="${PANEL_PORT:-12889}"
    IP=$(hostname -I 2>/dev/null | awk '{print $1}')
    [ -z "$IP" ] && IP="YOUR_SERVER_IP"

    echo ""
    echo "======================================"
    info "Room 安装完成！"
    echo "======================================"
    echo ""
    echo "  访问:     http://${IP}:${PORT}"
    echo "  目录:     $INSTALL_DIR"
    echo "  配置:     $CONFIG_FILE"
    echo "  数据库:   PostgreSQL ($DB_USER@$DB_HOST:$DB_PORT/$DB_NAME)"
    echo "  缓存:     Redis ($REDIS_HOST:$REDIS_PORT)"
    echo ""
    echo "管理命令:"
    echo "  状态:     systemctl status $SERVICE_NAME"
    echo "  日志:     journalctl -u $SERVICE_NAME -f"
    echo "  更新:     curl -sL https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/install.sh | sudo bash -s update"
    echo "  卸载:     curl -sL https://raw.githubusercontent.com/${GITHUB_REPO}/main/scripts/install.sh | sudo bash -s uninstall"
    echo ""
}

update_service() {
    if [ ! -f "$INSTALL_DIR/room" ]; then error "未安装"; exit 1; fi
    info "更新 Room..."
    # 加载已有配置（从 systemd Environment 安全提取）
    if [ -f /etc/systemd/system/${SERVICE_NAME}.service ]; then
        DB_HOST=$(grep -oP 'DB_HOST=\K[^"]*' /etc/systemd/system/${SERVICE_NAME}.service 2>/dev/null || echo "localhost")
        DB_PORT=$(grep -oP 'DB_PORT=\K[^"]*' /etc/systemd/system/${SERVICE_NAME}.service 2>/dev/null || echo "5432")
        DB_NAME=$(grep -oP 'DB_NAME=\K[^"]*' /etc/systemd/system/${SERVICE_NAME}.service 2>/dev/null || echo "room")
        DB_USER=$(grep -oP 'DB_USER=\K[^"]*' /etc/systemd/system/${SERVICE_NAME}.service 2>/dev/null || echo "room")
        DB_PASS=$(grep -oP 'DB_PASSWORD=\K[^"]*' /etc/systemd/system/${SERVICE_NAME}.service 2>/dev/null || echo "room")
        PANEL_PORT=$(grep -oP 'PORT=\K[^"]*' /etc/systemd/system/${SERVICE_NAME}.service 2>/dev/null || echo "12889")
        REDIS_HOST="localhost"; REDIS_PORT="6379"; REDIS_PREFIX="room:"; REDIS_USER=""; REDIS_PASS=""
    fi

    systemctl stop ${SERVICE_NAME}.service || true
    cp "$INSTALL_DIR/room" "$INSTALL_DIR/room.bak" 2>/dev/null || true
    download_release; install_files; create_service
    if start_service; then
        rm -f "$INSTALL_DIR/room.bak"; show_done
    else
        error "更新失败，回滚..."
        mv "$INSTALL_DIR/room.bak" "$INSTALL_DIR/room" 2>/dev/null || true
        systemctl start ${SERVICE_NAME}.service || true; exit 1
    fi
}

uninstall_service() {
    [ ! -f "$INSTALL_DIR/room" ] && { error "未安装"; exit 1; }
    info "卸载 Room..."
    systemctl stop ${SERVICE_NAME}.service 2>/dev/null || true
    systemctl disable ${SERVICE_NAME}.service 2>/dev/null || true
    rm -f /etc/systemd/system/${SERVICE_NAME}.service
    systemctl daemon-reload
    rm -rf "$INSTALL_DIR"
    info "卸载完成（PostgreSQL/Redis 未删除）"
}

main() {
    case "$1" in
        update)    check_root; check_arch; get_latest_version; update_service ;;
        uninstall) check_root; uninstall_service ;;
        *)
            check_root; check_arch
            collect_config   # 交互式收集配置
            install_deps; get_latest_version
            setup_database; setup_redis
            save_config
            download_release; install_files; create_service
            if start_service; then show_done
            else error "安装失败: journalctl -u $SERVICE_NAME -n 50"; exit 1; fi
            ;;
    esac
}

main "$@"
