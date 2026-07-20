// Package xconfig 提供 Xray-core 配置的生成与管理
// 通过 gRPC 调用 Xray-core 的 HandlerService API 实现热更新
package xconfig

import (
	"encoding/json"
	"fmt"
)

// XrayConfig 表示一份完整的 Xray 配置文件
type XrayConfig struct {
	Log       *LogConfig       `json:"log,omitempty"`
	API       *APIConfig       `json:"api,omitempty"`
	DNS       *DNSConfig       `json:"dns,omitempty"`
	Routing   *RoutingConfig   `json:"routing,omitempty"`
	Policy    *PolicyConfig    `json:"policy,omitempty"`
	Inbounds  []InboundConfig  `json:"inbounds,omitempty"`
	Outbounds []OutboundConfig `json:"outbounds,omitempty"`
	Stats     *StatsConfig     `json:"stats,omitempty"`
}

// LogConfig 日志配置
type LogConfig struct {
	LogLevel string `json:"loglevel,omitempty"` // debug, info, warning, error, none
}

// APIConfig gRPC API 配置
type APIConfig struct {
	Tag      string   `json:"tag,omitempty"`
	Listen   string   `json:"listen,omitempty"` // 127.0.0.1:38889
	Services []string `json:"services,omitempty"`
}

// DNSConfig 内置 DNS 配置
type DNSConfig struct {
	Servers []DNSServer `json:"servers,omitempty"`
}

// DNSServer DNS 服务器
type DNSServer struct {
	Address      string   `json:"address,omitempty"`
	Domains      []string `json:"domains,omitempty"`
	QueryStrategy string  `json:"queryStrategy,omitempty"`
}

// RoutingConfig 路由配置
type RoutingConfig struct {
	DomainStrategy string          `json:"domainStrategy,omitempty"`
	Rules          []RoutingRule   `json:"rules,omitempty"`
	Balancers      []BalancerConfig `json:"balancers,omitempty"`
}

// RoutingRule 路由规则
type RoutingRule struct {
	Type        string   `json:"type,omitempty"` // field
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	Network     string   `json:"network,omitempty"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
	User        []string `json:"user,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
}

// BalancerConfig 负载均衡器
type BalancerConfig struct {
	Tag          string   `json:"tag"`
	Selector     []string `json:"selector"`
	FallbackTag  string   `json:"fallbackTag,omitempty"`
	StrategyType string   `json:"strategyType,omitempty"` // random, roundRobin, leastPing, leastLoad
}

// PolicyConfig 策略配置
type PolicyConfig struct {
	Levels map[int]PolicyLevel `json:"levels,omitempty"`
}

// PolicyLevel 用户等级策略
type PolicyLevel struct {
	StatsUserUplink   bool `json:"statsUserUplink,omitempty"`
	StatsUserDownlink bool `json:"statsUserDownlink,omitempty"`
	Handshake         int  `json:"handshake,omitempty"`
	ConnectionIdle    int  `json:"connectionIdle,omitempty"`
}

// InboundConfig 入站配置（对应 Xray InboundObject）
type InboundConfig struct {
	Listen         string                   `json:"listen,omitempty"`
	Port           interface{}              `json:"port"` // number or string
	Protocol       string                   `json:"protocol"`
	Settings       json.RawMessage          `json:"settings,omitempty"`
	StreamSettings *StreamSettingsConfig    `json:"streamSettings,omitempty"`
	Tag            string                   `json:"tag"`
	Sniffing       *SniffingConfig          `json:"sniffing,omitempty"`
}

// OutboundConfig 出站配置
type OutboundConfig struct {
	SendThrough    string                `json:"sendThrough,omitempty"`
	Protocol       string                `json:"protocol"`
	Settings       json.RawMessage       `json:"settings,omitempty"`
	Tag            string                `json:"tag"`
	StreamSettings *StreamSettingsConfig  `json:"streamSettings,omitempty"`
	ProxySettings  *ProxySettingsConfig   `json:"proxySettings,omitempty"`
	Mux            *MuxConfig            `json:"mux,omitempty"`
}

