# WebView2 演示程序

这是一个使用 WebView2 构建的抖音直播监控演示程序。

## 快速开始

### 方法 1: 浏览器模式（推荐用于开发）

```bash
# 在 server-go 目录下
go run cmd/webview_demo/main.go
```

然后在浏览器中访问: http://localhost:18889

### 方法 2: WebView2 窗口模式

#### 前置条件

1. **安装 WebView2 Runtime**
   - 下载: https://developer.microsoft.com/microsoft-edge/webview2/
   - 或使用 PowerShell 自动安装:
     ```powershell
     Invoke-WebRequest -Uri "https://go.microsoft.com/fwlink/p/?LinkId=2124703" -OutFile "MicrosoftEdgeWebview2Setup.exe"
     .\MicrosoftEdgeWebview2Setup.exe
     ```

2. **安装 Go WebView 库**
   ```bash
   go get github.com/webview/webview
   ```

#### 启用 WebView2 窗口

1. 编辑 `main.go`
2. 取消第 33-58 行的注释（/* 到 */）
3. 编译运行:
   ```bash
   go build -o webview_demo.exe
   ./webview_demo.exe
   ```

## 功能特性

### 📊 实时统计
- 总房间数
- 在线房间数
- 礼物总数
- 总价值（钻石）

### 📡 房间监控
- 实时房间列表
- 在线状态显示
- 观众数统计

### 🎁 礼物记录
- 礼物详细信息
- 送礼用户
- 接收主播
- 钻石价值

### 📋 日志系统
- 实时操作日志
- 错误提示
- 时间戳记录

## API 接口

### GET /
主页面

### GET /api/rooms
获取房间列表

**响应示例**:
```json
[
  {
    "room_id": "7404883888",
    "room_title": "测试直播间",
    "status": "online",
    "viewers": 1234
  }
]
```

### GET /api/gifts
获取礼物记录

**响应示例**:
```json
[
  {
    "time": "11-21 15:30:00",
    "gift": "玫瑰花",
    "count": 10,
    "diamond": 50,
    "receiver": "主播A",
    "sender": "用户123"
  }
]
```

### GET /api/stats
获取统计数据

**响应示例**:
```json
{
  "total_rooms": 2,
  "online_rooms": 1,
  "total_gifts": 3,
  "total_value": 3050
}
```

## Go 和 JavaScript 通信

### 从 JavaScript 调用 Go

```javascript
// 调用 Go 函数
const response = goMessage("Hello from JS!");
console.log(response); // "Go 收到: Hello from JS!"

// 获取数据
const gifts = JSON.parse(getGiftRecords());
console.log(gifts);
```

### 从 Go 调用 JavaScript

```go
// 执行 JavaScript 代码
w.Eval("alert('Hello from Go!');")

// 更新数据
data := getStats()
js := fmt.Sprintf("updateStats(%s)", data)
w.Eval(js)
```

## 自定义开发

### 添加新的 API 端点

```go
// 在 main.go 的 startServer 函数中添加
mux.HandleFunc("/api/custom", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "自定义端点"
    })
})
```

### 绑定新的 Go 函数

```go
// 取消注释后，在 main 函数中添加
w.Bind("customFunction", func(param string) string {
    // 处理逻辑
    return "result"
})
```

### 修改 UI 样式

编辑 `htmlTemplate` 常量中的 CSS 部分。

## 故障排除

### 问题 1: 端口被占用

**错误信息**: `listen tcp :18889: bind: address already in use`

**解决方法**:
```bash
# 查找占用端口的进程
netstat -ano | findstr :18889

# 结束进程
taskkill /PID <进程ID> /F

# 或修改端口号
# 在 main.go 中修改 port: 18889 为其他值
```

### 问题 2: WebView2 Runtime 未安装

**错误信息**: `WebView2 Runtime is not installed`

**解决方法**: 安装 WebView2 Runtime（见"前置条件"）

### 问题 3: CORS 错误

**错误信息**: `Access to fetch at ... has been blocked by CORS policy`

**解决方法**: 已在代码中配置 CORS 中间件，如果仍有问题，检查请求 URL 是否正确。

## 性能优化

### 减少 API 调用频率

```javascript
// 修改自动刷新间隔（默认 5 秒）
setInterval(() => {
    loadStats();
}, 10000); // 改为 10 秒
```

### 使用缓存

```javascript
// 添加简单的缓存机制
let cachedData = null;
let cacheTime = 0;
const CACHE_TTL = 5000; // 5 秒

async function loadDataWithCache() {
    const now = Date.now();
    if (cachedData && (now - cacheTime < CACHE_TTL)) {
        return cachedData;
    }
    
    const response = await fetch('/api/data');
    cachedData = await response.json();
    cacheTime = now;
    return cachedData;
}
```

## 安全建议

1. **不要在生产环境中使用默认配置**
2. **添加身份验证**
3. **使用 HTTPS**
4. **验证所有输入数据**
5. **限制 API 访问频率**

## 扩展阅读

- [WebView2 官方文档](https://docs.microsoft.com/microsoft-edge/webview2/)
- [Go WebView 库](https://github.com/webview/webview)
- [完整测试指南](../../WEBVIEW2_TEST_GUIDE.md)

## 许可证

与主项目相同
