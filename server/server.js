/**
 * WebSocket服务器
 * 用于接收浏览器插件发送的URL和请求监控数据
 * 
 * 安装依赖：npm install
 * 运行服务器：npm start
 */

const WebSocket = require('ws');

// 创建WebSocket服务器，监听8080端口的/monitor路径
const wss = new WebSocket.Server({ 
  port: 8080,
  path: '/monitor'
});

console.log('='.repeat(60));
console.log('WebSocket服务器已启动');
console.log('地址: ws://localhost:8080/monitor');
console.log('='.repeat(60));
console.log('');
console.log('监控内容:');
console.log('  - 地址栏URL变化');
console.log('  - 所有网络请求');
console.log('');
console.log('等待客户端连接...');
console.log('');

// 存储所有连接的客户端
const clients = new Set();
let messageCount = 0;

wss.on('connection', (ws, req) => {
  const clientIp = req.socket.remoteAddress;
  console.log(`[${new Date().toISOString()}] 新客户端已连接 (IP: ${clientIp})`);
  console.log(`当前连接数: ${wss.clients.size}`);
  console.log('');
  
  clients.add(ws);

  // 处理接收到的消息
  ws.on('message', (message) => {
    try {
      const data = JSON.parse(message.toString());
      messageCount++;
      
      // 根据消息类型进行不同的处理
      switch (data.type) {
        case 'connection':
          console.log(`[${new Date().toISOString()}] ✅ 客户端连接确认`);
          console.log('');
          break;
          
        case 'url_change':
          console.log(`[${new Date().toISOString()}] 🔄 地址栏URL变化`);
          console.log(`  URL: ${data.url}`);
          console.log(`  标题: ${data.title}`);
          console.log(`  标签页: ${data.tabId}`);
          console.log('');
          break;
          
        case 'request':
          // 网络请求，只输出主请求，避免日志过多
          if (data.resourceType === 'main_frame') {
            console.log(`[${new Date().toISOString()}] 📡 网络请求 (主页面)`);
            console.log(`  URL: ${data.url}`);
            console.log(`  方法: ${data.method}`);
            console.log(`  标签页: ${data.tabId}`);
            console.log('');
          }
          // 子资源请求不打印，但已接收并可处理
          break;
          
        case 'request_completed':
          // 请求完成，只输出主请求
          if (data.resourceType === 'main_frame') {
            console.log(`[${new Date().toISOString()}] ✅ 请求完成 (主页面)`);
            console.log(`  URL: ${data.url}`);
            console.log(`  状态码: ${data.statusCode}`);
            console.log('');
          }
          break;
          
        default:
          console.log(`[${new Date().toISOString()}] ⚠️  未知消息类型: ${data.type}`);
          console.log('');
      }
      
      // 每100条消息显示一次统计
      if (messageCount % 100 === 0) {
        console.log(`📊 已接收 ${messageCount} 条消息`);
        console.log('');
      }
      
    } catch (error) {
      console.error(`[${new Date().toISOString()}] ❌ 解析消息失败:`, error.message);
      console.log('原始消息:', message.toString().substring(0, 200));
      console.log('');
    }
  });

  // 处理连接关闭
  ws.on('close', (code, reason) => {
    clients.delete(ws);
    console.log(`[${new Date().toISOString()}] 客户端已断开连接`);
    console.log(`  关闭代码: ${code}`);
    console.log(`  当前连接数: ${wss.clients.size}`);
    console.log('');
  });

  // 处理错误
  ws.on('error', (error) => {
    console.error(`[${new Date().toISOString()}] ❌ WebSocket错误:`, error.message);
    console.log('');
  });

  // 发送欢迎消息
  ws.send(JSON.stringify({
    type: 'welcome',
    message: '欢迎连接到URL和请求监控服务器',
    timestamp: new Date().toISOString()
  }));
});

// 处理服务器错误
wss.on('error', (error) => {
  console.error('❌ 服务器错误:', error);
});

// 优雅关闭
process.on('SIGINT', () => {
  console.log('');
  console.log('正在关闭服务器...');
  console.log(`总共接收了 ${messageCount} 条消息`);
  
  // 关闭所有客户端连接
  wss.clients.forEach((client) => {
    client.close(1000, '服务器正在关闭');
  });
  
  wss.close(() => {
    console.log('服务器已关闭');
    process.exit(0);
  });
});

// 定期清理断开的连接
setInterval(() => {
  wss.clients.forEach((client) => {
    if (client.readyState === WebSocket.CLOSED) {
      clients.delete(client);
    }
  });
}, 30000);
