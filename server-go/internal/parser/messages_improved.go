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
	case "WebcastRoomMessage":
		parsed, err = parseRoomMessageImproved(payload, result)
	case "WebcastMatchAgainstScoreMessage":
		parsed, err = parseMatchAgainstScoreMessageImproved(payload, result)
	case "WebcastRankUpdateMessage":
		parsed, err = parseRankUpdateMessageImproved(payload, result)
	case "WebcastLinkMicMessage":
		parsed, err = parseLinkMicMessageImproved(payload, result)
	case "WebcastLinkMicBattle":
		parsed, err = parseLinkMicBattleImproved(payload, result)
	case "WebcastLinkMicArmies":
		parsed, err = parseLinkMicArmiesImproved(payload, result)
	case "WebcastInRoomBannerMessage":
		parsed, err = parseInRoomBannerMessageImproved(payload, result)
	case "WebcastProductChangeMessage":
		parsed, err = parseProductChangeMessageImproved(payload, result)
	case "WebcastCommonTextMessage":
		parsed, err = parseCommonTextMessageImproved(payload, result)
	case "WebcastBarrageMessage":
		parsed, err = parseBarrageMessageImproved(payload, result)
	case "WebcastRoomRankMessage":
		parsed, err = parseRoomRankMessageImproved(payload, result)
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
	var toUser *User
	var gift *GiftStruct
	var giftId string
	var repeatCount string
	var comboCount string
	var groupCount string
	var repeatEnd string
	var totalCoin string
	var timestamp string
	var logId string
	var sendType int32
	var publicArea int32
	
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
		case 3: // fanTicketCount
			bb.ReadVarint64(false)
		case 4: // groupCount (数量)
			groupCount, _ = bb.ReadVarint64(false)
		case 5: // repeatCount (重复次数)
			repeatCount, _ = bb.ReadVarint64(false)
		case 6: // comboCount (连击次数)
			comboCount, _ = bb.ReadVarint64(false)
		case 7: // user (发送者)
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 8: // toUser (接收者，用于礼物PK)
			oldLimit, _ := bb.PushTemporaryLength()
			toUser, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 9: // repeatEnd (是否连击结束, 0=继续, 1=结束)
			repeatEnd, _ = bb.ReadVarint64(false)
		case 10: // textEffect
			skipLengthDelimitedField(bb)
		case 11: // groupId
			bb.ReadVarint64(false)
		case 12: // incomeTaskGifts
			bb.ReadVarint64(false)
		case 13: // roomFanTicketCount
			bb.ReadVarint64(false)
		case 14: // priority
			skipLengthDelimitedField(bb)
		case 15: // gift (关键：礼物详情)
			oldLimit, _ := bb.PushTemporaryLength()
			gift, _ = DecodeGiftStructImproved(bb)
			bb.limit = oldLimit
		case 16: // logId
			length, _ := bb.ReadVarint32()
			logId, _ = bb.ReadString(int(length))
		case 17: // sendType (1=普通礼物, 2=投喂礼物)
			sendType, _ = bb.ReadVarint32()
		case 18: // publicAreaCommon
			skipLengthDelimitedField(bb)
		case 19: // trayDisplayText
			skipLengthDelimitedField(bb)
		case 20: // bannedDisplayText
			skipLengthDelimitedField(bb)
		case 21: // timestamp
			timestamp, _ = bb.ReadVarint64(false)
		case 22: // diyGiftInfo
			skipLengthDelimitedField(bb)
		case 23: // totalCoin (总价值)
			totalCoin, _ = bb.ReadVarint64(false)
		case 24: // sendGiftProfitCoreUserInfo
			skipLengthDelimitedField(bb)
		case 25: // toUserInfo
			skipLengthDelimitedField(bb)
		case 26: // comboGift
			bb.ReadByte()
		case 27: // monitorExtra
			skipLengthDelimitedField(bb)
		case 28: // anchorGift
			bb.ReadByte()
		case 29: // publicArea (是否公屏显示)
			publicArea, _ = bb.ReadVarint32()
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
	result["userLevel"] = user.Level
	
	// 接收者信息
	if toUser != nil {
		result["toUser"] = toUser.Nickname
		result["toUserId"] = toUser.ID
	}
	
	// 礼物信息
	if gift != nil {
		result["giftName"] = gift.Name
		result["giftId"] = gift.ID
		result["diamondCount"] = gift.DiamondCount
	} else if giftId != "" {
		result["giftId"] = giftId
	}
	
	// 数量计算 (优先级: groupCount > repeatCount > comboCount)
	if groupCount != "" && groupCount != "0" {
		result["giftCount"] = groupCount
	} else if repeatCount != "" && repeatCount != "0" {
		result["giftCount"] = repeatCount
	} else if comboCount != "" && comboCount != "0" {
		result["giftCount"] = comboCount
	} else {
		result["giftCount"] = "1"
	}
	
	// 连击信息
	if comboCount != "" && comboCount != "0" {
		result["comboCount"] = comboCount
	}
	if repeatEnd != "" {
		result["repeatEnd"] = repeatEnd
		result["isComboEnd"] = repeatEnd == "1"
	}
	
	// 其他信息
	if totalCoin != "" && totalCoin != "0" {
		result["totalCoin"] = totalCoin
	}
	if timestamp != "" {
		result["giftTimestamp"] = timestamp
	}
	if logId != "" {
		result["logId"] = logId
	}
	if sendType > 0 {
		result["sendType"] = sendType
	}
	if publicArea > 0 {
		result["publicArea"] = publicArea
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
		case 1: // id (用户ID)
			user.ID, _ = bb.ReadVarint64(false)
		case 2: // shortId (短ID)
			user.ShortID, _ = bb.ReadVarint64(false)
		case 3: // nickname (昵称)
			length, _ := bb.ReadVarint32()
			user.Nickname, _ = bb.ReadString(int(length))
		case 4: // gender (性别: 1=男, 2=女)
			user.Gender, _ = bb.ReadVarint32()
		case 5: // signature (签名)
			skipLengthDelimitedField(bb)
		case 6: // level (等级)
			user.Level, _ = bb.ReadVarint32()
		case 7: // birthday (生日)
			bb.ReadVarint64(false)
		case 8: // telephone (电话)
			skipLengthDelimitedField(bb)
		case 9: // avatarThumb (头像缩略图)
			skipLengthDelimitedField(bb)
		case 10: // avatarMedium (头像中图)
			skipLengthDelimitedField(bb)
		case 11: // avatarLarge (头像大图)
			skipLengthDelimitedField(bb)
		case 12: // verified (是否认证)
			bb.ReadByte()
		case 13: // experience (经验值)
			bb.ReadVarint32()
		case 14: // city (城市)
			skipLengthDelimitedField(bb)
		case 15: // status (状态)
			bb.ReadVarint32()
		case 16: // createTime (创建时间)
			bb.ReadVarint64(false)
		case 17: // modifyTime (修改时间)
			bb.ReadVarint64(false)
		case 18: // secret (隐私设置)
			bb.ReadVarint32()
		case 19: // shareQrcodeUri (分享二维码)
			skipLengthDelimitedField(bb)
		case 20: // incomeSharePercent (收益分成比例)
			bb.ReadVarint32()
		case 21: // badgeImageList (徽章列表)
			skipLengthDelimitedField(bb)
		case 22: // followInfo (关注信息)
			skipLengthDelimitedField(bb)
		case 23: // payGrade (付费等级)
			skipLengthDelimitedField(bb)
		case 24: // fansClub (粉丝团)
			skipLengthDelimitedField(bb)
		case 25: // border (边框)
			skipLengthDelimitedField(bb)
		case 26: // specialId (特殊ID)
			skipLengthDelimitedField(bb)
		case 27: // avatarBorder (头像边框)
			skipLengthDelimitedField(bb)
		case 28: // medal (勋章)
			skipLengthDelimitedField(bb)
		case 29: // realTimeIcons (实时图标)
			skipLengthDelimitedField(bb)
		case 30: // newRealTimeIcons (新实时图标)
			skipLengthDelimitedField(bb)
		case 31: // topVip (顶级VIP)
			skipLengthDelimitedField(bb)
		case 32: // userAttr (用户属性)
			skipLengthDelimitedField(bb)
		case 33: // ownRoom (自己的直播间)
			skipLengthDelimitedField(bb)
		case 34: // payScores (付费积分)
			bb.ReadVarint64(false)
		case 35: // ticketCount (票数)
			bb.ReadVarint64(false)
		case 36: // anchorInfo (主播信息)
			skipLengthDelimitedField(bb)
		case 37: // linkMicStats (连麦统计)
			skipLengthDelimitedField(bb)
		case 38: // displayId (显示ID)
			skipLengthDelimitedField(bb)
		case 39: // withCommercePermission (商业权限)
			bb.ReadByte()
		case 40: // withFusionShopEntry (融合商店入口)
			bb.ReadByte()
		case 41: // verifyStatus (认证状态)
			bb.ReadVarint32()
		case 42: // enterpriseVerifyReason (企业认证原因)
			skipLengthDelimitedField(bb)
		case 43: // needsToGetToast (是否需要获取提示)
			bb.ReadByte()
		case 44: // isBlock (是否被拉黑)
			bb.ReadByte()
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
		case 1: // image (礼物图片)
			skipLengthDelimitedField(bb)
		case 2: // describe (描述)
			skipLengthDelimitedField(bb)
		case 3: // notify (通知)
			bb.ReadByte()
		case 4: // duration (持续时间)
			bb.ReadVarint64(false)
		case 5: // id (礼物ID)
			gift.ID, _ = bb.ReadVarint64(false)
		case 6: // forLinkMic (是否连麦礼物)
			bb.ReadByte()
		case 7: // type (类型)
			bb.ReadVarint32()
		case 8: // diamondCount (抖币价值)
			gift.DiamondCount, _ = bb.ReadVarint32()
		case 9: // isDisplayedOnPanel (是否显示在面板)
			bb.ReadByte()
		case 10: // primaryEffectId (主效果ID)
			bb.ReadVarint64(false)
		case 11: // giftLabelIcon (礼物标签图标)
			skipLengthDelimitedField(bb)
		case 12: // name (礼物名称) - 注意：可能是field 12或16
			length, _ := bb.ReadVarint32()
			if gift.Name == "" {
				gift.Name, _ = bb.ReadString(int(length))
			} else {
				bb.ReadString(int(length))
			}
		case 13: // region (区域)
			skipLengthDelimitedField(bb)
		case 14: // manual (手动)
			bb.ReadByte()
		case 15: // forFansclub (粉丝团礼物)
			bb.ReadByte()
		case 16: // name (礼物名称) - 注意：可能是field 12或16
			length, _ := bb.ReadVarint32()
			if gift.Name == "" {
				gift.Name, _ = bb.ReadString(int(length))
			} else {
				bb.ReadString(int(length))
			}
		case 17: // goldEffect (金币效果)
			skipLengthDelimitedField(bb)
		case 18: // colorInfos (颜色信息)
			skipLengthDelimitedField(bb)
		case 19: // eventName (事件名称)
			skipLengthDelimitedField(bb)
		case 20: // landingPageUrl (落地页URL)
			skipLengthDelimitedField(bb)
		case 21: // stayTime (停留时间)
			bb.ReadVarint32()
		case 22: // icon (图标)
			skipLengthDelimitedField(bb)
		case 23: // actionType (动作类型)
			bb.ReadVarint32()
		case 24: // anchorEventInfo (主播事件信息)
			skipLengthDelimitedField(bb)
		case 25: // userEventInfo (用户事件信息)
			skipLengthDelimitedField(bb)
		case 26: // giftPanelBanner (礼物面板横幅)
			skipLengthDelimitedField(bb)
		case 27: // fullScreenEffect (全屏效果)
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

// parseRoomMessageImproved 直播间消息解析
func parseRoomMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var content string
	var roomStatus int32
	
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
		case 2: // content
			length, _ := bb.ReadVarint32()
			content, _ = bb.ReadString(int(length))
		case 3: // roomStatus (1=开播, 2=下播)
			roomStatus, _ = bb.ReadVarint32()
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "直播间消息"
	result["content"] = content
	result["roomStatus"] = roomStatus
	
	if roomStatus == 1 {
		result["statusText"] = "开播"
	} else if roomStatus == 2 {
		result["statusText"] = "下播"
	}
	
	return true, nil
}

