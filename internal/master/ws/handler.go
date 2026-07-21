// Package ws 提供主控服务端 WebSocket 处理
// 接收 Agent 的 WebSocket 连接，处理认证、流量、心跳等消息
package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// AgentConnection 代表一个 Agent WebSocket 连接
type AgentConnection struct {
	ServerID   int64
	ServerName string
	Token      string
	Conn       *websocket.Conn
	LastPing   time.Time
	Connected  bool
	Version    string
	PublicIPv4 string
	PublicIPv6 string
	mu         sync.Mutex
	writeMu    sync.Mutex
}

// Handler WebSocket 连接处理器
type Handler struct {
	upgrader websocket.Upgrader
	conns    sync.Map // token -> *AgentConnection

	// 回调函数
	OnAuth      func(conn *AgentConnection) error
	OnTraffic   func(conn *AgentConnection, stats TrafficReport) error
	OnHeartbeat func(conn *AgentConnection, hb HeartbeatReport) error
	OnSpeed     func(conn *AgentConnection, speed SpeedReport) error
	OnDisconnect func(conn *AgentConnection)
}

// TrafficReport 流量上报
type TrafficReport struct {
	Stats interface{} `json:"stats"`
}

// HeartbeatReport 心跳上报
type HeartbeatReport struct {
	BootTime  int64  `json:"boot_time"`
	LocalTime int64  `json:"local_time"`
	PublicIPv4 string `json:"public_ipv4,omitempty"`
	PublicIPv6 string `json:"public_ipv6,omitempty"`
}

// SpeedReport 速度上报
type SpeedReport struct {
	UploadSpeed   int64 `json:"upload_speed"`
	DownloadSpeed int64 `json:"download_speed"`
}

// Message WebSocket 消息
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewHandler 创建 WebSocket 处理器
func NewHandler() *Handler {
	return &Handler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Agent 连接通常不发送 Origin header，直接放行
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				// 有 Origin 时校验其 host:port 是否匹配请求的 Host
				host := r.Host
				if h := r.Header.Get("X-Forwarded-Host"); h != "" {
					host = h
				}
				return originHostMatches(origin, host)
			},
		},
	}
}

// ServeHTTP 处理 WebSocket 升级请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[Master WS] Upgrade failed: %v", err)
		return
	}

	agentConn := &AgentConnection{
		Conn:      conn,
		LastPing:  time.Now(),
		Connected: true,
	}

	defer func() {
		agentConn.writeMu.Lock()
		conn.Close()
		agentConn.writeMu.Unlock()

		agentConn.mu.Lock()
		agentConn.Connected = false
		token := agentConn.Token
		agentConn.mu.Unlock()

		if h.OnDisconnect != nil {
			h.OnDisconnect(agentConn)
		}

		if token != "" {
			// CompareAndDelete 原子化删除，消除 TOCTOU：
			// 只删除仍指向当前连接的条目，重连后新连接不受影响
			h.conns.CompareAndDelete(token, agentConn)
		}
		log.Printf("[Master WS] Agent disconnected: %s", agentConn.ServerName)
	}()

	log.Printf("[Master WS] New connection from %s", r.RemoteAddr)

	// 消息循环
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[Master WS] Read error: %v", err)
			}
			return
		}

		agentConn.mu.Lock()
		agentConn.LastPing = time.Now()
		agentConn.mu.Unlock()

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[Master WS] Invalid message: %v", err)
			continue
		}

		log.Printf("[Master WS] Received: type=%s payload=%s", msg.Type, string(msg.Payload))

		h.handleMessage(agentConn, msg)
	}
}

