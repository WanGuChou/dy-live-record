# 礼物表格显示修复文档

## 修复日期
2025-11-21

## 问题描述

`gift_records` 表中有数据，但是 `roomTab.GiftTable` 页面没有显示数据。用户在"礼物记录" Tab 页看不到任何礼物记录。

## 问题分析

### 数据流程

#### 1. 浏览器插件连接流程
```
浏览器扩展发送消息
  ↓
handleDouyinMessage (websocket.go)
  ↓
saveMessage → saveGiftRecord (保存到 gift_records)
  ↓
AddParsedMessageWithDetail (通知 UI)
  ↓
recordParsedMessage (persist=false, 不再保存)
  ❌ 没有刷新礼物表格！
```

#### 2. 手动连接流程
```
手动建立房间连接
  ↓
handleManualEvent (manual_room.go)
  ↓
recordParsedMessage (persist=true)
  ↓
saveManualGiftRecord (保存到 gift_records)
  ❌ 没有刷新礼物表格！
```

### 根本原因

**礼物记录保存到数据库后，没有调用 `refreshRoomTables()` 刷新 UI 表格，导致虽然数据在数据库中，但界面上看不到。**

## 解决方案

### 1. 修复手动连接的礼物表格刷新

**文件**: `/workspace/server-go/internal/ui/fyne_ui.go`  
**位置**: `recordParsedMessage` 函数（2676-2692行）

在保存手动房间礼物记录成功后，立即刷新表格：

```go
if parsed.MessageType == "礼物消息" {
    ui.handleGiftAssignment(roomID, pair.Detail)

    // 保存礼物记录到 gift_records 表
    if persist && ui.db != nil {
        log.Printf("🎁 [房间 %s] 手动连接收到礼物消息，准备保存到 gift_records", roomID)
        if err := ui.saveManualGiftRecord(roomID, parsed); err != nil {
            log.Printf("❌ [房间 %s] 保存手动房间礼物记录失败: %v", roomID, err)
        } else {
            // 保存成功后刷新礼物表格 ✅ 新增
            if roomTab, ok := ui.roomTabs[roomID]; ok {
                log.Printf("🔄 [房间 %s] 刷新礼物表格", roomID)
                ui.refreshRoomTables(roomTab)
            }
        }
    }
}
```

### 2. 修复浏览器插件连接的礼物表格刷新

**文件**: `/workspace/server-go/internal/ui/fyne_ui.go`  
**位置**: `AddParsedMessageWithDetail` 函数（2550-2593行）

在接收到礼物消息通知后，刷新表格：

```go
func (ui *FyneUI) AddParsedMessageWithDetail(roomID string, message string, detail map[string]interface{}) {
    if detail != nil {
        if parsed, ok := detail["_parsed"].(*parser.ParsedProtoMessage); ok {
            ui.recordParsedMessage(roomID, parsed, false)
            // 如果是礼物消息，刷新礼物表格（因为 WebSocket 已经保存到数据库了）✅ 新增
            if parsed.MessageType == "礼物消息" {
                if roomTab, ok := ui.roomTabs[roomID]; ok {
                    log.Printf("🔄 [房间 %s] 浏览器插件礼物消息，刷新礼物表格", roomID)
                    ui.refreshRoomTables(roomTab)
                }
            }
            return
        }
    }
    
    // ... 其他处理 ...
    
    ui.recordParsedMessage(roomID, parsed, false)
    
    // 如果是礼物消息，刷新礼物表格 ✅ 新增
    if msgType == "礼物消息" {
        if roomTab, ok := ui.roomTabs[roomID]; ok {
            log.Printf("🔄 [房间 %s] 浏览器插件礼物消息，刷新礼物表格", roomID)
            ui.refreshRoomTables(roomTab)
        }
    }
}
```

## refreshRoomTables 函数说明

**文件**: `/workspace/server-go/internal/ui/fyne_ui.go`  
**位置**: 2906-2921行

此函数会重新从数据库加载数据并刷新所有相关表格：

```go
func (ui *FyneUI) refreshRoomTables(roomTab *RoomTab) {
    roomTab.GiftRows = ui.loadRoomGiftRows(roomTab.RoomID)        // 重新加载礼物数据
    roomTab.AnchorRows = ui.loadRoomAnchorRows(roomTab.RoomID)    // 重新加载主播数据
    roomTab.SegmentRows = ui.loadRoomSegmentRows(roomTab.RoomID)  // 重新加载分段数据

    if roomTab.GiftTable != nil {
        roomTab.GiftTable.Refresh()      // 刷新礼物表格 UI
    }
    if roomTab.AnchorTable != nil {
        roomTab.AnchorTable.Refresh()    // 刷新主播表格 UI
    }
    if roomTab.SegmentTable != nil {
        roomTab.SegmentTable.Refresh()   // 刷新分段表格 UI
    }
    ui.refreshRoomAnchorPicker(roomTab)
}
```

## 修复后的数据流程

