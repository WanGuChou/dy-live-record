# 下一步开发建议

## 🎯 立即可测试的功能

当前项目已完成 **90%**，以下功能可以立即测试：

### 1. 端到端测试流程

```bash
# 步骤 1: 启动 server-go
cd server-go
build.bat
dy-live-monitor.exe

# 步骤 2: 加载浏览器插件
# 在 Chrome 中访问 chrome://extensions/
# 启用开发者模式，加载 browser-monitor 目录

# 步骤 3: 打开抖音直播间
# 访问 https://live.douyin.com/[任意房间号]

# 步骤 4: 查看数据
# 在 server-go 主界面查看实时数据
# 礼物、消息会自动解析并存入数据库
```

---

## 🔧 剩余 10% 待实现功能

### 优先级 1 (重要但非阻塞)

#### 1.1 WebView2 Fallback 数据通道
**目的**: 当插件失效时，作为备用数据源

**实现思路**:
```go
// server-go/internal/fallback/webview.go
package fallback

import (
	webview "github.com/webview/webview_go"
)

// FallbackManager Fallback 管理器
type FallbackManager struct {
	webview webview.WebView
	roomID  string
}

// Start 启动 Fallback WebView2 实例
func (f *FallbackManager) Start(roomID string) error {
	f.roomID = roomID
	f.webview = webview.New(false)
	
	// 设置隐藏窗口
	f.webview.SetSize(1, 1, webview.HintNone)
	
	// 注入 JavaScript 拦截 WSS
	f.webview.Init(fmt.Sprintf(`
		// 拦截 WebSocket
		const OriginalWebSocket = window.WebSocket;
		window.WebSocket = function(url, protocols) {
			const ws = new OriginalWebSocket(url, protocols);
			ws.addEventListener('message', (event) => {
				// 发送到 server-go
				window.sendToGo(event.data);
			});
			return ws;
		};
	`))
	
	// 加载直播间
	f.webview.Navigate(fmt.Sprintf("https://live.douyin.com/%s", roomID))
	f.webview.Run()
	
	return nil
}
```

**触发条件**:
- 检测到插件心跳超时（10 秒无消息）
- 且浏览器已打开直播间（通过 CDP 检测）

---

#### 1.2 分段记分功能
**目的**: 支持一场直播多时段统计（如 PK 时段）

**数据库设计**:
```sql
-- server-go/internal/database/database.go
CREATE TABLE IF NOT EXISTS score_segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    room_id TEXT NOT NULL,
    segment_name TEXT NOT NULL,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    total_gift_value INTEGER DEFAULT 0,
    total_messages INTEGER DEFAULT 0,
    FOREIGN KEY (session_id) REFERENCES live_sessions(id)
);
```

**UI 实现**:
```html
<!-- server-go/internal/ui/webview.go -->
<button class="btn" onclick="createSegment()">📊 创建新分段</button>

<script>
async function createSegment() {
    const name = prompt('请输入分段名称（如：PK 第一轮）');
    if (!name) return;
    
    const result = await window.createSegment(currentRoom, name);
    if (result) {
        alert('分段创建成功！');
        loadSegments(currentRoom);
    }
}
</script>
```

---

### 优先级 2 (可选增强)

#### 2.1 浏览器插件打包与内嵌

**打包脚本**:
```bash
# browser-monitor/pack.sh
#!/bin/bash
zip -r browser-monitor.zip manifest.json background.js popup.html popup.js icons/
mv browser-monitor.zip ../server-go/assets/
```

**server-go 设置界面**:
```go
// server-go/internal/ui/settings.go
func (s *SettingsWindow) InstallPlugin() {
	// 解压插件到临时目录
	tempDir := filepath.Join(os.TempDir(), "browser-monitor")
	os.MkdirAll(tempDir, 0755)
	
	// 解压内嵌的 browser-monitor.zip
	unzip("assets/browser-monitor.zip", tempDir)
	
	// 打开浏览器扩展页面
	exec.Command("cmd", "/c", "start", "chrome://extensions/").Run()
	
	// 提示用户加载目录
	dialog.Message("请在浏览器中加载目录：%s", tempDir).Info()
}
```

---

#### 2.2 server-active 管理后台

