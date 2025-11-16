package fallback

import (
	"dy-live-monitor/internal/database"
	"dy-live-monitor/internal/parser"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"

	webview "github.com/webview/webview_go"
)

// FallbackManager Fallback WebView2 管理器
type FallbackManager struct {
	db           *database.DB
	roomID       string
	webview      webview.WebView
	parser       *parser.DouyinParser
	isRunning    bool
	mu           sync.Mutex
	dataCallback func([]byte) // 数据回调函数
}

// NewFallbackManager 创建 Fallback 管理器
func NewFallbackManager(db *database.DB, roomID string) *FallbackManager {
	return &FallbackManager{
		db:        db,
		roomID:    roomID,
		parser:    parser.NewDouyinParser(),
		isRunning: false,
	}
}

// SetDataCallback 设置数据回调
func (f *FallbackManager) SetDataCallback(callback func([]byte)) {
	f.dataCallback = callback
}

// Start 启动 Fallback WebView2 实例
func (f *FallbackManager) Start() error {
	f.mu.Lock()
	if f.isRunning {
		f.mu.Unlock()
		return fmt.Errorf("fallback already running")
	}
	f.isRunning = true
	f.mu.Unlock()

	log.Printf("🔄 [Fallback] 启动 WebView2 备用数据通道 (房间: %s)", f.roomID)

	// 创建隐藏的 WebView2 窗口
	f.webview = webview.New(false)
	defer f.webview.Destroy()

	// 设置极小窗口（几乎隐藏）
	f.webview.SetTitle(fmt.Sprintf("Fallback - Room %s", f.roomID))
	f.webview.SetSize(1, 1, webview.HintNone)

	// 绑定消息接收函数
	f.webview.Bind("sendToGo", func(data string) {
		// data 是 Base64 编码的 WebSocket 消息
		f.handleWebSocketMessage(data)
	})

	// 注入 JavaScript 拦截 WebSocket
	injectedJS := f.generateInjectionScript()
	f.webview.Init(injectedJS)

	// 加载直播间页面
	url := fmt.Sprintf("https://live.douyin.com/%s", f.roomID)
	f.webview.Navigate(url)

	log.Printf("✅ [Fallback] WebView2 已加载: %s", url)

	// 启动心跳检测（确保 Fallback 正常工作）
	go f.heartbeat()

	// 运行 WebView2 主循环
	f.webview.Run()

	f.mu.Lock()
	f.isRunning = false
	f.mu.Unlock()

	log.Printf("⏹️  [Fallback] WebView2 已停止 (房间: %s)", f.roomID)
	return nil
}

// Stop 停止 Fallback
func (f *FallbackManager) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.isRunning && f.webview != nil {
		f.webview.Terminate()
		log.Printf("🛑 [Fallback] 手动停止 (房间: %s)", f.roomID)
	}
}

// IsRunning 检查是否正在运行
func (f *FallbackManager) IsRunning() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.isRunning
}

// generateInjectionScript 生成注入脚本
func (f *FallbackManager) generateInjectionScript() string {
	return `
(function() {
	console.log('[Fallback] 注入 WebSocket 拦截脚本');

	// 保存原始 WebSocket
	const OriginalWebSocket = window.WebSocket;

	// 重写 WebSocket 构造函数
	window.WebSocket = function(url, protocols) {
		console.log('[Fallback] WebSocket 连接:', url);

		// 创建原始 WebSocket 实例
		const ws = new OriginalWebSocket(url, protocols);

		// 拦截 message 事件
		ws.addEventListener('message', function(event) {
			try {
				// 检查是否是抖音 WebSocket
				if (url.includes('webcast') || url.includes('douyin')) {
					// 将 ArrayBuffer 或 Blob 转换为 Base64
					if (event.data instanceof ArrayBuffer) {
						const bytes = new Uint8Array(event.data);
						const binary = String.fromCharCode.apply(null, bytes);
						const base64 = btoa(binary);
						
						// 发送到 Go 后端
						sendToGo(base64);
					} else if (event.data instanceof Blob) {
						const reader = new FileReader();
						reader.onloadend = function() {
							const bytes = new Uint8Array(reader.result);
							const binary = String.fromCharCode.apply(null, bytes);
							const base64 = btoa(binary);
							sendToGo(base64);
						};
						reader.readAsArrayBuffer(event.data);
					}
				}
			} catch (e) {
				console.error('[Fallback] 消息处理失败:', e);
			}
		});

		return ws;
	};

	// 保留原型链
	window.WebSocket.prototype = OriginalWebSocket.prototype;

	console.log('[Fallback] WebSocket 拦截脚本注入完成');
})();
`
}

// handleWebSocketMessage 处理拦截到的 WebSocket 消息
func (f *FallbackManager) handleWebSocketMessage(base64Data string) {
	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		log.Printf("❌ [Fallback] Base64 解码失败: %v", err)
		return
	}

	// 解析抖音消息
	url := fmt.Sprintf("https://live.douyin.com/%s", f.roomID)
	parsedMessages, err := f.parser.ParseMessage(base64Data, url)
	if err != nil {
		log.Printf("❌ [Fallback] 消息解析失败: %v", err)
		return
	}

	if len(parsedMessages) > 0 {
		log.Printf("✅ [Fallback] 成功解析 %d 条消息", len(parsedMessages))

		// 打印格式化消息
		formatted := f.parser.FormatMessage(parsedMessages)
		if formatted != "" {
			log.Println(formatted)
		}

		// 如果有回调函数，调用它
		if f.dataCallback != nil {
			f.dataCallback(data)
		}
	}
}

// heartbeat 心跳检测
func (f *FallbackManager) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !f.IsRunning() {
				return
			}
			log.Printf("💓 [Fallback] 心跳检测 (房间: %s)", f.roomID)
		}
	}
}
