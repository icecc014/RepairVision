-- RepairVision V3.0 领域数据迁移（对已存在数据卷手动执行一次）
-- 执行前请先 mysqldump 备份 order_db / worker_db / map_db 三个库。
-- 执行方式示例：
--   docker compose exec -T mysql sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" < /tmp/20260909_v3_0.sql'

USE order_db;

ALTER TABLE orders
  ADD COLUMN priority TINYINT NOT NULL DEFAULT 1 COMMENT '1普通 2紧急' AFTER fault_type,
  ADD COLUMN expect_minutes INT NOT NULL DEFAULT 30 COMMENT '预计维修时长(分钟)' AFTER priority,
  ADD COLUMN dispatched_at DATETIME DEFAULT NULL COMMENT '派单时间' AFTER worker_id,
  ADD COLUMN started_at DATETIME DEFAULT NULL COMMENT '工人开工时间' AFTER dispatched_at,
  ADD COLUMN completed_at DATETIME DEFAULT NULL COMMENT '完成时间' AFTER started_at;

USE worker_db;

ALTER TABLE users
  ADD COLUMN max_concurrent TINYINT NOT NULL DEFAULT 3 COMMENT '最大在途工单数' AFTER status;

CREATE TABLE IF NOT EXISTS worker_schedules (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  worker_id BIGINT NOT NULL,
  work_date DATE NOT NULL,
  shift_type VARCHAR(20) NOT NULL DEFAULT 'DAY'
    COMMENT 'DAY全天班 MORNING午班 AFTERNOON晚班 OFF休息',
  note VARCHAR(100),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_worker_date (worker_id, work_date),
  KEY idx_date (work_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