// handleMessage 分发处理不同类型的消息
func (h *Handler) handleMessage(conn *AgentConnection, msg Message) {
	switch msg.Type {
	case "auth":
		var auth struct {
			Token string `json:"token"`
			PublicIPv4 string `json:"public_ipv4,omitempty"`
			PublicIPv6 string `json:"public_ipv6,omitempty"`
			Version string `json:"agent_version,omitempty"`
		}
		if err := json.Unmarshal(msg.Payload, &auth); err != nil {
			h.sendAuthResult(conn, false, "Invalid auth payload")
			return
		}

		conn.mu.Lock()
		conn.Token = auth.Token
		conn.Version = auth.Version
		conn.PublicIPv4 = auth.PublicIPv4
		conn.PublicIPv6 = auth.PublicIPv6
		conn.mu.Unlock()

		if h.OnAuth != nil {
			if err := h.OnAuth(conn); err != nil {
				h.sendAuthResult(conn, false, err.Error())
				return
			}
		}

		h.conns.Store(auth.Token, conn)
		h.sendAuthResult(conn, true, "Authenticated")
		log.Printf("[Master WS] Agent authenticated: %s (v%s)", conn.ServerName, conn.Version)

	case "heartbeat":
		var hb HeartbeatReport
		if err := json.Unmarshal(msg.Payload, &hb); err != nil {
			return
		}
		if h.OnHeartbeat != nil {
			h.OnHeartbeat(conn, hb)
		}
		// 回复心跳确认
		ack, err := json.Marshal(map[string]int64{"server_time": time.Now().Unix()})
		if err != nil {
			log.Printf("[Master WS] heartbeat ack marshal error: %v", err)
			return
		}
		h.sendTo(conn, Message{Type: "heartbeat_ack", Payload: ack})

	case "traffic":
		var tr TrafficReport
		if err := json.Unmarshal(msg.Payload, &tr); err != nil {
			return
		}
		if h.OnTraffic != nil {
			h.OnTraffic(conn, tr)
		}

	case "speed":
		var sr SpeedReport
		if err := json.Unmarshal(msg.Payload, &sr); err != nil {
			return
		}
		if h.OnSpeed != nil {
			h.OnSpeed(conn, sr)
		}

	case "ping":
		h.sendTo(conn, Message{Type: "pong", Payload: json.RawMessage(`{}`)})

	default:
		log.Printf("[Master WS] Unknown message type: %s", msg.Type)
	}
}

// sendAuthResult 发送认证结果
func (h *Handler) sendAuthResult(conn *AgentConnection, success bool, message string) {
	payload, err := json.Marshal(map[string]interface{}{
		"success": success,
		"message": message,
	})
	if err != nil {
		log.Printf("[Master WS] auth result marshal error: %v", err)
		return
	}
	h.sendTo(conn, Message{
		Type:    "auth_result",
		Payload: payload,
	})
}

// SendConfigUpdate 向指定 Agent 推送配置更新
func (h *Handler) SendConfigUpdate(token string, config interface{}) error {
	payload, err := json.Marshal(config)
	if err != nil {
		return err
	}

	connI, ok := h.conns.Load(token)
	if !ok {
		return nil // Agent 不在线，静默跳过
	}

	conn := connI.(*AgentConnection)
	return h.sendTo(conn, Message{
		Type:    "config_update",
		Payload: payload,
	})
}

// BroadcastConfigUpdate 向所有在线 Agent 广播配置更新
func (h *Handler) BroadcastConfigUpdate(updates map[string]string) {
	payload, err := json.Marshal(updates)
	if err != nil {
		log.Printf("[Master WS] BroadcastConfigUpdate marshal failed: %v", err)
		return
	}
	msg := Message{Type: "config_update", Payload: payload}

	h.conns.Range(func(_, v interface{}) bool {
		conn := v.(*AgentConnection)
		h.sendTo(conn, msg)
		return true
	})
}

// GetConnection 获取 Agent 连接
func (h *Handler) GetConnection(token string) (*AgentConnection, bool) {
	v, ok := h.conns.Load(token)
	if !ok {
		return nil, false
	}
	return v.(*AgentConnection), true
}

// GetConnectedTokens 获取所有在线 Agent 的 token
func (h *Handler) GetConnectedTokens() []string {
	var tokens []string
	h.conns.Range(func(k, _ interface{}) bool {
		tokens = append(tokens, k.(string))
		return true
	})
	return tokens
}

// CleanStaleConnections 清理超时连接
func (h *Handler) CleanStaleConnections(timeout time.Duration) {
	cutoff := time.Now().Add(-timeout)

	h.conns.Range(func(key, value interface{}) bool {
		conn := value.(*AgentConnection)
		conn.mu.Lock()
		lastPing := conn.LastPing
		conn.mu.Unlock()

		if lastPing.Before(cutoff) {
			log.Printf("[Master WS] Closing stale connection: %s", conn.ServerName)
			conn.writeMu.Lock()
			conn.Conn.Close()
			conn.writeMu.Unlock()
			// CompareAndDelete 防重连竞态
			h.conns.CompareAndDelete(key, conn)
		}
		return true
	})
}

func (h *Handler) sendTo(conn *AgentConnection, msg Message) error {
	conn.writeMu.Lock()
	defer conn.writeMu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.Conn.WriteMessage(websocket.TextMessage, data)
}

// originHostMatches 检查 Origin URL 的 host:port 是否匹配请求的 Host
func originHostMatches(origin, host string) bool {
	// 提取 Origin URL 中的 host:port
	o := origin
	if len(o) > 8 && o[:8] == "https://" {
		o = o[8:]
	} else if len(o) > 7 && o[:7] == "http://" {
		o = o[7:]
	}
	// 去掉路径
	for i := 0; i < len(o); i++ {
		if o[i] == '/' {
			o = o[:i]
			break
		}
	}
	// 精确匹配
	return o == host
}
