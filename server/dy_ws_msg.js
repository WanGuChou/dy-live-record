/**
 * 抖音直播 WebSocket 消息解析器
 * 参考: https://github.com/skmcj/dycast
 * 
 * 解析来自 wss://webcast100-ws-web-hl.douyin.com 的消息
 * 支持的消息类型：
 * - WebcastChatMessage: 聊天消息
 * - WebcastGiftMessage: 礼物消息  
 * - WebcastLikeMessage: 点赞消息
 * - WebcastMemberMessage: 用户进入直播间
 * - WebcastSocialMessage: 关注消息
 * - WebcastRoomUserSeqMessage: 在线人数更新
 * - WebcastFansclubMessage: 粉丝团消息
 */

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
           url.includes('douyin.com') &&
           (url.includes('/webcast/') || url.includes('ws-web'));
  }

  /**
   * 解析 WebSocket 消息
   * @param {string} payloadData - WebSocket消息内容
   * @param {string} url - WebSocket URL
   * @returns {Object|null} 解析后的消息对象
   */
  parseMessage(payloadData, url = '') {
    if (!payloadData) return null;
    
    try {
      // 尝试解析为JSON
      const message = JSON.parse(payloadData);
      this.statistics.totalMessages++;

      // 检查是否为抖音直播消息格式
      if (message.method || message.payload) {
        return this.parseDouyinMessage(message);
      }

      return null;
    } catch (e) {
      // 如果不是JSON，可能是protobuf或其他二进制格式
      // 抖音直播使用protobuf，这里做简单的文本提取
      return this.parseBinaryMessage(payloadData);
    }
  }

  /**
   * 解析抖音消息（JSON格式）
   */
  parseDouyinMessage(message) {
    const result = {
      type: 'douyin_live',
      timestamp: new Date().toISOString(),
      parsed: true
    };

    // WebcastChatMessage - 聊天消息
    if (message.method === 'WebcastChatMessage' || message.type === 'chat') {
      this.statistics.chatCount++;
      return {
        ...result,
        messageType: '聊天消息',
        user: message.user?.nickname || message.nickname || '匿名用户',
        userId: message.user?.id || message.userId,
        content: message.content || message.text || '',
        userLevel: message.user?.level,
        userBadges: message.user?.badges || []
      };
    }

    // WebcastGiftMessage - 礼物消息
    if (message.method === 'WebcastGiftMessage' || message.type === 'gift') {
      this.statistics.giftCount++;
      return {
        ...result,
        messageType: '礼物消息',
        user: message.user?.nickname || '匿名用户',
        userId: message.user?.id,
        giftName: message.gift?.name || message.giftName || '未知礼物',
        giftId: message.gift?.id || message.giftId,
        giftCount: message.giftCount || message.count || 1,
        giftValue: message.gift?.diamondCount || 0,
        totalValue: (message.giftCount || 1) * (message.gift?.diamondCount || 0),
        comboCount: message.comboCount || 0,
        giftIcon: message.gift?.image?.urlList?.[0]
      };
    }

    // WebcastLikeMessage - 点赞消息
    if (message.method === 'WebcastLikeMessage' || message.type === 'like') {
      this.statistics.likeCount++;
      return {
        ...result,
        messageType: '点赞消息',
        user: message.user?.nickname || '匿名用户',
        userId: message.user?.id,
        likeCount: message.count || 1,
        totalLikes: message.total || 0
      };
    }

    // WebcastMemberMessage - 用户进入
    if (message.method === 'WebcastMemberMessage' || message.type === 'member') {
      this.statistics.memberCount++;
      return {
        ...result,
        messageType: '进入直播间',
        user: message.user?.nickname || '匿名用户',
        userId: message.user?.id,
        userLevel: message.user?.level,
        memberCount: message.memberCount || 0
      };
    }

    // WebcastSocialMessage - 关注消息
    if (message.method === 'WebcastSocialMessage' || message.type === 'social') {
      return {
        ...result,
        messageType: '关注消息',
        user: message.user?.nickname || '匿名用户',
        userId: message.user?.id,
        action: message.action || 'follow'
      };
    }

    // WebcastRoomUserSeqMessage - 在线人数
    if (message.method === 'WebcastRoomUserSeqMessage' || message.type === 'room_user_seq') {
      this.statistics.onlineUsers = message.total || message.onlineUserCount || 0;
      return {
        ...result,
        messageType: '在线人数',
        onlineCount: this.statistics.onlineUsers,
        totalUsers: message.totalUser || 0
      };
    }

    // WebcastFansclubMessage - 粉丝团消息
    if (message.method === 'WebcastFansclubMessage' || message.type === 'fansclub') {
      return {
        ...result,
        messageType: '粉丝团消息',
        user: message.user?.nickname || '匿名用户',
        content: message.content || '',
        fanLevel: message.fanTicket?.level || 0
      };
    }

    // 其他消息类型
    return {
      ...result,
      messageType: message.method || message.type || '未知消息',
      rawData: message
    };
  }

  /**
   * 解析二进制消息（Protobuf）
   * 抖音使用protobuf，这里做简单的文本提取
   */
  parseBinaryMessage(payloadData) {
    // 尝试从二进制数据中提取可读文本
    const textMatches = payloadData.match(/[\u4e00-\u9fa5a-zA-Z0-9]+/g);
    
    if (textMatches && textMatches.length > 0) {
      return {
        type: 'douyin_live',
        messageType: '二进制消息（未完全解析）',
        timestamp: new Date().toISOString(),
        parsed: false,
        extractedText: textMatches.slice(0, 10).join(' '),
        rawLength: payloadData.length
      };
    }

    return null;
  }

  /**
   * 格式化消息用于显示
   */
  formatMessage(parsedMessage) {
    if (!parsedMessage) return null;

    const lines = [];
    lines.push(`╔${'═'.repeat(78)}╗`);
    lines.push(`║ 🎬 抖音直播消息`);
    lines.push(`╠${'═'.repeat(78)}╣`);
    lines.push(`║ 消息类型: ${parsedMessage.messageType}`);
    lines.push(`║ 时间: ${parsedMessage.timestamp}`);

    switch (parsedMessage.messageType) {
      case '聊天消息':
        lines.push(`║ 用户: ${parsedMessage.user} ${parsedMessage.userLevel ? `[Lv.${parsedMessage.userLevel}]` : ''}`);
        lines.push(`║ 内容: ${parsedMessage.content}`);
        if (parsedMessage.userBadges && parsedMessage.userBadges.length > 0) {
          lines.push(`║ 徽章: ${parsedMessage.userBadges.map(b => b.name || b).join(', ')}`);
        }
        break;

      case '礼物消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        lines.push(`║ 礼物: ${parsedMessage.giftName} x ${parsedMessage.giftCount}`);
        lines.push(`║ 价值: ${parsedMessage.totalValue} 💎`);
        if (parsedMessage.comboCount > 0) {
          lines.push(`║ 连击: ${parsedMessage.comboCount}`);
        }
        break;

      case '点赞消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        lines.push(`║ 点赞数: ${parsedMessage.likeCount} ❤️`);
        break;

      case '进入直播间':
        lines.push(`║ 用户: ${parsedMessage.user} ${parsedMessage.userLevel ? `[Lv.${parsedMessage.userLevel}]` : ''}`);
        lines.push(`║ 当前人数: ${parsedMessage.memberCount}`);
        break;

      case '在线人数':
        lines.push(`║ 在线人数: ${parsedMessage.onlineCount} 👥`);
        break;

      case '关注消息':
        lines.push(`║ 用户: ${parsedMessage.user}`);
        lines.push(`║ 动作: ${parsedMessage.action === 'follow' ? '关注了主播' : parsedMessage.action}`);
        break;

      default:
        if (parsedMessage.user) {
          lines.push(`║ 用户: ${parsedMessage.user}`);
        }
        if (parsedMessage.content) {
          lines.push(`║ 内容: ${parsedMessage.content}`);
        }
        if (parsedMessage.extractedText) {
          lines.push(`║ 提取文本: ${parsedMessage.extractedText}`);
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
