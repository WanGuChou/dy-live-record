# 抖音直播监控系统 - 项目路线图

## 📋 项目概览

本项目是一套完整的抖音直播数据采集、分析和管理系统，采用C/S架构，包含三个核心组件：

1. **`server-go`** - Go语言桌面客户端（核心后端）
2. **`browser-monitor`** - 浏览器监控插件
3. **`server-active`** - Go语言许可证授权服务

---

## 🎯 当前状态

**阶段 1: 基础框架搭建 - 已完成 ✅**

### 已完成工作

#### ✅ 项目结构重组
- 创建 `server-go/` 目录（Go 核心服务）
- 创建 `server-active/` 目录（许可证服务）
- 重命名 `dy-live-record/brower-monitor/` → `browser-monitor/`

#### ✅ server-go 基础模块
- **配置管理** (`internal/config/`)
  - JSON 配置文件
  - 默认配置生成
  - 端口、数据库、许可证等配置项

- **数据库** (`internal/database/`)
  - SQLite 自动初始化
  - 完整表结构设计（rooms, live_sessions, gift_records, message_records, anchors）
  - 索引优化

- **许可证系统** (`internal/license/`)
  - RSA 2048 数字签名
  - 硬件指纹生成（CPU, 主板, 硬盘, MAC）
  - NTP 时间校验
  - 离线验证 + 在线激活

- **WebSocket 服务器** (`internal/server/`)
  - 接收插件数据
  - 多房间管理
  - 消息分类处理
  - 数据持久化框架

- **系统托盘 UI** (`internal/ui/`)
  - 托盘图标和菜单
  - 主界面、设置、许可证入口（框架）

---

## 🚀 下一步计划

### 阶段 2: 核心功能实现（当前阶段）

#### ⭐⭐⭐⭐⭐ 优先级 1: Protobuf 解析器移植

**任务**: 将 `server/dy_ws_msg.js` 完整移植到 Go  
**文件**: `server-go/internal/parser/douyin.go`  
**工作量**: 大（预计 2-3 天）

**子任务**:
1. [ ] ByteBuffer 实现（Go 版本）
2. [ ] `decodePushFrame()` - PushFrame 解析
3. [ ] `decodeResponse()` - Response 解析
4. [ ] `decodeMessage()` - Message 解析
5. [ ] `decodeUser()` - User 结构（包含 80+ 字段）
6. [ ] `decodeChatMessage()` - 聊天消息
7. [ ] `decodeGiftMessage()` - 礼物消息
8. [ ] `decodeGiftStruct()` - 礼物详情
9. [ ] `decodeLikeMessage()` - 点赞消息
10. [ ] `decodeMemberMessage()` - 进入直播间
11. [ ] GZIP 解压（`compress/gzip`）
12. [ ] 消息格式化输出

**技术要点**:
- JavaScript `Uint8Array` → Go `[]byte`
- Protobuf varint 编码/解码
- 嵌套结构递归解析
- Wire type 处理（0-5）

**参考资源**:
- `server/dy_ws_msg.js` (现有实现)
- `DOUYIN_USER_FIX.md` (User 结构详解)
- `DOUYIN_FIELD_FIX.md` (字段编号修复)

---

#### ⭐⭐⭐⭐ 优先级 2: 主界面开发

**任务**: 使用 webview2 创建数据看板  
**文件**: `server-go/internal/ui/main_window.go`  
**工作量**: 中（预计 2 天）

**子任务**:
1. [ ] 创建 webview2 窗口
2. [ ] HTML/CSS/JS 前端页面
3. [ ] Tab 标签页（多房间切换）
4. [ ] 实时数据统计卡片
5. [ ] 礼物记录表格
6. [ ] 消息记录列表
7. [ ] 历史记录查询
8. [ ] Go ↔ JS 双向通信

---

#### ⭐⭐⭐ 优先级 3: 插件调整

**任务**: 适配新的 server-go 后端  
**文件**: `browser-monitor/background.js`  
**工作量**: 小（预计 1 天）

**子任务**:
1. [ ] 修改服务器地址配置（默认 `localhost:8080`）
2. [ ] 添加离线缓存逻辑（`chrome.storage.local`）
3. [ ] 数据重推机制（检测服务器恢复后推送缓存）
4. [ ] 心跳检测（定期发送心跳包）

---

### 阶段 3: 高级功能

#### ⭐⭐⭐ Webview2 备用数据通道

**任务**: 实现 Fallback 机制  
**工作量**: 中（预计 2 天）

**子任务**:
1. [ ] 后台启动 webview2 实例
2. [ ] 注入 JavaScript 脚本到 `live.douyin.com`
3. [ ] 拦截 WSS 消息
4. [ ] 解析并注入主数据流
5. [ ] 心跳检测触发机制（10秒无数据）

---

#### ⭐⭐ 主播管理与礼物分配

**任务**: 主播配置和业绩计算  
**工作量**: 中（预计 1-2 天）

