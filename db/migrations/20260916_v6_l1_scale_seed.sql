-- V6 L1：标准层模板（复制 1 号宿舍楼）+ 16 栋宿舍楼 + 10 名工人（含泥瓦/木工）+ 新故障类型与技能
-- 说明：新楼栋的 layout_json 直接复用 1 号宿舍楼，即"标准层设计"

-- 1) 新故障类型：泥瓦维修 / 木维修
USE order_db;
INSERT INTO fault_types(code, name, category, auto_dispatch, parent_id, sort, status)
VALUES ('masonry','泥瓦维修','masonry',1,0,40,1), ('wood','木维修','wood',1,0,50,1)
ON DUPLICATE KEY UPDATE name = VALUES(name), category = VALUES(category), auto_dispatch = VALUES(auto_dispatch);

-- 2) 标准层模板变量
SET @tpl_layout := (SELECT layout_json FROM map_db.buildings WHERE code = 'A1' LIMIT 1);
SET @tpl_layout := COALESCE(@tpl_layout, (SELECT layout_json FROM map_db.buildings WHERE id = 1 LIMIT 1));
SET @tpl_w := (SELECT width FROM map_db.buildings ORDER BY id LIMIT 1);
SET @tpl_h := (SELECT height FROM map_db.buildings ORDER BY id LIMIT 1);
SET @tpl_f := (SELECT floors FROM map_db.buildings ORDER BY id LIMIT 1);
SET @tpl_fh := (SELECT floor_height FROM map_db.buildings ORDER BY id LIMIT 1);
SET @tpl_rp := (SELECT rooms_per_floor FROM map_db.buildings ORDER BY id LIMIT 1);

-- 3) 16 栋宿舍楼（A1..A8 / B1..B8），layout_json 复用标准层
INSERT INTO map_db.buildings(code, name, pos_x, pos_y, width, height, floors, floor_height, rooms_per_floor, layout_json)
SELECT CONCAT(t.prefix, t.n), CONCAT(t.prefix, t.n, ' 宿舍楼'),
       t.px, t.py, @tpl_w, @tpl_h, @tpl_f, @tpl_fh, @tpl_rp, @tpl_layout
FROM (
  SELECT 'A' AS prefix, a.n AS n, 140 + a.n * 130 AS px, 170 AS py
    FROM (SELECT 1 n UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8) a
  UNION ALL
  SELECT 'B' AS prefix, b.n AS n, 140 + b.n * 130 AS px, 430 AS py
    FROM (SELECT 1 n UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8) b
) t
WHERE NOT EXISTS (
  SELECT 1 FROM (SELECT code FROM map_db.buildings) x WHERE x.code = CONCAT(t.prefix, t.n)
);

-- 4) 新技能
INSERT IGNORE INTO worker_db.skills(name) VALUES ('泥瓦维修'), ('木维修');

-- 5) 10 名工人（密码复用现有账号 admin123 的哈希，默认并发 8）
SET @pwd := (SELECT password FROM worker_db.users WHERE username = 'worker1' LIMIT 1);
INSERT INTO worker_db.users(username, password, role, job_type, name, phone, building_id, status, max_concurrent)
SELECT t.username, @pwd, 2, t.job_type, t.name, NULL, NULL, 1, 8
FROM (
  SELECT 'water1' username, 2 job_type, '水工·刘师傅' name UNION ALL
  SELECT 'water2', 2, '水工·周师傅' UNION ALL
  SELECT 'water3', 2, '水工·吴师傅' UNION ALL
  SELECT 'water4', 2, '水工·郑师傅' UNION ALL
  SELECT 'elec1', 1, '电工·孙师傅' UNION ALL
  SELECT 'elec2', 1, '电工·马师傅' UNION ALL
  SELECT 'elec3', 1, '电工·朱师傅' UNION ALL
  SELECT 'mason1', 3, '泥瓦工·胡师傅' UNION ALL
  SELECT 'wood1', 4, '木工·林师傅' UNION ALL
  SELECT 'wood2', 4, '木工·何师傅'
) t
WHERE NOT EXISTS (
  SELECT 1 FROM (SELECT username FROM worker_db.users) u WHERE u.username = t.username
);

-- 6) 管辖楼栋（每栋至少 1 名水电 + 1 名其他工种）
INSERT IGNORE INTO worker_db.worker_buildings(worker_id, building_id)
SELECT u.id, b.id
FROM (
  SELECT 'water1' un, 'A1' bc UNION ALL SELECT 'water1','A2' UNION ALL SELECT 'water1','B5'
  UNION ALL SELECT 'water2','A3' UNION ALL SELECT 'water2','A4' UNION ALL SELECT 'water2','B6'
  UNION ALL SELECT 'water3','A5' UNION ALL SELECT 'water3','A6' UNION ALL SELECT 'water3','B7'
  UNION ALL SELECT 'water4','A7' UNION ALL SELECT 'water4','A8' UNION ALL SELECT 'water4','B8'
  UNION ALL SELECT 'elec1','B1' UNION ALL SELECT 'elec1','B2' UNION ALL SELECT 'elec1','A1'
  UNION ALL SELECT 'elec2','B3' UNION ALL SELECT 'elec2','B4' UNION ALL SELECT 'elec2','A2'
  UNION ALL SELECT 'elec3','B5' UNION ALL SELECT 'elec3','B6' UNION ALL SELECT 'elec3','A3'
  UNION ALL SELECT 'mason1','B7' UNION ALL SELECT 'mason1','B8' UNION ALL SELECT 'mason1','A4' UNION ALL SELECT 'mason1','A5'
  UNION ALL SELECT 'wood1','A6' UNION ALL SELECT 'wood1','A7' UNION ALL SELECT 'wood1','B1' UNION ALL SELECT 'wood1','B3'
  UNION ALL SELECT 'wood2','A8' UNION ALL SELECT 'wood2','B2' UNION ALL SELECT 'wood2','B4' UNION ALL SELECT 'wood2','B6'
) m
JOIN worker_db.users u ON u.username = m.un
JOIN map_db.buildings b ON b.code = m.bc;

-- 7) 按工种补技能（熟练度 3）
INSERT IGNORE INTO worker_db.worker_skills(worker_id, skill_id, proficiency)
SELECT u.id, s.id, 3
FROM worker_db.users u
JOIN worker_db.skills s ON s.name = CASE u.job_type
  WHEN 1 THEN '电维修' WHEN 2 THEN '水维修' WHEN 3 THEN '泥瓦维修' WHEN 4 THEN '木维修' END
WHERE u.role = 2 AND u.job_type BETWEEN 1 AND 4;