**简单 HTML 管理界面**:
```html
<!-- server-active/web/admin.html -->
<!DOCTYPE html>
<html>
<head>
    <title>许可证管理后台</title>
</head>
<body>
    <h1>许可证生成</h1>
    <form id="generateForm">
        <label>客户ID: <input name="customer_id" required></label><br>
        <label>有效天数: <input name="expiry_days" type="number" value="365"></label><br>
        <button type="submit">生成许可证</button>
    </form>
    
    <h2>许可证列表</h2>
    <table id="licenseTable">
        <thead>
            <tr><th>许可证Key</th><th>客户ID</th><th>过期时间</th><th>状态</th></tr>
        </thead>
        <tbody></tbody>
    </table>
    
    <script>
        document.getElementById('generateForm').onsubmit = async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);
            data.software_id = 'dy-live-monitor';
            data.expiry_days = parseInt(data.expiry_days);
            data.max_activations = 1;
            data.license_type = 'full';
            
            const response = await fetch('/api/v1/licenses/generate', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify(data)
            });
            
            const result = await response.json();
            alert('许可证生成成功！\n\n' + result.license_data);
            loadLicenses();
        };
        
        async function loadLicenses() {
            // 实现许可证列表加载
        }
    </script>
</body>
</html>
```

**集成到 Gin**:
```go
// server-active/main.go
router.StaticFile("/admin", "./web/admin.html")
```

---

## 📚 测试场景建议

### 场景 1: 基本功能测试
1. 启动 `server-go`
2. 加载插件
3. 打开抖音直播间
4. **验证**: 
   - 主界面是否显示房间信息
   - 礼物消息是否正确解析
   - 数据库是否正确存储

### 场景 2: 离线缓存测试
1. 启动插件，打开直播间
2. **关闭** `server-go`
3. 等待 1 分钟（模拟离线）
4. **重新启动** `server-go`
5. **验证**: 插件是否自动重推缓存的数据

### 场景 3: 主播管理测试
1. 在主界面添加主播（如：主播A）
2. 绑定礼物（如：玫瑰花,火箭）
3. 在直播间送"玫瑰花"
4. **验证**: `anchor_performance` 表是否记录了主播A的业绩

### 场景 4: 多房间测试
1. 打开浏览器的两个标签页
2. 分别访问两个不同的直播间
3. **验证**: 主界面是否自动创建两个 Tab

### 场景 5: 许可证测试
1. 启动 `server-active`
2. 生成许可证（使用 curl 或 Postman）
3. 将 `license_data` 粘贴到 `server-go` 激活窗口
4. **验证**: `server-go` 是否启动成功

---

## 🔍 已知 Bug 排查清单

在开始测试前，请确认：

- [ ] **Go 版本** >= 1.21
- [ ] **Windows 10/11** 已安装 WebView2 Runtime
- [ ] **SQLite 驱动** 已正确编译（需要 CGO）
- [ ] **MySQL** 已启动（如果测试 server-active）
- [ ] **Chrome/Edge** 版本 >= 88

**常见问题**:

1. **CGO 编译错误**（SQLite）
   ```bash
   # 安装 MinGW-w64
   choco install mingw
   # 或下载：https://sourceforge.net/projects/mingw-w64/
   ```

2. **WebView2 未安装**
   ```bash
   # 下载：https://developer.microsoft.com/en-us/microsoft-edge/webview2/
   ```

3. **端口冲突**
   ```bash
   # 检查端口占用
   netstat -ano | findstr "8090"
   # 修改 config.json 端口
   ```

---

## 📝 待补充的文档

1. **API 文档**: 完整的 `server-active` RESTful API 文档（Swagger/OpenAPI）
2. **部署文档**: 生产环境部署指南（HTTPS、MySQL 优化、防火墙配置）
3. **用户手册**: 面向最终用户的使用手册（带截图）
4. **开发者文档**: 如何扩展 Protobuf 解析器（适配新消息类型）

---

## 🎉 项目里程碑

- [x] **v1.0** - Node.js 原型（CDP 基础监控）
- [x] **v2.0** - Protobuf 解析器（Douyin 消息）
- [x] **v3.0** - Go 重构（90% 完成）
- [ ] **v3.1** - Fallback + 分段记分（预计 +5%）
- [ ] **v3.5** - 管理后台 UI（预计 +3%）
- [ ] **v4.0** - 多平台支持（B站、快手）

---

**最后更新**: 2025-11-15  
**当前版本**: v3.0.0  
**进度**: 🟢 90% 完成
