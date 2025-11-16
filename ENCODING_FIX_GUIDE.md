# 批处理脚本编码问题修复指南

## 🐛 问题症状

执行 `BUILD_WITH_FYNE.bat` 时出现：
```
'��平台支持' is not recognized as an internal or external command
'包浏览器插件...' is not recognized as an internal or external command
'4]' is not recognized as an internal or external command
```

## 🔍 问题原因

**Windows CMD 默认使用 GBK/ANSI 编码**，但我们的批处理脚本是 **UTF-8 编码**，导致中文字符被错误解析。

---

## ✅ 解决方案（4 种方法）

### 方法 1: 使用英文版本脚本（推荐）✨

```cmd
# 使用无中文字符的安全版本
.\BUILD_WITH_FYNE_SAFE.bat
```

**优点**: 
- ✅ 不依赖编码设置
- ✅ 100% 兼容所有 Windows 系统
- ✅ 功能完全一致

---

### 方法 2: 手动设置 UTF-8 编码

**在运行脚本前执行：**
```cmd
# 设置控制台为 UTF-8
chcp 65001

# 再运行脚本
.\BUILD_WITH_FYNE.bat
```

**永久设置（可选）：**
1. Win + R，输入 `regedit`
2. 找到：`HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Command Processor`
3. 新建字符串值：`AutoRun`，值为 `chcp 65001`
4. 重启 CMD

---

### 方法 3: 使用 PowerShell（推荐）

```powershell
# PowerShell 对 UTF-8 支持更好
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 运行脚本
.\BUILD_WITH_FYNE.bat
```

---

### 方法 4: 分步手动执行

如果上述方法都不行，可以手动执行每一步：

```cmd
# Step 1: 打包插件
cd browser-monitor
call pack.bat
cd ..

# Step 2: 下载依赖
cd server-go
go mod download
go mod tidy

# Step 3: 检查 GCC
where gcc
gcc --version

# Step 4: 编译
set CGO_ENABLED=1
go build -v -o dy-live-monitor.exe .

# Step 5: 运行
.\dy-live-monitor.exe
```

---

## 🔧 额外检查

### 1. 验证 GCC 安装

```cmd
where gcc
gcc --version
```

**预期输出**:
```
C:\mingw-w64\bin\gcc.exe
gcc.exe (x86_64-posix-seh-rev0, Built by MinGW-W64 project) 8.1.0
```

**如果未安装**:
```cmd
# 使用 Chocolatey 安装（推荐）
choco install mingw -y

# 或手动下载
# https://www.mingw-w64.org/
```

---

### 2. 验证 Go 环境

```cmd
go version
go env

# 检查 CGO
go env CGO_ENABLED
```

**预期输出**:
```
go version go1.21.0 windows/amd64
CGO_ENABLED=1
```

---

### 3. 设置 Go 代理（网络问题）

```cmd
# 使用国内镜像
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

---

## 📊 编码检查工具

### 检查文件编码

```cmd
# 使用 PowerShell
$content = Get-Content .\BUILD_WITH_FYNE.bat -Raw
$content.GetType().FullName

# 或使用 file 命令（需要 Git Bash）
file -i BUILD_WITH_FYNE.bat
```

**预期结果**: `charset=utf-8`

---

## 🎯 快速排查步骤

### Step 1: 尝试英文版本
```cmd
.\BUILD_WITH_FYNE_SAFE.bat
```
✅ 如果成功 → 确认是编码问题  
❌ 如果失败 → 继续 Step 2

### Step 2: 检查 GCC
```cmd
gcc --version
```
✅ 如果有输出 → 继续 Step 3  
❌ 如果错误 → 安装 MinGW-w64

### Step 3: 检查网络
```cmd
ping goproxy.cn
go mod download
```
✅ 如果成功 → 继续 Step 4  
❌ 如果失败 → 设置代理

### Step 4: 手动编译
```cmd
cd server-go
set CGO_ENABLED=1
go build -v -o dy-live-monitor.exe .
```
✅ 如果成功 → 运行程序  
❌ 如果失败 → 查看详细错误

---

## 📝 常见错误和解决方案

### 错误 1: `gcc: command not found`

**原因**: MinGW-w64 未安装或未添加到 PATH

**解决**:
```cmd
# 方法 1: Chocolatey
choco install mingw -y

# 方法 2: 手动安装
# 下载: https://sourceforge.net/projects/mingw-w64/
# 安装到: C:\mingw-w64
# 添加到 PATH: C:\mingw-w64\bin
```

---

### 错误 2: `go: downloading ... timeout`

**原因**: 网络问题，无法访问 Go 官方代理

**解决**:
```cmd
# 设置国内代理
set GOPROXY=https://goproxy.cn,direct
go mod download
```

---

### 错误 3: `cannot find package`

**原因**: 依赖缺失

**解决**:
```cmd
cd server-go
go mod tidy
go mod download
```

---

### 错误 4: `failed to initialize GL`

**原因**: 显卡驱动问题或虚拟机环境

**解决**:
```cmd
# 方法 1: 更新显卡驱动

# 方法 2: 使用系统托盘版本（无 GUI）
.\BUILD_NO_WEBVIEW2_FIXED.bat
```

---

## 🚀 推荐工作流程

### 开发/测试环境

```cmd
# 1. 使用英文版本脚本（避免编码问题）
.\BUILD_WITH_FYNE_SAFE.bat

# 2. 启用调试模式（跳过 License）
cd server-go
copy config.debug.json config.json

# 3. 运行
.\dy-live-monitor.exe
```

---

## 📞 获取帮助

### 查看详细文档
- [README_FYNE.md](README_FYNE.md) - Fyne GUI 详细说明
- [DEBUG_MODE.md](DEBUG_MODE.md) - 调试模式
- [README_ERRORS.md](README_ERRORS.md) - 错误排查
- [INSTALL_GUIDE.md](INSTALL_GUIDE.md) - 安装指南

### 提交问题
- GitHub Issues: https://github.com/WanGuChou/dy-live-record/issues

---

## ✨ 总结

| 方法 | 难度 | 推荐度 | 说明 |
|------|------|--------|------|
| BUILD_WITH_FYNE_SAFE.bat | ⭐ | ⭐⭐⭐⭐⭐ | 英文版本，无编码问题 |
| 手动设置 chcp 65001 | ⭐⭐ | ⭐⭐⭐⭐ | 需要每次设置 |
| 使用 PowerShell | ⭐⭐ | ⭐⭐⭐⭐ | 对 UTF-8 支持好 |
| 手动分步执行 | ⭐⭐⭐ | ⭐⭐⭐ | 调试时有用 |

**最佳实践**: 使用 `BUILD_WITH_FYNE_SAFE.bat`（英文版本）

---

**更新时间**: 2025-11-15  
**版本**: v3.2.1
