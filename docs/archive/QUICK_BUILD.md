# 快速编译指南

## 🚀 最简单的方法（推荐）

### 方法 1: 自动编译脚本

```cmd
# 在 CMD 中运行（不是 PowerShell）
cd C:\path\to\dy-live-record
.\BUILD_WITH_SDK.bat
```

**这个脚本会自动**：
- ✅ 检测 Windows SDK 安装
- ✅ 自动设置环境变量
- ✅ 打包浏览器插件
- ✅ 编译 server-go
- ✅ 编译 server-active
- ✅ 如果没有 SDK，自动编译无 WebView2 版本

---

### 方法 2: 手动设置 SDK 路径

如果自动脚本失败，手动设置：

```cmd
# 在 CMD 中运行
.\SET_SDK_PATH.bat
```

然后编译：

```cmd
cd server-go
go build -o dy-live-monitor.exe .
```

---

### 方法 3: 不使用 SDK（最快）

**直接使用无 WebView2 版本**：

```cmd
# 1. 打包插件
cd browser-monitor
.\pack.bat

# 2. 编译 server-go
cd ..\server-go
go mod tidy
go build -ldflags="-H windowsgui" -o dy-live-monitor.exe .

# 3. 编译 server-active
cd ..\server-active
go mod tidy
go build -o dy-live-license.exe .
```

**功能对比**：

| 功能 | 有 SDK | 无 SDK |
|------|--------|--------|
| 数据采集 | ✅ | ✅ |
| WebSocket | ✅ | ✅ |
| 数据存储 | ✅ | ✅ |
| 系统托盘 | ✅ | ✅ |
| 图形界面 | ✅ | ❌ |

**核心功能完全相同！**

---

## ❓ 选择哪个方法？

### 如果你想要图形界面
→ 使用 **方法 1**（自动编译脚本）

### 如果你想快速测试
→ 使用 **方法 3**（无 SDK 版本）

### 如果方法 1 失败
→ 使用 **方法 2**（手动设置 SDK）

---

## 🔧 PowerShell vs CMD 对照表

| 操作 | CMD | PowerShell |
|------|-----|------------|
| 设置变量 | `set VAR=value` | `$env:VAR = "value"` |
| 查看变量 | `echo %VAR%` | `echo $env:VAR` |
| 运行批处理 | `.\script.bat` | `cmd /c .\script.bat` |
| 路径空格 | 自动处理 | 需要引号 |

**建议**: 编译时使用 CMD，日常使用 PowerShell

---

## 🐛 常见错误

### 错误 1: "x86 无法识别"
**原因**: 在 PowerShell 中使用了 CMD 语法

**解决**: 
```powershell
# PowerShell 中运行批处理
cmd /c .\BUILD_WITH_SDK.bat

# 或切换到 CMD
```

### 错误 2: "EventToken.h 找不到"
**原因**: 未安装 Windows SDK 或路径未设置

**解决**: 
```cmd
# 运行自动脚本
.\BUILD_WITH_SDK.bat

# 或直接编译无 SDK 版本（方法 3）
```

### 错误 3: "go.sum 不一致"
**解决**:
```cmd
cd server-go
del go.sum
go mod tidy
go build .
```

---

## 📞 需要帮助？

查看详细文档：
- `WEBVIEW2_FIX.md` - WebView2 问题详解
- `BUILD_INSTRUCTIONS.md` - 完整编译说明
- `README_ERRORS.md` - 错误排查

---

**创建时间**: 2025-11-15  
**适用版本**: v3.1.2+
