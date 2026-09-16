SET NAMES utf8mb4;

-- 1) 楼栋：id 与编码一一对应
USE map_db;
UPDATE buildings SET code = CONCAT('T', id);  -- 先临时值避免唯一键冲突
UPDATE buildings SET code = CAST(id AS CHAR), name = CONCAT(id, '号宿舍楼');

-- 2) 删除旧演示工人（李工/王工/赵工/孙工）及其关联数据
USE worker_db;
DELETE FROM worker_skills WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM worker_buildings WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM worker_schedules WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM worker_leave_requests WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM users WHERE username IN ('worker1','worker2','worker3','worker4');

-- 3) 清空历史工单与派单记录
USE order_db;
DELETE FROM dispatch_records;
DELETE FROM order_feedbacks;
DELETE FROM order_duplicate_records;
DELETE FROM notifications;
DELETE FROM orders;

-- 4) 生成贴近真实的测试工单：16 栋 × 近 7 天 × 每栋每天 2~3 单（故障类型按真实比例分布）
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
  SELECT g.building_id, g.day_offset, g.seq,
         g.seq AS seq_in_day,
         DATE_ADD(DATE_ADD(DATE_SUB(CURDATE(), INTERVAL g.day_offset DAY), INTERVAL 8 HOUR),
                  INTERVAL ((g.building_id * 3 + g.seq * 7 + g.day_offset * 5) % 9) HOUR) AS ts,
         1 + ((g.building_id + g.seq + g.day_offset) % 6) AS floor,
         CONCAT(1 + ((g.building_id + g.seq + g.day_offset) % 6),
                LPAD(1 + ((g.building_id * 5 + g.seq * 3 + g.day_offset) % 16), 2, '0')) AS room,
         ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) AS m,
         CASE
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 40 THEN 'electric'
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 75 THEN 'water'
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 85 THEN 'masonry'
           WHEN ((g.building_id * 13 + g.seq * 7 + g.day_offset * 11) % 100) < 95 THEN 'wood'
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
    SELECT b.id AS building_id, d.d AS day_offset, s.s AS seq
    FROM map_db.buildings b
    CROSS JOIN (SELECT 0 d UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6) d
    CROSS JOIN (SELECT 1 s UNION SELECT 2 UNION SELECT 3) s
    WHERE s.s <= 2 + ((b.id + d.d) % 2)
  ) g
) r
JOIN order_db.fault_types ft ON ft.code = r.fault_code;

SELECT 'buildings' AS t, COUNT(*) AS n FROM map_db.buildings
UNION ALL SELECT 'workers', COUNT(*) FROM worker_db.users WHERE role = 2
UNION ALL SELECT 'orders', COUNT(*) FROM order_db.orders
UNION ALL SELECT 'per_building_day_max', MAX(c) FROM (
  SELECT building_id, DATE(created_at) d, COUNT(*) c FROM order_db.orders GROUP BY building_id, DATE(created_at)) x;
SELECT code, name FROM map_db.buildings ORDER BY CAST(code AS UNSIGNED);