// Background Service Worker - 使用 Chrome DevTools Protocol
// 监控所有网络请求和WebSocket消息

let wsConnection = null;
let serverUrl = '';
let isEnabled = false;
let filterKeywords = '';
let reconnectInterval = null;

// 存储活跃的调试会话
const activeTabs = new Map(); // tabId -> debuggee
const websockets = new Map(); // requestId -> WebSocket info

// 从存储中加载配置
async function loadConfig() {
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled', 'filterKeywords']);
  serverUrl = result.serverUrl || 'ws://localhost:8080/monitor'; // 使用新的 /monitor 路径
  isEnabled = result.isEnabled !== undefined ? result.isEnabled : false;
  filterKeywords = result.filterKeywords || 'live.douyin.com,webcast'; // 默认过滤抖音直播
  
  console.log('⚙️ CDP Monitor 配置已加载:', { serverUrl, isEnabled, filterKeywords });
  
  if (isEnabled) {
    connectWebSocket();
    await attachToAllTabs();
  }
}

// 检查URL是否匹配过滤关键字
function matchesFilter(url) {
  if (!filterKeywords || filterKeywords.trim() === '') {
    return true;
  }
  const keywords = filterKeywords.split(',').map(k => k.trim()).filter(k => k !== '');
  return keywords.some(keyword => url.includes(keyword));
}

// 连接到服务器的WebSocket
function connectWebSocket() {
  if (!serverUrl || wsConnection?.readyState === WebSocket.OPEN) {
    return;
  }

  try {
    console.log('🔌 正在连接WebSocket服务器:', serverUrl);
    wsConnection = new WebSocket(serverUrl);

    wsConnection.onopen = () => {
      console.log('✅ WebSocket服务器连接已建立');
      sendMessage({
        type: 'connection',
        status: 'connected',
        method: 'CDP',
        filterKeywords: filterKeywords,
        timestamp: new Date().toISOString()
      });
      
      if (reconnectInterval) {
        clearInterval(reconnectInterval);
        reconnectInterval = null;
      }
    };

    wsConnection.onmessage = (event) => {
      console.log('📥 收到服务器消息:', event.data);
    };

    wsConnection.onerror = (error) => {
      console.error('❌ WebSocket服务器错误:', error);
    };

    wsConnection.onclose = () => {
      console.log('🔌 WebSocket服务器连接已关闭');
      wsConnection = null;
      
      if (isEnabled && !reconnectInterval) {
        reconnectInterval = setInterval(() => {
          console.log('🔄 尝试重新连接服务器...');
          connectWebSocket();
        }, 5000);
      }
    };
  } catch (error) {
    console.error('❌ WebSocket服务器连接失败:', error);
  }
}

// 断开WebSocket连接
function disconnectWebSocket() {
  if (reconnectInterval) {
    clearInterval(reconnectInterval);
    reconnectInterval = null;
  }
  
  if (wsConnection) {
    wsConnection.close();
    wsConnection = null;
  }
}

// 发送消息到服务器
function sendMessage(data) {
  if (wsConnection?.readyState === WebSocket.OPEN) {
    try {
      wsConnection.send(JSON.stringify(data));
    } catch (error) {
      console.error('❌ 发送消息到服务器失败:', error);
    }
  }
}

// 附加调试器到标签页
async function attachDebugger(tabId) {
  if (activeTabs.has(tabId)) {
    console.log(`⚠️ 标签页 ${tabId} 已经附加调试器`);
    return;
  }

  const debuggee = { tabId: tabId };
  
  try {
    await chrome.debugger.attach(debuggee, '1.3');
    console.log(`✅ 调试器已附加到标签页 ${tabId}`);
    
    // 启用 Network 域
    await chrome.debugger.sendCommand(debuggee, 'Network.enable');
    console.log(`📡 Network 已启用 (标签页 ${tabId})`);
    
    activeTabs.set(tabId, debuggee);
  } catch (error) {
    console.error(`❌ 附加调试器失败 (标签页 ${tabId}):`, error.message);
  }
}

// 分离调试器
async function detachDebugger(tabId) {
  if (!activeTabs.has(tabId)) {
    return;
  }

  const debuggee = activeTabs.get(tabId);
  
  try {
    await chrome.debugger.detach(debuggee);
    console.log(`🔓 调试器已从标签页 ${tabId} 分离`);
  } catch (error) {
    console.error(`❌ 分离调试器失败 (标签页 ${tabId}):`, error.message);
  }
  
  activeTabs.delete(tabId);
}

// 附加到所有现有标签页
async function attachToAllTabs() {
  const tabs = await chrome.tabs.query({});
  console.log(`🔍 发现 ${tabs.length} 个标签页`);
  
  for (const tab of tabs) {
    // 过滤掉 chrome:// 和 edge:// 等特殊页面
    if (tab.url && !tab.url.startsWith('chrome://') && !tab.url.startsWith('edge://') && !tab.url.startsWith('chrome-extension://')) {
      await attachDebugger(tab.id);
    }
  }
}

