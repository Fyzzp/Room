// Package grpc 提供主控端 gRPC 客户端
// 主控通过 gRPC 调用远程 Agent 的 AgentService
package grpc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient 封装与单个 Agent 的 gRPC 连接
type AgentClient struct {
	ServerID  int64
	Address   string
	Token     string
	conn      *grpc.ClientConn
	mu        sync.Mutex
	connected bool
}

// ClientManager 管理所有 Agent 的 gRPC 连接
type ClientManager struct {
	mu      sync.RWMutex
	clients map[int64]*AgentClient // serverID -> client
}

// NewClientManager 创建客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[int64]*AgentClient),
	}
}

// Connect 连接到 Agent
func (m *ClientManager) Connect(serverID int64, address, token string) (*AgentClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已有连接，先关闭
	if existing, ok := m.clients[serverID]; ok {
		existing.Close()
	}

	client := &AgentClient{
		ServerID: serverID,
		Address:  address,
		Token:    token,
	}

	if err := client.dial(); err != nil {
		return nil, err
	}

	m.clients[serverID] = client
	return client, nil
}

// Get 获取已连接的 Agent 客户端
func (m *ClientManager) Get(serverID int64) (*AgentClient, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[serverID]
	return c, ok
}

// Remove 移除 Agent 连接
func (m *ClientManager) Remove(serverID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.clients[serverID]; ok {
		c.Close()
		delete(m.clients, serverID)
	}
}

// CloseAll 关闭所有连接
func (m *ClientManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, c := range m.clients {
		c.Close()
		delete(m.clients, id)
	}
}

// dial 建立 gRPC 连接
func (c *AgentClient) dial() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("dial agent %s: %w", c.Address, err)
	}

	c.conn = conn
	c.connected = true
	log.Printf("[gRPC] Connected to agent %d at %s", c.ServerID, c.Address)
	return nil
}

// Close 关闭连接
func (c *AgentClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.connected = false
	}
}

// IsConnected 检查连接状态
func (c *AgentClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// Conn 返回底层 gRPC 连接（用于创建 Service Client）
func (c *AgentClient) Conn() *grpc.ClientConn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}

// HealthCheck 调用 Agent 的健康检查
func (c *AgentClient) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	// TODO: 等 proto 生成后使用生成的客户端
	// client := pb.NewAgentServiceClient(c.Conn())
	// return client.HealthCheck(ctx, &pb.HealthCheckRequest{})
	return map[string]interface{}{
		"ok": c.IsConnected(),
	}, nil
}

// AddInbound 通过 gRPC 向 Agent 添加 Xray 入站
func (c *AgentClient) AddInbound(ctx context.Context, inboundJSON string) error {
	// TODO: 实现 protobuf 序列化后调用
	log.Printf("[gRPC] AddInbound to agent %d", c.ServerID)
	return nil
}

// RemoveInbound 通过 gRPC 删除 Xray 入站
func (c *AgentClient) RemoveInbound(ctx context.Context, tag string) error {
	log.Printf("[gRPC] RemoveInbound %s from agent %d", tag, c.ServerID)
	return nil
}
