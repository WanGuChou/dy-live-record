# CDP Monitor 使用指南

## 🎯 版本 v2.0.0 - Chrome DevTools Protocol

这是一个重大版本升级，插件现在使用 **Chrome DevTools Protocol (CDP)** 来深度监控浏览器的所有网络活动。

---

## 🚀 快速开始

### 1. 启动服务器

```bash
cd server
npm install   # 首次运行需要安装依赖
npm start     # 启动服务器
```

**服务器启动后会显示：**
```
════════════════════════════════════════════════════════════════════════════════
CDP Monitor 服务器已启动
地址: ws://localhost:8080/monitor
════════════════════════════════════════════════════════════════════════════════

监控内容:
  ✅ 所有 HTTP/HTTPS 请求 (使用 Chrome DevTools Protocol)
  ✅ WebSocket 连接创建
  ✅ WebSocket 握手过程
  ✅ WebSocket 发送的所有消息
  ✅ WebSocket 接收的所有消息
  ✅ WebSocket 连接关闭

等待客户端连接...
```

---

### 2. 安装插件

⚠️ **重要**: 如果之前安装了旧版本，请先完全卸载旧版本！

1. 打开浏览器
   - Chrome: `chrome://extensions/`
   - Edge: `edge://extensions/`

2. 启用"开发者模式"

3. 点击"加载已解压的扩展程序"

4. 选择 `dy-live-record/brower-monitor` 目录

5. **授予debugger权限** (系统会提示)

---

### 3. 配置插件

1. 点击插件图标打开配置面板

2. 配置服务器地址：
   ```
   ws://localhost:8080/monitor
   ```

3. （可选）设置过滤关键字：
   ```
   例如: live,video,websocket
   ```
   留空则监控所有请求

4. 打开"启用监控"开关

5. 点击"保存配置"

6. （可选）点击"测试连接"验证服务器连接

---

## 📊 监控的内容

### HTTP/HTTPS 请求

插件会捕获：
- ✅ 请求方法 (GET, POST, PUT, DELETE等)
- ✅ 完整URL
- ✅ 请求头 (Headers)
- ✅ POST数据 (如果有)
- ✅ 响应状态码
- ✅ 响应头
- ✅ MIME类型
- ✅ 资源类型 (document, script, stylesheet, image等)

**服务器日志示例：**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 📤 HTTP 请求 #1
├──────────────────────────────────────────────────────────────────────────────┤
│ 方法: GET
│ URL: https://example.com/api/data
│ 资源类型: fetch
│ 标签页ID: 123
│ 请求ID: 1234.1
│ 请求头:
│     accept: application/json
│     user-agent: Mozilla/5.0...
│ 时间: 2025-11-15T14:30:00.000Z
└──────────────────────────────────────────────────────────────────────────────┘
```

---

### WebSocket 完整生命周期

#### 1. WebSocket 创建

```
╔══════════════════════════════════════════════════════════════════════════════╗
║ 🔌 WebSocket 创建 #1
╠══════════════════════════════════════════════════════════════════════════════╣
║ 完整URL: wss://example.com/live/socket
║ 标签页ID: 123
║ 请求ID: 5678.1
║ 时间: 2025-11-15T14:30:00.000Z
╚══════════════════════════════════════════════════════════════════════════════╝
```

#### 2. WebSocket 握手

**握手请求：**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 🤝 WebSocket 握手请求
├──────────────────────────────────────────────────────────────────────────────┤
│ URL: wss://example.com/live/socket
│ 请求ID: 5678.1
│ 握手请求头:
│     Upgrade: websocket
│     Connection: Upgrade
│     Sec-WebSocket-Key: xxxxx
│     Sec-WebSocket-Version: 13
│ 时间: 2025-11-15T14:30:00.100Z
└──────────────────────────────────────────────────────────────────────────────┘
```

**握手响应：**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ ✅ WebSocket 握手响应
├──────────────────────────────────────────────────────────────────────────────┤
│ 状态码: 101 Switching Protocols
│ URL: wss://example.com/live/socket
│ 握手响应头:
│     Upgrade: websocket
│     Connection: Upgrade
│     Sec-WebSocket-Accept: xxxxx
│ 时间: 2025-11-15T14:30:00.200Z
└──────────────────────────────────────────────────────────────────────────────┘
```

#### 3. WebSocket 消息发送

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 📤 WebSocket 发送消息
├──────────────────────────────────────────────────────────────────────────────┤
│ WebSocket URL: wss://example.com/live/socket
│ 请求ID: 5678.1
│ Opcode: 1 (text frame)
│ Mask: true
│ 消息内容:
│   {"type":"subscribe","channel":"live-123"}
│ 消息长度: 42 字符
│ 时间: 2025-11-15T14:30:01.000Z
└──────────────────────────────────────────────────────────────────────────────┘
```

