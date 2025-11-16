# 动态房间 Tab 使用指南

## 功能说明

server-go 现在支持动态创建房间 Tab，每个抖音直播间会自动创建一个独立的监控标签页，实时显示原始 WebSocket 消息和解析后的结构化数据。

---

## 核心功能

### 1. 自动房间检测

- ✅ 当浏览器打开 `https://live.douyin.com/46387032209`
- ✅ server-go 自动创建 `🏠 房间 46387032209` Tab
- ✅ 无需手动配置，全自动检测

### 2. 双视图实时显示

#### 左侧：原始 WebSocket 消息
- 显示原始 URL 和 Payload 数据
- 截取前 200 字符（避免过长）
- 格式：
  ```
  [15:30:45] URL: wss://webcast...
  Payload: base64_encoded_data...
  ```

#### 右侧：解析后的消息
- 结构化显示消息内容
- 格式：
  ```
  [15:30:45] 类型: 礼物消息 | 用户: 张三 | 礼物: 火箭 x1
  [15:30:46] 类型: 聊天消息 | 用户: 李四 | 内容: 666
  [15:30:47] 类型: 进入直播间 | 用户: 王五
  ```

### 3. 消息统计

每个房间 Tab 顶部显示实时统计：
```
房间: 46387032209 | 原始消息: 50 条 | 解析消息: 48 条
```

### 4. 自动滚动

- ✅ 新消息自动滚动到底部
- ✅ 保留最新 100 条消息（自动清理旧消息）
- ✅ 高性能 List Widget，支持大量消息

---

## 使用步骤

### 步骤 1：启动 server-go

```bash
cd /workspace/server-go

# 使用调试配置
copy config.debug.json config.json

# 启动
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

### 步骤 2：配置浏览器插件

#### 2.1 安装插件

1. 打开 Chrome/Edge 浏览器
2. 访问 `chrome://extensions/` 或 `edge://extensions/`
3. 启用 "开发者模式"
4. 点击 "加载已解压的扩展程序"
5. 选择 `/workspace/browser-monitor` 文件夹

#### 2.2 配置服务器地址

插件会自动连接到 `ws://localhost:8080/monitor`

如果需要修改，编辑 `browser-monitor/config.js`：

```javascript
const CONFIG = {
  SERVER_URL: 'ws://localhost:8080/monitor',  // 修改这里
  // ...
};
```

### 步骤 3：访问抖音直播间

1. 打开浏览器
2. 访问任意抖音直播间，例如：
   - `https://live.douyin.com/46387032209`
   - `https://live.douyin.com/任意房间号`

### 步骤 4：查看实时监控

#### 4.1 自动创建房间 Tab

server-go 控制台输出：
```
🔌 收到 WebSocket 连接请求: 127.0.0.1:xxxxx
✅ WebSocket 连接成功: 127.0.0.1:xxxxx
🎬 创建房间: 46387032209 (Session: 1)
🎬 创建房间 Tab: 46387032209
✅ 房间 Tab 创建成功: 46387032209
```

server-go GUI 界面：
- 自动添加 `🏠 房间 46387032209` Tab
- 自动切换到新创建的 Tab

#### 4.2 查看实时消息

**左侧（原始消息）：**
```
[15:30:45] URL: wss://webcast-hl.douyincdn.com...
Payload: CgoIAhDGhAIY2A0SYgpgCkMSQTE4M...

[15:30:46] URL: wss://webcast-hl.douyincdn.com...
Payload: CgoIAhDGhAIY2A0SYgpgCkMSQTE4M...
```

**右侧（解析后消息）：**
```
[15:30:45] 类型: 聊天消息 | 用户: 张三 | 内容: 主播厉害啊
[15:30:46] 类型: 礼物消息 | 用户: 李四 | 礼物: 玫瑰 x1
[15:30:47] 类型: 进入直播间 | 用户: 王五
[15:30:48] 类型: 点赞消息 | 点赞数: 5
[15:30:49] ❌ 解析失败: Invalid wire type: 6
```

---

## UI 布局说明

### 主界面结构

