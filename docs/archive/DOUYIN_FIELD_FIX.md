# 抖音消息字段编号修复

## 问题诊断

### 错误症状
```
[Douyin] 解析 WebcastGiftMessage 失败: Invalid wire type: 6
[Douyin] 解析 WebcastGiftMessage 失败: Invalid wire type: 7
[Douyin] 解析 WebcastMemberMessage 失败: Invalid wire type: 7

礼物消息输出：
║ 用户: undefined
║ 礼物: undefined
```

### 根本原因

**Wire type 6 和 7 是无效的**（Protobuf 只有 wire type 0-5）。这说明在解析过程中 **ByteBuffer offset 错位**。

#### 为什么会错位？

1. **字段编号（field number）错误**
   - 我之前猜测 GiftMessage.gift 是 field 9
   - 实际上是 **field 15**
   - 导致在 field 9-14 之间无法正确跳过字段

2. **GiftStruct 字段编号错误**
   - 我之前认为：id=field 1, name=field 2, diamondCount=field 10
   - 实际上是：**id=field 5, name=field 16, diamondCount=field 12**
   - 导致读取到错误位置的数据

3. **连锁反应**
   - 字段 A 读取错误 → offset 偏移
   - 字段 B 从错误位置读取 tag
   - tag 的低 3 位（wire type）变成无效值 6 或 7
   - 抛出 `Invalid wire type` 错误

## 修复方案

### 1. 查阅 dycast 源码，确认正确的字段编号

参考 `dycast/src/core/model.ts`：

#### GiftMessage（部分字段）
```typescript
case 1: // Common common
case 2: // int64 giftId
case 3: // int64 fanTicketCount
case 4: // int64 groupCount
case 5: // int64 repeatCount ✅
case 6: // int64 comboCount ✅
case 7: // User user ✅
case 8: // User toUser
case 9: // int32 repeatEnd
...
case 15: // GiftStruct gift ✅ 重要！不是 field 9
```

#### GiftStruct（部分字段）
```typescript
case 1: // Image image
case 2: // string describe
case 3: // bool notify
case 4: // int64 duration
case 5: // int64 id ✅ 重要！不是 field 1
...
case 12: // int32 diamondCount ✅ 重要！不是 field 10
...
case 16: // string name ✅ 重要！不是 field 2
```

### 2. 修正所有解码函数

#### Before（错误）
```javascript
function decodeGiftMessage(binary) {
  // ...
  switch (fieldNumber) {
    case 7: // user ✅ 正确
      // ...
      break;
    case 9: // gift ❌ 错误！应该是 field 15
      message.gift = decodeGiftStruct(bb);
      break;
  }
}

function decodeGiftStruct(bb) {
  // ...
  switch (fieldNumber) {
    case 1: // id ❌ 错误！应该是 field 5
      gift.id = readVarint64(bb, false);
      break;
    case 2: // name ❌ 错误！应该是 field 16
      gift.name = readString(bb, readVarint32(bb));
      break;
    case 10: // diamondCount ❌ 错误！应该是 field 12
      gift.diamondCount = readVarint32(bb);
      break;
  }
}
```

#### After（正确）
```javascript
function decodeGiftMessage(binary) {
  // ...
  switch (fieldNumber) {
    case 7: // user ✅
      // ...
      break;
    case 15: // gift ✅ 修正
      message.gift = decodeGiftStruct(bb);
      break;
    default:
      skipUnknownField(bb, tag & 7); // 跳过其他字段
  }
}

function decodeGiftStruct(bb) {
  // ...
  switch (fieldNumber) {
    case 5: // id ✅ 修正
      gift.id = readVarint64(bb, false);
      break;
    case 12: // diamondCount ✅ 修正
      gift.diamondCount = readVarint32(bb);
      break;
    case 16: // name ✅ 修正
      gift.name = readString(bb, readVarint32(bb));
      break;
    default:
      skipUnknownField(bb, tag & 7); // 跳过其他字段
  }
}
```

### 3. 优化循环结构（使用 dycast 标准）

#### Before
```javascript
while (!isAtEnd(bb)) {
  const tag = readVarint32(bb);
  const wireType = tag & 7;
  const fieldNumber = tag >>> 3;
  
  if (fieldNumber === 0) break; // 可能无法正确退出
  
  switch (fieldNumber) {
    // ...
    default:
      skipUnknownField(bb, wireType);
  }
}
```

#### After
```javascript
end_of_message: while (!isAtEnd(bb)) {
  const tag = readVarint32(bb);
  const fieldNumber = tag >>> 3;
  
  switch (fieldNumber) {
    case 0:
      break end_of_message; // 明确退出循环
    
    // ... 其他 case
    
    default:
      skipUnknownField(bb, tag & 7); // 统一使用 tag & 7
  }
}
```

## 修复后的完整字段映射

### 所有已实现的消息类型

