# WebSocket 服务器

用于接收浏览器插件发送的URL监控数据的WebSocket服务器。

## 功能特性

- ✅ 接收浏览器插件发送的URL变化数据
- ✅ 支持多客户端并发连接
- ✅ 实时消息处理和日志输出
- ✅ 自动清理断开的连接
- ✅ 优雅关闭机制

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动服务器

```bash
# 生产模式
npm start

# 开发模式（自动重启）
npm run dev
```

服务器将在 `ws://localhost:8080/monitor` 上运行。

## 消息类型

服务器可以接收以下类型的消息：

### 1. 连接建立
```json
{
  "type": "connection",
  "status": "connected",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 2. URL变化
```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://example.com/page",
  "title": "Page Title",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 3. 新标签页创建
```json
{
  "type": "tab_created",
  "tabId": 12345,
  "url": "https://example.com",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 4. 标签页关闭
```json
{
  "type": "tab_closed",
  "tabId": 12345,
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 5. 标签页激活
```json
{
  "type": "tab_activated",
  "tabId": 12345,
  "url": "https://example.com",
  "title": "Page Title",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

## 配置

### 修改端口

编辑 `server.js` 文件，修改以下配置：

```javascript
const wss = new WebSocket.Server({ 
  port: 8080,        // 修改端口号
  path: '/monitor'   // 修改路径
});
```

### 环境变量（可选）

您可以创建 `.env` 文件来配置环境变量：

```env
PORT=8080
WS_PATH=/monitor
```

然后在代码中使用：

```javascript
const port = process.env.PORT || 8080;
const path = process.env.WS_PATH || '/monitor';
```

## 扩展功能

### 数据持久化

您可以将接收到的数据保存到数据库：

```javascript
// 安装 MySQL 客户端：npm install mysql2
const mysql = require('mysql2/promise');

const pool = mysql.createPool({
  host: 'localhost',
  user: 'root',
  password: 'password',
  database: 'url_monitor'
});

// 在消息处理中保存数据
ws.on('message', async (message) => {
  try {
    const data = JSON.parse(message.toString());
    
    if (data.type === 'url_change') {
      await pool.execute(
        'INSERT INTO url_logs (tab_id, url, title, timestamp) VALUES (?, ?, ?, ?)',
        [data.tabId, data.url, data.title, data.timestamp]
      );
    }
  } catch (error) {
    console.error('保存数据失败:', error);
  }
});
```

### 消息广播

向所有连接的客户端广播消息：

```javascript
function broadcast(message) {
  wss.clients.forEach((client) => {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(message));
    }
  });
}

// 使用示例
broadcast({
  type: 'notification',
  message: '系统通知',
  timestamp: new Date().toISOString()
});
```

### 身份验证

添加简单的token验证：

```javascript
wss.on('connection', (ws, req) => {
  const token = new URL(req.url, 'ws://localhost').searchParams.get('token');
  
  if (token !== 'your-secret-token') {
    ws.close(1008, '未授权');
    return;
  }
  
  // 继续处理连接...
});
```

## 部署到生产环境

### 使用 PM2

```bash
# 安装 PM2
npm install -g pm2

# 启动服务器
pm2 start server.js --name "url-monitor-server"

# 查看状态
pm2 status

# 查看日志
pm2 logs url-monitor-server

# 设置开机自启
pm2 startup
pm2 save
```

### 使用 Docker

创建 `Dockerfile`：

```dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install --production

COPY server.js ./

EXPOSE 8080

CMD ["npm", "start"]
```

构建和运行：

```bash
docker build -t url-monitor-server .
docker run -p 8080:8080 url-monitor-server
```

### 使用 Nginx 反向代理

Nginx 配置示例：

```nginx
upstream websocket {
    server localhost:8080;
}

server {
    listen 80;
    server_name your-domain.com;

    location /monitor {
        proxy_pass http://websocket;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 调试技巧

### 查看详细日志

所有消息都会输出到控制台，包括：
- 连接和断开事件
- 接收到的消息内容
- 错误信息

### 测试连接

使用 `wscat` 工具测试：

```bash
# 安装 wscat
npm install -g wscat

# 连接服务器
wscat -c ws://localhost:8080/monitor

# 发送测试消息
{"type":"connection","status":"connected","timestamp":"2025-11-15T10:30:00.000Z"}
```

## 性能优化

### 限制连接数

```javascript
const MAX_CLIENTS = 100;

wss.on('connection', (ws, req) => {
  if (wss.clients.size > MAX_CLIENTS) {
    ws.close(1008, '服务器连接数已满');
    return;
  }
  // 继续处理...
});
```

### 消息速率限制

```javascript
const rateLimit = new Map();

ws.on('message', (message) => {
  const now = Date.now();
  const lastTime = rateLimit.get(ws) || 0;
  
  if (now - lastTime < 100) { // 限制每100ms一条消息
    console.log('消息过于频繁，已忽略');
    return;
  }
  
  rateLimit.set(ws, now);
  // 处理消息...
});
```

## 技术栈

- **Node.js** - 运行环境
- **ws** - WebSocket库

## 许可证

MIT License

## 作者

DY Live Record Team
