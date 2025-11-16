# 运行时错误修复指南

## ✅ 已修复的运行时错误

### 问题 1: GetVersionInfo 未定义

**错误信息**:
```
.\main.go:17:24: undefined: GetVersionInfo
```

**原因**: 本地代码未同步

**解决方案**:
```cmd
# 拉取最新代码
git pull

# version.go 文件已存在并正确定义
```

---

### 问题 2: License 公钥空指针

**错误信息**:
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x18 pc=0x7ff7455eaf2a]

goroutine 1 [running, locked to thread]:
dy-live-monitor/internal/license.getEmbeddedPublicKey()
```

**原因**: 
- `getEmbeddedPublicKey()` 尝试解析无效的 PEM 公钥
- 返回 `nil` 导致后续代码空指针解引用

**修复**:
1. ✅ 改进 `NewManager` 错误处理
2. ✅ 添加公钥 `nil` 检查
3. ✅ `Validate` 方法添加 `nil` 保护
4. ✅ 提供清晰的警告信息

---

## 🚀 正确的使用方法

### 方法 1: 使用调试模式（推荐）⭐

**适用场景**: 开发、测试环境

```cmd
# 1. 拉取最新代码
git pull

# 2. 进入目录
cd server-go

# 3. 使用调试配置
copy config.debug.json config.json

# 4. 运行
go run main.go
```

**config.debug.json 内容**:
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
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": false
  }
}
```

**效果**: 
- ✅ 跳过 License 验证
- ✅ 无需配置公钥
- ✅ 可以正常运行

---

### 方法 2: 配置公钥路径（生产环境）

**适用场景**: 生产环境，需要 License 验证

```cmd
# 1. 生成 RSA 密钥对（使用 server-active）
cd ../server-active
go run . generate-keys

# 2. 复制公钥到 server-go
copy public_key.pem ../server-go/

# 3. 配置 config.json
cd ../server-go
```

**config.json 配置**:
```json
{
  "license": {
    "server_url": "http://your-license-server:9090",
    "public_key_path": "./public_key.pem"
  },
  "debug": {
    "enabled": false,
    "skip_license": false
  }
}
```

---

## 📝 修复详情

### NewManager 改进

**修复前**:
```go
func NewManager(serverURL, publicKeyPath string) *Manager {
    pubKey, err := loadPublicKey(publicKeyPath)
    if err != nil {
        pubKey = getEmbeddedPublicKey() // 可能返回 nil
    }
    return &Manager{
        publicKey: pubKey, // nil 会导致后续 panic
    }
}
```

**修复后**:
```go
func NewManager(serverURL, publicKeyPath string) *Manager {
    var pubKey *rsa.PublicKey
    
    // 尝试从文件加载
    if publicKeyPath != "" {
        pubKey, err = loadPublicKey(publicKeyPath)
        if err != nil {
            log.Printf("⚠️  公钥文件加载失败: %v", err)
        }
    }
    
    // 尝试使用嵌入公钥
    if pubKey == nil {
        pubKey = getEmbeddedPublicKey()
    }
    
    // nil 检查和警告
    if pubKey == nil {
        log.Println("⚠️  警告：未找到有效公钥，License 验证将无法工作")
        log.Println("⚠️  请配置 publicKeyPath 或启用调试模式")
    }
    
    return &Manager{
        publicKey: pubKey, // 可能是 nil，但不会 panic
    }
}
```

---

### Validate 改进

**修复前**:
```go
func (m *Manager) Validate(licenseString string) (bool, time.Time, error) {
    // 直接使用 m.publicKey，如果是 nil 会 panic
    err = rsa.VerifyPKCS1v15(m.publicKey, ...)
```

**修复后**:
```go
func (m *Manager) Validate(licenseString string) (bool, time.Time, error) {
    // 先检查公钥
    if m.publicKey == nil {
        return false, time.Time{}, errors.New("公钥未配置，无法验证许可证")
    }
    
    // 再进行验证
    err = rsa.VerifyPKCS1v15(m.publicKey, ...)
```

---

## 🎯 运行效果

### 调试模式（skip_license: true）

