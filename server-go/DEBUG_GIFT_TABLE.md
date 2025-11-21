# 礼物表格调试指南

## 问题现象
gift_records 表中有数据，但 roomTab.GiftTable 页面没有显示数据。

## 已添加的调试日志

现在程序会输出详细的调试日志来追踪问题，包括：

### 1. 初始化时的日志
```
🏗️  [房间 123456] 初始化礼物表格
📊 [房间 123456] 开始加载礼物记录
🔍 [房间 123456] 执行查询: WHERE room_id = '123456'
✅ [房间 123456] 加载了 5 条礼物记录（包含表头共 6 行）
📊 [房间 123456] 初始化时加载了 6 行数据
📐 [房间 123456] 表格尺寸: 6 行 x 6 列
✅ [房间 123456] 礼物表格初始化完成
```

### 2. 刷新时的日志
```
🔄 [房间 123456] refreshRoomTables 开始刷新表格
📊 [房间 123456] 开始加载礼物记录
🔍 [房间 123456] 执行查询: WHERE room_id = '123456'
✅ [房间 123456] 加载了 5 条礼物记录（包含表头共 6 行）
📊 [房间 123456] GiftRows 更新完成，当前行数: 6
🔄 [房间 123456] 刷新 GiftTable UI
✅ [房间 123456] refreshRoomTables 完成
```

### 3. 可能的错误日志
```
⚠️  [房间 123456] 数据库连接为空
❌ [房间 123456] 查询礼物记录失败: sql error message
⚠️  [房间 123456] 扫描记录失败: scan error
⚠️  [房间 123456] GiftRows 为空，返回 0 行
⚠️  [房间 123456] GiftTable 为 nil，无法刷新
```

## 调试步骤

### 步骤 1: 检查数据库中的实际数据

运行以下 SQL 查询来查看数据库中的礼物记录：

```sql
-- 查看所有房间的礼物记录统计
SELECT room_id, COUNT(*) as count 
FROM gift_records 
GROUP BY room_id 
ORDER BY count DESC;

-- 查看特定房间的礼物记录
SELECT 
    msg_id,
    room_id,
    user_nickname,
    gift_name,
    gift_count,
    gift_diamond_value,
    anchor_name,
    create_time
FROM gift_records 
WHERE room_id = 'YOUR_ROOM_ID'  -- 替换为实际的房间ID
ORDER BY create_time DESC 
LIMIT 10;

-- 查看所有 room_id 的唯一值
SELECT DISTINCT room_id FROM gift_records ORDER BY room_id;
```

### 步骤 2: 启动程序并观察日志

1. 启动 server-go
2. 连接到直播间（手动或浏览器插件）
3. 等待收到礼物消息
4. 观察控制台输出的日志

**关键检查点**：
- 是否看到 "🏗️ 初始化礼物表格" 日志？
- room_id 的值是什么？
- 查询返回了多少条记录？
- GiftRows 的行数是多少？
- 是否有任何错误或警告日志？

### 步骤 3: 对比 room_id

**重要：room_id 必须完全匹配！**

检查以下几点：

1. **数据库中保存的 room_id**：
```sql
SELECT DISTINCT room_id FROM gift_records;
```

2. **UI 中使用的 room_id**：
查看日志中的 `WHERE room_id = 'xxx'`

3. **可能的不匹配原因**：
   - 数据库中的 room_id 有前缀或后缀（如 "room_123456"）
   - UI 中的 room_id 有前缀或后缀
   - 大小写不同
   - 空格或特殊字符

### 步骤 4: 验证保存和查询的 room_id 一致性

#### 保存时的日志（应该看到）：
```
✅ [手动房间 123456] 礼物记录已保存到 gift_records 表，recordID: 42, msgID: xxx
```
或
```
✅ [房间 123456] 礼物记录已保存到 gift_records 表，recordID: 43, msgID: xxx
```

#### 查询时的日志（应该看到）：
```
🔍 [房间 123456] 执行查询: WHERE room_id = '123456'
```

**确认两个日志中的 room_id 完全一致！**

## 常见问题和解决方案

### 问题 1: 加载了 0 条记录

**日志示例**：
```
✅ [房间 123456] 加载了 0 条礼物记录（包含表头共 1 行）
```

**可能原因**：
1. 数据库中确实没有该 room_id 的记录
2. room_id 不匹配

