# 🔨 构建说明 - Windows 平台

## ⚠️ 构建前必读

### 第一次构建？请按照以下顺序执行！

```
┌─────────────────────────────────────────────────────────┐
│                    构建流程图                            │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  1️⃣  安装 MinGW-w64（GCC 编译器）                        │
│      ↓                                                   │
│  2️⃣  验证 gcc 可用                                       │
│      ↓                                                   │
│  3️⃣  运行 BUILD_ALL.bat                                 │
│      ├─ 打包 browser-monitor                            │
│      ├─ 下载 server-go 依赖                             │
│      ├─ 编译 server-go                                  │
│      ├─ 下载 server-active 依赖                         │
│      └─ 编译 server-active                              │
│      ↓                                                   │
│  4️⃣  验证构建产物                                        │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## 📋 准备工作

### Step 1: 安装 MinGW-w64

**为什么需要？**
- SQLite（`mattn/go-sqlite3`）需要 CGO 支持
- CGO 需要 GCC 编译器

**推荐方法：使用 Chocolatey**

```powershell
# 1. 以管理员身份打开 PowerShell

# 2. 安装 Chocolatey（如果未安装）
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# 3. 安装 MinGW-w64
choco install mingw -y

# 4. 重启命令行窗口
```

**手动安装方法**:

1. 下载：https://sourceforge.net/projects/mingw-w64/
2. 选择版本：
   - Architecture: `x86_64`
   - Threads: `posix`
   - Exception: `seh`
3. 安装到 `C:\mingw-w64`
4. 添加到环境变量：
   - 右键"此电脑" → "属性" → "高级系统设置" → "环境变量"
   - 编辑 `Path`，添加 `C:\mingw-w64\bin`
5. **重启命令行窗口**

---

### Step 2: 验证 GCC 安装

```bash
# 打开新的命令行窗口
gcc --version
```

**预期输出**:
```
gcc.exe (x86_64-posix-seh-rev0, Built by MinGW-W64 project) 8.1.0
...
```

**如果显示 `gcc: command not found`**:
- 确认 MinGW-w64 已安装
- 确认 `C:\mingw-w64\bin` 已添加到 PATH
- **重启命令行窗口**（重要！）

---

### Step 3: 配置 Go 环境（可选，国内用户推荐）

```bash
# 设置 Go 代理（加速依赖下载）
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

# 确保 CGO 已启用
go env -w CGO_ENABLED=1
```

---

## 🚀 开始构建

### 一键构建所有组件

```bash
# 以管理员身份运行（推荐）
BUILD_ALL.bat
```

**构建过程**:
```
========================================
Building All Components
========================================

[1/3] Packing browser-monitor...
  → 打包插件文件
  → 输出: server-go/assets/browser-monitor.zip
  ✅ [SUCCESS] browser-monitor packed successfully

[2/3] Downloading server-go dependencies...
  → go mod tidy
  → 下载 gorilla/websocket, mattn/go-sqlite3, webview_go, systray
  ✅ 依赖下载完成

[3/3] Building server-go...
  → 编译 Release 版本
  → 输出: dy-live-monitor.exe
  ✅ [SUCCESS] server-go built successfully

[4/5] Downloading server-active dependencies...
  → go mod tidy
  → 下载 gin, go-sql-driver/mysql, google/uuid
  ✅ 依赖下载完成

[5/5] Building server-active...
  → 编译 Release 版本
  → 输出: dy-live-license-server.exe
  ✅ [SUCCESS] server-active built successfully

========================================
Build Summary
========================================
Status: ALL BUILDS SUCCEEDED!

Output files:
  - server-go/dy-live-monitor.exe
  - server-go/assets/browser-monitor.zip
  - server-active/dy-live-license-server.exe
