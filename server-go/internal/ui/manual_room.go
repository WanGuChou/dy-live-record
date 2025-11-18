package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	douyinLive "github.com/jwwsjlm/douyinLive"
	generatedmsg "github.com/jwwsjlm/douyinLive/generated"
	newdouyin "github.com/jwwsjlm/douyinLive/generated/new_douyin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

	ui.roomConnMu.Lock()
	if _, exists := ui.manualRooms[roomID]; exists {
		ui.roomConnMu.Unlock()
		return fmt.Errorf("房间 %s 已在监听中", roomID)
	}
	ui.roomConnMu.Unlock()

	logger := log.New(os.Stdout, fmt.Sprintf("[手动房间 %s] ", roomID), log.LstdFlags)
	live, err := douyinLive.NewDouyinLive(roomID, logger)
	if err != nil {
		return err
	}

	conn := &manualRoomConnection{
		roomID: roomID,
		live:   live,
	}

	conn.subscriptionID = live.Subscribe(func(eventData *newdouyin.Webcast_Im_Message) {
		ui.handleManualEvent(roomID, eventData)
	})

	ui.roomConnMu.Lock()
	ui.manualRooms[roomID] = conn
	ui.roomConnMu.Unlock()

	ui.AddOrUpdateRoom(roomID)
	ui.updateOverviewStatus(fmt.Sprintf("状态: 房间 %s 已连接", roomID))

	go func() {
		live.Start()
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

	ui.AddOrUpdateRoom(roomID)

	rawSummary := fmt.Sprintf("Method: %s | MsgID: %d | Payload: %d bytes", eventData.Method, eventData.MsgId, len(eventData.Payload))
	ui.AddRawMessage(roomID, rawSummary)

	msgInstance, err := generatedmsg.GetMessageInstance(eventData.Method)
	if err != nil {
		ui.AddParsedMessage(roomID, fmt.Sprintf("未注册的消息类型: %s", eventData.Method))
		return
	}

	protoMsg, ok := msgInstance.(proto.Message)
	if !ok {
		ui.AddParsedMessage(roomID, fmt.Sprintf("消息类型 %s 不支持当前解析方式", eventData.Method))
		return
	}

	if err := proto.Unmarshal(eventData.Payload, protoMsg); err != nil {
		ui.AddParsedMessage(roomID, fmt.Sprintf("解析 %s 失败: %v", eventData.Method, err))
		return
	}

	display, detail := ui.buildManualMessageDisplay(eventData.Method, protoMsg)
	ui.AddParsedMessageWithDetail(roomID, display, detail)
}

func (ui *FyneUI) buildManualMessageDisplay(method string, message proto.Message) (string, map[string]interface{}) {
	detail := map[string]interface{}{
		"method": method,
	}

	switch m := message.(type) {
	case *newdouyin.Webcast_Im_ChatMessage:
		user := safeNickname(m.User)
		detail["user"] = user
		detail["userId"] = m.GetUser().GetId()
		detail["content"] = m.GetContent()
		return fmt.Sprintf("聊天 | %s: %s", user, m.GetContent()), detail

	case *newdouyin.Webcast_Im_GiftMessage:
		describeGiftMessage(detail, m)
		return fmt.Sprintf("礼物 | %s -> %s: %s x%d",
			safeNickname(m.User),
			safeNickname(m.ToUser),
			giftName(m.Gift),
			m.GetGroupCount(),
		), detail

	case *newdouyin.Webcast_Im_LikeMessage:
		user := safeNickname(m.User)
		detail["user"] = user
		detail["count"] = m.GetCount()
		detail["total"] = m.GetTotal()
		return fmt.Sprintf("点赞 | %s 点赞 %d 次 (总计 %d)", user, m.GetCount(), m.GetTotal()), detail

	case *newdouyin.Webcast_Im_MemberMessage:
		user := safeNickname(m.User)
		detail["user"] = user
		detail["memberCount"] = m.GetMemberCount()
		return fmt.Sprintf("进场 | %s 加入直播间，当前人数 %d", user, m.GetMemberCount()), detail

	case *newdouyin.Webcast_Im_SocialMessage:
		user := safeNickname(m.User)
		detail["user"] = user
		return fmt.Sprintf("关注 | %s 关注了主播", user), detail

	default:
		raw, _ := protojson.Marshal(message)
		detail["payload"] = string(raw)
		return fmt.Sprintf("%s 消息", method), detail
	}
}

func describeGiftMessage(detail map[string]interface{}, msg *newdouyin.Webcast_Im_GiftMessage) {
	user := msg.GetUser()
	toUser := msg.GetToUser()
	gift := msg.GetGift()

	detail["user"] = safeNickname(user)
	detail["userId"] = user.GetId()
	detail["toUser"] = safeNickname(toUser)
	detail["toUserId"] = toUser.GetId()
	detail["giftName"] = giftName(gift)
	detail["diamondCount"] = gift.GetDiamondCount()
	detail["groupCount"] = msg.GetGroupCount()
	detail["repeatCount"] = msg.GetRepeatCount()

	log.Printf("礼物消息:user=%d %s ->%s %s 组数%d 钻石：%d ",
		user.GetId(),
		user.GetNickname(),
		toUser.GetNickname(),
		gift.GetName(),
		msg.GetGroupCount(),
		gift.GetDiamondCount(),
	)
}

func safeNickname(user *newdouyin.Webcast_Data_User) string {
	if user == nil {
		return "匿名"
	}
	if nick := strings.TrimSpace(user.GetNickname()); nick != "" {
		return nick
	}
	return fmt.Sprintf("用户%d", user.GetId())
}

func giftName(gift *newdouyin.Webcast_Data_GiftStruct) string {
	if gift == nil {
		return "未知礼物"
	}
	if name := strings.TrimSpace(gift.GetName()); name != "" {
		return name
	}
	return fmt.Sprintf("礼物%d", gift.GetId())
}
