# DY Live Record - 浏览器URL和请求监控项目

## 项目概述

这是一个抖音直播录制相关的项目，包含浏览器监控插件，用于捕获和记录浏览器地址栏URL变化和所有网络请求。

## 项目结构

```
dy-live-record/
├── brower-monitor/          # Chrome/Edge 浏览器扩展插件
│   ├── manifest.json        # 扩展配置文件
│   ├── background.js        # 后台服务脚本
│   ├── popup.html          # 配置界面HTML
│   ├── popup.js            # 配置界面脚本
│   ├── icons/              # 图标文件夹
│   │   └── README.md       # 图标说明
│   ├── README.md           # 插件详细文档
│   └── QUICKSTART.md       # 快速入门指南
└── server/                  # WebSocket服务器
    ├── server.js            # 服务器主程序
    ├── package.json         # Node.js依赖配置
    ├── .gitignore          # Git忽略文件
    └── README.md           # 服务器文档
```

## 快速开始

### 1. 安装浏览器插件

详细步骤请参考：[brower-monitor/README.md](./dy-live-record/brower-monitor/README.md)

**简要步骤：**
1. 在Chrome/Edge中打开扩展管理页面
2. 开启"开发者模式"
3. 加载 `dy-live-record/brower-monitor` 文件夹
4. 配置WebSocket服务器地址
5. 启用监控

### 2. 启动WebSocket服务器

```bash
cd server
npm install
npm start
```

服务器将在 `ws://localhost:8080/monitor` 上运行

详细文档请参考：[server/README.md](./server/README.md)

## 主要功能

### 浏览器插件功能
- ✅ 监控地址栏URL变化
- ✅ 捕获所有网络请求（页面、图片、脚本、API等）
- ✅ **所有请求都打印到插件控制台日志** ⭐ 新增
- ✅ **关键字过滤功能（只发送匹配的请求）** ⭐ 新增
- ✅ 通过WebSocket实时发送数据到服务器
- ✅ 自动重连机制（断线后每5秒重连）
- ✅ 简洁的配置界面
- ✅ 实时连接状态显示

### 监控内容

#### 1. 地址栏URL变化
- 用户在地址栏输入新URL
- 点击链接导航到新页面

#### 2. 所有网络请求
- 主页面请求 (main_frame)
- 子页面请求 (sub_frame/iframe)
- CSS样式表 (stylesheet)
- JavaScript脚本 (script)
- 图片资源 (image)
- AJAX请求 (xmlhttprequest)
- WebSocket连接 (websocket)
- 媒体文件 (media)
- 其他资源类型

### WebSocket消息类型

| 类型 | 说明 | 包含信息 |
|------|------|---------|
| `connection` | 连接建立 | 状态、时间戳 |
| `url_change` | 地址栏URL变化 | URL、标题、标签页ID |
| `request` | 网络请求发起 | URL、方法、资源类型、请求ID |
| `request_completed` | 请求完成 | URL、状态码、资源类型 |

## 技术栈

### 浏览器插件
- **Manifest V3** - Chrome扩展最新标准
- **Service Worker** - 后台脚本
- **Chrome APIs** - tabs, webRequest, storage
- **WebSocket** - 实时通信

### WebSocket服务器
- **Node.js** - 运行环境
- **ws** - WebSocket库
- **支持多客户端并发连接**
- **自动清理断开的连接**
- **优雅关闭机制**

## 使用场景

1. **URL和请求监控**
   - 记录用户浏览历史和所有网络请求
   - 分析用户访问模式和API调用
   - 监控特定网站的URL和请求

2. **直播录制辅助**
   - 检测直播间URL和媒体请求
   - 自动触发录制任务
   - 捕获视频流URL

3. **网络调试和分析**
   - 实时查看所有网络请求
   - 分析API调用模式
   - 监控资源加载情况

## 数据格式示例

### URL变化
```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://example.com",
  "title": "页面标题",
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

### 网络请求
```json
{
  "type": "request",
  "requestId": "12345",
  "url": "https://example.com/api/data",
  "method": "GET",
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "frameId": 0,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

### 请求完成
```json
{
  "type": "request_completed",
  "requestId": "12345",
  "url": "https://example.com/api/data",
  "method": "GET",
  "statusCode": 200,
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

## 注意事项

⚠️ **重要：**
- 网页加载时可能产生大量请求（图片、CSS、JS、API等）
- 服务器需要能够处理高频率的消息
- 建议在服务器端进行数据过滤和存储
- 监控所有请求可能略微影响浏览器性能

## 安全说明

- 插件需要访问所有网站的权限以监控URL和请求
- 所有数据通过WebSocket发送，请确保服务器端安全
- 建议在生产环境使用 `wss://` (WebSocket Secure)
- 敏感数据应在服务器端加密存储

## 开发计划

- [ ] 添加请求过滤规则（只监控特定域名或资源类型）
- [ ] 添加数据本地存储功能
- [ ] 支持HTTP REST API作为备选通信方式
- [ ] 添加统计和分析界面
- [ ] 支持Firefox浏览器

## 文档

- **功能总结**: [FEATURE_SUMMARY.md](./FEATURE_SUMMARY.md) - 详细功能说明 ⭐
- **测试指南**: [TEST_GUIDE.md](./TEST_GUIDE.md) - 如何测试新功能 ⭐
- **使用说明**: [USAGE.md](./USAGE.md) - 快速使用指南
- **更新日志**: [CHANGELOG.md](./CHANGELOG.md) - 版本更新记录
- **插件文档**: [dy-live-record/brower-monitor/README.md](./dy-live-record/brower-monitor/README.md)
- **快速入门**: [dy-live-record/brower-monitor/QUICKSTART.md](./dy-live-record/brower-monitor/QUICKSTART.md)
- **服务器文档**: [server/README.md](./server/README.md)
- **项目结构**: [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md)

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如有问题，请在GitHub上提交Issue。

---

**最后更新时间：** 2025-11-15
