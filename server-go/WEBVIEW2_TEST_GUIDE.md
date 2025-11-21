# WebView2 测试指南

## 概述

本指南介绍如何在 server-go 项目中使用和测试 WebView2 窗口。WebView2 是 Microsoft 提供的现代化 Web 渲染引擎，可以在 Go 应用程序中嵌入完整的 Web 浏览器功能。

## 项目结构

```
server-go/
├── internal/ui/
│   └── webview_test.go          # WebView2 单元测试
├── cmd/webview_demo/
│   └── main.go                  # WebView2 演示程序
└── WEBVIEW2_TEST_GUIDE.md      # 本文档
```

## 前置要求

### 1. 系统要求

- **操作系统**: Windows 10/11
- **WebView2 Runtime**: 必须安装

### 2. 安装 WebView2 Runtime

**方法 A: 自动下载安装**
```powershell
# 使用 PowerShell
Invoke-WebRequest -Uri "https://go.microsoft.com/fwlink/p/?LinkId=2124703" -OutFile "MicrosoftEdgeWebview2Setup.exe"
.\MicrosoftEdgeWebview2Setup.exe
```

**方法 B: 手动下载**
访问: https://developer.microsoft.com/microsoft-edge/webview2/

### 3. 安装 Go WebView 库

```bash
cd server-go
go get github.com/webview/webview
```

## 运行测试

### 1. 运行单元测试

```bash
# 运行所有 WebView2 测试
go test -v ./internal/ui -run TestWebView2

# 运行特定测试
go test -v ./internal/ui -run TestWebView2BasicWindow

# 运行基准测试
go test -v ./internal/ui -bench BenchmarkWebView2Creation
```

### 2. 运行演示程序

#### 方法 A: 浏览器模式（无需 WebView2）
```bash
cd cmd/webview_demo
go run main.go
```

然后在浏览器中访问: http://localhost:18889

#### 方法 B: WebView2 窗口模式
1. 取消 `cmd/webview_demo/main.go` 中的注释代码
2. 编译运行:
```bash
cd cmd/webview_demo
go build -o webview_demo.exe
./webview_demo.exe
```

## 测试内容

### 1. 基础窗口测试 (TestWebView2BasicWindow)

测试 WebView2 窗口的基本创建和配置：
- 窗口标题设置
- 窗口尺寸设置
- 调试模式开关

### 2. HTML 加载测试 (TestWebView2WithHTML)

测试加载 HTML 内容的能力：
- 内联 HTML 渲染
- CSS 样式支持
- JavaScript 执行
- 交互功能

### 3. 通信测试 (TestWebView2Communication)

测试 Go 和 JavaScript 之间的双向通信：
- JavaScript 调用 Go 函数
- Go 向 JavaScript 传递数据
- JSON 数据序列化/反序列化

### 4. 本地服务器测试 (TestWebView2WithLocalServer)

测试通过本地 HTTP 服务器加载页面：
- HTTP 服务器启动
- 页面加载
- API 端点调用

### 5. 多窗口测试 (TestWebView2MultipleWindows)

测试同时创建和管理多个 WebView2 窗口。

### 6. 性能测试 (TestWebView2Performance)

测试 WebView2 的性能指标：
- 窗口创建时间
- 内存占用
- 响应速度

## 演示程序功能

### 主要功能

1. **统计面板**
   - 总房间数
   - 在线房间数
   - 礼物总数
   - 总价值（钻石）

2. **房间监控**
   - 房间列表显示
   - 在线状态
   - 观众数统计

3. **礼物记录**
   - 礼物详细信息
   - 送礼用户
   - 接收主播

4. **日志系统**
   - 实时日志输出
   - 操作记录
   - 错误提示

### API 端点

- `GET /` - 主页面
- `GET /api/rooms` - 房间列表
- `GET /api/gifts` - 礼物记录
- `GET /api/stats` - 统计数据

## Go 和 JavaScript 通信示例

### Go 端绑定函数

```go
// 在 Go 中定义可被 JavaScript 调用的函数
w.Bind("goMessage", func(msg string) string {
    log.Printf("收到来自 JS 的消息: %s", msg)
    return fmt.Sprintf("Go 收到: %s", msg)
})

w.Bind("getGiftRecords", func() string {
    records := getGiftRecordsFromDB()
    data, _ := json.Marshal(records)
    return string(data)
})
```

### JavaScript 端调用

```javascript
// 在 JavaScript 中调用 Go 函数
const response = goMessage("Hello from JS!");
console.log(response); // "Go 收到: Hello from JS!"

// 获取数据
const gifts = JSON.parse(getGiftRecords());
console.log(gifts);
```