// StreamSettingsConfig 传输配置
type StreamSettingsConfig struct {
	Method        string                `json:"method,omitempty"` // raw, ws, grpc, mkcp, httpupgrade, xhttp, hysteria
	Security      string                `json:"security,omitempty"` // none, tls, reality
	WSSettings    *WSSettingsConfig     `json:"wsSettings,omitempty"`
	GRPCSettings  *GRPCSettingsConfig   `json:"grpcSettings,omitempty"`
	TLSSettings   *TLSSettingsConfig    `json:"tlsSettings,omitempty"`
	RealitySettings *RealitySettingsConfig `json:"realitySettings,omitempty"`
	Sockopt       *SockoptConfig        `json:"sockopt,omitempty"`
}

// SniffingConfig 嗅探配置
type SniffingConfig struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride,omitempty"`
	RouteOnly    bool     `json:"routeOnly,omitempty"`
}

// WSSettingsConfig WebSocket 配置
type WSSettingsConfig struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// GRPCSettingsConfig gRPC 配置
type GRPCSettingsConfig struct {
	ServiceName string `json:"serviceName,omitempty"`
	MultiMode   bool   `json:"multiMode,omitempty"`
}

// TLSSettingsConfig TLS 配置
type TLSSettingsConfig struct {
	ServerName    string        `json:"serverName,omitempty"`
	ALPN          []string      `json:"alpn,omitempty"`
	Fingerprint   string        `json:"fingerprint,omitempty"`
	Certificates  []Certificate `json:"certificates,omitempty"`
}

// Certificate TLS 证书
type Certificate struct {
	CertificateFile string   `json:"certificateFile,omitempty"`
	KeyFile         string   `json:"keyFile,omitempty"`
	Certificate     []string `json:"certificate,omitempty"`
	Key             []string `json:"key,omitempty"`
}

