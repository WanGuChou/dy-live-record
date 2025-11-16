# 抖音直播 WebSocket 消息解析技术文档

## 概述

本文档详细说明抖音直播 WebSocket 消息的解析技术实现。

参考项目：https://github.com/skmcj/dycast

## 消息结构

### 1. 完整消息层次

```
Base64编码的WebSocket数据
  ↓
PushFrame (外层Protobuf结构)
├── logId: uint64
├── service: uint32
├── method: string
├── headers_list: Header[]
│   └── Header
│       ├── key: string
│       └── value: string
└── payloadBinary: bytes (GZIP压缩)
    ↓ (解压后)
    Response (内层Protobuf结构)
    └── messagesList: Message[]
        └── Message
            ├── method: string (消息类型)
            └── payload: bytes (具体消息内容)
```

### 2. Protobuf Wire Types

```javascript
0: Varint (int32, int64, uint32, uint64, bool, enum)
1: 64-bit (fixed64, double)
2: Length-delimited (string, bytes, embedded messages)
3: Start group (deprecated)
4: End group (deprecated)
5: 32-bit (fixed32, float)
```

## 解析流程

### 步骤 1: 解析 PushFrame

```javascript
parsePushFrame(buffer) {
  let offset = 0;
  const frame = {};
  
  while (offset < buffer.length) {
    // 读取 Tag: (field_number << 3) | wire_type
    const tag = buffer[offset++];
    const wireType = tag & 0x07;
    const fieldNumber = tag >> 3;
    
    if (wireType === 2) { // Length-delimited
      const length = this.readVarint(buffer, offset);
      offset += this.varintSize(length);
      const value = buffer.slice(offset, offset + length);
      offset += length;
      
      // 字段映射
      if (fieldNumber === 1) frame.logId = value;
      else if (fieldNumber === 3) frame.method = value.toString('utf8');
      else if (fieldNumber === 4) frame.headersList = this.parseHeadersList(value);
      else if (fieldNumber === 5) frame.payloadBinary = value;
    }
  }
  
  return frame;
}
```

### 步骤 2: 读取 Varint

Protobuf 使用 Varint 编码来压缩整数：

```javascript
readVarint(buffer, offset) {
  let result = 0;
  let shift = 0;
  
  for (let i = 0; i < 10; i++) {
    const byte = buffer[offset + i];
    
    // 取低7位
    result |= (byte & 0x7f) << shift;
    
    // 如果最高位为0，表示这是最后一个字节
    if ((byte & 0x80) === 0) {
      return result;
    }
    
    shift += 7;
  }
  
  return result;
}
```

**示例**：
- `0x08` → `8`
- `0x96 0x01` → `150` (0x96 = 0b10010110, 0x01 = 0b00000001)
  - 第一字节：低7位 = 0010110 (22)
  - 第二字节：低7位 = 0000001 (1)
  - 结果：22 + (1 << 7) = 22 + 128 = 150

### 步骤 3: 解析 headers_list

```javascript
parseHeadersList(buffer) {
  const headers = {};
  let offset = 0;
  
  while (offset < buffer.length) {
    const tag = buffer[offset++];
    const wireType = tag & 0x07;
    const fieldNumber = tag >> 3;
    
    // fieldNumber === 3 表示 Header 消息
    if (wireType === 2 && fieldNumber === 3) {
      const length = this.readVarint(buffer, offset);
      offset += this.varintSize(length);
      const headerData = buffer.slice(offset, offset + length);
      offset += length;
      
      const header = this.parseHeader(headerData);
      if (header && header.key) {
        headers[header.key] = header.value;
      }
    }
  }
  
  return headers;
}
```

### 步骤 4: GZIP 解压

```javascript
async parseResponse(frame) {
  let payload = frame.payloadBinary;
  
  // 检查是否需要解压
  const compressType = frame.headersList?.['compress_type'];
  
  if (compressType === 'gzip') {
    payload = await gunzip(payload);
  }
  
  // 解析 Response 结构
  const response = {};
  let offset = 0;
  
  while (offset < payload.length) {
    const tag = payload[offset++];
    const wireType = tag & 0x07;
    const fieldNumber = tag >> 3;
    
    if (wireType === 2 && fieldNumber === 1) {
      // messagesList 字段
      const length = this.readVarint(payload, offset);
      offset += this.varintSize(length);
      const value = payload.slice(offset, offset + length);
      offset += length;
      
      if (!response.messagesList) {
        response.messagesList = [];
      }
      response.messagesList.push(value);
    }
  }
  
  return response;
}
```

