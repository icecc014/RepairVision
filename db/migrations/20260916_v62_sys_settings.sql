-- V6.2 通用键值配置表（首个用途：区域概览自动备份保留份数）
create table if not exists order_db.sys_settings (
  setting_key varchar(64) not null primary key,
  setting_value varchar(255) not null default '',
  remark varchar(100) default '',
  updated_at datetime default current_timestamp on update current_timestamp
) engine=InnoDB default charset=utf8mb4;

insert ignore into order_db.sys_settings(setting_key, setting_value, remark)
values ('campus_backup_keep', '5', '区域概览自动备份保留份数（1~50）');
