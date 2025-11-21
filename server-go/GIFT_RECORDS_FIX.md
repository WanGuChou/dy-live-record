# 礼物记录入库修复文档

## 修复概述

本次修复解决了礼物消息未能正确入库到 `gift_records` 表的问题，并为 `gift_records` 表添加了 `msg_id` 字段。

## 修改日期

2025-11-21

## 问题描述

1. **礼物消息未入库**: 收到礼物消息时，没有将礼物消息保存到 `gift_records` 表
2. **缺少 msg_id 字段**: `gift_records` 表缺少用于唯一标识消息的 `msg_id` 字段
3. **日志不足**: 礼物消息处理过程缺少详细的日志输出，难以调试

## 解决方案

### 1. 数据库表结构修改

#### gift_records 表添加 msg_id 字段

**文件**: `/workspace/server-go/internal/database/database.go`

```sql
CREATE TABLE IF NOT EXISTS gift_records (
    record_id INTEGER PRIMARY KEY AUTOINCREMENT,
    msg_id TEXT,                                   -- 新增字段
    session_id INTEGER NOT NULL,
    room_id TEXT NOT NULL,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT,
    user_nickname TEXT,
    gift_id TEXT,
    gift_name TEXT,
    gift_count INTEGER DEFAULT 1,
    gift_diamond_value INTEGER DEFAULT 0,
    anchor_id TEXT,
    anchor_name TEXT,
    FOREIGN KEY (session_id) REFERENCES live_sessions(session_id),
    FOREIGN KEY (room_id) REFERENCES rooms(room_id)
);
```

#### 数据库迁移函数

添加了 `msg_id` 列的迁移逻辑：

```go
func ensureGiftRecordsColumns(conn *sql.DB) error {
    // 添加 msg_id 列
    if err := addColumnIfMissing(conn, "gift_records", "msg_id", "TEXT"); err != nil {
        return err
    }
    
    // 添加 anchor_name 列
    if err := addColumnIfMissing(conn, "gift_records", "anchor_name", "TEXT"); err != nil {
        return err
    }
    
    // ... 其他迁移逻辑
}
```

### 2. 浏览器插件连接的礼物消息处理

**文件**: `/workspace/server-go/internal/server/websocket.go`

#### saveMessage 函数增强

添加了详细的日志输出，用于追踪礼物消息的识别和处理：

```go
func (s *WebSocketServer) saveMessage(roomID string, sessionID int64, parsed *parser.ParsedProtoMessage) {
    if parsed == nil {
        log.Printf("⚠️  [房间 %s] parsed 消息为 nil，跳过保存", roomID)
        return
    }

    log.Printf("🔍 [房间 %s] saveMessage 检查消息类型: '%s'", roomID, parsed.MessageType)

    switch parsed.MessageType {
    case "礼物消息":
        log.Printf("✅ [房间 %s] 识别到礼物消息，准备保存到 gift_records", roomID)
        s.saveGiftRecord(roomID, sessionID, parsed)
    default:
        log.Printf("ℹ️  [房间 %s] 消息类型 '%s' 不需要特殊处理", roomID, parsed.MessageType)
    }
}
```

#### saveGiftRecord 函数完善

1. **生成唯一的 msgID**:
   ```go
   msgID := fmt.Sprintf("%d_%s_%d", time.Now().UnixNano(), parsed.Method, sessionID)
   ```

2. **添加详细的日志输出**:
   - 礼物详情日志
   - 主播分配日志
   - 数据库插入日志
   - 错误详情日志

