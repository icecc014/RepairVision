SET NAMES utf8mb4;
USE worker_db;
SET @pwd := (SELECT password FROM users WHERE username = 'dorm1' LIMIT 1);
UPDATE users SET name = CONCAT(building_id, '号宿舍管理员') WHERE role = 3 AND building_id BETWEEN 1 AND 16;
INSERT INTO users(username, password, role, job_type, name, phone, building_id, status, max_concurrent)
SELECT CONCAT('dorm', t.n), @pwd, 3, 0, CONCAT(t.n, '号宿舍管理员'), NULL, t.n, 1, 8
FROM (SELECT 5 n UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9 UNION SELECT 10
      UNION SELECT 11 UNION SELECT 12 UNION SELECT 13 UNION SELECT 14 UNION SELECT 15 UNION SELECT 16) t
WHERE NOT EXISTS (SELECT 1 FROM (SELECT username FROM users) u WHERE u.username = CONCAT('dorm', t.n));
SELECT COUNT(*) AS dorms FROM users WHERE role = 3;
SELECT username, name, building_id FROM users WHERE role = 3 ORDER BY building_id;