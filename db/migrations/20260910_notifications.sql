USE order_db;

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL COMMENT '接收用户',
  role TINYINT NOT NULL DEFAULT 0 COMMENT '接收角色 1管理员 2工人 3宿管',
  type VARCHAR(30) NOT NULL COMMENT 'create/dispatch/start/complete/cancel/feedback',
  title VARCHAR(100) NOT NULL,
  content VARCHAR(255) NOT NULL,
  order_id BIGINT DEFAULT NULL,
  is_read TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  KEY idx_user_read (user_id, is_read, id),
  KEY idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
