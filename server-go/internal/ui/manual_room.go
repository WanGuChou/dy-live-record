package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	douyinLive "github.com/jwwsjlm/douyinLive"
	newdouyin "github.com/jwwsjlm/douyinLive/generated/new_douyin"

	"dy-live-monitor/internal/parser"
)

type manualRoomConnection struct {
	roomID         string
	live           *douyinLive.DouyinLive
	subscriptionID string
}

// startManualRoom launches a standalone Douyin WSS session for a room.
func (ui *FyneUI) startManualRoom(roomID string) error {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return errors.New("房间号不能为空")
	}

	log.Printf("🚀 [手动房间 %s] 准备建立连接...", roomID)

	ui.roomConnMu.Lock()
	if _, exists := ui.manualRooms[roomID]; exists {
		ui.roomConnMu.Unlock()
		log.Printf("⚠️  [手动房间 %s] 已在监听中，跳过", roomID)
		return fmt.Errorf("房间 %s 已在监听中", roomID)
	}
	ui.roomConnMu.Unlock()

	// 确保 rooms 表中有记录
	if ui.db != nil {
		if err := ui.ensureManualRoomRecord(roomID); err != nil {
			log.Printf("⚠️  [手动房间 %s] 创建房间记录失败: %v", roomID, err)
		}
	}

	logger := log.New(os.Stdout, fmt.Sprintf("[手动房间 %s] ", roomID), log.LstdFlags)
	live, err := douyinLive.NewDouyinLive(roomID, logger)
	if err != nil {
		log.Printf("❌ [手动房间 %s] 创建连接失败: %v", roomID, err)
		return err
	}

	log.Printf("✅ [手动房间 %s] 连接对象创建成功", roomID)

	conn := &manualRoomConnection{
		roomID: roomID,
		live:   live,
	}

	conn.subscriptionID = live.Subscribe(func(eventData *newdouyin.Webcast_Im_Message) {
		ui.handleManualEvent(roomID, eventData)
	})

	log.Printf("📡 [手动房间 %s] 事件订阅已注册", roomID)

	ui.roomConnMu.Lock()
	ui.manualRooms[roomID] = conn
	ui.roomConnMu.Unlock()

	ui.AddOrUpdateRoom(roomID)
	ui.updateOverviewStatus(fmt.Sprintf("状态: 房间 %s 已连接", roomID))

	log.Printf("✅ [手动房间 %s] 房间已添加到监控列表", roomID)

	go func() {
		log.Printf("🔄 [手动房间 %s] 开始监听消息...", roomID)
		live.Start()
		log.Printf("⏹️  [手动房间 %s] 监听已停止", roomID)
		ui.cleanupManualRoom(roomID)
		ui.updateOverviewStatus(fmt.Sprintf("状态: 房间 %s 连接结束", roomID))
	}()

	return nil
}

func (ui *FyneUI) stopManualRoom(roomID string) {
	conn := ui.detachManualRoom(roomID)
	if conn == nil {
		return
	}

	if conn.subscriptionID != "" {
		conn.live.Unsubscribe(conn.subscriptionID)
	}
	conn.live.Close()
}

func (ui *FyneUI) cleanupManualRoom(roomID string) {
	conn := ui.detachManualRoom(roomID)
	if conn == nil {
		return
	}
	if conn.subscriptionID != "" {
		conn.live.Unsubscribe(conn.subscriptionID)
	}
}

func (ui *FyneUI) detachManualRoom(roomID string) *manualRoomConnection {
	ui.roomConnMu.Lock()
	defer ui.roomConnMu.Unlock()

	conn, exists := ui.manualRooms[roomID]
	if !exists {
		return nil
	}

	delete(ui.manualRooms, roomID)
	return conn
}

func (ui *FyneUI) handleManualEvent(roomID string, eventData *newdouyin.Webcast_Im_Message) {
	if eventData == nil {
		return
	}

	log.Printf("📩 [手动房间 %s] 收到事件: %s", roomID, eventData.Method)

	// 确保 rooms 表中有记录
	if ui.db != nil {
		if err := ui.ensureManualRoomRecord(roomID); err != nil {
			log.Printf("⚠️  [手动房间 %s] 确保房间记录失败: %v", roomID, err)
		}
	}

	ui.AddOrUpdateRoom(roomID)

	parsed, err := parser.ParseProtoMessage(eventData.Method, eventData.Payload)
	if err != nil {
		log.Printf("❌ [手动房间 %s] 解析 %s 失败: %v", roomID, eventData.Method, err)
		ui.AddParsedMessage(roomID, fmt.Sprintf("解析 %s 失败: %v", eventData.Method, err))
		return
	}

	log.Printf("✅ [手动房间 %s] 消息解析成功: %s - %s", roomID, parsed.MessageType, parsed.Method)

	// 如果是礼物消息，额外打印详情
	if parsed.MessageType == "礼物消息" {
		giftName := fmt.Sprintf("%v", parsed.Detail["giftName"])
		user := fmt.Sprintf("%v", parsed.Detail["user"])
		count := parsed.Detail["groupCount"]
		diamond := parsed.Detail["diamondCount"]
		log.Printf("🎁 [手动房间 %s] 礼物详情: %s 送出 %s x%v (💎%v)", roomID, user, giftName, count, diamond)
	}

	ui.recordParsedMessage(roomID, parsed, true)
}

