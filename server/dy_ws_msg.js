/**
 * 抖音直播 WebSocket 消息解析器
 * 参考: https://github.com/skmcj/dycast
 * 
 * 解析来自 wss://webcast100-ws-web-hl.douyin.com 的消息
 * 支持 Protobuf + GZIP 压缩格式
 */

const zlib = require('zlib');
const { promisify } = require('util');

const gunzip = promisify(zlib.gunzip);
const inflate = promisify(zlib.inflate);

class DouyinWSMessageParser {
  constructor() {
    // 消息类型映射
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
      'WebcastRoomStatsMessage': '直播间统计',
      'WebcastRoomMessage': '直播间消息',
      'WebcastLinkMicBattle': '连麦PK',
      'WebcastLinkMicArmies': '连麦军团'
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

  /**
   * 检测是否为抖音直播 WebSocket URL
   */
  isDouyinLiveWS(url) {
    if (!url) return false;
    return url.includes('webcast') && 
           url.includes('douyin.com');
  }

  /**
   * 解析 WebSocket 消息
   * @param {string} payloadData - WebSocket消息内容（可能是base64）
   * @param {string} url - WebSocket URL
   * @returns {Object|null} 解析后的消息对象
   */
  async parseMessage(payloadData, url = '') {
    if (!payloadData) return null;
    
    try {
      // 将payload转为Buffer
      let buffer;
      if (typeof payloadData === 'string') {
        // 尝试base64解码
        try {
          buffer = Buffer.from(payloadData, 'base64');
        } catch (e) {
          buffer = Buffer.from(payloadData);
        }
      } else {
        buffer = Buffer.from(payloadData);
      }

      // 解析外层 PushFrame
      const pushFrame = this.parsePushFrame(buffer);
      if (!pushFrame) {
        return null;
      }

      // 解析内层 Response
      const response = await this.parseResponse(pushFrame);
      if (!response) {
        return null;
      }

      // 解析具体消息
      return this.parseMessages(response);
    } catch (e) {
      console.error('解析消息失败:', e.message);
      return null;
    }
  }

  /**
   * 解析 PushFrame（外层结构）
   */
  parsePushFrame(buffer) {
    try {
      let offset = 0;
      const frame = {};

      while (offset < buffer.length) {
        // 读取字段类型和编号
        const tag = buffer[offset++];
        if (!tag) break;

        const wireType = tag & 0x07;
        const fieldNumber = tag >> 3;

        if (wireType === 2) { // Length-delimited
          const length = this.readVarint(buffer, offset);
          offset += this.varintSize(length);

          const value = buffer.slice(offset, offset + length);
          offset += length;

          // 字段映射
          if (fieldNumber === 1) {
            frame.logId = value.readBigUInt64LE ? value.readBigUInt64LE(0) : 0;
          } else if (fieldNumber === 2) {
            frame.service = value.readUInt32LE(0);
          } else if (fieldNumber === 3) {
            frame.method = value.toString('utf8');
          } else if (fieldNumber === 4) {
            // 这是重要的 headers_list
            frame.headersList = this.parseHeadersList(value);
          } else if (fieldNumber === 5) {
            // 这是 payload（压缩的Response）
            frame.payloadBinary = value;
          }
        } else if (wireType === 0) { // Varint
          const value = this.readVarint(buffer, offset);
          offset += this.varintSize(value);
          
          if (fieldNumber === 2) {
            frame.service = value;
          }
        } else {
          // 跳过未知字段
          break;
        }
      }

      return frame;
    } catch (e) {
      console.error('解析PushFrame失败:', e.message);
      return null;
    }
  }

  /**
   * 解析 headers_list
   */
  parseHeadersList(buffer) {
    const headers = {};
    let offset = 0;

    while (offset < buffer.length) {
      const tag = buffer[offset++];
      if (!tag) break;

      const wireType = tag & 0x07;
      const fieldNumber = tag >> 3;

      if (wireType === 2 && fieldNumber === 3) {
        const length = this.readVarint(buffer, offset);
        offset += this.varintSize(length);

        const headerData = buffer.slice(offset, offset + length);
        offset += length;

        // 解析单个header
        const header = this.parseHeader(headerData);
        if (header && header.key) {
          headers[header.key] = header.value;
        }
      } else {
        break;
      }
    }

    return headers;
  }

  /**
   * 解析单个 header
   */
  parseHeader(buffer) {
    const header = {};
    let offset = 0;

    while (offset < buffer.length) {
      const tag = buffer[offset++];
      if (!tag) break;

      const wireType = tag & 0x07;
      const fieldNumber = tag >> 3;

      if (wireType === 2) {
        const length = this.readVarint(buffer, offset);
        offset += this.varintSize(length);

        const value = buffer.slice(offset, offset + length);
        offset += length;

        if (fieldNumber === 1) {
          header.key = value.toString('utf8');
        } else if (fieldNumber === 2) {
          header.value = value.toString('utf8');
        }
      } else {
        break;
      }
    }

    return header;
  }

  /**
   * 解析 Response（解压后的内层结构）
   */
  async parseResponse(frame) {
    try {
      if (!frame.payloadBinary) return null;

      // 检查是否需要解压
      const compressType = frame.headersList?.['compress_type'];
      let payload = frame.payloadBinary;

      if (compressType === 'gzip') {
        try {
          payload = await gunzip(payload);
        } catch (e) {
          console.error('GZIP解压失败:', e.message);
          return null;
        }
      }

      // 解析Response结构
      const response = {};
      let offset = 0;

      while (offset < payload.length) {
        const tag = payload[offset++];
        if (!tag) break;

        const wireType = tag & 0x07;
        const fieldNumber = tag >> 3;

        if (wireType === 2) {
          const length = this.readVarint(payload, offset);
          offset += this.varintSize(length);

          const value = payload.slice(offset, offset + length);
          offset += length;

          if (fieldNumber === 1) {
            // messages_list
            if (!response.messagesList) {
              response.messagesList = [];
            }
            response.messagesList.push(value);
          }
        } else if (wireType === 0) {
          const value = this.readVarint(payload, offset);
          offset += this.varintSize(value);
        } else {
          break;
        }
      }

      return response;
    } catch (e) {
      console.error('解析Response失败:', e.message);
      return null;
    }
  }

  /**
   * 解析具体消息列表
   */
  parseMessages(response) {
    if (!response.messagesList || response.messagesList.length === 0) {
      return null;
    }

    const results = [];

    for (const msgBuffer of response.messagesList) {
      try {
        const message = this.parseMessage_inner(msgBuffer);
        if (message) {
          results.push(message);
        }
      } catch (e) {
        console.error('解析单条消息失败:', e.message);
      }
    }

    return results.length > 0 ? results : null;
  }

  /**
   * 解析单条消息
   */
  parseMessage_inner(buffer) {
    const message = {};
    let offset = 0;

    while (offset < buffer.length) {
      const tag = buffer[offset++];
      if (!tag) break;

      const wireType = tag & 0x07;
      const fieldNumber = tag >> 3;

      if (wireType === 2) {
        const length = this.readVarint(buffer, offset);
        offset += this.varintSize(length);

        const value = buffer.slice(offset, offset + length);
        offset += length;

        if (fieldNumber === 1) {
          // method 字段
          message.method = value.toString('utf8');
        } else if (fieldNumber === 2) {
          // payload 字段（具体消息内容）
          message.payload = value;
        }
      } else if (wireType === 0) {
        const value = this.readVarint(buffer, offset);
        offset += this.varintSize(value);
      } else {
        break;
      }
    }

    // 根据method解析payload
    if (message.method && message.payload) {
      return this.parseMessagePayload(message.method, message.payload);
    }

    return null;
  }

  /**
   * 根据消息类型解析payload
   */
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
      // 提取文本信息
      const texts = this.extractTexts(payload);

      // 根据不同消息类型提取特定字段
      if (method === 'WebcastChatMessage') {
        this.statistics.chatCount++;
        return {
          ...result,
          messageType: '聊天消息',
          user: texts[0] || '匿名用户',
          content: texts[1] || texts[texts.length - 1] || '',
          allTexts: texts
        };
      }

      if (method === 'WebcastGiftMessage') {
        this.statistics.giftCount++;
        return {
          ...result,
          messageType: '礼物消息',
          user: texts[0] || '匿名用户',
          giftName: texts.find(t => t.includes('礼物') || t.length < 10) || texts[1] || '未知礼物',
          allTexts: texts
        };
      }

      if (method === 'WebcastLikeMessage') {
        this.statistics.likeCount++;
        return {
          ...result,
          messageType: '点赞消息',
          user: texts[0] || '匿名用户',
          allTexts: texts
        };
      }

      if (method === 'WebcastMemberMessage') {
        this.statistics.memberCount++;
        return {
          ...result,
          messageType: '进入直播间',
          user: texts[0] || '匿名用户',
          allTexts: texts
        };
      }

      if (method === 'WebcastRoomUserSeqMessage') {
        // 尝试提取在线人数
        const numbers = this.extractNumbers(payload);
        if (numbers.length > 0) {
          this.statistics.onlineUsers = numbers[0];
        }
        return {
          ...result,
          messageType: '在线人数',
          onlineCount: this.statistics.onlineUsers,
          numbers: numbers
        };
      }

      if (method === 'WebcastSocialMessage') {
        return {
          ...result,
          messageType: '关注消息',
          user: texts[0] || '匿名用户',
          allTexts: texts
        };
      }

      // 其他消息类型
      return {
        ...result,
        texts: texts
      };
    } catch (e) {
      return result;
    }
  }

  /**
   * 从Buffer中提取文本
   */
  extractTexts(buffer) {
    const texts = [];
    const str = buffer.toString('utf8');
    
    // 匹配中文、英文、数字的连续字符串
    const regex = /[\u4e00-\u9fa5a-zA-Z0-9]{2,}/g;
    const matches = str.match(regex);
    
    if (matches) {
      const seen = new Set();
      for (const match of matches) {
        // 过滤掉太长的（可能是乱码）和重复的
        if (match.length >= 2 && match.length <= 50 && !seen.has(match)) {
          // 过滤掉看起来像ID的纯数字
          if (!/^\d+$/.test(match) || match.length < 10) {
            texts.push(match);
            seen.add(match);
          }
        }
      }
    }
    
    return texts.slice(0, 20);
  }

  /**
   * 从Buffer中提取数字
   */
  extractNumbers(buffer) {
    const numbers = [];
    
    for (let i = 0; i < buffer.length - 4; i++) {
      if (buffer[i] < 0x80) {
        const num = buffer.readUInt32LE(i);
        if (num > 0 && num < 10000000) {
          numbers.push(num);
        }
      }
    }
    
    return numbers.slice(0, 5);
  }

  /**
   * 读取Varint
   */
  readVarint(buffer, offset) {
    let result = 0;
    let shift = 0;

    for (let i = 0; i < 10; i++) {
      if (offset + i >= buffer.length) break;

      const byte = buffer[offset + i];
      result |= (byte & 0x7f) << shift;

      if ((byte & 0x80) === 0) {
        return result;
      }

      shift += 7;
    }

    return result;
  }

  /**
   * 计算Varint占用的字节数
   */
  varintSize(value) {
    let size = 0;
    while (value > 0) {
      size++;
      value >>= 7;
    }
    return size || 1;
  }

  /**
   * 格式化消息用于显示
   */
  formatMessage(parsedMessages) {
    if (!parsedMessages) return null;

    // 如果是数组，格式化每条消息
    if (Array.isArray(parsedMessages)) {
      return parsedMessages.map(msg => this.formatSingleMessage(msg)).filter(Boolean).join('\n\n');
    }

    return this.formatSingleMessage(parsedMessages);
  }

  /**
   * 格式化单条消息
   */
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
        lines.push(`║ 内容: ${parsedMessage.content}`);
        break;

      case '礼物消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        lines.push(`║ 礼物: ${parsedMessage.giftName}`);
        break;

      case '点赞消息':
        lines.push(`║ 用户: ${parsedMessage.user} ❤️`);
        break;

      case '进入直播间':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        break;

      case '在线人数':
        lines.push(`║ 在线人数: ${parsedMessage.onlineCount} 👥`);
        break;

      case '关注消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        lines.push(`║ 动作: 关注了主播`);
        break;

      default:
        if (parsedMessage.user) {
          lines.push(`║ 用户: ${parsedMessage.user}`);
        }
        if (parsedMessage.texts && parsedMessage.texts.length > 0) {
          lines.push(`║ 提取信息: ${parsedMessage.texts.slice(0, 3).join(', ')}`);
        }
    }

    lines.push(`╚${'═'.repeat(78)}╝`);
    return lines.join('\n');
  }

  /**
   * 获取统计信息
   */
  getStatistics() {
    return {
      ...this.statistics,
      timestamp: new Date().toISOString()
    };
  }

  /**
   * 重置统计信息
   */
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

  /**
   * 格式化统计信息
   */
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