```
┌─────────────────────────────────────────────────────────────┐
│  抖音直播监控系统 v3.2.1 [调试模式]                           │
├─────────────────────────────────────────────────────────────┤
│  📊 统计卡片                                                  │
│  礼物总数: 0 | 消息总数: 0 | 礼物总值: 0 钻石 | 在线用户: 0   │
├─────────────────────────────────────────────────────────────┤
│ [📊 数据概览] [🎁 礼物记录] [🏠 房间 46387032209] [⚙️ 设置]   │
├─────────────────────────────────────────────────────────────┤
│  房间: 46387032209 | 原始消息: 50 条 | 解析消息: 48 条        │
├──────────────────────────┬──────────────────────────────────┤
│ 📡 原始 WebSocket 消息   │ 📋 解析后的消息                   │
│                          │                                  │
│ [15:30:45] URL: wss://...│ [15:30:45] 类型: 聊天消息        │
│ Payload: CgoIAhDG...     │ 用户: 张三 | 内容: 666           │
│                          │                                  │
│ [15:30:46] URL: wss://...│ [15:30:46] 类型: 礼物消息        │
│ Payload: CgoIAhDG...     │ 用户: 李四 | 礼物: 玫瑰 x1       │
│                          │                                  │
│ ↓ 自动滚动               │ ↓ 自动滚动                       │
└──────────────────────────┴──────────────────────────────────┘
```

### Tab 标签页说明

| Tab | 功能 |
|-----|------|
| 📊 数据概览 | 系统整体说明和状态 |
| 🎁 礼物记录 | 礼物统计表格 |
| 💬 消息记录 | 消息统计表格 |
| 👤 主播管理 | 主播配置和礼物绑定 |
| 📈 分段记分 | 分段统计功能 |
| 🏠 房间 {ID} | **动态创建的房间监控 Tab** |
| ⚙️ 设置 | 系统设置 |

---

## 多房间监控

### 支持同时监控多个房间

**场景：** 同时打开多个直播间

1. **浏览器 Tab 1**：`https://live.douyin.com/46387032209`
2. **浏览器 Tab 2**：`https://live.douyin.com/123456789`

**server-go 自动创建：**
- `🏠 房间 46387032209` Tab
- `🏠 房间 123456789` Tab

**每个房间独立监控：**
- 独立的消息列表
- 独立的统计信息
- 独立的数据存储

---

## 消息类型说明

### 解析后的消息类型

| 类型 | 说明 | 示例 |
|------|------|------|
| **聊天消息** | 用户发送的弹幕 | `类型: 聊天消息 \| 用户: 张三 \| 内容: 666` |
| **礼物消息** | 用户赠送的礼物 | `类型: 礼物消息 \| 用户: 李四 \| 礼物: 火箭 x1` |
| **进入直播间** | 用户进入直播间 | `类型: 进入直播间 \| 用户: 王五` |
| **关注消息** | 用户关注主播 | `类型: 关注消息 \| 用户: 赵六` |
| **点赞消息** | 点赞统计 | `类型: 点赞消息 \| 点赞数: 1000` |
| **❌ 解析失败** | 消息解析错误 | `❌ 解析失败: Invalid wire type: 6` |

---

## 性能特性

### 1. 高性能列表

- 使用 Fyne 的 `widget.List`
- 虚拟化渲染，只渲染可见项
- 支持数万条消息流畅滚动

### 2. 内存管理

- 每个房间最多保留 100 条消息
- 自动清理旧消息
- 避免内存泄漏

### 3. 线程安全

- UI 更新在主线程
- WebSocket 消息处理在独立 goroutine
- 使用接口解耦

---

## 调试技巧

### 1. 查看控制台日志

**server-go 控制台会显示：**
```
🔌 收到 WebSocket 连接请求: 127.0.0.1:xxxxx
✅ WebSocket 连接成功: 127.0.0.1:xxxxx
💓 收到心跳
🎬 创建新房间: 46387032209 (Session: 1)
╔══════════════════════════════════════════════════════════════╗
║ 🎬 抖音直播消息
╠══════════════════════════════════════════════════════════════╣
║ 消息类型: 聊天消息
║ 时间: 2025-11-16T15:30:45Z
║ 用户: 张三
║ 内容: 666
╚══════════════════════════════════════════════════════════════╝
```

### 2. 查看解析错误

如果解析失败，右侧会显示：
```
[15:30:45] ❌ 解析失败: Invalid wire type: 6
```

控制台同时输出：
```
❌ [房间 46387032209] 解析失败: Invalid wire type: 6
```

### 3. 测试消息流

**方法 1：** 在直播间发送弹幕
- 输入 "测试消息"
- 观察 server-go 是否实时显示

**方法 2：** 送礼物
- 送一个免费礼物
- 观察礼物消息是否正确解析

**方法 3：** 点赞
- 连续点赞
- 观察点赞统计是否累加

---

## 故障排除

### 问题 1：没有自动创建房间 Tab

**原因：**
- 插件未连接到服务器
- 房间 URL 不正确

**解决方法：**
1. 确认 server-go 已启动
2. 确认插件已安装并启用
3. 打开浏览器开发者工具（F12），查看 Console 是否有错误
4. 确认访问的是 `live.douyin.com` 域名

