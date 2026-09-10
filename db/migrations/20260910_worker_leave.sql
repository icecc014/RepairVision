USE worker_db;

CREATE TABLE IF NOT EXISTS worker_leave_requests (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  worker_id BIGINT NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  reason VARCHAR(200) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1待审批 2已通过 3已驳回 4已撤销',
  reviewer_id BIGINT DEFAULT NULL,
  review_note VARCHAR(200),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_worker_status (worker_id, status),
  KEY idx_range (start_date, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
