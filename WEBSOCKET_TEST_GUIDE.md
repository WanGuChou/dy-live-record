# WebSocket 服务器测试指南

## 问题说明

启动 `server-go` 后，WebSocket 无法连接的问题已修复。

### 修复内容

1. **启动时序问题**
   - ✅ 添加了启动确认机制 (`started` channel)
   - ✅ `Start()` 方法现在等待服务器真正监听后才返回
   - ✅ 避免了"日志显示启动成功但服务器还未就绪"的问题

2. **日志改进**
   - ✅ 详细的启动日志
   - ✅ WebSocket 连接请求日志
   - ✅ 连接成功/失败详细信息

3. **健康检查增强**
   - ✅ 返回 JSON 格式数据
   - ✅ 显示客户端数量、房间数量
   - ✅ 提供所有可用端点信息

4. **测试工具**
   - ✅ `TEST_WEBSOCKET.bat` - 命令行测试脚本
   - ✅ `TEST_WEBSOCKET.html` - Web 可视化测试工具

---

## 快速测试步骤

### 步骤 1：启动服务器

```bash
cd /workspace/server-go

# 使用调试配置
copy config.debug.json config.json

# 启动程序
go run main.go
```

### 步骤 2：确认启动成功

**预期日志输出：**

```
🔍 正在查找系统中文字体...
✅ 找到中文字体: C:\Windows\Fonts\msyh.ttf
2025/11/16 xx:xx:xx main.go:30: 🚀 抖音直播监控系统 v3.2.1 启动...
2025/11/16 xx:xx:xx checker.go:35: 🔍 开始检查系统依赖...
✅ 所有依赖检查通过
✅ 数据库初始化成功
⚠️  调试模式已启用，跳过 License 验证
📡 正在启动 WebSocket 服务器 (端口: 8080)...
🌐 WebSocket 服务器正在启动，监听端口: 8080
📍 WebSocket 地址: ws://localhost:8080/ws
📍 健康检查地址: http://localhost:8080/health
✅ WebSocket 服务器启动成功！
   📍 连接地址: ws://localhost:8080/ws
   📍 健康检查: http://localhost:8080/health
   💡 提示: 浏览器插件需连接到此地址
✅ 启动图形界面...
```

**如果看到以上日志，说明服务器启动成功！**

---

## 测试方法一：使用批处理脚本

### 1. 运行测试脚本

```bash
cd /workspace/server-go
TEST_WEBSOCKET.bat
```

### 2. 预期输出

```
========================================
WebSocket 服务器连接测试
========================================

[1/3] 读取配置文件...
配置端口: 8080

[2/3] 测试健康检查接口...
{"status":"ok","port":8080,"clients":0,"rooms":0,"endpoints":{"health":"http://localhost:8080/health","websocket":"ws://localhost:8080/ws"}}
✅ 健康检查成功！

[3/3] WebSocket 连接信息
========================================
WebSocket 地址: ws://localhost:8080/ws
健康检查地址:   http://localhost:8080/health
========================================
```

**如果显示 "✅ 健康检查成功"，说明服务器工作正常！**

---

## 测试方法二：使用 Web 测试工具 (推荐)

### 1. 打开测试页面

用浏览器打开：
```
/workspace/server-go/TEST_WEBSOCKET.html
```

或者双击文件 `TEST_WEBSOCKET.html`

### 2. 测试步骤

#### (1) 自动健康检查

- 页面加载后会自动测试健康检查接口
- 查看日志区域是否显示 "✅ 健康检查成功"

#### (2) 连接 WebSocket

1. 确认 WebSocket 地址：`ws://localhost:8080/ws`
2. 点击 **"连接"** 按钮
3. 观察日志输出

**成功连接日志示例：**
```
[xx:xx:xx] 正在连接到: ws://localhost:8080/ws
[xx:xx:xx] ✅ WebSocket 连接成功！
```

**同时，server-go 控制台会显示：**
```
🔌 收到 WebSocket 连接请求: 127.0.0.1:xxxxx
✅ WebSocket 连接成功: 127.0.0.1:xxxxx
```

#### (3) 发送测试消息

