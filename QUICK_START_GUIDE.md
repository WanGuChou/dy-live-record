# 快速开始指南

## ✅ 完成的修改

### 1. WebSocket 路径更新
- **旧路径**: `ws://localhost:8080/ws`
- **新路径**: `ws://localhost:8080/monitor` ✅

### 2. 动态房间 Tab 功能
- ✅ 自动检测直播间房间号
- ✅ 自动创建房间 Tab（如：`🏠 房间 46387032209`）
- ✅ 实时显示原始 WebSocket 消息
- ✅ 实时显示解析后的结构化消息
- ✅ 50/50 分割视图

### 3. 消息解析改进
- ✅ 参考 dycast 项目的 Protobuf 解析逻辑
- ✅ ByteBuffer 实现完整
- ✅ 解析失败时显示错误信息

### 4. UI 增强
- ✅ 双视图显示：原始消息 | 解析消息
- ✅ 自动滚动到底部
- ✅ 消息数量限制（最新 100 条）
- ✅ 实时统计：原始消息数 | 解析消息数

---

## 🚀 快速启动步骤

### 步骤 1：启动 server-go

```bash
cd /workspace/server-go

# 复制调试配置
copy config.debug.json config.json

# 启动程序
go run main.go
```

**预期输出：**
```
🔍 正在查找系统中文字体...
✅ 找到中文字体: C:\Windows\Fonts\msyh.ttf
🚀 抖音直播监控系统 v3.2.1 启动...
✅ 数据库初始化成功
⚠️  调试模式已启用，跳过 License 验证
📡 正在启动 WebSocket 服务器 (端口: 8080)...
🌐 WebSocket 服务器正在启动，监听端口: 8080
📍 WebSocket 地址: ws://localhost:8080/monitor
📍 健康检查地址: http://localhost:8080/health
✅ WebSocket 服务器启动成功！
   📍 连接地址: ws://localhost:8080/monitor
   📍 健康检查: http://localhost:8080/health
   💡 提示: 浏览器插件需连接到此地址
✅ 启动图形界面...
```

### 步骤 2：安装浏览器插件

1. 打开 Chrome/Edge
2. 访问 `chrome://extensions/` 或 `edge://extensions/`
3. 启用"开发者模式"
4. 点击"加载已解压的扩展程序"
5. 选择 `/workspace/browser-monitor` 文件夹
6. 确认插件已安装并启用

**插件自动配置：**
- 服务器地址：`ws://localhost:8080/monitor`
- 过滤关键字：`live.douyin.com,webcast`

### 步骤 3：访问抖音直播间

打开浏览器，访问：
```
https://live.douyin.com/46387032209
```

或任意其他抖音直播间。

### 步骤 4：查看实时监控

#### 4.1 server-go 控制台

```
🔌 收到 WebSocket 连接请求: 127.0.0.1:xxxxx
✅ WebSocket 连接成功: 127.0.0.1:xxxxx
🎬 创建新房间: 46387032209 (Session: 1)
🎬 创建房间 Tab: 46387032209
✅ 房间 Tab 创建成功: 46387032209
```

#### 4.2 server-go GUI 界面

**自动创建 Tab：**
- 新增 `🏠 房间 46387032209` 标签页
- 自动切换到该 Tab

**双视图显示：**

```
┌─────────────────────────┬─────────────────────────┐
│ 📡 原始 WebSocket 消息  │ 📋 解析后的消息         │
├─────────────────────────┼─────────────────────────┤
│ [15:30:45]              │ [15:30:45]              │
│ URL: wss://webcast...   │ 类型: 聊天消息          │
│ Payload: CgoIAhDG...    │ 用户: 张三              │
│                         │ 内容: 666               │
├─────────────────────────┼─────────────────────────┤
│ [15:30:46]              │ [15:30:46]              │
│ URL: wss://webcast...   │ 类型: 礼物消息          │
│ Payload: CgoIAhDG...    │ 用户: 李四              │
│                         │ 礼物: 玫瑰 x1           │
└─────────────────────────┴─────────────────────────┘
```