// 分离所有调试器
async function detachAllDebuggers() {
  console.log(`🔓 正在分离所有调试器...`);
  const tabIds = Array.from(activeTabs.keys());
  
  for (const tabId of tabIds) {
    await detachDebugger(tabId);
  }
}

// ============ CDP 事件处理器 ============

chrome.debugger.onEvent.addListener((source, method, params) => {
  if (!isEnabled) return;
  
  const tabId = source.tabId;
  
  // Network.requestWillBeSent - 请求即将发送
  if (method === 'Network.requestWillBeSent') {
    const request = params.request;
    const requestId = params.requestId;
    
    console.log(`📤 [请求] ${request.method} ${request.url}`);
    console.log(`   RequestID: ${requestId}, TabID: ${tabId}`);
    
    const data = {
      type: 'cdp_request',
      tabId: tabId,
      requestId: requestId,
      url: request.url,
      method: request.method,
      headers: request.headers,
      postData: request.postData,
      resourceType: params.type,
      timestamp: new Date().toISOString()
    };
    
    if (matchesFilter(request.url)) {
      sendMessage(data);
    }
  }
  
  // Network.responseReceived - 收到响应
  else if (method === 'Network.responseReceived') {
    const response = params.response;
    const requestId = params.requestId;
    
    console.log(`📥 [响应] ${response.status} ${response.url}`);
    console.log(`   RequestID: ${requestId}`);
    
    const data = {
      type: 'cdp_response',
      tabId: tabId,
      requestId: requestId,
      url: response.url,
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
      mimeType: response.mimeType,
      resourceType: params.type,
      timestamp: new Date().toISOString()
    };
    
    if (matchesFilter(response.url)) {
      sendMessage(data);
    }
  }
  
  // Network.webSocketCreated - WebSocket 创建
  else if (method === 'Network.webSocketCreated') {
    const url = params.url;
    const requestId = params.requestId;
    
    console.log(`🔌 [WebSocket 创建] ${url}`);
    console.log(`   RequestID: ${requestId}, TabID: ${tabId}`);
    
    websockets.set(requestId, {
      url: url,
      tabId: tabId,
      createdAt: new Date().toISOString()
    });
    
    const data = {
      type: 'websocket_created',
      tabId: tabId,
      requestId: requestId,
      url: url,
      timestamp: new Date().toISOString()
    };
    
    if (matchesFilter(url)) {
      sendMessage(data);
    }
  }
  
  // Network.webSocketWillSendHandshakeRequest - WebSocket 握手请求
  else if (method === 'Network.webSocketWillSendHandshakeRequest') {
    const requestId = params.requestId;
    const request = params.request;
    
    console.log(`🤝 [WebSocket 握手请求]`);
    console.log(`   RequestID: ${requestId}`);
    console.log(`   Headers:`, request.headers);
    
    const wsInfo = websockets.get(requestId);
    const data = {
      type: 'websocket_handshake_request',
      tabId: tabId,
      requestId: requestId,
      url: wsInfo?.url,
      headers: request.headers,
      timestamp: new Date().toISOString()
    };
    
    if (!wsInfo || matchesFilter(wsInfo.url)) {
      sendMessage(data);
    }
  }
  
  // Network.webSocketHandshakeResponseReceived - WebSocket 握手响应
  else if (method === 'Network.webSocketHandshakeResponseReceived') {
    const requestId = params.requestId;
    const response = params.response;
    
    console.log(`✅ [WebSocket 握手响应]`);
    console.log(`   RequestID: ${requestId}`);
    console.log(`   Status: ${response.status}`);
    
    const wsInfo = websockets.get(requestId);
    const data = {
      type: 'websocket_handshake_response',
      tabId: tabId,
      requestId: requestId,
      url: wsInfo?.url,
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
      timestamp: new Date().toISOString()
    };
    
    if (!wsInfo || matchesFilter(wsInfo.url)) {
      sendMessage(data);
    }
  }
  
  // Network.webSocketFrameSent - WebSocket 发送消息
  else if (method === 'Network.webSocketFrameSent') {
    const requestId = params.requestId;
    const frame = params.response;
    
    const wsInfo = websockets.get(requestId);
    console.log(`📤 [WebSocket 发送] ${wsInfo?.url || requestId}`);
    console.log(`   Opcode: ${frame.opcode}, PayloadData: ${frame.payloadData?.substring(0, 100)}`);
    
    const data = {
      type: 'websocket_frame_sent',
      tabId: tabId,
      requestId: requestId,
      url: wsInfo?.url,
      opcode: frame.opcode,
      mask: frame.mask,
      payloadData: frame.payloadData,
      timestamp: new Date().toISOString()
    };
    
    if (!wsInfo || matchesFilter(wsInfo.url)) {
      sendMessage(data);
    }
  }
  
  // Network.webSocketFrameReceived - WebSocket 接收消息
  else if (method === 'Network.webSocketFrameReceived') {
    const requestId = params.requestId;
    const frame = params.response;
    
    const wsInfo = websockets.get(requestId);
    console.log(`📥 [WebSocket 接收] ${wsInfo?.url || requestId}`);
    console.log(`   Opcode: ${frame.opcode}, PayloadData: ${frame.payloadData?.substring(0, 100)}`);
    
    const data = {
      type: 'websocket_frame_received',
      tabId: tabId,
      requestId: requestId,
      url: wsInfo?.url,
      opcode: frame.opcode,
      mask: frame.mask,
      payloadData: frame.payloadData,
      timestamp: new Date().toISOString()
    };
    
    if (!wsInfo || matchesFilter(wsInfo.url)) {
      sendMessage(data);
    }
  }
  
  // Network.webSocketClosed - WebSocket 关闭
  else if (method === 'Network.webSocketClosed') {
    const requestId = params.requestId;
    
    const wsInfo = websockets.get(requestId);
    console.log(`🔌 [WebSocket 关闭] ${wsInfo?.url || requestId}`);
    console.log(`   RequestID: ${requestId}`);
    
    const data = {
      type: 'websocket_closed',
      tabId: tabId,
      requestId: requestId,
      url: wsInfo?.url,
      timestamp: new Date().toISOString()
    };
    
    if (!wsInfo || matchesFilter(wsInfo.url)) {
      sendMessage(data);
    }
    
    websockets.delete(requestId);
  }
  
  // Network.webSocketFrameError - WebSocket 错误
  else if (method === 'Network.webSocketFrameError') {
    const requestId = params.requestId;
    const errorMessage = params.errorMessage;
    
    const wsInfo = websockets.get(requestId);
    console.log(`❌ [WebSocket 错误] ${wsInfo?.url || requestId}`);
    console.log(`   Error: ${errorMessage}`);
    
    const data = {
      type: 'websocket_error',
      tabId: tabId,
      requestId: requestId,
      url: wsInfo?.url,
      errorMessage: errorMessage,
      timestamp: new Date().toISOString()
    };
    
    if (!wsInfo || matchesFilter(wsInfo.url)) {
      sendMessage(data);
    }
  }
});

