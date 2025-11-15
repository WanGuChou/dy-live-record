# 使用说明

## 快速开始（3步）

### 步骤 1: 启动服务器

```bash
cd server
npm install
npm start
```

**预期输出：**
```
============================================================
WebSocket服务器已启动
地址: ws://localhost:8080/monitor
============================================================

监控内容:
  - 地址栏URL变化
  - 所有网络请求

等待客户端连接...
```

---

### 步骤 2: 安装插件

1. **Chrome**: 打开 `chrome://extensions/`
2. **Edge**: 打开 `edge://extensions/`
3. 开启 **"开发者模式"**
4. 点击 **"加载已解压的扩展程序"** (Chrome) 或 **"加载解压缩的扩展"** (Edge)
5. 选择 `dy-live-record/brower-monitor` 文件夹

---

### 步骤 3: 配置并启用

1. 点击浏览器工具栏中的 **插件图标**
2. 输入服务器地址: `ws://localhost:8080/monitor`
3. 点击 **"测试连接"** （应该显示成功）
4. 点击 **"保存配置"**
5. 开启 **"启用监控"** 开关（变为蓝色）

✅ **完成！** 现在插件会自动捕获URL和所有请求。

---

## 验证是否工作

### 在服务器终端看到：

```
[时间] ✅ 客户端连接确认

[时间] 🔄 地址栏URL变化
  URL: https://www.baidu.com/
  标题: 百度一下，你就知道
  标签页: 12345

[时间] 📡 网络请求 (主页面)
  URL: https://www.baidu.com/
  方法: GET
  标签页: 12345

[时间] ✅ 请求完成 (主页面)
  URL: https://www.baidu.com/
  状态码: 200
```

---

## 监控的内容

### 1. 地址栏URL变化
- ✅ 用户在地址栏输入URL并按回车
- ✅ 点击链接跳转
- ✅ 前端路由变化 (如 React Router)

### 2. 所有网络请求
- ✅ 页面请求 (HTML)
- ✅ API请求 (AJAX/Fetch)
- ✅ 图片、CSS、JavaScript
- ✅ WebSocket连接
- ✅ 媒体文件 (视频、音频)

**注意：** 服务器只打印主页面请求，避免日志过多。所有请求都已发送到服务器，可以在服务器端代码中自定义处理。

---

## 查看详细日志

### 浏览器插件日志

1. Chrome: `chrome://extensions/` → 找到插件 → 点击 **"Service Worker"**
2. Edge: `edge://extensions/` → 找到插件 → 点击 **"检查视图"**

**看到的日志：**
```javascript
配置已加载: {serverUrl: "ws://localhost:8080/monitor", isEnabled: true}
正在连接WebSocket: ws://localhost:8080/monitor
WebSocket连接已建立
地址栏URL变化: https://www.baidu.com/
主请求: https://www.baidu.com/
```

---

## 接收到的数据格式

### URL变化
```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://www.baidu.com/",
  "title": "百度一下，你就知道",
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

### 网络请求
```json
{
  "type": "request",
  "requestId": "12345",
  "url": "https://www.baidu.com/api/data",
  "method": "GET",
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "frameId": 0,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

### 请求完成
```json
{
  "type": "request_completed",
  "requestId": "12345",
  "url": "https://www.baidu.com/api/data",
  "method": "GET",
  "statusCode": 200,
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

---

## 常见资源类型 (resourceType)

| 类型 | 说明 |
|------|------|
| `main_frame` | 主页面 (HTML) |
| `sub_frame` | iframe |
| `stylesheet` | CSS文件 |
| `script` | JavaScript文件 |
| `image` | 图片 (jpg, png, gif等) |
| `font` | 字体文件 |
| `xmlhttprequest` | AJAX/Fetch请求 |
| `websocket` | WebSocket连接 |
| `media` | 媒体文件 (视频、音频) |
| `other` | 其他类型 |

---

## 自定义服务器处理

编辑 `server/server.js`，在消息处理部分添加自定义逻辑：

```javascript
ws.on('message', (message) => {
  const data = JSON.parse(message.toString());
  
  switch (data.type) {
    case 'url_change':
      // 处理URL变化
      console.log('用户访问了:', data.url);
      // 保存到数据库...
      break;
      
    case 'request':
      // 处理网络请求
      if (data.resourceType === 'xmlhttprequest') {
        console.log('API请求:', data.url);
        // 分析API调用...
      }
      break;
      
    case 'request_completed':
      // 处理请求完成
      if (data.statusCode >= 400) {
        console.log('请求失败:', data.url, data.statusCode);
        // 记录错误...
      }
      break;
  }
});
```

---

## 过滤特定请求

如果只想监控特定类型的请求，可以在 `background.js` 中添加过滤：

```javascript
chrome.webRequest.onBeforeRequest.addListener(
  (details) => {
    if (!isEnabled) return;
    
    // 只监控API请求
    if (details.type !== 'xmlhttprequest') {
      return;
    }
    
    // 只监控特定域名
    if (!details.url.includes('example.com')) {
      return;
    }
    
    // 发送到服务器
    sendMessage({...});
  },
  { urls: ['<all_urls>'] }
);
```

---

## 性能考虑

### 请求量
- 一个普通网页可能产生 **50-200** 个请求
- 包括图片、CSS、JS、API等
- 服务器需要能处理高频消息

### 优化建议
1. **服务器端过滤**: 只处理需要的请求类型
2. **批量处理**: 积累一定数量后批量写入数据库
3. **异步处理**: 使用消息队列 (Redis, RabbitMQ)
4. **限流**: 对频繁的请求进行限流

---

## 故障排查

### 插件不工作？
1. 检查"启用监控"开关是否打开（蓝色）
2. 查看插件Service Worker日志是否有错误
3. 确认服务器正在运行

### 服务器收不到消息？
1. 检查服务器地址是否正确: `ws://localhost:8080/monitor`
2. 确认使用 `ws://` 而不是 `wss://`
3. 重启服务器和浏览器

### 消息太多？
1. 在服务器端添加过滤逻辑
2. 只打印重要的消息类型
3. 考虑使用日志文件而不是控制台输出

---

## 下一步

- 📖 查看完整文档: [README.md](./README.md)
- 📝 查看更新日志: [CHANGELOG.md](./CHANGELOG.md)
- 🔧 查看服务器文档: [server/README.md](./server/README.md)
- 🚀 查看快速入门: [dy-live-record/brower-monitor/QUICKSTART.md](./dy-live-record/brower-monitor/QUICKSTART.md)

---

**祝使用愉快！** 🎉