---

## 🧪 测试 WebSocket 连接

### 方法 1：使用 HTML 测试工具

```bash
# 用浏览器打开
/workspace/server-go/TEST_WEBSOCKET.html
```

点击"连接"按钮，观察是否成功。

### 方法 2：使用批处理测试

```bash
cd /workspace/server-go
TEST_WEBSOCKET.bat
```

### 方法 3：使用 curl

```bash
curl http://localhost:8080/health
```

**预期输出：**
```json
{
  "status": "ok",
  "port": 8080,
  "clients": 0,
  "rooms": 0,
  "endpoints": {
    "websocket": "ws://localhost:8080/monitor",
    "health": "http://localhost:8080/health"
  }
}
```

---

## 📊 功能演示

### 场景 1：单房间监控

1. 打开 `https://live.douyin.com/46387032209`
2. server-go 自动创建 `🏠 房间 46387032209` Tab
3. 实时显示：
   - 左侧：原始 WebSocket 数据
   - 右侧：聊天、礼物、进入直播间等消息

### 场景 2：多房间监控

1. 打开 `https://live.douyin.com/46387032209`
2. 新标签页打开 `https://live.douyin.com/123456789`
3. server-go 自动创建两个 Tab：
   - `🏠 房间 46387032209`
   - `🏠 房间 123456789`
4. 每个房间独立监控、独立显示

### 场景 3：消息类型

**解析后的消息示例：**

| 类型 | 显示内容 |
|------|---------|
| 聊天消息 | `类型: 聊天消息 \| 用户: 张三 \| 内容: 666` |
| 礼物消息 | `类型: 礼物消息 \| 用户: 李四 \| 礼物: 火箭 x1` |
| 进入直播间 | `类型: 进入直播间 \| 用户: 王五` |
| 关注消息 | `类型: 关注消息 \| 用户: 赵六` |
| 点赞消息 | `类型: 点赞消息 \| 点赞数: 1000` |
| 解析失败 | `❌ 解析失败: Invalid wire type: 6` |

---

## 🔧 配置文件

### config.debug.json（推荐开发使用）

```json
{
  "server": {
    "port": 8080
  },
  "database": {
    "path": "data.db"
  },
  "license": {
    "server_url": "http://localhost:8081",
    "offline_grace_days": 7,
    "validation_interval": 60
  },
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": true
  }
}
```

### browser-monitor/config.js

```javascript
const CONFIG = {
  SERVER_URL: 'ws://localhost:8080/monitor',
  RECONNECT_INTERVAL: 3000,
  MAX_RECONNECT_ATTEMPTS: 10,
  MONITOR_DOUYIN: true,
  URL_FILTERS: ['live.douyin.com', 'webcast'],
  DEBUG: true
};
```

---

## ❓ 常见问题

### Q1: 浏览器插件连接失败

**症状：** 插件显示"连接失败"或控制台报错

**解决方法：**
1. 确认 server-go 已启动
2. 测试 `curl http://localhost:8080/health`
3. 检查端口是否被占用：`netstat -ano | findstr :8080`
4. 重新加载插件
5. 查看浏览器控制台（F12）错误信息

### Q2: 没有自动创建房间 Tab

**症状：** 打开直播间后 UI 没有新 Tab

**解决方法：**
1. 确认访问的是 `live.douyin.com` 域名
2. 查看 server-go 控制台是否有 `🎬 创建新房间` 日志
3. 检查插件是否正常连接
4. 确认插件过滤关键字配置正确

### Q3: 只有原始消息，没有解析消息

**症状：** 左侧有数据，右侧没有或显示"解析失败"

