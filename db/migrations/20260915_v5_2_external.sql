-- V5.2 补充：工单"标记外援"
-- 其他类型工单不自动派单，管理员可协商内部工人或标记为外援处理。

USE order_db;

ALTER TABLE orders
  ADD COLUMN external_mark TINYINT NOT NULL DEFAULT 0 COMMENT '1=已转外援/外部人员处理' AFTER manual_review;