# 详细测试指南 - WebSocket和刷新捕获

## 🔧 重大改进

### 新增功能
1. ✅ **webNavigation权限** - 捕获页面导航和刷新
2. ✅ **onBeforeNavigate监听** - 捕获所有导航事件
3. ✅ **onCommitted监听** - 捕获页面提交（包括reload）
4. ✅ **onBeforeSendHeaders监听** - 专门捕获WebSocket升级请求
5. ✅ **onErrorOccurred监听** - 捕获请求错误
6. ✅ **请求计数器** - 显示捕获的请求总数

### 修复的问题
- ❌ **之前**: WebSocket连接没有被捕获
- ✅ **现在**: 通过检查`Upgrade: websocket`头部，专门捕获WS连接
- ❌ **之前**: 刷新页面可能遗漏某些请求
- ✅ **现在**: 添加了webNavigation监听器，确保捕获所有刷新事件

---

## 📋 测试步骤

### 步骤1: 重新加载插件

⚠️ **非常重要**: manifest.json已修改，必须重新加载

1. 打开 `chrome://extensions/`
2. 找到插件
3. 点击 **刷新按钮** 🔄
4. 点击 **"Service Worker"** 查看日志

**预期看到：**
```
🎯 URL & Request Monitor 已初始化
📊 版本: 1.0.1
🔍 监控内容: 所有URL变化和网络请求（包括WebSocket）
⚙️ 配置已加载: ...
```

---

### 步骤2: 测试刷新页面捕获

#### 测试A: 基本刷新
1. 访问 `https://www.baidu.com`
2. 按 **F5** 刷新
3. 查看Service Worker Console

**预期日志：**
```
🔄 页面导航: https://www.baidu.com/
  标签页: 123, 时间戳: 1234567890

🚀 页面已提交 [reload]: https://www.baidu.com/
  ✅ 匹配过滤条件，发送到服务器

📄 [1] main_frame: https://www.baidu.com/
  方法: GET
  标签页: 123
  ✅ 发送

🎨 [2] stylesheet: https://www.baidu.com/style.css
  方法: GET
  标签页: 123
  ✅ 发送

📜 [3] script: https://www.baidu.com/app.js
  方法: GET
  标签页: 123
  ✅ 发送

✅ 完成 [200] main_frame: https://www.baidu.com/
✅ 完成 [200] stylesheet: https://www.baidu.com/style.css
...
```

#### 测试B: 硬刷新
1. 按 **Ctrl+Shift+R** (或 Cmd+Shift+R)
2. 查看日志

**预期：**
- 应该看到 `transitionType: "reload"`
- 所有资源重新加载
- 包括缓存的资源

---

### 步骤3: 测试WebSocket连接捕获

#### 方法1: 使用WebSocket测试网站

访问这些网站测试：
- https://www.websocket.org/echo.html
- https://socketsbay.com/test-websockets
- 任何使用WebSocket的网站

**预期日志：**
```
📦 [10] other: wss://echo.websocket.org/
  方法: GET
  标签页: 123
  ✅ 发送

🔌🔌 WebSocket升级请求: wss://echo.websocket.org/
  标签页: 123
  ✅ 发送WebSocket升级请求
```

#### 方法2: 创建测试页面

创建 `test-websocket.html`:
```html
<!DOCTYPE html>
<html>
<head>
  <title>WebSocket测试</title>
</head>
<body>
  <h1>WebSocket连接测试</h1>
  <button onclick="connectWS()">连接WebSocket</button>
  <div id="log"></div>
  
  <script>
    function connectWS() {
      const log = document.getElementById('log');
      log.innerHTML += '<p>正在连接...</p>';
      
      const ws = new WebSocket('wss://echo.websocket.org/');
      
      ws.onopen = () => {
        log.innerHTML += '<p>✅ 已连接</p>';
        ws.send('Hello WebSocket!');
      };
      
      ws.onmessage = (e) => {
        log.innerHTML += '<p>📥 收到: ' + e.data + '</p>';
      };
      
      ws.onerror = (e) => {
        log.innerHTML += '<p>❌ 错误</p>';
      };
    }
  </script>
</body>
</html>
```

打开此页面，点击按钮，查看日志。

**预期看到：**
1. 初始页面请求
2. 🔌🔌 WebSocket升级请求

---

### 步骤4: 测试各种导航类型

#### A. 地址栏输入
```
在地址栏输入: https://www.google.com
按回车
```

**预期：**
```
🌐 地址栏URL变化: https://www.google.com/
🔄 页面导航: https://www.google.com/
🚀 页面已提交 [typed]: https://www.google.com/
```

#### B. 点击链接
```
在百度搜索结果中点击任意链接
```

**预期：**
```
🚀 页面已提交 [link]: https://...
```

#### C. 前进/后退
```
点击浏览器的后退按钮
```

**预期：**
```
🚀 页面已提交 [forward_back]: https://...
```

---

### 步骤5: 验证所有请求类型

访问一个复杂的网站（如bilibili.com），应该看到：

