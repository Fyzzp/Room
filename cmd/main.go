// Master 是 Xray 多机管理面板的主控服务
// 职责：
//   - 提供 Web 管理界面 (React SPA 嵌入式)
//   - 管理用户、服务器、入站/出站/路由配置
//   - 通过 gRPC/WebSocket 与多个 Agent 通信
//   - 流量统计汇总与展示
//   - 生成订阅链接（支持多客户端格式）
package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/Fyzzp/Room/internal/master/handler"
)

// Config 主控配置
type Config struct {
	Port        string `json:"port"`
	DBHost      string `json:"db_host"`
	DBPort      string `json:"db_port"`
	DBUser      string `json:"db_user"`
	DBPassword  string `json:"db_password"`
	DBName      string `json:"db_name"`
	RedisAddr   string `json:"redis_addr"`
	JWTSecret   string `json:"jwt_secret"`
	LogLevel    string `json:"log_level"`
}

func main() {
	configPath := flag.String("c", "config.json", "配置文件路径")
	flag.Parse()

	cfg := loadMasterConfig(*configPath)

	// 如果 JWT 密钥为空，自动生成一个（生产环境应通过环境变量或配置文件设置）
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = generateRandomSecret(32)
		log.Printf("[WARNING] JWT_SECRET not set, generated random secret. "+
			"Set JWT_SECRET environment variable for persistence across restarts.")
	}

	log.Printf("=== Room Panel Master v0.1.0 ===")
	log.Printf("Port: %s | DB: %s:%s/%s | Redis: %s", cfg.Port, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.RedisAddr)

	// 连接 PostgreSQL
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("[DB] PostgreSQL connected")

	// TODO: Redis 连接
	// redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	// 运行数据库迁移
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 预置管理员账户（从环境变量，安装脚本设置）
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminEmail != "" && adminPass != "" {
		var existing int
		db.QueryRow("SELECT COUNT(*) FROM users").Scan(&existing)
		if existing == 0 {
			hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
			if err == nil {
				db.Exec("INSERT INTO users (email, password_hash, role) VALUES ($1, $2, 'admin') ON CONFLICT (email) DO NOTHING", adminEmail, string(hash))
				log.Printf("[INIT] Admin account created: %s", adminEmail)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置 HTTP 路由
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"version": "0.1.0",
			"time":    time.Now().Unix(),
		})
	})

	// 认证相关
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonError(w, 405, "Method not allowed")
			return
		}
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, 400, "Invalid JSON")
			return
		}
		if req.Email == "" || req.Password == "" {
			jsonError(w, 400, "邮箱和密码不能为空")
			return
		}

		var userID int
		var role, hash string
		err := db.QueryRow("SELECT id, role, password_hash FROM users WHERE email=$1", req.Email).Scan(&userID, &role, &hash)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
			jsonError(w, 401, "邮箱或密码错误")
			return
		}

		token := generateJWT(cfg.JWTSecret, req.Email, role)
		jsonOK(w, map[string]interface{}{
			"token": token,
			"user":  map[string]interface{}{"email": req.Email, "role": role},
		})
	})

	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonError(w, 405, "Method not allowed")
			return
		}
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, 400, "Invalid JSON")
			return
		}
		if req.Email == "" || len(req.Password) < 8 {
			jsonError(w, 400, "邮箱不能为空且密码至少8位")
			return
		}

		// 检查是否已存在
		var exists int
		err := db.QueryRow("SELECT 1 FROM users WHERE email=$1", req.Email).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			jsonError(w, 500, "数据库错误")
			return
		}
		if exists == 1 {
			jsonError(w, 409, "该邮箱已注册")
			return
		}

		// 所有注册用户均为普通用户（管理员由安装时预置）
		role := "user"

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			jsonError(w, 500, "密码加密失败")
			return
		}

		_, err = db.Exec("INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)", req.Email, string(hash), role)
		if err != nil {
			jsonError(w, 500, "注册失败: "+err.Error())
			return
		}

		token := generateJWT(cfg.JWTSecret, req.Email, role)
		jsonOK(w, map[string]interface{}{
			"token": token,
			"user":  map[string]interface{}{"email": req.Email, "role": role},
		})
	})

	// 创建 handler 实例
	h := handler.NewMasterHandler(db)

	// 服务器管理（需要登录）
	mux.HandleFunc("/api/servers", authMiddleware(cfg.JWTSecret, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListServers(w, r)
		case http.MethodPost:
			h.CreateServer(w, r)
		case http.MethodDelete:
			h.DeleteServer(w, r)
		default:
			jsonError(w, 405, "Method not allowed")
		}
	}))

	// 入站/出站/路由管理
	mux.HandleFunc("/api/inbounds", func(w http.ResponseWriter, r *http.Request) {
		// TODO: CRUD 入站
		json.NewEncoder(w).Encode(map[string]string{"message": "TODO: inbounds"})
	})

	// 流量统计
	mux.HandleFunc("/api/traffic", func(w http.ResponseWriter, r *http.Request) {
		// TODO: 流量查询
		json.NewEncoder(w).Encode(map[string]string{"message": "TODO: traffic"})
	})

	// 订阅
	mux.HandleFunc("/api/subscribe/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: 订阅链接生成
		json.NewEncoder(w).Encode(map[string]string{"message": "TODO: subscribe"})
	})

	// Agent 远程管理端点
	mux.HandleFunc("/api/remote/install-script", func(w http.ResponseWriter, r *http.Request) {
		// TODO: 生成 Agent 一键安装脚本
		token := r.URL.Query().Get("token")
		script := generateInstallScript(r.Host, token)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(script))
	})

	// WebSocket 端点（Agent 连接）
	mux.HandleFunc("/api/agent/ws", func(w http.ResponseWriter, r *http.Request) {
		// TODO: WebSocket upgrade + agent 通信
		json.NewEncoder(w).Encode(map[string]string{"message": "TODO: ws upgrade"})
	})

	// 静态文件（SPA fallback: 非文件路径返回 index.html）
	fs := http.FileServer(http.Dir("./web/dist"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// API 路由不处理
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
			http.NotFound(w, r)
			return
		}
		// 尝试服务静态资源（仅 /assets/ 路径，path.Clean 防穿越）
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			p := filepath.Clean("./web/dist" + r.URL.Path)
			if strings.HasPrefix(p, "web/dist") {
				if _, err := os.Stat(p); err == nil {
					fs.ServeHTTP(w, r)
					return
				}
			}
		}
		// SPA fallback: 所有其他路径返回 index.html
		http.ServeFile(w, r, "./web/dist/index.html")
	})

	// CORS 中间件
	handler := corsMiddleware(mux)

	// 启动 HTTP 服务器
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[HTTP] Master panel listening on http://0.0.0.0:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
	cancel()
	_ = ctx
}

