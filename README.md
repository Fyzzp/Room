# Room — Xray 多机管理主控面板

Room 是一个基于 Xray-core 的多服务器管理面板，采用主控+Agent 架构，统一管理分布在不同 VPS 上的 Xray 服务。

## 功能

- 🖥️ **多服务器管理** — 一个面板管理所有远程 Xray 节点
- 📊 **实时流量** — 各服务器流量统计与速度监控
- 🔧 **远程配置** — 在线管理入站/出站/路由规则
- 🔐 **证书管理** — ACME 自动申请/续期 SSL 证书
- 📦 **套餐管理** — 用户套餐与流量限额
- 🔄 **节点同步** — 入站变更自动同步到订阅节点
- 📡 **订阅管理** — 支持 Clash/Surge/Shadowrocket 等多种客户端

## 一键安装

```bash
curl -fsSL https://raw.githubusercontent.com/Fyzzp/Room/main/scripts/install.sh | sudo bash
```

安装完成后访问 `http://服务器IP:12889`。

## 技术栈

- 后端：Go + PostgreSQL + Redis
- 前端：React 19 + Vite + TailwindCSS 4 + shadcn/ui
- 通信：gRPC + WebSocket（加密 + 自动回退）
- 部署：Docker Compose 一键部署

## 子节点部署

在主控面板添加服务器后，会生成一键安装命令，在远程 VPS 执行即可自动安装 [Room-Agent](https://github.com/Fyzzp/Room-Agent) 并接入主控。

## 开发

```bash
# 启动基础设施
docker compose up -d postgres redis

# 运行主控
go run ./cmd -c config.dev.json

# 构建前端
cd web && npm install && npm run dev
```

## 许可证

MIT License
