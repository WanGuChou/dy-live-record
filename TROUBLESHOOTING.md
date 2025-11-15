# 故障排查指南

## 问题：浏览器有WebSocket连接但服务器未收到消息

### 📋 诊断步骤

#### 1. 检查服务器是否正在运行

```bash
# 进入server目录
cd server

# 启动服务器
npm start
```

**预期输出：**
```
============================================================
✅ WebSocket服务器已成功启动
============================================================

📡 服务器信息:
  - 端口: 8080
  - 路径: /monitor

🌐 连接地址:
  - 本地: ws://localhost:8080/monitor
  - 局域网: ws://192.168.x.x:8080/monitor

⏳ 等待客户端连接...
```

如果没有看到此输出，说明服务器未成功启动。

---

#### 2. 检查浏览器插件配置

1. 点击浏览器工具栏中的插件图标
2. 确认配置：
   - **服务器地址**: `ws://localhost:8080/monitor` （注意是 `ws://` 不是 `wss://`）
   - **启用监控**: 开关已打开（蓝色）
   - **连接状态**: 显示绿色"已连接"

3. 如果显示"连接中..."或红色"未连接"，说明连接有问题

---

#### 3. 查看浏览器插件日志

**Chrome浏览器：**
```
1. 打开 chrome://extensions/
2. 找到 "URL & WebSocket Monitor" 插件
3. 点击 "Service Worker" 链接
4. 在弹出的DevTools中查看Console标签
```

**Edge浏览器：**
```
1. 打开 edge://extensions/
2. 找到插件，点击 "检查视图"
3. 查看Console标签
```

**应该看到的日志：**
```
[时间] [URL Monitor] 开始加载配置...
[时间] [URL Monitor] 配置已加载: {serverUrl: "ws://localhost:8080/monitor", isEnabled: true, ...}
[时间] [URL Monitor] 监控已启用，开始连接WebSocket...
[时间] [URL Monitor] 🔌 正在创建WebSocket连接... {url: "ws://localhost:8080/monitor"}
[时间] [URL Monitor] ✅ WebSocket连接成功建立！
[时间] [URL Monitor] 📤 发送连接确认消息: {type: "connection", status: "connected", ...}
```

---

#### 4. 常见问题和解决方案

##### 问题A: 插件日志显示 "❌ WebSocket错误"

**可能原因：**
- 服务器未启动
- 服务器地址配置错误
- 防火墙阻止连接

**解决方案：**
1. 确保服务器正在运行（执行 `npm start`）
2. 确认地址是 `ws://localhost:8080/monitor` 而不是 `wss://`
3. 检查防火墙设置，允许端口8080
4. 尝试重启服务器和浏览器

---

##### 问题B: 插件日志显示 "⚠️ WebSocket未连接，消息未发送"

**可能原因：**
- "启用监控"开关未打开
- WebSocket连接尚未建立
- 连接已断开

**解决方案：**
1. 打开插件配置界面
2. 确认"启用监控"开关是打开的（蓝色）
3. 点击"保存配置"
4. 查看连接状态指示器是否变为绿色

---

##### 问题C: 服务器日志中没有 "新客户端已连接"

**可能原因：**
- 浏览器插件配置的地址错误
- 服务器路径配置不匹配
- CORS或网络问题

**解决方案：**
1. **检查地址完整性**：
   - 插件配置: `ws://localhost:8080/monitor`
   - 服务器配置: `path: '/monitor'`
   - 两者必须匹配！

2. **检查端口是否被占用**：
   ```bash
   # Linux/Mac
   lsof -i :8080
   
   # Windows
   netstat -ano | findstr :8080
   ```

3. **尝试更换端口**：
   - 编辑 `server/server.js`，修改 `port: 8080` 为其他端口
   - 同步修改插件配置中的地址

---

##### 问题D: 插件日志显示连接成功，但服务器没反应

**可能原因：**
- 运行了多个服务器实例
- 服务器代码有问题
- 查看了错误的服务器日志

**解决方案：**
1. **停止所有服务器实例**：
   ```bash
   # 查找并停止Node.js进程
   ps aux | grep node
   kill <进程ID>
   ```

2. **重新启动服务器**：
   ```bash
   cd server
   npm start
   ```

3. **确认服务器正在接收请求**：
   - 服务器启动后应该显示 "⏳ 等待客户端连接..."
   - 当浏览器插件连接时，应该显示 "🎉 新客户端已连接"

---

##### 问题E: 使用了 wss:// 而不是 ws://

**症状：**
- 插件日志显示连接错误
- 错误信息包含 SSL、TLS 或证书相关

**解决方案：**
1. 在插件配置中修改地址为 `ws://localhost:8080/monitor`
2. 本地开发环境使用 `ws://`（非加密）
3. 生产环境才使用 `wss://`（需要SSL证书）

---

### 🔍 完整诊断流程

按以下顺序检查：

```
1. ✅ 服务器已启动并显示 "WebSocket服务器已启动"
   ↓
2. ✅ 浏览器插件已安装并启用
   ↓
3. ✅ 插件配置中地址是 ws://localhost:8080/monitor
   ↓
4. ✅ "启用监控" 开关已打开
   ↓
5. ✅ 插件Service Worker日志显示 "✅ WebSocket连接成功建立！"
   ↓
6. ✅ 服务器日志显示 "🎉 新客户端已连接"
   ↓
7. ✅ 测试：打开新标签页，服务器应该收到 url_change 消息
```

---

### 🧪 手动测试连接

使用 `wscat` 工具测试服务器：

```bash
# 安装 wscat
npm install -g wscat

# 连接到服务器
wscat -c ws://localhost:8080/monitor

# 应该看到：
# connected (press CTRL+C to quit)
# < {"type":"welcome","message":"欢迎连接到URL监控服务器","timestamp":"..."}

# 发送测试消息
> {"type":"connection","status":"connected","timestamp":"2025-11-15T10:00:00.000Z"}

# 服务器应该输出收到的消息
```

---

### 📞 仍然无法解决？

请提供以下信息：

1. **服务器日志**（完整输出）
2. **浏览器插件Service Worker日志**（完整Console输出）
3. **插件配置截图**
4. **操作系统和浏览器版本**
   - 操作系统: Windows/Mac/Linux
   - 浏览器: Chrome/Edge + 版本号
5. **测试步骤**（你做了什么操作）

---

### 💡 快速测试脚本

创建文件 `test-connection.js`:

```javascript
// 简单的WebSocket客户端测试
const WebSocket = require('ws');

console.log('开始测试WebSocket连接...');
const ws = new WebSocket('ws://localhost:8080/monitor');

ws.on('open', () => {
  console.log('✅ 连接成功！');
  ws.send(JSON.stringify({
    type: 'test',
    message: 'Hello from test script',
    timestamp: new Date().toISOString()
  }));
  console.log('✅ 消息已发送');
});

ws.on('message', (data) => {
  console.log('✅ 收到服务器响应:', data.toString());
  ws.close();
  process.exit(0);
});

ws.on('error', (error) => {
  console.error('❌ 连接失败:', error.message);
  process.exit(1);
});

ws.on('close', () => {
  console.log('连接已关闭');
});
```

运行测试：
```bash
node test-connection.js
```

如果这个脚本能成功连接并收到服务器响应，说明服务器工作正常，问题出在浏览器插件。

---

**最后更新**: 2025-11-15
