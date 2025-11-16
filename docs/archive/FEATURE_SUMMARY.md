# 功能总结

## 🎯 完成的功能

### 1. ✅ 所有请求都打印到插件日志

**实现方式：**
- 使用emoji图标标记不同类型的请求
- 每个请求都输出到Service Worker控制台
- 包括请求发起和完成状态

**日志格式：**
```
📄 请求 [main_frame]: https://example.com/
🎨 请求 [stylesheet]: https://example.com/style.css
📜 请求 [script]: https://example.com/app.js
🖼️ 请求 [image]: https://example.com/logo.png
🔗 请求 [xmlhttprequest]: https://api.example.com/data
✅ 请求完成 [200]: https://example.com/
```

**Emoji图标对照表：**
- 📄 `main_frame` - 主页面
- 🖼️ `sub_frame` - iframe
- 🎨 `stylesheet` - CSS
- 📜 `script` - JavaScript
- 🖼️ `image` - 图片
- 🔤 `font` - 字体
- 🔗 `xmlhttprequest` - AJAX/Fetch
- 🔌 `websocket` - WebSocket
- 🎬 `media` - 媒体文件
- 📦 `other` - 其他

---

### 2. ✅ 关键字过滤功能

**功能说明：**
- 在插件配置界面可以设置过滤关键字
- 支持多个关键字，用逗号分隔
- 只有URL包含任一关键字的请求才发送到服务器
- 所有请求仍然打印到日志，但会标记是否发送

**配置示例：**
```
过滤关键字: live,video,stream
```

**工作原理：**
1. 留空 → 发送所有请求
2. 填写关键字 → 只发送匹配的请求
3. 多个关键字 → 任意匹配即可

**日志输出：**
```
📄 请求 [main_frame]: https://live.example.com/
  ✅ 匹配过滤条件，发送到服务器

📄 请求 [main_frame]: https://other.example.com/
  ⚠️ 不匹配过滤条件，跳过发送
```

---

### 3. ✅ 捕获刷新页面请求

**实现方式：**
- 使用 `chrome.webRequest.onBeforeRequest` 监听所有请求
- 包括页面刷新触发的请求
- 监听URL: `<all_urls>`

**刷新页面时的行为：**
1. 触发 `main_frame` 请求（主页面）
2. 触发所有资源请求（CSS、JS、图片等）
3. 所有请求都会被捕获并打印日志
4. 符合过滤条件的请求发送到服务器

**测试方法：**
- 按 F5 或点击刷新按钮
- 查看Service Worker Console
- 应该看到一系列请求日志

---

## 📋 配置选项

### 插件配置界面

| 选项 | 说明 | 示例 |
|------|------|------|
| **服务器地址** | WebSocket服务器URL | `ws://localhost:8080/monitor` |
| **过滤关键字** | 用逗号分隔的关键字列表 | `live,video,stream` 或留空 |
| **启用监控** | 开关，控制是否监控 | 开启（蓝色） / 关闭（灰色） |

---

## 🔄 工作流程

```
1. 用户操作浏览器
   ↓
2. 触发URL变化或网络请求
   ↓
3. background.js 捕获事件
   ↓
4. 打印日志到 Service Worker Console
   ↓
5. 检查过滤关键字
   ↓
6a. 匹配 → 发送到服务器 ✅
6b. 不匹配 → 跳过发送 ⚠️
   ↓
7. 服务器接收并处理（如果发送）
```

---

## 📊 数据流

### 客户端 → 服务器

**消息类型1: 连接建立**
```json
{
  "type": "connection",
  "status": "connected",
  "filterKeywords": "live,video",
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

**消息类型2: URL变化**
```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://example.com",
  "title": "页面标题",
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

**消息类型3: 网络请求**
```json
{
  "type": "request",
  "requestId": "12345",
  "url": "https://example.com/api/data",
  "method": "GET",
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "frameId": 0,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

**消息类型4: 请求完成**
```json
{
  "type": "request_completed",
  "requestId": "12345",
  "url": "https://example.com/api/data",
  "method": "GET",
  "statusCode": 200,
  "resourceType": "xmlhttprequest",
  "tabId": 12345,
  "timestamp": "2025-11-15T10:00:00.000Z"
}
```

---

## 🎨 用户界面

### 插件配置弹窗
- **宽度**: 450px
- **设计**: 紫色渐变背景 + 白色卡片
- **元素**:
  - 连接状态指示器（红色/绿色圆点）
  - 服务器地址输入框
  - 过滤关键字输入框 ⭐ 新增
  - 启用监控开关
  - 保存配置 / 测试连接 按钮
  - 提示信息框

---

## 🔧 技术实现

### background.js 关键代码

```javascript
// 加载配置（包括过滤关键字）
async function loadConfig() {
  const result = await chrome.storage.local.get([
    'serverUrl', 
    'isEnabled', 
    'filterKeywords'  // ⭐ 新增
  ]);
  filterKeywords = result.filterKeywords || '';
}

// 检查URL是否匹配过滤条件
function matchesFilter(url) {
  if (!filterKeywords || filterKeywords.trim() === '') {
    return true;  // 没有关键字，全部发送
  }
  
  const keywords = filterKeywords.split(',')
    .map(k => k.trim())
    .filter(k => k !== '');
  
  return keywords.some(keyword => url.includes(keyword));
}

// 监听所有请求并打印日志
chrome.webRequest.onBeforeRequest.addListener(
  (details) => {
    // ⭐ 打印所有请求到控制台
    console.log(`📦 请求 [${details.type}]:`, details.url);
    
    // ⭐ 检查过滤条件
    if (matchesFilter(details.url)) {
      console.log('  ✅ 匹配过滤条件，发送到服务器');
      sendMessage(data);
    } else {
      console.log('  ⚠️ 不匹配过滤条件，跳过发送');
    }
  },
  { urls: ['<all_urls>'] }
);
```

---

## 📝 使用场景

### 场景1: 监控所有请求（不过滤）
```
配置: 
  - 过滤关键字: (留空)
  - 启用监控: ✅

结果:
  - 所有URL和请求都打印到日志
  - 所有请求都发送到服务器
```

### 场景2: 只监控直播相关请求
```
配置:
  - 过滤关键字: live,video,stream,m3u8
  - 启用监控: ✅

结果:
  - 所有请求都打印到日志
  - 只有包含这些关键字的请求发送到服务器
  - 日志中标记哪些发送，哪些跳过
```

### 场景3: 调试特定网站
```
配置:
  - 过滤关键字: douyin.com,tiktok.com
  - 启用监控: ✅

结果:
  - 只有这些域名的请求发送到服务器
  - 其他网站的请求被过滤
```

---

## ⚠️ 注意事项

1. **日志性能**
   - 所有请求都打印到Console
   - 访问资源多的网站会产生大量日志
   - 建议定期清空Console

2. **过滤关键字**
   - 大小写敏感
   - 支持URL的任意部分匹配
   - 多个关键字用逗号分隔

3. **服务器压力**
   - 不过滤时会产生大量消息
   - 建议设置合适的过滤关键字
   - 服务器端也可以进一步过滤

---

## 📚 相关文档

- [TEST_GUIDE.md](./TEST_GUIDE.md) - 详细测试指南
- [USAGE.md](./USAGE.md) - 使用说明
- [CHANGELOG.md](./CHANGELOG.md) - 更新日志
- [README.md](./README.md) - 项目说明

---

**更新时间**: 2025-11-15
