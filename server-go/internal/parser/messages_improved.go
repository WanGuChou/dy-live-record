package parser

import (
	"fmt"
	"log"
	"time"
)

// ParseMessagePayloadImproved 改进的消息解析（带详细日志）
func ParseMessagePayloadImproved(method string, payload []byte) map[string]interface{} {
	result := map[string]interface{}{
		"type":        "douyin_live",
		"messageType": method,
		"method":      method,
		"timestamp":   time.Now().Format(time.RFC3339),
		"parsed":      false,
	}

	// 添加调试日志
	if len(payload) == 0 {
		log.Printf("⚠️  [%s] Payload 为空", method)
		return result
	}

	var err error
	var parsed bool

	switch method {
	case "WebcastChatMessage":
		parsed, err = parseChatMessageImproved(payload, result)
	case "WebcastGiftMessage":
		parsed, err = parseGiftMessageImproved(payload, result)
	case "WebcastLikeMessage":
		parsed, err = parseLikeMessageImproved(payload, result)
	case "WebcastMemberMessage":
		parsed, err = parseMemberMessageImproved(payload, result)
	case "WebcastSocialMessage":
		parsed, err = parseSocialMessageImproved(payload, result)
	case "WebcastRoomUserSeqMessage":
		parsed, err = parseRoomUserSeqMessageImproved(payload, result)
	case "WebcastRoomStatsMessage":
		parsed, err = parseRoomStatsMessageImproved(payload, result)
	case "WebcastControlMessage":
		parsed, err = parseControlMessageImproved(payload, result)
	case "WebcastFansclubMessage":
		parsed, err = parseFansclubMessageImproved(payload, result)
	case "WebcastEmojiChatMessage":
		parsed, err = parseEmojiChatMessageImproved(payload, result)
	default:
		log.Printf("⚠️  [%s] 未知消息类型", method)
		return result
	}

	if err != nil {
		log.Printf("❌ [%s] 解析失败: %v (Payload 长度: %d)", method, err, len(payload))
		result["error"] = err.Error()
	} else if parsed {
		result["parsed"] = true
	}

	return result
}

// parseChatMessageImproved 改进的聊天消息解析
func parseChatMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var user *User
	var content string
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common (嵌套消息，需要跳过)
			if err := skipLengthDelimitedField(bb); err != nil {
				return false, fmt.Errorf("跳过 common 失败: %w", err)
			}
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 3: // content
			length, _ := bb.ReadVarint32()
			content, _ = bb.ReadString(int(length))
		case 4: // visibleToSender (bool)
			bb.ReadByte()
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	if user == nil || content == "" {
		return false, fmt.Errorf("缺少必要字段: user=%v, content=%s", user != nil, content)
	}
	
	result["messageType"] = "聊天消息"
	result["user"] = user.Nickname
	result["userId"] = user.ID
	result["content"] = content
	result["level"] = user.Level
	
	return true, nil
}

// parseGiftMessageImproved 改进的礼物消息解析
func parseGiftMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var user *User
	var gift *GiftStruct
	var giftId string
	var repeatCount string
	var comboCount string
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			if err := skipLengthDelimitedField(bb); err != nil {
				return false, fmt.Errorf("跳过 common 失败: %w", err)
			}
		case 2: // giftId
			giftId, _ = bb.ReadVarint64(false)
		case 4: // fanTicketCount
			bb.ReadVarint64(false)
		case 5: // groupCount / repeatCount
			repeatCount, _ = bb.ReadVarint64(false)
		case 6: // repeatEnd
			comboCount, _ = bb.ReadVarint64(false)
		case 7: // textEffect
			skipLengthDelimitedField(bb)
		case 8: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 9: // toUser
			skipLengthDelimitedField(bb)
		case 10: // roomId
			bb.ReadVarint64(false)
		case 11: // timestamp
			bb.ReadVarint64(false)
		case 15: // gift (关键：礼物详情)
			oldLimit, _ := bb.PushTemporaryLength()
			gift, _ = DecodeGiftStructImproved(bb)
			bb.limit = oldLimit
		case 23: // comboCount (另一个可能的字段)
			comboCount, _ = bb.ReadVarint64(false)
		case 25: // monitorExtra
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	if user == nil {
		return false, fmt.Errorf("缺少用户信息")
	}
	
	result["messageType"] = "礼物消息"
	result["user"] = user.Nickname
	result["userId"] = user.ID
	
	if gift != nil {
		result["giftName"] = gift.Name
		result["giftId"] = gift.ID
		result["diamondCount"] = gift.DiamondCount
	} else if giftId != "" {
		result["giftId"] = giftId
	}
	
	if repeatCount != "" && repeatCount != "0" {
		result["giftCount"] = repeatCount
	} else if comboCount != "" && comboCount != "0" {
		result["giftCount"] = comboCount
	} else {
		result["giftCount"] = "1"
	}
	
	return true, nil
}

// parseLikeMessageImproved 改进的点赞消息解析
func parseLikeMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var user *User
	var count string
	var total string
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 3: // count
			count, _ = bb.ReadVarint64(false)
		case 4: // total
			total, _ = bb.ReadVarint64(false)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "点赞消息"
	if user != nil {
		result["user"] = user.Nickname
		result["userId"] = user.ID
	} else {
		result["user"] = "匿名用户"
	}
	result["count"] = count
	result["total"] = total
	
	return true, nil
}

