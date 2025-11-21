# WebView2 快速开始 ⚡

## 5 分钟快速测试

### 步骤 1: 安装依赖（首次）

```bash
# 安装 WebView2 Runtime (Windows)
# 访问: https://developer.microsoft.com/microsoft-edge/webview2/

# 安装 Go 库
go get github.com/webview/webview
```

### 步骤 2: 运行测试

**Windows 用户**:
```bash
# 双击运行
TEST_WEBVIEW2.bat

# 或命令行
go test -v ./internal/ui -run TestWebView2
```

**Linux/Mac 用户**:
```bash
# 只能进行语法检查
./TEST_WEBVIEW2.sh
```

### 步骤 3: 启动演示

```bash
# 方法 A: 浏览器模式（无需 WebView2）
go run cmd/webview_demo/main.go
# 然后访问: http://localhost:18889

# 方法 B: WebView2 窗口模式（需要取消注释）
cd cmd/webview_demo
go build -o webview_demo.exe
./webview_demo.exe
```

## 项目结构

```
server-go/
├── internal/ui/
│   └── webview_test.go           # 单元测试 ✅
├── cmd/webview_demo/
│   ├── main.go                   # 演示程序 ✅
│   └── README.md                 # 演示文档 ✅
├── TEST_WEBVIEW2.bat             # Windows 测试脚本 ✅
├── TEST_WEBVIEW2.sh              # Linux/Mac 脚本 ✅
├── WEBVIEW2_TEST_GUIDE.md        # 完整指南 ✅
└── WEBVIEW2_QUICK_START.md       # 本文档 ✅
```

## 测试清单

### ✅ 单元测试 (6个)

| 测试名称 | 功能 | 命令 |
|---------|------|------|
| TestWebView2BasicWindow | 基础窗口创建 | `go test -run TestWebView2BasicWindow` |
| TestWebView2WithHTML | HTML 加载 | `go test -run TestWebView2WithHTML` |
| TestWebView2Communication | Go-JS 通信 | `go test -run TestWebView2Communication` |
| TestWebView2WithLocalServer | 本地服务器 | `go test -run TestWebView2WithLocalServer` |
| TestWebView2MultipleWindows | 多窗口支持 | `go test -run TestWebView2MultipleWindows` |
| TestWebView2Performance | 性能指标 | `go test -run TestWebView2Performance` |

### ✅ 基准测试 (1个)

```bash
go test -bench BenchmarkWebView2Creation -benchmem
```

## 演示功能

### 📊 实时统计面板
- 总房间数
- 在线房间数  
- 礼物总数
- 总价值（钻石）

### 📡 房间监控
- 房间列表
- 在线状态
- 观众数

### 🎁 礼物记录
- 详细信息
- 送礼用户
- 接收主播

### 📋 日志系统
- 实时输出
- 操作记录

## API 端点

| 端点 | 方法 | 说明 |
|-----|------|------|
| `/` | GET | 主页面 |
| `/api/rooms` | GET | 房间列表 |
| `/api/gifts` | GET | 礼物记录 |
| `/api/stats` | GET | 统计数据 |

## Go ↔ JavaScript 通信示例

### JavaScript 调用 Go

```javascript
// 在 WebView2 窗口中
const response = goMessage("Hello!");
console.log(response);

const gifts = JSON.parse(getGiftRecords());
console.log(gifts);
```

### Go 调用 JavaScript

```go
// 在 Go 代码中
w.Eval("alert('Hello from Go!');")

data := getStats()
js := fmt.Sprintf("updateStats(%s)", data)
w.Eval(js)
```

## 常见命令

```bash
# 运行所有测试
go test -v ./internal/ui -run TestWebView2

# 启动演示（浏览器模式）
go run cmd/webview_demo/main.go

# 编译演示程序
cd cmd/webview_demo && go build -o webview_demo.exe

# 格式化代码
go fmt ./internal/ui/webview_test.go

# 查看测试覆盖率
go test -cover ./internal/ui
```

## 快速排查

### ❌ 测试跳过: "WebView2 仅支持 Windows 平台"
**原因**: 在非 Windows 系统上运行  
**解决**: 在 Windows 上运行，或只进行语法检查

### ❌ 端口被占用: "bind: address already in use"
**原因**: 端口 18889 已被占用  
**解决**: 
```bash
netstat -ano | findstr :18889
taskkill /PID <进程ID> /F
```

### ❌ WebView2 Runtime 未安装
**原因**: 系统未安装 WebView2 Runtime  
**解决**: 访问 https://developer.microsoft.com/microsoft-edge/webview2/

### ❌ Go 库未安装: "cannot find package"
**原因**: webview 包未安装  
**解决**: 
```bash
go get github.com/webview/webview
```

## 性能指标

### 预期性能

| 指标 | 目标值 | 说明 |
|-----|--------|------|
| 窗口创建时间 | < 100ms | 首次创建 |
| 内存占用 | < 50MB | 基础窗口 |
| API 响应时间 | < 10ms | 本地请求 |
| UI 刷新率 | 60 FPS | 流畅渲染 |

### 基准测试结果示例

```
BenchmarkWebView2Creation-8    1000    1234567 ns/op    12345 B/op    123 allocs/op
```

## 下一步

### 🚀 开发集成

1. 在主程序中集成 WebView2
2. 连接到真实的数据库
3. 实现实时数据推送
4. 添加用户认证

### 📚 学习资源

- **完整指南**: [WEBVIEW2_TEST_GUIDE.md](WEBVIEW2_TEST_GUIDE.md)
- **演示文档**: [cmd/webview_demo/README.md](cmd/webview_demo/README.md)
- **官方文档**: https://docs.microsoft.com/microsoft-edge/webview2/
- **Go 库文档**: https://github.com/webview/webview

### 🔧 自定义开发

```go
// 添加新的 Go 函数绑定
w.Bind("myFunction", func(param string) string {
    // 你的逻辑
    return "result"
})

// 添加新的 API 端点
mux.HandleFunc("/api/myendpoint", func(w http.ResponseWriter, r *http.Request) {
    // 你的逻辑
    json.NewEncoder(w).Encode(data)
})
```

## 技术栈

- **Go**: 1.24.2+
- **WebView2**: Microsoft Edge WebView2
- **前端**: HTML5 + CSS3 + JavaScript (ES6+)
- **通信**: HTTP/REST API + JavaScript Bridge

## 支持平台

| 平台 | WebView2 窗口 | 浏览器模式 | 单元测试 |
|-----|--------------|-----------|---------|
| Windows 10/11 | ✅ | ✅ | ✅ |
| Linux | ❌ | ✅ | ⚠️ (仅语法) |
| macOS | ❌ | ✅ | ⚠️ (仅语法) |

## 许可证

与主项目相同

---

**快速链接**:
- [完整测试指南](WEBVIEW2_TEST_GUIDE.md)
- [演示程序文档](cmd/webview_demo/README.md)
- [主项目 README](README.md)

**问题反馈**: 在项目 Issues 中提交

**最后更新**: 2025-11-21
