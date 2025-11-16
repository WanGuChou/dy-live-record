package server

import (
	"database/sql"
	"log"
	"strings"
)

// GiftAllocator 礼物分配器
type GiftAllocator struct {
	db *sql.DB
}

// NewGiftAllocator 创建礼物分配器
func NewGiftAllocator(db *sql.DB) *GiftAllocator {
	return &GiftAllocator{db: db}
}

// AllocateGift 分配礼物给主播
// 返回分配的主播ID，如果未分配则返回空字符串
func (ga *GiftAllocator) AllocateGift(giftName string, messageContent string) (string, error) {
	// 1. 首先检查礼物是否已绑定到某个主播
	anchorID, err := ga.getAnchorByBoundGift(giftName)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if anchorID != "" {
		log.Printf("🎁 礼物 [%s] 已绑定到主播 [%s]", giftName, anchorID)
		return anchorID, nil
	}

	// 2. 解析消息内容，查找 @主播名 或 "送给XX" 等指令
	anchorID, err = ga.parseMessageForAnchor(messageContent)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if anchorID != "" {
		log.Printf("🎁 从消息 [%s] 中识别到主播 [%s]", messageContent, anchorID)
		return anchorID, nil
	}

	return "", nil
}

// getAnchorByBoundGift 根据绑定的礼物获取主播ID
func (ga *GiftAllocator) getAnchorByBoundGift(giftName string) (string, error) {
	var anchorID, boundGifts string
	rows, err := ga.db.Query(`SELECT anchor_id, bound_gifts FROM anchors`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&anchorID, &boundGifts); err != nil {
			continue
		}

		// 检查礼物是否在绑定列表中
		gifts := strings.Split(boundGifts, ",")
		for _, gift := range gifts {
			if strings.TrimSpace(gift) == giftName {
				return anchorID, nil
			}
		}
	}

	return "", sql.ErrNoRows
}

// parseMessageForAnchor 从消息中解析主播名称
func (ga *GiftAllocator) parseMessageForAnchor(message string) (string, error) {
	if message == "" {
		return "", sql.ErrNoRows
	}

	// 获取所有主播
	rows, err := ga.db.Query(`SELECT anchor_id, anchor_name FROM anchors`)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var anchorID, anchorName string
	for rows.Next() {
		if err := rows.Scan(&anchorID, &anchorName); err != nil {
			continue
		}

		// 检查消息中是否包含 @主播名、送给主播名、给主播名 等关键词
		if strings.Contains(message, "@"+anchorName) ||
			strings.Contains(message, "送给"+anchorName) ||
			strings.Contains(message, "给"+anchorName) ||
			strings.Contains(message, anchorName) {
			return anchorID, nil
		}
	}

	return "", sql.ErrNoRows
}

// RecordAnchorPerformance 记录主播业绩
func (ga *GiftAllocator) RecordAnchorPerformance(anchorID string, giftName string, giftValue int) error {
	// 创建或更新主播业绩记录表
	_, err := ga.db.Exec(`
		CREATE TABLE IF NOT EXISTS anchor_performance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			anchor_id TEXT NOT NULL,
			gift_name TEXT NOT NULL,
			gift_value INTEGER NOT NULL,
			recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 插入业绩记录
	_, err = ga.db.Exec(`
		INSERT INTO anchor_performance (anchor_id, gift_name, gift_value)
		VALUES (?, ?, ?)
	`, anchorID, giftName, giftValue)

	if err == nil {
		log.Printf("📊 主播 [%s] 业绩记录: %s (价值: %d 💎)", anchorID, giftName, giftValue)
	}

	return err
}
