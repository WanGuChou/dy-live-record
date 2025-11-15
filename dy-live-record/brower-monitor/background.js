// Background Service Worker - 监控URL和所有网络请求

let wsConnection = null;
let serverUrl = '';
let isEnabled = false;
let reconnectInterval = null;

// 从存储中加载配置
async function loadConfig() {
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled']);
  serverUrl = result.serverUrl || 'ws://localhost:8080/monitor';
  isEnabled = result.isEnabled !== undefined ? result.isEnabled : false;
  
  console.log('配置已加载:', { serverUrl, isEnabled });
  
  if (isEnabled) {
    connectWebSocket();
  }
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
      console.log('WebSocket连接已建立');
      sendMessage({
        type: 'connection',
        status: 'connected',
        timestamp: new Date().toISOString()
      });
      
      if (reconnectInterval) {
        clearInterval(reconnectInterval);
        reconnectInterval = null;
      }
    };

    wsConnection.onmessage = (event) => {
      console.log('收到服务器消息:', event.data);
    };

    wsConnection.onerror = (error) => {
      console.error('WebSocket错误:', error);
    };

    wsConnection.onclose = () => {
      console.log('WebSocket连接已关闭');
      wsConnection = null;
      
      if (isEnabled && !reconnectInterval) {
        reconnectInterval = setInterval(() => {
          console.log('尝试重新连接...');
          connectWebSocket();
        }, 5000);
      }
    };
  } catch (error) {
    console.error('WebSocket连接失败:', error);
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
      console.error('发送消息失败:', error);
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
    
    console.log('地址栏URL变化:', data.url);
    sendMessage(data);
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
    
    // 只在控制台简要输出，避免日志过多
    if (details.type === 'main_frame') {
      console.log('主请求:', data.url);
    }
    
    sendMessage(data);
  },
  { urls: ['<all_urls>'] },
  ['requestBody']
);

// 监听网络请求完成（可选，获取响应状态）
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
    
    sendMessage(data);
  },
  { urls: ['<all_urls>'] }
);

// 监听来自popup的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'updateConfig') {
    serverUrl = request.serverUrl;
    isEnabled = request.isEnabled;
    
    console.log('配置已更新:', { serverUrl, isEnabled });
    
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
      serverUrl: serverUrl
    });
  }
  
  return true;
});

// 扩展安装或更新时
chrome.runtime.onInstalled.addListener(() => {
  console.log('扩展已安装/更新');
  loadConfig();
});

// 扩展启动时
chrome.runtime.onStartup.addListener(() => {
  console.log('扩展已启动');
  loadConfig();
});

// 初始化
loadConfig();
