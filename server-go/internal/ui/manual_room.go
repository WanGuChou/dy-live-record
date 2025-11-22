package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	douyinLive "dy-live-monitor/internal/jwwsjlm/douyinLive"
	newdouyin "dy-live-monitor/internal/jwwsjlm/douyinLive/generated/new_douyin"
	"github.com/tidwall/gjson"

	"dy-live-monitor/internal/parser"
)

type manualRoomConnection struct {
	liveRoomID     string
	roomID         string
	liveName       string
	wsURL          string
	live           *douyinLive.DouyinLive
	subscriptionID string
}

// startManualRoom launches a standalone Douyin WSS session for a room.
func (ui *FyneUI) startManualRoom(liveRoomID string) (string, error) {
	liveRoomID = strings.TrimSpace(liveRoomID)
	if liveRoomID == "" {
		return "", errors.New("live_room_id 不能为空")
	}

	log.Printf("🚀 [手动房间 %s] 准备建立连接...", liveRoomID)

	logger := log.New(os.Stdout, fmt.Sprintf("[手动房间 %s] ", liveRoomID), log.LstdFlags)
	live, err := douyinLive.NewDouyinLive(liveRoomID, logger)
	if err != nil {
		log.Printf("❌ [手动房间 %s] 创建连接失败: %v", liveRoomID, err)
		return "", err
	}

	roomID, liveName, wsURL, err := live.ConnectionInfo()
	if err != nil {
		log.Printf("❌ [手动房间 %s] 获取房间信息失败: %v", liveRoomID, err)
		return "", err
	}
	if strings.TrimSpace(roomID) == "" {
		return "", fmt.Errorf("未获取到 room_id")
	}

	ui.roomConnMu.Lock()
	if _, exists := ui.manualRooms[roomID]; exists {
		ui.roomConnMu.Unlock()
		log.Printf("⚠️  [手动房间 %s] 已在监听中，跳过", roomID)
		return "", fmt.Errorf("房间 %s 已在监听中", roomID)
	}
	ui.roomConnMu.Unlock()

	// 确保 rooms 表中有记录
	if ui.db != nil {
		roomTitle := fmt.Sprintf("房间-%s", liveName)
		if err := ui.ensureRoomMetadata(roomID, liveRoomID, roomTitle, wsURL); err != nil {
			log.Printf("⚠️  [手动房间 %s] 创建房间记录失败: %v", roomID, err)
		}
	}

	// 设置数据库连接
	if ui.db != nil {
		live.SetDB(ui.db)
		log.Printf("✅ [手动房间 %s] 数据库连接已设置", roomID)
	}

	log.Printf("✅ [手动房间 %s] 连接对象创建成功", roomID)

	conn := &manualRoomConnection{
		liveRoomID: liveRoomID,
		roomID:     roomID,
		liveName:   liveName,
		wsURL:      wsURL,
		live:       live,
	}

	conn.subscriptionID = live.Subscribe(func(eventData *newdouyin.Webcast_Im_Message) {
		ui.handleManualEvent(roomID, eventData)
	})

	log.Printf("📡 [手动房间 %s] 事件订阅已注册", roomID)

	ui.roomConnMu.Lock()
	ui.manualRooms[roomID] = conn
	ui.roomConnMu.Unlock()

	ui.AddOrUpdateRoom(roomID)
	displayName := ui.lookupRoomTitle(roomID)
	ui.updateOverviewStatus(fmt.Sprintf("状态: %s 已连接", displayName))

	log.Printf("✅ [手动房间 %s] 房间已添加到监控列表", roomID)

	go func(display string) {
		log.Printf("🔄 [手动房间 %s] 开始监听消息...", roomID)
		live.Start()
		log.Printf("⏹️  [手动房间 %s] 监听已停止", roomID)
		ui.cleanupManualRoom(roomID)
		ui.updateOverviewStatus(fmt.Sprintf("状态: %s 连接结束", display))
	}(displayName)

	return displayName, nil
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
		if err := ui.ensureRoomMetadata(roomID, "", "", ""); err != nil {
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

// ensureRoomMetadata upserts basic room information.
func (ui *FyneUI) ensureRoomMetadata(roomID, liveRoomID, roomTitle, wsURL string) error {
	if ui.db == nil || strings.TrimSpace(roomID) == "" {
		return nil
	}

	var count int
	if err := ui.db.QueryRow(`SELECT COUNT(*) FROM rooms WHERE room_id = ?`, roomID).Scan(&count); err != nil {
		return fmt.Errorf("查询房间记录失败: %w", err)
	}

	if count > 0 {
		setClauses := []string{"last_seen_at = CURRENT_TIMESTAMP"}
		args := make([]interface{}, 0, 4)
		if strings.TrimSpace(roomTitle) != "" {
			setClauses = append(setClauses, "room_title = ?")
			args = append(args, roomTitle)
		}
		if strings.TrimSpace(liveRoomID) != "" {
			setClauses = append(setClauses, "live_room_id = ?")
			args = append(args, liveRoomID)
		}
		if strings.TrimSpace(wsURL) != "" {
			setClauses = append(setClauses, "ws_url = ?")
			args = append(args, wsURL)
		}
		args = append(args, roomID)
		query := fmt.Sprintf("UPDATE rooms SET %s WHERE room_id = ?", strings.Join(setClauses, ", "))
		_, err := ui.db.Exec(query, args...)
		return err
	}

	_, err := ui.db.Exec(`
		INSERT INTO rooms (room_id, live_room_id, room_title, ws_url, anchor_name, first_seen_at, last_seen_at)
		VALUES (?, ?, ?, ?, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, roomID, liveRoomID, roomTitle, wsURL)
	return err
}

// saveManualGiftRecord 保存手动房间的礼物记录到 gift_records 表
func (ui *FyneUI) saveManualGiftRecord(roomID string, parsed *parser.ParsedProtoMessage) error {
	if ui.db == nil || parsed == nil || parsed.Detail == nil {
		return fmt.Errorf("数据库或消息数据无效")
	}

	log.Printf("🎁 [手动房间 %s] 开始保存礼物记录", roomID)

	// 生成 msgID
	msgID := gjson.Get(parsed.RawJSON, "common.msgId")

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
	anchorID := normalizeAnchorID(toString(detail["anchorId"]))
	anchorName := normalizeAnchorName(toString(detail["anchorName"]))

	log.Printf("🎁 [手动房间 %s] 礼物详情 - 用户: %s(%s), 礼物: %s(%s) x%d, 钻石: %d",
		roomID, userNickname, userID, giftName, giftID, giftCount, diamondCount)

	log.Printf("💾 [手动房间 %s] 准备插入 gift_records 表，msgID: %s", roomID, msgID)

	result, err := ui.db.Exec(`
		INSERT INTO gift_records (
			msg_id, room_id, user_id, user_nickname, gift_id, gift_name, 
			gift_count, gift_diamond_value, anchor_id, anchor_name
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msgID, roomID, userID, userNickname, giftID, giftName, giftCount, diamondCount, anchorID, anchorName)

	if err != nil {
		log.Printf("❌ [手动房间 %s] 保存礼物记录失败: %v", roomID, err)
		return err
	}

	recordID, _ := result.LastInsertId()
	log.Printf("✅ [手动房间 %s] 礼物记录已保存到 gift_records 表，recordID: %d, msgID: %s", roomID, recordID, msgID)

	if anchorID != "" {
		if strings.TrimSpace(anchorName) == "" {
			anchorName = ui.lookupAnchorName(anchorID)
		}
		ui.ensureGlobalAnchor(anchorID, anchorName)
		ui.ensureRoomAnchorRecord(roomID, anchorID, anchorName)
	}

	return nil
}

// 辅助函数：转换接口类型为字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
