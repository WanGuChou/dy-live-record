// Background Service Worker for URL and WebSocket Monitoring

let wsConnection = null;
let serverUrl = '';
let isEnabled = false;
let reconnectInterval = null;

// 调试日志函数
function debugLog(message, data = null) {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] [URL Monitor] ${message}`);
  if (data) {
    console.log(data);
  }
}

// 从存储中加载配置
async function loadConfig() {
  debugLog('开始加载配置...');
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled']);
  serverUrl = result.serverUrl || 'ws://localhost:8080/monitor';
  isEnabled = result.isEnabled !== undefined ? result.isEnabled : false;
  
  debugLog('配置已加载:', {
    serverUrl: serverUrl,
    isEnabled: isEnabled,
    wsState: wsConnection?.readyState
  });
  
  if (isEnabled) {
    debugLog('监控已启用，开始连接WebSocket...');
    connectWebSocket();
  } else {
    debugLog('监控未启用，跳过连接');
  }
}

// 连接WebSocket
function connectWebSocket() {
  debugLog('connectWebSocket 被调用', {
    serverUrl: serverUrl,
    currentState: wsConnection?.readyState,
    isEnabled: isEnabled
  });
  
  if (!serverUrl) {
    debugLog('❌ 错误：服务器URL为空，无法连接');
    return;
  }
  
  if (wsConnection?.readyState === WebSocket.OPEN) {
    debugLog('⚠️ WebSocket已经连接，跳过重复连接');
    return;
  }
  
  if (wsConnection?.readyState === WebSocket.CONNECTING) {
    debugLog('⚠️ WebSocket正在连接中，跳过重复连接');
    return;
  }

  try {
    debugLog('🔌 正在创建WebSocket连接...', { url: serverUrl });
    wsConnection = new WebSocket(serverUrl);

    wsConnection.onopen = () => {
      debugLog('✅ WebSocket连接成功建立！', {
        readyState: wsConnection.readyState,
        url: serverUrl
      });
      
      const connectionMsg = {
        type: 'connection',
        status: 'connected',
        timestamp: new Date().toISOString()
      };
      
      debugLog('📤 发送连接确认消息:', connectionMsg);
      sendMessage(connectionMsg);
      
      // 清除重连定时器
      if (reconnectInterval) {
        debugLog('清除重连定时器');
        clearInterval(reconnectInterval);
        reconnectInterval = null;
      }
    };

    wsConnection.onmessage = (event) => {
      debugLog('📥 收到服务器消息:', {
        data: event.data,
        type: event.type
      });
      
      try {
        const data = JSON.parse(event.data);
        debugLog('解析后的消息:', data);
      } catch (e) {
        debugLog('消息不是JSON格式:', event.data);
      }
    };

    wsConnection.onerror = (error) => {
      debugLog('❌ WebSocket错误:', {
        error: error,
        readyState: wsConnection?.readyState,
        url: serverUrl
      });
      console.error('WebSocket详细错误:', error);
    };

    wsConnection.onclose = (event) => {
      debugLog('🔌 WebSocket连接已关闭', {
        code: event.code,
        reason: event.reason,
        wasClean: event.wasClean,
        url: serverUrl
      });
      
      wsConnection = null;
      
      // 如果启用状态，则尝试重连
      if (isEnabled && !reconnectInterval) {
        debugLog('⏰ 设置5秒后自动重连');
        reconnectInterval = setInterval(() => {
          debugLog('🔄 尝试重新连接WebSocket...');
          connectWebSocket();
        }, 5000);
      }
    };
  } catch (error) {
    debugLog('❌ 创建WebSocket连接时发生异常:', error);
    console.error('WebSocket连接异常详情:', error);
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
  const state = wsConnection?.readyState;
  const stateNames = {
    0: 'CONNECTING',
    1: 'OPEN',
    2: 'CLOSING',
    3: 'CLOSED'
  };
  
  debugLog('尝试发送消息', {
    messageType: data.type,
    wsState: state !== undefined ? `${state} (${stateNames[state]})` : 'null',
    isConnected: state === WebSocket.OPEN
  });
  
  if (wsConnection?.readyState === WebSocket.OPEN) {
    try {
      const jsonData = JSON.stringify(data);
      wsConnection.send(jsonData);
      debugLog('✅ 消息发送成功:', data);
    } catch (error) {
      debugLog('❌ 发送消息时出错:', error);
      console.error('发送消息失败详情:', error);
    }
  } else {
    debugLog(`⚠️ WebSocket未连接，消息未发送 (状态: ${state !== undefined ? stateNames[state] : 'null'})`, data);
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
