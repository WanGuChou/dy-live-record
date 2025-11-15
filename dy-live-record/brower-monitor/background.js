// Background Service Worker - 监控URL和所有网络请求
// 包括WebSocket连接、刷新页面等所有场景

let wsConnection = null;
let serverUrl = '';
let isEnabled = false;
let filterKeywords = ''; // 过滤关键字，逗号分隔
let reconnectInterval = null;
let requestCount = 0; // 请求计数

// 从存储中加载配置
async function loadConfig() {
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled', 'filterKeywords']);
  serverUrl = result.serverUrl || 'ws://localhost:8080/monitor';
  isEnabled = result.isEnabled !== undefined ? result.isEnabled : false;
  filterKeywords = result.filterKeywords || '';
  
  console.log('⚙️ 配置已加载:', { serverUrl, isEnabled, filterKeywords });
  
  if (isEnabled) {
    connectWebSocket();
  }
}

// 检查URL是否匹配过滤关键字
function matchesFilter(url) {
  // 如果没有设置过滤关键字，全部发送
  if (!filterKeywords || filterKeywords.trim() === '') {
    return true;
  }
  
  // 分割关键字并检查
  const keywords = filterKeywords.split(',').map(k => k.trim()).filter(k => k !== '');
  
  // 只要匹配任一关键字就发送
  return keywords.some(keyword => url.includes(keyword));
}

// 连接WebSocket
function connectWebSocket() {
  if (!serverUrl || wsConnection?.readyState === WebSocket.OPEN) {
    return;
  }

  try {
    console.log('🔌 正在连接WebSocket:', serverUrl);
    wsConnection = new WebSocket(serverUrl);

    wsConnection.onopen = () => {
      console.log('✅ WebSocket连接已建立');
      sendMessage({
        type: 'connection',
        status: 'connected',
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
      console.error('❌ WebSocket错误:', error);
    };

    wsConnection.onclose = () => {
      console.log('🔌 WebSocket连接已关闭');
      wsConnection = null;
      
      if (isEnabled && !reconnectInterval) {
        reconnectInterval = setInterval(() => {
          console.log('🔄 尝试重新连接...');
          connectWebSocket();
        }, 5000);
      }
    };
  } catch (error) {
    console.error('❌ WebSocket连接失败:', error);
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
      console.error('❌ 发送消息失败:', error);
    }
  }
}

// emoji图标映射
const emojiMap = {
  'main_frame': '📄',
  'sub_frame': '🖼️',
  'stylesheet': '🎨',
  'script': '📜',
  'image': '🖼️',
  'font': '🔤',
  'xmlhttprequest': '🔗',
  'fetch': '🔗',
  'websocket': '🔌',
  'webtransport': '🚄',
  'media': '🎬',
  'object': '📦',
  'ping': '📡',
  'csp_report': '🛡️',
  'other': '📦'
};

// 打印请求日志
function logRequest(type, url, details = {}) {
  requestCount++;
  const emoji = emojiMap[type] || '📦';
  console.log(`${emoji} [${requestCount}] ${type}: ${url}`);
  
  // 打印额外信息
  if (details.method) {
    console.log(`  方法: ${details.method}`);
  }
  if (details.statusCode) {
    console.log(`  状态码: ${details.statusCode}`);
  }
  if (details.tabId >= 0) {
    console.log(`  标签页: ${details.tabId}`);
  }
}

// ============ 监听地址栏URL变化 ============
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (!isEnabled) return;

  if (changeInfo.url) {
    const data = {
      type: 'url_change',
      tabId: tabId,
      url: changeInfo.url,
      title: tab.title || '',
      timestamp: new Date().toISOString()
    };
    
    console.log('🌐 地址栏URL变化:', data.url);
    
    // 检查是否匹配过滤条件
    if (matchesFilter(data.url)) {
      console.log('  ✅ 匹配过滤条件，发送到服务器');
      sendMessage(data);
    } else {
      console.log('  ⚠️ 不匹配过滤条件，跳过发送');
    }
  }
});

// ============ 监听页面导航（捕获刷新等） ============
chrome.webNavigation.onBeforeNavigate.addListener((details) => {
  if (!isEnabled) return;
  
  // frameId === 0 表示主框架（不是iframe）
  if (details.frameId === 0) {
    console.log('🔄 页面导航:', details.url);
    console.log(`  标签页: ${details.tabId}, 时间戳: ${details.timeStamp}`);
  }
});

// ============ 监听页面提交（表单提交、刷新确认） ============
chrome.webNavigation.onCommitted.addListener((details) => {
  if (!isEnabled) return;
  
  if (details.frameId === 0) {
    console.log(`🚀 页面已提交 [${details.transitionType}]:`, details.url);
    
    // transitionType可能是: reload, typed, link, auto_bookmark等
    const data = {
      type: 'navigation_committed',
      tabId: details.tabId,
      url: details.url,
      transitionType: details.transitionType,
      transitionQualifiers: details.transitionQualifiers,
      timestamp: new Date().toISOString()
    };
    
    if (matchesFilter(data.url)) {
      sendMessage(data);
    }
  }
});

