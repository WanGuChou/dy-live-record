// Background Service Worker for URL and WebSocket Monitoring

let wsConnection = null;
let serverUrl = '';
let isEnabled = false;
let reconnectInterval = null;

// 从存储中加载配置
async function loadConfig() {
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled']);
  serverUrl = result.serverUrl || 'ws://localhost:8080/monitor';
  isEnabled = result.isEnabled !== undefined ? result.isEnabled : false;
  
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
    wsConnection = new WebSocket(serverUrl);

    wsConnection.onopen = () => {
      console.log('WebSocket连接已建立');
      sendMessage({
        type: 'connection',
        status: 'connected',
        timestamp: new Date().toISOString()
      });
      
      // 清除重连定时器
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
      
      // 如果启用状态，则尝试重连
      if (isEnabled && !reconnectInterval) {
        reconnectInterval = setInterval(() => {
          console.log('尝试重新连接WebSocket...');
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
      console.log('消息已发送:', data);
    } catch (error) {
      console.error('发送消息失败:', error);
    }
  } else {
    console.warn('WebSocket未连接，消息未发送');
  }
}

// 监听标签页URL变化
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (!isEnabled) return;

  if (changeInfo.url) {
    const urlData = {
      type: 'url_change',
      tabId: tabId,
      url: changeInfo.url,
      title: tab.title || '',
      timestamp: new Date().toISOString()
    };
    
    console.log('URL变化:', urlData);
    sendMessage(urlData);
  }
});

// 监听新标签页创建
chrome.tabs.onCreated.addListener((tab) => {
  if (!isEnabled) return;

  const tabData = {
    type: 'tab_created',
    tabId: tab.id,
    url: tab.url || '',
    timestamp: new Date().toISOString()
  };
  
  console.log('新标签页:', tabData);
  sendMessage(tabData);
});

// 监听标签页关闭
chrome.tabs.onRemoved.addListener((tabId, removeInfo) => {
  if (!isEnabled) return;

  const tabData = {
    type: 'tab_closed',
    tabId: tabId,
    timestamp: new Date().toISOString()
  };
  
  console.log('标签页关闭:', tabData);
  sendMessage(tabData);
});

// 监听标签页激活
chrome.tabs.onActivated.addListener(async (activeInfo) => {
  if (!isEnabled) return;

  try {
    const tab = await chrome.tabs.get(activeInfo.tabId);
    const tabData = {
      type: 'tab_activated',
      tabId: activeInfo.tabId,
      url: tab.url || '',
      title: tab.title || '',
      timestamp: new Date().toISOString()
    };
    
    console.log('标签页激活:', tabData);
    sendMessage(tabData);
  } catch (error) {
    console.error('获取标签页信息失败:', error);
  }
});

// 监听来自popup的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'updateConfig') {
    serverUrl = request.serverUrl;
    isEnabled = request.isEnabled;
    
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
