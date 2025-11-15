# DY Live Record - CDP深度监控项目

## 🎯 项目概述

这是一个抖音直播录制相关的项目，包含基于 **Chrome DevTools Protocol (CDP)** 的浏览器深度监控插件，能够捕获和记录：
- ✅ 所有 HTTP/HTTPS 网络请求
- ✅ WebSocket 连接的完整生命周期
- ✅ WebSocket 发送和接收的所有消息内容
- ✅ 请求和响应的完整头部信息

**最新版本**: v2.2.0 🚀  
**重大更新**: 使用 Chrome DevTools Protocol 实现深度监控  
**新功能**: 
- 🎬 抖音直播WebSocket消息自动解析（Protobuf + GZIP）⭐⭐⭐
- 📝 WebSocket消息自动保存到日志文件（按日期和小时分组）⭐⭐

---

## 📦 项目结构

```
dy-live-record/
├── brower-monitor/          # Chrome/Edge 浏览器扩展插件 (CDP版本)
│   ├── manifest.json        # 扩展配置文件 (需要debugger权限)
│   ├── background.js        # CDP监控核心逻辑
│   ├── popup.html          # 配置界面HTML
│   ├── popup.js            # 配置界面脚本
│   ├── icons/              # 图标文件夹
│   └── README.md           # 插件文档
└── server/                  # WebSocket服务器
    ├── server.js            # 服务器主程序 (支持CDP消息格式)
    ├── package.json         # Node.js依赖配置
    └── README.md           # 服务器文档
```

---

## 🚀 快速开始

### 1. 启动WebSocket服务器

```bash
cd server
npm install   # 首次运行需要
npm start
```

**服务器启动后会显示：**
```
════════════════════════════════════════════════════════════════════════════════
CDP Monitor 服务器已启动
地址: ws://localhost:8080/monitor
════════════════════════════════════════════════════════════════════════════════
```

### 2. 安装浏览器插件

⚠️ **重要**: 如果之前安装了v1.x版本，请先完全卸载！

1. 打开 Chrome/Edge 扩展管理页面
   - Chrome: `chrome://extensions/`
   - Edge: `edge://extensions/`

2. 启用"开发者模式"

3. 点击"加载已解压的扩展程序"

4. 选择 `dy-live-record/brower-monitor` 文件夹

5. **授予 debugger 权限** (系统会提示)

### 3. 配置插件

1. 点击插件图标
2. 设置服务器地址: `ws://localhost:8080/monitor`
3. （可选）设置过滤关键字
4. 打开"启用监控"开关
5. 点击"保存配置"

### 4. 开始监控

访问任意网页，服务器会实时显示：
- 📤 所有HTTP请求
- 📥 所有HTTP响应
- 🔌 WebSocket连接创建
- 🤝 WebSocket握手过程
- 📤📥 WebSocket所有消息

---

## ✨ 核心功能

### HTTP/HTTPS 请求监控

- ✅ 完整的请求URL
- ✅ HTTP方法 (GET, POST, PUT, DELETE等)
- ✅ 完整的请求头
- ✅ POST数据内容
- ✅ 响应状态码
- ✅ 完整的响应头
- ✅ MIME类型
- ✅ 资源类型分类

### WebSocket 深度监控 ⭐ 核心特性

**这是v2.0最重要的功能！**

#### 1. 连接生命周期
- 🔌 WebSocket 创建 (完整URL)
- 🤝 握手请求 (包含所有头部)
- ✅ 握手响应 (包含Sec-WebSocket-Accept等)
- 🔌 连接关闭

#### 2. 消息内容捕获
- 📤 **发送的所有消息** (完整内容)
- 📥 **接收的所有消息** (完整内容)
- 🔢 消息类型 (text/binary/ping/pong)
- 📏 消息长度统计

