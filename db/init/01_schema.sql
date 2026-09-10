-- RepairVision P0/P1 基础库结构（与 AGENTS/02_技术数据库.md 对齐）
-- 单 MySQL 实例承载 order_db / worker_db / map_db 三个独立库

CREATE DATABASE IF NOT EXISTS order_db CHARACTER SET utf8mb4;
USE order_db;

CREATE TABLE orders (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_no VARCHAR(32) NOT NULL UNIQUE COMMENT 'REP+日期+流水号',
  title VARCHAR(100) NOT NULL,
  description TEXT NOT NULL,
  building_id INT NOT NULL,
  room VARCHAR(20) NOT NULL,
  floor TINYINT NOT NULL DEFAULT 1 COMMENT '故障楼层，3D故障点定位用',
  fault_type VARCHAR(20) NOT NULL COMMENT '电维修/水维修/其他',
  priority TINYINT NOT NULL DEFAULT 1 COMMENT '1普通 2紧急',
  expect_minutes INT NOT NULL DEFAULT 30 COMMENT '预计维修时长(分钟)',
  status TINYINT DEFAULT 1 COMMENT '1待派单 2已派单 3维修中 4已完成 5已取消',
  is_merged TINYINT DEFAULT 0 COMMENT '0独立单 1主单(已合并子单) 2子单(被并入主单，不再派单)',
  main_order_id BIGINT DEFAULT NULL COMMENT '子单指向主单ID；独立单/主单为NULL',
  worker_id BIGINT DEFAULT NULL,
  dispatched_at DATETIME DEFAULT NULL COMMENT '派单时间',
  started_at DATETIME DEFAULT NULL COMMENT '工人开工时间',
  completed_at DATETIME DEFAULT NULL COMMENT '完成时间',
  reporter_id BIGINT NOT NULL COMMENT '宿管用户ID',
  source VARCHAR(20) DEFAULT 'dormitory' COMMENT '来源：dormitory/admin',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_building_status (building_id, status),
  KEY idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE order_duplicate_records (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  main_order_id BIGINT NOT NULL,
  sub_order_id BIGINT NOT NULL,
  similarity_score DECIMAL(5,4) DEFAULT 0.00,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_main_sub (main_order_id, sub_order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE dispatch_records (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id BIGINT NOT NULL,
  worker_id BIGINT NOT NULL,
  score DECIMAL(5,4) NOT NULL,
  skill_score DECIMAL(5,4),
  distance_score DECIMAL(5,4),
  load_score DECIMAL(5,4),
  status TINYINT DEFAULT 1 COMMENT '1派单中 2已接受 3已拒绝',
  dispatched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  KEY idx_order (order_id),
  KEY idx_worker (worker_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE fault_types (
  id INT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(20) NOT NULL UNIQUE COMMENT '如 electric/water/other',
  name VARCHAR(50) NOT NULL COMMENT '显示名：电维修/水维修/其他',
  parent_id INT DEFAULT 0 COMMENT '0为一级分类',
  sort INT DEFAULT 0,
  status TINYINT DEFAULT 1 COMMENT '1启用 0停用',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE dispatch_rule_config (
  id INT AUTO_INCREMENT PRIMARY KEY,
  rule_key VARCHAR(50) NOT NULL UNIQUE COMMENT 'skill_weight/distance_weight/load_weight/auto_dispatch_enabled 等',
  rule_value DECIMAL(10,4) NOT NULL DEFAULT 0,
  enabled TINYINT DEFAULT 1,
  remark VARCHAR(100),
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE operation_logs (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  op_id VARCHAR(64) NOT NULL COMMENT '幂等键（uuid），防止Kafka重复消费导致重复落库',
  user_id BIGINT,
  username VARCHAR(50),
  role TINYINT,
  module VARCHAR(50) COMMENT 'order/worker/map/user 等',
  action VARCHAR(50) COMMENT 'create/update/dispatch/complete/cancel 等',
  method VARCHAR(10),
  path VARCHAR(200),
  request_body TEXT,
  response_code INT,
  ip VARCHAR(45),
  cost_ms INT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_op_id (op_id),
  KEY idx_user_time (user_id, created_at),
  KEY idx_module_action (module, action),
  KEY idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE order_feedbacks (
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

CREATE TABLE notifications (
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
CREATE DATABASE IF NOT EXISTS worker_db CHARACTER SET utf8mb4;
USE worker_db;

CREATE TABLE users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL COMMENT 'bcrypt加密',
  role TINYINT NOT NULL COMMENT '1超级管理员 2工人 3宿管',
  name VARCHAR(50) NOT NULL,
  phone VARCHAR(15),
  building_id INT DEFAULT NULL COMMENT '宿管绑定楼栋',
  status TINYINT DEFAULT 1 COMMENT '1启用 0禁用',
  max_concurrent TINYINT NOT NULL DEFAULT 3 COMMENT '最大在途工单数',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  KEY idx_role_building (role, building_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE skills (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(50) NOT NULL UNIQUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE worker_skills (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  worker_id BIGINT NOT NULL,
  skill_id INT NOT NULL,
  proficiency TINYINT DEFAULT 2 COMMENT '1初级 2中级 3高级',
  UNIQUE KEY uk_worker_skill (worker_id, skill_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE worker_buildings (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  worker_id BIGINT NOT NULL,
  building_id INT NOT NULL,
  UNIQUE KEY uk_worker_building (worker_id, building_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE worker_schedules (
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

CREATE TABLE worker_leave_requests (
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
CREATE DATABASE IF NOT EXISTS map_db CHARACTER SET utf8mb4;
USE map_db;

CREATE TABLE buildings (
  id INT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(20) NOT NULL UNIQUE COMMENT '楼栋编号 A1',
  name VARCHAR(100) NOT NULL,
  pos_x DECIMAL(10,2) NOT NULL COMMENT '2D X坐标',
  pos_y DECIMAL(10,2) NOT NULL COMMENT '2D Y坐标',
  width DECIMAL(10,2) DEFAULT 50,
  height DECIMAL(10,2) DEFAULT 30,
  floors TINYINT DEFAULT 1 COMMENT '楼层数（3D用）',
  floor_height DECIMAL(5,2) DEFAULT 3.5 COMMENT '层高（3D用）',
  rooms_per_floor TINYINT DEFAULT 10 COMMENT '每层房间数（3D用）',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE fault_markers (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  order_id BIGINT NOT NULL UNIQUE,
  building_id INT NOT NULL,
  floor INT NOT NULL COMMENT '故障所在楼层',
  room_number VARCHAR(10) NOT NULL COMMENT '房间号如 211',
  offset_x DECIMAL(10,2) DEFAULT 0 COMMENT '3D场景内偏移',
  offset_y DECIMAL(10,2) DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  KEY idx_building_floor (building_id, floor)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;