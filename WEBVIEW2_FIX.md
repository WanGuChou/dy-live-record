# WebView2 编译问题修复指南

## 🐛 问题描述

编译 server-go 时出现以下错误：
```
fatal error: EventToken.h: No such file or directory
  978 | #include "EventToken.h"
```

## 🔍 原因分析

`EventToken.h` 是 Windows 10 SDK 的一部分，属于 WinRT 头文件。

错误原因：
1. 未安装 Windows 10 SDK
2. MinGW-w64 无法找到 Windows SDK 头文件
3. `webview_go` 库依赖完整的 Windows SDK

## ✅ 解决方案

### 方案 1: 安装 Windows 10 SDK（推荐）

#### 下载并安装
1. 访问：https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
2. 下载最新版 Windows 10 SDK
3. 运行安装程序
4. **确保选择安装**：
   - Windows SDK for Desktop C++ Apps
   - Windows SDK C++ Headers
   - Windows SDK C++ Libraries

#### 安装后配置
```bash
# 设置 Windows SDK 路径
set INCLUDE=%INCLUDE%;C:\Program Files (x86)\Windows Kits\10\Include\<version>\um
set INCLUDE=%INCLUDE%;C:\Program Files (x86)\Windows Kits\10\Include\<version>\shared

# 重新编译
cd server-go
go build
```

---

### 方案 2: 使用无 WebView2 版本（临时方案）

如果不需要图形界面，可以禁用 WebView2：

#### 已完成的修改
1. `go.mod` - 注释掉 webview_go 依赖
2. `main.go` - 注释掉 WebView2 主窗口代码

#### 编译测试
```bash
cd server-go
go mod tidy
go build -v -o dy-live-monitor.exe .
```

**功能影响**:
- ✅ 系统托盘正常工作
- ✅ WebSocket 服务器正常工作
- ✅ 数据采集和存储正常
- ❌ 无法显示 WebView2 主界面
- ✅ 可通过其他方式（如 Web 浏览器）查看数据

---

### 方案 3: 使用 Visual Studio（最简单）

#### 安装 Visual Studio Community
1. 下载：https://visualstudio.microsoft.com/downloads/
2. 安装时选择：
   - "Desktop development with C++"
   - Windows 10 SDK
3. 重启系统
4. 使用 VS 的命令提示符编译

```bash
# 使用 VS Developer Command Prompt
cd server-go
go build
```

---

### 方案 4: 手动配置 MinGW + Windows SDK

#### 1. 下载 Windows SDK
从 Microsoft 下载并安装 Windows 10 SDK

#### 2. 创建符号链接
```bash
# 以管理员身份运行
mklink /D "C:\mingw-w64\include\EventToken.h" "C:\Program Files (x86)\Windows Kits\10\Include\<version>\um\EventToken.h"
```

#### 3. 设置环境变量
```bash
set CGO_CFLAGS=-IC:/Program Files (x86)/Windows Kits/10/Include/<version>/um
set CGO_LDFLAGS=-LC:/Program Files (x86)/Windows Kits/10/Lib/<version>/um/x64
```

---

## 🚀 推荐方案对比

| 方案 | 难度 | 时间 | WebView2 支持 | 推荐度 |
|------|------|------|--------------|--------|
| 安装 Windows SDK | 中等 | 30-60分钟 | ✅ 完整支持 | ⭐⭐⭐⭐⭐ |
| 禁用 WebView2 | 简单 | 5分钟 | ❌ 禁用 | ⭐⭐⭐⭐ |
| 安装 VS | 简单 | 60-120分钟 | ✅ 完整支持 | ⭐⭐⭐ |
| 手动配置 | 困难 | 30-60分钟 | ⚠️ 可能不稳定 | ⭐⭐ |

---

## 🎯 当前项目状态

**已应用方案 2（无 WebView2 版本）**

修改文件：
- `server-go/go.mod` - 注释 webview_go
- `server-go/main.go` - 禁用主窗口

优点：
- ✅ 立即可以编译
- ✅ 核心功能不受影响
- ✅ 不需要额外安装

缺点：
- ❌ 无图形界面

---

## 📝 如何恢复 WebView2

当安装好 Windows SDK 后：

### 1. 恢复 go.mod
```go
require (
    github.com/webview/webview_go v0.0.0-20230901181450-5a14030a9070
)
```

### 2. 恢复 main.go
```go
// 5. 启动主窗口
mainWindow := ui.NewMainWindow(db, wsServer)
go ui.RunSystemTray(cfg, db, wsServer, licenseManager)
mainWindow.Show()
```

### 3. 重新编译
```bash
cd server-go
go mod tidy
go build
```

---

## 🔧 验证 Windows SDK 安装

```bash
# 检查 Windows SDK 路径
dir "C:\Program Files (x86)\Windows Kits\10\Include"

# 查找 EventToken.h
dir /s "C:\Program Files (x86)\Windows Kits\10\Include\*EventToken.h"
```

**预期输出**:
```
C:\Program Files (x86)\Windows Kits\10\Include\10.0.xxxxx.0\um\EventToken.h
```

---

## 📞 技术支持

如需帮助：
1. 查看 `BUILD_INSTRUCTIONS.md`
2. 查看 `README_ERRORS.md`
3. GitHub Issues: https://github.com/WanGuChou/dy-live-record/issues

---

**最后更新**: 2025-11-15  
**版本**: v3.1.1  
**状态**: 🟡 WebView2 临时禁用
