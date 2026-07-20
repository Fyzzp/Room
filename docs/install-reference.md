# 妙妙屋X install.sh 研究笔记

## 部署模式
- systemd 服务管理（无 Docker）
- 单二进制文件部署
- 直接下载 GitHub Releases 二进制
- SQLite 数据库（零外部依赖）

## 支持的四种模式
- install — 全新安装
- update — 更新二进制 + 备份 + 启动失败自动回滚
- uninstall — 卸载 + 可选保留数据
- reinstall — 覆盖安装保留数据

## 关键设计
- 架构自动检测（amd64/arm64）
- 依赖安装（curl/jq）
- 交互式/非交互式兼容（检测 stdin 是否 tty）
- 版本号持久化到 .version 文件
- 端口号可配置（交互式询问或环境变量）
- systemd 环境变量注入
- 二进制下载失败自动回退
