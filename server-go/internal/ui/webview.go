package ui

import (
	"dy-live-monitor/internal/database"
	"dy-live-monitor/internal/server"
	"encoding/json"
	"fmt"
	"log"

	webview "github.com/webview/webview_go"
)

// MainWindow 主窗口
type MainWindow struct {
	webview webview.WebView
	db      *database.DB
	wsServer *server.WebSocketServer
}

// NewMainWindow 创建主窗口
func NewMainWindow(db *database.DB, wsServer *server.WebSocketServer) *MainWindow {
	return &MainWindow{
		db:       db,
		wsServer: wsServer,
	}
}

// Show 显示主窗口
func (w *MainWindow) Show() {
	// 创建 webview
	w.webview = webview.New(false)
	defer w.webview.Destroy()

	w.webview.SetTitle("抖音直播监控系统")
	w.webview.SetSize(1280, 800, webview.HintNone)

	// 绑定 Go 函数供 JavaScript 调用
	w.bindFunctions()

	// 加载 HTML 内容
	html := w.generateHTML()
	w.webview.SetHtml(html)

	// 运行主循环
	w.webview.Run()
}

// bindFunctions 绑定 Go 函数
func (w *MainWindow) bindFunctions() {
	// 获取房间列表
	w.webview.Bind("getRooms", func() string {
		rooms, err := w.getRooms()
		if err != nil {
			log.Printf("❌ 获取房间列表失败: %v", err)
			return "[]"
		}
		data, _ := json.Marshal(rooms)
		return string(data)
	})

	// 获取房间详情
	w.webview.Bind("getRoomDetails", func(roomID string) string {
		details, err := w.getRoomDetails(roomID)
		if err != nil {
			log.Printf("❌ 获取房间详情失败: %v", err)
			return "{}"
		}
		data, _ := json.Marshal(details)
		return string(data)
	})

	// 获取礼物记录
	w.webview.Bind("getGiftRecords", func(roomID string, limit int) string {
		records, err := w.getGiftRecords(roomID, limit)
		if err != nil {
			log.Printf("❌ 获取礼物记录失败: %v", err)
			return "[]"
		}
		data, _ := json.Marshal(records)
		return string(data)
	})

	// 获取消息记录
	w.webview.Bind("getMessageRecords", func(roomID string, limit int) string {
		records, err := w.getMessageRecords(roomID, limit)
		if err != nil {
			log.Printf("❌ 获取消息记录失败: %v", err)
			return "[]"
		}
		data, _ := json.Marshal(records)
		return string(data)
	})

	// 获取主播列表
	w.webview.Bind("getAnchors", func() string {
		anchors, err := w.getAnchors()
		if err != nil {
			log.Printf("❌ 获取主播列表失败: %v", err)
			return "[]"
		}
		data, _ := json.Marshal(anchors)
		return string(data)
	})

	// 保存主播
	w.webview.Bind("saveAnchor", func(anchorJSON string) string {
		var anchor map[string]interface{}
		if err := json.Unmarshal([]byte(anchorJSON), &anchor); err != nil {
			return `{"success": false, "message": "JSON解析失败"}`
		}

		if err := w.saveAnchor(anchor); err != nil {
			return fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error())
		}
		return `{"success": true}`
	})

	// 删除主播
	w.webview.Bind("deleteAnchor", func(anchorID string) string {
		if err := w.deleteAnchor(anchorID); err != nil {
			return fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error())
		}
		return `{"success": true}`
	})

	// 创建分段
	w.webview.Bind("createSegment", func(roomID, segmentName string) string {
		result, err := w.createSegment(roomID, segmentName)
		if err != nil {
			return fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error())
		}
		data, _ := json.Marshal(result)
		return string(data)
	})

	// 结束分段
	w.webview.Bind("endSegment", func(segmentID string) string {
		if err := w.endSegment(segmentID); err != nil {
			return fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error())
		}
		return `{"success": true}`
	})

	// 获取分段列表
	w.webview.Bind("getSegments", func(roomID string) string {
		segments, err := w.getSegments(roomID)
		if err != nil {
			log.Printf("❌ 获取分段列表失败: %v", err)
			return "[]"
		}
		data, _ := json.Marshal(segments)
		return string(data)
	})
}

