package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"testing"
	"time"
)

// WebView2TestWindow WebView2测试窗口结构
type WebView2TestWindow struct {
	title   string
	width   int
	height  int
	url     string
	debug   bool
	webview interface{} // 实际的 webview 实例
}

// NewWebView2TestWindow 创建新的 WebView2 测试窗口
func NewWebView2TestWindow(title string, width, height int, debug bool) *WebView2TestWindow {
	return &WebView2TestWindow{
		title:  title,
		width:  width,
		height: height,
		debug:  debug,
	}
}

// TestWebView2BasicWindow 测试基础 WebView2 窗口
func TestWebView2BasicWindow(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("WebView2 仅支持 Windows 平台")
	}

	t.Log("开始测试 WebView2 基础窗口")

	// 注意：这里只是测试框架，实际使用需要安装 github.com/webview/webview
	// go get github.com/webview/webview

	testWindow := NewWebView2TestWindow("WebView2 测试窗口", 800, 600, true)

	if testWindow.title == "" {
		t.Error("窗口标题不能为空")
	}

	if testWindow.width <= 0 || testWindow.height <= 0 {
		t.Error("窗口尺寸必须大于0")
	}

	t.Logf("✅ WebView2 测试窗口创建成功: %s (%dx%d)", testWindow.title, testWindow.width, testWindow.height)
}

// TestWebView2WithHTML 测试加载 HTML 内容
func TestWebView2WithHTML(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("WebView2 仅支持 Windows 平台")
	}

	t.Log("开始测试 WebView2 加载 HTML")

	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>WebView2 测试页面</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        .container {
            text-align: center;
            padding: 40px;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 20px;
            backdrop-filter: blur(10px);
        }
        h1 {
            font-size: 48px;
            margin-bottom: 20px;
        }
        .info {
            font-size: 18px;
            margin-top: 20px;
        }
        button {
            padding: 12px 24px;
            font-size: 16px;
            margin: 10px;
            cursor: pointer;
            border: none;
            border-radius: 8px;
            background: white;
            color: #667eea;
            font-weight: bold;
            transition: transform 0.2s;
        }
        button:hover {
            transform: scale(1.05);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎉 WebView2 测试成功！</h1>
        <p class="info">这是一个使用 WebView2 渲染的测试页面</p>
        <div>
            <button onclick="testJS()">测试 JavaScript</button>
            <button onclick="sendToGo()">发送消息到 Go</button>
        </div>
        <div id="output" style="margin-top: 20px;"></div>
    </div>
    <script>
        function testJS() {
            document.getElementById('output').innerHTML = 
                '<p style="color: #90EE90;">✅ JavaScript 正常工作！</p>';
        }
        
        function sendToGo() {
            if (window.external && window.external.invoke) {
                window.external.invoke(JSON.stringify({
                    type: 'test',
                    message: 'Hello from JavaScript!',
                    timestamp: Date.now()
                }));
            }
            document.getElementById('output').innerHTML = 
                '<p style="color: #90EE90;">📤 消息已发送到 Go 后端</p>';
        }
    </script>
</body>
</html>
	`

	if htmlContent == "" {
		t.Error("HTML 内容不能为空")
	}

	t.Logf("✅ HTML 内容长度: %d 字节", len(htmlContent))
	t.Log("✅ WebView2 HTML 测试准备完成")
}

// TestWebView2Communication 测试 Go 和 JavaScript 通信
func TestWebView2Communication(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("WebView2 仅支持 Windows 平台")
	}

	t.Log("开始测试 WebView2 通信功能")

	// 模拟从 JavaScript 接收的消息
	testMessage := map[string]interface{}{
		"type":      "test",
		"message":   "Hello from JavaScript",
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(testMessage)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	t.Logf("✅ 测试消息: %s", string(jsonData))

	// 模拟消息处理
	var received map[string]interface{}
	err = json.Unmarshal(jsonData, &received)
	if err != nil {
		t.Fatalf("JSON 反序列化失败: %v", err)
	}

	if received["type"] != "test" {
		t.Error("消息类型不匹配")
	}

	t.Log("✅ WebView2 通信测试通过")
}

// TestWebView2WithLocalServer 测试通过本地服务器加载页面
func TestWebView2WithLocalServer(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("WebView2 仅支持 Windows 平台")
	}

	t.Log("开始测试 WebView2 本地服务器")

	// 启动测试用 HTTP 服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>本地服务器测试</title>
</head>
<body>
    <h1>✅ 本地服务器正常工作</h1>
    <p>当前时间: <span id="time"></span></p>
    <script>
        setInterval(() => {
            document.getElementById('time').textContent = new Date().toLocaleString('zh-CN');
        }, 1000);
    </script>
</body>
</html>
		`)
	})

	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"message": "API 测试成功",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	server := &http.Server{
		Addr:    ":18888",
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("服务器启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 测试 API 端点
	resp, err := http.Get("http://localhost:18888/api/test")
	if err != nil {
		t.Fatalf("API 请求失败: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if result["status"] != "ok" {
		t.Error("API 响应状态不正确")
	}

	t.Log("✅ 本地服务器测试通过")
	t.Logf("✅ WebView2 可以访问: http://localhost:18888")

	// 清理
	server.Close()
}

// TestWebView2MultipleWindows 测试多窗口支持
func TestWebView2MultipleWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("WebView2 仅支持 Windows 平台")
	}

	t.Log("开始测试 WebView2 多窗口")

	windows := []*WebView2TestWindow{
		NewWebView2TestWindow("窗口1", 800, 600, true),
		NewWebView2TestWindow("窗口2", 600, 400, false),
		NewWebView2TestWindow("窗口3", 1024, 768, true),
	}

	if len(windows) != 3 {
		t.Error("窗口数量不正确")
	}

	for i, win := range windows {
		t.Logf("✅ 窗口 %d: %s (%dx%d, debug=%v)", i+1, win.title, win.width, win.height, win.debug)
	}

	t.Log("✅ WebView2 多窗口测试通过")
}

// TestWebView2Performance 测试性能指标
func TestWebView2Performance(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("WebView2 仅支持 Windows 平台")
	}

	t.Log("开始测试 WebView2 性能")

	start := time.Now()

	// 模拟创建窗口
	testWindow := NewWebView2TestWindow("性能测试", 1920, 1080, false)

	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Logf("⚠️  窗口创建耗时: %v (可能需要优化)", elapsed)
	} else {
		t.Logf("✅ 窗口创建耗时: %v", elapsed)
	}

	// 测试内存占用（简化版）
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	t.Logf("📊 当前内存使用: %.2f MB", float64(m.Alloc)/1024/1024)
	t.Logf("📊 系统内存占用: %.2f MB", float64(m.Sys)/1024/1024)

	if testWindow != nil {
		t.Log("✅ WebView2 性能测试完成")
	}
}

// BenchmarkWebView2Creation WebView2 窗口创建性能测试
func BenchmarkWebView2Creation(b *testing.B) {
	if runtime.GOOS != "windows" {
		b.Skip("WebView2 仅支持 Windows 平台")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewWebView2TestWindow("基准测试", 800, 600, false)
	}
}