**解决方法**：
```sql
-- 检查数据库中是否有任何礼物记录
SELECT COUNT(*) FROM gift_records;

-- 检查是否有该 room_id 的记录
SELECT COUNT(*) FROM gift_records WHERE room_id = 'YOUR_ROOM_ID';

-- 查看实际保存的 room_id
SELECT DISTINCT room_id FROM gift_records;
```

### 问题 2: 数据库连接为空

**日志示例**：
```
⚠️  [房间 123456] 数据库连接为空
```

**原因**：UI 层没有正确获取数据库连接

**解决方法**：
检查 `FyneUI` 初始化时是否正确设置了 `db` 字段。

### 问题 3: GiftTable 为 nil

**日志示例**：
```
⚠️  [房间 123456] GiftTable 为 nil，无法刷新
```

**原因**：表格还没有初始化就尝试刷新

**解决方法**：
确保在刷新之前先调用了 `initRoomGiftTable`。

### 问题 4: 查询失败

**日志示例**：
```
❌ [房间 123456] 查询礼物记录失败: sql error message
```

**原因**：SQL 查询出错

**解决方法**：
1. 检查数据库表结构是否正确
2. 确认字段名称（create_time, timestamp, gift_name 等）

### 问题 5: room_id 格式不一致

**场景**：
- 保存时：`room_id = "7404883888"`
- 查询时：`WHERE room_id = '7404883888'`
- 但数据库中实际是：`room_id = "room_7404883888"`

**解决方法**：
需要统一 room_id 的格式。检查以下位置：
1. `ensureRoomRecord` (websocket.go)
2. `ensureManualRoomRecord` (manual_room.go)
3. `saveGiftRecord` (websocket.go)
4. `saveManualGiftRecord` (manual_room.go)

## 测试用 SQL

### 手动插入测试数据
```sql
-- 插入测试礼物记录
INSERT INTO gift_records (
    msg_id, session_id, room_id, user_id, user_nickname, 
    gift_id, gift_name, gift_count, gift_diamond_value, 
    anchor_id, anchor_name, create_time
) VALUES (
    'test_msg_001', 1, 'YOUR_ROOM_ID', 'user001', '测试用户',
    'gift001', '测试礼物', 5, 100,
    'anchor001', '测试主播', datetime('now')
);

-- 验证插入
SELECT * FROM gift_records WHERE room_id = 'YOUR_ROOM_ID';
```

### 清理测试数据
```sql
-- 删除测试数据
DELETE FROM gift_records WHERE msg_id = 'test_msg_001';
```

## 预期的正常流程

### 1. 初始化房间
```
🏗️  [房间 7404883888] 初始化礼物表格
📊 [房间 7404883888] 开始加载礼物记录
🔍 [房间 7404883888] 执行查询: WHERE room_id = '7404883888'
✅ [房间 7404883888] 加载了 0 条礼物记录（包含表头共 1 行）
📊 [房间 7404883888] 初始化时加载了 1 行数据
📐 [房间 7404883888] 表格尺寸: 1 行 x 6 列
✅ [房间 7404883888] 礼物表格初始化完成
```

### 2. 收到礼物消息
```
🎁 [房间 7404883888] 收到礼物消息: 张三 送出 玫瑰花 x10 (价值 50 钻石)
💾 [房间 7404883888] 准备插入 gift_records 表，msgID: 1732185600123456789_WebcastGiftMessage_1
✅ [房间 7404883888] 礼物记录已保存到 gift_records 表，recordID: 1, msgID: 1732185600123456789_WebcastGiftMessage_1
```

### 3. 刷新表格
```
🔄 [房间 7404883888] 浏览器插件礼物消息，刷新礼物表格
🔄 [房间 7404883888] refreshRoomTables 开始刷新表格
📊 [房间 7404883888] 开始加载礼物记录
🔍 [房间 7404883888] 执行查询: WHERE room_id = '7404883888'
✅ [房间 7404883888] 加载了 1 条礼物记录（包含表头共 2 行）
📊 [房间 7404883888] GiftRows 更新完成，当前行数: 2
🔄 [房间 7404883888] 刷新 GiftTable UI
✅ [房间 7404883888] refreshRoomTables 完成
```

## 下一步

1. **启动程序**，观察日志输出
2. **记录所有相关日志**，特别是：
   - 保存礼物记录时的 room_id
   - 查询礼物记录时的 room_id
   - 返回的记录数量
3. **执行 SQL 查询**，确认数据库中的实际数据
4. **对比 room_id**，确保保存和查询使用的是同一个值

如果仍然有问题，请提供：
- 完整的相关日志输出
- SQL 查询结果
- room_id 的实际值
