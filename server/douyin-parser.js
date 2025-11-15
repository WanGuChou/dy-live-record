/**
 * 抖音直播WebSocket消息解析器
 * 参考: https://github.com/skmcj/dycast
 * 
 * 抖音直播使用Protobuf格式传输消息
 * 主要消息类型：
 * - WebcastChatMessage: 聊天消息
 * - WebcastGiftMessage: 礼物消息
 * - WebcastLikeMessage: 点赞消息
 * - WebcastMemberMessage: 进入直播间
 * - WebcastSocialMessage: 关注消息
 * - WebcastRoomUserSeqMessage: 在线人数
 */

const zlib = require('zlib');
const { promisify } = require('util');

const gunzip = promisify(zlib.gunzip);
const inflate = promisify(zlib.inflate);

/**
 * 抖音消息解析器
 */
class DouyinMessageParser {
  constructor() {
    this.messageTypes = {
      'WebcastChatMessage': 'chat',
      'WebcastGiftMessage': 'gift',
      'WebcastLikeMessage': 'like',
      'WebcastMemberMessage': 'member',
      'WebcastSocialMessage': 'social',
      'WebcastRoomUserSeqMessage': 'viewer_count',
      'WebcastFansclubMessage': 'fansclub',
      'WebcastControlMessage': 'control',
      'WebcastEmojiChatMessage': 'emoji',
      'WebcastRoomStatsMessage': 'room_stats'
    };
  }

  /**
   * 检测是否是抖音WebSocket消息
   */
  isDouyinMessage(url) {
    if (!url) return false;
    return url.includes('webcast') && 
           url.includes('douyin.com') || 
           url.includes('douyincdn.com');
  }

  /**
   * 解析WebSocket消息
   * @param {string|Buffer} payload - 消息负载
   * @returns {Object} 解析后的消息
   */
  async parseMessage(payload) {
    try {
      // 将字符串转为Buffer
      let buffer;
      if (typeof payload === 'string') {
        // 尝试base64解码
        try {
          buffer = Buffer.from(payload, 'base64');
        } catch (e) {
          buffer = Buffer.from(payload);
        }
      } else {
        buffer = payload;
      }

      // 检查是否压缩
      const compressed = this.isCompressed(buffer);
      if (compressed) {
        buffer = await this.decompress(buffer);
      }

      // 解析Protobuf消息
      return this.parseProtobuf(buffer);
    } catch (error) {
      console.error('解析消息失败:', error.message);
      return {
        error: error.message,
        rawLength: payload?.length || 0
      };
    }
  }

  /**
   * 检测是否压缩（GZIP或Deflate）
   */
  isCompressed(buffer) {
    if (buffer.length < 2) return false;
    
    // GZIP magic number: 0x1f 0x8b
    if (buffer[0] === 0x1f && buffer[1] === 0x8b) {
      return 'gzip';
    }
    
    // Deflate (zlib): 0x78
    if (buffer[0] === 0x78) {
      return 'deflate';
    }
    
    return false;
  }

  /**
   * 解压缩消息
   */
  async decompress(buffer) {
    const type = this.isCompressed(buffer);
    
    if (type === 'gzip') {
      return await gunzip(buffer);
    } else if (type === 'deflate') {
      return await inflate(buffer);
    }
    
    return buffer;
  }

  /**
   * 解析Protobuf消息（简化版）
   * 注：完整解析需要.proto文件和protobuf库
   */
  parseProtobuf(buffer) {
    try {
      // 简单的字段提取
      const result = {
        type: 'unknown',
        data: {},
        raw: buffer.toString('hex').substring(0, 100) + '...'
      };

      // 尝试提取消息类型
      const messageType = this.extractMessageType(buffer);
      if (messageType) {
        result.type = this.messageTypes[messageType] || messageType;
        result.messageType = messageType;
      }

      // 尝试提取可读文本
      const texts = this.extractTexts(buffer);
      if (texts.length > 0) {
        result.texts = texts;
      }

      // 尝试提取数字字段
      const numbers = this.extractNumbers(buffer);
      if (numbers.length > 0) {
        result.numbers = numbers;
      }

      return result;
    } catch (error) {
      return {
        error: 'Protobuf解析失败: ' + error.message,
        hex: buffer.toString('hex').substring(0, 200)
      };
    }
  }

  /**
   * 提取消息类型名称
   */
  extractMessageType(buffer) {
    // 在buffer中搜索已知的消息类型名称
    const bufferStr = buffer.toString('utf8', 0, Math.min(buffer.length, 500));
    
    for (const typeName of Object.keys(this.messageTypes)) {
      if (bufferStr.includes(typeName)) {
        return typeName;
      }
    }
    
    return null;
  }

  /**
   * 提取可读文本
   */
  extractTexts(buffer) {
    const texts = [];
    const str = buffer.toString('utf8');
    
    // 匹配中文、英文、数字的连续字符串
    const regex = /[\u4e00-\u9fa5a-zA-Z0-9]{2,}/g;
    const matches = str.match(regex);
    
    if (matches) {
      // 过滤掉乱码和重复
      const seen = new Set();
      for (const match of matches) {
        if (match.length >= 2 && match.length <= 100 && !seen.has(match)) {
          texts.push(match);
          seen.add(match);
        }
      }
    }
    
    return texts.slice(0, 20); // 最多20个
  }

  /**
   * 提取数字字段（可能是用户ID、礼物ID等）
   */
  extractNumbers(buffer) {
    const numbers = [];
    
    // Protobuf的varint编码检测
    for (let i = 0; i < buffer.length - 4; i++) {
      // 跳过高位字节
      if (buffer[i] < 0x80) {
        // 读取32位整数
        const num = buffer.readUInt32LE(i);
        if (num > 0 && num < 0xFFFFFF) {
          numbers.push(num);
        }
      }
    }
    
    return numbers.slice(0, 10); // 最多10个
  }

  /**
   * 简化的消息格式化（用于日志输出）
   */
  formatMessage(parsed) {
    const lines = [];
    
    lines.push(`消息类型: ${parsed.messageType || parsed.type || 'unknown'}`);
    
    if (parsed.texts && parsed.texts.length > 0) {
      lines.push(`文本内容:`);
      parsed.texts.forEach(text => {
        lines.push(`  - ${text}`);
      });
    }
    
    if (parsed.numbers && parsed.numbers.length > 0) {
      lines.push(`数字字段: ${parsed.numbers.slice(0, 5).join(', ')}`);
    }
    
    if (parsed.error) {
      lines.push(`错误: ${parsed.error}`);
    }
    
    return lines.join('\n');
  }
}

/**
 * 创建解析器单例
 */
const parser = new DouyinMessageParser();

module.exports = {
  DouyinMessageParser,
  parser,
  isDouyinMessage: (url) => parser.isDouyinMessage(url),
  parseMessage: (payload) => parser.parseMessage(payload),
  formatMessage: (parsed) => parser.formatMessage(parsed)
};
