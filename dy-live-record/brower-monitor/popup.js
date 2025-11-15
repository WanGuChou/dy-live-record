// Popup script for configuration

const serverUrlInput = document.getElementById('serverUrl');
const enableToggle = document.getElementById('enableToggle');
const saveBtn = document.getElementById('saveBtn');
const testBtn = document.getElementById('testBtn');
const statusDot = document.getElementById('statusDot');
const statusText = document.getElementById('statusText');
const messageDiv = document.getElementById('message');

// 加载配置
async function loadConfig() {
  const result = await chrome.storage.local.get(['serverUrl', 'isEnabled']);
  
  serverUrlInput.value = result.serverUrl || 'ws://localhost:8080/monitor';
  enableToggle.checked = result.isEnabled || false;
  
  updateStatus();
}

// 更新状态显示
async function updateStatus() {
  try {
    const response = await chrome.runtime.sendMessage({ action: 'getStatus' });
    
    if (response.isConnected) {
      statusDot.classList.add('connected');
      statusText.textContent = '已连接';
    } else {
      statusDot.classList.remove('connected');
      statusText.textContent = response.isEnabled ? '连接中...' : '未连接';
    }
  } catch (error) {
    console.error('获取状态失败:', error);
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
      isEnabled: isEnabled
    });
    
    // 通知background script更新配置
    await chrome.runtime.sendMessage({
      action: 'updateConfig',
      serverUrl: serverUrl,
      isEnabled: isEnabled
    });
    
    showMessage('配置已保存', 'success');
    
    // 更新状态
    setTimeout(updateStatus, 500);
  } catch (error) {
    showMessage('保存配置失败: ' + error.message, 'error');
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
  
  showMessage('正在测试连接...', 'success');
  
  try {
    const testWs = new WebSocket(serverUrl);
    
    testWs.onopen = () => {
      showMessage('连接测试成功！', 'success');
      testWs.close();
    };
    
    testWs.onerror = (error) => {
      showMessage('连接测试失败，请检查服务器地址', 'error');
      console.error('连接测试失败:', error);
    };
  } catch (error) {
    showMessage('连接测试失败: ' + error.message, 'error');
    console.error('连接测试失败:', error);
  }
}

// 事件监听
saveBtn.addEventListener('click', saveConfig);
testBtn.addEventListener('click', testConnection);
enableToggle.addEventListener('change', () => {
  if (enableToggle.checked) {
    showMessage('监控已启用', 'success');
  } else {
    showMessage('监控已禁用', 'success');
  }
});

// 定期更新状态
setInterval(updateStatus, 2000);

// 初始化
loadConfig();
