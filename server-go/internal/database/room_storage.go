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
	MsgID       string
	RoomID      string
	Method      string
	MessageType string
	Display     string
	UserID      string
	UserName    string
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
		msg_id TEXT,
		room_id TEXT,
		create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		method TEXT,
		message_type TEXT,
		display TEXT,
		user_id TEXT,
		user_name TEXT,
		anchor_id TEXT,
		raw_payload BLOB,
		parsed_json TEXT,
		source TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_%s_method ON %s(method);
	CREATE INDEX IF NOT EXISTS idx_%s_type ON %s(message_type);
	CREATE INDEX IF NOT EXISTS idx_%s_room ON %s(room_id);
	CREATE INDEX IF NOT EXISTS idx_%s_time ON %s(create_time);
	`, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName, tableName)

	// 如果表已存在，添加缺失的列
	if err := db.ensureRoomMessageColumns(roomID); err != nil {
		return err
	}

	_, err := db.conn.Exec(schema)
	return err
}

// ensureRoomMessageColumns 确保房间消息表包含所有需要的列
func (db *DB) ensureRoomMessageColumns(roomID string) error {
	tableName := db.roomMessageTable(roomID)

	// 检查表是否存在
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, tableName).Scan(&count)
	if err != nil || count == 0 {
		return nil // 表不存在，无需迁移
	}

	// 添加缺失的列
	if err := addColumnIfMissing(db.conn, tableName, "msg_id", "TEXT"); err != nil {
		return err
	}
	if err := addColumnIfMissing(db.conn, tableName, "room_id", "TEXT"); err != nil {
		return err
	}

	// 检查是否需要迁移 timestamp 到 create_time
	hasTimestamp, err := columnExists(db.conn, tableName, "timestamp")
	if err != nil {
		return err
	}
	hasCreateTime, err := columnExists(db.conn, tableName, "create_time")
	if err != nil {
		return err
	}

	if hasTimestamp && !hasCreateTime {
		if err := addColumnIfMissing(db.conn, tableName, "create_time", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"); err != nil {
			return err
		}
		// 迁移数据
		query := fmt.Sprintf(`UPDATE %s SET create_time = timestamp WHERE create_time IS NULL`, tableName)
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
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
			msg_id, room_id, create_time, method, message_type, display,
			user_id, user_name, anchor_id, raw_payload, parsed_json, source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, table),
		record.MsgID,
		record.RoomID,
		record.SentAt,
		record.Method,
		record.MessageType,
		record.Display,
		record.UserID,
		record.UserName,
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
