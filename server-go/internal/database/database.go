package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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

// GetConn 获取底层的 sql.DB 连接（用于需要 *sql.DB 的场景）
func (db *DB) GetConn() *sql.DB {
	return db.conn
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

	-- 房间专属主播信息
	CREATE TABLE IF NOT EXISTS room_anchors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id TEXT NOT NULL,
		anchor_id TEXT NOT NULL,
		anchor_name TEXT,
		avatar_url TEXT,
		bound_gifts TEXT,
		gift_count INTEGER DEFAULT 0,
		score INTEGER DEFAULT 0,
		UNIQUE(room_id, anchor_id)
	);

	-- 礼物信息
	CREATE TABLE IF NOT EXISTS gifts (
		gift_id TEXT PRIMARY KEY,
		gift_name TEXT,
		diamond_value INTEGER DEFAULT 0,
		icon TEXT,
		version TEXT,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 房间礼物绑定
	CREATE TABLE IF NOT EXISTS room_gift_bindings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id TEXT NOT NULL,
		gift_name TEXT NOT NULL,
		anchor_id TEXT,
		UNIQUE(room_id, gift_name)
	);

	-- 索引
	CREATE INDEX IF NOT EXISTS idx_gifts_session ON gift_records(session_id);
	CREATE INDEX IF NOT EXISTS idx_gifts_room ON gift_records(room_id);
	CREATE INDEX IF NOT EXISTS idx_gifts_timestamp ON gift_records(timestamp);
	CREATE INDEX IF NOT EXISTS idx_messages_session ON message_records(session_id);
	CREATE INDEX IF NOT EXISTS idx_messages_room ON message_records(room_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_room ON live_sessions(room_id);
	CREATE INDEX IF NOT EXISTS idx_room_anchors_room ON room_anchors(room_id);
	CREATE INDEX IF NOT EXISTS idx_room_gift_binding ON room_gift_bindings(room_id);
	`

	if _, err := db.conn.Exec(schema); err != nil {
		return err
	}
	return ensureAnchorExtraColumns(db.conn)
}

// GetConnection 获取原始数据库连接（用于复杂查询）
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

func ensureAnchorExtraColumns(conn *sql.DB) error {
	if err := addColumnIfMissing(conn, "anchors", "avatar_url", "TEXT"); err != nil {
		return err
	}
	if err := addColumnIfMissing(conn, "anchors", "is_deleted", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := addColumnIfMissing(conn, "anchors", "deleted_at", "TIMESTAMP"); err != nil {
		return err
	}
	return nil
}

func addColumnIfMissing(conn *sql.DB, table, column, definition string) error {
	exists, err := columnExists(conn, table, column)
	if err != nil || exists {
		return err
	}
	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	if _, err := conn.Exec(stmt); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return nil
		}
		return err
	}
	return nil
}

func columnExists(conn *sql.DB, table, column string) (bool, error) {
	rows, err := conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if strings.EqualFold(name, column) {
			return true, nil
		}
	}
	return false, rows.Err()
}
