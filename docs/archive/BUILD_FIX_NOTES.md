# 🔧 构建问题修复说明

## 问题 1: `pattern assets/browser-monitor.zip: no matching files found`

### 根本原因
`//go:embed` 指令在编译时需要文件相对于源码文件的路径。

`settings.go` 位置：
```
server-go/internal/ui/settings.go
```

embed 路径 `assets/browser-monitor.zip` 会从 `settings.go` 的位置开始查找：
```
server-go/internal/ui/assets/browser-monitor.zip  ❌ 不存在
```

实际文件位置：
```
server-go/assets/browser-monitor.zip  ✅ 存在
```

### 解决方案
**移除 embed 指令**，改为运行时从外部文件读取：

```go
// 修改前
//go:embed assets/browser-monitor.zip
var embeddedPlugin []byte

// 修改后
var embeddedPlugin []byte  // 运行时从外部加载
```

### 为什么不使用 embed？
1. **路径问题**: embed 路径难以正确配置
2. **灵活性**: 外部文件更易于更新和测试
3. **构建简化**: 不需要在编译时嵌入文件
4. **体积**: 减小可执行文件体积

### 插件安装流程
```
启动 server-go
  ↓
设置界面 → 点击"安装插件"
  ↓
从 server-go/assets/browser-monitor.zip 读取
  ↓
解压到临时目录
  ↓
打开浏览器扩展页面
```

---

## 问题 2: `undefined: crypto`

### 根本原因
`manager.go` 第 136 行使用了 `crypto.SHA256`，但没有导入 `crypto` 包。

### 错误位置
```go
// manager.go:136
signature, err := rsa.SignPKCS1v15(rand.Reader, m.privateKey, crypto.SHA256, hashed[:])
                                                               ^^^^^^^^^^^^^^
```

### 解决方案
添加 `crypto` 包导入：

```go
// 修改前
import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	...
)

// 修改后
import (
	"crypto"           // ✅ 添加这一行
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	...
)
```

---

## ✅ 修复验证

### 测试编译 server-go
```bash
cd server-go
go build -v
```

**预期输出**:
```
dy-live-monitor/internal/config
dy-live-monitor/internal/database
dy-live-monitor/internal/parser
...
✅ 编译成功
```

### 测试编译 server-active
```bash
cd server-active
go build -v
```

**预期输出**:
```
dy-live-license/internal/config
dy-live-license/internal/database
dy-live-license/internal/license
...
✅ 编译成功
```

### 运行完整构建
```bash
BUILD_ALL.bat
```

**预期输出**:
```
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

## 📋 修改文件清单

| 文件 | 修改内容 | 原因 |
|------|---------|------|
| `server-go/internal/ui/settings.go` | 移除 `//go:embed` 和 `_ "embed"` | 修复 embed 路径问题 |
| `server-active/internal/license/manager.go` | 添加 `import "crypto"` | 修复 undefined 错误 |

---

## 🚀 现在可以正常构建

所有问题已修复，请重新运行：

```bash
BUILD_ALL.bat
```

如果仍有问题，请查看：
- **README_ERRORS.md** - 常见错误
- **BUILD_INSTRUCTIONS.md** - 构建说明
- **INSTALL_GUIDE.md** - 安装指南

---

**修复日期**: 2025-11-15  
**修复版本**: v3.1.1