========================================
```

---

## ✅ 验证构建

### 检查构建产物

```bash
dir server-go\dy-live-monitor.exe
dir server-go\assets\browser-monitor.zip
dir server-active\dy-live-license-server.exe
```

**预期输出**:
```
server-go\dy-live-monitor.exe            ~15-25 MB
server-go\assets\browser-monitor.zip     ~500 KB
server-active\dy-live-license-server.exe ~15-20 MB
```

### 测试运行

```bash
cd server-go
dy-live-monitor.exe
```

**预期输出**:
```
🚀 抖音直播监控系统 v3.1.0 (2025-11-15) 启动...
🔍 开始检查系统依赖...
✅ WebView2 Runtime: 已安装
✅ SQLite Driver (CGO): CGO_ENABLED=true, GCC=true
✅ 网络连接: 正常
✅ 磁盘空间: 可写
✅ 所有依赖检查通过
✅ 数据库初始化成功
🌐 WebSocket 服务器监听: :8090
✅ 启动系统托盘...
```

---

## 🐛 常见构建错误

### 错误 1: `gcc: command not found`

**完整错误**:
```
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in %PATH%
```

**原因**: 未安装 MinGW-w64 或未添加到 PATH

**解决方案**:
1. 安装 MinGW-w64（见 Step 1）
2. 确认 `C:\mingw-w64\bin` 在 PATH 中
3. **重启命令行窗口**
4. 验证：`gcc --version`

---

### 错误 2: `missing go.sum entry`

**完整错误**:
```
internal\database\database.go:8:2: missing go.sum entry for module providing package github.com/mattn/go-sqlite3
```

**原因**: 缺少依赖包

**解决方案**:
```bash
cd server-go
go mod tidy

# 如果仍然失败，删除 go.sum 重新生成
del go.sum
go mod tidy
```

---

### 错误 3: `pattern assets/*: no matching files found`

**完整错误**:
```
internal\ui\settings.go:15:12: pattern assets/*: no matching files found
```

**原因**: browser-monitor 还未打包

**解决方案**:
```bash
# 先打包插件
cd browser-monitor
pack.bat

# 验证生成的文件
dir ..\server-go\assets\browser-monitor.zip

# 然后重新编译 server-go
cd ..\server-go
build.bat
```

---

### 错误 4: `go: downloading ... connection timed out`

**完整错误**:
```
go: downloading github.com/gorilla/websocket v1.5.1
dial tcp: i/o timeout
```

**原因**: 网络问题，无法下载依赖

**解决方案**:

**方法 1: 设置国内代理**
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

**方法 2: 使用 HTTP 代理**
```bash
set HTTP_PROXY=http://127.0.0.1:7890
set HTTPS_PROXY=http://127.0.0.1:7890
go mod tidy
```

**方法 3: 手动下载**
```bash
# 多次重试
go mod download
go mod download
go mod download
```

---

### 错误 5: `undefined: embeddedPlugin`

**完整错误**:
```
internal\ui\settings.go:45:6: undefined: embeddedPlugin
```

**原因**: embed 包导入问题

**解决方案**:
```bash
cd server-go
go mod tidy
go build -v
```

---

## 🔧 单独构建各组件

### 只构建 server-go

```bash
# 1. 确保插件已打包
cd browser-monitor
pack.bat
cd ..

# 2. 下载依赖
cd server-go
go mod tidy

# 3. 编译
build.bat
```

### 只构建 server-active

```bash
cd server-active
go mod tidy
build.bat
```

### 只打包 browser-monitor

```bash
cd browser-monitor
pack.bat
```

---

## 📊 构建性能优化

### 加速依赖下载

```bash
# 使用国内镜像
go env -w GOPROXY=https://goproxy.cn,direct

# 并行下载
go mod download -x
```

### 减小可执行文件体积

```bash
# 编译时去除调试信息
go build -ldflags="-s -w" -o dy-live-monitor.exe

# 使用 UPX 压缩（可选）
upx --best dy-live-monitor.exe
```

---

## 🚀 生产环境构建

### 发布版本构建

```bash
# 设置版本号
set VERSION=v3.1.0
set BUILD_DATE=2025-11-15

# 编译（嵌入版本信息）
cd server-go
go build -ldflags="-s -w -X main.Version=%VERSION% -X main.BuildDate=%BUILD_DATE%" -o dy-live-monitor.exe

# 验证版本
dy-live-monitor.exe --version
```

---

## 📝 构建检查清单

构建完成后，请检查：

- [ ] `gcc --version` 输出正常
- [ ] `server-go/dy-live-monitor.exe` 存在且 > 10 MB
- [ ] `server-go/assets/browser-monitor.zip` 存在
- [ ] `server-active/dy-live-license-server.exe` 存在且 > 10 MB
- [ ] `dy-live-monitor.exe` 可以正常启动
- [ ] 系统托盘显示图标
- [ ] WebSocket 服务器监听 :8090

---

## 📞 技术支持

如果仍然遇到构建问题：

1. 查看 **INSTALL_GUIDE.md** - 详细安装指南
2. 查看 **QUICK_START.md** - 快速开始
3. 查看 **server-go/README.md** - 后端文档

或访问 GitHub Issues:
- https://github.com/WanGuChou/dy-live-record/issues

---

**最后更新**: 2025-11-15  
**版本**: v3.1.0  
**适用平台**: Windows 10/11