```
📄 [1] main_frame: https://www.bilibili.com/
📄 [2] sub_frame: https://www.bilibili.com/iframe/...
🎨 [3] stylesheet: https://www.bilibili.com/style.css
📜 [4] script: https://www.bilibili.com/app.js
🖼️ [5] image: https://www.bilibili.com/logo.png
🔤 [6] font: https://www.bilibili.com/font.woff2
🔗 [7] xmlhttprequest: https://api.bilibili.com/data
🔗 [8] fetch: https://api.bilibili.com/user
🔌 [9] websocket: wss://broadcast.chat.bilibili.com/
🎬 [10] media: https://video.bilibili.com/video.mp4
📦 [11] other: https://...
```

---

## 🔍 验证清单

### 基础功能
- [ ] 插件已重新加载
- [ ] Service Worker正常运行
- [ ] 服务器已连接
- [ ] 启用监控开关已打开

### 刷新捕获
- [ ] F5刷新能看到所有请求
- [ ] Ctrl+Shift+R硬刷新能看到所有请求
- [ ] 日志显示 `🔄 页面导航`
- [ ] 日志显示 `🚀 页面已提交 [reload]`
- [ ] 所有资源都被重新请求

### WebSocket捕获
- [ ] 能看到 `🔌🔌 WebSocket升级请求`
- [ ] WebSocket URL正确
- [ ] 包含wss://或ws://协议
- [ ] 发送到服务器

### 所有请求类型
- [ ] main_frame (主页面)
- [ ] stylesheet (CSS)
- [ ] script (JS)
- [ ] image (图片)
- [ ] font (字体)
- [ ] xmlhttprequest (AJAX)
- [ ] fetch (Fetch API)
- [ ] websocket (WebSocket)
- [ ] media (媒体)
- [ ] other (其他)

### 过滤功能
- [ ] 留空关键字时发送所有请求
- [ ] 设置关键字后只发送匹配的
- [ ] 日志正确标记 ✅发送 或 ⚠️跳过

---

## 🐛 故障排查

### Q1: 刷新页面还是看不到某些请求

**检查：**
1. 确保Service Worker在运行（不要关闭DevTools）
2. 清空Console后再刷新
3. 检查"启用监控"开关
4. 查看是否有JavaScript错误

**解决：**
```javascript
// 在Console中执行，查看状态
chrome.runtime.sendMessage({action: 'getStatus'}, console.log);
```

### Q2: WebSocket连接还是没有显示

**可能原因：**
1. 网站没有使用WebSocket
2. 使用了其他协议（如WebTransport）
3. WebSocket在iframe中（检查sub_frame请求）

**测试WebSocket：**
```javascript
// 在任意页面的Console中执行
const ws = new WebSocket('wss://echo.websocket.org/');
ws.onopen = () => console.log('WebSocket opened');
```

然后查看Service Worker日志。

### Q3: 请求计数不准确

**原因：**
- Service Worker重启会重置计数
- 这是正常的，仅用于调试

**查看当前计数：**
```javascript
chrome.runtime.sendMessage({action: 'getStatus'}, r => {
  console.log('请求总数:', r.requestCount);
});
```

---

## 📊 日志解读

### emoji含义

| Emoji | 类型 | 说明 |
|-------|------|------|
| 🌐 | URL变化 | 地址栏URL改变 |
| 🔄 | 导航 | 页面导航开始 |
| 🚀 | 提交 | 页面提交（已确定跳转） |
| 📄 | main_frame | 主页面 |
| 🖼️ | sub_frame | iframe |
| 🎨 | stylesheet | CSS |
| 📜 | script | JavaScript |
| 🖼️ | image | 图片 |
| 🔤 | font | 字体 |
| 🔗 | XHR/fetch | API请求 |
| 🔌 | websocket | WebSocket |
| 🎬 | media | 视频/音频 |
| 📦 | other | 其他 |
| 🔌🔌 | WS升级 | WebSocket升级请求 |
| ✅ | 成功 | 请求成功 |
| ❌ | 失败 | 请求失败 |
| ⚠️ | 跳过 | 不匹配过滤条件 |

### 数字含义

`[123]` = 这是第123个请求（自Service Worker启动以来）

---

## 💡 高级测试

### 测试1: 同时打开多个标签页

```
打开3个标签页，分别访问不同网站
观察日志中的 tabId
确认能区分不同标签页的请求
```

### 测试2: 测试AJAX密集的网站

```
访问: https://twitter.com 或 https://weibo.com
滚动页面加载更多内容
应该看到大量的 🔗 xmlhttprequest
```

### 测试3: 测试视频网站

```
访问: https://www.youtube.com
播放视频
应该看到:
  - 🎬 media 请求（视频片段）
  - 🔗 API请求（播放器数据）
```

---

## ✅ 测试成功标志

如果看到以下所有内容，说明插件工作完美：

1. ✅ 刷新页面时看到 `🔄 页面导航` 和 `🚀 页面已提交 [reload]`
2. ✅ 所有资源请求都打印到日志
3. ✅ WebSocket连接显示 `🔌🔌 WebSocket升级请求`
4. ✅ 每个请求有编号 `[1]`, `[2]`, `[3]`...
5. ✅ 过滤功能正常工作
6. ✅ 服务器收到所有匹配的请求

---

**更新时间**: 2025-11-15  
**版本**: 1.0.1
