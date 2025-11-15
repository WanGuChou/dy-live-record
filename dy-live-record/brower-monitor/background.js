// Background Service Worker - 监控URL和所有网络请求

let wsConnection = null;
let serverUrl = '';
let isEnabled = false;
let filterKeywords = ''; // 过滤关键字，逗号分隔
let reconnectInterval = null;

// 从存储中加载配置
async function loadConfig() {
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled', 'filterKeywords']);
  serverUrl = result.serverUrl || 'ws://localhost:8080/monitor';
  isEnabled = result.isEnabled !== undefined ? result.isEnabled : false;
  filterKeywords = result.filterKeywords || '';
  
  console.log('配置已加载:', { serverUrl, isEnabled, filterKeywords });
  
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
    console.log('正在连接WebSocket:', serverUrl);
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

// 监听地址栏URL变化
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

// 监听所有网络请求发起
chrome.webRequest.onBeforeRequest.addListener(
  (details) => {
    if (!isEnabled) return;
    
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
    
    // 打印所有请求到控制台
    const emoji = {
      'main_frame': '📄',
      'sub_frame': '🖼️',
      'stylesheet': '🎨',
      'script': '📜',
      'image': '🖼️',
      'font': '🔤',
      'xmlhttprequest': '🔗',
      'websocket': '🔌',
      'media': '🎬',
      'other': '📦'
    };
    
    console.log(`${emoji[details.type] || '📦'} 请求 [${details.type}]:`, data.url);
    
    // 检查是否匹配过滤条件
    if (matchesFilter(data.url)) {
      console.log('  ✅ 匹配过滤条件，发送到服务器');
      sendMessage(data);
    } else {
      console.log('  ⚠️ 不匹配过滤条件，跳过发送');
    }
  },
  { urls: ['<all_urls>'] },
  ['requestBody']
);

// 监听网络请求完成
chrome.webRequest.onCompleted.addListener(
  (details) => {
    if (!isEnabled) return;
    
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
    
    // 打印完成状态
    const statusEmoji = details.statusCode >= 200 && details.statusCode < 300 ? '✅' : 
                        details.statusCode >= 400 ? '❌' : '⚠️';
    console.log(`${statusEmoji} 请求完成 [${details.statusCode}]:`, data.url);
    
    // 检查是否匹配过滤条件
    if (matchesFilter(data.url)) {
      sendMessage(data);
    }
  },
  { urls: ['<all_urls>'] },
  ['responseHeaders']
);

// 监听来自popup的消息
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
      filterKeywords: filterKeywords
    });
  }
  
  return true;
});

// 扩展安装或更新时
chrome.runtime.onInstalled.addListener(() => {
  console.log('🔧 扩展已安装/更新');
  loadConfig();
});

// 扩展启动时
chrome.runtime.onStartup.addListener(() => {
  console.log('🚀 扩展已启动');
  loadConfig();
});

// 初始化
console.log('🎯 初始化 URL & 请求监控插件');
loadConfig();
