# 新功能快速使用指南

## 🎯 新功能概述

本次更新添加了以下主要功能：
1. ✨ 原始消息与解析消息的关联查看
2. 🔍 点击查看完整的消息解析详情
3. 📊 支持更多消息类型（新增12种）
4. 💎 更详细的礼物信息解析

## 🚀 快速开始

### 1. 编译运行

#### Windows 环境
```bash
cd server-go
go build
.\dy-live-monitor.exe
```

#### Linux 环境（需要安装GUI依赖）
```bash
# 安装依赖
sudo apt-get install libgl1-mesa-dev xorg-dev

# 编译运行
cd server-go
go build
./dy-live-monitor
```

### 2. 使用新功能

#### 📡 查看原始消息
1. 程序启动后，安装并启动浏览器插件
2. 访问抖音直播间
3. 在"🏠 房间 XXX" Tab 中，左侧显示"📡 原始 WebSocket 消息"

#### 📋 查看解析消息
1. 右侧显示"📋 解析后的消息"
2. 解析消息包含消息类型、用户、内容等关键信息

#### 🔗 消息关联查看
1. **点击左侧原始消息**：
   - 右侧对应的解析消息会自动被选中
   - 并滚动到该消息位置
   - 便于对比原始数据和解析结果

2. **点击右侧解析消息**：
   - 弹出详情对话框
   - 显示完整的消息信息

#### 🔍 查看详情对话框
详情对话框包含以下内容：

```
📅 时间: 2025-11-17 14:30:45

📡 原始消息:
URL: wss://...
Payload: ABCD1234...

📋 解析后消息:
类型: 礼物消息 | 用户: 张三 | 礼物: 小心心 x10

🔍 详细信息:
  messageType: 礼物消息
  user: 张三
  userId: 123456789
  userLevel: 35
  giftName: 小心心
  giftId: 5678
  giftCount: 10
  diamondCount: 1
  totalCoin: 10
  comboCount: 10
  isComboEnd: true
  ...
```

**详情对话框功能：**
- 📋 **复制详情**：点击"复制详情"按钮，将所有信息复制到剪贴板
- ❌ **关闭窗口**：点击"关闭"按钮或窗口关闭图标

## 💡 使用技巧

### 1. 调试消息解析
```
1. 点击原始消息，查看原始数据
2. 查看右侧解析结果
3. 如果解析失败，会显示错误信息
4. 点击详情查看所有解析的字段
```

### 2. 监控礼物信息
```
礼物消息现在包含更多信息：
- 基础信息：用户、礼物名称、数量
- 连击信息：连击次数、是否连击结束
- 价值信息：单价、总价
- 其他：接收者（PK场景）、发送类型等
```

### 3. 查看特殊消息
```
新增支持的消息类型：
- 直播间消息（开播/下播）
- PK消息（比分、状态）
- 连麦消息
- 榜单更新
- 商品变化
- 等等...
```

## 📊 消息类型说明

### 常见消息类型

| 消息类型 | 说明 | 主要字段 |
|---------|------|---------|
| 聊天消息 | 用户发送的文字聊天 | user, content, level |
| 礼物消息 | 用户赠送礼物 | user, giftName, giftCount, diamondCount |
| 点赞消息 | 用户点赞 | user, count, total |
| 进入直播间 | 用户进入直播间 | user, memberCount |
| 关注消息 | 用户关注主播 | user, followCount |
| 在线人数 | 直播间在线人数更新 | total, totalUser |
| 直播间统计 | 直播间数据统计 | displayMiddle |

### 新增消息类型

| 消息类型 | 说明 | 主要字段 |
|---------|------|---------|
| 直播间消息 | 开播/下播通知 | content, roomStatus, statusText |
| PK消息 | PK比分和状态 | matchScore, ownScore, againstScore |
| 连麦消息 | 连麦相关信息 | scene, micStatus |
| 榜单更新 | 榜单排名变化 | rankType |
| 房间横幅 | 房间横幅消息 | content |
| 商品变化 | 商品上下架 | updateType, productId |
| 弹幕消息 | 弹幕内容 | content |
| 表情消息 | 表情聊天 | user, content, emojiId |

## 🎁 礼物消息详解

### 礼物数量说明
礼物数量的计算使用以下优先级：
1. **groupCount** - 礼物数量（最常用）
2. **repeatCount** - 重复次数
3. **comboCount** - 连击次数
4. **默认值** - 1

### 连击礼物
连击礼物会收到多条消息：
```
第1条: comboCount=1, isComboEnd=false
第2条: comboCount=2, isComboEnd=false
第3条: comboCount=3, isComboEnd=false
...
最后: comboCount=10, isComboEnd=true
```

只有当 `isComboEnd=true` 时，连击才算结束。

### 礼物PK
当在PK场景下，礼物消息会包含接收者信息：
```
user: 张三       // 发送者
toUser: 李四     // 接收者（主播）
giftName: 小心心
giftCount: 10
```

## ⚙️ 配置说明

### 调试模式
在 `config.json` 中启用调试模式：
```json
{
  "debug": {
    "enabled": true,
    "verboseLog": true
  }
}
```

启用后会输出更详细的解析日志。

### 消息限制
- 每个房间最多保留 **100 条**原始消息
- 每个房间最多保留 **100 条**解析消息
- 超过限制时，旧消息会被自动清除

## 🐛 常见问题

### Q: 为什么有些消息显示"❌ 解析失败"？
A: 可能的原因：
1. 消息类型暂不支持
2. 协议版本更新导致字段变化
3. 消息格式异常

解决方法：
- 查看详情对话框中的错误信息
- 检查原始消息数据
- 反馈给开发者

### Q: 点击消息没有反应？
A: 请确保：
1. 消息列表已经加载
2. 点击的是消息文本区域
3. 查看控制台是否有错误日志

### Q: 礼物数量显示不正确？
A: 请注意：
1. 连击礼物需要等待 `isComboEnd=true`
2. 不同礼物可能使用不同的数量字段
3. 可以在详情对话框中查看所有字段

### Q: Linux 环境编译失败？
A: 需要安装 GUI 依赖：
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev
```

## 📝 反馈与建议

如果您遇到问题或有改进建议，请：
1. 查看详细日志
2. 截图或记录错误信息
3. 提供复现步骤
4. 反馈给开发团队

## 📚 更多文档

- [详细改进文档](./PARSER_IMPROVEMENTS.md)
- [修改总结](./CHANGES_SUMMARY.md)
- [项目README](./README.md)

---

**祝使用愉快！** 🎉
