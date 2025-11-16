# URL & 请求监控 - 浏览器插件

监控浏览器地址栏URL变化和所有网络请求，并通过WebSocket发送到服务器。

## 功能

- ✅ 监控地址栏URL变化
- ✅ 捕获所有网络请求（包括AJAX、图片、脚本等）
- ✅ 实时通过WebSocket发送到服务器
- ✅ 自动重连机制
- ✅ 简洁的配置界面

## 安装

### 1. 加载插件到浏览器

**Chrome浏览器：**
1. 打开 `chrome://extensions/`
2. 开启"开发者模式"
3. 点击"加载已解压的扩展程序"
4. 选择 `brower-monitor` 文件夹

**Edge浏览器：**
1. 打开 `edge://extensions/`
2. 开启"开发人员模式"
3. 点击"加载解压缩的扩展"
4. 选择 `brower-monitor` 文件夹

### 2. 配置插件

1. 点击浏览器工具栏中的插件图标
2. 输入WebSocket服务器地址：`ws://localhost:8080/monitor`
3. 点击"测试连接"
4. 点击"保存配置"
5. 开启"启用监控"开关

## 使用

### 启动服务器

```bash
# 启动 Go 后端服务
cd ../server-go
.\dy-live-monitor.exe
```

服务器将在 `ws://localhost:8080` 运行。

### 监控数据

插件会自动发送以下数据到服务器：

#### 1. 地址栏URL变化
当用户在地址栏输入新URL或点击链接时触发。

```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://example.com",
  "title": "页面标题",
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

#### 2. 网络请求发起
浏览器发起的所有请求（页面、图片、脚本、API等）。

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

**resourceType 类型：**
- `main_frame`: 主页面
- `sub_frame`: iframe
- `stylesheet`: CSS文件
- `script`: JavaScript文件
- `image`: 图片
- `xmlhttprequest`: AJAX请求
- `websocket`: WebSocket连接
- 等等...

#### 3. 请求完成
请求完成后的响应状态。

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

## 服务器端示例

查看 `../../server/` 目录中的完整服务器实现。

### 简单示例

```javascript
const WebSocket = require('ws');

const wss = new WebSocket.Server({ port: 8080, path: '/monitor' });

wss.on('connection', (ws) => {
  console.log('客户端已连接');

  ws.on('message', (message) => {
    const data = JSON.parse(message);
    
    switch(data.type) {
      case 'url_change':
        console.log('URL变化:', data.url);
        // 处理URL变化...
        break;
        
      case 'request':
        console.log('网络请求:', data.url);
        // 处理网络请求...
        break;
        
      case 'request_completed':
        console.log('请求完成:', data.url, '状态码:', data.statusCode);
        // 处理请求完成...
        break;
    }
  });
});
```

## 注意事项

- **请求量大**：网页可能产生大量请求（图片、CSS、JS等），服务器需要能处理高频消息
- **仅主请求**：服务器端示例只打印主页面请求，避免日志过多
- **性能影响**：监控所有请求可能略微影响浏览器性能
- **数据存储**：建议在服务器端对数据进行过滤和存储

## 权限说明

插件需要以下权限：
- `tabs`：访问标签页信息
- `webRequest`：监听网络请求
- `storage`：保存配置
- `activeTab`：访问活动标签页
- `<all_urls>`：监控所有网站

## 调试

查看插件日志：
- Chrome: `chrome://extensions/` → 点击"Service Worker"
- Edge: `edge://extensions/` → 点击"检查视图"

## 许可证

MIT License
