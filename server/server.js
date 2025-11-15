/**
 * WebSocket服务器
 * 用于接收浏览器插件发送的CDP监控数据
 * 包括所有网络请求和WebSocket消息
 * 特别支持抖音直播WebSocket消息自动解析
 * 
 * 安装依赖：npm install
 * 运行服务器：npm start
 */

const WebSocket = require('ws');
const douyinParser = require('./dy_ws_msg');

// 创建WebSocket服务器，监听8080端口的/monitor路径
const wss = new WebSocket.Server({ 
  port: 8080,
  path: '/monitor'
});

console.log('='.repeat(80));
console.log('CDP Monitor 服务器已启动');
console.log('地址: ws://localhost:8080/monitor');
console.log('='.repeat(80));
console.log('');
console.log('监控内容:');
console.log('  ✅ 所有 HTTP/HTTPS 请求 (使用 Chrome DevTools Protocol)');
console.log('  ✅ WebSocket 连接创建');
console.log('  ✅ WebSocket 握手过程');
console.log('  ✅ WebSocket 发送的所有消息');
console.log('  ✅ WebSocket 接收的所有消息');
console.log('  ✅ WebSocket 连接关闭');
console.log('  🎬 抖音直播消息自动解析');
console.log('');
console.log('等待客户端连接...');
console.log('');

// 存储所有连接的客户端
const clients = new Set();
let messageCount = 0;
let requestCount = 0;
let websocketCount = 0;
let douyinMessageCount = 0; // 抖音直播消息计数

// 辅助函数：截断长字符串
function truncate(str, maxLength = 500) {
  if (!str) return '';
  if (str.length <= maxLength) return str;
  return str.substring(0, maxLength) + '... [截断]';
}

// 辅助函数：格式化headers
function formatHeaders(headers) {
  if (!headers || typeof headers !== 'object') return '';
  const lines = [];
  for (const [key, value] of Object.entries(headers)) {
    lines.push(`    ${key}: ${truncate(String(value), 200)}`);
  }
  return lines.join('\n');
}

