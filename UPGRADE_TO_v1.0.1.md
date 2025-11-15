# 🚀 升级到 v1.0.1 指南

## ⚡ 快速升级

### 步骤1: 重新加载插件（必须）

⚠️ **非常重要**: manifest.json已更改，必须重新加载插件

1. 打开 `chrome://extensions/` 或 `edge://extensions/`
2. 找到 "URL & Request Monitor" 插件
3. 点击刷新按钮 🔄
4. 查看是否有错误

### 步骤2: 验证版本

点击 **"Service Worker"** 链接，应该看到：

```
🎯 URL & Request Monitor 已初始化
📊 版本: 1.0.1
🔍 监控内容: 所有URL变化和网络请求（包括WebSocket）
```

如果看到 `版本: 1.0.1`，说明升级成功！

### 步骤3: 测试新功能

#### A. 测试刷新捕获

1. 访问任意网站（如 https://www.baidu.com）
2. 按 **F5** 刷新
3. 查看Service Worker日志

**应该看到：**
```
🔄 页面导航: https://www.baidu.com/
🚀 页面已提交 [reload]: https://www.baidu.com/
📄 [1] main_frame: https://www.baidu.com/
🎨 [2] stylesheet: ...
📜 [3] script: ...
```

✅ 如果看到 `🔄 页面导航` 和 `🚀 页面已提交`，表示刷新捕获正常！

#### B. 测试WebSocket捕获

**方法1: 使用测试网站**

访问 https://www.websocket.org/echo.html

**应该看到：**
```
🔌🔌 WebSocket升级请求: wss://echo.websocket.org/
  标签页: 123
  ✅ 发送WebSocket升级请求
```

**方法2: 手动测试**

在任意网页的Console中执行：
```javascript
const ws = new WebSocket('wss://echo.websocket.org/');
ws.onopen = () => console.log('✅ WS connected');
```

✅ 如果看到 `🔌🔌 WebSocket升级请求`，表示WebSocket捕获正常！

---

## 🆕 新功能一览

### 1. WebSocket专门捕获 🔌

**之前**: WebSocket连接不显示  
**现在**: 显示 `🔌🔌 WebSocket升级请求`

### 2. 完整的页面刷新捕获 🔄

**之前**: 刷新可能遗漏某些请求  
**现在**: 
- 显示 `🔄 页面导航`
- 显示 `🚀 页面已提交 [reload]`
- 捕获所有资源请求

### 3. 请求计数器 📊

每个请求都有编号：
```
📄 [1] main_frame: ...
🎨 [2] stylesheet: ...
📜 [3] script: ...
```

### 4. 请求错误捕获 ❌

现在会记录失败的请求：
```
❌ 请求错误 [net::ERR_CONNECTION_REFUSED]: https://...
```

### 5. 导航类型识别 🚀

显示不同的导航方式：
- `[reload]` - 刷新
- `[typed]` - 地址栏输入
- `[link]` - 点击链接
- `[forward_back]` - 前进/后退

---

## 🔧 技术变更

### manifest.json
```diff
{
  "permissions": [
    "tabs",
    "webRequest",
+   "webNavigation",  ← 新增权限
    "storage"
  ]
}
```

### background.js

**新增监听器：**
- ✅ `chrome.webNavigation.onBeforeNavigate` - 导航开始
- ✅ `chrome.webNavigation.onCommitted` - 导航提交
- ✅ `chrome.webRequest.onBeforeSendHeaders` - 捕获WS升级
- ✅ `chrome.webRequest.onErrorOccurred` - 捕获错误

---

## 🐛 故障排查

### Q: 重新加载后看不到Service Worker

**解决：**
1. 完全卸载插件
2. 重新加载插件文件夹
3. 检查是否有错误提示

### Q: 还是看不到WebSocket连接

**检查：**
1. 确认版本是 `1.0.1`
2. 确认"启用监控"开关打开
3. 确认服务器已连接
4. 尝试访问 https://www.websocket.org/echo.html

### Q: 刷新页面还是遗漏请求

**检查：**
1. 清空Console后再刷新
2. 确保Service Worker在运行（不要关闭DevTools）
3. 查看是否有JavaScript错误

---

## 📊 验证清单

升级完成后，确认以下内容：

- [ ] 版本显示为 1.0.1
- [ ] F5刷新能看到 `🔄 页面导航`
- [ ] 能看到 `🚀 页面已提交 [reload]`
- [ ] WebSocket连接显示 `🔌🔌 WebSocket升级请求`
- [ ] 每个请求有编号 `[1]`, `[2]`...
- [ ] 请求过滤功能正常
- [ ] 服务器能收到消息

---

## 📚 详细文档

- **改进说明**: [IMPROVEMENTS.md](./IMPROVEMENTS.md)
- **详细测试**: [DETAILED_TEST.md](./DETAILED_TEST.md)
- **主文档**: [README.md](./README.md)

---

## 💡 遇到问题？

1. 查看 [DETAILED_TEST.md](./DETAILED_TEST.md) 进行完整测试
2. 查看 [IMPROVEMENTS.md](./IMPROVEMENTS.md) 了解技术细节
3. 确认所有文件都已更新

---

**升级时间**: 2025-11-15  
**版本**: v1.0.0 → v1.0.1  
**重要性**: 🔴 重要更新（修复关键功能）