// parseMatchAgainstScoreMessageImproved PK消息解析
func parseMatchAgainstScoreMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var battleStatus int32
	var matchScore string
	var ownScore string
	var againstScore string
	
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
		case 2: // matchScore
			matchScore, _ = bb.ReadVarint64(false)
		case 3: // ownScore
			ownScore, _ = bb.ReadVarint64(false)
		case 4: // againstScore
			againstScore, _ = bb.ReadVarint64(false)
		case 5: // battleStatus (1=进行中, 2=结束)
			battleStatus, _ = bb.ReadVarint32()
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "PK消息"
	result["matchScore"] = matchScore
	result["ownScore"] = ownScore
	result["againstScore"] = againstScore
	result["battleStatus"] = battleStatus
	
	return true, nil
}

// parseRankUpdateMessageImproved 榜单更新消息解析
func parseRankUpdateMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var rankType int32
	
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
		case 2: // rankType
			rankType, _ = bb.ReadVarint32()
		case 3: // rankList (repeated)
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "榜单更新"
	result["rankType"] = rankType
	
	return true, nil
}

// parseLinkMicMessageImproved 连麦消息解析
func parseLinkMicMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var scene int32
	var micStatus int32
	
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
		case 2: // scene (1=连麦, 2=PK)
			scene, _ = bb.ReadVarint32()
		case 3: // micStatus
			micStatus, _ = bb.ReadVarint32()
		case 4: // anchorList
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "连麦消息"
	result["scene"] = scene
	result["micStatus"] = micStatus
	
	return true, nil
}