## 实际应用场景

### 场景 1: 实时监控面板

```go
// 实时推送数据到前端
w.Bind("subscribeUpdates", func() {
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        for range ticker.C {
            data := getCurrentStats()
            js := fmt.Sprintf("updateStats(%s)", data)
            w.Eval(js)
        }
    }()
})
```

### 场景 2: 数据可视化

```go
// 提供图表数据
w.Bind("getChartData", func(roomID string) string {
    data := getGiftTrendsChart(roomID)
    json, _ := json.Marshal(data)
    return string(json)
})
```

### 场景 3: 配置管理

```go
// 保存配置
w.Bind("saveConfig", func(configJSON string) bool {
    var config Config
    if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
        return false
    }
    return saveConfigToDB(config)
})
```

## 常见问题

### Q1: WebView2 Runtime 未安装

**错误信息**:
```
WebView2 Runtime is not installed
```

**解决方法**:
安装 WebView2 Runtime（见"前置要求"部分）

### Q2: CGO 编译错误

**错误信息**:
```
# github.com/webview/webview
cgo: C compiler not found
```

**解决方法**:
安装 MinGW-w64 或 TDM-GCC

### Q3: 窗口无法显示

**可能原因**:
- 防火墙阻止
- 端口被占用
- 权限不足

**解决方法**:
```bash
# 检查端口
netstat -ano | findstr :18889

# 以管理员身份运行
```

### Q4: JavaScript 无法调用 Go 函数

**可能原因**:
- 函数未正确绑定
- 在浏览器模式下运行（而非 WebView2）

**解决方法**:
确保在 `w.Bind()` 之后再 `w.Navigate()`

## 性能优化建议

### 1. 窗口创建优化

```go
// 使用窗口池
var windowPool = sync.Pool{
    New: func() interface{} {
        return webview.New(false)
    },
}
```

### 2. 数据传输优化

```go
// 使用压缩传输大数据
w.Bind("getLargeData", func() string {
    data := getLargeDataset()
    compressed := compress(data)
    return base64.StdEncoding.EncodeToString(compressed)
})
```

### 3. 内存管理

```go
// 及时释放资源
defer w.Destroy()

// 定期清理缓存
w.Eval("localStorage.clear()")
```

## 安全考虑

### 1. 内容安全策略 (CSP)

```html
<meta http-equiv="Content-Security-Policy" 
      content="default-src 'self'; script-src 'self' 'unsafe-inline'">
```

### 2. 输入验证

```go
w.Bind("processInput", func(input string) string {
    // 验证和清理输入
    cleaned := sanitizeInput(input)
    return processData(cleaned)
})
```

### 3. HTTPS 通信

```go
// 使用 HTTPS 加载远程资源
w.Navigate("https://your-secure-domain.com")
```

## 调试技巧

### 1. 启用开发者工具

```go
w := webview.New(true) // true 启用调试模式
```

### 2. 控制台日志

```javascript
// 在 JavaScript 中
console.log("Debug info:", data);

// 在 Go 中查看
w.Bind("log", func(msg string) {
    log.Println("[WebView JS]", msg)
})
```

### 3. 网络监控

在开发者工具的 Network 标签中查看所有网络请求。

## 进阶用法

### 1. 自定义协议

```go
// 注册自定义协议处理
w.Bind("openRoom", func(roomID string) {
    // app://open-room/123456
    openRoomInNewWindow(roomID)
})
```

### 2. 文件操作

```go
w.Bind("saveFile", func(filename, content string) bool {
    return ioutil.WriteFile(filename, []byte(content), 0644) == nil
})

w.Bind("loadFile", func(filename string) string {
    data, _ := ioutil.ReadFile(filename)
    return string(data)
})
```

### 3. 系统集成

```go
w.Bind("showNotification", func(title, message string) {
    // 显示系统通知
    showSystemNotification(title, message)
})
```

## 资源链接

- **WebView2 官方文档**: https://docs.microsoft.com/microsoft-edge/webview2/
- **Go WebView 库**: https://github.com/webview/webview
- **示例项目**: https://github.com/webview/webview/tree/master/examples

## 总结

WebView2 为 Go 应用提供了强大的 Web UI 能力：
- ✅ 现代化的 UI 界面
- ✅ 跨平台支持（主要是 Windows）
- ✅ 完整的 Web 技术栈
- ✅ Go 和 JavaScript 无缝通信
- ✅ 高性能渲染引擎

使用本指南中的测试和示例，你可以快速在 server-go 项目中集成 WebView2 功能。
