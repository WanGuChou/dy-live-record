package parser

import (
	"fmt"
	"strings"
)

// DouyinParser 抖音消息解析器
type DouyinParser struct {
	statistics Statistics
}

// Statistics 统计信息
type Statistics struct {
	TotalMessages int
	ChatCount     int
	GiftCount     int
	LikeCount     int
	MemberCount   int
	OnlineUsers   int
}

// NewDouyinParser 创建解析器
func NewDouyinParser() *DouyinParser {
	return &DouyinParser{
		statistics: Statistics{},
	}
}

// ParseMessage 解析抖音消息（使用完整的 Protobuf 逻辑）
func (p *DouyinParser) ParseMessage(payloadData, url string) ([]map[string]interface{}, error) {
	// 调用 protobuf.go 中的解析函数
	messages, err := ParseDouyinMessage(payloadData, url)
	if err != nil {
		return nil, err
	}

	// 更新统计
	for _, msg := range messages {
		p.statistics.TotalMessages++
		
		if msgType, ok := msg["messageType"].(string); ok {
			switch msgType {
			case "聊天消息":
				p.statistics.ChatCount++
			case "礼物消息":
				p.statistics.GiftCount++
			case "点赞消息":
				p.statistics.LikeCount++
			case "进入直播间":
				p.statistics.MemberCount++
			}
		}
	}

	return messages, nil
}

// FormatMessage 格式化消息（用于控制台输出）
func (p *DouyinParser) FormatMessage(messages []map[string]interface{}) string {
	if len(messages) == 0 {
		return ""
	}

	var result []string
	for _, msg := range messages {
		formatted := p.formatSingleMessage(msg)
		if formatted != "" {
			result = append(result, formatted)
		}
	}

	return strings.Join(result, "\n\n")
}

// formatSingleMessage 格式化单条消息
func (p *DouyinParser) formatSingleMessage(msg map[string]interface{}) string {
	lines := []string{
		"╔" + strings.Repeat("═", 78) + "╗",
		"║ 🎬 抖音直播消息",
		"╠" + strings.Repeat("═", 78) + "╣",
	}

	// 消息类型
	if msgType, ok := msg["messageType"].(string); ok {
		lines = append(lines, fmt.Sprintf("║ 消息类型: %s", msgType))
	}

	// 时间戳
	if timestamp, ok := msg["timestamp"].(string); ok {
		lines = append(lines, fmt.Sprintf("║ 时间: %s", timestamp))
	}

	// 根据消息类型添加详细信息
	msgType, _ := msg["messageType"].(string)
	
	switch msgType {
	case "聊天消息":
		if user, ok := msg["user"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 用户: %s", user))
		}
		if level, ok := msg["level"].(int32); ok {
			lines = append(lines, fmt.Sprintf("║ 等级: %d", level))
		}
		if content, ok := msg["content"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 内容: %s", content))
		}

	case "礼物消息":
		if user, ok := msg["user"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 用户: %s", user))
		}
		if giftName, ok := msg["giftName"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 礼物: %s", giftName))
		}
		if giftCount, ok := msg["giftCount"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 数量: %s", giftCount))
		}
		if diamondCount, ok := msg["diamondCount"].(int32); ok {
			lines = append(lines, fmt.Sprintf("║ 价值: %d 💎", diamondCount))
		}

	case "点赞消息":
		if user, ok := msg["user"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 用户: %s ❤️", user))
		}
		if count, ok := msg["count"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 点赞数: %s", count))
		}

	case "进入直播间":
		if user, ok := msg["user"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 用户: %s", user))
		}
		if memberCount, ok := msg["memberCount"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 当前人数: %s", memberCount))
		}

	case "在线人数":
		if total, ok := msg["total"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 在线人数: %s 👥", total))
		}
		if totalUser, ok := msg["totalUser"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 累计观看: %s", totalUser))
		}

	case "直播间统计":
		if displayMiddle, ok := msg["displayMiddle"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 在线观众: %s 👥", displayMiddle))
		}

	case "关注消息":
		if user, ok := msg["user"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 用户: %s", user))
		}
		lines = append(lines, "║ 动作: 关注了主播")

	default:
		if method, ok := msg["method"].(string); ok {
			lines = append(lines, fmt.Sprintf("║ 方法: %s", method))
		}
	}

	lines = append(lines, "╚"+strings.Repeat("═", 78)+"╝")
	return strings.Join(lines, "\n")
}

// GetStatistics 获取统计信息
func (p *DouyinParser) GetStatistics() Statistics {
	return p.statistics
}

// ResetStatistics 重置统计
func (p *DouyinParser) ResetStatistics() {
	p.statistics = Statistics{}
}

// FormatStatistics 格式化统计信息
func (p *DouyinParser) FormatStatistics() string {
	stats := p.GetStatistics()
	lines := []string{
		"╔" + strings.Repeat("═", 78) + "╗",
		"║ 📊 抖音直播统计",
		"╠" + strings.Repeat("═", 78) + "╣",
		fmt.Sprintf("║ 总消息数: %d", stats.TotalMessages),
		fmt.Sprintf("║ 聊天消息: %d", stats.ChatCount),
		fmt.Sprintf("║ 礼物消息: %d", stats.GiftCount),
		fmt.Sprintf("║ 点赞消息: %d", stats.LikeCount),
		fmt.Sprintf("║ 进入直播间: %d", stats.MemberCount),
		fmt.Sprintf("║ 当前在线: %d 👥", stats.OnlineUsers),
		"╚" + strings.Repeat("═", 78) + "╝",
	}
	return strings.Join(lines, "\n")
}
