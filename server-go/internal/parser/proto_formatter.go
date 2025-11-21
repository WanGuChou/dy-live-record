package parser

import (
	"fmt"
	"strings"
	"time"

	generatedmsg "dy-live-monitor/internal/jwwsjlm/douyinLive/generated"
	newdouyin "dy-live-monitor/internal/jwwsjlm/douyinLive/generated/new_douyin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ParsedProtoMessage 表示通过 protobuf 解析后的直播消息
type ParsedProtoMessage struct {
	Method      string
	Display     string
	Detail      map[string]interface{}
	Proto       proto.Message
	RawJSON     string
	RawPayload  []byte
	ReceivedAt  time.Time
	MessageID   string
	MessageType string
}

// ParseProtoMessage 解析 protobuf 消息并格式化为可展示的数据
func ParseProtoMessage(method string, payload []byte) (*ParsedProtoMessage, error) {
	instance, err := generatedmsg.GetMessageInstance(method)
	if err != nil {
		return nil, fmt.Errorf("未注册的 protobuf 消息: %s: %w", method, err)
	}

	protoMsg, ok := instance.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("消息 %s 无法转换为 proto.Message", method)
	}

	if err := proto.Unmarshal(payload, protoMsg); err != nil {
		return nil, fmt.Errorf("解析消息 %s 失败: %w", method, err)
	}

	display, detail := formatDisplay(method, protoMsg)
	rawJSON, _ := protojson.Marshal(protoMsg)

	return &ParsedProtoMessage{
		Method:      method,
		Display:     display,
		Detail:      detail,
		Proto:       protoMsg,
		RawJSON:     string(rawJSON),
		RawPayload:  payload,
		ReceivedAt:  time.Now(),
		MessageID:   fmt.Sprintf("%p-%d", protoMsg, time.Now().UnixNano()),
		MessageType: detail["messageType"].(string),
	}, nil
}

func formatDisplay(method string, message proto.Message) (string, map[string]interface{}) {
	detail := map[string]interface{}{
		"method":      method,
		"messageType": method,
	}

	switch msg := message.(type) {
	case *newdouyin.Webcast_Im_ChatMessage:
		user := safeNickname(msg.User)
		detail["messageType"] = "聊天消息"
		detail["user"] = user
		detail["userId"] = msg.GetUser().GetId()
		detail["content"] = msg.GetContent()
		return fmt.Sprintf("💬 %s: %s", user, msg.GetContent()), detail

	case *newdouyin.Webcast_Im_GiftMessage:
		return formatGiftMessage(detail, msg)

	case *newdouyin.Webcast_Im_LikeMessage:
		user := safeNickname(msg.User)
		detail["messageType"] = "点赞消息"
		detail["user"] = user
		detail["count"] = msg.GetCount()
		detail["total"] = msg.GetTotal()
		return fmt.Sprintf("❤️ %s 点赞 %d 次 (总计 %d)", user, msg.GetCount(), msg.GetTotal()), detail

	case *newdouyin.Webcast_Im_MemberMessage:
		user := safeNickname(msg.User)
		detail["messageType"] = "进场消息"
		detail["user"] = user
		detail["memberCount"] = msg.GetMemberCount()
		return fmt.Sprintf("🚪 %s 进入直播间，当前人数 %d", user, msg.GetMemberCount()), detail

	case *newdouyin.Webcast_Im_SocialMessage:
		user := safeNickname(msg.User)
		detail["messageType"] = "关注消息"
		detail["user"] = user
		return fmt.Sprintf("⭐ %s 关注了主播", user), detail

	default:
		raw, _ := protojson.Marshal(message)
		detail["messageType"] = method
		detail["payload"] = string(raw)
		return fmt.Sprintf("📦 %s (%T)", method, message), detail
	}
}

func formatGiftMessage(detail map[string]interface{}, msg *newdouyin.Webcast_Im_GiftMessage) (string, map[string]interface{}) {
	user := msg.GetUser()
	toUser := msg.GetToUser()
	gift := msg.GetGift()

	detail["messageType"] = "礼物消息"
	detail["user"] = safeNickname(user)
	detail["userId"] = user.GetId()

	if toUser != nil {
		detail["toUser"] = safeNickname(toUser)
		detail["toUserId"] = toUser.GetId()
		detail["anchorId"] = toUser.GetId()
		detail["anchorName"] = safeNickname(toUser)
	}

	if gift != nil {
		detail["giftName"] = giftName(gift)
		detail["giftId"] = gift.GetId()
		detail["diamondCount"] = int(gift.GetDiamondCount())
	}

	detail["groupCount"] = msg.GetGroupCount()
	detail["repeatCount"] = msg.GetRepeatCount()
	detail["totalCount"] = msg.GetTotalCount()

	display := fmt.Sprintf("🎁 %s 送出 %s x%d",
		safeNickname(user),
		giftName(gift),
		msg.GetGroupCount(),
	)

	if toUser != nil {
		display = fmt.Sprintf("%s -> %s", display, safeNickname(toUser))
	}

	return display, detail
}

func safeNickname(user *newdouyin.Webcast_Data_User) string {
	if user == nil {
		return "匿名"
	}
	if nick := strings.TrimSpace(user.GetNickname()); nick != "" {
		return nick
	}
	if user.GetDisplayId() != "" {
		return user.GetDisplayId()
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