// 数据库查询函数

func (w *MainWindow) getRooms() ([]map[string]interface{}, error) {
	rows, err := w.db.GetConnection().Query(`
		SELECT room_id, room_title, anchor_name, last_seen_at
		FROM rooms
		ORDER BY last_seen_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []map[string]interface{}
	for rows.Next() {
		var roomID, roomTitle, anchorName, lastSeenAt string
		if err := rows.Scan(&roomID, &roomTitle, &anchorName, &lastSeenAt); err != nil {
			continue
		}
		rooms = append(rooms, map[string]interface{}{
			"room_id":      roomID,
			"room_title":   roomTitle,
			"anchor_name":  anchorName,
			"last_seen_at": lastSeenAt,
		})
	}
	return rooms, nil
}

func (w *MainWindow) getRoomDetails(roomID string) (map[string]interface{}, error) {
	// 获取当前场次统计
	var totalGiftsValue, totalMessages int
	err := w.db.GetConnection().QueryRow(`
		SELECT COALESCE(SUM(gift_diamond_value), 0), COUNT(*)
		FROM gift_records
		WHERE room_id = ?
	`, roomID).Scan(&totalGiftsValue, &totalMessages)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"room_id":           roomID,
		"total_gifts_value": totalGiftsValue,
		"total_messages":    totalMessages,
	}, nil
}

func (w *MainWindow) getGiftRecords(roomID string, limit int) ([]map[string]interface{}, error) {
	rows, err := w.db.GetConnection().Query(`
		SELECT timestamp, user_nickname, gift_name, gift_count, gift_diamond_value
		FROM gift_records
		WHERE room_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var timestamp, userNickname, giftName, giftCount string
		var diamondValue int
		if err := rows.Scan(&timestamp, &userNickname, &giftName, &giftCount, &diamondValue); err != nil {
			continue
		}
		records = append(records, map[string]interface{}{
			"timestamp":      timestamp,
			"user_nickname":  userNickname,
			"gift_name":      giftName,
			"gift_count":     giftCount,
			"diamond_value":  diamondValue,
		})
	}
	return records, nil
}