wss.on('connection', (ws, req) => {
  const clientIp = req.socket.remoteAddress;
  console.log(`╔${'═'.repeat(78)}╗`);
  console.log(`║ [${new Date().toISOString()}] 新客户端已连接`);
  console.log(`║ IP: ${clientIp}`);
  console.log(`║ 当前连接数: ${wss.clients.size}`);
  console.log(`╚${'═'.repeat(78)}╝`);
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
          console.log(`┌${'─'.repeat(78)}┐`);
          console.log(`│ [${new Date().toISOString()}] ✅ 客户端连接确认`);
          console.log(`│ 监控方法: ${data.method || 'CDP'}`);
          if (data.filterKeywords) {
            console.log(`│ 过滤关键字: ${data.filterKeywords}`);
          }
          console.log(`└${'─'.repeat(78)}┘`);
          console.log('');
          break;
          
        // ========== HTTP/HTTPS 请求 ==========
        case 'cdp_request':
          requestCount++;
          console.log(`┌${'─'.repeat(78)}┐`);
          console.log(`│ 📤 HTTP 请求 #${requestCount}`);
          console.log(`├${'─'.repeat(78)}┤`);
          console.log(`│ 方法: ${data.method}`);
          console.log(`│ URL: ${data.url}`);
          console.log(`│ 资源类型: ${data.resourceType || 'unknown'}`);
          console.log(`│ 标签页ID: ${data.tabId}`);
          console.log(`│ 请求ID: ${data.requestId}`);
          if (data.headers && Object.keys(data.headers).length > 0) {
            console.log(`│ 请求头:`);
            console.log(formatHeaders(data.headers).split('\n').map(line => `│ ${line}`).join('\n'));
          }
          if (data.postData) {
            console.log(`│ POST数据: ${truncate(data.postData, 300)}`);
          }
          console.log(`│ 时间: ${data.timestamp}`);
          console.log(`└${'─'.repeat(78)}┘`);
          console.log('');
          break;
          
        case 'cdp_response':
          console.log(`┌${'─'.repeat(78)}┐`);
          console.log(`│ 📥 HTTP 响应`);
          console.log(`├${'─'.repeat(78)}┤`);
          console.log(`│ 状态码: ${data.status} ${data.statusText || ''}`);
          console.log(`│ URL: ${data.url}`);
          console.log(`│ 资源类型: ${data.resourceType || 'unknown'}`);
          console.log(`│ MIME类型: ${data.mimeType || 'unknown'}`);
          console.log(`│ 请求ID: ${data.requestId}`);
          if (data.headers && Object.keys(data.headers).length > 0) {
            console.log(`│ 响应头:`);
            console.log(formatHeaders(data.headers).split('\n').map(line => `│ ${line}`).join('\n'));
          }
          console.log(`│ 时间: ${data.timestamp}`);
          console.log(`└${'─'.repeat(78)}┘`);
          console.log('');
          break;
          
        // ========== WebSocket 生命周期 ==========
        case 'websocket_created':
          websocketCount++;
          const isDouyinWS = data.url && douyinParser.isDouyinLiveWS(data.url);
          console.log(`╔${'═'.repeat(78)}╗`);
          console.log(`║ 🔌 WebSocket 创建 #${websocketCount}${isDouyinWS ? ' [抖音直播]' : ''}`);
          console.log(`╠${'═'.repeat(78)}╣`);
          console.log(`║ 完整URL: ${data.url}`);
          console.log(`║ 标签页ID: ${data.tabId}`);
          console.log(`║ 请求ID: ${data.requestId}`);
          if (isDouyinWS) {
            console.log(`║ ⭐ 抖音直播WebSocket，将自动解析消息内容`);
          }
          console.log(`║ 时间: ${data.timestamp}`);
          console.log(`╚${'═'.repeat(78)}╝`);
          console.log('');
          break;
          
        case 'websocket_handshake_request':
          console.log(`┌${'─'.repeat(78)}┐`);
          console.log(`│ 🤝 WebSocket 握手请求`);
          console.log(`├${'─'.repeat(78)}┤`);
          console.log(`│ URL: ${data.url || '(继承自创建事件)'}`);
          console.log(`│ 请求ID: ${data.requestId}`);
          if (data.headers && Object.keys(data.headers).length > 0) {
            console.log(`│ 握手请求头:`);
            console.log(formatHeaders(data.headers).split('\n').map(line => `│ ${line}`).join('\n'));
          }
          console.log(`│ 时间: ${data.timestamp}`);
          console.log(`└${'─'.repeat(78)}┘`);
          console.log('');
          break;
          
        case 'websocket_handshake_response':
          console.log(`┌${'─'.repeat(78)}┐`);
          console.log(`│ ✅ WebSocket 握手响应`);
          console.log(`├${'─'.repeat(78)}┤`);
          console.log(`│ 状态码: ${data.status} ${data.statusText || ''}`);
          console.log(`│ URL: ${data.url || '(继承自创建事件)'}`);
          console.log(`│ 请求ID: ${data.requestId}`);
          if (data.headers && Object.keys(data.headers).length > 0) {
            console.log(`│ 握手响应头:`);
            console.log(formatHeaders(data.headers).split('\n').map(line => `│ ${line}`).join('\n'));
          }
          console.log(`│ 时间: ${data.timestamp}`);
          console.log(`└${'─'.repeat(78)}┘`);
          console.log('');
          break;
          
        // ========== WebSocket 消息 ==========
        case 'websocket_frame_sent':
          // 检测是否为抖音直播消息
          if (data.url && douyinParser.isDouyinLiveWS(data.url)) {
            douyinMessageCount++;
            const parsed = douyinParser.parseMessage(data.payloadData, data.url);
            if (parsed) {
              const formatted = douyinParser.formatMessage(parsed);
              if (formatted) {
                console.log(formatted);
                console.log('');
                break;
              }
            }
          }
          
          // 非抖音消息或解析失败，显示原始格式
          console.log(`┌${'─'.repeat(78)}┐`);
          console.log(`│ 📤 WebSocket 发送消息`);
          console.log(`├${'─'.repeat(78)}┤`);
          console.log(`│ WebSocket URL: ${data.url || '(未知)'}`);
          console.log(`│ 请求ID: ${data.requestId}`);
          console.log(`│ Opcode: ${data.opcode} ${getOpcodeDescription(data.opcode)}`);
          console.log(`│ Mask: ${data.mask}`);
          if (data.payloadData) {
            console.log(`│ 消息内容:`);
            console.log(`│   ${truncate(data.payloadData, 1000)}`);
            console.log(`│ 消息长度: ${data.payloadData.length} 字符`);
          }
          console.log(`│ 时间: ${data.timestamp}`);
          console.log(`└${'─'.repeat(78)}┘`);
          console.log('');
          break;
          
        case 'websocket_frame_received':
          // 检测是否为抖音直播消息
          if (data.url && douyinParser.isDouyinLiveWS(data.url)) {
            douyinMessageCount++;
            const parsed = douyinParser.parseMessage(data.payloadData, data.url);
            if (parsed) {
              const formatted = douyinParser.formatMessage(parsed);
              if (formatted) {
                console.log(formatted);
                console.log('');
                break;
              }
            }
          }
          
          // 非抖音消息或解析失败，显示原始格式
          console.log(`┌${'─'.repeat(78)}┐`);
          console.log(`│ 📥 WebSocket 接收消息`);
          console.log(`├${'─'.repeat(78)}┤`);
          console.log(`│ WebSocket URL: ${data.url || '(未知)'}`);
          console.log(`│ 请求ID: ${data.requestId}`);
          console.log(`│ Opcode: ${data.opcode} ${getOpcodeDescription(data.opcode)}`);
          console.log(`│ Mask: ${data.mask}`);
          if (data.payloadData) {
            console.log(`│ 消息内容:`);
            console.log(`│   ${truncate(data.payloadData, 1000)}`);
            console.log(`│ 消息长度: ${data.payloadData.length} 字符`);
          }
          console.log(`│ 时间: ${data.timestamp}`);
          console.log(`└${'─'.repeat(78)}┘`);
          console.log('');
          break;
          
        case 'websocket_closed':
          console.log(`╔${'═'.repeat(78)}╗`);
          console.log(`║ 🔌 WebSocket 已关闭`);
          console.log(`╠${'═'.repeat(78)}╣`);
          console.log(`║ WebSocket URL: ${data.url || '(未知)'}`);
          console.log(`║ 请求ID: ${data.requestId}`);
          console.log(`║ 时间: ${data.timestamp}`);
          console.log(`╚${'═'.repeat(78)}╝`);
          console.log('');
          break;
          
        case 'websocket_error':
          console.log(`╔${'═'.repeat(78)}╗`);
          console.log(`║ ❌ WebSocket 错误`);
          console.log(`╠${'═'.repeat(78)}╣`);
          console.log(`║ WebSocket URL: ${data.url || '(未知)'}`);
          console.log(`║ 请求ID: ${data.requestId}`);
          console.log(`║ 错误消息: ${data.errorMessage}`);
          console.log(`║ 时间: ${data.timestamp}`);
          console.log(`╚${'═'.repeat(78)}╝`);
          console.log('');
          break;
          
        default:
          console.log(`⚠️  未知消息类型: ${data.type}`);
          console.log('完整消息:', JSON.stringify(data, null, 2).substring(0, 500));
          console.log('');
      }
      
      // 每50条消息显示一次统计
      if (messageCount % 50 === 0) {
        console.log(`╔${'═'.repeat(78)}╗`);
        console.log(`║ 📊 统计信息`);
        console.log(`╠${'═'.repeat(78)}╣`);
        console.log(`║ 总消息数: ${messageCount}`);
        console.log(`║ HTTP请求数: ${requestCount}`);
        console.log(`║ WebSocket连接数: ${websocketCount}`);
        console.log(`║ 抖音直播消息: ${douyinMessageCount}`);
        console.log(`╚${'═'.repeat(78)}╝`);
        console.log('');
        
        // 如果有抖音消息，显示抖音统计
        if (douyinMessageCount > 0) {
          console.log(douyinParser.formatStatistics());
          console.log('');
        }
      }
      
    } catch (error) {
      console.error(`❌ 解析消息失败:`, error.message);
      console.log('原始消息:', message.toString().substring(0, 500));
      console.log('');
    }
  });

  // 处理连接关闭
  ws.on('close', (code, reason) => {
    clients.delete(ws);
    console.log(`╔${'═'.repeat(78)}╗`);
    console.log(`║ [${new Date().toISOString()}] 客户端已断开连接`);
    console.log(`║ 关闭代码: ${code}`);
    console.log(`║ 原因: ${reason || '(无)'}`);
    console.log(`║ 当前连接数: ${wss.clients.size}`);
    console.log(`╚${'═'.repeat(78)}╝`);
    console.log('');
  });

  // 处理错误
  ws.on('error', (error) => {
    console.error(`❌ WebSocket错误:`, error.message);
    console.log('');
  });

  // 发送欢迎消息
  ws.send(JSON.stringify({
    type: 'welcome',
    message: '欢迎连接到CDP监控服务器（支持抖音直播解析）',
    timestamp: new Date().toISOString()
  }));
});