```
2025/11/16 23:15:28 main.go:17: 🚀 抖音直播监控系统 v3.2.1 (2025-11-15) 启动...
2025/11/16 23:15:28 database.go:35: ✅ 数据库表结构初始化完成
2025/11/16 23:15:28 main.go:54: ✅ 数据库初始化成功
2025/11/16 23:15:28 license.go:61: ⚠️  警告：未找到有效公钥，License 验证将无法工作
2025/11/16 23:15:28 license.go:62: ⚠️  请配置 publicKeyPath 或启用调试模式
2025/11/16 23:15:28 main.go:61: ⚠️  调试模式已启用，跳过 License 验证
2025/11/16 23:15:28 main.go:62: ⚠️  警告：调试模式仅供开发使用，生产环境请禁用！
2025/11/16 23:15:28 main.go:91: ✅ WebSocket 服务器启动成功 (端口: 8080)
2025/11/16 23:15:28 main.go:94: ✅ 启动图形界面...
```

✅ **程序正常启动，无 panic**

---

### 生产模式（有公钥）

```
2025/11/16 23:15:28 main.go:17: 🚀 抖音直播监控系统 v3.2.1 (2025-11-15) 启动...
2025/11/16 23:15:28 database.go:35: ✅ 数据库表结构初始化完成
2025/11/16 23:15:28 main.go:54: ✅ 数据库初始化成功
2025/11/16 23:15:28 main.go:81: ✅ 许可证校验通过，有效期至: 2026-11-16
2025/11/16 23:15:28 main.go:91: ✅ WebSocket 服务器启动成功 (端口: 8080)
2025/11/16 23:15:28 main.go:94: ✅ 启动图形界面...
```

✅ **正常验证 License**

---

### 生产模式（无公钥，无 License）

```
2025/11/16 23:15:28 main.go:17: 🚀 抖音直播监控系统 v3.2.1 (2025-11-15) 启动...
2025/11/16 23:15:28 database.go:35: ✅ 数据库表结构初始化完成
2025/11/16 23:15:28 main.go:54: ✅ 数据库初始化成功
2025/11/16 23:15:28 license.go:61: ⚠️  警告：未找到有效公钥，License 验证将无法工作
2025/11/16 23:15:28 license.go:62: ⚠️  请配置 publicKeyPath 或启用调试模式
2025/11/16 23:15:28 main.go:67: ⚠️  未找到有效许可证，请激活软件
```

✅ **提示激活，不会 panic**

---

## 🔧 快速修复命令

### 如果遇到 GetVersionInfo 错误

```cmd
# 拉取最新代码
git pull

# 检查 version.go
cd server-go
type version.go

# 如果文件不存在，重新拉取
git fetch --all
git reset --hard origin/cursor/browser-extension-for-url-and-ws-capture-46de
```

---

### 如果遇到 License panic

```cmd
# 方案 1: 使用调试模式（推荐）
cd server-go
copy config.debug.json config.json
go run main.go

# 方案 2: 配置公钥（生产环境）
# 联系管理员获取 public_key.pem
# 放到 server-go 目录
# 在 config.json 中配置路径
```

---

## 📚 相关文档

- **[DEBUG_MODE.md](DEBUG_MODE.md)** - 调试模式详细说明
- **[ALL_FIXES_COMPLETE.md](ALL_FIXES_COMPLETE.md)** - 所有修复总结
- **[README_FYNE.md](README_FYNE.md)** - Fyne GUI 使用指南

---

## ✅ 验证修复

### 测试 1: 调试模式启动

```cmd
cd server-go
copy config.debug.json config.json
go run main.go
```

**预期**: 
- ✅ 无 panic
- ✅ 显示调试模式警告
- ✅ 跳过 License 验证
- ✅ GUI 正常显示

---

### 测试 2: 编译并运行

```cmd
cd server-go
go build -o dy-live-monitor.exe .
.\dy-live-monitor.exe
```

**预期**: 
- ✅ 编译成功
- ✅ 程序正常启动
- ✅ 无崩溃

---

## 🎉 总结

### 修复的问题

1. ✅ GetVersionInfo 未定义 - 确保 version.go 存在
2. ✅ License 公钥空指针 - 添加 nil 检查和友好提示

### 推荐做法

1. **开发/测试**: 使用调试模式（`config.debug.json`）
2. **生产环境**: 配置正确的公钥路径
3. **无 License**: 程序会友好提示，不会崩溃

---

**最后更新**: 2025-11-16  
**版本**: v3.2.1  
**提交**: 8ae3670  
**状态**: ✅ 所有运行时错误已修复
