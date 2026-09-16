-- V6.1 区域概览「我的画布」模板库
-- 说明：campus_layouts 仍是"当前生效地图"（宿管端/工人端只读读取），
--       本表存放管理员命名的快照，可随时"使用此模板"应用到当前生效地图。
create table if not exists order_db.campus_layout_templates (
  id bigint not null auto_increment primary key,
  name varchar(100) not null unique,
  grid_cols int not null default 100,
  grid_rows int not null default 80,
  layout_json longtext,
  source varchar(20) not null default 'manual',
  created_by bigint not null default 0,
  created_at datetime default current_timestamp,
  updated_at datetime default current_timestamp on update current_timestamp,
  key idx_source (source)
) engine=InnoDB default charset=utf8mb4;
