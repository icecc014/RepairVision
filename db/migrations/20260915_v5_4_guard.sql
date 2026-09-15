-- V5.4 双阈值保护：预留"预警线 / 保护线"与自动派单开关配置（可在"派单规则"页调整）
USE order_db;

INSERT INTO dispatch_rule_config(rule_key, rule_value, enabled, remark) VALUES
  ('backlog_warn_ratio', 3, 1, '预警线：待派工单数 > 在岗人数 × 该倍数'),
  ('backlog_guard_ratio', 5, 1, '保护线：待派工单数 > 在岗人数 × 该倍数（暂停自动派单）'),
  ('backlog_warn_hours', 2, 1, '预警线：存在等待超过该小时数未派出的工单'),
  ('backlog_guard_hours', 4, 1, '保护线：存在等待超过该小时数未派出的工单（暂停自动派单）'),
  ('auto_dispatch_paused', 0, 1, '1 = 已暂停自动派单（保护模式，人工处置）')
ON DUPLICATE KEY UPDATE remark = VALUES(remark);