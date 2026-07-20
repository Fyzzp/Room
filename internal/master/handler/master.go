// Package handler 提供主控 HTTP API 处理器
package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// MasterHandler 主控 HTTP 处理器
type MasterHandler struct {
	db *sql.DB
	// redisClient *redis.Client
	wsHandler interface{} // *ws.Handler，避免循环引用
}

// NewMasterHandler 创建处理器
func NewMasterHandler(db *sql.DB) *MasterHandler {
	return &MasterHandler{
		db: db,
	}
}

// SetWSHandler 注入 WebSocket 处理器
func (h *MasterHandler) SetWSHandler(ws interface{}) {
	h.wsHandler = ws
}

// JSON 响应工具
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	jsonResponse(w, http.StatusOK, data)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// HealthCheck 健康检查
func (h *MasterHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]interface{}{
		"status":    "ok",
		"version":   "0.1.0",
		"db":        h.db != nil,
		"connected": 0,
	})
}

// ListServers 获取服务器列表
func (h *MasterHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, name, token, ip_address, ip_address_v6, listen_port,
		       connection_mode, xray_mode, status, last_heartbeat,
		       upload_speed, download_speed, created_at
		FROM servers ORDER BY created_at DESC
	`)
	if err != nil {
		jsonError(w, 500, "Database error")
		return
	}
	defer rows.Close()

	type Server struct {
		ID             int64   `json:"id"`
		Name           string  `json:"name"`
		Token          string  `json:"token"`
		IPAddress      *string `json:"ip_address"`
		IPAddressV6    *string `json:"ip_address_v6"`
		ListenPort     int     `json:"listen_port"`
		ConnectionMode string  `json:"connection_mode"`
		XrayMode       string  `json:"xray_mode"`
		Status         string  `json:"status"`
		LastHeartbeat  *string `json:"last_heartbeat"`
		UploadSpeed    int64   `json:"upload_speed"`
		DownloadSpeed  int64   `json:"download_speed"`
		CreatedAt      string  `json:"created_at"`
	}

	var servers []Server
	for rows.Next() {
		var s Server
		var lastHb sql.NullString
		if err := rows.Scan(&s.ID, &s.Name, &s.Token, &s.IPAddress, &s.IPAddressV6,
			&s.ListenPort, &s.ConnectionMode, &s.XrayMode, &s.Status,
			&lastHb, &s.UploadSpeed, &s.DownloadSpeed, &s.CreatedAt); err != nil {
			continue
		}
		if lastHb.Valid {
			s.LastHeartbeat = &lastHb.String
		}
		servers = append(servers, s)
	}

	if servers == nil {
		servers = []Server{}
	}

	jsonOK(w, map[string]interface{}{"servers": servers})
}

// CreateServer 添加服务器
func (h *MasterHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, 405, "Method not allowed")
		return
	}

	var req struct {
		Name           string `json:"name"`
		ConnectionMode string `json:"connection_mode"`
		XrayMode       string `json:"xray_mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "Invalid JSON")
		return
	}

	if req.Name == "" {
		jsonError(w, 400, "Name is required")
		return
	}

	if req.ConnectionMode == "" {
		req.ConnectionMode = "websocket"
	}
	if req.XrayMode == "" {
		req.XrayMode = "external"
	}

	// 生成随机 token
	token := generateToken(32)

	var id int64
	err := h.db.QueryRow(`
		INSERT INTO servers (name, token, connection_mode, xray_mode)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, req.Name, token, req.ConnectionMode, req.XrayMode).Scan(&id)

	if err != nil {
		jsonError(w, 500, "Failed to create server: "+err.Error())
		return
	}

	jsonOK(w, map[string]interface{}{
		"id":    id,
		"name":  req.Name,
		"token": token,
	})
}

// DeleteServer 删除服务器
func (h *MasterHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		jsonError(w, 405, "Method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		jsonError(w, 400, "id is required")
		return
	}

	result, err := h.db.Exec("DELETE FROM servers WHERE id = $1", id)
	if err != nil {
		jsonError(w, 500, "Failed to delete server")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, 404, "Server not found")
		return
	}

	jsonOK(w, map[string]string{"message": "Server deleted"})
}

// GetInstallScript 获取 Agent 安装脚本
func (h *MasterHandler) GetInstallScript(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		jsonError(w, 400, "token is required")
		return
	}

	// 获取主控 URL（优先使用 r.Host）
	masterHost := r.Host
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		masterHost = forwarded
	}

	protocol := "https"
	if r.TLS == nil {
		protocol = "http"
	}

	masterURL := protocol + "://" + masterHost

	// shell-escape 防止命令注入
	masterURL = shellEscape(masterURL)
	token = shellEscape(token)

	script := `#!/bin/bash
set -e
echo "=== Xray Panel Agent Installer ==="
MASTER_URL="` + masterURL + `"
TOKEN="` + token + `"

# 检测架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)  ARCH_NAME="amd64" ;;
    aarch64|arm64) ARCH_NAME="arm64" ;;
    *) echo "Unsupported: $ARCH"; exit 1 ;;
esac

echo "[1/4] Downloading agent..."
curl -fsSL --connect-timeout 10 --max-time 180 \
    -o /tmp/xray-agent \
    "https://github.com/Fyzzp/Room/releases/latest/download/xray-agent-linux-${ARCH_NAME}"
chmod +x /tmp/xray-agent
mv /tmp/xray-agent /usr/local/bin/xray-agent

echo "[2/4] Creating config..."
mkdir -p /etc/xray-agent
cat > /etc/xray-agent/config.yaml << EOF
mode: remote
master_url: ${MASTER_URL}
token: ${TOKEN}
connection_mode: auto
xray_mode: external
xray_config_path: /usr/local/etc/xray/config.json
listen_port: 23889
EOF

echo "[3/4] Installing Xray..."
if ! command -v xray >/dev/null 2>&1; then
    bash -c "$(curl -L https://github.com/XTLS/Xray-install/raw/main/install-release.sh)" @ install
fi

echo "[4/4] Creating service..."
cat > /etc/systemd/system/xray-agent.service << EOF
[Unit]
Description=Xray Panel Agent
After=network.target
[Service]
Type=simple
ExecStart=/usr/local/bin/xray-agent -c /etc/xray-agent/config.yaml
Restart=always
RestartSec=5
[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable xray-agent
systemctl start xray-agent

echo ""
echo "=== Installation Complete ==="
echo "Check status: systemctl status xray-agent"
`

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(script))
}

// generateToken 生成加密安全随机 token
func generateToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// shellEscape 转义字符串以防止 shell 命令注入
func shellEscape(s string) string {
	result := make([]byte, 0, len(s)+2)
	result = append(result, '\'')
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			result = append(result, '\'', '\\', '\'', '\'')
		} else {
			result = append(result, s[i])
		}
	}
	result = append(result, '\'')
	return string(result)
}
