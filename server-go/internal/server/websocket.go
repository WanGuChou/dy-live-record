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
	SessionID   int64
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

	// 提取房间号
	roomID := extractRoomID(url)
	if roomID == "" {
		return
	}

	// 获取或创建房间管理器
	room := s.getOrCreateRoom(roomID)

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
		return
	}

	// 存储到数据库
	for _, msg := range parsedMessages {
		s.saveMessage(roomID, room.SessionID, msg)

		if s.uiUpdater != nil {
			detailCopy := cloneDetail(msg.Detail)
			detailCopy["_parsed"] = msg
			s.uiUpdater.AddParsedMessageWithDetail(roomID, msg.Display, detailCopy)
		}

		if err := s.PersistRoomMessage(roomID, msg, "browser"); err != nil {
			log.Printf("⚠️  保存房间消息失败: %v", err)
		}
	}

	log.Printf("📨 房间 %s 收到 %d 条消息", roomID, len(parsedMessages))
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

		// 创建新的直播场次
		sessionID := s.createLiveSession(roomID)
		room.SessionID = sessionID

		s.rooms[roomID] = room
		log.Printf("🎬 创建新房间: %s (Session: %d)", roomID, sessionID)
	}

	return room
}

// createLiveSession 创建直播场次
func (s *WebSocketServer) createLiveSession(roomID string) int64 {
	result, err := s.db.GetConnection().Exec(
		"INSERT INTO live_sessions (room_id) VALUES (?)",
		roomID,
	)
	if err != nil {
		log.Printf("❌ 创建场次失败: %v", err)
		return 0
	}

	sessionID, _ := result.LastInsertId()
	return sessionID
}

// saveMessage 保存消息到数据库
func (s *WebSocketServer) saveMessage(roomID string, sessionID int64, parsed *parser.ParsedProtoMessage) {
	if parsed == nil {
		return
	}

	switch parsed.MessageType {
	case "礼物消息":
		s.saveGiftRecord(roomID, sessionID, parsed)
	case "聊天消息", "进入直播间", "关注消息":
		s.saveMessageRecord(roomID, sessionID, parsed)
	}
}

func (s *WebSocketServer) PersistRoomMessage(roomID string, parsed *parser.ParsedProtoMessage, source string) error {
	if s.db == nil || parsed == nil {
		return nil
	}

	detail := parsed.Detail

	record := &database.RoomMessageRecord{
		RoomID:      roomID,
		Method:      parsed.Method,
		MessageType: parsed.MessageType,
		Display:     parsed.Display,
		UserID:      toString(detail["userId"]),
		UserName:    toString(detail["user"]),
		GiftName:    toString(detail["giftName"]),
		GiftCount:   toInt(detail["groupCount"]),
		GiftValue:   toInt(detail["diamondCount"]),
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
func (s *WebSocketServer) saveGiftRecord(roomID string, sessionID int64, parsed *parser.ParsedProtoMessage) {
	detail := parsed.Detail
	userNickname := toString(detail["user"])
	giftName := toString(detail["giftName"])
	giftCount := toString(detail["groupCount"])
	diamondCount := toInt(detail["diamondCount"])
	content := toString(detail["content"])
	anchorID := toString(detail["anchorId"])

	_, err := s.db.GetConnection().Exec(`
		INSERT INTO gift_records (
			session_id, room_id, user_nickname, gift_name, gift_count, gift_diamond_value
		) VALUES (?, ?, ?, ?, ?, ?)
	`, sessionID, roomID, userNickname, giftName, giftCount, diamondCount)

	if err != nil {
		log.Printf("❌ 保存礼物记录失败: %v", err)
		return
	}

	// 尝试分配礼物给主播
	if anchorID == "" {
		var err error
		anchorID, err = s.giftAllocator.AllocateGift(giftName, content)
		if err != nil {
			return
		}
	}
	if anchorID != "" {
		// 记录主播业绩
		if err := s.giftAllocator.RecordAnchorPerformance(anchorID, giftName, diamondCount); err != nil {
			log.Printf("❌ 记录主播业绩失败: %v", err)
		}
	}
}

// saveMessageRecord 保存消息记录
func (s *WebSocketServer) saveMessageRecord(roomID string, sessionID int64, parsed *parser.ParsedProtoMessage) {
	detail := parsed.Detail
	messageType := toString(detail["messageType"])
	userNickname := toString(detail["user"])
	content := toString(detail["content"])

	_, err := s.db.GetConnection().Exec(`
		INSERT INTO message_records (
			session_id, room_id, message_type, user_nickname, content
		) VALUES (?, ?, ?, ?, ?)
	`, sessionID, roomID, messageType, userNickname, content)

	if err != nil {
		log.Printf("❌ 保存消息记录失败: %v", err)
	}
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
