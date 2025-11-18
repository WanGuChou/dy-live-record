package database

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var roomTableSanitizer = regexp.MustCompile(`[^a-zA-Z0-9]`)

// RoomMessageRecord 描述单条房间消息
type RoomMessageRecord struct {
	RoomID      string
	Method      string
	MessageType string
	Display     string
	UserID      string
	UserName    string
	GiftName    string
	GiftCount   int
	GiftValue   int
	AnchorID    string
	RawPayload  []byte
	ParsedJSON  string
	Source      string
	SentAt      time.Time
}

// EnsureRoomTables 确保房间相关的动态表已创建
func (db *DB) EnsureRoomTables(roomID string) error {
	tableName := db.roomMessageTable(roomID)
	schema := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		method TEXT,
		message_type TEXT,
		display TEXT,
		user_id TEXT,
		user_name TEXT,
		gift_name TEXT,
		gift_count INTEGER DEFAULT 0,
		gift_value INTEGER DEFAULT 0,
		anchor_id TEXT,
		raw_payload BLOB,
		parsed_json TEXT,
		source TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_%s_method ON %s(method);
	CREATE INDEX IF NOT EXISTS idx_%s_type ON %s(message_type);
	`, tableName, tableName, tableName, tableName, tableName)

	_, err := db.conn.Exec(schema)
	return err
}

// InsertRoomMessage 写入房间消息记录
func (db *DB) InsertRoomMessage(record *RoomMessageRecord) error {
	if record == nil {
		return fmt.Errorf("record 不能为空")
	}

	if err := db.EnsureRoomTables(record.RoomID); err != nil {
		return err
	}

	table := db.roomMessageTable(record.RoomID)

	_, err := db.conn.Exec(fmt.Sprintf(`
		INSERT INTO %s (
			timestamp, method, message_type, display,
			user_id, user_name, gift_name, gift_count, gift_value,
			anchor_id, raw_payload, parsed_json, source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, table),
		record.SentAt,
		record.Method,
		record.MessageType,
		record.Display,
		record.UserID,
		record.UserName,
		record.GiftName,
		record.GiftCount,
		record.GiftValue,
		record.AnchorID,
		record.RawPayload,
		record.ParsedJSON,
		record.Source,
	)

	return err
}

func (db *DB) roomMessageTable(roomID string) string {
	return RoomMessageTableName(roomID)
}

// RoomMessageTableName 生成房间消息表名称
func RoomMessageTableName(roomID string) string {
	safe := roomTableSanitizer.ReplaceAllString(roomID, "_")
	if safe == "" {
		safe = "unknown"
	}
	return fmt.Sprintf("room_%s_messages", strings.ToLower(safe))
}