// WebSocket Opcode 说明
function getOpcodeDescription(opcode) {
  const opcodes = {
    0: '(continuation frame)',
    1: '(text frame)',
    2: '(binary frame)',
    8: '(connection close)',
    9: '(ping)',
    10: '(pong)'
  };
  return opcodes[opcode] || '(unknown)';
}

// 处理服务器错误
wss.on('error', (error) => {
  console.error('❌ 服务器错误:', error);
});

// 优雅关闭
process.on('SIGINT', () => {
  console.log('');
  console.log('╔═════════════════════════════════════════╗');
  console.log('║ 正在关闭服务器...                      ║');
  console.log('╠═════════════════════════════════════════╣');
  console.log(`║ 总消息数: ${messageCount.toString().padEnd(28)} ║`);
  console.log(`║ HTTP请求数: ${requestCount.toString().padEnd(26)} ║`);
  console.log(`║ WebSocket连接数: ${websocketCount.toString().padEnd(22)} ║`);
  console.log(`║ 抖音直播消息: ${douyinMessageCount.toString().padEnd(24)} ║`);
  console.log('╚═════════════════════════════════════════╝');
  
  // 显示抖音直播统计
  if (douyinMessageCount > 0) {
    console.log('');
    console.log(douyinParser.formatStatistics());
  }
  
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
