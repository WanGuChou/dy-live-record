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

	"github.com/gorilla/websocket"
)

// WebSocketServer WebSocket服务器
type WebSocketServer struct {
	port      int
	db        *database.DB
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	rooms     map[string]*RoomManager
	roomsMu   sync.RWMutex
	upgrader  websocket.Upgrader
}

// RoomManager 房间管理器
type RoomManager struct {
	RoomID      string
	SessionID   int64
	LastMessage int64
	Parser      *parser.DouyinParser
}

// NewWebSocketServer 创建WebSocket服务器
func NewWebSocketServer(port int, db *database.DB) *WebSocketServer {
	return &WebSocketServer{
		port:    port,
		db:      db,
		clients: make(map[*websocket.Conn]bool),
		rooms:   make(map[string]*RoomManager),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源（生产环境应限制）
			},
		},
	}
}

// Start 启动WebSocket服务器
func (s *WebSocketServer) Start() error {
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/health", s.handleHealth)
	
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 WebSocket 服务器监听: %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleWebSocket 处理WebSocket连接
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket 升级失败: %v", err)
		return
	}

	log.Printf("✅ 新客户端连接: %s", conn.RemoteAddr())

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

	// 解析抖音消息
	parsedMessages, err := room.Parser.ParseMessage(payloadData, url)
	if err != nil {
		log.Printf("❌ [房间 %s] 解析失败: %v", roomID, err)
		return
	}

	if parsedMessages == nil || len(parsedMessages) == 0 {
		return
	}

	// 存储到数据库
	for _, msg := range parsedMessages {
		s.saveMessage(roomID, room.SessionID, msg)
	}

	// 打印格式化消息
	formatted := room.Parser.FormatMessage(parsedMessages)
	if formatted != "" {
		log.Println(formatted)
	}
}

// handleRequest 处理HTTP请求记录
func (s *WebSocketServer) handleRequest(data map[string]interface{}) {
	// 可选：记录所有HTTP请求
	url, _ := data["url"].(string)
	log.Printf("🌐 请求: %s", url)
}

// handleHealth 健康检查接口
func (s *WebSocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
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
			Parser: parser.NewDouyinParser(),
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
func (s *WebSocketServer) saveMessage(roomID string, sessionID int64, msg map[string]interface{}) {
	messageType, _ := msg["messageType"].(string)

	switch messageType {
	case "礼物消息":
		s.saveGiftRecord(roomID, sessionID, msg)
	case "聊天消息", "进入直播间", "关注消息":
		s.saveMessageRecord(roomID, sessionID, msg)
	}
}

// saveGiftRecord 保存礼物记录
func (s *WebSocketServer) saveGiftRecord(roomID string, sessionID int64, msg map[string]interface{}) {
	userNickname, _ := msg["user"].(string)
	giftName, _ := msg["giftName"].(string)
	giftCount, _ := msg["giftCount"].(string)
	diamondCount, _ := msg["diamondCount"].(int)

	_, err := s.db.GetConnection().Exec(`
		INSERT INTO gift_records (
			session_id, room_id, user_nickname, gift_name, gift_count, gift_diamond_value
		) VALUES (?, ?, ?, ?, ?, ?)
	`, sessionID, roomID, userNickname, giftName, giftCount, diamondCount)

	if err != nil {
		log.Printf("❌ 保存礼物记录失败: %v", err)
	}
}

// saveMessageRecord 保存消息记录
func (s *WebSocketServer) saveMessageRecord(roomID string, sessionID int64, msg map[string]interface{}) {
	messageType, _ := msg["messageType"].(string)
	userNickname, _ := msg["user"].(string)
	content, _ := msg["content"].(string)

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
