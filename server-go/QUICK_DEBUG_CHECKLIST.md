# 礼物表格不显示 - 快速调试清单

## 🎯 问题
gift_records 表中有数据，但 UI 上看不到。

## ✅ 已完成的修改

1. ✅ 添加了详细的调试日志到 `loadRoomGiftRows`
2. ✅ 添加了详细的调试日志到 `refreshRoomTables`
3. ✅ 添加了详细的调试日志到 `initRoomGiftTable`
4. ✅ 在礼物消息保存后自动刷新表格

## 🔍 快速调试步骤

### 步骤 1: 启动程序并观察日志（30 秒）

启动 server-go 后，在控制台中查找：

```
🏗️  [房间 XXXXX] 初始化礼物表格
📊 [房间 XXXXX] 开始加载礼物记录
🔍 [房间 XXXXX] 执行查询: WHERE room_id = 'XXXXX'
```

**记录下 room_id 的值！**

### 步骤 2: 检查数据库（1 分钟）

打开数据库（通常在 server-go 目录下），执行：

```sql
-- 快速检查
SELECT DISTINCT room_id FROM gift_records;
```

**对比这个 room_id 是否和步骤 1 中的一致！**

### 步骤 3: 查看具体记录（30 秒）

```sql
-- 将 'YOUR_ROOM_ID' 替换为步骤 1 中看到的 room_id
SELECT COUNT(*) FROM gift_records WHERE room_id = 'YOUR_ROOM_ID';
```

**如果返回 0，说明 room_id 不匹配！**

## 🎯 最可能的问题

### 问题 A: room_id 不匹配 (90% 可能)

**症状**：
- 日志显示：`WHERE room_id = '7404883888'`
- 数据库中：`SELECT DISTINCT room_id` 返回其他值

**原因**：
- 保存时用了一个 room_id
- 查询时用了另一个 room_id

**解决**：
查看保存时的日志：
```
✅ [房间 XXXXX] 礼物记录已保存到 gift_records 表
```
和查询时的日志：
```
🔍 [房间 XXXXX] 执行查询: WHERE room_id = 'XXXXX'
```
确保两个 XXXXX 完全一致！

---

### 问题 B: 数据库中确实没有数据 (5% 可能)

**症状**：
```sql
SELECT COUNT(*) FROM gift_records;
-- 返回 0
```

**原因**：
礼物消息没有正确保存。

**解决**：
查看保存时的日志，应该看到：
```
✅ [房间 XXXXX] 礼物记录已保存到 gift_records 表，recordID: X
```

如果没有看到，说明保存失败。

---

### 问题 C: 表格没有刷新 (3% 可能)

**症状**：
- 日志显示加载了记录：`✅ [房间 XXXXX] 加载了 5 条礼物记录`
- 但 UI 上看不到

**原因**：
表格 UI 没有刷新。

**解决**：
查看日志中是否有：
```
🔄 [房间 XXXXX] 刷新 GiftTable UI
```

如果有 `⚠️  [房间 XXXXX] GiftTable 为 nil`，说明表格未初始化。

---

### 问题 D: 数据库连接问题 (2% 可能)

**症状**：
```
⚠️  [房间 XXXXX] 数据库连接为空
```

**原因**：
UI 层没有正确获取数据库连接。

**解决**：
检查程序启动日志，确认数据库已正确初始化。

## 📋 完整的日志检查清单

收到礼物消息时，应该看到以下完整日志序列：

```
# 1. 保存礼物消息
🎁 [房间 7404883888] 开始处理礼物记录，SessionID: 1
💾 [房间 7404883888] 准备插入 gift_records 表，msgID: xxx
✅ [房间 7404883888] 礼物记录已保存到 gift_records 表，recordID: 1

# 2. 触发刷新
🔄 [房间 7404883888] 浏览器插件礼物消息，刷新礼物表格
（或）
🔄 [房间 7404883888] 刷新礼物表格

# 3. 刷新表格
🔄 [房间 7404883888] refreshRoomTables 开始刷新表格

# 4. 加载数据
📊 [房间 7404883888] 开始加载礼物记录
🔍 [房间 7404883888] 执行查询: WHERE room_id = '7404883888'
✅ [房间 7404883888] 加载了 1 条礼物记录（包含表头共 2 行）

# 5. 更新 UI
📊 [房间 7404883888] GiftRows 更新完成，当前行数: 2
🔄 [房间 7404883888] 刷新 GiftTable UI
✅ [房间 7404883888] refreshRoomTables 完成
```

**如果缺少任何一步，说明流程有问题！**

## 🔧 快速测试方法

### 手动插入测试数据

```sql
-- 使用你在日志中看到的 room_id
INSERT INTO gift_records (
    msg_id, session_id, room_id, user_nickname, 
    gift_name, gift_count, gift_diamond_value
) VALUES (
    'test_001', 1, '7404883888', '测试用户',
    '测试礼物', 10, 100
);

-- 验证
SELECT * FROM gift_records WHERE msg_id = 'test_001';
```

然后在程序中：
1. 切换到该房间的"礼物记录" Tab
2. 点击"刷新"按钮（如果有）
3. 观察是否显示

如果显示了测试数据，说明问题是 **room_id 不匹配**！

## 📞 需要帮助？

提供以下信息：

1. **保存时的日志**（查找包含 "礼物记录已保存" 的行）
2. **查询时的日志**（查找包含 "执行查询" 的行）
3. **SQL 查询结果**：
```sql
SELECT DISTINCT room_id FROM gift_records;
```
4. **记录数量**：
```sql
SELECT COUNT(*) FROM gift_records;
```

## 📄 详细文档

- `DEBUG_GIFT_TABLE.md` - 完整调试指南
- `check_gift_records.sql` - SQL 调试脚本
- `GIFT_TABLE_DEBUG_SUMMARY.md` - 技术总结

## 🎉 预期结果

修复后，收到礼物消息时应该：
1. ✅ 看到保存日志
2. ✅ 看到刷新日志
3. ✅ 看到加载日志（显示记录数量）
4. ✅ UI 上实时显示礼物记录

## ⏱️ 预计调试时间

- 如果是 room_id 不匹配：**5-10 分钟**
- 如果是其他问题：**10-30 分钟**

现在可以开始测试了！🚀