#### 3. 详细信息
- WebSocket完整URL (包括wss://路径)
- 请求ID关联
- Opcode类型
- Mask状态
- 时间戳

### 配置功能

- 🔧 自定义服务器地址
- 🔍 关键字过滤 (减少日志量)
- 🎚️ 一键启用/禁用
- 📊 实时状态显示
- 📈 活跃标签页和WebSocket统计

---

## 📊 使用场景

### 1. 直播平台调试
监控直播平台的WebSocket消息流：
```
- 弹幕消息
- 礼物通知
- 在线人数更新
- 直播状态变化
```

### 2. API调试
查看前端发送的所有API请求：
```
- 请求参数
- 响应内容
- 请求头信息
- 错误信息
```

### 3. WebSocket应用开发
调试实时通信应用：
```
- 聊天消息
- 游戏状态同步
- 实时数据推送
- 心跳包监控
```

### 4. 安全审计
分析网站的网络通信：
```
- 数据传输内容
- 加密情况
- 第三方请求
- 隐私数据泄露检测
```

---

## 🔬 技术架构

### CDP vs 传统方法

| 特性 | chrome.webRequest | **CDP (v2.0)** |
|------|-------------------|----------------|
| HTTP请求 | ✅ | ✅ |
| 完整请求头 | ⚠️ 部分 | ✅ 完整 |
| POST数据 | ⚠️ 有限 | ✅ 完整 |
| WebSocket创建 | ⚠️ 间接 | ✅ 直接 |
| WebSocket消息 | ❌ 不能 | ✅ **完整捕获** |
| 响应内容 | ❌ 不能 | ✅ 可获取 |

### 使用的CDP域

- `Network.enable` - 启用网络监控
- `Network.requestWillBeSent` - 请求发送前
- `Network.responseReceived` - 收到响应
- `Network.webSocketCreated` - WebSocket创建
- `Network.webSocketWillSendHandshakeRequest` - 握手请求
- `Network.webSocketHandshakeResponseReceived` - 握手响应
- `Network.webSocketFrameSent` - 发送消息
- `Network.webSocketFrameReceived` - 接收消息
- `Network.webSocketClosed` - 连接关闭
- `Network.webSocketFrameError` - 错误

---

## 📚 详细文档

### 核心文档 ⭐
- **[🎬 抖音直播快速开始](./DOUYIN_QUICK_START.md)** - 5分钟监控抖音直播 ⭐⭐⭐
- **[CDP使用指南](./CDP_USAGE.md)** - 完整的使用说明和示例
- **[CDP测试指南](./CDP_TEST.md)** - 详细的测试步骤和验证清单
- **[隐藏调试横幅](./HIDE_DEBUGGER_BANNER.md)** - 如何隐藏"正在调试此浏览器"提示

### 抖音直播相关 🎬
- **[抖音解析详细说明](./server/README_DOUYIN.md)** - 消息类型和功能说明
- **[抖音解析技术文档](./DOUYIN_PARSER_TECH.md)** - Protobuf解析原理和实现细节
- **[WebSocket日志记录](./server/README_LOGS.md)** - 日志文件自动保存和管理 ⭐

### 其他文档
- **[插件文档](./dy-live-record/brower-monitor/README.md)** - 插件技术细节
- **[服务器文档](./server/README.md)** - 服务器配置和API
- **[项目结构](./PROJECT_STRUCTURE.md)** - 完整的项目文件说明

### 历史文档
- **[v1.x 功能总结](./FEATURE_SUMMARY.md)** - 旧版本功能
- **[v1.x 测试指南](./TEST_GUIDE.md)** - 旧版本测试
- **[更新日志](./CHANGELOG.md)** - 版本历史

---

## ⚠️ 重要注意事项

### 1. Debugger权限

v2.0版本需要 `debugger` 权限来使用CDP。

**用户体验：**
- 浏览器会显示 "正在调试此浏览器"
- **这是正常现象**，因为插件使用了Chrome DevTools Protocol
- 不影响浏览器的正常使用

**如何隐藏这个提示？**
- 查看详细指南: [HIDE_DEBUGGER_BANNER.md](./HIDE_DEBUGGER_BANNER.md)
- Windows用户: 双击运行 `dy-live-record/brower-monitor/START_CHROME.bat`
- macOS/Linux用户: 运行 `./dy-live-record/brower-monitor/start-chrome.sh`

### 2. 性能影响

- CDP监控比传统方法稍微增加资源占用
- 对大多数网站，影响可忽略不计
- WebSocket消息量大时，日志会较多

**优化建议：**
- 使用过滤关键字减少不必要的监控
- 仅在需要调试时启用监控

### 3. 隐私和安全

- 插件仅在本地运行，不上传数据
- 所有数据发送到用户自己的服务器
- 建议仅在开发/测试环境使用

---

## 🧪 快速测试

### 测试 HTTP 请求

```bash
# 1. 启动服务器和插件
# 2. 访问任意网站
curl https://www.baidu.com

# 服务器会显示所有HTTP请求
```

### 测试 WebSocket

```javascript
// 在浏览器Console中执行
const ws = new WebSocket('wss://echo.websocket.org/');
ws.onopen = () => ws.send('Hello CDP!');
ws.onmessage = (e) => console.log('Received:', e.data);

// 服务器会显示：
// 🔌 WebSocket 创建
// 🤝 握手请求/响应
// 📤 发送: Hello CDP!
// 📥 接收: Hello CDP!
```

---

## 🐛 故障排查

### Q: 插件安装失败

**解决：**
1. 完全卸载旧版本
2. 清除浏览器缓存
3. 确认授予debugger权限
4. 重新加载插件

### Q: 看不到WebSocket消息

**检查：**
1. 确认版本是 v2.0.0
2. 查看Service Worker日志
3. 确认过滤关键字配置正确
4. 确认监控已启用

### Q: 服务器收不到数据

**检查：**
1. 服务器是否运行
2. 插件配置的服务器地址是否正确
3. 防火墙是否阻止8080端口
4. 使用"测试连接"功能验证

详细故障排查请参考: [CDP_TEST.md](./CDP_TEST.md)

---

## 🔄 版本升级

### 从 v1.x 升级到 v2.0

1. **完全卸载旧版本插件**
   ```
   chrome://extensions/ → 移除旧插件
   ```

2. **安装新版本**
   ```
   加载 brower-monitor 文件夹
   授予 debugger 权限
   ```

3. **更新配置**
   ```
   服务器地址保持不变
   重新配置过滤关键字（如需要）
   ```

4. **验证功能**
   ```
   查看版本是否为 v2.0.0
   测试WebSocket捕获功能
   ```

---

## 📈 性能基准

### 测试环境
- CPU: Intel i5
- RAM: 8GB
- 浏览器: Chrome 120

### 性能数据

| 场景 | 请求数 | 内存占用 | CPU占用 |
|------|--------|----------|---------|
| 简单网页 | 20-50 | +10MB | +2% |
| 复杂网页 | 100-200 | +20MB | +5% |
| WebSocket密集 | 持续 | +30MB | +8% |

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发环境设置

```bash
# 克隆项目
git clone https://github.com/WanGuChou/dy-live-record.git
cd dy-live-record

# 安装服务器依赖
cd server
npm install

# 加载插件到浏览器
# chrome://extensions/ → 加载 brower-monitor/
```

---

## 📄 许可证

MIT License

---

## 🔗 相关链接

- **Chrome DevTools Protocol**: https://chromedevtools.github.io/devtools-protocol/
- **WebSocket协议**: https://datatracker.ietf.org/doc/html/rfc6455
- **Chrome扩展开发**: https://developer.chrome.com/docs/extensions/

---

## 📞 联系方式

- **Issue**: https://github.com/WanGuChou/dy-live-record/issues
- **Email**: wangguocheng16@gmail.com

---

**版本**: v2.2.0  
**更新时间**: 2025-11-15  
**技术栈**: Chrome DevTools Protocol 1.3, WebSocket, Node.js, Protobuf, GZIP  
**核心特性**: WebSocket消息完整捕获 + 抖音直播消息解析 + 自动日志记录 🎯🎬📝
