package server

import (
	"dy-live-monitor/internal/database"
	"dy-live-monitor/internal/parser"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

// UIUpdater UI更新接口
type UIUpdater interface {
	AddOrUpdateRoom(roomID string)
	AddParsedMessage(roomID string, message string)
	AddParsedMessageWithDetail(roomID string, message string, detail map[string]interface{})
}

// WebSocketServer WebSocket服务器
type WebSocketServer struct {
	port          int
	db            *database.DB
	giftAllocator *GiftAllocator
	clients       map[*websocket.Conn]bool
	clientsMu     sync.RWMutex
	rooms         map[string]*RoomManager
	roomsMu       sync.RWMutex
	upgrader      websocket.Upgrader
	started       chan bool // 用于通知服务器已启动
	uiUpdater     UIUpdater // UI更新器
}

// RoomManager 房间管理器
type RoomManager struct {
	RoomID      string
	LastMessage int64
}

// NewWebSocketServer 创建WebSocket服务器
func NewWebSocketServer(port int, db *database.DB) *WebSocketServer {
	return &WebSocketServer{
		port:          port,
		db:            db,
		giftAllocator: NewGiftAllocator(db.GetConnection()),
		clients:       make(map[*websocket.Conn]bool),
		rooms:         make(map[string]*RoomManager),
		started:       make(chan bool, 1),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源（生产环境应限制）
			},
		},
	}
}

// Start 启动WebSocket服务器
func (s *WebSocketServer) Start() error {
	http.HandleFunc("/monitor", s.handleWebSocket)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)

	// 在单独的 goroutine 中启动服务器
	go func() {
		log.Printf("🌐 WebSocket 服务器正在启动，监听端口: %d", s.port)
		log.Printf("📍 WebSocket 地址: ws://localhost:%d/monitor", s.port)
		log.Printf("📍 健康检查地址: http://localhost:%d/health", s.port)

		// 通知服务器已准备好监听
		s.started <- true

		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("❌ WebSocket 服务器启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	<-s.started
	return nil
}

// handleWebSocket 处理WebSocket连接
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔌 收到 WebSocket 连接请求: %s", r.RemoteAddr)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket 升级失败: %v", err)
		return
	}

	log.Printf("✅ WebSocket 连接成功: %s", conn.RemoteAddr())

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
		log.Printf("👋 客户端断开: %s", conn.RemoteAddr())
	}()

	// 读取消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("⚠️  WebSocket 错误: %v", err)
			}
			break
		}

		// 处理消息
		s.handleMessage(message)
	}
}

// handleMessage 处理接收到的消息
func (s *WebSocketServer) handleMessage(message []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		log.Printf("❌ JSON 解析失败: %v", err)
		return
	}

	msgType, ok := data["type"].(string)
	if !ok {
		log.Println("⚠️  消息缺少 type 字段")
		return
	}

	// 根据消息类型处理
	switch msgType {
	case "websocket_frame_received", "websocket_frame_sent":
		s.handleDouyinMessage(data)
	case "cdp_request":
		s.handleRequest(data)
	case "heartbeat":
		// 心跳消息，用于检测插件存活
		log.Println("💓 收到心跳")
	default:
		log.Printf("⚠️  未知消息类型: %s", msgType)
	}
}

// SetUIUpdater 设置UI更新器
func (s *WebSocketServer) SetUIUpdater(updater UIUpdater) {
	s.uiUpdater = updater
}

