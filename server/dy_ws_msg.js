/**
 * 抖音直播 WebSocket 消息解析器
 * 完全按照 dycast 项目实现: https://github.com/skmcj/dycast
 * 
 * 使用 Protobuf + GZIP 压缩格式解析抖音直播消息
 */

const pako = require('pako');

// ============ ByteBuffer 实现 (来自 dycast/model.ts) ============

function createByteBuffer(bytes) {
  return {
    bytes: bytes || new Uint8Array(1024),
    offset: 0,
    limit: bytes ? bytes.length : 0
  };
}

function advance(bb, count) {
  const offset = bb.offset;
  if (offset + count > bb.limit) {
    throw new Error('Read past limit');
  }
  bb.offset += count;
  return offset;
}

function readByte(bb) {
  return bb.bytes[advance(bb, 1)];
}

function readBytes(bb, count) {
  const offset = advance(bb, count);
  return bb.bytes.subarray(offset, offset + count);
}

function readVarint32(bb) {
  let c = 0;
  let value = 0;
  let b;
  do {
    b = readByte(bb);
    if (c < 32) value |= (b & 0x7f) << c;
    c += 7;
  } while (b & 0x80);
  return value;
}

function readVarint64(bb, unsigned) {
  let part0 = 0;
  let part1 = 0;
  let part2 = 0;
  let b;

  b = readByte(bb);
  part0 = b & 0x7f;
  if (b & 0x80) {
    b = readByte(bb);
    part0 |= (b & 0x7f) << 7;
    if (b & 0x80) {
      b = readByte(bb);
      part0 |= (b & 0x7f) << 14;
      if (b & 0x80) {
        b = readByte(bb);
        part0 |= (b & 0x7f) << 21;
        if (b & 0x80) {
          b = readByte(bb);
          part1 = b & 0x7f;
          if (b & 0x80) {
            b = readByte(bb);
            part1 |= (b & 0x7f) << 7;
            if (b & 0x80) {
              b = readByte(bb);
              part1 |= (b & 0x7f) << 14;
              if (b & 0x80) {
                b = readByte(bb);
                part1 |= (b & 0x7f) << 21;
                if (b & 0x80) {
                  b = readByte(bb);
                  part2 = b & 0x7f;
                  if (b & 0x80) {
                    b = readByte(bb);
                    part2 |= (b & 0x7f) << 7;
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  const low = part0 | (part1 << 28);
  const high = (part1 >>> 4) | (part2 << 24);
  
  if (high === 0) {
    return String(low >>> 0);
  }
  
  return String(high * 4294967296 + (low >>> 0));
}

function readString(bb, count) {
  const offset = advance(bb, count);
  const bytes = bb.bytes;
  const fromCharCode = String.fromCharCode;
  const invalid = '\uFFFD';
  let text = '';

  for (let i = 0; i < count; i++) {
    let c1 = bytes[i + offset], c2, c3, c4, c;

    if ((c1 & 0x80) === 0) {
      text += fromCharCode(c1);
    }
    else if ((c1 & 0xe0) === 0xc0) {
      if (i + 1 >= count) text += invalid;
      else {
        c2 = bytes[i + offset + 1];
        if ((c2 & 0xc0) !== 0x80) text += invalid;
        else {
          c = ((c1 & 0x1f) << 6) | (c2 & 0x3f);
          if (c < 0x80) text += invalid;
          else {
            text += fromCharCode(c);
            i++;
          }
        }
      }
    }
    else if ((c1 & 0xf0) === 0xe0) {
      if (i + 2 >= count) text += invalid;
      else {
        c2 = bytes[i + offset + 1];
        c3 = bytes[i + offset + 2];
        if (((c2 | (c3 << 8)) & 0xc0c0) !== 0x8080) text += invalid;
        else {
          c = ((c1 & 0x0f) << 12) | ((c2 & 0x3f) << 6) | (c3 & 0x3f);
          if (c < 0x0800 || (c >= 0xd800 && c <= 0xdfff)) text += invalid;
          else {
            text += fromCharCode(c);
            i += 2;
          }
        }
      }
    }
    else if ((c1 & 0xf8) === 0xf0) {
      if (i + 3 >= count) text += invalid;
      else {
        c2 = bytes[i + offset + 1];
        c3 = bytes[i + offset + 2];
        c4 = bytes[i + offset + 3];
        if (((c2 | (c3 << 8) | (c4 << 16)) & 0xc0c0c0) !== 0x808080) text += invalid;
        else {
          c = ((c1 & 0x07) << 0x12) | ((c2 & 0x3f) << 0x0c) | ((c3 & 0x3f) << 0x06) | (c4 & 0x3f);
          if (c < 0x10000 || c > 0x10ffff) text += invalid;
          else {
            c -= 0x10000;
            text += fromCharCode((c >> 10) + 0xd800, (c & 0x3ff) + 0xdc00);
            i += 3;
          }
        }
      }
    } else text += invalid;
  }

  return text;
}

function isAtEnd(bb) {
  return bb.offset >= bb.limit;
}

function skipUnknownField(bb, type) {
  switch (type) {
    case 0: // Varint
      while (readByte(bb) & 0x80);
      break;
    case 1: // 64-bit
      advance(bb, 8);
      break;
    case 2: // Length-delimited
      advance(bb, readVarint32(bb));
      break;
    case 3: // Start group (deprecated)
      while (!isAtEnd(bb)) {
        const tag = readVarint32(bb);
        const wireType = tag & 7;
        
        if (wireType === 4) {
          break;
        }
        
        skipUnknownField(bb, wireType);
      }
      break;
    case 4: // End group (deprecated)
      break;
    case 5: // 32-bit
      advance(bb, 4);
      break;
    default:
      throw new Error('Invalid wire type: ' + type);
  }
}

function pushTemporaryLength(bb) {
  const length = readVarint32(bb);
  const limit = bb.limit;
  bb.limit = bb.offset + length;
  return limit;
}

// ============ Protobuf 解码函数 (来自 dycast/model.ts) ============

function decodePushFrame(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 1: // seqId
        message.seqId = readVarint64(bb, true);
        break;
      case 2: // logId
        message.logId = readVarint64(bb, true);
        break;
      case 3: // service
        message.service = readVarint64(bb, true);
        break;
      case 4: // method
        message.method = readVarint64(bb, true);
        break;
      case 5: // headersList (map<string, string>)
        {
          const outerLimit = pushTemporaryLength(bb);
          let key, value;
          while (!isAtEnd(bb)) {
            const tag2 = readVarint32(bb);
            const fieldNumber2 = tag2 >>> 3;
            if (fieldNumber2 === 0) break;
            if (fieldNumber2 === 1) {
              key = readString(bb, readVarint32(bb));
            } else if (fieldNumber2 === 2) {
              value = readString(bb, readVarint32(bb));
            } else {
              skipUnknownField(bb, tag2 & 7);
            }
          }
          if (key !== undefined && value !== undefined) {
            if (!message.headersList) message.headersList = {};
            message.headersList[key] = value;
          }
          bb.limit = outerLimit;
        }
        break;
      case 6: // payloadEncoding
        message.payloadEncoding = readString(bb, readVarint32(bb));
        break;
      case 7: // payloadType
        message.payloadType = readString(bb, readVarint32(bb));
        break;
      case 8: // payload (bytes)
        message.payload = readBytes(bb, readVarint32(bb));
        break;
      case 9: // lodIdNew
        message.lodIdNew = readString(bb, readVarint32(bb));
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeResponse(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 1: // messages (repeated Message)
        {
          const limit = pushTemporaryLength(bb);
          if (!message.messages) message.messages = [];
          message.messages.push(decodeMessage(bb));
          bb.limit = limit;
        }
        break;
      case 2: // cursor
        message.cursor = readString(bb, readVarint32(bb));
        break;
      case 3: // fetchInterval
        message.fetchInterval = readVarint64(bb, false);
        break;
      case 4: // now
        message.now = readVarint64(bb, false);
        break;
      case 5: // internalExt
        message.internalExt = readString(bb, readVarint32(bb));
        break;
      case 6: // fetchType
        message.fetchType = readVarint32(bb);
        break;
      case 8: // heartbeatDuration
        message.heartbeatDuration = readVarint64(bb, false);
        break;
      case 9: // needAck
        message.needAck = !!readByte(bb);
        break;
      case 10: // pushServer
        message.pushServer = readString(bb, readVarint32(bb));
        break;
      case 11: // liveCursor
        message.liveCursor = readString(bb, readVarint32(bb));
        break;
      case 12: // historyNoMore
        message.historyNoMore = !!readByte(bb);
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeMessage(bb) {
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 1: // method (string)
        message.method = readString(bb, readVarint32(bb));
        break;
      case 2: // payload (bytes)
        message.payload = readBytes(bb, readVarint32(bb));
        break;
      case 3: // msgId
        message.msgId = readVarint64(bb, false);
        break;
      case 4: // msgType
        message.msgType = readVarint32(bb);
        break;
      case 5: // offset
        message.offset = readVarint64(bb, false);
        break;
      case 6: // needWrdsStore
        message.needWrdsStore = !!readByte(bb);
        break;
      case 7: // wrdsVersion
        message.wrdsVersion = readVarint64(bb, false);
        break;
      case 8: // wrdsSubKey
        message.wrdsSubKey = readString(bb, readVarint32(bb));
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

// ============ 消息解码函数 (简化版，只提取关键字段) ============

function decodeUser(binary) {
  const bb = createByteBuffer(binary);
  const user = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 1: // id
        user.id = readVarint64(bb, false);
        break;
      case 2: // shortId
        user.shortId = readVarint64(bb, false);
        break;
      case 3: // nickname
        user.nickname = readString(bb, readVarint32(bb));
        break;
      case 6: // level
        user.level = readVarint32(bb);
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return user;
}

function decodeChatMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 2: // user
        {
          const limit = pushTemporaryLength(bb);
          message.user = decodeUser(bb);
          bb.limit = limit;
        }
        break;
      case 3: // content
        message.content = readString(bb, readVarint32(bb));
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeGiftMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 2: // giftId
        message.giftId = readVarint64(bb, false);
        break;
      case 5: // repeatCount
        message.repeatCount = readVarint64(bb, false);
        break;
      case 6: // comboCount
        message.comboCount = readVarint64(bb, false);
        break;
      case 7: // user
        {
          const limit = pushTemporaryLength(bb);
          message.user = decodeUser(bb);
          bb.limit = limit;
        }
        break;
      case 9: // gift (GiftStruct)
        {
          const limit = pushTemporaryLength(bb);
          message.gift = decodeGiftStruct(bb);
          bb.limit = limit;
        }
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeGiftStruct(bb) {
  const gift = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 1: // giftId
        gift.id = readVarint64(bb, false);
        break;
      case 2: // name
        gift.name = readString(bb, readVarint32(bb));
        break;
      case 10: // diamondCount
        gift.diamondCount = readVarint32(bb);
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return gift;
}

function decodeLikeMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 2: // user
        {
          const limit = pushTemporaryLength(bb);
          message.user = decodeUser(bb);
          bb.limit = limit;
        }
        break;
      case 3: // count
        message.count = readVarint64(bb, false);
        break;
      case 4: // total
        message.total = readVarint64(bb, false);
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeMemberMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 2: // user
        {
          const limit = pushTemporaryLength(bb);
          message.user = decodeUser(bb);
          bb.limit = limit;
        }
        break;
      case 3: // memberCount
        message.memberCount = readVarint64(bb, false);
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeSocialMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 2: // user
        {
          const limit = pushTemporaryLength(bb);
          message.user = decodeUser(bb);
          bb.limit = limit;
        }
        break;
      case 3: // followCount
        message.followCount = readVarint64(bb, false);
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeRoomUserSeqMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 2: // total (在线人数)
        message.total = readVarint64(bb, false);
        break;
      case 3: // totalUser (总观看人数)
        message.totalUser = readVarint64(bb, false);
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

function decodeRoomStatsMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const wireType = tag & 7;
    const fieldNumber = tag >>> 3;

    if (fieldNumber === 0) break;

    switch (fieldNumber) {
      case 2: // displayShort (简短显示)
        message.displayShort = readString(bb, readVarint32(bb));
        break;
      case 3: // displayMiddle (中等显示)
        message.displayMiddle = readString(bb, readVarint32(bb));
        break;
      case 4: // displayLong (完整显示)
        message.displayLong = readString(bb, readVarint32(bb));
        break;
      default:
        skipUnknownField(bb, wireType);
    }
  }

  return message;
}

// ============ 抖音消息解析器 ============

class DouyinWSMessageParser {
  constructor() {
    this.messageTypes = {
      'WebcastChatMessage': '聊天消息',
      'WebcastGiftMessage': '礼物消息',
      'WebcastLikeMessage': '点赞消息',
      'WebcastMemberMessage': '进入直播间',
      'WebcastSocialMessage': '关注消息',
      'WebcastRoomUserSeqMessage': '在线人数',
      'WebcastFansclubMessage': '粉丝团消息',
      'WebcastControlMessage': '直播间控制',
      'WebcastEmojiChatMessage': '表情消息',
      'WebcastRoomStatsMessage': '直播间统计'
    };

    this.statistics = {
      totalMessages: 0,
      chatCount: 0,
      giftCount: 0,
      likeCount: 0,
      memberCount: 0,
      onlineUsers: 0
    };
  }

  isDouyinLiveWS(url) {
    if (!url) return false;
    return url.includes('webcast') && url.includes('douyin.com');
  }

  async parseMessage(payloadData, url = '') {
    if (!payloadData) return null;
    
    try {
      let buffer;
      if (typeof payloadData === 'string') {
        buffer = Buffer.from(payloadData, 'base64');
      } else if (Buffer.isBuffer(payloadData)) {
        buffer = payloadData;
      } else {
        buffer = Buffer.from(payloadData);
      }

      const uint8Array = new Uint8Array(buffer);

      // 1. 解析 PushFrame
      const pushFrame = decodePushFrame(uint8Array);
      
      if (!pushFrame || !pushFrame.payload) {
        return null;
      }

      // 2. 检查是否需要 GZIP 解压
      let payload = pushFrame.payload;
      const compressType = pushFrame.headersList?.['compress_type'];
      
      if (compressType === 'gzip') {
        try {
          payload = pako.ungzip(payload);
        } catch (e) {
          console.error('[Douyin] GZIP 解压失败:', e.message);
          return null;
        }
      }

      // 3. 解析 Response
      const response = decodeResponse(payload);
      
      if (!response || !response.messages || response.messages.length === 0) {
        return null;
      }

      // 4. 解析每条消息 - 按照 dycast 的 _dealMessage 逻辑
      const results = [];
      for (const msg of response.messages) {
        if (msg.method && msg.payload) {
          const parsed = this.parseMessagePayload(msg.method, msg.payload);
          if (parsed) {
            results.push(parsed);
          }
        }
      }

      return results.length > 0 ? results : null;
    } catch (e) {
      console.error('[Douyin] 解析消息失败:', e.message);
      return null;
    }
  }

  parseMessagePayload(method, payload) {
    this.statistics.totalMessages++;

    const result = {
      type: 'douyin_live',
      messageType: this.messageTypes[method] || method,
      method: method,
      timestamp: new Date().toISOString(),
      parsed: true
    };

    try {
      // 按照 dycast 的 switch 逻辑解析不同类型的消息
      switch (method) {
        case 'WebcastChatMessage': {
          const message = decodeChatMessage(payload);
          this.statistics.chatCount++;
          return {
            ...result,
            messageType: '聊天消息',
            user: message.user?.nickname || '匿名用户',
            userId: message.user?.id,
            content: message.content || '',
            level: message.user?.level
          };
        }

        case 'WebcastGiftMessage': {
          const message = decodeGiftMessage(payload);
          this.statistics.giftCount++;
          return {
            ...result,
            messageType: '礼物消息',
            user: message.user?.nickname || '匿名用户',
            userId: message.user?.id,
            giftName: message.gift?.name || '未知礼物',
            giftId: message.gift?.id,
            giftCount: message.repeatCount || message.comboCount || '1',
            diamondCount: message.gift?.diamondCount
          };
        }

        case 'WebcastLikeMessage': {
          const message = decodeLikeMessage(payload);
          this.statistics.likeCount++;
          return {
            ...result,
            messageType: '点赞消息',
            user: message.user?.nickname || '匿名用户',
            userId: message.user?.id,
            count: message.count || '1',
            total: message.total
          };
        }

        case 'WebcastMemberMessage': {
          const message = decodeMemberMessage(payload);
          this.statistics.memberCount++;
          return {
            ...result,
            messageType: '进入直播间',
            user: message.user?.nickname || '匿名用户',
            userId: message.user?.id,
            memberCount: message.memberCount
          };
        }

        case 'WebcastSocialMessage': {
          const message = decodeSocialMessage(payload);
          return {
            ...result,
            messageType: '关注消息',
            user: message.user?.nickname || '匿名用户',
            userId: message.user?.id,
            followCount: message.followCount
          };
        }

        case 'WebcastRoomUserSeqMessage': {
          const message = decodeRoomUserSeqMessage(payload);
          const total = message.total || '0';
          this.statistics.onlineUsers = parseInt(total) || 0;
          return {
            ...result,
            messageType: '在线人数',
            total: total,
            totalUser: message.totalUser || '0'
          };
        }

        case 'WebcastRoomStatsMessage': {
          const message = decodeRoomStatsMessage(payload);
          const displayMiddle = message.displayMiddle || '0';
          this.statistics.onlineUsers = parseInt(displayMiddle) || 0;
          return {
            ...result,
            messageType: '直播间统计',
            displayShort: message.displayShort,
            displayMiddle: displayMiddle,
            displayLong: message.displayLong
          };
        }

        default:
          // 其他消息类型，返回基本信息
          return result;
      }
    } catch (e) {
      console.error(`[Douyin] 解析 ${method} 失败:`, e.message);
      return result;
    }
  }

  formatMessage(parsedMessages) {
    if (!parsedMessages) return null;

    if (Array.isArray(parsedMessages)) {
      return parsedMessages.map(msg => this.formatSingleMessage(msg)).filter(Boolean).join('\n\n');
    }

    return this.formatSingleMessage(parsedMessages);
  }

  formatSingleMessage(parsedMessage) {
    if (!parsedMessage) return null;

    const lines = [];
    lines.push(`╔${'═'.repeat(78)}╗`);
    lines.push(`║ 🎬 抖音直播消息`);
    lines.push(`╠${'═'.repeat(78)}╣`);
    lines.push(`║ 消息类型: ${parsedMessage.messageType}`);
    lines.push(`║ 时间: ${parsedMessage.timestamp}`);

    switch (parsedMessage.messageType) {
      case '聊天消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        if (parsedMessage.level) {
          lines.push(`║ 等级: ${parsedMessage.level}`);
        }
        if (parsedMessage.content) {
          lines.push(`║ 内容: ${parsedMessage.content}`);
        }
        break;

      case '礼物消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        lines.push(`║ 礼物: ${parsedMessage.giftName}`);
        if (parsedMessage.giftCount) {
          lines.push(`║ 数量: ${parsedMessage.giftCount}`);
        }
        if (parsedMessage.diamondCount) {
          lines.push(`║ 价值: ${parsedMessage.diamondCount} 💎`);
        }
        break;

      case '点赞消息':
        lines.push(`║ 用户: ${parsedMessage.user} ❤️`);
        if (parsedMessage.count) {
          lines.push(`║ 点赞数: ${parsedMessage.count}`);
        }
        break;

      case '进入直播间':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        if (parsedMessage.memberCount) {
          lines.push(`║ 当前人数: ${parsedMessage.memberCount}`);
        }
        break;

      case '在线人数':
        lines.push(`║ 在线人数: ${parsedMessage.total} 👥`);
        if (parsedMessage.totalUser) {
          lines.push(`║ 累计观看: ${parsedMessage.totalUser}`);
        }
        break;

      case '直播间统计':
        if (parsedMessage.displayMiddle) {
          lines.push(`║ 在线观众: ${parsedMessage.displayMiddle} 👥`);
        }
        break;

      case '关注消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        lines.push(`║ 动作: 关注了主播`);
        break;

      default:
        lines.push(`║ 方法: ${parsedMessage.method}`);
    }

    lines.push(`╚${'═'.repeat(78)}╝`);
    return lines.join('\n');
  }

  getStatistics() {
    return {
      ...this.statistics,
      timestamp: new Date().toISOString()
    };
  }

  resetStatistics() {
    this.statistics = {
      totalMessages: 0,
      chatCount: 0,
      giftCount: 0,
      likeCount: 0,
      memberCount: 0,
      onlineUsers: 0
    };
  }

  formatStatistics() {
    const stats = this.getStatistics();
    const lines = [];
    lines.push(`╔${'═'.repeat(78)}╗`);
    lines.push(`║ 📊 抖音直播统计`);
    lines.push(`╠${'═'.repeat(78)}╣`);
    lines.push(`║ 总消息数: ${stats.totalMessages}`);
    lines.push(`║ 聊天消息: ${stats.chatCount}`);
    lines.push(`║ 礼物消息: ${stats.giftCount}`);
    lines.push(`║ 点赞消息: ${stats.likeCount}`);
    lines.push(`║ 进入直播间: ${stats.memberCount}`);
    lines.push(`║ 当前在线: ${stats.onlineUsers} 👥`);
    lines.push(`║ 更新时间: ${stats.timestamp}`);
    lines.push(`╚${'═'.repeat(78)}╝`);
    return lines.join('\n');
  }
}

// 导出单例
module.exports = new DouyinWSMessageParser();
