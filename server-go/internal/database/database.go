package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// DB 数据库包装
type DB struct {
	conn *sql.DB
}

// Init 初始化数据库
func Init(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	db := &DB{conn: conn}

	// 初始化表结构
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("初始化数据库结构失败: %w", err)
	}

	log.Println("✅ 数据库表结构初始化完成")
	return db, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// initSchema 初始化数据库表结构
func (db *DB) initSchema() error {
	schema := `
	-- 房间信息表
	CREATE TABLE IF NOT EXISTS rooms (
		room_id TEXT PRIMARY KEY,
		room_title TEXT,
		anchor_name TEXT,
		first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 直播场次表
	CREATE TABLE IF NOT EXISTS live_sessions (
		session_id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id TEXT NOT NULL,
		start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_time TIMESTAMP,
		total_gifts_value INTEGER DEFAULT 0,
		total_messages INTEGER DEFAULT 0,
		total_members INTEGER DEFAULT 0,
		FOREIGN KEY (room_id) REFERENCES rooms(room_id)
	);

	-- 礼物记录表
	CREATE TABLE IF NOT EXISTS gift_records (
		record_id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		room_id TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		user_id TEXT,
		user_nickname TEXT,
		gift_id TEXT,
		gift_name TEXT,
		gift_count INTEGER DEFAULT 1,
		gift_diamond_value INTEGER DEFAULT 0,
		anchor_id TEXT,
		FOREIGN KEY (session_id) REFERENCES live_sessions(session_id),
		FOREIGN KEY (room_id) REFERENCES rooms(room_id)
	);

	-- 消息记录表
	CREATE TABLE IF NOT EXISTS message_records (
		record_id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		room_id TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		message_type TEXT NOT NULL,
		user_id TEXT,
		user_nickname TEXT,
		content TEXT,
		FOREIGN KEY (session_id) REFERENCES live_sessions(session_id),
		FOREIGN KEY (room_id) REFERENCES rooms(room_id)
	);

	-- 主播配置表
	CREATE TABLE IF NOT EXISTS anchors (
		anchor_id TEXT PRIMARY KEY,
		anchor_name TEXT NOT NULL,
		bound_gifts TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 索引
	CREATE INDEX IF NOT EXISTS idx_gifts_session ON gift_records(session_id);
	CREATE INDEX IF NOT EXISTS idx_gifts_room ON gift_records(room_id);
	CREATE INDEX IF NOT EXISTS idx_gifts_timestamp ON gift_records(timestamp);
	CREATE INDEX IF NOT EXISTS idx_messages_session ON message_records(session_id);
	CREATE INDEX IF NOT EXISTS idx_messages_room ON message_records(room_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_room ON live_sessions(room_id);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// GetConnection 获取原始数据库连接（用于复杂查询）
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}
