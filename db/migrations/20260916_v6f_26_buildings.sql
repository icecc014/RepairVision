SET NAMES utf8mb4;

-- ============ 1) 新增 10 栋教学楼（总 26 栋），标准层沿用 1 号楼 ============
USE map_db;
SET @tpl_layout := (SELECT layout_json FROM buildings WHERE code = '1' LIMIT 1);
SET @tpl_w := (SELECT width FROM buildings WHERE id = 1);
SET @tpl_h := (SELECT height FROM buildings WHERE id = 1);
SET @tpl_f := (SELECT floors FROM buildings WHERE id = 1);
SET @tpl_fh := (SELECT floor_height FROM buildings WHERE id = 1);
SET @tpl_rp := (SELECT rooms_per_floor FROM buildings WHERE id = 1);

INSERT INTO buildings(code, name, pos_x, pos_y, width, height, floors, floor_height, rooms_per_floor, layout_json)
SELECT CAST(t.n AS CHAR), CONCAT(t.n, '号教学楼'), 0, 0, @tpl_w, @tpl_h, @tpl_f, @tpl_fh, @tpl_rp, @tpl_layout
FROM (SELECT 17 n UNION SELECT 18 UNION SELECT 19 UNION SELECT 20 UNION SELECT 21
      UNION SELECT 22 UNION SELECT 23 UNION SELECT 24 UNION SELECT 25 UNION SELECT 26) t
WHERE NOT EXISTS (SELECT 1 FROM (SELECT code FROM buildings) x WHERE x.code = CAST(t.n AS CHAR));

-- ============ 2) 补齐 26 位宿管（dorm17~dorm26，一栋一位） ============
USE worker_db;
SET @pwd := (SELECT password FROM users WHERE username = 'dorm1' LIMIT 1);
INSERT INTO users(username, password, role, job_type, name, phone, building_id, status, max_concurrent)
SELECT CONCAT('dorm', t.n), @pwd, 3, 0, CONCAT(t.n, '号教学楼管理员'), NULL, t.n, 1, 8
FROM (SELECT 17 n UNION SELECT 18 UNION SELECT 19 UNION SELECT 20 UNION SELECT 21
      UNION SELECT 22 UNION SELECT 23 UNION SELECT 24 UNION SELECT 25 UNION SELECT 26) t
WHERE NOT EXISTS (SELECT 1 FROM (SELECT username FROM users) u WHERE u.username = CONCAT('dorm', t.n));

-- ============ 3) 清空旧工单，按“并非每栋每天都报修”的真实规律重造 ============
USE order_db;
DELETE FROM dispatch_records;
DELETE FROM order_feedbacks;
DELETE FROM order_duplicate_records;
DELETE FROM notifications;
DELETE FROM orders;

INSERT INTO orders(order_no, title, description, building_id, room, floor, fault_type,
                   priority, expect_minutes, status, is_merged, manual_review, external_mark,
                   dispatch_locked, main_order_id, worker_id,
                   dispatched_at, started_at, completed_at, reporter_id, source, created_at, updated_at)
SELECT
  CONCAT('REP', DATE_FORMAT(r.ts, '%Y%m%d'), LPAD(r.building_id, 2, '0'), LPAD(r.seq_in_day, 2, '0')),
  CONCAT(ft.name, '（', r.room, '室）'),
  CONCAT(r.room, '室 ', ft.name, '，请尽快处理'),
  r.building_id, r.room, r.floor, r.fault_code,
  1 + (r.m % 3), 30, r.st, 0,
  CASE WHEN r.fault_code = 'other' THEN 1 ELSE 0 END, 0, 0, NULL,
  CASE WHEN r.fault_code = 'other' THEN NULL ELSE (
     SELECT u.id FROM worker_db.worker_buildings wb
     JOIN worker_db.users u ON u.id = wb.worker_id
     WHERE wb.building_id = r.building_id AND u.status = 1
       AND u.job_type = CASE r.fault_code WHEN 'electric' THEN 1 WHEN 'water' THEN 2 WHEN 'masonry' THEN 3 WHEN 'wood' THEN 4 ELSE 0 END
     ORDER BY u.id LIMIT 1
  ) END,
  CASE WHEN r.st IN (2,3,4) THEN DATE_ADD(r.ts, INTERVAL 20 MINUTE) END,
  CASE WHEN r.st IN (3,4) THEN DATE_ADD(r.ts, INTERVAL 50 MINUTE) END,
  CASE WHEN r.st = 4 THEN DATE_ADD(r.ts, INTERVAL 110 MINUTE) END,
  COALESCE((SELECT d.id FROM worker_db.users d WHERE d.role = 3 AND d.building_id = r.building_id LIMIT 1),
           (SELECT MIN(d.id) FROM worker_db.users d WHERE d.role = 3)),
  'dormitory', r.ts, r.ts