| 消息类型 | 关键字段 | Field Number | Wire Type |
|---------|---------|--------------|-----------|
| **ChatMessage** | | | |
| | user | 2 | 2 (length-delimited) |
| | content | 3 | 2 (length-delimited) |
| **GiftMessage** | | | |
| | giftId | 2 | 0 (varint) |
| | repeatCount | 5 | 0 (varint) |
| | comboCount | 6 | 0 (varint) |
| | user | 7 | 2 (length-delimited) |
| | gift | 15 | 2 (length-delimited) |
| **GiftStruct** | | | |
| | id | 5 | 0 (varint) |
| | diamondCount | 12 | 0 (varint) |
| | name | 16 | 2 (length-delimited) |
| **LikeMessage** | | | |
| | user | 2 | 2 (length-delimited) |
| | count | 3 | 0 (varint) |
| | total | 4 | 0 (varint) |
| **MemberMessage** | | | |
| | user | 2 | 2 (length-delimited) |
| | memberCount | 3 | 0 (varint) |
| **User** | | | |
| | id | 1 | 0 (varint) |
| | shortId | 2 | 0 (varint) |
| | nickname | 3 | 2 (length-delimited) |
| | level | 6 | 0 (varint) |

## 测试验证

### 预期输出（修复后）

#### 礼物消息
```
╔══════════════════════════════════════════════════════════════════════════════╗
║ 🎬 抖音直播消息
╠══════════════════════════════════════════════════════════════════════════════╣
║ 消息类型: 礼物消息
║ 时间: 2025-11-15T16:56:12.155Z
║ 用户: 张三
║ 礼物: 玫瑰花
║ 数量: 10
║ 价值: 100 💎
╚══════════════════════════════════════════════════════════════════════════════╝
```

#### 聊天消息
```
╔══════════════════════════════════════════════════════════════════════════════╗
║ 🎬 抖音直播消息
╠══════════════════════════════════════════════════════════════════════════════╣
║ 消息类型: 聊天消息
║ 时间: 2025-11-15T16:56:15.456Z
║ 用户: 李四
║ 等级: 20
║ 内容: 主播好！
╚══════════════════════════════════════════════════════════════════════════════╝
```

### 不应再出现的错误

❌ `Invalid wire type: 6`  
❌ `Invalid wire type: 7`  
❌ `用户: undefined`  
❌ `礼物: undefined`

## 调试技巧

### 1. 添加调试日志

```javascript
function decodeGiftMessage(binary) {
  const bb = createByteBuffer(binary);
  const message = {};

  end_of_message: while (!isAtEnd(bb)) {
    const tag = readVarint32(bb);
    const fieldNumber = tag >>> 3;
    const wireType = tag & 7;
    
    console.log(`[DEBUG] GiftMessage field ${fieldNumber}, wire type ${wireType}`);
    
    switch (fieldNumber) {
      // ...
    }
  }
  
  return message;
}
```

### 2. 验证 ByteBuffer 状态

```javascript
console.log(`[DEBUG] BB offset=${bb.offset}, limit=${bb.limit}, remaining=${bb.limit - bb.offset}`);
```

### 3. 十六进制查看原始数据

```javascript
function debugPayload(payload) {
  const hex = Array.from(payload.slice(0, 50))
    .map(b => b.toString(16).padStart(2, '0'))
    .join(' ');
  console.log(`[DEBUG] Payload hex: ${hex}`);
}
```

## 如何避免此类错误

1. **始终参考源代码**：不要猜测字段编号，直接查看 dycast 的 `.ts` 文件
2. **使用调试日志**：在解析失败时打印 field number 和 wire type
3. **完整测试**：测试所有消息类型（聊天、礼物、点赞、进入直播间等）
4. **渐进开发**：先实现一个消息类型，验证通过后再添加其他类型

## Protobuf Wire Type 参考

| Wire Type | 含义 | 使用场景 |
|-----------|------|---------|
| 0 | Varint | int32, int64, uint32, uint64, bool, enum |
| 1 | 64-bit | fixed64, sfixed64, double |
| 2 | Length-delimited | string, bytes, embedded messages, packed repeated fields |
| 3 | Start group | **已废弃** |
| 4 | End group | **已废弃** |
| 5 | 32-bit | fixed32, sfixed32, float |
| 6 | **无效** | ❌ 不存在 |
| 7 | **无效** | ❌ 不存在 |

**如果遇到 wire type 6 或 7，说明 offset 已经错位！**

## 总结

✅ **修复内容**：
- 修正 GiftMessage.gift 字段编号：9 → 15
- 修正 GiftStruct.id 字段编号：1 → 5
- 修正 GiftStruct.name 字段编号：2 → 16
- 修正 GiftStruct.diamondCount 字段编号：10 → 12
- 优化循环结构，使用 `end_of_message` 标签
- 统一使用 `tag & 7` 获取 wire type

✅ **预期结果**：
- 不再出现 `Invalid wire type` 错误
- 礼物消息正确显示用户名、礼物名、数量、价值
- 所有消息类型完整解析

---

**参考资源**：
- dycast model.ts: https://github.com/skmcj/dycast/blob/main/src/core/model.ts
- Protobuf Encoding: https://protobuf.dev/programming-guides/encoding/
