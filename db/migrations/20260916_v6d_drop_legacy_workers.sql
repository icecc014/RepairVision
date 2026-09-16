SET NAMES utf8mb4;
USE worker_db;
DELETE FROM worker_skills WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM worker_buildings WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM worker_schedules WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM worker_leave_requests WHERE worker_id IN (SELECT id FROM users WHERE username IN ('worker1','worker2','worker3','worker4'));
DELETE FROM users WHERE username IN ('worker1','worker2','worker3','worker4');
SELECT COUNT(*) AS workers FROM users WHERE role = 2;