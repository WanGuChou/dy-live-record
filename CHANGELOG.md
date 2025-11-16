# Changelog

所有项目的重要变更都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

---

## [3.1.0] - 2025-11-15 - 🎉 完整版发布

### Added
- **依赖自动检查** (`server-go/internal/dependencies/`)
  - WebView2 Runtime 自动检测和安装
  - SQLite 驱动 (CGO) 检测
  - 网络连接检测
  - 磁盘空间检测
  
- **分段记分功能** (`server-go/internal/database/segments.go`)
  - `score_segments` 数据库表
  - 创建/结束分段 API
  - 分段统计（礼物总值、消息数）
  - UI 界面（新增「分段记分」标签页）

- **WebView2 Fallback 数据通道** (`server-go/internal/fallback/`)
  - 隐藏 WebView2 窗口
  - JavaScript 注入拦截 WebSocket
  - Base64 编码传输
  - 心跳检测机制

- **浏览器插件管理** (`server-go/internal/ui/settings.go`)
  - 插件打包脚本 (`browser-monitor/pack.bat|sh`)
  - 内嵌式部署（embed.FS）
  - 一键安装功能

- **管理后台 UI** (`server-active/web/admin.html`)
  - 生成新许可证表单
  - 许可证列表显示
  - 查看/撤销操作
  - 新增 `GET /api/v1/licenses/list` API

- **构建脚本**
  - `BUILD_ALL.bat` - Windows 全量构建
  - `BUILD_ALL.sh` - Linux/Mac 全量构建
  
- **文档**
  - `QUICK_START.md` - 快速开始指南
  - `COMPLETION_REPORT.md` - 完整功能报告
  - `RELEASE_NOTES.md` - 发布说明
  - `CHANGELOG.md` - 变更日志

### Changed
- `server-go/main.go` - 集成依赖检查器
- `server-go/internal/ui/webview.go` - 新增分段记分 UI
- `server-active/internal/api/routes.go` - 添加许可证列表 API
- `server-active/internal/license/manager.go` - 新增 `ListAllLicenses()` 方法

---

## [3.0.0] - 2025-11-15 - 🚀 Go 重构版

### Added
- **完整的 Protobuf 解析器** (`server-go/internal/parser/`)
  - ByteBuffer 实现
  - 所有 Douyin 消息类型解码
  - GZIP 解压缩支持

- **WebView2 主界面** (`server-go/internal/ui/webview.go`)
  - 多房间标签页
  - 数据概览看板
  - 礼物/消息记录表
  - 主播管理界面

- **主播管理与礼物分配** (`server-go/internal/server/gift_allocation.go`)
  - 礼物自动绑定
  - 消息内容解析（@主播名识别）
  - 主播业绩记录

- **SQLite 数据持久化** (`server-go/internal/database/`)
  - 房间信息、直播场次、礼物记录、消息记录、主播配置

- **许可证系统** (`server-go/internal/license/`)
  - RSA 2048 客户端校验
  - 硬件指纹采集（Windows）
  - NTP 时间同步

- **浏览器插件适配** (`browser-monitor/`)
  - 离线数据缓存 (`chrome.storage.local`)
  - 心跳机制（30 秒）
  - 自动重推机制

- **许可证授权服务** (`server-active/`)
  - MySQL 数据库
  - RESTful API（生成、校验、转移、撤销）
  - RSA 2048 私钥签名

### Changed
- 从 Node.js 重构为 Go 语言
- 从内存存储改为 SQLite 持久化
- 新增系统托盘 UI

---

## [2.2.0] - 2025-11-08 - Protobuf 解析器

### Added
- **抖音 WebSocket 消息解析** (`server/dy_ws_msg.js`)
  - 完整的 Protobuf 解析器（移植自 skmcj/dycast）
  - ByteBuffer 实现
  - GZIP 解压缩（使用 pako）
  - 所有消息类型解码

- **WebSocket 日志** (`server/`)
  - 按日期和小时分组
  - 自动管理日志文件

### Fixed
- 修复 Protobuf wire type 3/4 处理
- 修复 GiftMessage/User 字段编号错误
- 修复嵌套结构解析

---

## [2.0.0] - 2025-11-01 - CDP 深度监控

### Added
- **Chrome DevTools Protocol 集成** (`brower-monitor/`)
  - 浏览器请求监控
  - WebSocket 消息拦截
  - 完整生命周期捕获

- **Node.js WebSocket 服务器** (`server/`)
  - 接收浏览器插件数据
  - 消息格式化输出

### Changed
- 从简单 HTTP 监控升级为 CDP 深度监控

---

## [1.0.0] - 2025-10-25 - 初始版本

### Added
- 基础浏览器扩展（Manifest V3）
- 简单的 HTTP 请求监控
- WebSocket 连接检测

---

## 图例

- `Added` - 新增功能
- `Changed` - 功能变更
- `Deprecated` - 即将废弃的功能
- `Removed` - 已移除的功能
- `Fixed` - 错误修复
- `Security` - 安全修复

---

**维护者**: AI Assistant (Claude Sonnet 4.5)  
**最后更新**: 2025-11-15