// 监听调试器分离事件
chrome.debugger.onDetach.addListener((source, reason) => {
  const tabId = source.tabId;
  console.log(`🔓 调试器已分离 (标签页 ${tabId}), 原因: ${reason}`);
  activeTabs.delete(tabId);
});

// ============ 标签页事件监听 ============

// 新标签页创建
chrome.tabs.onCreated.addListener(async (tab) => {
  if (!isEnabled) return;
  
  console.log(`📑 新标签页创建: ${tab.id}`);
  
  // 等待标签页加载
  setTimeout(async () => {
    if (tab.url && !tab.url.startsWith('chrome://') && !tab.url.startsWith('edge://')) {
      await attachDebugger(tab.id);
    }
  }, 500);
});

// 标签页更新
chrome.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
  if (!isEnabled) return;
  
  if (changeInfo.status === 'loading' && tab.url) {
    if (!tab.url.startsWith('chrome://') && !tab.url.startsWith('edge://') && !tab.url.startsWith('chrome-extension://')) {
      if (!activeTabs.has(tabId)) {
        await attachDebugger(tabId);
      }
    }
  }
});

// 标签页关闭
chrome.tabs.onRemoved.addListener(async (tabId) => {
  console.log(`📑 标签页关闭: ${tabId}`);
  await detachDebugger(tabId);
});

// ============ 监听来自popup的消息 ============

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'updateConfig') {
    serverUrl = request.serverUrl;
    const wasEnabled = isEnabled;
    isEnabled = request.isEnabled;
    filterKeywords = request.filterKeywords || '';
    
    console.log('⚙️ 配置已更新:', { serverUrl, isEnabled, filterKeywords });
    
    if (isEnabled && !wasEnabled) {
      // 从禁用变为启用
      connectWebSocket();
      attachToAllTabs();
    } else if (!isEnabled && wasEnabled) {
      // 从启用变为禁用
      disconnectWebSocket();
      detachAllDebuggers();
    } else if (isEnabled) {
      // 保持启用状态
      connectWebSocket();
    }
    
    sendResponse({ success: true });
  } else if (request.action === 'getStatus') {
    sendResponse({
      isEnabled: isEnabled,
      isConnected: wsConnection?.readyState === WebSocket.OPEN,
      serverUrl: serverUrl,
      filterKeywords: filterKeywords,
      activeTabs: activeTabs.size,
      activeWebSockets: websockets.size
    });
  }
  
  return true;
});

// ============ 扩展生命周期 ============

chrome.runtime.onInstalled.addListener(() => {
  console.log('🔧 CDP Monitor 已安装/更新');
  loadConfig();
});

chrome.runtime.onStartup.addListener(() => {
  console.log('🚀 CDP Monitor 已启动');
  loadConfig();
});

// Service Worker 启动
console.log('🎯 CDP Network & WebSocket Monitor 已初始化');
console.log('📊 版本: 2.0.0');
console.log('🔍 使用 Chrome DevTools Protocol 监控所有请求和WebSocket消息');
loadConfig();