### 步骤 5: 解析单条消息

```javascript
parseMessage_inner(buffer) {
  const message = {};
  let offset = 0;
  
  while (offset < buffer.length) {
    const tag = buffer[offset++];
    const wireType = tag & 0x07;
    const fieldNumber = tag >> 3;
    
    if (wireType === 2) {
      const length = this.readVarint(buffer, offset);
      offset += this.varintSize(length);
      const value = buffer.slice(offset, offset + length);
      offset += length;
      
      if (fieldNumber === 1) {
        message.method = value.toString('utf8');  // "WebcastChatMessage"
      } else if (fieldNumber === 2) {
        message.payload = value;  // 具体消息内容
      }
    }
  }
  
  return message;
}
```

### 步骤 6: 提取消息内容

```javascript
parseMessagePayload(method, payload) {
  const result = {
    type: 'douyin_live',
    messageType: this.messageTypes[method] || method,
    method: method,
    timestamp: new Date().toISOString(),
    parsed: true
  };
  
  // 提取可读文本
  const texts = this.extractTexts(payload);
  
  // 根据消息类型特殊处理
  if (method === 'WebcastChatMessage') {
    return {
      ...result,
      messageType: '聊天消息',
      user: texts[0] || '匿名用户',
      content: texts[1] || texts[texts.length - 1] || '',
      allTexts: texts
    };
  }
  
  if (method === 'WebcastGiftMessage') {
    return {
      ...result,
      messageType: '礼物消息',
      user: texts[0] || '匿名用户',
      giftName: texts.find(t => t.includes('礼物') || t.length < 10) || texts[1] || '未知礼物',
      allTexts: texts
    };
  }
  
  // ... 其他消息类型
}
```

### 步骤 7: 文本提取算法

由于没有完整的 `.proto` 定义文件，使用启发式方法提取文本：

```javascript
extractTexts(buffer) {
  const texts = [];
  const str = buffer.toString('utf8');
  
  // 正则匹配中文、英文、数字的连续字符串
  const regex = /[\u4e00-\u9fa5a-zA-Z0-9]{2,}/g;
  const matches = str.match(regex);
  
  if (matches) {
    const seen = new Set();
    for (const match of matches) {
      // 过滤条件：
      // 1. 长度在 2-50 之间（避免乱码）
      // 2. 不重复
      // 3. 如果是纯数字，长度要小于10（避免ID）
      if (match.length >= 2 && match.length <= 50 && !seen.has(match)) {
        if (!/^\d+$/.test(match) || match.length < 10) {
          texts.push(match);
          seen.add(match);
        }
      }
    }
  }
  
  return texts.slice(0, 20); // 最多返回20个
}
```

## 支持的消息类型

| Method | 中文名称 | 主要字段 |
|--------|---------|---------|
| WebcastChatMessage | 聊天消息 | user, content |
| WebcastGiftMessage | 礼物消息 | user, giftName, count |
| WebcastLikeMessage | 点赞消息 | user, count |
| WebcastMemberMessage | 进入直播间 | user |
| WebcastSocialMessage | 关注消息 | user |
| WebcastRoomUserSeqMessage | 在线人数 | onlineCount |
| WebcastFansclubMessage | 粉丝团消息 | user, level |
| WebcastControlMessage | 直播间控制 | - |
| WebcastEmojiChatMessage | 表情消息 | user, emoji |
| WebcastRoomStatsMessage | 直播间统计 | - |
| WebcastLinkMicBattle | 连麦PK | - |
| WebcastLinkMicArmies | 连麦军团 | - |

## 输出格式

### 聊天消息示例

```
╔══════════════════════════════════════════════════════════════════════════════╗
║ 🎬 抖音直播消息
╠══════════════════════════════════════════════════════════════════════════════╣
║ 消息类型: 聊天消息
║ 时间: 2025-11-15T15:42:50.428Z
║ 用户: 用户昵称
║ 内容: 消息内容
╚══════════════════════════════════════════════════════════════════════════════╝
```

### 礼物消息示例