1. 在 "测试消息" 输入框中输入 JSON 消息（默认是心跳消息）：
   ```json
   {"type":"heartbeat","timestamp":1234567890}
   ```

2. 点击 **"发送测试消息"** 按钮

3. 观察日志输出

**预期日志：**
- 测试页面：`📤 发送消息: {"type":"heartbeat","timestamp":1234567890}`
- server-go 控制台：`💓 收到心跳`

### 3. 功能说明

| 按钮 | 功能 |
|------|------|
| **连接** | 连接到 WebSocket 服务器 |
| **断开** | 断开 WebSocket 连接 |
| **测试健康检查** | 测试 `/health` 接口 |
| **清空日志** | 清空日志显示区域 |
| **发送测试消息** | 发送自定义 JSON 消息 |

### 4. 状态指示

- ⭕ **未连接**：红色背景，未建立连接
- ✅ **已连接**：绿色背景，连接成功

---

## 测试方法三：使用 curl 命令

### 1. 测试健康检查接口

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
    "health": "http://localhost:8080/health",
    "websocket": "ws://localhost:8080/ws"
  }
}
```

### 2. 使用 wscat 测试 WebSocket (可选)

如果已安装 `wscat` (Node.js 工具)：

```bash
# 安装 wscat
npm install -g wscat

# 连接到 WebSocket
wscat -c ws://localhost:8080/ws

# 连接成功后，发送测试消息
> {"type":"heartbeat","timestamp":1234567890}
```

---

## 常见问题排查

### 问题 1：健康检查失败

**错误信息：**
```
❌ 健康检查失败！
curl: (7) Failed to connect to localhost port 8080
```

**原因：**
- dy-live-monitor.exe 未运行
- 端口 8080 未正确启动

**解决方法：**
1. 确认程序正在运行：`tasklist | findstr dy-live-monitor`
2. 检查启动日志是否有错误
3. 确认端口配置正确（查看 `config.json`）

---

### 问题 2：端口被占用

**错误信息：**
```
❌ WebSocket 服务器启动失败: listen tcp :8080: bind: Only one usage of each socket address
```

**解决方法：**

#### 方法 1：查找并关闭占用端口的进程

```bash
# 查找占用端口的进程
netstat -ano | findstr :8080

# 关闭进程 (PID 是上面查到的进程ID)
taskkill /PID <进程ID> /F
```

#### 方法 2：修改端口

编辑 `config.json`：
```json
{
  "server": {
    "port": 8081
  }
}
```

重启程序。

---

### 问题 3：WebSocket 连接超时

**错误信息（浏览器控制台）：**
```
WebSocket connection to 'ws://localhost:8080/ws' failed
```

**原因：**
- 防火墙阻止连接
- 服务器未正确启动
- 端口配置错误

**解决方法：**
1. **检查防火墙**：
   - Windows 防火墙 → 允许应用通过防火墙
   - 添加 `dy-live-monitor.exe`

2. **确认服务器状态**：
   ```bash
   curl http://localhost:8080/health
   ```

3. **检查端口配置**：
   - 浏览器插件配置的端口
   - `config.json` 中的端口
   - 确保两者一致

---

### 问题 4：浏览器插件无法连接

**原因：**
- 插件配置的服务器地址错误
- 插件未正确加载

**解决方法：**

1. **检查插件配置**：
   - 打开插件选项页面
   - 确认服务器地址为：`ws://localhost:8080/ws`

2. **重新加载插件**：
   - Chrome：`chrome://extensions/`
   - Edge：`edge://extensions/`
   - 点击 "重新加载" 按钮

3. **查看插件日志**：
   - 打开开发者工具（F12）
   - 切换到 Console 标签
   - 查看是否有错误信息

---

## 验证完整流程

### 1. 启动服务器

```bash
cd /workspace/server-go
go run main.go
```

### 2. 测试健康检查

```bash
curl http://localhost:8080/health
```

预期输出：`{"status":"ok",...}`

### 3. 测试 WebSocket 连接

打开 `TEST_WEBSOCKET.html`，点击 "连接"，观察日志。

### 4. 安装浏览器插件

1. 打开 Chrome/Edge 扩展页面
2. 启用 "开发者模式"
3. 加载 `/workspace/browser-monitor` 文件夹