#### 4. WebSocket 消息接收

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ 📥 WebSocket 接收消息
├──────────────────────────────────────────────────────────────────────────────┤
│ WebSocket URL: wss://example.com/live/socket
│ 请求ID: 5678.1
│ Opcode: 1 (text frame)
│ Mask: false
│ 消息内容:
│   {"type":"message","data":"Hello from server!"}
│ 消息长度: 47 字符
│ 时间: 2025-11-15T14:30:01.100Z
└──────────────────────────────────────────────────────────────────────────────┘
```

#### 5. WebSocket 关闭

```
╔══════════════════════════════════════════════════════════════════════════════╗
║ 🔌 WebSocket 已关闭
╠══════════════════════════════════════════════════════════════════════════════╣
║ WebSocket URL: wss://example.com/live/socket
║ 请求ID: 5678.1
║ 时间: 2025-11-15T14:30:10.000Z
╚══════════════════════════════════════════════════════════════════════════════╝
```

---

## 🔬 技术细节

### CDP vs 旧版本的区别

| 特性 | 旧版本 (v1.x) | 新版本 (v2.0 CDP) |
|------|---------------|-------------------|
| 监控方式 | chrome.webRequest API | Chrome DevTools Protocol |
| WebSocket消息 | ❌ 不能捕获 | ✅ 完整捕获 |
| 请求头 | ⚠️ 部分 | ✅ 完整 |
| POST数据 | ⚠️ 有限 | ✅ 完整 |
| 响应头 | ⚠️ 部分 | ✅ 完整 |
| 性能影响 | 小 | 小到中等 |

### WebSocket Opcode 说明

| Opcode | 类型 | 说明 |
|--------|------|------|
| 0 | continuation frame | 延续帧 |
| 1 | text frame | 文本消息 |
| 2 | binary frame | 二进制消息 |
| 8 | connection close | 连接关闭 |
| 9 | ping | Ping帧 |
| 10 | pong | Pong帧 |

---

## ⚠️ 重要注意事项

### 1. 调试提示

启用监控后，浏览器会显示：

```
正在调试此浏览器
```

**这是正常现象！** 因为插件使用了Chrome DevTools Protocol。

### 2. 性能影响

- CDP监控会比旧版本略微增加资源占用
- 对于大多数网站，影响可忽略不计
- 如果网站有大量WebSocket消息，可能会产生大量日志

**建议：**
- 使用过滤关键字减少不必要的监控
- 仅在需要调试时启用监控

### 3. 权限要求

CDP监控需要 `debugger` 权限，这比旧版本的权限更高。

**用户授权流程：**
1. 安装插件时会提示授权
2. 点击"添加扩展程序"
3. 浏览器会显示调试提示

---

## 🧪 测试场景

### 测试 HTTP 请求

1. 启动服务器和插件
2. 访问任意网站 (如 https://www.baidu.com)
3. 查看服务器日志

**预期看到：**
- 主页面请求 (document)
- CSS文件 (stylesheet)
- JavaScript文件 (script)
- 图片 (image)
- API请求 (fetch/xmlhttprequest)

### 测试 WebSocket

**方法1: 使用测试网站**

访问 https://www.websocket.org/echo.html

**方法2: 在Console中测试**

```javascript
// 创建WebSocket连接
const ws = new WebSocket('wss://echo.websocket.org/');

ws.onopen = () => {
  console.log('WebSocket opened');
  ws.send('Hello Server!');
};

ws.onmessage = (e) => {
  console.log('Received:', e.data);
};
```

**预期在服务器看到：**
1. 🔌 WebSocket 创建
2. 🤝 握手请求
3. ✅ 握手响应
4. 📤 发送消息: "Hello Server!"
5. 📥 接收消息: "Hello Server!"

---

## 🐛 故障排查

### Q1: 插件安装后无法工作

**检查：**
1. 是否授予了 `debugger` 权限
2. 是否重新加载了插件
3. Service Worker 是否运行

**解决：**
```
1. 卸载插件
2. 清除所有数据
3. 重新安装并授权
```

### Q2: 看不到WebSocket消息

**检查：**
1. 确认版本是 v2.0.0
2. 确认监控已启用
3. 确认过滤关键字正确
4. 查看Service Worker日志

**Service Worker日志位置：**
```
chrome://extensions/ 
→ 找到插件 
→ 点击 "Service Worker"
```

### Q3: 服务器收不到消息

**检查：**
1. 服务器是否运行
2. 插件配置的服务器地址是否正确
3. 查看插件的连接状态
4. 防火墙是否阻止8080端口

---

## 📈 性能优化建议

### 1. 使用过滤关键字

```
# 只监控包含这些关键字的请求
live,video,websocket,api
```

### 2. 仅在需要时启用

不调试时关闭监控开关，减少资源占用。

### 3. 定期重启

长时间运行后，可以重启插件释放内存：
```
chrome://extensions/
→ 找到插件
→ 点击刷新按钮
```

---

## 📚 相关文档

- **主文档**: [README.md](../README.md)
- **技术细节**: [IMPROVEMENTS.md](../IMPROVEMENTS.md)
- **服务器文档**: [server/README.md](../server/README.md)

---

## 🎓 高级用法

### 自定义过滤逻辑

修改 `background.js` 中的 `matchesFilter` 函数：

```javascript
function matchesFilter(url) {
  // 自定义过滤逻辑
  // 例如：只监控特定域名
  return url.includes('example.com') || url.includes('api.mysite.com');
}
```

### 处理服务器端数据

修改 `server.js` 来添加自定义处理逻辑：

```javascript
case 'websocket_frame_received':
  // 解析消息
  try {
    const payload = JSON.parse(data.payloadData);
    // 做些什么...
    console.log('解析的WebSocket消息:', payload);
  } catch (e) {
    // 不是JSON消息
  }
  break;
```

---

**版本**: v2.0.0  
**更新时间**: 2025-11-15  
**使用CDP**: Chrome DevTools Protocol 1.3