### 1. 浏览器插件连接（修复后）
```
浏览器扩展发送消息
  ↓
handleDouyinMessage (websocket.go)
  ↓
saveMessage → saveGiftRecord (保存到 gift_records) ✅
  ↓
AddParsedMessageWithDetail (通知 UI)
  ↓
recordParsedMessage (persist=false)
  ↓
检测到礼物消息 → refreshRoomTables() ✅ 新增
  ↓
loadRoomGiftRows (从数据库加载)
  ↓
GiftTable.Refresh() (刷新 UI)
  ↓
✅ 用户看到礼物记录！
```

### 2. 手动连接（修复后）
```
手动建立房间连接
  ↓
handleManualEvent (manual_room.go)
  ↓
recordParsedMessage (persist=true)
  ↓
saveManualGiftRecord (保存到 gift_records) ✅
  ↓
保存成功 → refreshRoomTables() ✅ 新增
  ↓
loadRoomGiftRows (从数据库加载)
  ↓
GiftTable.Refresh() (刷新 UI)
  ↓
✅ 用户看到礼物记录！
```

## 日志输出

修复后，收到礼物消息时会看到以下日志：

### 手动连接
```
🎁 [手动房间 123456] 手动连接收到礼物消息，准备保存到 gift_records
🎁 [手动房间 123456] 开始保存礼物记录
💾 [手动房间 123456] 准备插入 gift_records 表，msgID: 1732185600123456789_WebcastGiftMessage_5, sessionID: 5
✅ [手动房间 123456] 礼物记录已保存到 gift_records 表，recordID: 42, msgID: 1732185600123456789_WebcastGiftMessage_5
🔄 [房间 123456] 刷新礼物表格
```

### 浏览器插件连接
```
🔍 [房间 123456] saveMessage 检查消息类型: '礼物消息'
✅ [房间 123456] 识别到礼物消息，准备保存到 gift_records
🎁 [房间 123456] 开始处理礼物记录，SessionID: 1
💾 [房间 123456] 准备插入 gift_records 表，msgID: 1732185700987654321_WebcastGiftMessage_1
✅ [房间 123456] 礼物记录已保存到 gift_records 表，recordID: 43, msgID: 1732185700987654321_WebcastGiftMessage_1
🔄 [房间 123456] 浏览器插件礼物消息，刷新礼物表格
```

## 技术细节

### 为什么需要在两个地方刷新？

1. **手动连接** (`recordParsedMessage` with `persist=true`)
   - 直接在 UI 层保存数据
   - 需要在保存成功后立即刷新

2. **浏览器插件连接** (`AddParsedMessageWithDetail`)
   - WebSocket 层已经保存数据（`saveGiftRecord`）
   - UI 层只是接收通知，不再保存
   - 需要在接收通知后刷新表格

### refreshRoomTables 的作用

- 重新从数据库加载最新数据（`loadRoomGiftRows`）
- 更新内存中的数据行（`GiftRows`）
- 通知 Fyne 表格控件刷新 UI（`GiftTable.Refresh()`）
- 同时刷新主播和分段数据（保持一致性）

### 为什么之前没有显示？

表格只在以下情况更新：
1. 初始创建时（`initRoomGiftTable`）
2. 手动点击"刷新"按钮
3. 其他操作触发 `refreshRoomTables`

但是收到新礼物消息时，并没有触发刷新，导致：
- ✅ 数据已保存到数据库
- ❌ UI 表格没有更新
- ❌ 用户看不到新数据

## 验证方法

### 1. 测试手动连接
1. 启动应用程序
2. 手动连接到直播间
3. 等待收到礼物消息
4. 观察日志输出（应看到刷新日志）
5. 切换到"礼物记录" Tab
6. ✅ 应该能看到礼物记录

### 2. 测试浏览器插件连接
1. 启动应用程序
2. 通过浏览器扩展连接到直播间
3. 等待收到礼物消息
4. 观察日志输出（应看到刷新日志）
5. 切换到"礼物记录" Tab
6. ✅ 应该能看到礼物记录

### 3. 验证数据库
```sql
-- 查看礼物记录
SELECT msg_id, room_id, user_nickname, gift_name, gift_count, create_time 
FROM gift_records 
ORDER BY create_time DESC 
LIMIT 10;
```

## 修改的文件

1. `/workspace/server-go/internal/ui/fyne_ui.go`
   - 修改 `recordParsedMessage` 函数（添加手动连接的刷新逻辑）
   - 修改 `AddParsedMessageWithDetail` 函数（添加浏览器插件的刷新逻辑）

## 性能考虑

- `refreshRoomTables` 会重新查询数据库（限制 200 条）
- 只在收到礼物消息时触发，不会频繁刷新
- 刷新操作在 UI 线程执行，Fyne 会自动处理
- 对用户体验影响极小

## 总结

通过在两个关键位置添加 `refreshRoomTables()` 调用：
1. ✅ 手动连接保存礼物记录后
2. ✅ 浏览器插件接收礼物消息通知后

确保礼物表格能够实时显示最新的数据库记录，解决了"有数据但不显示"的问题。