### 5. 访问抖音直播间

1. 打开 https://live.douyin.com/任意房间号
2. 观察 server-go 控制台日志
3. 应该看到：
   - `🔌 收到 WebSocket 连接请求`
   - `✅ WebSocket 连接成功`
   - `🎬 创建新房间: xxxxx`

### 6. 查看数据

- server-go 控制台会实时显示礼物、弹幕消息
- Fyne GUI 界面会显示统计数据

---

## 日志说明

### 正常启动日志标识

| 符号 | 含义 |
|------|------|
| 🚀 | 程序启动 |
| 🔍 | 检查依赖/查找字体 |
| ✅ | 操作成功 |
| 📡 | 正在启动服务器 |
| 🌐 | WebSocket 相关 |
| 📍 | 地址信息 |
| 💡 | 提示信息 |
| ⚠️ | 警告（非致命） |
| ❌ | 错误（致命） |

### WebSocket 连接日志

| 符号 | 含义 |
|------|------|
| 🔌 | 收到连接请求 |
| ✅ | 连接成功 |
| 👋 | 客户端断开 |
| 💓 | 心跳消息 |
| 📨 | 收到消息 |
| 🎬 | 创建新房间 |
| 🎁 | 礼物消息 |
| 💬 | 聊天消息 |

---

## 配置文件说明

### config.json 结构

```json
{
  "server": {
    "port": 8080              // WebSocket 服务器端口
  },
  "database": {
    "path": "data.db"         // SQLite 数据库文件路径
  },
  "license": {
    "server_url": "...",      // License 服务器地址
    "public_key_path": "..."  // RSA 公钥路径
  },
  "debug": {
    "enabled": true,          // 启用调试模式
    "skip_license": true,     // 跳过 License 验证
    "verbose_log": true       // 详细日志
  }
}
```

### 调试模式配置

**开发环境（推荐）：**
```json
{
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": true
  }
}
```

**生产环境：**
```json
{
  "debug": {
    "enabled": false,
    "skip_license": false,
    "verbose_log": false
  }
}
```

---

## 性能指标

### 正常运行指标

- **启动时间**：< 3 秒
- **WebSocket 连接时间**：< 100ms
- **消息处理延迟**：< 50ms
- **内存占用**：< 100MB（无负载）
- **CPU 占用**：< 5%（无负载）

### 监控命令

```bash
# 查看进程
tasklist | findstr dy-live-monitor

# 查看端口占用
netstat -ano | findstr :8080

# 查看详细进程信息（PowerShell）
Get-Process dy-live-monitor | Format-List *
```

---

## 下一步

### ✅ WebSocket 测试通过后

1. **安装浏览器插件**
   - 加载 `browser-monitor` 文件夹
   - 确认插件已启用

2. **配置插件**
   - 设置服务器地址：`ws://localhost:8080/ws`
   - 保存配置

3. **测试数据采集**
   - 访问抖音直播间
   - 观察 server-go 日志
   - 检查数据是否正确解析

4. **查看数据统计**
   - Fyne GUI 界面实时显示
   - 礼物记录、消息记录等

---

## 技术支持

### 相关文件

- `/workspace/server-go/main.go` - 主程序入口
- `/workspace/server-go/internal/server/websocket.go` - WebSocket 服务器实现
- `/workspace/server-go/TEST_WEBSOCKET.bat` - 批处理测试脚本
- `/workspace/server-go/TEST_WEBSOCKET.html` - Web 测试工具

### 日志文件

- 控制台日志：实时输出
- 数据库：`data.db` (SQLite)

### 更新日志

- **v3.2.1** (2025-11-16)
  - ✅ 修复 WebSocket 启动时序问题
  - ✅ 添加启动确认机制
  - ✅ 增强健康检查接口
  - ✅ 创建测试工具

---

**测试完成后，请运行以下命令确认一切正常：**

```bash
# 1. 测试健康检查
curl http://localhost:8080/health

# 2. 打开 Web 测试工具
start TEST_WEBSOCKET.html

# 3. 连接并发送测试消息
```

如果以上测试都通过，说明 WebSocket 服务器工作正常！ 🎉
