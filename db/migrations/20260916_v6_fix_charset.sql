SET NAMES utf8mb4;

-- 1) 楼栋名称：新插入的 12 栋按 code 重写（A1/A2/B1/B2 为历史楼栋，保留其原名）
USE map_db;
UPDATE buildings SET name = CONCAT(code, ' 宿舍楼') WHERE code IN ('A3','A4','A5','A6','A7','A8','B3','B4','B5','B6','B7','B8');

-- 2) 新工人姓名
USE worker_db;
UPDATE users SET name = '水工·刘师傅' WHERE username = 'water1';
UPDATE users SET name = '水工·周师傅' WHERE username = 'water2';
UPDATE users SET name = '水工·吴师傅' WHERE username = 'water3';
UPDATE users SET name = '水工·郑师傅' WHERE username = 'water4';
UPDATE users SET name = '电工·孙师傅' WHERE username = 'elec1';
UPDATE users SET name = '电工·马师傅' WHERE username = 'elec2';
UPDATE users SET name = '电工·朱师傅' WHERE username = 'elec3';
UPDATE users SET name = '泥瓦工·胡师傅' WHERE username = 'mason1';
UPDATE users SET name = '木工·林师傅' WHERE username = 'wood1';
UPDATE users SET name = '木工·何师傅' WHERE username = 'wood2';

-- 3) 新故障类型名称
USE order_db;
UPDATE fault_types SET name = '泥瓦维修' WHERE code = 'masonry';
UPDATE fault_types SET name = '木维修' WHERE code = 'wood';
SELECT code, name, category FROM fault_types ORDER BY sort, id;
SET NAMES utf8mb4;
USE order_db;
UPDATE fault_types SET name = '照明维修' WHERE code = 'lamp';