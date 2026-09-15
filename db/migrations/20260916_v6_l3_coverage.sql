USE worker_db;
-- V6 L3-a：4 类工种覆盖全部 16 栋楼（每栋楼 4 类工种各至少 1 名候选人）
DELETE wb FROM worker_db.worker_buildings wb
JOIN worker_db.users u ON u.id = wb.worker_id
WHERE u.username IN ('water1','water2','water3','water4','elec1','elec2','elec3','mason1','wood1','wood2');

INSERT IGNORE INTO worker_db.worker_buildings(worker_id, building_id)
SELECT u.id, b.id FROM (
  SELECT 'water1' un, 'A1' bc UNION ALL SELECT 'water1','A2' UNION ALL SELECT 'water1','A3' UNION ALL SELECT 'water1','A4'
  UNION ALL SELECT 'water2','A5' UNION ALL SELECT 'water2','A6' UNION ALL SELECT 'water2','A7' UNION ALL SELECT 'water2','A8'
  UNION ALL SELECT 'water3','B1' UNION ALL SELECT 'water3','B2' UNION ALL SELECT 'water3','B3' UNION ALL SELECT 'water3','B4'
  UNION ALL SELECT 'water4','B5' UNION ALL SELECT 'water4','B6' UNION ALL SELECT 'water4','B7' UNION ALL SELECT 'water4','B8'
  UNION ALL SELECT 'elec1','A1' UNION ALL SELECT 'elec1','A2' UNION ALL SELECT 'elec1','A3' UNION ALL SELECT 'elec1','A4' UNION ALL SELECT 'elec1','A5' UNION ALL SELECT 'elec1','A6'
  UNION ALL SELECT 'elec2','A7' UNION ALL SELECT 'elec2','A8' UNION ALL SELECT 'elec2','B1' UNION ALL SELECT 'elec2','B2' UNION ALL SELECT 'elec2','B3' UNION ALL SELECT 'elec2','B4'
  UNION ALL SELECT 'elec3','B5' UNION ALL SELECT 'elec3','B6' UNION ALL SELECT 'elec3','B7' UNION ALL SELECT 'elec3','B8'
  UNION ALL SELECT 'mason1','A1' UNION ALL SELECT 'mason1','A2' UNION ALL SELECT 'mason1','A3' UNION ALL SELECT 'mason1','A4'
  UNION ALL SELECT 'mason1','A5' UNION ALL SELECT 'mason1','A6' UNION ALL SELECT 'mason1','A7' UNION ALL SELECT 'mason1','A8'
  UNION ALL SELECT 'mason1','B1' UNION ALL SELECT 'mason1','B2' UNION ALL SELECT 'mason1','B3' UNION ALL SELECT 'mason1','B4'
  UNION ALL SELECT 'mason1','B5' UNION ALL SELECT 'mason1','B6' UNION ALL SELECT 'mason1','B7' UNION ALL SELECT 'mason1','B8'
  UNION ALL SELECT 'wood1','A1' UNION ALL SELECT 'wood1','A2' UNION ALL SELECT 'wood1','A3' UNION ALL SELECT 'wood1','A4'
  UNION ALL SELECT 'wood1','A5' UNION ALL SELECT 'wood1','A6' UNION ALL SELECT 'wood1','A7' UNION ALL SELECT 'wood1','A8'
  UNION ALL SELECT 'wood2','B1' UNION ALL SELECT 'wood2','B2' UNION ALL SELECT 'wood2','B3' UNION ALL SELECT 'wood2','B4'
  UNION ALL SELECT 'wood2','B5' UNION ALL SELECT 'wood2','B6' UNION ALL SELECT 'wood2','B7' UNION ALL SELECT 'wood2','B8'
) m
JOIN worker_db.users u ON u.username = m.un
JOIN map_db.buildings b ON b.code = m.bc;