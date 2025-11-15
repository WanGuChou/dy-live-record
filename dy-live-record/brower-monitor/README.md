# URL & WebSocket Monitor - 浏览器插件

一个用于监控浏览器URL变化并通过WebSocket将数据发送到服务器的Chrome/Edge扩展插件。

## 功能特性

- ✅ 实时监控所有标签页的URL变化
- ✅ 捕获标签页的创建、关闭和激活事件
- ✅ 通过WebSocket将数据实时发送到服务器
- ✅ 自动重连机制
- ✅ 简洁美观的配置界面
- ✅ 连接状态实时显示

## 目录结构

```
brower-monitor/
├── manifest.json          # 扩展配置文件
├── background.js          # 后台服务脚本（URL监控+WebSocket客户端）
├── popup.html            # 弹出窗口HTML
├── popup.js              # 弹出窗口脚本
├── icons/                # 图标文件夹
│   ├── icon16.png
│   ├── icon32.png
│   ├── icon48.png
│   └── icon128.png
└── README.md             # 说明文档
```

## 安装步骤

### 1. 准备图标文件

在 `icons` 文件夹中放置以下尺寸的图标：
- icon16.png (16x16 像素)
- icon32.png (32x32 像素)
- icon48.png (48x48 像素)
- icon128.png (128x128 像素)

### 2. 加载到浏览器

#### Chrome浏览器
1. 打开Chrome浏览器
2. 在地址栏输入 `chrome://extensions/`
3. 开启右上角的"开发者模式"
4. 点击"加载已解压的扩展程序"
5. 选择 `brower-monitor` 文件夹

#### Edge浏览器
1. 打开Edge浏览器
2. 在地址栏输入 `edge://extensions/`
3. 开启左下角的"开发人员模式"
4. 点击"加载解压缩的扩展"
5. 选择 `brower-monitor` 文件夹

## 使用方法

### 1. 配置WebSocket服务器

1. 点击浏览器工具栏中的插件图标
2. 在弹出的配置窗口中输入WebSocket服务器地址
   - 格式：`ws://your-server:port/path` 或 `wss://your-server:port/path`
   - 示例：`ws://localhost:8080/monitor`
3. 点击"测试连接"按钮验证服务器是否可访问
4. 点击"保存配置"按钮保存设置

### 2. 启用监控

1. 在配置窗口中，切换"启用监控"开关
2. 插件将开始监控所有标签页的URL变化
3. 状态指示器会显示当前连接状态

### 3. 查看状态

- **已连接**：绿色指示灯，表示已成功连接到服务器
- **连接中...**：红色指示灯闪烁，表示正在尝试连接
- **未连接**：红色指示灯，表示未连接或连接失败

## WebSocket消息格式

插件会发送以下类型的JSON消息到服务器：

### 1. 连接建立
```json
{
  "type": "connection",
  "status": "connected",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 2. URL变化
```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://example.com/page",
  "title": "Page Title",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 3. 新标签页创建
```json
{
  "type": "tab_created",
  "tabId": 12345,
  "url": "https://example.com",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 4. 标签页关闭
```json
{
  "type": "tab_closed",
  "tabId": 12345,
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

### 5. 标签页激活
```json
{
  "type": "tab_activated",
  "tabId": 12345,
  "url": "https://example.com",
  "title": "Page Title",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

## 服务器端示例

### Node.js WebSocket服务器示例

```javascript
const WebSocket = require('ws');

const wss = new WebSocket.Server({ port: 8080, path: '/monitor' });

wss.on('connection', (ws) => {
  console.log('新客户端已连接');

  ws.on('message', (message) => {
    try {
      const data = JSON.parse(message);
      console.log('收到消息:', data);
      
      // 处理不同类型的消息
      switch (data.type) {
        case 'connection':
          console.log('客户端连接确认');
          break;
        case 'url_change':
          console.log(`URL变化: ${data.url}`);
          break;
        case 'tab_created':
          console.log(`新标签页: ${data.tabId}`);
          break;
        case 'tab_closed':
          console.log(`关闭标签页: ${data.tabId}`);
          break;
        case 'tab_activated':
          console.log(`激活标签页: ${data.tabId} - ${data.url}`);
          break;
      }
    } catch (error) {
      console.error('解析消息失败:', error);
    }
  });

  ws.on('close', () => {
    console.log('客户端已断开');
  });
});

console.log('WebSocket服务器运行在 ws://localhost:8080/monitor');
```

### Spring Boot WebSocket服务器示例

```java
import org.springframework.stereotype.Component;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;
import com.fasterxml.jackson.databind.ObjectMapper;

@Component
public class MonitorWebSocketHandler extends TextWebSocketHandler {
    
    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public void afterConnectionEstablished(WebSocketSession session) {
        System.out.println("新客户端已连接: " + session.getId());
    }

    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) {
        try {
            String payload = message.getPayload();
            Map<String, Object> data = objectMapper.readValue(payload, Map.class);
            
            String type = (String) data.get("type");
            System.out.println("收到消息类型: " + type);
            
            // 处理消息
            switch (type) {
                case "url_change":
                    System.out.println("URL变化: " + data.get("url"));
                    break;
                case "tab_created":
                    System.out.println("新标签页: " + data.get("tabId"));
                    break;
                // 处理其他类型...
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
```

## 技术特性

- **Manifest V3**：使用最新的Chrome扩展标准
- **Service Worker**：后台脚本使用Service Worker模式
- **自动重连**：WebSocket断开后每5秒尝试重连
- **本地存储**：使用Chrome Storage API保存配置
- **现代UI**：响应式设计，美观的渐变色界面

## 权限说明

插件需要以下权限：
- `tabs`：访问标签页信息
- `webRequest`：监听网络请求
- `storage`：保存配置信息
- `activeTab`：访问活动标签页
- `<all_urls>`：监控所有网站的URL变化

## 调试技巧

### 1. 查看后台日志
- Chrome: `chrome://extensions/` → 点击"Service Worker"
- Edge: `edge://extensions/` → 点击"检查视图"

### 2. 查看popup日志
- 右键点击插件图标 → 选择"检查弹出内容"

### 3. 常见问题
- **无法连接服务器**：检查服务器地址是否正确，防火墙是否允许连接
- **消息未发送**：确保"启用监控"开关已打开
- **频繁断连**：检查网络状态和服务器稳定性

## 开发环境

- Manifest Version: 3
- 兼容浏览器：
  - Chrome 88+
  - Edge 88+
  - 其他基于Chromium的浏览器

## 许可证

MIT License

## 作者

DY Live Record Team

## 更新日志

### v1.0.0 (2025-11-15)
- ✨ 初始版本发布
- ✨ 实现URL监控功能
- ✨ 实现WebSocket通信
- ✨ 实现配置界面
- ✨ 实现自动重连机制
