USE order_db;

CREATE TABLE IF NOT EXISTS order_feedbacks (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id BIGINT NOT NULL UNIQUE COMMENT '已完工工单',
  building_id INT NOT NULL,
  worker_id BIGINT DEFAULT NULL,
  rating TINYINT NOT NULL DEFAULT 5 COMMENT '1-5星',
  comment VARCHAR(500),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  KEY idx_worker (worker_id),
  KEY idx_rating (rating)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
