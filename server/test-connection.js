#!/usr/bin/env node
/**
 * WebSocket连接测试脚本
 * 用于验证服务器是否正常工作
 * 
 * 使用方法：
 *   1. 确保服务器正在运行（npm start）
 *   2. 在新终端运行：node test-connection.js
 */

const WebSocket = require('ws');

console.log('='.repeat(60));
console.log('🧪 WebSocket连接测试');
console.log('='.repeat(60));
console.log('');

const serverUrl = 'ws://localhost:8080/monitor';
console.log(`📡 目标服务器: ${serverUrl}`);
console.log('⏳ 正在连接...');
console.log('');

const ws = new WebSocket(serverUrl);

// 超时检测
const timeout = setTimeout(() => {
  console.error('❌ 连接超时（5秒）');
  console.log('');
  console.log('可能的原因：');
  console.log('  1. 服务器未启动');
  console.log('  2. 端口8080被占用');
  console.log('  3. 防火墙阻止连接');
  console.log('');
  console.log('解决方案：');
  console.log('  - 在另一个终端运行：cd server && npm start');
  console.log('  - 检查服务器是否正常启动');
  ws.close();
  process.exit(1);
}, 5000);

ws.on('open', () => {
  clearTimeout(timeout);
  console.log('✅ 连接成功建立！');
  console.log('');
  
  // 发送测试消息
  const testMessage = {
    type: 'test',
    message: 'Hello from test script',
    timestamp: new Date().toISOString()
  };
  
  console.log('📤 发送测试消息:');
  console.log(JSON.stringify(testMessage, null, 2));
  console.log('');
  
  ws.send(JSON.stringify(testMessage));
  console.log('✅ 消息已发送');
  console.log('⏳ 等待服务器响应...');
  console.log('');
});

ws.on('message', (data) => {
  console.log('📥 收到服务器响应:');
  try {
    const parsed = JSON.parse(data.toString());
    console.log(JSON.stringify(parsed, null, 2));
  } catch (e) {
    console.log(data.toString());
  }
  console.log('');
  console.log('='.repeat(60));
  console.log('✅ 测试成功！服务器工作正常');
  console.log('='.repeat(60));
  console.log('');
  
  // 关闭连接
  setTimeout(() => {
    ws.close();
    process.exit(0);
  }, 500);
});

ws.on('error', (error) => {
  clearTimeout(timeout);
  console.error('❌ 连接失败！');
  console.log('');
  console.log('错误信息:', error.message);
  console.log('');
  console.log('可能的原因：');
  console.log('  1. 服务器未启动');
  console.log('  2. 服务器地址或端口错误');
  console.log('  3. 防火墙或网络问题');
  console.log('');
  console.log('解决方案：');
  console.log('  1. 启动服务器：cd server && npm start');
  console.log('  2. 检查服务器是否显示 "WebSocket服务器已启动"');
  console.log('  3. 确认服务器运行在 ws://localhost:8080/monitor');
  console.log('');
  process.exit(1);
});

ws.on('close', (code, reason) => {
  console.log('🔌 连接已关闭');
  if (code !== 1000) {
    console.log(`  关闭代码: ${code}`);
    console.log(`  关闭原因: ${reason || '(无)'}`);
  }
});