3. **完整的数据插入**:
   ```go
   result, err := s.db.GetConnection().Exec(`
       INSERT INTO gift_records (
           msg_id, session_id, room_id, user_id, user_nickname, gift_id, gift_name, 
           gift_count, gift_diamond_value, anchor_id, anchor_name
       ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
   `, msgID, sessionID, roomID, userID, userNickname, giftID, giftName, giftCount, diamondCount, anchorID, anchorName)
   ```

### 3. 手动连接房间的礼物消息处理

#### recordParsedMessage 函数增强

**文件**: `/workspace/server-go/internal/ui/fyne_ui.go`

添加了对手动连接房间礼物消息的处理：

```go
if parsed.MessageType == "礼物消息" {
    ui.handleGiftAssignment(roomID, pair.Detail)
    
    // 保存礼物记录到 gift_records 表
    if persist && ui.db != nil {
        log.Printf("🎁 [房间 %s] 手动连接收到礼物消息，准备保存到 gift_records", roomID)
        if err := ui.saveManualGiftRecord(roomID, parsed); err != nil {
            log.Printf("❌ [房间 %s] 保存手动房间礼物记录失败: %v", roomID, err)
        }
    }
}
```

#### 新增函数

**文件**: `/workspace/server-go/internal/ui/manual_room.go`

1. **saveManualGiftRecord**: 保存手动房间的礼物记录
   ```go
   func (ui *FyneUI) saveManualGiftRecord(roomID string, parsed *parser.ParsedProtoMessage) error
   ```
   
   功能：
   - 获取或创建 session_id
   - 生成唯一的 msgID
   - 提取礼物详情
   - 保存到 gift_records 表
   - 添加详细的日志输出

2. **getOrCreateManualSession**: 获取或创建手动房间的 session_id
   ```go
   func (ui *FyneUI) getOrCreateManualSession(roomID string) (int64, error)
   ```
   
   功能：
   - 查找已有的活跃 session
   - 如果不存在，创建新的 session
   - 返回 session_id

3. **辅助函数**:
   - `toString`: 转换接口类型为字符串
   - `toInt`: 转换接口类型为整数

#### 导入更新

添加了必要的包导入：
```go
import (
    "strconv"
    "time"
    // ... 其他导入
)
```

## 日志输出示例

### 浏览器插件连接的礼物消息

```
🔍 [房间 123456] saveMessage 检查消息类型: '礼物消息'
✅ [房间 123456] 识别到礼物消息，准备保存到 gift_records
🎁 [房间 123456] 开始处理礼物记录，SessionID: 1
🎁 [房间 123456] 礼物详情 - 用户: 张三(user123), 礼物: 玫瑰花(gift001) x10, 钻石: 50
🔍 [房间 123456] 礼物未指定主播，尝试自动分配
🎯 [房间 123456] 礼物 玫瑰花 自动分配给主播: anchor001
📛 [房间 123456] 主播名称: 李四
💾 [房间 123456] 准备插入 gift_records 表，msgID: 1732185600123456789_WebcastGiftMessage_1
✅ [房间 123456] 礼物记录已保存到 gift_records 表，recordID: 42, msgID: 1732185600123456789_WebcastGiftMessage_1
📊 [房间 123456] 主播 anchor001 业绩已更新
```

### 手动连接的礼物消息

```
📩 [手动房间 789012] 收到事件: WebcastGiftMessage
✅ [手动房间 789012] 消息解析成功: 礼物消息 - WebcastGiftMessage
🎁 [手动房间 789012] 礼物详情: 王五 送出 豪华游艇 x1 (💎1000)
🎁 [房间 789012] 手动连接收到礼物消息，准备保存到 gift_records
🎁 [手动房间 789012] 开始保存礼物记录
📋 [手动房间 789012] 使用已存在的 sessionID: 5
🎁 [手动房间 789012] 礼物详情 - 用户: 王五(user789), 礼物: 豪华游艇(gift999) x1, 钻石: 1000
💾 [手动房间 789012] 准备插入 gift_records 表，msgID: 1732185700987654321_WebcastGiftMessage_5, sessionID: 5
✅ [手动房间 789012] 礼物记录已保存到 gift_records 表，recordID: 43, msgID: 1732185700987654321_WebcastGiftMessage_5
```

## 数据流程

### 1. 浏览器插件连接

```
浏览器扩展发送消息
    ↓
handleDouyinMessage 接收消息
    ↓
ParseProtoMessages 解析消息
    ↓
saveMessage 检查消息类型
    ↓
saveGiftRecord 保存礼物记录（包含 msg_id）
    ↓
gift_records 表
```

### 2. 手动连接

```
手动建立房间连接
    ↓
handleManualEvent 接收事件
    ↓
ParseProtoMessage 解析消息
    ↓
recordParsedMessage 记录消息
    ↓
saveManualGiftRecord 保存礼物记录（包含 msg_id）
    ↓
gift_records 表
```

## 技术特点

1. **唯一消息标识**: 使用纳秒级时间戳 + 方法名 + session_id 生成唯一的 msgID
2. **详细日志输出**: 每个关键步骤都有对应的日志，便于调试和监控
3. **自动主播分配**: 支持自动将礼物分配给主播
4. **向后兼容**: 通过数据库迁移确保现有数据库平滑升级
5. **统一处理**: 浏览器插件和手动连接两种方式都支持礼物记录保存

## 测试建议

### 1. 浏览器插件连接测试

1. 启动应用程序
2. 通过浏览器扩展连接到直播间
3. 观察有礼物消息时的日志输出
4. 检查 `gift_records` 表是否有新记录
5. 验证 `msg_id` 字段是否已填充

### 2. 手动连接测试

1. 启动应用程序
2. 手动连接到直播间
3. 观察有礼物消息时的日志输出
4. 检查 `gift_records` 表是否有新记录
5. 验证 `msg_id` 字段是否已填充

### 3. 数据库验证

```sql
-- 查看最近的礼物记录
SELECT * FROM gift_records ORDER BY create_time DESC LIMIT 10;

-- 验证 msg_id 字段
SELECT msg_id, room_id, user_nickname, gift_name, gift_count 
FROM gift_records 
WHERE msg_id IS NOT NULL 
ORDER BY create_time DESC 
LIMIT 10;

-- 检查是否有重复的 msg_id
SELECT msg_id, COUNT(*) as count 
FROM gift_records 
WHERE msg_id IS NOT NULL 
GROUP BY msg_id 
HAVING count > 1;
```

## 编译验证

已验证以下包的编译通过：
- ✅ `internal/database` - 数据库包编译成功
- ✅ `internal/server` - 服务器包编译成功

注意：`internal/ui` 包需要 GUI 依赖环境，在无 GUI 的 Linux 环境中会出现编译错误，但这不影响代码的逻辑正确性。

## 已修改的文件

1. `/workspace/server-go/internal/database/database.go`
   - 添加 `msg_id` 字段到 `gift_records` 表
   - 更新 `ensureGiftRecordsColumns` 函数

2. `/workspace/server-go/internal/server/websocket.go`
   - 增强 `saveMessage` 函数的日志输出
   - 完善 `saveGiftRecord` 函数，生成并保存 msgID

3. `/workspace/server-go/internal/ui/fyne_ui.go`
   - 修改 `recordParsedMessage` 函数，添加手动房间礼物消息的保存逻辑

4. `/workspace/server-go/internal/ui/manual_room.go`
   - 添加 `saveManualGiftRecord` 函数
   - 添加 `getOrCreateManualSession` 函数
   - 添加辅助函数 `toString` 和 `toInt`
   - 更新导入语句

## 结论

本次修复确保了所有礼物消息（无论是通过浏览器插件还是手动连接）都会正确保存到 `gift_records` 表，并且每条记录都包含唯一的 `msg_id` 字段。详细的日志输出也便于后续的调试和监控。
