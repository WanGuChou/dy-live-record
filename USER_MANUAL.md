# 用户使用手册

## 🚨 如果遇到 `undefined: GetVersionInfo` 错误

这说明您本地缺少 `version.go` 文件。

---

## ✅ 解决方案（3 种方法）

### 方法 1: 强制更新（推荐）⭐

```cmd
REM 强制拉取最新代码
git fetch --all
git reset --hard origin/cursor/browser-extension-for-url-and-ws-capture-46de

REM 进入目录
cd server-go

REM 检查文件
dir *.go
```

**应该看到**: `main.go` 和 `version.go`

---

### 方法 2: 手动创建 version.go

如果方法 1 不行，手动创建文件：

**文件路径**: `server-go/version.go`

**文件内容**:
```go
package main

const (
	// Version 版本号
	Version = "v3.2.1"

	// BuildDate 构建日期
	BuildDate = "2025-11-15"

	// AppName 应用名称
	AppName = "抖音直播监控系统"

	// AppNameEN 应用英文名称
	AppNameEN = "Douyin Live Monitor"
)

// GetVersionInfo 获取版本信息
func GetVersionInfo() string {
	return AppName + " " + Version + " (" + BuildDate + ")"
}
```

**创建步骤**:
1. 在 `server-go` 目录下
2. 创建新文件 `version.go`
3. 复制上面的内容
4. 保存文件

---

### 方法 3: 注释掉版本信息（临时）

如果急着测试，可以临时注释掉：

**编辑 `server-go/main.go`**:

```go
func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// log.Println("🚀 " + GetVersionInfo() + " 启动...")
	log.Println("🚀 抖音直播监控系统 v3.2.1 启动...")
	
	// ... 其他代码
}
```

---

## 🎯 完整测试流程

### Step 1: 更新代码

```cmd
REM 方法 A: 普通更新
git pull

REM 方法 B: 强制更新（推荐）
git fetch --all
git reset --hard origin/cursor/browser-extension-for-url-and-ws-capture-46de
```

---

### Step 2: 验证文件

```cmd
cd server-go
dir *.go
```

**必须看到**:
```
main.go
version.go    <-- 必须存在！
```

---

### Step 3: 配置调试模式

```cmd
copy config.debug.json config.json
```

---

### Step 4: 运行程序

```cmd
go run main.go
```

**预期输出**:
```
2025/11/16 23:30:00 main.go:17: 🚀 抖音直播监控系统 v3.2.1 (2025-11-15) 启动...
2025/11/16 23:30:00 checker.go:66: ✅ 所有依赖检查通过
2025/11/16 23:30:00 database.go:35: ✅ 数据库表结构初始化完成
2025/11/16 23:30:00 main.go:54: ✅ 数据库初始化成功
2025/11/16 23:30:00 main.go:61: ⚠️  调试模式已启用，跳过 License 验证
2025/11/16 23:30:00 main.go:91: ✅ WebSocket 服务器启动成功 (端口: 8080)
2025/11/16 23:30:00 main.go:94: ✅ 启动图形界面...
```

✅ **成功！**

---

## 🔍 故障排查

### 检查 1: version.go 文件是否存在

```cmd
cd server-go
type version.go
```

**应该看到**: 文件内容，包含 `GetVersionInfo` 函数

**如果报错**: 文件不存在，使用方法 2 手动创建

---

### 检查 2: main.go 和 version.go 在同一目录

```cmd
cd server-go
dir *.go
```

**应该看到**:
```
main.go
version.go
```

**都在 `server-go` 目录下**

---

### 检查 3: 包名是否一致

**version.go 第一行**:
```go
package main
```

**main.go 第一行**:
```go
package main
```

**必须都是 `package main`**

---

### 检查 4: 清理缓存

```cmd
cd server-go
go clean -cache
go clean -modcache
go mod tidy
go build
```

---

## 📝 config.debug.json 内容

**文件路径**: `server-go/config.debug.json`

**如果文件不存在，创建它**:

```json
{
  "server": {
    "port": 8080
  },
  "database": {
    "path": "./data.db"
  },
  "license": {
    "server_url": "",
    "public_key_path": ""
  },
  "browser": {
    "startup_params": "--silent-debugger-extension-api"
  },
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": false
  }
}
```

---

## 🚀 一键修复脚本

复制以下命令，一次执行：

```cmd
REM ========================================
REM 完整修复和启动流程
REM ========================================

REM 1. 强制更新代码
git fetch --all
git reset --hard origin/cursor/browser-extension-for-url-and-ws-capture-46de

REM 2. 进入目录
cd server-go

REM 3. 检查文件
echo 检查 Go 文件...
dir *.go

REM 4. 清理缓存
echo 清理缓存...
go clean -cache
go mod tidy

REM 5. 配置调试模式
if not exist config.json (
    echo 复制调试配置...
    copy config.debug.json config.json
)

REM 6. 测试编译
echo 测试编译...
go build

REM 7. 运行程序
echo 启动程序...
go run main.go
```

---

## 📞 仍然无法解决？

### 提供以下信息

```cmd
REM 1. 检查 Git 状态
git status

REM 2. 检查分支
git branch

REM 3. 检查最新提交
git log --oneline -3

REM 4. 检查文件
cd server-go
dir *.go

REM 5. 尝试读取 version.go
type version.go
```

**将输出发送给我，我会帮您诊断。**

---

## ✅ 成功标志

### 文件检查
```cmd
cd server-go
dir *.go
```
**应该看到**: `main.go` 和 `version.go`

### 编译测试
```cmd
go build
```
**应该**: 无错误，生成 `dy-live-monitor.exe`

### 运行测试
```cmd
go run main.go
```
**应该**: 显示 "🚀 抖音直播监控系统 v3.2.1 (2025-11-15) 启动..."

---

## 🎉 祝您使用愉快！

如果还有问题，请：
1. 使用强制更新（方法 1）
2. 提供详细的错误信息
3. 检查文件是否存在

---

**最后更新**: 2025-11-16  
**版本**: v3.2.1  
**状态**: 🟢 所有问题都有解决方案
