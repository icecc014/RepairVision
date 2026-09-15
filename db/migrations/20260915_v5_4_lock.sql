-- V5.4 工单锁定：锁定后不参与自动派单/批量派单（管理员仍可手动改派）
USE order_db;

ALTER TABLE orders
  ADD COLUMN dispatch_locked TINYINT NOT NULL DEFAULT 0 COMMENT '1=锁定，不参与自动调整' AFTER external_mark;