// ============ 监听所有网络请求发起 ============
chrome.webRequest.onBeforeRequest.addListener(
  (details) => {
    if (!isEnabled) return;
    
    // 打印所有请求到控制台
    logRequest(details.type, details.url, {
      method: details.method,
      tabId: details.tabId
    });
    
    const data = {
      type: 'request',
      requestId: details.requestId,
      url: details.url,
      method: details.method,
      resourceType: details.type,
      tabId: details.tabId,
      frameId: details.frameId,
      timestamp: new Date().toISOString()
    };
    
    // 检查是否匹配过滤条件
    if (matchesFilter(data.url)) {
      console.log('  ✅ 发送');
      sendMessage(data);
    } else {
      console.log('  ⚠️ 跳过');
    }
  },
  { urls: ['<all_urls>'] }
);

// ============ 监听请求发送头部（捕获WebSocket升级） ============
chrome.webRequest.onBeforeSendHeaders.addListener(
  (details) => {
    if (!isEnabled) return;
    
    // 检查是否是WebSocket升级请求
    const headers = details.requestHeaders || [];
    const upgradeHeader = headers.find(h => h.name.toLowerCase() === 'upgrade');
    
    if (upgradeHeader && upgradeHeader.value.toLowerCase() === 'websocket') {
      console.log('🔌🔌 WebSocket升级请求:', details.url);
      console.log(`  标签页: ${details.tabId}`);
      
      const data = {
        type: 'websocket_upgrade',
        requestId: details.requestId,
        url: details.url,
        method: details.method,
        tabId: details.tabId,
        timestamp: new Date().toISOString()
      };
      
      if (matchesFilter(data.url)) {
        console.log('  ✅ 发送WebSocket升级请求');
        sendMessage(data);
      }
    }
  },
  { urls: ['<all_urls>'] },
  ['requestHeaders']
);

// ============ 监听网络请求完成 ============
chrome.webRequest.onCompleted.addListener(
  (details) => {
    if (!isEnabled) return;
    
    const statusEmoji = details.statusCode >= 200 && details.statusCode < 300 ? '✅' : 
                        details.statusCode >= 400 ? '❌' : '⚠️';
    console.log(`${statusEmoji} 完成 [${details.statusCode}] ${details.type}: ${details.url}`);
    
    const data = {
      type: 'request_completed',
      requestId: details.requestId,
      url: details.url,
      method: details.method,
      statusCode: details.statusCode,
      resourceType: details.type,
      tabId: details.tabId,
      timestamp: new Date().toISOString()
    };
    
    // 检查是否匹配过滤条件
    if (matchesFilter(data.url)) {
      sendMessage(data);
    }
  },
  { urls: ['<all_urls>'] }
);

// ============ 监听请求错误 ============
chrome.webRequest.onErrorOccurred.addListener(
  (details) => {
    if (!isEnabled) return;
    
    console.log(`❌ 请求错误 [${details.error}]:`, details.url);
    
    const data = {
      type: 'request_error',
      requestId: details.requestId,
      url: details.url,
      method: details.method,
      error: details.error,
      resourceType: details.type,
      tabId: details.tabId,
      timestamp: new Date().toISOString()
    };
    
    if (matchesFilter(data.url)) {
      sendMessage(data);
    }
  },
  { urls: ['<all_urls>'] }
);

// ============ 监听来自popup的消息 ============
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'updateConfig') {
    serverUrl = request.serverUrl;
    isEnabled = request.isEnabled;
    filterKeywords = request.filterKeywords || '';
    
    console.log('⚙️ 配置已更新:', { serverUrl, isEnabled, filterKeywords });
    
    if (isEnabled) {
      connectWebSocket();
    } else {
      disconnectWebSocket();
    }
    
    sendResponse({ success: true });
  } else if (request.action === 'getStatus') {
    sendResponse({
      isEnabled: isEnabled,
      isConnected: wsConnection?.readyState === WebSocket.OPEN,
      serverUrl: serverUrl,
      filterKeywords: filterKeywords,
      requestCount: requestCount
    });
  }
  
  return true;
});

// ============ 扩展安装或更新时 ============
chrome.runtime.onInstalled.addListener(() => {
  console.log('🔧 扩展已安装/更新');
  requestCount = 0;
  loadConfig();
});

// ============ 扩展启动时 ============
chrome.runtime.onStartup.addListener(() => {
  console.log('🚀 扩展已启动');
  requestCount = 0;
  loadConfig();
});

// ============ Service Worker启动时 ============
console.log('🎯 URL & Request Monitor 已初始化');
console.log('📊 版本: 1.0.1');
console.log('🔍 监控内容: 所有URL变化和网络请求（包括WebSocket）');
loadConfig();
