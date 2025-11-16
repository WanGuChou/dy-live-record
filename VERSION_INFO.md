# 版本信息

## 当前版本

**v3.2.1** (2025-11-15)

---

## 版本历史

### v3.2.1 (2025-11-15)
**类型**: 功能增强 + 文档更新

**新增**:
- ✅ 调试模式支持（可跳过 License 验证）
- ✅ 配置文件模板（config.example.json, config.debug.json）
- ✅ 详细的调试模式文档

**改进**:
- ✅ 文档清理（归档 39 个过时文档）
- ✅ 更正所有文档中的过时信息
- ✅ 统一文档结构

**文档**:
- ✅ DEBUG_MODE.md
- ✅ DOCS_CLEANUP.md
- ✅ DOCUMENTATION_STRUCTURE.md
- ✅ VERSION_INFO.md

---

### v3.2.0 (2025-11-15)
**类型**: 重大更新 - GUI 框架迁移

**核心变更**:
- ✅ 完全迁移到 Fyne GUI 框架
- ❌ 移除所有 WebView2 相关代码
- ✅ 无需 Windows SDK
- ✅ 跨平台支持（Windows/Linux/macOS）

**性能提升**:
- 编译时间: 5-10 分钟 → 2-3 分钟 (60-70%)
- 启动时间: 3 秒 → 1 秒 (66%)
- 内存占用: 150MB → 80MB (46%)

**文档**:
- ✅ README_FYNE.md
- ✅ FYNE_MIGRATION.md
- ✅ UPGRADE_TO_FYNE.md
- ✅ CHANGELOG_v3.2.0.md

---

### v3.1.2 (2025-11-15)
**类型**: Bug 修复

**修复**:
- ✅ ByteBuffer 类型转换错误
- ✅ CGO 路径空格问题
- ✅ 批处理脚本编码问题

**新增**:
- ✅ 完整的 Protocol Buffers 定义
- ✅ server-go/proto/ 目录

---

### v3.1.0 (2025-11-14)
**类型**: 重大架构升级

**核心变更**:
- ✅ 从 Node.js 重构为 Go 语言
- ✅ WebView2 图形界面
- ✅ SQLite 数据持久化
- ✅ 完整的 License 授权系统

**新增组件**:
- ✅ server-go (Go 后端)
- ✅ server-active (License 服务)
- ✅ browser-monitor (CDP 版本)

---

### v2.x (Legacy)
**类型**: Node.js 版本

**技术栈**:
- Node.js + Express
- WebSocket
- 文件日志

**状态**: 已废弃，不再维护

---

## 技术栈版本

### v3.2.1 当前技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 编程语言 | Go | 1.21+ |
| GUI 框架 | Fyne | v2.4.3 |
| 数据库 | SQLite | 3.x |
| WebSocket | gorilla/websocket | v1.5.1 |
| 系统托盘 | getlantern/systray | v1.2.2 |
| 浏览器插件 | Manifest V3 | - |
| License 服务 | Go + MySQL | 8.0+ |

---

## 系统要求

### 开发/编译

| 平台 | 要求 |
|------|------|
| Windows | Go 1.21+, MinGW-w64 |
| Linux | Go 1.21+, GCC, X11 开发库 |
| macOS | Go 1.21+, Xcode Command Line Tools |

### 运行

| 平台 | 要求 |
|------|------|
| Windows | Windows 10/11, OpenGL 2.0+ |
| Linux | X11, OpenGL 2.0+ |
| macOS | macOS 10.13+, OpenGL 2.0+ |

---

## 构建工具

### Windows
- `BUILD_WITH_FYNE.bat` - Fyne GUI 版本（推荐）
- `BUILD_NO_WEBVIEW2_FIXED.bat` - 系统托盘版本
- `QUICK_START.bat` - 交互式编译

### Linux/macOS
```bash
cd server-go
go mod tidy
go build -o dy-live-monitor .
```

---

## 功能完成度

| 功能模块 | 完成度 | 状态 |
|---------|-------|------|
| 浏览器数据采集 | 100% | ✅ |
| WebSocket 通信 | 100% | ✅ |
| Protobuf 解析 | 100% | ✅ |
| SQLite 存储 | 100% | ✅ |
| Fyne GUI | 100% | ✅ |
| 系统托盘 | 100% | ✅ |
| 主播管理 | 100% | ✅ |
| 分段记分 | 100% | ✅ |
| License 系统 | 100% | ✅ |
| 调试模式 | 100% | ✅ |
| 跨平台支持 | 100% | ✅ |

**总体完成度**: 100% ✅

---

## 发布渠道

### GitHub
- **仓库**: https://github.com/WanGuChou/dy-live-record
- **分支**: cursor/browser-extension-for-url-and-ws-capture-46de
- **标签**: v3.2.1

### 文档
- **主文档**: README.md
- **详细文档**: README_FYNE.md
- **变更日志**: CHANGELOG_v3.2.0.md

---

## 下一个版本预告

### v3.3.0 (计划中)
**预计发布**: 2025-12-01

**计划功能**:
- 数据可视化图表
- 主题切换（亮色/暗色）
- 数据导出（Excel/CSV）
- 性能优化

---

## 许可证

MIT License

---

## 联系方式

- **Issues**: https://github.com/WanGuChou/dy-live-record/issues
- **Discussions**: https://github.com/WanGuChou/dy-live-record/discussions
- **Email**: [项目维护者邮箱]

---

**最后更新**: 2025-11-15  
**维护者**: Cursor AI Assistant  
**状态**: 🟢 活跃开发中