FROM (
  SELECT g.building_id, g.day_offset, g.seq AS seq_in_day,
         DATE_ADD(DATE_ADD(DATE_SUB(CURDATE(), INTERVAL g.day_offset DAY), INTERVAL 8 HOUR),
                  INTERVAL ((g.building_id * 3 + g.seq * 7 + g.day_offset * 5) % 9) HOUR) AS ts,
         1 + ((g.building_id + g.seq + g.day_offset) % 6) AS floor,
         CONCAT(1 + ((g.building_id + g.seq + g.day_offset) % 6),
                LPAD(1 + ((g.building_id * 5 + g.seq * 3 + g.day_offset) % 16), 2, '0')) AS room,
         ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) AS m,
         CASE
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 45 THEN 'electric'
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 80 THEN 'water'
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 90 THEN 'masonry'
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 97 THEN 'wood'
           ELSE 'other'
         END AS fault_code,
         CASE
           WHEN g.day_offset >= 2 THEN 4
           WHEN g.day_offset = 1 THEN (CASE WHEN ((g.building_id + g.seq) % 10) < 2 THEN 3 ELSE 4 END)
           ELSE (CASE WHEN ((g.building_id + g.seq) % 10) < 3 THEN 1
                      WHEN ((g.building_id + g.seq) % 10) < 6 THEN 2
                      ELSE 3 END)
         END AS st
  FROM (
    -- 每栋楼每天按真实概率决定"当天是否有报修"：30% 无报修、40% 1 单、22% 2 单、8% 3 单
    SELECT bd.building_id, bd.day_offset, s.s AS seq
    FROM (
      SELECT b.id AS building_id, d.d AS day_offset,
             CASE WHEN ((b.id * 7 + d.d * 13) % 100) < 30 THEN 0
                  WHEN ((b.id * 7 + d.d * 13) % 100) < 70 THEN 1
                  WHEN ((b.id * 7 + d.d * 13) % 100) < 92 THEN 2
                  ELSE 3 END AS cnt
      FROM map_db.buildings b
      CROSS JOIN (SELECT 0 d UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6) d
    ) bd
    JOIN (SELECT 1 s UNION SELECT 2 UNION SELECT 3) s ON s.s <= bd.cnt
  ) g
) r
JOIN order_db.fault_types ft ON ft.code = r.fault_code;

SELECT 'buildings' AS t, COUNT(*) AS n FROM map_db.buildings
UNION ALL SELECT 'dorms', COUNT(*) FROM worker_db.users WHERE role = 3
UNION ALL SELECT 'workers', COUNT(*) FROM worker_db.users WHERE role = 2
UNION ALL SELECT 'orders', COUNT(*) FROM order_db.orders
UNION ALL SELECT 'building_days_with_orders', COUNT(*) FROM (SELECT building_id, DATE(created_at) d FROM order_db.orders GROUP BY building_id, DATE(created_at)) x
UNION ALL SELECT 'max_per_building_day', MAX(c) FROM (SELECT building_id, DATE(created_at) d, COUNT(*) c FROM order_db.orders GROUP BY building_id, DATE(created_at)) y;