// handleDouyinMessage 处理抖音消息
func (s *WebSocketServer) handleDouyinMessage(data map[string]interface{}) {
	url, _ := data["url"].(string)
	payloadData, _ := data["payloadData"].(string)

	if url == "" || payloadData == "" {
		return
	}

	// 打印 WSS 链接地址
	log.Printf("🔗 WSS 链接: %s", url)

	// 提取房间号
	roomID := extractRoomID(url)
	if roomID == "" {
		log.Printf("⚠️  无法从 URL 提取房间号: %s", url)
		return
	}

	log.Printf("📍 提取到房间号: %s", roomID)

	// 确保 rooms 表中有记录
	if err := s.ensureRoomRecord(roomID, url); err != nil {
		log.Printf("⚠️  确保房间记录失败 (房间 %s): %v", roomID, err)
	}

	// 获取或创建房间管理器
	_ = s.getOrCreateRoom(roomID)

	// 通知UI创建房间Tab
	if s.uiUpdater != nil {
		s.uiUpdater.AddOrUpdateRoom(roomID)
	}

	// 解析抖音消息
	parsedMessages, err := parser.ParseWebcastPayload(payloadData)
	if err != nil {
		log.Printf("❌ [房间 %s] 解析失败: %v", roomID, err)
		if s.uiUpdater != nil {
			s.uiUpdater.AddParsedMessage(roomID, fmt.Sprintf("❌ 解析失败: %v", err))
		}
		return
	}

	if len(parsedMessages) == 0 {
		log.Printf("ℹ️  [房间 %s] 解析结果为空", roomID)
		return
	}

	log.Printf("✅ [房间 %s] 成功解析 %d 条消息", roomID, len(parsedMessages))

	// 存储到数据库
	for i, msg := range parsedMessages {
		log.Printf("📝 [房间 %s] 处理消息 %d/%d: %s - %s", roomID, i+1, len(parsedMessages), msg.MessageType, msg.Method)

		s.saveMessage(roomID, msg)

		if s.uiUpdater != nil {
			detailCopy := cloneDetail(msg.Detail)
			detailCopy["_parsed"] = msg
			s.uiUpdater.AddParsedMessageWithDetail(roomID, msg.Display, detailCopy)
		}

		if err := s.PersistRoomMessage(roomID, msg, "browser"); err != nil {
			log.Printf("⚠️  [房间 %s] 保存房间消息失败: %v", roomID, err)
		} else {
			log.Printf("✅ [房间 %s] 房间消息已保存", roomID)
		}
	}

	log.Printf("📨 [房间 %s] 批量处理完成，共 %d 条消息", roomID, len(parsedMessages))
}

// handleRequest 处理HTTP请求记录
func (s *WebSocketServer) handleRequest(data map[string]interface{}) {
	// 可选：记录所有HTTP请求
	url, _ := data["url"].(string)
	log.Printf("🌐 请求: %s", url)
}

// handleHealth 健康检查接口
func (s *WebSocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	s.clientsMu.RLock()
	clientCount := len(s.clients)
	s.clientsMu.RUnlock()

	s.roomsMu.RLock()
	roomCount := len(s.rooms)
	s.roomsMu.RUnlock()

	response := map[string]interface{}{
		"status":  "ok",
		"port":    s.port,
		"clients": clientCount,
		"rooms":   roomCount,
		"endpoints": map[string]string{
			"websocket": fmt.Sprintf("ws://localhost:%d/monitor", s.port),
			"health":    fmt.Sprintf("http://localhost:%d/health", s.port),
		},
	}

	json.NewEncoder(w).Encode(response)

	log.Printf("💊 健康检查: 客户端=%d, 房间=%d", clientCount, roomCount)
}

// getOrCreateRoom 获取或创建房间管理器
func (s *WebSocketServer) getOrCreateRoom(roomID string) *RoomManager {
	s.roomsMu.Lock()
	defer s.roomsMu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		// 创建新房间
		room = &RoomManager{
			RoomID: roomID,
		}

		// 确保 rooms 表中有记录
		if err := s.ensureRoomRecord(roomID, ""); err != nil {
			log.Printf("⚠️  确保房间记录失败: %v", err)
		}

		s.rooms[roomID] = room
		log.Printf("🎬 创建新房间: %s", roomID)
	}

	return room
}

// saveMessage 保存消息到数据库
func (s *WebSocketServer) saveMessage(roomID string, parsed *parser.ParsedProtoMessage) {
	if parsed == nil {
		log.Printf("⚠️  [房间 %s] parsed 消息为 nil，跳过保存", roomID)
		return
	}

	log.Printf("🔍 [房间 %s] saveMessage 检查消息类型: '%s'", roomID, parsed.MessageType)

	switch parsed.MessageType {
	case "礼物消息":
		log.Printf("✅ [房间 %s] 识别到礼物消息，准备保存到 gift_records", roomID)
		s.saveGiftRecord(roomID, parsed)
	default:
		log.Printf("ℹ️  [房间 %s] 消息类型 '%s' 不需要特殊处理", roomID, parsed.MessageType)
	}
}

func (s *WebSocketServer) PersistRoomMessage(roomID string, parsed *parser.ParsedProtoMessage, source string) error {
	if s.db == nil || parsed == nil {
		return nil
	}

	detail := parsed.Detail

	// 生成 msgId

	var record = &database.RoomMessageRecord{
		MsgID:       gjson.Get(parsed.RawJSON, "common.msgId").String(),
		RoomID:      roomID,
		Method:      parsed.Method,
		MessageType: parsed.MessageType,
		Display:     parsed.Display,
		UserID:      toString(detail["userId"]),
		UserName:    toString(detail["user"]),
		AnchorID:    toString(detail["anchorId"]),
		RawPayload:  parsed.RawPayload,
		ParsedJSON:  parsed.RawJSON,
		Source:      source,
		SentAt:      parsed.ReceivedAt,
	}
	if record.SentAt.IsZero() {
		record.SentAt = time.Now()
	}

	return s.db.InsertRoomMessage(record)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int32:
		return int(val)
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		if val == "" {
			return 0
		}
		var i int
		fmt.Sscanf(val, "%d", &i)
		return i
	default:
		return 0
	}
}

