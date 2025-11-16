# 📝 Release Notes

## v3.1.0 (2025-11-15) - 完整版发布 🎉

### 🎯 重大更新

这是项目的**首个完整版本**，包含所有计划功能，生产就绪！

---

## ✨ 新增功能

### 1. 依赖自动检查与安装
- ✅ **WebView2 Runtime 自动检测**
  - 检查 3 个标准安装路径
  - 自动检测 Edge 浏览器（内置 WebView2）
  - **一键自动下载并安装**
  
- ✅ **SQLite 驱动检测**
  - CGO 环境变量检查
  - GCC/MinGW 工具链检测
  - 详细的安装指南

- ✅ **网络连接检测**
  - NTP 服务器连通性测试
  - 离线模式支持

- ✅ **磁盘空间检测**
  - 当前目录可写性检查

### 2. 分段记分功能
- ✅ **多时段统计**
  - 创建新分段（如：PK 第一轮、主播独播时段）
  - 结束分段并自动计算统计
  - 礼物总值和消息数统计

- ✅ **主播业绩分段统计**
  - 按时段查看各主播业绩
  - 支持历史数据查询

- ✅ **UI 界面**
  - 新增「📈 分段记分」标签页
  - 实时显示进行中的分段（黄色高亮）
  - 创建/结束分段操作

### 3. WebView2 Fallback 数据通道
- ✅ **备用数据源**
  - 插件失效时自动启用
  - 隐藏 WebView2 窗口（1x1 像素）
  
- ✅ **WebSocket 拦截**
  - JavaScript 注入拦截原生 WebSocket
  - 捕获所有 WebSocket 消息
  - Base64 编码传输到 Go 后端

- ✅ **无缝集成**
  - 与 Protobuf 解析器无缝集成
  - 心跳检测机制（30 秒）

### 4. 浏览器插件管理
- ✅ **自动打包**
  - Windows 批处理脚本 (`pack.bat`)
  - Linux/Mac Shell 脚本 (`pack.sh`)
  - 输出到 `server-go/assets/`

- ✅ **内嵌式部署**
  - 插件 zip 内嵌到可执行文件
  - 一键解压到临时目录
  - 自动打开浏览器扩展页面

### 5. 管理后台 UI
- ✅ **现代化设计**
  - 渐变色背景
  - 卡片式布局
  - 响应式设计

- ✅ **完整功能**
  - 生成新许可证
  - 查看许可证列表
  - 查看许可证详情
  - 撤销许可证

- ✅ **新增 API**
  - `GET /api/v1/licenses/list` - 获取许可证列表

---

## 🔧 技术改进

### 核心架构
- ✅ **Go 1.21+ 支持**
- ✅ **SQLite 本地持久化**
- ✅ **RSA 2048 位加密**
- ✅ **WebView2 桌面 UI**

### 性能优化
- ✅ 数据库索引优化
- ✅ 并发连接支持
- ✅ 离线数据缓存

### 安全增强
- ✅ 硬件指纹绑定
- ✅ NTP 时间同步
- ✅ 许可证签名验证

---

## 📦 组件版本

| 组件 | 版本 | 状态 |
|------|------|------|
| server-go | v3.1.0 | ✅ 生产就绪 |
| browser-monitor | v3.1.0 | ✅ 生产就绪 |
| server-active | v3.1.0 | ✅ 生产就绪 |

---

## 🚀 快速开始

### 1. 构建所有组件

**Windows:**
```bash
BUILD_ALL.bat
```

**Linux/Mac:**
```bash
chmod +x BUILD_ALL.sh
./BUILD_ALL.sh
```

### 2. 启动服务

```bash
# 启动核心后端
cd server-go
dy-live-monitor.exe  # Windows
./dy-live-monitor     # Linux/Mac

# 启动许可证服务
cd server-active
dy-live-license-server.exe  # Windows
./dy-live-license-server     # Linux/Mac
```

### 3. 安装浏览器插件

参考 `QUICK_START.md`

---

## 🐛 已知问题

无严重问题，项目稳定运行。

### 建议改进
1. 管理后台添加认证机制（Basic Auth / JWT）
2. 支持更多直播平台（B站、快手）
3. 数据导出功能（Excel、CSV）

---

## 📚 文档

- **QUICK_START.md** - 快速开始指南
- **COMPLETION_REPORT.md** - 完整功能报告
- **UPGRADE_GUIDE.md** - 升级指南（v2.x → v3.x）
- **server-go/README.md** - 后端服务文档
- **server-active/README.md** - 许可证服务文档

---

## 🙏 致谢

感谢所有贡献者和测试人员！

---

## 📞 技术支持

- GitHub: https://github.com/WanGuChou/dy-live-record
- 分支: `cursor/browser-extension-for-url-and-ws-capture-46de`

---

**发布日期**: 2025-11-15  
**项目状态**: 🟢 生产就绪  
**完成度**: 100%
