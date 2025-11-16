# 调试模式使用指南

## 🐛 什么是调试模式？

调试模式允许开发者在本地测试时跳过某些验证步骤，简化开发流程。

**⚠️ 警告**: 调试模式仅供开发使用，**不要在生产环境启用**！

---

## 🚀 快速启用调试模式

### 方法 1: 使用预设配置文件（推荐）

```bash
# 1. 复制调试配置
cd server-go
copy config.debug.json config.json

# 2. 运行程序
.\dy-live-monitor.exe
```

### 方法 2: 手动编辑配置

编辑 `server-go/config.json`：

```json
{
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": true
  }
}
```

---

## ⚙️ 调试选项说明

### 1. `enabled` - 启用调试模式

```json
"enabled": true
```

**作用**:
- 在窗口标题显示 `[调试模式]`
- 在状态栏显示调试标识
- 启用其他调试功能

**UI 显示**:
```
┌─────────────────────────────────────────────┐
│  抖音直播监控系统 v3.2.0 [调试模式]           │
├─────────────────────────────────────────────┤
│  礼物: 0  消息: 0  总值: 0  在线: 0  ⚠️调试模式│
└─────────────────────────────────────────────┘
```

---

### 2. `skip_license` - 跳过 License 验证

```json
"skip_license": true
```

**作用**:
- 启动时不验证 License
- 不检查硬件指纹
- 不连接 License 服务器

**控制台输出**:
```
🔐 初始化许可证系统...
⚠️  调试模式已启用，跳过 License 验证
⚠️  警告：调试模式仅供开发使用，生产环境请禁用！
```

**使用场景**:
- ✅ 本地开发测试
- ✅ 功能验证
- ✅ 离线开发环境
- ❌ 生产部署
- ❌ 用户环境

---

### 3. `verbose_log` - 详细日志

```json
"verbose_log": true
```

**作用**:
- 输出更详细的调试信息
- 记录所有网络请求
- 显示完整的错误堆栈

**使用场景**:
- 排查问题
- 性能分析
- 理解程序流程

---

## 📋 完整配置示例

### 生产环境配置 (`config.json`)

```json
{
  "server": {
    "port": 8080
  },
  "database": {
    "path": "data.db"
  },
  "license": {
    "key": "your-license-key-here",
    "server_url": "https://license.example.com",
    "offline_grace_days": 7,
    "validation_interval": 60
  },
  "debug": {
    "enabled": false,
    "skip_license": false,
    "verbose_log": false
  }
}
```

### 开发环境配置 (`config.debug.json`)

```json
{
  "server": {
    "port": 8080
  },
  "database": {
    "path": "data.db"
  },
  "license": {
    "key": "",
    "server_url": "http://localhost:8081",
    "offline_grace_days": 7,
    "validation_interval": 60
  },
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": true
  }
}
```

---

## 🔄 快速切换

### 开发 → 生产

```bash
cd server-go

# 使用生产配置
copy config.example.json config.json

# 或手动编辑 config.json，设置：
# "debug": { "enabled": false, "skip_license": false }
```

### 生产 → 开发

```bash
cd server-go

# 使用调试配置
copy config.debug.json config.json
```

---

## 🎨 UI 标识

### 正常模式
```
窗口标题: 抖音直播监控系统 v3.2.0
状态栏: 礼物: 123  消息: 456  总值: 78900  在线: 234
```

### 调试模式
```
窗口标题: 抖音直播监控系统 v3.2.0 [调试模式]
状态栏: 礼物: 123  消息: 456  总值: 78900  在线: 234  ⚠️调试模式
```

### 数据概览 Tab
调试模式下会显示额外的警告信息：

```
📊 实时监控说明

1. 打开浏览器并安装插件
2. 访问抖音直播间
3. 插件会自动采集数据
4. 数据实时显示在这里

当前功能：
✅ 礼物统计
✅ 消息记录
✅ 主播管理
✅ 分段记分
✅ 数据持久化

⚠️  调试模式已启用
⚠️  License 验证已跳过（仅供调试）
⚠️  详细日志已启用

❗ 警告：调试模式仅供开发使用，
   生产环境请在 config.json 中禁用！
```

---

## 🔒 安全注意事项

### ⚠️ 永远不要在生产环境启用调试模式

**原因**:
1. 跳过 License 验证 = 任何人都能使用
2. 详细日志可能暴露敏感信息
3. 调试功能可能影响性能

### ✅ 最佳实践

1. **开发时**: 使用 `config.debug.json`
2. **测试时**: 使用 `config.json` + `skip_license: false`
3. **生产时**: 使用 `config.json` + 所有调试选项关闭

---

## 📝 常见场景

### 场景 1: 本地开发新功能

```bash
# 1. 启用调试模式
copy config.debug.json config.json

# 2. 开发和测试
.\dy-live-monitor.exe

# 3. 完成后恢复
copy config.example.json config.json
```

### 场景 2: 在没有 License 服务器的环境测试

```json
{
  "debug": {
    "enabled": true,
    "skip_license": true,  // 跳过验证
    "verbose_log": false   // 不需要详细日志
  }
}
```

### 场景 3: 排查 License 验证问题

```json
{
  "debug": {
    "enabled": true,
    "skip_license": false,  // 仍然验证
    "verbose_log": true     // 查看详细日志
  }
}
```

---

## 🐛 调试技巧

### 1. 查看详细日志

```bash
# Windows
.\dy-live-monitor.exe > debug.log 2>&1

# Linux/macOS
./dy-live-monitor > debug.log 2>&1
```

### 2. 测试 License 功能

```bash
# 临时禁用 skip_license
# 编辑 config.json:
"skip_license": false

# 然后测试 License 验证流程
```

### 3. 性能分析

```json
{
  "debug": {
    "enabled": true,
    "skip_license": true,
    "verbose_log": true  // 启用以查看性能数据
  }
}
```

---

## 📚 相关命令

### 查看当前配置

```bash
cd server-go
type config.json  # Windows
cat config.json   # Linux/macOS
```

### 验证配置格式

```bash
# 使用 jq 验证（如果已安装）
jq . config.json
```

### 备份配置

```bash
# 备份当前配置
copy config.json config.backup.json

# 恢复配置
copy config.backup.json config.json
```

---

## ❓ 常见问题

### Q1: 调试模式下 License 服务器连接失败？

**A**: 正常！启用 `skip_license` 后不会连接 License 服务器。

---

### Q2: 忘记关闭调试模式就部署了？

**A**: 
1. 立即停止程序
2. 编辑 `config.json`，设置 `enabled: false`
3. 重启程序

---

### Q3: 如何确认调试模式已启用？

**A**: 检查以下标识：
- 窗口标题包含 `[调试模式]`
- 状态栏显示 `⚠️调试模式`
- 控制台输出 "调试模式已启用"

---

### Q4: 调试模式会影响数据采集吗？

**A**: 不会！调试模式只影响：
- License 验证
- 日志详细程度
- UI 显示

数据采集功能完全正常。

---

## 🎯 总结

### ✅ 调试模式的好处
- 快速本地测试
- 无需 License 服务器
- 便于开发调试

### ⚠️ 使用原则
- 仅在开发环境使用
- 测试后及时关闭
- 生产环境严格禁用

### 📝 推荐工作流

```
1. 开发 → config.debug.json（skip_license: true）
2. 测试 → config.json（skip_license: false）
3. 生产 → config.json（debug.enabled: false）
```

---

**最后更新**: 2025-11-15  
**适用版本**: v3.2.0+  
**配置文件**: `server-go/config.json`
