# WebSocket 消息解析改进文档

## 概述
本次更新对 server-go 目录下的代码进行了全面改进，主要包括：
1. UI交互改进：原始消息与解析消息的关联显示
2. 消息解析逻辑增强：支持更多消息类型和字段
3. 详情查看功能：点击消息可查看完整的解析内容

## 1. UI交互改进

### 1.1 消息对（MessagePair）机制
新增 `MessagePair` 结构体，用于关联原始消息和解析后的消息：

```go
type MessagePair struct {
    RawMessage    string                 // 原始消息文本
    ParsedMessage string                 // 解析后的简要显示
    ParsedDetail  map[string]interface{} // 完整的解析数据
    Timestamp     time.Time              // 消息时间戳
}
```

### 1.2 交互功能
- **点击原始消息**：自动选中并滚动到对应的解析消息
- **点击解析消息**：弹出详情窗口，显示完整的消息内容
- **详情窗口功能**：
  - 显示原始消息、解析消息和所有解析字段
  - 支持复制到剪贴板
  - 良好的格式化显示

### 1.3 改进的方法
- `AddRawMessage()` - 添加原始消息时创建消息对
- `AddParsedMessageWithDetail()` - 添加解析消息时保存完整详情
- `showMessageDetail()` - 显示消息详情对话框

## 2. 消息解析增强

### 2.1 新增消息类型支持

#### 直播间相关
- `WebcastRoomMessage` - 直播间消息（开播/下播通知）
- `WebcastRoomStatsMessage` - 直播间统计
- `WebcastRoomUserSeqMessage` - 在线人数
- `WebcastRoomRankMessage` - 房间排行榜

#### PK/连麦相关
- `WebcastMatchAgainstScoreMessage` - PK消息（比分、状态）
- `WebcastLinkMicMessage` - 连麦消息
- `WebcastLinkMicBattle` - 连麦PK
- `WebcastLinkMicArmies` - 连麦军团

#### 榜单相关
- `WebcastRankUpdateMessage` - 榜单更新

#### 其他消息
- `WebcastInRoomBannerMessage` - 房间横幅
- `WebcastProductChangeMessage` - 商品变化
- `WebcastCommonTextMessage` - 通用文本消息
- `WebcastBarrageMessage` - 弹幕消息
- `WebcastFansclubMessage` - 粉丝团消息
- `WebcastEmojiChatMessage` - 表情消息

### 2.2 礼物消息解析增强

#### 新增字段支持
```go
// 基础信息
- giftId          // 礼物ID
- giftName        // 礼物名称
- diamondCount    // 钻石价值

// 发送者/接收者
- user            // 发送者信息
- userId          // 发送者ID
- userLevel       // 发送者等级
- toUser          // 接收者（用于礼物PK）
- toUserId        // 接收者ID

// 数量信息（支持多种字段）
- groupCount      // 礼物数量（优先级最高）
- repeatCount     // 重复次数
- comboCount      // 连击次数
- giftCount       // 最终计算的数量

// 连击信息
- repeatEnd       // 是否连击结束
- isComboEnd      // 连击是否结束（布尔值）

// 其他信息
- totalCoin       // 总价值
- timestamp       // 时间戳
- logId           // 日志ID
- sendType        // 发送类型（1=普通，2=投喂）
- publicArea      // 是否公屏显示
```

#### 数量计算优先级
```
groupCount > repeatCount > comboCount > 默认值1
```

### 2.3 用户信息解析增强

支持 80+ 字段的用户信息解析，包括：
- 基础信息：ID、昵称、性别、等级
- 头像信息：缩略图、中图、大图
- 认证信息：认证状态、企业认证
- 付费信息：付费等级、付费积分
- 粉丝团信息：粉丝团、边框、勋章
- 直播信息：主播信息、连麦统计
- 其他：城市、签名、经验值等

### 2.4 礼物信息解析增强

支持 27+ 字段的礼物信息解析，包括：
- 基础信息：ID、名称、类型
- 价值信息：钻石数量
- 显示信息：图标、图片、描述
- 效果信息：主效果ID、全屏效果
- 特殊标识：连麦礼物、粉丝团礼物
- 其他：持续时间、停留时间等

## 3. 技术实现细节

### 3.1 Protobuf 解析改进
- 完善了 ByteBuffer 的字段跳过逻辑
- 改进了 varint 读取的准确性
- 增强了嵌套结构的解析能力

### 3.2 错误处理
- 添加了详细的错误日志
- 记录解析失败的消息类型和原因
- 在UI中显示解析错误信息

### 3.3 UI更新机制
```go
// 旧接口
type UIUpdater interface {
    AddOrUpdateRoom(roomID string)
    AddRawMessage(roomID string, message string)
    AddParsedMessage(roomID string, message string)
}

// 新接口（新增方法）
type UIUpdater interface {
    AddOrUpdateRoom(roomID string)
    AddRawMessage(roomID string, message string)
    AddParsedMessage(roomID string, message string)
    AddParsedMessageWithDetail(roomID string, message string, detail map[string]interface{})
}
```

## 4. 使用示例

### 4.1 查看消息详情
1. 运行程序后，访问抖音直播间
2. 在界面中可以看到两个列表：
   - 左侧：📡 原始 WebSocket 消息
   - 右侧：📋 解析后的消息
3. 点击原始消息：右侧对应的解析消息会被选中
4. 点击解析消息：弹出详情窗口
5. 在详情窗口中可以：
   - 查看原始消息
   - 查看解析后的简要信息
   - 查看所有解析字段的详细信息
   - 复制详情到剪贴板

### 4.2 礼物消息示例
解析后的礼物消息包含：
```
类型: 礼物消息
用户: 张三
用户等级: 35
礼物: 小心心
礼物数量: 10
连击次数: 10
是否连击结束: true
钻石价值: 1
总价值: 10
```

## 5. 参考资源
本次改进参考了以下开源项目：
1. https://github.com/yughghbkg/DouyinLiveWebFetcher-pro
2. https://github.com/skmcj/dycast
3. https://lf-webcast-platform.bytetos.com/obj/webcast-platform-cdn/webcast/douyin_live/9569.6aac901a.js

## 6. 未来改进计划
- [ ] 添加更多消息类型的支持
- [ ] 改进连麦和PK消息的详细解析
- [ ] 添加消息统计和分析功能
- [ ] 支持消息过滤和搜索
- [ ] 导出消息记录功能

## 7. 注意事项
1. 礼物数量的计算依赖多个字段，请确保正确处理
2. 连击消息需要等待 `repeatEnd=1` 才表示连击结束
3. 部分字段可能在不同版本的协议中有所不同
4. 建议在生产环境中启用详细日志以便调试

## 8. 编译说明
由于项目使用了 Fyne GUI 框架，在 Linux 环境下编译需要安装以下依赖：
```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-dev xorg-dev

# 或者使用无GUI模式编译（仅限服务器端）
go build -tags nofyne
```

在 Windows 环境下可以直接编译运行。