// parseLinkMicBattleImproved 连麦PK消息解析
func parseLinkMicBattleImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var battleStatus int32
	var battleDuration int32
	
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
		case 2: // battleStatus
			battleStatus, _ = bb.ReadVarint32()
		case 3: // battleDuration
			battleDuration, _ = bb.ReadVarint32()
		case 4: // battleItems
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "连麦PK"
	result["battleStatus"] = battleStatus
	result["battleDuration"] = battleDuration
	
	return true, nil
}

// parseLinkMicArmiesImproved 连麦军团消息解析
func parseLinkMicArmiesImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
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
		case 2: // armies
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "连麦军团"
	
	return true, nil
}

// parseInRoomBannerMessageImproved 房间横幅消息解析
func parseInRoomBannerMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
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
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // content
			length, _ := bb.ReadVarint32()
			content, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "房间横幅"
	result["content"] = content
	
	return true, nil
}

// parseProductChangeMessageImproved 商品变化消息解析
func parseProductChangeMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var updateType int32
	var productId string
	
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
		case 2: // updateType (1=上架, 2=下架)
			updateType, _ = bb.ReadVarint32()
		case 3: // productId
			length, _ := bb.ReadVarint32()
			productId, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "商品变化"
	result["updateType"] = updateType
	result["productId"] = productId
	
	return true, nil
}

