// 配置文件
const CONFIG = {
  // WebSocket 服务器地址
  SERVER_URL: 'ws://localhost:8080/monitor',
  
  // 重连配置
  RECONNECT_INTERVAL: 3000, // 3秒
  MAX_RECONNECT_ATTEMPTS: 10,
  
  // 监控配置
  MONITOR_DOUYIN: true, // 是否监控抖音
  
  // 过滤配置
  URL_FILTERS: [
    'live.douyin.com',
    'webcast'
  ],
  
  // 调试模式
  DEBUG: true
};
