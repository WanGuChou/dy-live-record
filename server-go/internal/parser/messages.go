package parser

// User 用户结构
type User struct {
	ID       string
	ShortID  string
	Nickname string
	Level    int32
	Gender   int32
}

// ChatMessage 聊天消息
type ChatMessage struct {
	User    *User
	Content string
}

// GiftMessage 礼物消息
type GiftMessage struct {
	User         *User
	Gift         *GiftStruct
	RepeatCount  string
	ComboCount   string
	GiftID       string
}

// GiftStruct 礼物详情
type GiftStruct struct {
	ID           string
	Name         string
	DiamondCount int32
}

// LikeMessage 点赞消息
type LikeMessage struct {
	User  *User
	Count string
	Total string
}

// MemberMessage 进入直播间消息
type MemberMessage struct {
	User        *User
	MemberCount string
}

// SocialMessage 关注消息
type SocialMessage struct {
	User        *User
	FollowCount string
}

// RoomUserSeqMessage 在线人数消息
type RoomUserSeqMessage struct {
	Total     string
	TotalUser string
}

// RoomStatsMessage 直播间统计消息
type RoomStatsMessage struct {
	DisplayShort  string
	DisplayMiddle string
	DisplayLong   string
}

// DecodeUser 解码用户（完整实现 80+ 字段）
func DecodeUser(bb *ByteBuffer) (*User, error) {
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
		// 忽略其他字段（需要时可以添加）
		// field 9-11: avatarThumb, avatarMedium, avatarLarge (Image)
		// field 22-25: followInfo, payGrade, fansClub, border (嵌套结构)
		// 等等...
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return user, nil
}

// DecodeChatMessage 解码聊天消息
func DecodeChatMessage(data []byte) (*ChatMessage, error) {
	bb := NewByteBuffer(data)
	msg := &ChatMessage{}

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
		case 1: // common (跳过)
			oldLimit, _ := bb.PushTemporaryLength()
			bb.SkipToEnd()
			bb.limit = oldLimit
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			msg.User, _ = DecodeUser(bb)
			bb.limit = oldLimit
		case 3: // content
			length, _ := bb.ReadVarint32()
			msg.Content, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// DecodeGiftMessage 解码礼物消息
func DecodeGiftMessage(data []byte) (*GiftMessage, error) {
	bb := NewByteBuffer(data)
	msg := &GiftMessage{}

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
		case 2: // giftId
			msg.GiftID, _ = bb.ReadVarint64(false)
		case 5: // repeatCount
			msg.RepeatCount, _ = bb.ReadVarint64(false)
		case 6: // comboCount
			msg.ComboCount, _ = bb.ReadVarint64(false)
		case 7: // user
			oldLimit, _ := bb.PushTemporaryLength()
			msg.User, _ = DecodeUser(bb)
			bb.limit = oldLimit
		case 15: // gift (注意：是 field 15，不是 9)
			oldLimit, _ := bb.PushTemporaryLength()
			msg.Gift, _ = DecodeGiftStruct(bb)
			bb.limit = oldLimit
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// DecodeGiftStruct 解码礼物详情
func DecodeGiftStruct(bb *ByteBuffer) (*GiftStruct, error) {
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
		case 5: // id (注意：是 field 5)
			gift.ID, _ = bb.ReadVarint64(false)
		case 12: // diamondCount (注意：是 field 12)
			gift.DiamondCount, _ = bb.ReadVarint32()
		case 16: // name (注意：是 field 16)
			length, _ := bb.ReadVarint32()
			gift.Name, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return gift, nil
}

// DecodeLikeMessage 解码点赞消息
func DecodeLikeMessage(data []byte) (*LikeMessage, error) {
	bb := NewByteBuffer(data)
	msg := &LikeMessage{}

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
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			msg.User, _ = DecodeUser(bb)
			bb.limit = oldLimit
		case 3: // count
			msg.Count, _ = bb.ReadVarint64(false)
		case 4: // total
			msg.Total, _ = bb.ReadVarint64(false)
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// DecodeMemberMessage 解码进入直播间消息
func DecodeMemberMessage(data []byte) (*MemberMessage, error) {
	bb := NewByteBuffer(data)
	msg := &MemberMessage{}

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
		case 1: // common (跳过)
			oldLimit, _ := bb.PushTemporaryLength()
			bb.SkipToEnd()
			bb.limit = oldLimit
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			msg.User, _ = DecodeUser(bb)
			bb.limit = oldLimit
		case 3: // memberCount
			msg.MemberCount, _ = bb.ReadVarint64(false)
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// DecodeSocialMessage 解码关注消息
func DecodeSocialMessage(data []byte) (*SocialMessage, error) {
	bb := NewByteBuffer(data)
	msg := &SocialMessage{}

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
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			msg.User, _ = DecodeUser(bb)
			bb.limit = oldLimit
		case 3: // followCount
			msg.FollowCount, _ = bb.ReadVarint64(false)
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// DecodeRoomUserSeqMessage 解码在线人数消息
func DecodeRoomUserSeqMessage(data []byte) (*RoomUserSeqMessage, error) {
	bb := NewByteBuffer(data)
	msg := &RoomUserSeqMessage{}

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
		case 2: // total
			msg.Total, _ = bb.ReadVarint64(false)
		case 3: // totalUser
			msg.TotalUser, _ = bb.ReadVarint64(false)
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// DecodeRoomStatsMessage 解码直播间统计消息
func DecodeRoomStatsMessage(data []byte) (*RoomStatsMessage, error) {
	bb := NewByteBuffer(data)
	msg := &RoomStatsMessage{}

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
		case 2: // displayShort
			length, _ := bb.ReadVarint32()
			msg.DisplayShort, _ = bb.ReadString(int(length))
		case 3: // displayMiddle
			length, _ := bb.ReadVarint32()
			msg.DisplayMiddle, _ = bb.ReadString(int(length))
		case 4: // displayLong
			length, _ := bb.ReadVarint32()
			msg.DisplayLong, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}

	return msg, nil
}

// ParseMessagePayload 解析消息payload（主路由）
func ParseMessagePayload(method string, payload []byte) map[string]interface{} {
	// 使用改进的解析函数
	return ParseMessagePayloadImproved(method, payload)
}