func cloneDetail(detail map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(detail))
	for k, v := range detail {
		result[k] = v
	}
	return result
}

// saveGiftRecord 保存礼物记录
func (s *WebSocketServer) saveGiftRecord(roomID string, parsed *parser.ParsedProtoMessage) {
	log.Printf("🎁 [房间 %s] 开始处理礼物记录", roomID)

	detail := parsed.Detail
	if detail == nil {
		log.Printf("❌ [房间 %s] 礼物消息 Detail 为空", roomID)
		return
	}

	// 生成 msgId
	msgID := gjson.Get(parsed.RawJSON, "common.msgId")

	userID := toString(detail["userId"])
	userNickname := toString(detail["user"])
	giftID := toString(detail["giftId"])
	giftName := toString(detail["giftName"])
	giftCount := toInt(detail["groupCount"])
	if giftCount == 0 {
		giftCount = 1
	}
	diamondCount := toInt(detail["diamondCount"])
	content := toString(detail["content"])
	anchorID := sanitizeAnchorID(toString(detail["anchorId"]))
	anchorName := sanitizeAnchorName(toString(detail["anchorName"]))

	log.Printf("🎁 [房间 %s] 礼物详情 - 用户: %s(%s), 礼物: %s(%s) x%d, 钻石: %d",
		roomID, userNickname, userID, giftName, giftID, giftCount, diamondCount)

	// 尝试分配礼物给主播
	if anchorID == "" {
		log.Printf("🔍 [房间 %s] 礼物未指定主播，尝试自动分配", roomID)
		var err error
		anchorID, err = s.giftAllocator.AllocateGift(giftName, content)
		if err == nil && anchorID != "" {
			log.Printf("🎯 [房间 %s] 礼物 %s 自动分配给主播: %s", roomID, giftName, anchorID)
			// 查询主播名称
			var name string
			err := s.db.GetConnection().QueryRow(`SELECT anchor_name FROM anchors WHERE anchor_id = ?`, anchorID).Scan(&name)
			if err == nil {
				anchorName = name
				log.Printf("📛 [房间 %s] 主播名称: %s", roomID, anchorName)
			}
		} else if err != nil {
			log.Printf("⚠️  [房间 %s] 自动分配主播失败: %v", roomID, err)
		}
	} else {
		log.Printf("✅ [房间 %s] 礼物已指定主播: %s (%s)", roomID, anchorName, anchorID)
	}

	log.Printf("💾 [房间 %s] 准备插入 gift_records 表，msgID: %s", roomID, msgID)

	result, err := s.db.GetConnection().Exec(`
		INSERT INTO gift_records (
			msg_id, room_id, user_id, user_nickname, gift_id, gift_name, 
			gift_count, gift_diamond_value, anchor_id, anchor_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msgID, roomID, userID, userNickname, giftID, giftName, giftCount, diamondCount, anchorID, anchorName)

	if err != nil {
		log.Printf("❌ [房间 %s] 保存礼物记录失败: %v", roomID, err)
		log.Printf("❌ [房间 %s] 失败的数据: msgID=%s, userNickname=%s, giftName=%s",
			roomID, msgID, userNickname, giftName)
		return
	}

	recordID, _ := result.LastInsertId()
	log.Printf("✅ [房间 %s] 礼物记录已保存到 gift_records 表，recordID: %d, msgID: %s", roomID, recordID, msgID)

	if anchorID != "" {
		if anchorName == "" {
			anchorName = s.lookupAnchorName(anchorID)
		}

		s.ensureGlobalAnchor(anchorID, anchorName)
		s.ensureRoomAnchorRecord(roomID, anchorID, anchorName)

		// 记录主播业绩
		if err := s.giftAllocator.RecordAnchorPerformance(anchorID, giftName, diamondCount); err != nil {
			log.Printf("❌ [房间 %s] 记录主播业绩失败: %v", roomID, err)
		} else {
			log.Printf("📊 [房间 %s] 主播 %s 业绩已更新", roomID, anchorID)
		}
	}
}

func (s *WebSocketServer) ensureRoomAnchorRecord(roomID, anchorID, anchorName string) {
	anchorID = sanitizeAnchorID(anchorID)
	if s.db == nil || anchorID == "" {
		return
	}
	anchorName = sanitizeAnchorName(anchorName)
	if anchorName == "" {
		anchorName = s.lookupAnchorName(anchorID)
	}
	_, err := s.db.GetConnection().Exec(`
		INSERT INTO room_anchors (room_id, anchor_id, anchor_name, gift_count, score)
		VALUES (?, ?, ?, 0, 0)
		ON CONFLICT(room_id, anchor_id) DO UPDATE SET anchor_name=excluded.anchor_name
	`, roomID, anchorID, anchorName)
	if err != nil {
		log.Printf("⚠️  [房间 %s] 同步 room_anchors 失败: %v", roomID, err)
	}
}

func (s *WebSocketServer) ensureGlobalAnchor(anchorID, anchorName string) {
	anchorID = sanitizeAnchorID(anchorID)
	if s.db == nil || anchorID == "" {
		return
	}
	anchorName = sanitizeAnchorName(anchorName)
	if anchorName == "" {
		anchorName = anchorID
	}
	_, err := s.db.GetConnection().Exec(`
		INSERT INTO anchors (anchor_id, anchor_name, bound_gifts, created_at, updated_at)
		VALUES (?, ?, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(anchor_id) DO UPDATE SET 
			anchor_name=CASE 
				WHEN excluded.anchor_name = '' THEN anchors.anchor_name
				ELSE excluded.anchor_name
			END,
			updated_at=CURRENT_TIMESTAMP
	`, anchorID, anchorName)
	if err != nil {
		log.Printf("⚠️  同步 anchors 失败: %v", err)
	}
}

func (s *WebSocketServer) lookupAnchorName(anchorID string) string {
	anchorID = sanitizeAnchorID(anchorID)
	if s.db == nil || anchorID == "" {
		return ""
	}
	var name string
	err := s.db.GetConnection().QueryRow(`SELECT anchor_name FROM anchors WHERE anchor_id = ?`, anchorID).Scan(&name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(name)
}

// ensureRoomRecord 确保 rooms 表中有房间记录
func (s *WebSocketServer) ensureRoomRecord(roomID, wsURL string) error {
	if s.db == nil || roomID == "" {
		return nil
	}

	// 检查是否已存在
	var count int
	err := s.db.GetConnection().QueryRow(`SELECT COUNT(*) FROM rooms WHERE room_id = ?`, roomID).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询房间记录失败: %w", err)
	}

	if count > 0 {
		// 已存在，更新 last_seen_at 及必要字段
		setClauses := []string{"last_seen_at = CURRENT_TIMESTAMP"}
		args := make([]interface{}, 0, 2)
		if strings.TrimSpace(wsURL) != "" {
			setClauses = append(setClauses, "ws_url = ?")
			args = append(args, wsURL)
		}
		args = append(args, roomID)
		query := fmt.Sprintf("UPDATE rooms SET %s WHERE room_id = ?", strings.Join(setClauses, ", "))
		_, err := s.db.GetConnection().Exec(query, args...)
		if err != nil {
			log.Printf("⚠️  [房间 %s] 更新 last_seen_at 失败: %v", roomID, err)
		} else {
			log.Printf("🔄 [房间 %s] 房间记录已更新", roomID)
		}
		return nil
	}

	// 不存在，插入新记录
	_, err = s.db.GetConnection().Exec(`
		INSERT INTO rooms (room_id, room_title, anchor_name, ws_url, first_seen_at, last_seen_at)
		VALUES (?, '', '', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roomID, wsURL)

	if err != nil {
		return fmt.Errorf("插入房间记录失败: %w", err)
	}

	log.Printf("✅ [房间 %s] 新房间记录已创建", roomID)
	return nil
}

func sanitizeAnchorID(val string) string {
	val = strings.TrimSpace(val)
	if val == "" || val == "<nil>" {
		return ""
	}
	return val
}

func sanitizeAnchorName(val string) string {
	val = strings.TrimSpace(val)
	if val == "<nil>" {
		return ""
	}
	return val
}

// extractRoomID 从URL中提取房间号
func extractRoomID(url string) string {
	// 从 URL 参数中提取 room_id 或 wss_push_room_id
	if idx := strings.Index(url, "room_id="); idx >= 0 {
		start := idx + 8
		end := strings.IndexAny(url[start:], "&")
		if end > 0 {
			return url[start : start+end]
		}
		return url[start:]
	}
	if idx := strings.Index(url, "wss_push_room_id="); idx >= 0 {
		start := idx + 17
		end := strings.IndexAny(url[start:], "&")
		if end > 0 {
			return url[start : start+end]
		}
		return url[start:]
	}
	return "unknown"
}
