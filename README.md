# DY Live Record - 浏览器URL监控项目

## 项目概述

这是一个抖音直播录制相关的项目，包含浏览器监控插件，用于捕获和记录浏览器URL变化。

## 项目结构

```
dy-live-record/
└── brower-monitor/          # Chrome/Edge 浏览器扩展插件
    ├── manifest.json        # 扩展配置文件
    ├── background.js        # 后台服务脚本
    ├── popup.html          # 配置界面HTML
    ├── popup.js            # 配置界面脚本
    ├── example-server.js   # WebSocket测试服务器
    ├── package.json        # Node.js依赖配置
    ├── icons/              # 图标文件夹
    │   └── README.md       # 图标说明
    └── README.md           # 插件详细文档
```

## 快速开始

### 1. 安装浏览器插件

详细步骤请参考：[brower-monitor/README.md](./dy-live-record/brower-monitor/README.md)

**简要步骤：**
1. 准备图标文件（或使用占位图标）
2. 在Chrome/Edge中加载解压缩的扩展
3. 配置WebSocket服务器地址
4. 启用监控

### 2. 启动测试服务器

```bash
cd dy-live-record/brower-monitor
npm install
npm start
```

服务器将在 `ws://localhost:8080/monitor` 上运行

## 主要功能

### 浏览器插件功能
- ✅ 监控所有标签页的URL变化
- ✅ 捕获标签页创建、关闭和激活事件
- ✅ 通过WebSocket实时发送数据到服务器
- ✅ 自动重连机制（断线后每5秒重连）
- ✅ 美观的配置界面
- ✅ 实时连接状态显示

### WebSocket消息类型
- `connection` - 连接建立
- `url_change` - URL变化
- `tab_created` - 新标签页创建
- `tab_closed` - 标签页关闭
- `tab_activated` - 标签页激活

## 技术栈

### 浏览器插件
- **Manifest V3** - Chrome扩展最新标准
- **Service Worker** - 后台脚本
- **Chrome APIs** - tabs, storage, webRequest
- **WebSocket** - 实时通信

### 测试服务器
- **Node.js** - 运行环境
- **ws** - WebSocket库

## 使用场景

1. **URL监控和分析**
   - 记录用户浏览历史
   - 分析用户访问模式
   - 监控特定网站访问

2. **直播录制辅助**
   - 检测直播间URL
   - 自动触发录制任务
   - 记录直播时长和URL变化

3. **数据采集**
   - 实时采集浏览数据
   - 用户行为分析
   - 网站访问统计

## 安全说明

- 插件需要访问所有网站的权限以监控URL变化
- 所有数据通过WebSocket发送，请确保服务器端安全
- 建议在生产环境使用 `wss://` (WebSocket Secure)
- 敏感数据应在服务器端加密存储

## 开发计划

- [ ] 添加数据过滤规则（只监控特定域名）
- [ ] 添加数据本地存储功能
- [ ] 支持HTTP REST API作为备选通信方式
- [ ] 添加统计和分析界面
- [ ] 支持Firefox浏览器

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如有问题，请在GitHub上提交Issue。

---

**最后更新时间：** 2025-11-15
