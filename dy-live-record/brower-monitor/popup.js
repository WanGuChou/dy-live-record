// Popup script for CDP Monitor configuration

const serverUrlInput = document.getElementById('serverUrl');
const filterKeywordsInput = document.getElementById('filterKeywords');
const enableToggle = document.getElementById('enableToggle');
const saveBtn = document.getElementById('saveBtn');
const testBtn = document.getElementById('testBtn');
const statusDot = document.getElementById('statusDot');
const statusText = document.getElementById('statusText');
const messageDiv = document.getElementById('message');
const activeTabsCount = document.getElementById('activeTabsCount');
const activeWebSocketsCount = document.getElementById('activeWebSocketsCount');

// 加载配置
async function loadConfig() {
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled', 'filterKeywords']);
  
  serverUrlInput.value = result.serverUrl || 'ws://localhost:8080/monitor';
  filterKeywordsInput.value = result.filterKeywords || '';
  enableToggle.checked = result.isEnabled || false;
  
  updateStatus();
}

// 更新状态显示
async function updateStatus() {
  try {
    const response = await chrome.runtime.sendMessage({ action: 'getStatus' });
    
    // 更新连接状态
    if (response.isConnected) {
      statusDot.classList.add('connected');
      statusText.textContent = '已连接';
    } else {
      statusDot.classList.remove('connected');
      statusText.textContent = response.isEnabled ? '连接中...' : '未连接';
    }
    
    // 更新统计信息
    if (activeTabsCount) {
      activeTabsCount.textContent = response.activeTabs || 0;
    }
    if (activeWebSocketsCount) {
      activeWebSocketsCount.textContent = response.activeWebSockets || 0;
    }
  } catch (error) {
    console.error('获取状态失败:', error);
    statusDot.classList.remove('connected');
    statusText.textContent = '错误';
    if (activeTabsCount) activeTabsCount.textContent = '?';
    if (activeWebSocketsCount) activeWebSocketsCount.textContent = '?';
  }
}

// 显示消息
function showMessage(text, type = 'success') {
  messageDiv.textContent = text;
  messageDiv.className = `message ${type}`;
  messageDiv.style.display = 'block';
  
  setTimeout(() => {
    messageDiv.style.display = 'none';
  }, 3000);
}

// 保存配置
async function saveConfig() {
  const serverUrl = serverUrlInput.value.trim();
  const filterKeywords = filterKeywordsInput.value.trim();
  const isEnabled = enableToggle.checked;
  
  if (!serverUrl) {
    showMessage('请输入服务器地址', 'error');
    return;
  }
  
  // 验证URL格式
  if (!serverUrl.startsWith('ws://') && !serverUrl.startsWith('wss://')) {
    showMessage('服务器地址必须以 ws:// 或 wss:// 开头', 'error');
    return;
  }
  
  try {
    // 保存到存储
    await chrome.storage.local.set({
      serverUrl: serverUrl,
      filterKeywords: filterKeywords,
      isEnabled: isEnabled
    });
    
    // 通知background script更新配置
    await chrome.runtime.sendMessage({
      action: 'updateConfig',
      serverUrl: serverUrl,
      filterKeywords: filterKeywords,
      isEnabled: isEnabled
    });
    
    let msg = '✅ 配置已保存';
    if (isEnabled) {
      msg += ' - 监控已启用';
      if (filterKeywords) {
        msg += ` (过滤: ${filterKeywords})`;
      }
    } else {
      msg += ' - 监控已禁用';
    }
    showMessage(msg, 'success');
    
    // 更新状态
    setTimeout(updateStatus, 500);
  } catch (error) {
    showMessage('❌ 保存配置失败: ' + error.message, 'error');
    console.error('保存配置失败:', error);
  }
}

// 测试连接
async function testConnection() {
  const serverUrl = serverUrlInput.value.trim();
  
  if (!serverUrl) {
    showMessage('请输入服务器地址', 'error');
    return;
  }
  
  if (!serverUrl.startsWith('ws://') && !serverUrl.startsWith('wss://')) {
    showMessage('服务器地址必须以 ws:// 或 wss:// 开头', 'error');
    return;
  }
  
  showMessage('🔄 正在测试连接...', 'success');
  
  try {
    const testWs = new WebSocket(serverUrl);
    
    const timeout = setTimeout(() => {
      testWs.close();
      showMessage('⏱️ 连接超时，请检查服务器是否运行', 'error');
    }, 5000);
    
    testWs.onopen = () => {
      clearTimeout(timeout);
      showMessage('✅ 连接测试成功！', 'success');
      testWs.close();
    };
    
    testWs.onerror = (error) => {
      clearTimeout(timeout);
      showMessage('❌ 连接测试失败，请检查服务器地址和端口', 'error');
      console.error('连接测试失败:', error);
    };
  } catch (error) {
    showMessage('❌ 连接测试失败: ' + error.message, 'error');
    console.error('连接测试失败:', error);
  }
}

// 事件监听
saveBtn.addEventListener('click', saveConfig);
testBtn.addEventListener('click', testConnection);
enableToggle.addEventListener('change', () => {
  if (enableToggle.checked) {
    showMessage('⚡ 监控将在保存配置后启用', 'success');
  } else {
    showMessage('⏸️ 监控将在保存配置后禁用', 'success');
  }
});

// 快捷键：Enter键保存
serverUrlInput.addEventListener('keypress', (e) => {
  if (e.key === 'Enter') {
    saveConfig();
  }
});

filterKeywordsInput.addEventListener('keypress', (e) => {
  if (e.key === 'Enter') {
    saveConfig();
  }
});

// 定期更新状态
setInterval(updateStatus, 2000);

// 初始化
loadConfig();
console.log('🔬 CDP Monitor Popup 已初始化');