**子任务**:
1. [ ] 主播增删改查 UI
2. [ ] 礼物绑定规则配置
3. [ ] 弹幕指令识别（如 "@主播A 刷火箭"）
4. [ ] 自动计算业绩
5. [ ] 业绩报表生成

---

#### ⭐ 数据刷新与分段记分

**任务**: 数据手动刷新和时段划分  
**工作量**: 小（预计 0.5 天）

**子任务**:
1. [ ] 手动刷新钻石/积分按钮
2. [ ] 分段记分（快照当前统计）
3. [ ] PK 时段标记

---

### 阶段 4: server-active 许可证服务

#### ⭐⭐⭐⭐ 许可证生成与管理

**任务**: 创建许可证授权服务器  
**文件**: `server-active/`  
**工作量**: 大（预计 3 天）

**子任务**:
1. [ ] MySQL 数据库设计
   - `licenses` 表
   - `activation_records` 表
   - `customers` 表

2. [ ] API 接口
   - `POST /api/v1/licenses/generate` - 生成许可证
   - `POST /api/v1/licenses/validate` - 校验许可证
   - `POST /api/v1/licenses/transfer` - 转移许可证
   - `GET /api/v1/licenses/:key` - 查询许可证

3. [ ] 许可证生成逻辑
   - RSA 私钥签名
   - 硬件指纹绑定
   - Base64 编码

4. [ ] 管理后台（可选）
   - Web 界面
   - 许可证列表
   - 激活记录查询

---

### 阶段 5: 打包与部署

#### ⭐⭐⭐ Windows 安装程序

**任务**: 创建安装包  
**工具**: Inno Setup 或 WiX Toolset  
**工作量**: 中（预计 1 天）

**子任务**:
1. [ ] 创建安装脚本
2. [ ] 嵌入 WebView2 Runtime
3. [ ] 安装插件提示
4. [ ] 桌面快捷方式
5. [ ] 开机自启动选项
6. [ ] 卸载程序

---

## 📅 时间估算

| 阶段 | 任务 | 工作量 | 预计完成 |
|------|------|--------|----------|
| **阶段 1** | 基础框架 | 已完成 | ✅ |
| **阶段 2** | Protobuf 解析器 | 2-3 天 | 🚧 |
| **阶段 2** | 主界面开发 | 2 天 | ⏳ |
| **阶段 2** | 插件调整 | 1 天 | ⏳ |
| **阶段 3** | Webview2 Fallback | 2 天 | ⏳ |
| **阶段 3** | 主播管理 | 1-2 天 | ⏳ |
| **阶段 3** | 分段记分 | 0.5 天 | ⏳ |
| **阶段 4** | server-active | 3 天 | ⏳ |
| **阶段 5** | 打包部署 | 1 天 | ⏳ |
| **总计** | | **12-14 天** | |

---

## 🔧 技术栈汇总

### server-go
- **语言**: Go 1.21+
- **UI**: systray (托盘) + webview2 (主界面)
- **数据库**: SQLite3 (`mattn/go-sqlite3`)
- **WebSocket**: `gorilla/websocket`
- **加密**: RSA 2048, SHA-256

### browser-monitor
- **类型**: Chrome/Edge 插件 (Manifest V3)
- **核心API**: `chrome.debugger` (CDP)
- **通信**: WebSocket Client

### server-active
- **语言**: Go 1.21+
- **数据库**: MySQL 8.0+
- **Web框架**: Gin / Echo
- **加密**: RSA 2048

---

## 📝 文档清单

### 已完成文档
- ✅ `server-go/README.md` - Go 服务器说明
- ✅ `PROJECT_ROADMAP.md` - 本文档

### 待创建文档
- [ ] `server-active/README.md` - 许可证服务说明
- [ ] `browser-monitor/README.md` - 插件更新说明
- [ ] `DEPLOYMENT.md` - 部署指南
- [ ] `USER_MANUAL.md` - 用户手册
- [ ] `DEVELOPER_GUIDE.md` - 开发者指南

---

## 🎯 里程碑

1. **M1: 基础框架** ✅ (已完成)
   - 项目结构
   - 核心模块框架

2. **M2: 数据采集** 🚧 (进行中)
   - Protobuf 解析器
   - 插件适配

3. **M3: 数据展示** ⏳
   - 主界面
   - 数据看板

4. **M4: 完整功能** ⏳
   - 主播管理
   - 分段记分
   - Fallback 机制

5. **M5: 许可证系统** ⏳
   - server-active
   - 激活流程

6. **M6: 正式发布** ⏳
   - 打包
   - 文档
   - 用户测试

---

## 💡 下一步行动

### 立即开始（阶段 2）

1. **Protobuf 解析器移植** (最高优先级)
   - 创建 `server-go/internal/parser/bytebuffer.go`
   - 移植 `decodePushFrame`, `decodeResponse`, `decodeMessage`
   - 单元测试

2. **插件适配**
   - 修改 `browser-monitor/background.js`
   - 测试与 server-go 的连接

3. **简单主界面**
   - 创建 HTML 模板
   - 显示实时数据

---

**🚀 Let's Build It!**
