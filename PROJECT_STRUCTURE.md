# 项目结构说明

## 完整目录树

```
/workspace/
├── README.md                                    # 项目总览文档
├── PROJECT_STRUCTURE.md                         # 本文件：项目结构说明
│
├── dy-live-record/                              # 直播录制项目主目录
│   └── brower-monitor/                          # 浏览器监控插件
│       ├── manifest.json                        # Chrome扩展配置文件（Manifest V3）
│       ├── background.js                        # 后台Service Worker脚本
│       ├── popup.html                           # 插件配置界面HTML
│       ├── popup.js                             # 插件配置界面脚本
│       ├── .gitignore                           # Git忽略文件
│       ├── README.md                            # 插件详细文档
│       ├── QUICKSTART.md                        # 5分钟快速入门指南
│       └── icons/                               # 插件图标目录
│           └── README.md                        # 图标使用说明
│
└── server/                                      # WebSocket服务器（独立目录）
    ├── server.js                                # WebSocket服务器主程序
    ├── package.json                             # Node.js依赖配置
    ├── .gitignore                               # Git忽略文件
    └── README.md                                # 服务器详细文档
```

## 目录说明

### 1. `/workspace/` - 项目根目录
- **README.md**: 项目整体说明文档，包含快速开始指南
- **PROJECT_STRUCTURE.md**: 本文件，详细的项目结构说明

### 2. `/workspace/dy-live-record/brower-monitor/` - 浏览器插件
Chrome/Edge浏览器扩展插件，用于监控URL变化并发送到服务器。

**核心文件：**
- `manifest.json`: 扩展配置，定义权限、background脚本、popup等
- `background.js`: 后台脚本，负责URL监控和WebSocket通信
- `popup.html/js`: 用户配置界面，用于设置服务器地址和启用/禁用监控

**文档：**
- `README.md`: 完整的插件文档，包含安装、使用、消息格式等
- `QUICKSTART.md`: 快速入门指南，5分钟上手

**图标：**
- `icons/`: 存放不同尺寸的插件图标（16x16、32x32、48x48、128x128）

### 3. `/workspace/server/` - WebSocket服务器
独立的Node.js WebSocket服务器，接收浏览器插件发送的数据。

**核心文件：**
- `server.js`: WebSocket服务器主程序，处理连接和消息
- `package.json`: 依赖配置（ws库）

**文档：**
- `README.md`: 服务器详细文档，包含配置、扩展、部署等

## 组件关系

```
┌─────────────────┐         WebSocket          ┌─────────────────┐
│  浏览器插件      │ ═══════════════════════> │  WebSocket服务器 │
│  (brower-monitor)│    ws://localhost:8080    │  (server/)       │
│                  │         /monitor           │                  │
│  - 监控URL变化   │                            │  - 接收消息      │
│  - 发送消息      │                            │  - 处理数据      │
│  - 自动重连      │ <═══════════════════════ │  - 日志输出      │
└─────────────────┘      确认消息(ACK)         └─────────────────┘
```

## 数据流

1. **用户操作浏览器** → 触发URL变化、标签页操作
2. **浏览器插件捕获事件** → background.js监听
3. **组装JSON消息** → 包含type、url、timestamp等
4. **通过WebSocket发送** → 实时传输到服务器
5. **服务器接收处理** → server.js处理并记录日志
6. **可选：发送确认** → 服务器向客户端发送ACK

## 消息格式

所有消息都是JSON格式，包含以下字段：

```javascript
{
  "type": "消息类型",          // connection, url_change, tab_created, etc.
  "tabId": 12345,             // 标签页ID（如适用）
  "url": "https://...",       // URL地址（如适用）
  "title": "页面标题",        // 页面标题（如适用）
  "timestamp": "ISO 8601时间戳"
}
```

## 部署说明

### 开发环境
1. **启动服务器**: `cd server && npm install && npm start`
2. **安装插件**: 在Chrome/Edge中加载 `brower-monitor` 目录
3. **配置插件**: 设置服务器地址为 `ws://localhost:8080/monitor`

### 生产环境
1. **服务器部署**: 
   - 使用PM2或Docker部署到云服务器
   - 配置Nginx反向代理
   - 使用 `wss://` 加密连接

2. **插件分发**:
   - 打包插件目录为.zip文件
   - 上传到Chrome Web Store（可选）
   - 或者直接分发目录供用户手动安装

## 技术栈

### 浏览器插件
- **标准**: Chrome Extension Manifest V3
- **API**: Chrome Tabs API, Storage API, WebRequest API
- **通信**: WebSocket (原生)

### 服务器
- **运行时**: Node.js 18+
- **WebSocket库**: ws (^8.14.2)
- **开发工具**: nodemon (可选)

## 扩展计划

### 短期
- [ ] 添加URL过滤规则
- [ ] 实现本地数据缓存
- [ ] 支持批量消息发送

### 中期
- [ ] 集成Spring Boot后端
- [ ] 添加数据库持久化
- [ ] 实现用户认证机制

### 长期
- [ ] 支持Firefox浏览器
- [ ] 开发管理后台
- [ ] 添加数据统计和可视化

## 文档导航

- **快速开始**: [README.md](./README.md)
- **插件文档**: [dy-live-record/brower-monitor/README.md](./dy-live-record/brower-monitor/README.md)
- **快速入门**: [dy-live-record/brower-monitor/QUICKSTART.md](./dy-live-record/brower-monitor/QUICKSTART.md)
- **服务器文档**: [server/README.md](./server/README.md)

## 许可证

MIT License

## 更新记录

- **2025-11-15**: 
  - ✅ 初始项目创建
  - ✅ 浏览器插件开发完成
  - ✅ WebSocket服务器独立到 `/server/` 目录
  - ✅ 完善文档体系

---

**维护者**: DY Live Record Team
