/**
 * WebSocket服务器
 * 用于接收浏览器插件发送的URL监控数据
 * 
 * 安装依赖：npm install
 * 运行服务器：npm start
 */

const WebSocket = require('ws');
const os = require('os');

// 获取本机IP地址
function getLocalIpAddress() {
  const interfaces = os.networkInterfaces();
  for (const name of Object.keys(interfaces)) {
    for (const iface of interfaces[name]) {
      if (iface.family === 'IPv4' && !iface.internal) {
        return iface.address;
      }
    }
  }
  return 'localhost';
}

// 创建WebSocket服务器，监听8080端口的/monitor路径
const wss = new WebSocket.Server({ 
  port: 8080,
  path: '/monitor'
});

const localIp = getLocalIpAddress();

console.log('='.repeat(60));
console.log('✅ WebSocket服务器已成功启动');
console.log('='.repeat(60));
console.log('');
console.log('📡 服务器信息:');
console.log(`  - 端口: 8080`);
console.log(`  - 路径: /monitor`);
console.log('');
console.log('🌐 连接地址:');
console.log(`  - 本地: ws://localhost:8080/monitor`);
console.log(`  - 局域网: ws://${localIp}:8080/monitor`);
console.log('');
console.log('💡 提示: 在浏览器插件中配置上述任一地址');
console.log('='.repeat(60));
console.log('');
console.log('⏳ 等待客户端连接...');
console.log('');

// 存储所有连接的客户端
const clients = new Set();

wss.on('connection', (ws, req) => {
  const clientIp = req.socket.remoteAddress;
  const clientPort = req.socket.remotePort;
  const userAgent = req.headers['user-agent'] || '未知';
  const origin = req.headers['origin'] || '未知';
  
  console.log('='.repeat(60));
  console.log(`[${new Date().toISOString()}] 🎉 新客户端已连接`);
  console.log('='.repeat(60));
  console.log('客户端信息:');
  console.log(`  - IP地址: ${clientIp}`);
  console.log(`  - 端口: ${clientPort}`);
  console.log(`  - User-Agent: ${userAgent.substring(0, 80)}...`);
  console.log(`  - Origin: ${origin}`);
  console.log(`  - 当前总连接数: ${wss.clients.size}`);
  console.log('');
  
  clients.add(ws);

  // 处理接收到的消息
  ws.on('message', (message) => {
    try {
      const data = JSON.parse(message.toString());
      console.log(`[${new Date().toISOString()}] 收到消息:`);
      console.log(`  类型: ${data.type}`);
      
      // 根据消息类型进行不同的处理
      switch (data.type) {
        case 'connection':
          console.log(`  状态: ${data.status}`);
          console.log('  ✅ 客户端连接确认');
          break;
          
        case 'url_change':
          console.log(`  标签页ID: ${data.tabId}`);
          console.log(`  URL: ${data.url}`);
          console.log(`  标题: ${data.title}`);
          console.log('  🔄 URL已变化');
          break;
          
        case 'tab_created':
          console.log(`  标签页ID: ${data.tabId}`);
          console.log(`  URL: ${data.url || '(空)'}`);
          console.log('  ➕ 创建了新标签页');
          break;
          
        case 'tab_closed':
          console.log(`  标签页ID: ${data.tabId}`);
          console.log('  ❌ 标签页已关闭');
          break;
          
        case 'tab_activated':
          console.log(`  标签页ID: ${data.tabId}`);
          console.log(`  URL: ${data.url}`);
          console.log(`  标题: ${data.title}`);
          console.log('  👆 标签页已激活');
          break;
          
        default:
          console.log('  ⚠️  未知消息类型');
      }
      
      console.log(`  时间戳: ${data.timestamp}`);
      console.log('-'.repeat(60));
      console.log('');
      
      // 可选：向客户端发送确认消息
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'ack',
          originalType: data.type,
          received: true,
          timestamp: new Date().toISOString()
        }));
      }
      
    } catch (error) {
      console.error(`[${new Date().toISOString()}] ❌ 解析消息失败:`, error.message);
      console.log('原始消息:', message.toString());
      console.log('');
    }
  });

  // 处理连接关闭
  ws.on('close', (code, reason) => {
    clients.delete(ws);
    console.log(`[${new Date().toISOString()}] 客户端已断开连接`);
    console.log(`  关闭代码: ${code}`);
    console.log(`  关闭原因: ${reason || '(无)'}`);
    console.log(`  当前连接数: ${wss.clients.size}`);
    console.log('');
  });

  // 处理错误
  ws.on('error', (error) => {
    console.error(`[${new Date().toISOString()}] ❌ WebSocket错误:`, error.message);
    console.log('');
  });

  // 发送欢迎消息
  try {
    const welcomeMsg = {
      type: 'welcome',
      message: '欢迎连接到URL监控服务器',
      timestamp: new Date().toISOString()
    };
    ws.send(JSON.stringify(welcomeMsg));
    console.log(`[${new Date().toISOString()}] 📤 已发送欢迎消息`);
    console.log('');
  } catch (error) {
    console.error(`[${new Date().toISOString()}] ❌ 发送欢迎消息失败:`, error.message);
  }
});

// 处理服务器错误
wss.on('error', (error) => {
  console.error('❌ 服务器错误:', error);
});

// 优雅关闭
process.on('SIGINT', () => {
  console.log('');
  console.log('正在关闭服务器...');
  
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