// ensureManualRoomRecord 确保手动房间在 rooms 表中有记录
func (ui *FyneUI) ensureManualRoomRecord(roomID string) error {
	if ui.db == nil || roomID == "" {
		return nil
	}

	// 检查是否已存在
	var count int
	err := ui.db.QueryRow(`SELECT COUNT(*) FROM rooms WHERE room_id = ?`, roomID).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询房间记录失败: %w", err)
	}

	if count > 0 {
		// 已存在，更新 last_seen_at
		_, err := ui.db.Exec(`
			UPDATE rooms SET last_seen_at = CURRENT_TIMESTAMP WHERE room_id = ?
		`, roomID)
		if err != nil {
			log.Printf("⚠️  [手动房间 %s] 更新 last_seen_at 失败: %v", roomID, err)
		} else {
			log.Printf("🔄 [手动房间 %s] 房间记录已更新", roomID)
		}
		return nil
	}

	// 不存在，插入新记录
	_, err = ui.db.Exec(`
		INSERT INTO rooms (room_id, room_title, anchor_name, first_seen_at, last_seen_at)
		VALUES (?, '[手动连接]', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roomID)
	
	if err != nil {
		return fmt.Errorf("插入房间记录失败: %w", err)
	}

	log.Printf("✅ [手动房间 %s] 新房间记录已创建", roomID)
	return nil
}

// saveManualGiftRecord 保存手动房间的礼物记录到 gift_records 表
func (ui *FyneUI) saveManualGiftRecord(roomID string, parsed *parser.ParsedProtoMessage) error {
	if ui.db == nil || parsed == nil || parsed.Detail == nil {
		return fmt.Errorf("数据库或消息数据无效")
	}

	log.Printf("🎁 [手动房间 %s] 开始保存礼物记录", roomID)

	// 获取或创建 session_id
	sessionID, err := ui.getOrCreateManualSession(roomID)
	if err != nil {
		return fmt.Errorf("获取 session_id 失败: %w", err)
	}

	// 生成 msgID
	msgID := fmt.Sprintf("%d_%s_%d", time.Now().UnixNano(), parsed.Method, sessionID)
	
	detail := parsed.Detail
	userID := toString(detail["userId"])
	userNickname := toString(detail["user"])
	giftID := toString(detail["giftId"])
	giftName := toString(detail["giftName"])
	giftCount := toInt(detail["groupCount"])
	if giftCount == 0 {
		giftCount = 1
	}
	diamondCount := toInt(detail["diamondCount"])
	anchorID := toString(detail["anchorId"])
	anchorName := toString(detail["anchorName"])

	log.Printf("🎁 [手动房间 %s] 礼物详情 - 用户: %s(%s), 礼物: %s(%s) x%d, 钻石: %d", 
		roomID, userNickname, userID, giftName, giftID, giftCount, diamondCount)

	log.Printf("💾 [手动房间 %s] 准备插入 gift_records 表，msgID: %s, sessionID: %d", roomID, msgID, sessionID)

	result, err := ui.db.Exec(`
		INSERT INTO gift_records (
			msg_id, session_id, room_id, user_id, user_nickname, gift_id, gift_name, 
			gift_count, gift_diamond_value, anchor_id, anchor_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msgID, sessionID, roomID, userID, userNickname, giftID, giftName, giftCount, diamondCount, anchorID, anchorName)

	if err != nil {
		log.Printf("❌ [手动房间 %s] 保存礼物记录失败: %v", roomID, err)
		return err
	}

	recordID, _ := result.LastInsertId()
	log.Printf("✅ [手动房间 %s] 礼物记录已保存到 gift_records 表，recordID: %d, msgID: %s", roomID, recordID, msgID)
	
	return nil
}

// getOrCreateManualSession 获取或创建手动房间的 session_id
func (ui *FyneUI) getOrCreateManualSession(roomID string) (int64, error) {
	// 查找是否已有活跃的 session
	var sessionID int64
	err := ui.db.QueryRow(`
		SELECT session_id FROM live_sessions 
		WHERE room_id = ? 
		ORDER BY start_time DESC 
		LIMIT 1
	`, roomID).Scan(&sessionID)
	
	if err == nil {
		log.Printf("📋 [手动房间 %s] 使用已存在的 sessionID: %d", roomID, sessionID)
		return sessionID, nil
	}

	// 没有找到，创建新的 session
	result, err := ui.db.Exec(`INSERT INTO live_sessions (room_id) VALUES (?)`, roomID)
	if err != nil {
		return 0, fmt.Errorf("创建 live_session 失败: %w", err)
	}

	sessionID, _ = result.LastInsertId()
	log.Printf("✨ [手动房间 %s] 创建新 sessionID: %d", roomID, sessionID)
	return sessionID, nil
}

// 辅助函数：转换接口类型为字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
