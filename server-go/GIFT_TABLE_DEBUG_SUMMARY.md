# 礼物表格调试 - 修改总结

## 修改日期
2025-11-21

## 问题
gift_records 表中有数据，但 roomTab.GiftTable 页面仍然没有显示数据。

## 本次修改

### 添加了详细的调试日志

#### 1. loadRoomGiftRows 函数 (3529-3585行)

**添加的日志**：
- 📊 开始加载礼物记录
- ⚠️  数据库连接为空（如果 db 为 nil）
- 🔍 执行查询（显示 room_id）
- ❌ 查询失败（显示错误信息）
- ⚠️  扫描记录失败（显示错误信息）
- ✅ 加载完成（显示记录数量和总行数）

**关键输出**：
```go
log.Printf("📊 [房间 %s] 开始加载礼物记录", roomID)
log.Printf("🔍 [房间 %s] 执行查询: WHERE room_id = '%s'", roomID, roomID)
log.Printf("✅ [房间 %s] 加载了 %d 条礼物记录（包含表头共 %d 行）", roomID, recordCount, len(rows))
```

#### 2. refreshRoomTables 函数 (2927-2952行)

**添加的日志**：
- 🔄 开始刷新表格
- 📊 GiftRows 更新完成（显示行数）
- 🔄 刷新 GiftTable UI
- ⚠️  GiftTable 为 nil（无法刷新）
- ✅ 刷新完成

**关键输出**：
```go
log.Printf("🔄 [房间 %s] refreshRoomTables 开始刷新表格", roomTab.RoomID)
log.Printf("📊 [房间 %s] GiftRows 更新完成，当前行数: %d", roomTab.RoomID, len(roomTab.GiftRows))
log.Printf("✅ [房间 %s] refreshRoomTables 完成", roomTab.RoomID)
```

#### 3. initRoomGiftTable 函数 (2954-2989行)

**添加的日志**：
- 🏗️  初始化礼物表格
- 📊 初始化时加载的行数
- ⚠️  GiftRows 为空（如果没有数据）
- 📐 表格尺寸（行数 x 列数）
- ✅ 初始化完成

**关键输出**：
```go
log.Printf("🏗️  [房间 %s] 初始化礼物表格", roomTab.RoomID)
log.Printf("📊 [房间 %s] 初始化时加载了 %d 行数据", roomTab.RoomID, len(roomTab.GiftRows))
log.Printf("📐 [房间 %s] 表格尺寸: %d 行 x %d 列", roomTab.RoomID, rows, cols)
```

## 调试流程

### 正常流程的日志输出

#### 步骤 1: 初始化房间
```
🏗️  [房间 7404883888] 初始化礼物表格
📊 [房间 7404883888] 开始加载礼物记录
🔍 [房间 7404883888] 执行查询: WHERE room_id = '7404883888'
✅ [房间 7404883888] 加载了 0 条礼物记录（包含表头共 1 行）
📊 [房间 7404883888] 初始化时加载了 1 行数据
📐 [房间 7404883888] 表格尺寸: 1 行 x 6 列
✅ [房间 7404883888] 礼物表格初始化完成
```

#### 步骤 2: 收到并保存礼物消息
```
🎁 [房间 7404883888] 开始处理礼物记录，SessionID: 1
🎁 [房间 7404883888] 礼物详情 - 用户: 张三(user123), 礼物: 玫瑰花(gift001) x10, 钻石: 50
💾 [房间 7404883888] 准备插入 gift_records 表，msgID: 1732185600123456789_WebcastGiftMessage_1
✅ [房间 7404883888] 礼物记录已保存到 gift_records 表，recordID: 1, msgID: 1732185600123456789_WebcastGiftMessage_1
```

#### 步骤 3: 刷新表格
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

### 异常情况的日志

#### 情况 1: 数据库连接为空
```
📊 [房间 7404883888] 开始加载礼物记录
⚠️  [房间 7404883888] 数据库连接为空
```

#### 情况 2: 查询失败
```
📊 [房间 7404883888] 开始加载礼物记录
🔍 [房间 7404883888] 执行查询: WHERE room_id = '7404883888'
❌ [房间 7404883888] 查询礼物记录失败: no such table: gift_records
```

