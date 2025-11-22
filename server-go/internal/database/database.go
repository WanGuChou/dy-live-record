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
	conn, err := sql.Open("sqlite3", buildSQLiteDSN(dbPath))
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	conn.SetMaxOpenConns(1)

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
		live_room_id TEXT,
		room_title TEXT,
		anchor_name TEXT,
		ws_url TEXT,
		first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 礼物记录表
	CREATE TABLE IF NOT EXISTS gift_records (
		record_id INTEGER PRIMARY KEY AUTOINCREMENT,
		msg_id TEXT,
		room_id TEXT NOT NULL,
		create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		user_id TEXT,
		user_nickname TEXT,
		gift_id TEXT,
		gift_name TEXT,
		gift_count INTEGER DEFAULT 1,
		gift_diamond_value INTEGER DEFAULT 0,
		anchor_id TEXT,
		anchor_name TEXT,
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
	CREATE INDEX IF NOT EXISTS idx_gifts_room ON gift_records(room_id);
	CREATE INDEX IF NOT EXISTS idx_gifts_timestamp ON gift_records(create_time);
	CREATE INDEX IF NOT EXISTS idx_room_anchors_room ON room_anchors(room_id);
	CREATE INDEX IF NOT EXISTS idx_room_gift_binding ON room_gift_bindings(room_id);
	`

	if _, err := db.conn.Exec(schema); err != nil {
		return err
	}
	if err := ensureAnchorExtraColumns(db.conn); err != nil {
		return err
	}
	if err := ensureGiftTable(db.conn); err != nil {
		return err
	}
	if err := ensureRoomsExtraColumns(db.conn); err != nil {
		return err
	}
	if err := ensureGiftRecordsColumns(db.conn); err != nil {
		return err
	}
	return dropMessageRecordsTable(db.conn)
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

func tableExists(conn *sql.DB, table string) (bool, error) {
	var count int
	err := conn.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func ensureGiftTable(conn *sql.DB) error {
	exists, err := tableExists(conn, "gifts")
	if err != nil {
		return err
	}
	if !exists {
		_, err := conn.Exec(`
			CREATE TABLE IF NOT EXISTS gifts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				gift_id TEXT UNIQUE,
				gift_name TEXT,
				diamond_value INTEGER DEFAULT 0,
				icon_url TEXT,
				icon_local TEXT,
				version TEXT,
				is_deleted INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
		`)
		return err
	}

	hasID, err := columnExists(conn, "gifts", "id")
	if err != nil {
		return err
	}
	if !hasID {
		_, err := conn.Exec(`
			ALTER TABLE gifts RENAME TO gifts_backup;
			CREATE TABLE IF NOT EXISTS gifts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				gift_id TEXT UNIQUE,
				gift_name TEXT,
				diamond_value INTEGER DEFAULT 0,
				icon_url TEXT,
				icon_local TEXT,
				version TEXT,
				is_deleted INTEGER DEFAULT 0,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);
			INSERT INTO gifts (gift_id, gift_name, diamond_value, icon_local, version, updated_at)
			SELECT gift_id, gift_name, diamond_value, icon, version, updated_at FROM gifts_backup;
			DROP TABLE gifts_backup;
		`)
		if err != nil {
			return err
		}
	}

	if err := addColumnIfMissing(conn, "gifts", "icon_url", "TEXT"); err != nil {
		return err
	}
	if err := addColumnIfMissing(conn, "gifts", "icon_local", "TEXT"); err != nil {
		return err
	}
	if err := addColumnIfMissing(conn, "gifts", "is_deleted", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := addColumnIfMissing(conn, "gifts", "created_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"); err != nil {
		return err
	}
	if err := addColumnIfMissing(conn, "gifts", "updated_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"); err != nil {
		return err
	}
	return nil
}

func ensureRoomsExtraColumns(conn *sql.DB) error {
	if err := addColumnIfMissing(conn, "rooms", "live_room_id", "TEXT"); err != nil {
		return err
	}
	return addColumnIfMissing(conn, "rooms", "ws_url", "TEXT")
}

func buildSQLiteDSN(path string) string {
	if strings.TrimSpace(path) == "" {
		return path
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%s_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on", path, separator)
}

func ensureGiftRecordsColumns(conn *sql.DB) error {
	// 添加 msg_id 列
	if err := addColumnIfMissing(conn, "gift_records", "msg_id", "TEXT"); err != nil {
		return err
	}

	// 添加 anchor_name 列
	if err := addColumnIfMissing(conn, "gift_records", "anchor_name", "TEXT"); err != nil {
		return err
	}

	// 检查是否存在 timestamp 列，如果存在但没有 create_time 列，则需要迁移数据
	hasTimestamp, err := columnExists(conn, "gift_records", "timestamp")
	if err != nil {
		return err
	}
	hasCreateTime, err := columnExists(conn, "gift_records", "create_time")
	if err != nil {
		return err
	}

	// 如果 timestamp 存在但 create_time 不存在，迁移数据
	if hasTimestamp && !hasCreateTime {
		// 添加 create_time 列
		if err := addColumnIfMissing(conn, "gift_records", "create_time", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"); err != nil {
			return err
		}
		// 将 timestamp 的数据复制到 create_time
		if _, err := conn.Exec(`UPDATE gift_records SET create_time = timestamp WHERE create_time IS NULL`); err != nil {
			return err
		}
	}

	return nil
}

func dropMessageRecordsTable(conn *sql.DB) error {
	exists, err := tableExists(conn, "message_records")
	if err != nil {
		return err
	}
	if exists {
		_, err := conn.Exec(`DROP TABLE IF EXISTS message_records`)
		return err
	}
	return nil
}
