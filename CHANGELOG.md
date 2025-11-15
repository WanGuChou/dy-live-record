# 更新日志

## [当前版本] - 2025-11-15

### 重大更新
- 🔄 代码已还原到 commit 690b73f
- ✨ 新增：捕获浏览器所有网络请求
- ✨ 新增：监控地址栏URL变化
- 🗑️ 移除：标签页创建、关闭、激活等监听（简化功能）

### 功能说明

#### 监控内容
1. **地址栏URL变化**
   - 用户在地址栏输入新URL
   - 点击链接导航到新页面
   
2. **所有网络请求**
   - 主页面请求 (main_frame)
   - 子页面请求 (sub_frame/iframe)
   - CSS样式表 (stylesheet)
   - JavaScript脚本 (script)
   - 图片资源 (image)
   - AJAX请求 (xmlhttprequest)
   - WebSocket连接 (websocket)
   - 媒体文件 (media)
   - 其他资源类型

#### 数据格式

**URL变化：**
```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://example.com",
  "title": "页面标题",
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

**网络请求：**
```json
{
  "type": "request",
  "requestId": "12345",
  "url": "https://example.com/api/data",
  "method": "GET",
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "frameId": 0,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

**请求完成：**
```json
{
  "type": "request_completed",
  "requestId": "12345",
  "url": "https://example.com/api/data",
  "method": "GET",
  "statusCode": 200,
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

### 服务器更新

- ✨ 新增对 `request` 和 `request_completed` 消息类型的处理
- 🎨 优化日志输出，只显示主页面请求，避免日志过多
- 📊 添加消息计数统计
- ♻️ 改进错误处理

### 性能优化

- 只在控制台输出主请求日志，避免日志过载
- 所有请求都会发送到服务器，但控制台只显示重要信息
- 服务器端可以根据需要过滤和处理不同类型的请求

### 注意事项

⚠️ **重要：**
- 网页加载时可能产生大量请求（图片、CSS、JS、API等）
- 服务器需要能够处理高频率的消息
- 建议在服务器端进行数据过滤和存储
- 监控所有请求可能略微影响浏览器性能

### 使用方法

1. **启动服务器**
   ```bash
   cd server
   npm install
   npm start
   ```

2. **配置插件**
   - 地址：`ws://localhost:8080/monitor`
   - 开启"启用监控"开关

3. **查看数据**
   - 服务器终端会显示URL变化和主请求
   - 所有请求数据都已发送到服务器
   - 可以在服务器端进行自定义处理

### 文档更新

- ✅ 更新了 README.md
- ✅ 更新了 dy-live-record/brower-monitor/README.md
- ✅ 添加了 CHANGELOG.md

---

## 历史版本

### v1.0.0 - 2025-11-15 (commit 690b73f)
- ✨ 初始版本
- ✅ 基础URL监控
- ✅ WebSocket通信
- ✅ 配置界面
- ✅ 自动重连
- ✅ 自动生成图标
