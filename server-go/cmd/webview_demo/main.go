package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	// 注意：需要先安装 webview 库
	// go get github.com/webview/webview
	// Windows 上需要安装 WebView2 Runtime: https://developer.microsoft.com/microsoft-edge/webview2/
)

// WebView2Demo WebView2 演示程序
type WebView2Demo struct {
	port int
}

// NewWebView2Demo 创建新的演示实例
func NewWebView2Demo() *WebView2Demo {
	return &WebView2Demo{
		port: 18889,
	}
}

func main() {
	if runtime.GOOS != "windows" {
		fmt.Println("❌ WebView2 仅支持 Windows 平台")
		os.Exit(1)
	}

	fmt.Println("🚀 启动 WebView2 演示程序")
	fmt.Println("=" + string(make([]byte, 50)) + "=")

	demo := NewWebView2Demo()

	// 启动本地服务器
	go demo.startServer()

	fmt.Printf("📡 本地服务器启动于: http://localhost:%d\n", demo.port)
	fmt.Println("🪟 准备创建 WebView2 窗口...")

	// 这里是实际使用 webview 的代码示例
	// 取消下面的注释来使用真实的 webview（需要先安装依赖）

	/*
		import "github.com/webview/webview"

		w := webview.New(true)
		defer w.Destroy()

		w.SetTitle("抖音直播监控 - WebView2 演示")
		w.SetSize(1200, 800, webview.HintNone)

		// 绑定 Go 函数到 JavaScript
		w.Bind("goMessage", func(msg string) string {
			log.Printf("📨 收到来自 JS 的消息: %s", msg)
			return fmt.Sprintf("Go 收到: %s", msg)
		})

		// 绑定数据查询函数
		w.Bind("getGiftRecords", func() string {
			records := []map[string]interface{}{
				{
					"time":     "11-21 15:30:00",
					"gift":     "玫瑰花",
					"count":    10,
					"diamond":  50,
					"receiver": "主播A",
					"sender":   "用户123",
				},
				{
					"time":     "11-21 15:31:00",
					"gift":     "豪华游艇",
					"count":    1,
					"diamond":  1000,
					"receiver": "主播B",
					"sender":   "用户456",
				},
			}
			data, _ := json.Marshal(records)
			return string(data)
		})

		// 加载页面
		w.Navigate(fmt.Sprintf("http://localhost:%d", demo.port))
		w.Run()
	*/

	// 模拟运行（因为实际 webview 需要依赖）
	fmt.Println("\n⚠️  WebView2 演示模式")
	fmt.Println("要运行真实的 WebView2 窗口，请:")
	fmt.Println("1. 安装 WebView2 Runtime: https://developer.microsoft.com/microsoft-edge/webview2/")
	fmt.Println("2. 安装 Go 包: go get github.com/webview/webview")
	fmt.Println("3. 取消 main.go 中的注释代码")
	fmt.Printf("4. 在浏览器中访问: http://localhost:%d\n", demo.port)

	// 保持服务器运行
	select {}
}

// startServer 启动本地 HTTP 服务器
func (d *WebView2Demo) startServer() {
	mux := http.NewServeMux()

	// 主页
	mux.HandleFunc("/", d.handleIndex)

	// API 端点
	mux.HandleFunc("/api/rooms", d.handleRooms)
	mux.HandleFunc("/api/gifts", d.handleGifts)
	mux.HandleFunc("/api/stats", d.handleStats)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", d.port),
		Handler: d.corsMiddleware(mux),
	}

	log.Printf("🌐 HTTP 服务器启动: http://localhost:%d", d.port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}

// corsMiddleware CORS 中间件
func (d *WebView2Demo) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleIndex 主页处理
func (d *WebView2Demo) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlTemplate)
}