**解决方法：**
1. 这是正常的，部分消息类型可能解析失败
2. 查看控制台 `❌ 解析失败` 错误详情
3. 参考 dycast 项目对比 Protobuf 结构
4. 提交 Issue 包含错误日志

### Q4: 消息不实时更新

**症状：** 消息不刷新或延迟很大

**解决方法：**
1. 查看 server-go 控制台是否有 `💓 收到心跳`
2. 确认 WebSocket 连接正常
3. 重启 server-go
4. 重新加载浏览器页面

### Q5: 中文显示乱码

**症状：** UI 界面中文显示为方块或乱码

**解决方法：**
1. 确认启动日志显示 `✅ 找到中文字体`
2. 安装 Microsoft YaHei 字体
3. 参考 `/workspace/CHINESE_FONT_FIX.md`

---

## 📚 完整文档

| 文档 | 说明 |
|------|------|
| `/workspace/README.md` | 项目主文档 |
| `/workspace/DYNAMIC_ROOM_TAB_GUIDE.md` | 动态房间 Tab 详细指南 |
| `/workspace/WEBSOCKET_TEST_GUIDE.md` | WebSocket 测试完整指南 |
| `/workspace/CHINESE_FONT_FIX.md` | 中文字体修复指南 |
| `/workspace/DEBUG_MODE.md` | 调试模式文档 |

---

## 📝 修改文件清单

### 1. 服务器端

| 文件 | 修改内容 |
|------|---------|
| `server-go/main.go` | WebSocket 路径更新，设置 UIUpdater |
| `server-go/internal/server/websocket.go` | `/ws` → `/monitor`，UIUpdater 接口，handleDouyinMessage 增强 |
| `server-go/internal/ui/fyne_ui.go` | 添加 RoomTab、AddOrUpdateRoom、AddRawMessage、AddParsedMessage |

### 2. 浏览器插件

| 文件 | 修改内容 |
|------|---------|
| `browser-monitor/background.js` | 默认服务器地址改为 `/monitor`，默认过滤关键字 |
| `browser-monitor/config.js` | 配置文件（新增） |

### 3. 测试工具

| 文件 | 修改内容 |
|------|---------|
| `TEST_WEBSOCKET.html` | WebSocket 地址更新为 `/monitor` |
| `TEST_WEBSOCKET.bat` | WebSocket 地址更新为 `/monitor` |

### 4. 文档

| 文件 | 说明 |
|------|------|
| `QUICK_START_GUIDE.md` | 快速开始指南（本文档） |
| `DYNAMIC_ROOM_TAB_GUIDE.md` | 动态房间 Tab 详细指南 |

---

## 🎯 核心特性

### ✅ 动态房间 Tab
- 自动检测房间号
- 自动创建 Tab
- 支持多房间同时监控

### ✅ 双视图显示
- 左侧：原始 WebSocket 消息（Base64 Protobuf 数据）
- 右侧：解析后的结构化消息（类型、用户、内容等）

### ✅ 实时更新
- WebSocket 实时推送
- UI 自动刷新
- 自动滚动到底部

### ✅ 性能优化
- 最多保留 100 条消息
- 虚拟化列表渲染
- 线程安全设计

---

## 🔗 快速链接

- **启动程序**: `cd server-go && go run main.go`
- **测试连接**: 打开 `TEST_WEBSOCKET.html`
- **健康检查**: `curl http://localhost:8080/health`
- **抖音直播**: `https://live.douyin.com/46387032209`

---

## ⚡ 版本信息

- **版本**: v3.2.1
- **日期**: 2025-11-16
- **更新内容**:
  - ✅ WebSocket 路径 `/ws` → `/monitor`
  - ✅ 动态房间 Tab 功能
  - ✅ 实时双视图消息显示
  - ✅ UIUpdater 接口
  - ✅ 中文字体支持
  - ✅ 完整测试工具

---

**祝使用愉快！🎉**

如有问题，请查看详细文档或提交 Issue。
