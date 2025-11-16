# ❌ 常见错误及解决方案

## 🎯 快速索引

| 错误类型 | 关键词 | 跳转 |
|---------|-------|------|
| GCC 相关 | `gcc not found`, `cgo` | [错误 1](#错误-1-gcc-command-not-found) |
| 依赖相关 | `missing go.sum`, `go mod` | [错误 2](#错误-2-missing-gosum-entry) |
| 插件相关 | `pattern assets/*`, `embed` | [错误 3](#错误-3-pattern-assets-no-matching-files-found) |
| 网络相关 | `timeout`, `connection refused` | [错误 4](#错误-4-go-downloading--connection-timed-out) |
| WebView2 | `WebView2 Runtime` | [错误 5](#错误-5-webview2-runtime-未安装) |
| MySQL | `connection refused`, `access denied` | [错误 6](#错误-6-mysql-连接失败) |
| 构建顺序 | `BUILD_ALL.bat` | [错误 7](#错误-7-buildallbat-执行顺序错误) |

---

## 🐛 详细错误说明

### 错误 1: `gcc: command not found`

#### 完整错误信息
```
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in %PATH%
```

#### 错误原因
- 未安装 MinGW-w64（GCC 编译器）
- MinGW-w64 未添加到系统 PATH
- 命令行窗口未重启（PATH 未生效）

#### 解决方案

**Step 1: 安装 MinGW-w64**

**方法 A: Chocolatey（推荐）**
```powershell
# 以管理员身份运行 PowerShell
choco install mingw -y
```

**方法 B: 手动安装**
1. 下载：https://sourceforge.net/projects/mingw-w64/
2. 选择：`x86_64-posix-seh`
3. 安装到：`C:\mingw-w64`

**Step 2: 添加到 PATH**
```powershell
# 临时添加（当前窗口）
set PATH=%PATH%;C:\mingw-w64\bin

# 永久添加
# 1. 右键"此电脑" → "属性"
# 2. "高级系统设置" → "环境变量"
# 3. 编辑 "Path"
# 4. 添加 "C:\mingw-w64\bin"
```

**Step 3: 验证**
```bash
# 关闭并重新打开命令行
gcc --version
```

**预期输出**:
```
gcc.exe (x86_64-posix-seh-rev0, Built by MinGW-W64 project) 8.1.0
```

---

### 错误 2: `missing go.sum entry`

#### 完整错误信息
```
internal\database\database.go:8:2: missing go.sum entry for module providing package github.com/mattn/go-sqlite3 (imported by dy-live-monitor/internal/database); to add:
        go get dy-live-monitor/internal/database
```

#### 错误原因
- `go.sum` 文件缺失或不完整
- 依赖包未下载
- `go.mod` 与实际导入不一致

#### 解决方案

**方法 1: 运行 go mod tidy**
```bash
cd server-go
go mod tidy
```

**方法 2: 删除 go.sum 重新生成**
```bash
cd server-go
del go.sum
go mod tidy
```

**方法 3: 手动下载依赖**
```bash
cd server-go
go mod download
go mod download github.com/mattn/go-sqlite3
go mod download github.com/gorilla/websocket
go mod download github.com/webview/webview_go
go mod download github.com/getlantern/systray
```

**方法 4: 清理缓存重新下载**
```bash
go clean -modcache
go mod tidy
```

---

### 错误 3: `pattern assets/*: no matching files found`

#### 完整错误信息
```
internal\ui\settings.go:15:12: pattern assets/*: no matching files found
```

#### 错误原因
- `browser-monitor.zip` 还未生成
- `server-go/assets/` 目录不存在
- `//go:embed` 找不到文件

#### 解决方案

**Step 1: 创建 assets 目录**
```bash
mkdir server-go\assets
```

**Step 2: 打包 browser-monitor**
```bash
cd browser-monitor
pack.bat
```

**Step 3: 验证文件生成**
```bash
dir ..\server-go\assets\browser-monitor.zip
```

**预期输出**:
```
2025/11/15  10:30            512,345 browser-monitor.zip
```

**Step 4: 重新编译 server-go**
```bash
cd ..\server-go
build.bat
```

**如果仍然失败**:

编辑 `server-go/internal/ui/settings.go`：
```go
// 修改前
//go:embed assets/*
var assets embed.FS

// 修改后
//go:embed assets/browser-monitor.zip
var embeddedPlugin []byte
```

然后重新编译。

---

### 错误 4: `go: downloading ... connection timed out`

#### 完整错误信息
```
go: downloading github.com/gorilla/websocket v1.5.1
dial tcp 142.251.42.113:443: i/o timeout
```

#### 错误原因
- 网络连接问题
- 防火墙阻止
- 无法访问 GitHub/Go 官方代理
- DNS 解析失败

#### 解决方案

**方法 1: 使用国内代理（推荐）**
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
go mod tidy
```

**方法 2: 使用 HTTP 代理**
```bash
# 设置代理（根据你的代理软件端口）
set HTTP_PROXY=http://127.0.0.1:7890
set HTTPS_PROXY=http://127.0.0.1:7890

# 下载依赖
go mod download
```

**方法 3: 多次重试**
```bash
# Go 会自动重试失败的下载
go mod download
go mod download
go mod download
```

**方法 4: 手动下载依赖包**
```bash
# 逐个下载
go get github.com/gorilla/websocket@v1.5.1
go get github.com/mattn/go-sqlite3@v1.14.18
go get github.com/webview/webview_go@latest
go get github.com/getlantern/systray@v1.2.2
```

**方法 5: 修改 DNS**
```bash
# 修改 hosts 文件
# C:\Windows\System32\drivers\etc\hosts

# 添加以下内容
140.82.114.4 github.com
185.199.108.133 raw.githubusercontent.com
```

---

### 错误 5: WebView2 Runtime 未安装

#### 错误信息
```
⚠️  检测到关键依赖缺失
WebView2 Runtime: 未检测到，需要安装 WebView2 Runtime
```

#### 错误原因
- Windows 10/11 未自带 WebView2
- WebView2 Runtime 未安装

#### 解决方案

**方法 1: 自动安装（推荐）**
```bash
# 启动 server-go
dy-live-monitor.exe

# 出现提示时输入 y
是否尝试自动安装 WebView2? (y/n): y

# 程序会自动下载并安装
```

**方法 2: 手动下载安装**
1. 访问：https://developer.microsoft.com/en-us/microsoft-edge/webview2/
2. 下载 "Evergreen Standalone Installer"
3. 运行安装程序
4. 重启 `dy-live-monitor.exe`

**方法 3: 使用 Edge 浏览器（包含 WebView2）**
- Windows 11 和 Windows 10 (20H1+) 自带 Edge
- Edge 包含 WebView2 Runtime
- 确保 Edge 已更新到最新版本

---

### 错误 6: MySQL 连接失败

#### 错误信息
```
❌ 数据库初始化失败: dial tcp 127.0.0.1:3306: connect: connection refused
```

#### 错误原因
- MySQL 服务未启动
- MySQL 端口不是 3306
- 用户名/密码错误
- 数据库不存在

#### 解决方案

**Step 1: 启动 MySQL**
```bash
# Windows
net start mysql80

# 或使用服务管理器
# Win+R → services.msc → 找到 MySQL80 → 启动
```

**Step 2: 检查 MySQL 连接**
```bash
mysql -u root -p
```

**Step 3: 创建数据库**
```sql
CREATE DATABASE dy_license CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EXIT;
```

**Step 4: 检查配置文件**

编辑 `server-active/config.json`：
```json
{
  "database": {
    "host": "localhost",
    "port": "3306",
    "user": "root",
    "password": "your_password",
    "database": "dy_license"
  }
}
```

**Step 5: 测试连接**
```bash
mysql -h localhost -P 3306 -u root -p dy_license
```

---

### 错误 7: BUILD_ALL.bat 执行顺序错误

#### 错误信息
```
[1/3] Building server-go...
internal\ui\settings.go:15:12: pattern assets/*: no matching files found
❌ server-go 编译失败

[2/3] Packing browser-monitor...
✅ browser-monitor 打包成功
```

#### 错误原因
- 旧版 `BUILD_ALL.bat` 先编译 server-go，后打包插件
- server-go 需要 `browser-monitor.zip`（embed）
- 构建顺序错误

#### 解决方案

**已修复**: 最新版 `BUILD_ALL.bat` 已修复顺序

**新的构建顺序**:
1. 打包 browser-monitor → 生成 .zip
2. 下载 server-go 依赖
3. 编译 server-go → 使用 .zip
4. 下载 server-active 依赖
5. 编译 server-active

**如果使用旧版**:

手动执行：
```bash
# 1. 先打包插件
cd browser-monitor
pack.bat
cd ..

# 2. 再编译 server-go
cd server-go
go mod tidy
build.bat
cd ..

# 3. 最后编译 server-active
cd server-active
go mod tidy
build.bat
cd ..
```

---

## 🔧 通用排查步骤

### 1. 检查环境变量

```bash
# 检查 Go 版本
go version

# 检查 Go 环境
go env

# 关键变量
CGO_ENABLED=1
GOPROXY=https://goproxy.cn,direct
GOPATH=...
```

### 2. 清理并重建

```bash
# 清理 Go 缓存
go clean -modcache
go clean -cache

# 删除编译产物
del server-go\dy-live-monitor.exe
del server-active\dy-live-license-server.exe

# 重新构建
BUILD_ALL.bat
```

### 3. 逐步调试

```bash
# 1. 只打包插件
cd browser-monitor
pack.bat

# 2. 只编译 server-go
cd ..\server-go
go mod tidy
go build -v

# 3. 只编译 server-active
cd ..\server-active
go mod tidy
go build -v
```

---

## 📞 获取帮助

如果上述方法都无法解决问题：

1. **查看详细文档**:
   - `INSTALL_GUIDE.md` - 安装指南
   - `BUILD_INSTRUCTIONS.md` - 构建说明
   - `QUICK_START.md` - 快速开始

2. **收集错误信息**:
   ```bash
   # 运行构建并保存日志
   BUILD_ALL.bat > build.log 2>&1
   
   # 查看日志
   type build.log
   ```

3. **GitHub Issues**:
   - https://github.com/WanGuChou/dy-live-record/issues
   - 附上完整错误日志

---

**最后更新**: 2025-11-15  
**版本**: v3.1.0