// RealitySettingsConfig REALITY 配置
type RealitySettingsConfig struct {
	Show        bool     `json:"show,omitempty"`
	Target      string   `json:"target"` // 回退目标: "example.com:443"
	ServerNames []string `json:"serverNames"`
	PrivateKey  string   `json:"privateKey"`
	ShortIds    []string `json:"shortIds"`
	// 客户端字段
	ServerName  string `json:"serverName,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Password    string `json:"password,omitempty"` // 公钥
	ShortId     string `json:"shortId,omitempty"`
	SpiderX     string `json:"spiderX,omitempty"`
}

// SockoptConfig 底层网络配置
type SockoptConfig struct {
	DomainStrategy string `json:"domainStrategy,omitempty"`
	TCPKeepAlive   bool   `json:"tcpKeepAlive,omitempty"`
}

// ProxySettingsConfig 链式代理
type ProxySettingsConfig struct {
	Tag            string `json:"tag,omitempty"`
	TransportLayer bool   `json:"transportLayer,omitempty"`
}

// MuxConfig 多路复用
type MuxConfig struct {
	Enabled         bool   `json:"enabled"`
	Concurrency     int    `json:"concurrency,omitempty"`
	XUDPConcurrency int    `json:"xudpConcurrency,omitempty"`
	XUDPProxyUDP443 string `json:"xudpProxyUDP443,omitempty"`
}

// StatsConfig 统计配置
type StatsConfig struct{}

// ========================
// Builder - Xray 配置构建器
// ========================

// ConfigBuilder 构建 Xray 配置
type ConfigBuilder struct {
	config XrayConfig
}

// NewConfigBuilder 创建新的配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: XrayConfig{
			Log:     &LogConfig{LogLevel: "warning"},
			Stats:   &StatsConfig{},
			Routing: &RoutingConfig{}, // 避免 nil pointer dereference
			API: &APIConfig{
				Tag:      "api",
				Listen:   "127.0.0.1:38889",
				Services: []string{"HandlerService", "StatsService", "RoutingService"},
			},
			Policy: &PolicyConfig{
				Levels: map[int]PolicyLevel{
					0: {StatsUserUplink: true, StatsUserDownlink: true, Handshake: 4, ConnectionIdle: 300},
					1: {StatsUserUplink: true, StatsUserDownlink: true, Handshake: 4, ConnectionIdle: 300},
				},
			},
		},
	}
}

// AddVLESSInbound 添加 VLESS 入站
func (b *ConfigBuilder) AddVLESSInbound(tag string, port uint32, users []VLESSUser, stream *StreamSettingsConfig, sniffing bool) *ConfigBuilder {
	settings, _ := json.Marshal(map[string]interface{}{
		"clients":    users,
		"decryption": "none",
	})
	inbound := InboundConfig{
		Listen:         "0.0.0.0",
		Port:           port,
		Protocol:       "vless",
		Settings:       settings,
		StreamSettings: stream,
		Tag:            tag,
	}
	if sniffing {
		inbound.Sniffing = &SniffingConfig{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		}
	}
	b.config.Inbounds = append(b.config.Inbounds, inbound)
	return b
}

// AddVMessInbound 添加 VMess 入站
func (b *ConfigBuilder) AddVMessInbound(tag string, port uint32, users []VMessUser, stream *StreamSettingsConfig, sniffing bool) *ConfigBuilder {
	settings, _ := json.Marshal(map[string]interface{}{
		"users": users,
	})
	inbound := InboundConfig{
		Listen:         "0.0.0.0",
		Port:           port,
		Protocol:       "vmess",
		Settings:       settings,
		StreamSettings: stream,
		Tag:            tag,
	}
	if sniffing {
		inbound.Sniffing = &SniffingConfig{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		}
	}
	b.config.Inbounds = append(b.config.Inbounds, inbound)
	return b
}

// AddTrojanInbound 添加 Trojan 入站
func (b *ConfigBuilder) AddTrojanInbound(tag string, port uint32, passwords []string, stream *StreamSettingsConfig, sniffing bool) *ConfigBuilder {
	users := make([]map[string]interface{}, len(passwords))
	for i, pwd := range passwords {
		users[i] = map[string]interface{}{
			"password": pwd,
			"level":    0,
			"email":    fmt.Sprintf("user%d@trojan.local", i),
		}
	}
	settings, _ := json.Marshal(map[string]interface{}{
		"users": users,
	})
	inbound := InboundConfig{
		Listen:         "0.0.0.0",
		Port:           port,
		Protocol:       "trojan",
		Settings:       settings,
		StreamSettings: stream,
		Tag:            tag,
	}
	if sniffing {
		inbound.Sniffing = &SniffingConfig{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		}
	}
	b.config.Inbounds = append(b.config.Inbounds, inbound)
	return b
}

// AddShadowsocksInbound 添加 Shadowsocks 入站
func (b *ConfigBuilder) AddShadowsocksInbound(tag string, port uint32, method, password string) *ConfigBuilder {
	settings, _ := json.Marshal(map[string]interface{}{
		"method":   method,
		"password": password,
		"network":  "tcp,udp",
	})
	inbound := InboundConfig{
		Listen:   "0.0.0.0",
		Port:     port,
		Protocol: "shadowsocks",
		Settings: settings,
		Tag:      tag,
	}
	b.config.Inbounds = append(b.config.Inbounds, inbound)
	return b
}

// AddFreedomOutbound 添加直连出站
func (b *ConfigBuilder) AddFreedomOutbound(tag string) *ConfigBuilder {
	settings, _ := json.Marshal(map[string]interface{}{
		"domainStrategy": "AsIs",
	})
	outbound := OutboundConfig{
		Protocol: "freedom",
		Settings: settings,
		Tag:      tag,
	}
	b.config.Outbounds = append(b.config.Outbounds, outbound)
	return b
}

// AddBlackholeOutbound 添加黑洞出站
func (b *ConfigBuilder) AddBlackholeOutbound(tag string) *ConfigBuilder {
	settings, _ := json.Marshal(map[string]interface{}{})
	outbound := OutboundConfig{
		Protocol: "blackhole",
		Settings: settings,
		Tag:      tag,
	}
	b.config.Outbounds = append(b.config.Outbounds, outbound)
	return b
}

// AddDirectRule 添加直连规则
func (b *ConfigBuilder) AddDirectRule(domain []string, outboundTag string) *ConfigBuilder {
	b.config.Routing.Rules = append(b.config.Routing.Rules, RoutingRule{
		Type:        "field",
		Domain:      domain,
		OutboundTag: outboundTag,
	})
	return b
}

// Build 返回最终的 Xray 配置
func (b *ConfigBuilder) Build() XrayConfig {
	return b.config
}

// ToJSON 将配置序列化为 JSON
func (b *ConfigBuilder) ToJSON() ([]byte, error) {
	return json.MarshalIndent(b.config, "", "  ")
}

// ========================
// 用户类型
// ========================

// VLESSUser VLESS 用户
type VLESSUser struct {
	ID    string `json:"id"`
	Email string `json:"email,omitempty"`
	Level int    `json:"level,omitempty"`
	Flow  string `json:"flow,omitempty"` // xtls-rprx-vision
}

// VMessUser VMess 用户
type VMessUser struct {
	ID       string `json:"id"`
	Email    string `json:"email,omitempty"`
	Level    int    `json:"level,omitempty"`
	Security string `json:"security,omitempty"` // auto, aes-128-gcm, chacha20-poly1305, none
}

// ========================
// 常用配置模板
// ========================

// RealityStream 创建 REALITY 传输配置（服务端）
func RealityStream(target string, serverNames []string, privateKey string, shortIds []string) *StreamSettingsConfig {
	return &StreamSettingsConfig{
		Method:   "raw",
		Security: "reality",
		RealitySettings: &RealitySettingsConfig{
			Target:      target,
			ServerNames: serverNames,
			PrivateKey:  privateKey,
			ShortIds:    shortIds,
		},
	}
}

// RealityClientStream 创建 REALITY 传输配置（客户端）
func RealityClientStream(serverName, fingerprint, publicKey, shortId, spiderX string) *StreamSettingsConfig {
	return &StreamSettingsConfig{
		Method:   "raw",
		Security: "reality",
		RealitySettings: &RealitySettingsConfig{
			ServerName:  serverName,
			Fingerprint: fingerprint,
			Password:    publicKey,
			ShortId:     shortId,
			SpiderX:     spiderX,
		},
	}
}

// TLSStream 创建 TLS 传输配置
func TLSStream(serverName string, certFile, keyFile string) *StreamSettingsConfig {
	return &StreamSettingsConfig{
		Method:   "raw",
		Security: "tls",
		TLSSettings: &TLSSettingsConfig{
			ServerName: serverName,
			ALPN:       []string{"h2", "http/1.1"},
			Certificates: []Certificate{
				{CertificateFile: certFile, KeyFile: keyFile},
			},
		},
	}
}

// WebSocketStream 创建 WebSocket 传输配置
func WebSocketStream(path string) *StreamSettingsConfig {
	return &StreamSettingsConfig{
		Method: "websocket",
		WSSettings: &WSSettingsConfig{
			Path: path,
		},
	}
}

// GRPCStream 创建 gRPC 传输配置
func GRPCStream(serviceName string, multiMode bool) *StreamSettingsConfig {
	return &StreamSettingsConfig{
		Method: "grpc",
		GRPCSettings: &GRPCSettingsConfig{
			ServiceName: serviceName,
			MultiMode:   multiMode,
		},
	}
}
