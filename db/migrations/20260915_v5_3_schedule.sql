-- V5.3 双休白班排班改造
-- 排班简化为"白班 / 休息"，每周休息天数默认 2（错峰轮休，可改为固定休息日），
-- 工作时段默认上午 08:00-12:00、下午 14:00-18:00，均可配置。

USE worker_db;

CREATE TABLE IF NOT EXISTS work_settings (
  id TINYINT NOT NULL PRIMARY KEY,
  rest_days_per_week TINYINT NOT NULL DEFAULT 2 COMMENT '每周休息天数 0~3',
  rest_mode VARCHAR(16) NOT NULL DEFAULT 'staggered' COMMENT 'staggered=错峰轮休 fixed=固定休息日',
  fixed_rest_weekdays VARCHAR(16) NOT NULL DEFAULT '6,7' COMMENT '固定休息日，1=周一 ... 7=周日',
  morning_start VARCHAR(5) NOT NULL DEFAULT '08:00' COMMENT '上午上班',
  morning_end VARCHAR(5) NOT NULL DEFAULT '12:00' COMMENT '上午下班',
  afternoon_start VARCHAR(5) NOT NULL DEFAULT '14:00' COMMENT '下午上班',
  afternoon_end VARCHAR(5) NOT NULL DEFAULT '18:00' COMMENT '下午下班',
  allow_force_start TINYINT NOT NULL DEFAULT 1 COMMENT '1=非工作时段允许强制开工',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT '排班与工作时段配置（单行）';

INSERT IGNORE INTO work_settings(id) VALUES (1);