```
╔══════════════════════════════════════════════════════════════════════════════╗
║ 🎬 抖音直播消息
╠══════════════════════════════════════════════════════════════════════════════╣
║ 消息类型: 礼物消息
║ 时间: 2025-11-15T15:42:50.428Z
║ 用户: 用户昵称
║ 礼物: 玫瑰花
╚══════════════════════════════════════════════════════════════════════════════╝
```

## 调试技巧

### 1. 查看原始二进制数据

```javascript
console.log('原始数据 (hex):', buffer.toString('hex'));
console.log('原始数据 (base64):', buffer.toString('base64'));
```

### 2. 查看Protobuf字段

```javascript
while (offset < buffer.length) {
  const tag = buffer[offset];
  const wireType = tag & 0x07;
  const fieldNumber = tag >> 3;
  console.log(`字段 ${fieldNumber}, Wire Type ${wireType}`);
  // ...
}
```

### 3. 查看解压后的数据

```javascript
const decompressed = await gunzip(payload);
console.log('解压后大小:', decompressed.length);
console.log('解压后前100字节:', decompressed.slice(0, 100).toString('utf8'));
```

## 性能优化

### 1. Buffer 复用

```javascript
const bufferPool = [];

function getBuffer(size) {
  if (bufferPool.length > 0) {
    const buffer = bufferPool.pop();
    if (buffer.length >= size) return buffer;
  }
  return Buffer.allocUnsafe(size);
}

function recycleBuffer(buffer) {
  if (buffer.length <= 8192) {
    bufferPool.push(buffer);
  }
}
```

### 2. 避免字符串转换

```javascript
// 不好的做法
const str = buffer.toString('utf8');
const matches = str.match(/pattern/g);

// 好的做法（直接操作Buffer）
for (let i = 0; i < buffer.length; i++) {
  if (buffer[i] >= 0x4e00 && buffer[i] <= 0x9fa5) {
    // 处理中文字符
  }
}
```

### 3. 缓存解析结果

```javascript
const messageCache = new Map();

function parseMessageCached(buffer) {
  const key = buffer.toString('base64');
  if (messageCache.has(key)) {
    return messageCache.get(key);
  }
  
  const result = parseMessage(buffer);
  messageCache.set(key, result);
  
  // 限制缓存大小
  if (messageCache.size > 1000) {
    const firstKey = messageCache.keys().next().value;
    messageCache.delete(firstKey);
  }
  
  return result;
}
```

## 故障排查

### 问题 1: 解压失败

**现象**：`GZIP解压失败: incorrect header check`

**原因**：
- payload 不是 GZIP 格式
- compress_type 读取错误

**解决**：
```javascript
if (compressType === 'gzip') {
  try {
    payload = await gunzip(payload);
  } catch (e) {
    console.error('GZIP解压失败，尝试原始数据:', e.message);
    // 继续使用原始payload
  }
}
```

### 问题 2: 提取不到文本

**现象**：`allTexts: []`

**原因**：
- Protobuf 二进制格式中文本被编码
- 正则表达式匹配失败

**解决**：
```javascript
// 尝试多种字符编码
function extractTexts(buffer) {
  const texts = [];
  
  // UTF-8
  const utf8Str = buffer.toString('utf8');
  texts.push(...extractFromString(utf8Str));
  
  // GBK/GB2312 (需要 iconv-lite)
  // const gbkStr = iconv.decode(buffer, 'gbk');
  // texts.push(...extractFromString(gbkStr));
  
  return texts;
}
```

### 问题 3: method 为空

**现象**：`method: undefined`

**原因**：
- PushFrame 解析失败
- 字段编号错误

**解决**：
```javascript
// 打印所有字段
while (offset < buffer.length) {
  const tag = buffer[offset++];
  const wireType = tag & 0x07;
  const fieldNumber = tag >> 3;
  
  console.log(`字段 #${fieldNumber}, Wire Type: ${wireType}`);
  
  // ... 继续解析
}
```

## 参考资源

- **Protocol Buffers 文档**: https://protobuf.dev/
- **Protocol Buffers Encoding**: https://protobuf.dev/programming-guides/encoding/
- **dycast 项目**: https://github.com/skmcj/dycast （核心代码来源）
- **dycast model.ts**: https://github.com/skmcj/dycast/blob/main/src/core/model.ts （ByteBuffer 实现）
- **pako**: https://github.com/nodeca/pako （GZIP 解压库）
- **抖音直播协议分析**: https://github.com/YunzhiYike/live-tool

## 许可证

MIT License