// handleRooms 房间列表 API
func (d *WebView2Demo) handleRooms(w http.ResponseWriter, r *http.Request) {
	rooms := []map[string]interface{}{
		{
			"room_id":    "7404883888",
			"room_title": "测试直播间",
			"status":     "online",
			"viewers":    1234,
		},
		{
			"room_id":    "7404883999",
			"room_title": "另一个直播间",
			"status":     "offline",
			"viewers":    0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// handleGifts 礼物记录 API
func (d *WebView2Demo) handleGifts(w http.ResponseWriter, r *http.Request) {
	gifts := []map[string]interface{}{
		{
			"time":     "11-21 15:30:00",
			"gift":     "玫瑰花",
			"count":    10,
			"diamond":  50,
			"receiver": "主播A",
			"sender":   "用户123",
		},
		{
			"time":     "11-21 15:31:00",
			"gift":     "豪华游艇",
			"count":    1,
			"diamond":  1000,
			"receiver": "主播B",
			"sender":   "用户456",
		},
		{
			"time":     "11-21 15:32:00",
			"gift":     "跑车",
			"count":    2,
			"diamond":  2000,
			"receiver": "主播A",
			"sender":   "用户789",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gifts)
}

// handleStats 统计数据 API
func (d *WebView2Demo) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_rooms":  2,
		"online_rooms": 1,
		"total_gifts":  3,
		"total_value":  3050,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HTML 模板
const htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>抖音直播监控 - WebView2 演示</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Microsoft YaHei', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            padding: 30px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        }

        h1 {
            color: #667eea;
            margin-bottom: 10px;
            font-size: 32px;
        }

        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 16px;
        }

        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .stat-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 12px;
            text-align: center;
        }

        .stat-value {
            font-size: 36px;
            font-weight: bold;
            margin-bottom: 5px;
        }

        .stat-label {
            font-size: 14px;
            opacity: 0.9;
        }

        .section {
            margin-bottom: 30px;
        }

        .section-title {
            font-size: 20px;
            color: #333;
            margin-bottom: 15px;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            background: white;
            border-radius: 8px;
            overflow: hidden;
        }

        th {
            background: #667eea;
            color: white;
            padding: 12px;
            text-align: left;
        }

        td {
            padding: 12px;
            border-bottom: 1px solid #eee;
        }

        tr:hover {
            background: #f5f5f5;
        }

        .btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            margin: 5px;
        }

        .btn:hover {
            background: #5568d3;
        }

        .actions {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }

        #log {
            background: #2d2d2d;
            color: #00ff00;
            padding: 15px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 13px;
            max-height: 200px;
            overflow-y: auto;
            margin-top: 20px;
        }

        .log-entry {
            margin-bottom: 5px;
        }

        .online {
            color: #4caf50;
            font-weight: bold;
        }

        .offline {
            color: #f44336;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎉 抖音直播监控 - WebView2 演示</h1>
        <p class="subtitle">使用 WebView2 构建的现代化监控界面</p>

        <div class="stats" id="stats">
            <div class="stat-card">
                <div class="stat-value" id="totalRooms">-</div>
                <div class="stat-label">总房间数</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="onlineRooms">-</div>
                <div class="stat-label">在线房间</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="totalGifts">-</div>
                <div class="stat-label">礼物总数</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="totalValue">-</div>
                <div class="stat-label">总价值（钻石）</div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">📡 监控房间</h2>
            <div class="actions">
                <button class="btn" onclick="loadRooms()">🔄 刷新房间</button>
                <button class="btn" onclick="testGoBinding()">📤 测试 Go 通信</button>
            </div>
            <table id="roomsTable">
                <thead>
                    <tr>
                        <th>房间ID</th>
                        <th>房间标题</th>
                        <th>状态</th>
                        <th>观众数</th>
                    </tr>
                </thead>
                <tbody id="roomsBody">
                    <tr><td colspan="4" style="text-align:center;">加载中...</td></tr>
                </tbody>
            </table>
        </div>

        <div class="section">
            <h2 class="section-title">🎁 礼物记录</h2>
            <div class="actions">
                <button class="btn" onclick="loadGifts()">🔄 刷新礼物</button>
            </div>
            <table id="giftsTable">
                <thead>
                    <tr>
                        <th>时间</th>
                        <th>礼物</th>
                        <th>数量</th>
                        <th>钻石</th>
                        <th>接收主播</th>
                        <th>送礼用户</th>
                    </tr>
                </thead>
                <tbody id="giftsBody">
                    <tr><td colspan="6" style="text-align:center;">加载中...</td></tr>
                </tbody>
            </table>
        </div>

        <div class="section">
            <h2 class="section-title">📋 日志输出</h2>
            <div id="log"></div>
        </div>
    </div>

    <script>
        // 日志输出
        function addLog(message) {
            const log = document.getElementById('log');
            const entry = document.createElement('div');
            entry.className = 'log-entry';
            const timestamp = new Date().toLocaleTimeString('zh-CN');
            entry.textContent = '[' + timestamp + '] ' + message;
            log.appendChild(entry);
            log.scrollTop = log.scrollHeight;
        }

        // 加载统计数据
        async function loadStats() {
            try {
                const response = await fetch('/api/stats');
                const data = await response.json();
                
                document.getElementById('totalRooms').textContent = data.total_rooms;
                document.getElementById('onlineRooms').textContent = data.online_rooms;
                document.getElementById('totalGifts').textContent = data.total_gifts;
                document.getElementById('totalValue').textContent = data.total_value.toLocaleString();
                
                addLog('✅ 统计数据已更新');
            } catch (error) {
                addLog('❌ 加载统计数据失败: ' + error.message);
            }
        }

        // 加载房间列表
        async function loadRooms() {
            try {
                const response = await fetch('/api/rooms');
                const rooms = await response.json();
                
                const tbody = document.getElementById('roomsBody');
                tbody.innerHTML = '';
                
                rooms.forEach(room => {
                    const row = tbody.insertRow();
                    row.innerHTML = 
                        '<td>' + room.room_id + '</td>' +
                        '<td>' + room.room_title + '</td>' +
                        '<td class="' + room.status + '">' + (room.status === 'online' ? '在线' : '离线') + '</td>' +
                        '<td>' + room.viewers.toLocaleString() + '</td>';
                });
                
                addLog('✅ 房间列表已刷新 (' + rooms.length + ' 个房间)');
            } catch (error) {
                addLog('❌ 加载房间失败: ' + error.message);
            }
        }

        // 加载礼物记录
        async function loadGifts() {
            try {
                const response = await fetch('/api/gifts');
                const gifts = await response.json();
                
                const tbody = document.getElementById('giftsBody');
                tbody.innerHTML = '';
                
                gifts.forEach(gift => {
                    const row = tbody.insertRow();
                    row.innerHTML = 
                        '<td>' + gift.time + '</td>' +
                        '<td>' + gift.gift + '</td>' +
                        '<td>' + gift.count + '</td>' +
                        '<td>' + gift.diamond + '</td>' +
                        '<td>' + gift.receiver + '</td>' +
                        '<td>' + gift.sender + '</td>';
                });
                
                addLog('✅ 礼物记录已刷新 (' + gifts.length + ' 条记录)');
            } catch (error) {
                addLog('❌ 加载礼物失败: ' + error.message);
            }
        }

        // 测试 Go 绑定（需要真实的 webview）
        function testGoBinding() {
            addLog('📤 尝试调用 Go 函数...');
            
            // 这需要在真实的 webview 环境中运行
            if (typeof goMessage !== 'undefined') {
                const response = goMessage('Hello from JavaScript!');
                addLog('📨 Go 响应: ' + response);
            } else {
                addLog('⚠️  Go 绑定不可用（当前为浏览器模式）');
                addLog('💡 提示：需要在 WebView2 窗口中运行才能使用 Go 绑定');
            }
        }

        // 页面加载时初始化
        window.addEventListener('load', () => {
            addLog('🚀 WebView2 演示页面已加载');
            addLog('📊 开始加载数据...');
            
            loadStats();
            loadRooms();
            loadGifts();
            
            // 每 5 秒自动刷新一次
            setInterval(() => {
                loadStats();
            }, 5000);
        });
    </script>
</body>
</html>
`
