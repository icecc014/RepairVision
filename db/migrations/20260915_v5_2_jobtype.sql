-- V5.2 工种分类与报修类型治理
-- 1) 工人分电工 / 水工 / 通用；2) 故障类型分电 / 水 / 其他，其他类型不自动派单；
-- 3) 工单增加"待管理员处置"标记（非电非水或其他类型时置 1，跳过自动派单）。

USE worker_db;

ALTER TABLE users
  ADD COLUMN job_type TINYINT NOT NULL DEFAULT 0 COMMENT '工种：0通用 1电工 2水工' AFTER role;

USE order_db;

ALTER TABLE fault_types
  ADD COLUMN category VARCHAR(20) NOT NULL DEFAULT 'other' COMMENT '类别：electric/water/other' AFTER name,
  ADD COLUMN auto_dispatch TINYINT NOT NULL DEFAULT 1 COMMENT '是否参与自动派单' AFTER category;

ALTER TABLE orders
  ADD COLUMN manual_review TINYINT NOT NULL DEFAULT 0 COMMENT '1=待管理员处置（不自动派单）' AFTER is_merged;

-- 现网故障类型归类
UPDATE fault_types SET category = 'electric', auto_dispatch = 1 WHERE code IN ('electric', 'lamp');
UPDATE fault_types SET category = 'water', auto_dispatch = 1 WHERE code = 'water';
UPDATE fault_types SET category = 'other', auto_dispatch = 0 WHERE code = 'other';
