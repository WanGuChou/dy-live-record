# 测试指南

## 测试新功能

### 1. 重新加载插件

1. 打开 `chrome://extensions/` (Chrome) 或 `edge://extensions/` (Edge)
2. 找到 "URL & WebSocket Monitor" 插件
3. 点击 **刷新按钮** 🔄

### 2. 查看Service Worker日志

1. 在扩展管理页面，点击 **"Service Worker"** 链接
2. 会弹出DevTools窗口
3. 切换到 **Console** 标签

### 3. 配置插件

1. 点击浏览器工具栏中的插件图标
2. 配置：
   - 服务器地址: `ws://localhost:8080/monitor`
   - 过滤关键字: 留空（测试所有请求）或填写如 `baidu,google`
3. 点击"测试连接"
4. 点击"保存配置"
5. 开启"启用监控"开关

### 4. 启动服务器

```bash
cd server
npm start
```

### 5. 测试场景

#### 场景A: 地址栏输入URL

1. 在地址栏输入 `https://www.baidu.com`
2. 按回车

**预期日志（Service Worker Console）：**
```
🌐 地址栏URL变化: https://www.baidu.com/
  ✅ 匹配过滤条件，发送到服务器
📄 请求 [main_frame]: https://www.baidu.com/
  ✅ 匹配过滤条件，发送到服务器
🎨 请求 [stylesheet]: https://www.baidu.com/style.css
  ✅ 匹配过滤条件，发送到服务器
📜 请求 [script]: https://www.baidu.com/app.js
  ✅ 匹配过滤条件，发送到服务器
...（更多资源请求）
✅ 请求完成 [200]: https://www.baidu.com/
```

**预期日志（服务器终端）：**
```
[时间] ✅ 客户端连接确认
[时间] 🔄 地址栏URL变化
  URL: https://www.baidu.com/
[时间] 📡 网络请求 (主页面)
  URL: https://www.baidu.com/
[时间] ✅ 请求完成 (主页面)
  URL: https://www.baidu.com/
  状态码: 200
```

---

#### 场景B: 刷新页面

1. 在当前页面按 **F5** 或点击刷新按钮
2. 观察日志

**预期日志：**
- Service Worker应该显示大量请求
- 包括 main_frame、stylesheet、script、image 等
- 服务器应该收到这些请求

**如果没有看到日志：**
- 检查"启用监控"开关是否打开
- 检查Service Worker是否在运行（点击Service Worker链接）
- 查看是否有错误信息

---

#### 场景C: 测试过滤功能

1. 在插件配置中设置过滤关键字: `baidu`
2. 保存配置
3. 访问 `https://www.baidu.com`
4. 访问 `https://www.google.com`

**预期结果：**
- 访问baidu.com时，所有请求都发送到服务器
- 访问google.com时，日志显示"不匹配过滤条件"，不发送

**Service Worker日志：**
```
# 访问 baidu.com
🌐 地址栏URL变化: https://www.baidu.com/
  ✅ 匹配过滤条件，发送到服务器
📄 请求 [main_frame]: https://www.baidu.com/
  ✅ 匹配过滤条件，发送到服务器

# 访问 google.com
🌐 地址栏URL变化: https://www.google.com/
  ⚠️ 不匹配过滤条件，跳过发送
📄 请求 [main_frame]: https://www.google.com/
  ⚠️ 不匹配过滤条件，跳过发送
```

---

#### 场景D: 查看所有类型的请求

1. 确保过滤关键字为空或包含要访问的网站
2. 访问一个资源丰富的网站（如 bilibili.com）
3. 观察Service Worker日志

**应该看到各种类型的请求：**
```
📄 请求 [main_frame]: https://www.bilibili.com/
🎨 请求 [stylesheet]: https://www.bilibili.com/style.css
📜 请求 [script]: https://www.bilibili.com/app.js
🖼️ 请求 [image]: https://www.bilibili.com/logo.png
🔤 请求 [font]: https://www.bilibili.com/font.woff2
🔗 请求 [xmlhttprequest]: https://api.bilibili.com/data
🎬 请求 [media]: https://video.bilibili.com/video.mp4
```

---

## 验证清单

- [ ] **插件已重新加载**
- [ ] **Service Worker日志可见**
- [ ] **服务器正在运行**
- [ ] **连接状态显示"已连接"**
- [ ] **启用监控开关已打开**
- [ ] **地址栏输入URL能看到日志**
- [ ] **刷新页面能看到请求日志**
- [ ] **过滤功能正常工作**
- [ ] **服务器收到消息**
- [ ] **所有请求类型都能捕获**

---

## 常见问题

### Q1: Service Worker日志中看不到任何输出

**解决：**
1. 确保点击了"Service Worker"链接
2. 尝试关闭DevTools窗口，重新点击"Service Worker"
3. 重新加载插件
4. 查看是否有错误信息

### Q2: 刷新页面没有日志

**解决：**
1. 确保"启用监控"开关是蓝色（已打开）
2. 在Service Worker Console中输入: `chrome.runtime.sendMessage({action: 'getStatus'}, console.log)`
3. 检查返回的状态
4. 查看是否有JavaScript错误

### Q3: 过滤不生效

**解决：**
1. 确保过滤关键字已保存
2. 重新打开插件配置，确认关键字显示正确
3. 关键字是大小写敏感的
4. 多个关键字用英文逗号分隔

### Q4: 服务器收不到消息

**解决：**
1. 检查连接状态是否为"已连接"
2. 检查过滤关键字设置
3. 在Service Worker日志中查看是否显示"发送到服务器"
4. 检查服务器终端是否显示"客户端已连接"

---

## 调试命令

在Service Worker Console中执行：

### 查看当前状态
```javascript
chrome.runtime.sendMessage({action: 'getStatus'}, response => {
  console.log('状态:', response);
});
```

### 查看存储的配置
```javascript
chrome.storage.local.get(['serverUrl', 'isEnabled', 'filterKeywords'], result => {
  console.log('配置:', result);
});
```

### 手动触发连接
```javascript
chrome.runtime.sendMessage({
  action: 'updateConfig',
  serverUrl: 'ws://localhost:8080/monitor',
  filterKeywords: '',
  isEnabled: true
}, response => {
  console.log('响应:', response);
});
```

---

## 预期行为总结

### 所有请求都打印到日志 ✅
- 插件Service Worker会显示所有捕获的请求
- 包括URL变化和网络请求
- 使用emoji标记不同类型

### 过滤功能 ✅
- 没有关键字 = 发送所有请求
- 有关键字 = 只发送匹配的请求
- 不匹配的请求仍然打印日志，但标记"不匹配过滤条件"

### 刷新页面 ✅
- 刷新会触发新的请求
- 所有资源重新加载
- 应该能看到完整的请求列表

---

**测试成功标志：**
- ✅ Service Worker日志显示所有请求
- ✅ 过滤功能正常工作
- ✅ 服务器收到符合条件的请求
- ✅ 刷新页面能捕获所有请求