func loadMasterConfig(path string) Config {
	cfg := Config{
		Port:       "12889",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "room",
		DBPassword: "room",
		DBName:     "room",
		RedisAddr:  "localhost:6379",
		JWTSecret:  "",
		LogLevel:   "info",
	}

	// 先读配置文件
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &cfg)
	}

	// 环境变量覆盖（优先级最高）
	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.DBPort = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DBUser = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DBPassword = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.RedisAddr = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}

	return cfg
}

// generateRandomSecret 生成加密安全随机字符串
func generateRandomSecret(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("crypto/rand failed: %v", err)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// sanitizeHost 去除 Host header 中的非法字符，防止 YAML 注入
func sanitizeHost(host string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-:[]", r) {
			return r
		}
		return -1
	}, host)
}

// shellEscape 转义字符串以防止 shell 命令注入
// 将单引号替换为 '\'' 并用单引号包裹整个字符串
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

// runMigrations 执行数据库迁移
func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) DEFAULT 'user',
			traffic_limit BIGINT DEFAULT 0,
			upload BIGINT DEFAULT 0,
			download BIGINT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS servers (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			token VARCHAR(128) UNIQUE NOT NULL,
			ip_address VARCHAR(45),
			ip_address_v6 VARCHAR(45),
			listen_port INTEGER DEFAULT 23889,
			connection_mode VARCHAR(20) DEFAULT 'websocket',
			xray_mode VARCHAR(20) DEFAULT 'embedded',
			status VARCHAR(20) DEFAULT 'pending',
			last_heartbeat TIMESTAMPTZ,
			upload_speed BIGINT DEFAULT 0,
			download_speed BIGINT DEFAULT 0,
			public_key_base64 TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS inbounds (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
			tag VARCHAR(128) NOT NULL,
			protocol VARCHAR(50) NOT NULL,
			port INTEGER NOT NULL,
			listen VARCHAR(45) DEFAULT '0.0.0.0',
			settings JSONB NOT NULL DEFAULT '{}',
			stream_settings JSONB DEFAULT '{}',
			sniffing JSONB DEFAULT '{}',
			status VARCHAR(20) DEFAULT 'active',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS outbounds (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
			tag VARCHAR(128) NOT NULL,
			protocol VARCHAR(50) NOT NULL,
			settings JSONB NOT NULL DEFAULT '{}',
			stream_settings JSONB DEFAULT '{}',
			mux JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS routing_rules (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
			priority INTEGER DEFAULT 0,
			rule JSONB NOT NULL DEFAULT '{}',
			outbound_tag VARCHAR(128),
			enabled BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS traffic_records (
			id SERIAL PRIMARY KEY,
			server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			inbound_tag VARCHAR(128),
			upload BIGINT DEFAULT 0,
			download BIGINT DEFAULT 0,
			recorded_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255),
			token VARCHAR(128) UNIQUE NOT NULL,
			nodes JSONB DEFAULT '[]',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Printf("[DB] Migrations complete (%d tables)", len(migrations))
	return nil
}

// generateInstallScript 生成 Agent 一键安装脚本
func generateInstallScript(masterHost, token string) string {
	// token 需要转义防止注入；masterHost 来自 HTTP Host header，已由 Go 校验
	masterHost = sanitizeHost(masterHost)
	token = shellEscape(token)

	return fmt.Sprintf(`#!/bin/bash
set -e
echo "=== Room Agent Installer ==="
echo "Master: %s"
echo ""

# 检测架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)  ARCH_NAME="amd64" ;;
    aarch64|arm64) ARCH_NAME="arm64" ;;
    *) echo "Unsupported: $ARCH"; exit 1 ;;
esac

# 下载 Agent
echo "[1/3] Downloading agent..."
curl -fsSL -o /tmp/room-agent "https://github.com/Fyzzp/Room-Agent/releases/latest/download/room-agent-linux-${ARCH_NAME}"
chmod +x /tmp/room-agent
mv /tmp/room-agent /usr/local/bin/room-agent

# 创建配置
echo "[2/3] Creating config..."
mkdir -p /etc/room-agent
cat > /etc/room-agent/config.yaml << EOF
mode: remote
master_url: https://%s
token: %s
connection_mode: auto
xray_mode: external
xray_config_path: /usr/local/etc/xray/config.json
listen_port: 23889
EOF

# 创建 systemd 服务
echo "[3/3] Creating service..."
cat > /etc/systemd/system/room-agent.service << EOF
[Unit]
Description=Room Agent
After=network.target
[Service]
Type=simple
ExecStart=/usr/local/bin/room-agent -c /etc/room-agent/config.yaml
Restart=always
RestartSec=5
[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable room-agent
systemctl start room-agent

echo ""
echo "=== Installation Complete ==="
echo "Check status: systemctl status room-agent"
`, masterHost, masterHost, token)
}

// corsMiddleware 处理跨域请求
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// jsonOK 返回 JSON 成功响应
func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// jsonError 返回 JSON 错误响应
func jsonError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// generateJWT 生成 JWT token
func generateJWT(secret, email, role string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	now := time.Now().Unix()
	claims, _ := json.Marshal(map[string]interface{}{
		"email": email,
		"role":  role,
		"iat":   now,
		"exp":   now + 7*24*3600,
	})
	payloadB64 := base64.RawURLEncoding.EncodeToString(claims)
	sigInput := header + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return sigInput + "." + sig
}

// parseJWT 验证并解析 JWT token
func parseJWT(secret, token string) (email, role string, ok bool) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return "", "", false
	}
	// 验证算法
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", false
	}
	var hdr struct{ Alg string `json:"alg"` }
	if json.Unmarshal(header, &hdr) != nil || hdr.Alg != "HS256" {
		return "", "", false
	}
	// 验证签名（时序安全）
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expectedSig := mac.Sum(nil)
	actualSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(expectedSig, actualSig) {
		return "", "", false
	}
	// 解析 payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}
	var claims struct {
		Email string `json:"email"`
		Role  string `json:"role"`
		Exp   int64  `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", "", false
	}
	if time.Now().Unix() > claims.Exp {
		return "", "", false
	}
	return claims.Email, claims.Role, true
}

// authMiddleware 验证 Authorization Bearer token
func authMiddleware(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			jsonError(w, 401, "未登录")
			return
		}
		token := auth[7:]
		if _, _, ok := parseJWT(secret, token); !ok {
			jsonError(w, 401, "token 无效或已过期")
			return
		}
		next(w, r)
	}
}