#### 情况 3: 没有匹配的记录
```
📊 [房间 7404883888] 开始加载礼物记录
🔍 [房间 7404883888] 执行查询: WHERE room_id = '7404883888'
✅ [房间 7404883888] 加载了 0 条礼物记录（包含表头共 1 行）
```

#### 情况 4: GiftTable 为 nil
```
🔄 [房间 7404883888] refreshRoomTables 开始刷新表格
📊 [房间 7404883888] GiftRows 更新完成，当前行数: 5
⚠️  [房间 7404883888] GiftTable 为 nil，无法刷新
```

## 关键检查点

### 1. room_id 一致性

**保存时**（从日志中查找）：
```
✅ [房间 7404883888] 礼物记录已保存到 gift_records 表
```

**查询时**（从日志中查找）：
```
🔍 [房间 7404883888] 执行查询: WHERE room_id = '7404883888'
```

**数据库中**（执行 SQL）：
```sql
SELECT DISTINCT room_id FROM gift_records;
```

**三者必须完全一致！**

### 2. 数据库查询结果

查看日志中的：
```
✅ [房间 7404883888] 加载了 X 条礼物记录
```

- 如果 X = 0，说明数据库中没有该 room_id 的记录
- 如果 X > 0，说明数据存在，可能是 UI 刷新的问题

### 3. 表格状态

查看日志中的：
```
📊 [房间 7404883888] GiftRows 更新完成，当前行数: X
```

- 如果 X = 1，只有表头，没有数据
- 如果 X > 1，有数据，检查表格是否正确刷新

## 辅助工具

### 1. DEBUG_GIFT_TABLE.md
详细的调试指南，包含：
- 所有可能的日志输出说明
- 常见问题和解决方案
- 完整的调试步骤

### 2. check_gift_records.sql
SQL 调试脚本，包含：
- 检查表结构
- 统计记录数量
- 查看 room_id 分布
- 对比不同表的 room_id

## 使用方法

### 1. 启动程序
```bash
cd server-go
go run main.go
```

### 2. 观察日志输出
- 查找所有带 emoji 图标的日志（📊 🔍 ✅ ⚠️ ❌）
- 特别关注 room_id 的值
- 记录加载的记录数量

### 3. 执行 SQL 查询
```bash
sqlite3 your_database.db < check_gift_records.sql
```

### 4. 对比分析
- 对比保存和查询的 room_id
- 确认数据库中的记录数量
- 检查表格是否正确初始化

## 可能的问题和解决方案

### 问题 1: room_id 格式不一致

**症状**：
- 保存时：`room_id = "7404883888"`
- 查询时：`WHERE room_id = '7404883888'`
- 数据库：实际存储的是 `"room_7404883888"`

**解决**：
需要统一所有地方的 room_id 格式。

### 问题 2: 数据保存在不同的 room_id 下

**症状**：
- 浏览器插件保存为：`room_id = "7404883888"`
- 手动连接保存为：`room_id = "manual_7404883888"`

**解决**：
确保两种连接方式使用相同的 room_id。

### 问题 3: 表格初始化时机问题

**症状**：
- `GiftTable` 为 nil
- 无法刷新

**解决**：
确保在调用 `refreshRoomTables` 之前已经调用了 `initRoomGiftTable`。

## 修改的文件

1. `/workspace/server-go/internal/ui/fyne_ui.go`
   - `loadRoomGiftRows` (3529-3585行) - 添加详细日志
   - `refreshRoomTables` (2927-2952行) - 添加详细日志
   - `initRoomGiftTable` (2954-2989行) - 添加详细日志

## 创建的文件

1. `/workspace/server-go/DEBUG_GIFT_TABLE.md` - 详细调试指南
2. `/workspace/server-go/check_gift_records.sql` - SQL 调试脚本
3. `/workspace/server-go/GIFT_TABLE_DEBUG_SUMMARY.md` - 本文档

## 下一步

1. **运行程序**，观察新增的日志输出
2. **记录关键信息**：
   - 保存时的 room_id
   - 查询时的 room_id
   - 加载的记录数量
   - 任何错误或警告
3. **执行 SQL 查询**，确认数据库状态
4. **提供反馈**：
   - 相关日志输出
   - SQL 查询结果
   - room_id 的实际值

根据这些信息，我们可以精确定位问题并提供针对性的解决方案。
