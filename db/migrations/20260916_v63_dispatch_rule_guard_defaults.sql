-- V6.3 派单保护线默认值放宽 + 新增"绝对单量"阈值规则
-- 说明：仅在值仍是旧默认值时才更新，避免覆盖管理员自定义过的配置。
insert into order_db.dispatch_rule_config(rule_key, rule_value, enabled, remark) values
  ('backlog_warn_min_orders', 10, 1, '预警线：待派 ≥ 该绝对单量（与人均倍数取较大者）'),
  ('backlog_guard_min_orders', 20, 1, '保护线：待派 ≥ 该绝对单量（与人均倍数取较大者）')
on duplicate key update remark = values(remark);

update order_db.dispatch_rule_config set rule_value = 5,
  remark = '预警线：待派 ≥ max(在岗×该倍数, 绝对单量) 或 最长等待超过设定小时'
  where rule_key = 'backlog_warn_ratio' and rule_value = 3;
update order_db.dispatch_rule_config set rule_value = 10,
  remark = '保护线：待派 ≥ max(在岗×该倍数, 绝对单量) 或 最长等待超过设定小时（暂停自动派单）'
  where rule_key = 'backlog_guard_ratio' and rule_value = 5;
update order_db.dispatch_rule_config set rule_value = 6,
  remark = '预警线：存在等待超过该小时数未派出的工单'
  where rule_key = 'backlog_warn_hours' and rule_value = 2;
update order_db.dispatch_rule_config set rule_value = 12,
  remark = '保护线：存在等待超过该小时数未派出的工单（暂停自动派单）'
  where rule_key = 'backlog_guard_hours' and rule_value = 4;