func (w *MainWindow) getMessageRecords(roomID string, limit int) ([]map[string]interface{}, error) {
	rows, err := w.db.GetConnection().Query(`
		SELECT timestamp, message_type, user_nickname, content
		FROM message_records
		WHERE room_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var timestamp, messageType, userNickname, content string
		if err := rows.Scan(&timestamp, &messageType, &userNickname, &content); err != nil {
			continue
		}
		records = append(records, map[string]interface{}{
			"timestamp":     timestamp,
			"message_type":  messageType,
			"user_nickname": userNickname,
			"content":       content,
		})
	}
	return records, nil
}

func (w *MainWindow) getAnchors() ([]map[string]interface{}, error) {
	rows, err := w.db.GetConnection().Query(`
		SELECT anchor_id, anchor_name, bound_gifts, created_at
		FROM anchors
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anchors []map[string]interface{}
	for rows.Next() {
		var anchorID, anchorName, boundGifts, createdAt string
		if err := rows.Scan(&anchorID, &anchorName, &boundGifts, &createdAt); err != nil {
			continue
		}
		anchors = append(anchors, map[string]interface{}{
			"anchor_id":   anchorID,
			"anchor_name": anchorName,
			"bound_gifts": boundGifts,
			"created_at":  createdAt,
		})
	}
	return anchors, nil
}

func (w *MainWindow) saveAnchor(anchor map[string]interface{}) error {
	anchorID, _ := anchor["anchor_id"].(string)
	anchorName, _ := anchor["anchor_name"].(string)
	boundGifts, _ := anchor["bound_gifts"].(string)

	_, err := w.db.GetConnection().Exec(`
		INSERT OR REPLACE INTO anchors (anchor_id, anchor_name, bound_gifts, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, anchorID, anchorName, boundGifts)

	return err
}

func (w *MainWindow) deleteAnchor(anchorID string) error {
	_, err := w.db.GetConnection().Exec(`
		DELETE FROM anchors WHERE anchor_id = ?
	`, anchorID)
	return err
}

func (w *MainWindow) createSegment(roomID, segmentName string) (map[string]interface{}, error) {
	// 获取当前房间的 session_id
	var sessionID int64
	err := w.db.GetConnection().QueryRow(`
		SELECT id FROM live_sessions WHERE room_id = ? ORDER BY created_at DESC LIMIT 1
	`, roomID).Scan(&sessionID)

	if err != nil {
		return nil, err
	}

	// 先结束当前活动分段（如果有）
	activeSegment, _ := w.db.GetActiveSegment(sessionID)
	if activeSegment != nil {
		w.db.EndSegment(activeSegment.ID)
	}

	// 创建新分段
	segmentID, err := w.db.CreateSegment(sessionID, roomID, segmentName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success":    true,
		"segment_id": segmentID,
	}, nil
}

func (w *MainWindow) endSegment(segmentIDStr string) error {
	var segmentID int64
	fmt.Sscanf(segmentIDStr, "%d", &segmentID)
	return w.db.EndSegment(segmentID)
}

func (w *MainWindow) getSegments(roomID string) ([]map[string]interface{}, error) {
	// 获取房间的 session_id
	var sessionID int64
	err := w.db.GetConnection().QueryRow(`
		SELECT id FROM live_sessions WHERE room_id = ? ORDER BY created_at DESC LIMIT 1
	`, roomID).Scan(&sessionID)

	if err != nil {
		return []map[string]interface{}{}, nil
	}

	segments, err := w.db.GetAllSegments(sessionID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for _, seg := range segments {
		endTime := ""
		if seg.EndTime != nil {
			endTime = seg.EndTime.Format("2006-01-02 15:04:05")
		} else {
			endTime = "进行中"
		}

		result = append(result, map[string]interface{}{
			"id":               seg.ID,
			"segment_name":     seg.SegmentName,
			"start_time":       seg.StartTime.Format("2006-01-02 15:04:05"),
			"end_time":         endTime,
			"total_gift_value": seg.TotalGiftValue,
			"total_messages":   seg.TotalMessages,
		})
	}

	return result, nil
}

// generateHTML 生成 HTML 页面
func (w *MainWindow) generateHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>抖音直播监控系统</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
            overflow: hidden;
        }
        .container {
            display: flex;
            height: 100vh;
        }
        .sidebar {
            width: 250px;
            background: #2c3e50;
            color: white;
            padding: 20px;
            overflow-y: auto;
        }
        .sidebar h2 {
            font-size: 18px;
            margin-bottom: 20px;
            color: #ecf0f1;
        }
        .room-item {
            padding: 12px;
            background: #34495e;
            border-radius: 8px;
            margin-bottom: 10px;
            cursor: pointer;
            transition: all 0.3s;
        }
        .room-item:hover {
            background: #3d566e;
            transform: translateX(5px);
        }
        .room-item.active {
            background: #667eea;
        }
        .room-item h3 {
            font-size: 14px;
            margin-bottom: 5px;
        }
        .room-item p {
            font-size: 12px;
            color: #bdc3c7;
        }
        .main-content {
            flex: 1;
            background: white;
            overflow-y: auto;
            padding: 30px;
        }
        .tabs {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            border-bottom: 2px solid #e0e0e0;
        }
        .tab {
            padding: 12px 24px;
            background: transparent;
            border: none;
            border-bottom: 3px solid transparent;
            cursor: pointer;
            font-size: 16px;
            font-weight: 500;
            color: #666;
            transition: all 0.3s;
        }
        .tab:hover {
            color: #667eea;
        }
        .tab.active {
            color: #667eea;
            border-bottom-color: #667eea;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
            animation: fadeIn 0.3s;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .stats-grid {
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
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .stat-card h3 {
            font-size: 14px;
            margin-bottom: 10px;
            opacity: 0.9;
        }
        .stat-card .value {
            font-size: 32px;
            font-weight: bold;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 10px rgba(0,0,0,0.05);
        }
        thead {
            background: #f8f9fa;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #e0e0e0;
        }
        th {
            font-weight: 600;
            color: #666;
            font-size: 14px;
        }
        td {
            font-size: 14px;
        }
        tbody tr:hover {
            background: #f8f9fa;
        }
        .btn {
            padding: 10px 20px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.3s;
        }
        .btn:hover {
            background: #5568d3;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        .btn-danger {
            background: #e74c3c;
        }
        .btn-danger:hover {
            background: #c0392b;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #999;
        }
        .empty-state h3 {
            font-size: 18px;
            margin-bottom: 10px;
        }
        .loading {
            text-align: center;
            padding: 40px;
            color: #999;
        }
        .anchor-form {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .form-group {
            margin-bottom: 15px;
        }
        .form-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: 500;
            color: #666;
        }
        .form-group input, .form-group textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-size: 14px;
        }
        .form-group textarea {
            min-height: 80px;
            resize: vertical;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="sidebar">
            <h2>🎬 监控房间</h2>
            <div id="roomList" class="loading">加载中...</div>
        </div>
        <div class="main-content">
            <div class="tabs">
                <button class="tab active" onclick="switchTab('overview')">📊 数据概览</button>
                <button class="tab" onclick="switchTab('gifts')">🎁 礼物记录</button>
                <button class="tab" onclick="switchTab('messages')">💬 消息记录</button>
                <button class="tab" onclick="switchTab('segments')">📈 分段记分</button>
                <button class="tab" onclick="switchTab('anchors')">👤 主播管理</button>
            </div>

            <div id="overview" class="tab-content active">
                <div class="stats-grid">
                    <div class="stat-card">
                        <h3>礼物总价值</h3>
                        <div class="value" id="totalGiftsValue">0 💎</div>
                    </div>
                    <div class="stat-card">
                        <h3>消息总数</h3>
                        <div class="value" id="totalMessages">0</div>
                    </div>
                </div>
                <div class="empty-state">
                    <h3>请选择左侧房间查看详情</h3>
                    <p>当浏览器打开抖音直播间后，这里会显示实时数据</p>
                </div>
            </div>

            <div id="gifts" class="tab-content">
                <table id="giftsTable">
                    <thead>
                        <tr>
                            <th>时间</th>
                            <th>用户</th>
                            <th>礼物</th>
                            <th>数量</th>
                            <th>价值(💎)</th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>

            <div id="messages" class="tab-content">
                <table id="messagesTable">
                    <thead>
                        <tr>
                            <th>时间</th>
                            <th>类型</th>
                            <th>用户</th>
                            <th>内容</th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>

            <div id="segments" class="tab-content">
                <div style="margin-bottom: 20px; display: flex; gap: 10px; align-items: center;">
                    <input type="text" id="segmentName" placeholder="输入分段名称（如：PK 第一轮）" style="flex: 1; padding: 10px; border: 1px solid #ddd; border-radius: 6px;">
                    <button class="btn" onclick="createNewSegment()">📊 创建新分段</button>
                    <button class="btn" onclick="endCurrentSegment()" style="background: #e74c3c;">⏹️ 结束当前分段</button>
                </div>
                
                <table id="segmentsTable">
                    <thead>
                        <tr>
                            <th>分段名称</th>
                            <th>开始时间</th>
                            <th>结束时间</th>
                            <th>礼物总值(💎)</th>
                            <th>消息数</th>
                            <th>状态</th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>

            <div id="anchors" class="tab-content">
                <div class="anchor-form">
                    <h3 style="margin-bottom: 15px;">添加/编辑主播</h3>
                    <div class="form-group">
                        <label>主播ID</label>
                        <input type="text" id="anchorId" placeholder="例如: anchor_001">
                    </div>
                    <div class="form-group">
                        <label>主播名称</label>
                        <input type="text" id="anchorName" placeholder="例如: 主播A">
                    </div>
                    <div class="form-group">
                        <label>绑定礼物（多个用逗号分隔）</label>
                        <textarea id="boundGifts" placeholder="例如: 玫瑰花,跑车,火箭"></textarea>
                    </div>
                    <button class="btn" onclick="saveAnchor()">保存主播</button>
                </div>
                <table id="anchorsTable">
                    <thead>
                        <tr>
                            <th>主播ID</th>
                            <th>主播名称</th>
                            <th>绑定礼物</th>
                            <th>创建时间</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody></tbody>
                </table>
            </div>
        </div>
    </div>

    <script>
        let currentRoom = null;

        // 切换标签页
        function switchTab(tabName) {
            document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
            
            event.target.classList.add('active');
            document.getElementById(tabName).classList.add('active');
            
            if (currentRoom) {
                if (tabName === 'gifts') loadGifts(currentRoom);
                if (tabName === 'messages') loadMessages(currentRoom);
                if (tabName === 'anchors') loadAnchors();
            }
        }

        // 加载房间列表
        async function loadRooms() {
            try {
                const rooms = JSON.parse(await getRooms());
                const roomList = document.getElementById('roomList');
                
                if (rooms.length === 0) {
                    roomList.innerHTML = '<div class="empty-state"><p>暂无房间</p></div>';
                    return;
                }
                
                roomList.innerHTML = rooms.map(room => `
                    <div class="room-item" onclick="selectRoom('${room.room_id}')">
                        <h3>${room.room_title || '直播间 ' + room.room_id}</h3>
                        <p>主播: ${room.anchor_name || '未知'}</p>
                        <p style="font-size: 11px;">${room.last_seen_at}</p>
                    </div>
                `).join('');
            } catch (e) {
                console.error('加载房间列表失败:', e);
            }
        }

        // 选择房间
        async function selectRoom(roomId) {
            currentRoom = roomId;
            document.querySelectorAll('.room-item').forEach(item => item.classList.remove('active'));
            event.target.closest('.room-item').classList.add('active');
            
            // 加载房间详情
            try {
                const details = JSON.parse(await getRoomDetails(roomId));
                document.getElementById('totalGiftsValue').textContent = details.total_gifts_value + ' 💎';
                document.getElementById('totalMessages').textContent = details.total_messages;
            } catch (e) {
                console.error('加载房间详情失败:', e);
            }
            
            // 加载当前标签页数据
            const activeTab = document.querySelector('.tab.active').textContent;
            if (activeTab.includes('礼物')) loadGifts(roomId);
            if (activeTab.includes('消息')) loadMessages(roomId);
        }

        // 加载礼物记录
        async function loadGifts(roomId) {
            try {
                const gifts = JSON.parse(await getGiftRecords(roomId, 100));
                const tbody = document.querySelector('#giftsTable tbody');
                
                if (gifts.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;">暂无礼物记录</td></tr>';
                    return;
                }
                
                tbody.innerHTML = gifts.map(gift => `
                    <tr>
                        <td>${new Date(gift.timestamp).toLocaleString('zh-CN')}</td>
                        <td>${gift.user_nickname}</td>
                        <td>${gift.gift_name}</td>
                        <td>${gift.gift_count}</td>
                        <td>${gift.diamond_value}</td>
                    </tr>
                `).join('');
            } catch (e) {
                console.error('加载礼物记录失败:', e);
            }
        }

        // 加载消息记录
        async function loadMessages(roomId) {
            try {
                const messages = JSON.parse(await getMessageRecords(roomId, 100));
                const tbody = document.querySelector('#messagesTable tbody');
                
                if (messages.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="4" style="text-align:center;">暂无消息记录</td></tr>';
                    return;
                }
                
                tbody.innerHTML = messages.map(msg => `
                    <tr>
                        <td>${new Date(msg.timestamp).toLocaleString('zh-CN')}</td>
                        <td>${msg.message_type}</td>
                        <td>${msg.user_nickname}</td>
                        <td>${msg.content || '-'}</td>
                    </tr>
                `).join('');
            } catch (e) {
                console.error('加载消息记录失败:', e);
            }
        }

        // 加载主播列表
        async function loadAnchors() {
            try {
                const anchors = JSON.parse(await getAnchors());
                const tbody = document.querySelector('#anchorsTable tbody');
                
                if (anchors.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;">暂无主播</td></tr>';
                    return;
                }
                
                tbody.innerHTML = anchors.map(anchor => `
                    <tr>
                        <td>${anchor.anchor_id}</td>
                        <td>${anchor.anchor_name}</td>
                        <td>${anchor.bound_gifts || '-'}</td>
                        <td>${new Date(anchor.created_at).toLocaleString('zh-CN')}</td>
                        <td>
                            <button class="btn btn-danger" onclick="deleteAnchor('${anchor.anchor_id}')">删除</button>
                        </td>
                    </tr>
                `).join('');
            } catch (e) {
                console.error('加载主播列表失败:', e);
            }
        }

        // 保存主播
        async function saveAnchor() {
            const anchor = {
                anchor_id: document.getElementById('anchorId').value,
                anchor_name: document.getElementById('anchorName').value,
                bound_gifts: document.getElementById('boundGifts').value
            };
            
            if (!anchor.anchor_id || !anchor.anchor_name) {
                alert('请填写主播ID和名称');
                return;
            }
            
            try {
                const result = JSON.parse(await window.saveAnchor(JSON.stringify(anchor)));
                if (result.success) {
                    alert('保存成功！');
                    document.getElementById('anchorId').value = '';
                    document.getElementById('anchorName').value = '';
                    document.getElementById('boundGifts').value = '';
                    loadAnchors();
                } else {
                    alert('保存失败: ' + result.message);
                }
            } catch (e) {
                alert('保存失败: ' + e.message);
            }
        }

        // 删除主播
        async function deleteAnchor(anchorId) {
            if (!confirm('确定要删除这个主播吗？')) return;
            
            try {
                const result = JSON.parse(await window.deleteAnchor(anchorId));
                if (result.success) {
                    alert('删除成功！');
                    loadAnchors();
                } else {
                    alert('删除失败: ' + result.message);
                }
            } catch (e) {
                alert('删除失败: ' + e.message);
            }
        }

        // 加载分段列表
        async function loadSegments(roomId) {
            try {
                const segments = JSON.parse(await getSegments(roomId));
                const tbody = document.querySelector('#segmentsTable tbody');
                
                if (segments.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;">暂无分段记录</td></tr>';
                    return;
                }
                
                tbody.innerHTML = segments.map(seg => `
                    <tr style="${seg.end_time === '进行中' ? 'background: #fff3cd;' : ''}">
                        <td><strong>${seg.segment_name}</strong></td>
                        <td>${seg.start_time}</td>
                        <td>${seg.end_time}</td>
                        <td>${seg.total_gift_value}</td>
                        <td>${seg.total_messages}</td>
                        <td>${seg.end_time === '进行中' ? '<span style="color: #28a745;">⏺️ 进行中</span>' : '<span style="color: #6c757d;">⏹️ 已结束</span>'}</td>
                    </tr>
                `).join('');
            } catch (e) {
                console.error('加载分段列表失败:', e);
            }
        }

        // 创建新分段
        async function createNewSegment() {
            if (!currentRoom) {
                alert('请先选择一个房间');
                return;
            }

            const segmentName = document.getElementById('segmentName').value;
            if (!segmentName) {
                alert('请输入分段名称');
                return;
            }
            
            try {
                const result = JSON.parse(await createSegment(currentRoom, segmentName));
                if (result.success) {
                    alert('分段创建成功！');
                    document.getElementById('segmentName').value = '';
                    loadSegments(currentRoom);
                } else {
                    alert('创建失败: ' + result.message);
                }
            } catch (e) {
                alert('创建失败: ' + e.message);
            }
        }

        // 结束当前分段
        async function endCurrentSegment() {
            if (!currentRoom) {
                alert('请先选择一个房间');
                return;
            }

            if (!confirm('确定要结束当前分段吗？')) return;
            
            try {
                // 获取当前活动分段
                const segments = JSON.parse(await getSegments(currentRoom));
                const activeSegment = segments.find(s => s.end_time === '进行中');
                
                if (!activeSegment) {
                    alert('当前没有进行中的分段');
                    return;
                }

                const result = JSON.parse(await endSegment(String(activeSegment.id)));
                if (result.success) {
                    alert('分段已结束！');
                    loadSegments(currentRoom);
                } else {
                    alert('结束失败: ' + result.message);
                }
            } catch (e) {
                alert('结束失败: ' + e.message);
            }
        }

        // 修改 switchTab 函数
        function switchTab(tabName) {
            document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
            
            event.target.classList.add('active');
            document.getElementById(tabName).classList.add('active');
            
            if (currentRoom) {
                if (tabName === 'gifts') loadGifts(currentRoom);
                if (tabName === 'messages') loadMessages(currentRoom);
                if (tabName === 'segments') loadSegments(currentRoom);
                if (tabName === 'anchors') loadAnchors();
            }
        }

        // 初始化
        loadRooms();
        setInterval(loadRooms, 5000); // 每5秒刷新房间列表
    </script>
</body>
</html>`
}
