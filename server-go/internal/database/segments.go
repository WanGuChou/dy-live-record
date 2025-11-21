package database

import (
	"database/sql"
	"time"
)

// ScoreSegment 分段记分数据结构
type ScoreSegment struct {
	ID              int64     `json:"id"`
	SessionID       int64     `json:"session_id"`
	RoomID          string    `json:"room_id"`
	SegmentName     string    `json:"segment_name"`
	StartTime       time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	TotalGiftValue  int       `json:"total_gift_value"`
	TotalMessages   int       `json:"total_messages"`
}

// CreateScoreSegmentsTable 创建分段记分表
func (db *DB) CreateScoreSegmentsTable() error {
	schema := `
	CREATE TABLE IF NOT EXISTS score_segments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		room_id TEXT NOT NULL,
		segment_name TEXT NOT NULL,
		start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		end_time TIMESTAMP,
		total_gift_value INTEGER DEFAULT 0,
		total_messages INTEGER DEFAULT 0
	);
	
	CREATE INDEX IF NOT EXISTS idx_segments_session ON score_segments(session_id);
	CREATE INDEX IF NOT EXISTS idx_segments_room ON score_segments(room_id);
	`

	_, err := db.conn.Exec(schema)
	return err
}

// CreateSegment 创建新分段
func (db *DB) CreateSegment(sessionID int64, roomID, segmentName string) (int64, error) {
	result, err := db.conn.Exec(`
		INSERT INTO score_segments (session_id, room_id, segment_name)
		VALUES (?, ?, ?)
	`, sessionID, roomID, segmentName)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// EndSegment 结束当前分段并计算统计
func (db *DB) EndSegment(segmentID int64) error {
	// 1. 获取分段信息
	var sessionID int64
	var roomID string
	var startTime time.Time

	err := db.conn.QueryRow(`
		SELECT session_id, room_id, start_time
		FROM score_segments
		WHERE id = ?
	`, segmentID).Scan(&sessionID, &roomID, &startTime)

	if err != nil {
		return err
	}

	// 2. 计算该时段的礼物总值
	var totalGiftValue sql.NullInt64
	err = db.conn.QueryRow(`
		SELECT SUM(gift_diamond_value)
		FROM gift_records
		WHERE session_id = ? AND timestamp >= ?
	`, sessionID, startTime).Scan(&totalGiftValue)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 3. 计算该时段的消息总数
	var totalMessages int
	err = db.conn.QueryRow(`
		SELECT COUNT(*)
		FROM message_records
		WHERE session_id = ? AND timestamp >= ?
	`, sessionID, startTime).Scan(&totalMessages)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 4. 更新分段记录
	giftValue := int(0)
	if totalGiftValue.Valid {
		giftValue = int(totalGiftValue.Int64)
	}

	_, err = db.conn.Exec(`
		UPDATE score_segments
		SET end_time = CURRENT_TIMESTAMP,
		    total_gift_value = ?,
		    total_messages = ?
		WHERE id = ?
	`, giftValue, totalMessages, segmentID)

	return err
}

// GetActiveSegment 获取当前活动分段
func (db *DB) GetActiveSegment(sessionID int64) (*ScoreSegment, error) {
	var segment ScoreSegment
	var endTime sql.NullTime

	err := db.conn.QueryRow(`
		SELECT id, session_id, room_id, segment_name, start_time, end_time, total_gift_value, total_messages
		FROM score_segments
		WHERE session_id = ? AND end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`, sessionID).Scan(
		&segment.ID,
		&segment.SessionID,
		&segment.RoomID,
		&segment.SegmentName,
		&segment.StartTime,
		&endTime,
		&segment.TotalGiftValue,
		&segment.TotalMessages,
	)

	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		segment.EndTime = &endTime.Time
	}

	return &segment, nil
}

// GetAllSegments 获取某场次的所有分段
func (db *DB) GetAllSegments(sessionID int64) ([]ScoreSegment, error) {
	rows, err := db.conn.Query(`
		SELECT id, session_id, room_id, segment_name, start_time, end_time, total_gift_value, total_messages
		FROM score_segments
		WHERE session_id = ?
		ORDER BY start_time ASC
	`, sessionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segments := make([]ScoreSegment, 0)
	for rows.Next() {
		var segment ScoreSegment
		var endTime sql.NullTime

		err := rows.Scan(
			&segment.ID,
			&segment.SessionID,
			&segment.RoomID,
			&segment.SegmentName,
			&segment.StartTime,
			&endTime,
			&segment.TotalGiftValue,
			&segment.TotalMessages,
		)

		if err != nil {
			continue
		}

		if endTime.Valid {
			segment.EndTime = &endTime.Time
		}

		segments = append(segments, segment)
	}

	return segments, nil
}

// GetSegmentStats 获取分段详细统计（包括主播业绩）
func (db *DB) GetSegmentStats(segmentID int64) (map[string]interface{}, error) {
	// 1. 获取分段基本信息
	var segment ScoreSegment
	var endTime sql.NullTime

	err := db.conn.QueryRow(`
		SELECT id, session_id, room_id, segment_name, start_time, end_time, total_gift_value, total_messages
		FROM score_segments
		WHERE id = ?
	`, segmentID).Scan(
		&segment.ID,
		&segment.SessionID,
		&segment.RoomID,
		&segment.SegmentName,
		&segment.StartTime,
		&endTime,
		&segment.TotalGiftValue,
		&segment.TotalMessages,
	)

	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		segment.EndTime = &endTime.Time
	}

	// 2. 获取该时段各主播业绩
	endTimeFilter := "CURRENT_TIMESTAMP"
	if segment.EndTime != nil {
		endTimeFilter = "?"
	}

	query := `
		SELECT ap.anchor_id, a.anchor_name, SUM(ap.gift_value) as total_value
		FROM anchor_performance ap
		JOIN anchors a ON ap.anchor_id = a.anchor_id
		WHERE ap.recorded_at >= ? AND ap.recorded_at <= ` + endTimeFilter + `
		GROUP BY ap.anchor_id, a.anchor_name
		ORDER BY total_value DESC
	`

	var rows *sql.Rows
	if segment.EndTime != nil {
		rows, err = db.conn.Query(query, segment.StartTime, segment.EndTime)
	} else {
		rows, err = db.conn.Query(query, segment.StartTime)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	anchorStats := make([]map[string]interface{}, 0)
	for rows.Next() {
		var anchorID, anchorName string
		var totalValue int

		if err := rows.Scan(&anchorID, &anchorName, &totalValue); err != nil {
			continue
		}

		anchorStats = append(anchorStats, map[string]interface{}{
			"anchor_id":   anchorID,
			"anchor_name": anchorName,
			"total_value": totalValue,
		})
	}

	// 3. 组装结果
	result := map[string]interface{}{
		"segment":      segment,
		"anchor_stats": anchorStats,
	}

	return result, nil
}
