# Protocol Buffers 定义

本目录包含抖音直播间消息的 Protocol Buffers 定义文件。

## 文件说明

- `douyin.proto` - 抖音直播间所有消息类型的完整定义

## 消息类型

### 核心消息结构
- `PushFrame` - WebSocket 推送帧
- `Response` - 服务器响应
- `Message` - 通用消息包装

### 用户相关
- `User` - 用户完整信息
- `Image` - 图片资源
- `FollowInfo` - 关注信息
- `PayGrade` - 付费等级
- `FansClub` - 粉丝团
- `Border` - 边框
- `AnchorInfo` - 主播信息

### 消息类型
- `ChatMessage` - 聊天消息
- `GiftMessage` - 礼物消息
- `LikeMessage` - 点赞消息
- `MemberMessage` - 成员消息（进入直播间）
- `SocialMessage` - 关注消息
- `RoomUserSeqMessage` - 房间用户序列
- `RoomStatsMessage` - 房间统计
- `ControlMessage` - 控制消息
- `RoomMessage` - 房间消息

## 参考资料

本定义基于以下项目：
1. https://github.com/skmcj/dycast
2. https://github.com/WanGuChou/DouyinBarrageGrab

## 使用说明

当前项目使用手动解析 Protobuf，不需要使用 protoc 生成代码。
这些 .proto 文件作为参考文档，帮助理解消息结构。

如果需要使用 protoc 生成 Go 代码：
```bash
protoc --go_out=. --go_opt=paths=source_relative douyin.proto
```

## 字段编号对照

### GiftMessage 关键字段
- field 1: common
- field 2: giftId
- field 7: user
- field 8: toUser
- field 13: diamondCount
- field 14: giftName
- field 15: gift (GiftStruct)

### GiftStruct 关键字段
- field 1: icon
- field 5: id
- field 11: diamondCount
- field 15: name

### User 关键字段
- field 1: id
- field 2: shortId
- field 3: nickname
- field 9: avatarThumb
- field 22: followInfo
- field 23: payGrade
- field 24: fansClub

## 注意事项

1. 所有字段编号必须与抖音实际协议保持一致
2. 嵌套消息需要正确处理 Length-delimited 编码
3. 字符串使用 UTF-8 编码
4. 数值类型使用 Varint 编码
