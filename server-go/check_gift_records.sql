-- 礼物记录调试 SQL 脚本
-- 使用方法：在 SQLite 客户端中执行这些查询

-- 1. 检查 gift_records 表结构
PRAGMA table_info(gift_records);

-- 2. 统计每个房间的礼物记录数量
SELECT 
    room_id, 
    COUNT(*) as record_count,
    MIN(create_time) as first_gift,
    MAX(create_time) as last_gift
FROM gift_records 
GROUP BY room_id 
ORDER BY record_count DESC;

-- 3. 查看所有唯一的 room_id
SELECT DISTINCT room_id FROM gift_records ORDER BY room_id;

-- 4. 查看最近的 10 条礼物记录
SELECT 
    record_id,
    msg_id,
    room_id,
    user_nickname,
    gift_name,
    gift_count,
    gift_diamond_value,
    anchor_name,
    create_time
FROM gift_records 
ORDER BY 
    COALESCE(create_time, timestamp) DESC 
LIMIT 10;

-- 5. 检查特定房间的礼物记录（替换 'YOUR_ROOM_ID' 为实际的房间ID）
SELECT 
    record_id,
    msg_id,
    user_nickname,
    gift_name,
    gift_count,
    gift_diamond_value,
    anchor_name,
    create_time
FROM gift_records 
WHERE room_id = 'YOUR_ROOM_ID'  -- ⚠️ 替换这里
ORDER BY create_time DESC 
LIMIT 20;

-- 6. 检查是否有空的 room_id
SELECT COUNT(*) as null_or_empty_room_id_count
FROM gift_records 
WHERE room_id IS NULL OR room_id = '';

-- 7. 检查 room_id 的长度分布
SELECT 
    LENGTH(room_id) as room_id_length, 
    COUNT(*) as count,
    GROUP_CONCAT(DISTINCT room_id) as sample_ids
FROM gift_records 
GROUP BY LENGTH(room_id)
ORDER BY count DESC;

-- 8. 检查 rooms 表中的房间列表
SELECT 
    room_id,
    live_room_id,
    room_title,
    anchor_name,
    first_seen_at,
    last_seen_at
FROM rooms
ORDER BY last_seen_at DESC;

-- 9. 对比 rooms 表和 gift_records 表的 room_id
SELECT 
    r.room_id as rooms_id,
    COUNT(gr.record_id) as gift_count
FROM rooms r
LEFT JOIN gift_records gr ON r.room_id = gr.room_id
GROUP BY r.room_id
ORDER BY gift_count DESC;

-- 10. 查找有礼物记录但不在 rooms 表中的房间
SELECT DISTINCT gr.room_id
FROM gift_records gr
LEFT JOIN rooms r ON gr.room_id = r.room_id
WHERE r.room_id IS NULL;

-- 11. 查看最近保存的礼物记录详情（包含所有字段）
SELECT *
FROM gift_records 
ORDER BY record_id DESC 
LIMIT 5;

-- 12. 统计礼物记录的时间分布
SELECT 
    DATE(COALESCE(create_time, timestamp)) as date,
    COUNT(*) as count
FROM gift_records
GROUP BY DATE(COALESCE(create_time, timestamp))
ORDER BY date DESC;