// parseMemberMessageImproved 改进的进入直播间消息解析
func parseMemberMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var user *User
	var memberCount string
	var action int32
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 3: // memberCount
			memberCount, _ = bb.ReadVarint64(false)
		case 4: // operator
			skipLengthDelimitedField(bb)
		case 8: // action (1=进入, 2=关注后进入)
			action, _ = bb.ReadVarint32()
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	if user == nil {
		return false, fmt.Errorf("缺少用户信息")
	}
	
	result["messageType"] = "进入直播间"
	result["user"] = user.Nickname
	result["userId"] = user.ID
	result["memberCount"] = memberCount
	result["action"] = action
	
	return true, nil
}

// parseSocialMessageImproved 改进的关注消息解析
func parseSocialMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var user *User
	var followCount string
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 3: // followCount
			followCount, _ = bb.ReadVarint64(false)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	if user == nil {
		return false, fmt.Errorf("缺少用户信息")
	}
	
	result["messageType"] = "关注消息"
	result["user"] = user.Nickname
	result["userId"] = user.ID
	result["followCount"] = followCount
	
	return true, nil
}

// parseRoomUserSeqMessageImproved 改进的在线人数消息解析
func parseRoomUserSeqMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var total string
	var totalUser string
	var totalPvForAnchor string
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // total
			total, _ = bb.ReadVarint64(false)
		case 3: // totalUser
			totalUser, _ = bb.ReadVarint64(false)
		case 4: // totalPvForAnchor
			totalPvForAnchor, _ = bb.ReadVarint64(false)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "在线人数"
	result["total"] = total
	result["totalUser"] = totalUser
	result["totalPvForAnchor"] = totalPvForAnchor
	
	return true, nil
}

// parseRoomStatsMessageImproved 改进的直播间统计消息解析
func parseRoomStatsMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var displayShort string
	var displayMiddle string
	var displayLong string
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // displayShort
			length, _ := bb.ReadVarint32()
			displayShort, _ = bb.ReadString(int(length))
		case 3: // displayMiddle
			length, _ := bb.ReadVarint32()
			displayMiddle, _ = bb.ReadString(int(length))
		case 4: // displayLong
			length, _ := bb.ReadVarint32()
			displayLong, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "直播间统计"
	result["displayShort"] = displayShort
	result["displayMiddle"] = displayMiddle
	result["displayLong"] = displayLong
	
	return true, nil
}

// parseControlMessageImproved 控制消息解析
func parseControlMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var action int32
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // action
			action, _ = bb.ReadVarint32()
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "控制消息"
	result["action"] = action
	
	return true, nil
}

// parseFansclubMessageImproved 粉丝团消息解析
func parseFansclubMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var user *User
	var type_ int32
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // type
			type_, _ = bb.ReadVarint32()
		case 4: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "粉丝团消息"
	if user != nil {
		result["user"] = user.Nickname
		result["userId"] = user.ID
	}
	result["type"] = type_
	
	return true, nil
}

// parseEmojiChatMessageImproved 表情聊天消息解析
func parseEmojiChatMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var user *User
	var content string
	var emojiId string
	
	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 3: // content
			length, _ := bb.ReadVarint32()
			content, _ = bb.ReadString(int(length))
		case 4: // emojiId
			length, _ := bb.ReadVarint32()
			emojiId, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "表情消息"
	if user != nil {
		result["user"] = user.Nickname
		result["userId"] = user.ID
	}
	result["content"] = content
	result["emojiId"] = emojiId
	
	return true, nil
}

// DecodeUserImproved 改进的用户解码（处理更多字段）
func DecodeUserImproved(bb *ByteBuffer) (*User, error) {
	user := &User{}

	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // id
			user.ID, _ = bb.ReadVarint64(false)
		case 2: // shortId
			user.ShortID, _ = bb.ReadVarint64(false)
		case 3: // nickname
			length, _ := bb.ReadVarint32()
			user.Nickname, _ = bb.ReadString(int(length))
		case 4: // gender
			user.Gender, _ = bb.ReadVarint32()
		case 6: // level
			user.Level, _ = bb.ReadVarint32()
		case 9, 10, 11: // avatarThumb, avatarMedium, avatarLarge (Image)
			skipLengthDelimitedField(bb)
		case 22, 23, 24, 25, 26: // followInfo, payGrade, fansClub, border, specialId
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return user, nil
}

// DecodeGiftStructImproved 改进的礼物结构解码
func DecodeGiftStructImproved(bb *ByteBuffer) (*GiftStruct, error) {
	gift := &GiftStruct{}

	for !bb.IsAtEnd() {
		tag, err := bb.ReadVarint32()
		if err != nil {
			break
		}

		fieldNumber := tag >> 3
		wireType := int(tag & 7)

		if fieldNumber == 0 {
			break
		}

		switch fieldNumber {
		case 1: // image (Image)
			skipLengthDelimitedField(bb)
		case 2: // describe
			skipLengthDelimitedField(bb)
		case 5: // id
			gift.ID, _ = bb.ReadVarint64(false)
		case 7: // type
			bb.ReadVarint32()
		case 12: // diamondCount
			gift.DiamondCount, _ = bb.ReadVarint32()
		case 16: // name
			length, _ := bb.ReadVarint32()
			gift.Name, _ = bb.ReadString(int(length))
		case 22: // icon (Image)
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return gift, nil
}

// skipLengthDelimitedField 跳过 length-delimited 字段（wire type 2）
func skipLengthDelimitedField(bb *ByteBuffer) error {
	length, err := bb.ReadVarint32()
	if err != nil {
		return err
	}
	_, err = bb.Advance(int(length))
	return err
}