// parseCommonTextMessageImproved 通用文本消息解析
func parseCommonTextMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
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
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // user
			oldLimit, _ := bb.PushTemporaryLength()
			user, _ = DecodeUserImproved(bb)
			bb.limit = oldLimit
		case 3: // content
			length, _ := bb.ReadVarint32()
			content, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "通用文本消息"
	if user != nil {
		result["user"] = user.Nickname
		result["userId"] = user.ID
	}
	result["content"] = content
	
	return true, nil
}

// parseBarrageMessageImproved 弹幕消息解析
func parseBarrageMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
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
		case 1: // common
			skipLengthDelimitedField(bb)
		case 2: // content
			length, _ := bb.ReadVarint32()
			content, _ = bb.ReadString(int(length))
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "弹幕消息"
	result["content"] = content
	
	return true, nil
}

// parseRoomRankMessageImproved 房间排行榜消息解析
func parseRoomRankMessageImproved(payload []byte, result map[string]interface{}) (bool, error) {
	bb := NewByteBuffer(payload)
	
	var rankType int32
	
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
		case 2: // rankType
			rankType, _ = bb.ReadVarint32()
		case 3: // rankItems
			skipLengthDelimitedField(bb)
		default:
			bb.SkipUnknownField(wireType)
		}
	}
	
	result["messageType"] = "房间排行榜"
	result["rankType"] = rankType
	
	return true, nil
}