### 问题 2：只有原始消息，没有解析消息

**原因：**
- Protobuf 解析失败
- 消息格式变化

**解决方法：**
1. 查看控制台是否有 `❌ 解析失败` 错误
2. 对比 dycast 项目的最新实现
3. 检查 `server-go/internal/parser/` 中的解析逻辑

### 问题 3：消息不实时更新

**原因：**
- WebSocket 连接断开
- UI 刷新失败

**解决方法：**
1. 查看控制台 `💓 收到心跳` 是否正常
2. 重启 server-go
3. 重新加载浏览器插件

### 问题 4：中文乱码

**原因：**
- 字体未正确加载

**解决方法：**
1. 确认启动日志显示 `✅ 找到中文字体`
2. 安装 Microsoft YaHei 字体
3. 查看 `/workspace/CHINESE_FONT_FIX.md`

---

## API 接口说明

### UIUpdater 接口

```go
type UIUpdater interface {
    AddOrUpdateRoom(roomID string)
    AddRawMessage(roomID string, message string)
    AddParsedMessage(roomID string, message string)
}
```

### 调用流程

```
WebSocket 收到消息
    ↓
提取房间号
    ↓
调用 AddOrUpdateRoom(roomID) → 创建 Tab（如果不存在）
    ↓
调用 AddRawMessage(roomID, msg) → 显示原始消息
    ↓
解析消息
    ↓
调用 AddParsedMessage(roomID, msg) → 显示解析结果
```

---

## 技术细节

### 1. 房间 Tab 结构

```go
type RoomTab struct {
    RoomID       string              // 房间号
    Tab          *container.TabItem  // Tab 项
    RawMessages  *widget.List        // 原始消息列表
    ParsedMsgs   *widget.List        // 解析消息列表
    RawData      []string            // 原始数据
    ParsedData   []string            // 解析数据
    StatsLabel   *widget.Label       // 统计标签
}
```

### 2. WebSocket 路径

- **新路径**：`ws://localhost:8080/monitor`
- **旧路径**：`ws://localhost:8080/ws`（已废弃）

### 3. 消息格式

**插件发送到 server-go：**
```json
{
  "type": "websocket_frame_received",
  "url": "wss://webcast-hl.douyincdn.com/...",
  "payloadData": "base64_encoded_protobuf_data"
}
```

**server-go 解析后：**
```json
{
  "messageType": "聊天消息",
  "user": "张三",
  "content": "666",
  "timestamp": "2025-11-16T15:30:45Z"
}
```

---

## 配置文件

### config.json

```json
{
  "server": {
    "port": 8080
  },
  "database": {
    "path": "data.db"
  },
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": true
  }
}
```

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

---

## 更新日志

### v3.2.1 (2025-11-16)

✅ **新增功能：**
- 动态房间 Tab 系统
- 实时原始消息显示
- 实时解析消息显示
- 自动房间检测
- 多房间同时监控

✅ **优化：**
- WebSocket 路径改为 `/monitor`
- UI 更新接口解耦
- 高性能消息列表
- 自动内存管理

✅ **修复：**
- WebSocket 启动时序问题
- 中文字体显示问题

---

## 下一步计划

### 短期（v3.3.0）
- [ ] 消息过滤功能
- [ ] 消息导出功能
- [ ] 房间 Tab 关闭按钮
- [ ] 消息搜索功能

### 中期（v3.4.0）
- [ ] 礼物价值实时统计
- [ ] 主播业绩自动分配
- [ ] 数据可视化图表
- [ ] 历史数据回放

### 长期（v4.0.0）
- [ ] 多平台支持（快手、B站）
- [ ] 云端数据同步
- [ ] 移动端查看
- [ ] AI 自动分析

---

## 相关文档

- [WebSocket 测试指南](/workspace/WEBSOCKET_TEST_GUIDE.md)
- [中文字体修复指南](/workspace/CHINESE_FONT_FIX.md)
- [UI 语言变更说明](/workspace/UI_LANGUAGE_CHANGE.md)
- [调试模式文档](/workspace/DEBUG_MODE.md)

---

## 技术支持

### GitHub Issues
https://github.com/WanGuChou/dy-live-record/issues

### 常见问题
参考 `/workspace/WEBSOCKET_TEST_GUIDE.md` 中的故障排除章节

---

**开始使用：**

```bash
cd /workspace/server-go
copy config.debug.json config.json
go run main.go
```

然后打开浏览器访问 `https://live.douyin.com/46387032209`，观察 server-go 自动创建房间 Tab！ 🚀
