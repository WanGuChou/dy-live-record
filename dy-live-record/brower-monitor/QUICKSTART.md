# 快速入门指南

## 📋 前置准备

1. Chrome浏览器（版本88+）或 Edge浏览器（版本88+）
2. Node.js（如需运行测试服务器）

## 🚀 5分钟快速开始

### 步骤 1: 准备图标（可选）

在 `icons` 文件夹中放置以下PNG图标文件：
- icon16.png (16x16)
- icon32.png (32x32)
- icon48.png (48x48)
- icon128.png (128x128)

**提示：** 如果暂时没有图标，可以使用任意PNG图片，浏览器会自动处理。

### 步骤 2: 加载插件到浏览器

#### Chrome浏览器
```
1. 打开 chrome://extensions/
2. 开启"开发者模式"（右上角开关）
3. 点击"加载已解压的扩展程序"
4. 选择 brower-monitor 文件夹
```

#### Edge浏览器
```
1. 打开 edge://extensions/
2. 开启"开发人员模式"（左下角开关）
3. 点击"加载解压缩的扩展"
4. 选择 brower-monitor 文件夹
```

### 步骤 3: 启动测试服务器

打开终端，执行以下命令：

```bash
# 进入插件目录
cd dy-live-record/brower-monitor

# 安装依赖
npm install

# 启动服务器
npm start
```

您应该看到类似输出：
```
============================================================
WebSocket服务器已启动
地址: ws://localhost:8080/monitor
============================================================
```

### 步骤 4: 配置插件

1. 点击浏览器工具栏中的插件图标
2. 在弹出窗口中：
   - **WebSocket服务器地址**: `ws://localhost:8080/monitor`
   - 点击"测试连接"按钮确认连接成功
   - 点击"保存配置"
   - 开启"启用监控"开关

### 步骤 5: 验证工作状态

1. **检查插件状态**
   - 打开插件弹窗
   - 确认状态指示器显示为绿色"已连接"

2. **检查服务器日志**
   - 在服务器终端中应该看到：
   ```
   [时间] 新客户端已连接
   [时间] 收到消息:
     类型: connection
     状态: connected
   ```

3. **测试URL监控**
   - 打开一个新标签页
   - 访问任意网站（如 https://www.baidu.com）
   - 在服务器终端中应该看到URL变化的日志

## 🎯 常见问题

### Q1: 连接测试失败
**A:** 确保WebSocket服务器正在运行，检查地址和端口是否正确

### Q2: 状态显示"连接中..."
**A:** 
- 检查防火墙设置
- 确认服务器地址格式正确（必须以 ws:// 或 wss:// 开头）
- 查看浏览器控制台是否有错误信息

### Q3: 没有收到URL变化消息
**A:** 
- 确认"启用监控"开关已打开
- 检查服务器连接状态
- 在浏览器中打开 chrome://extensions/ 点击 Service Worker 查看日志

### Q4: 图标不显示
**A:** 这不影响功能使用，可以稍后添加图标文件

## 📊 查看日志

### 查看插件后台日志
```
Chrome: chrome://extensions/ → 找到插件 → 点击"Service Worker"
Edge: edge://extensions/ → 找到插件 → 点击"检查视图"
```

### 查看Popup日志
```
右键点击插件图标 → "检查弹出内容"
```

## 🔧 测试消息格式

服务器收到的消息示例：

**URL变化：**
```json
{
  "type": "url_change",
  "tabId": 12345,
  "url": "https://www.example.com",
  "title": "Example Page",
  "timestamp": "2025-11-15T10:30:00.000Z"
}
```

**新标签页：**
```json
{
  "type": "tab_created",
  "tabId": 12346,
  "url": "https://www.google.com",
  "timestamp": "2025-11-15T10:30:05.000Z"
}
```

## 🌐 生产环境部署

### 使用远程WebSocket服务器

1. 部署WebSocket服务器到云服务器
2. 使用 `wss://` 协议（加密连接）
3. 在插件配置中输入远程地址，例如：
   ```
   wss://your-domain.com/monitor
   ```

### Spring Boot示例配置

如果您使用Spring Boot作为后端服务器，参考 `README.md` 中的Spring Boot WebSocket配置示例。

## 📝 下一步

- ✅ 了解更多消息类型：查看 [README.md](./README.md)
- ✅ 自定义服务器逻辑：编辑 `example-server.js`
- ✅ 集成到现有项目：参考消息格式文档

## 💡 提示

- 开发时推荐使用 `nodemon` 自动重启服务器：`npm run dev`
- 可以在浏览器开发者工具中查看详细的日志信息
- WebSocket断开后会自动尝试重连，无需手动操作

---

**祝您使用愉快！** 🎉

如有问题，请查看完整文档 [README.md](./README.md) 或提交Issue。
