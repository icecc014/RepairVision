-- V7 公共报修：报修人身份/联系方式、房间查询索引、敏感词表
alter table order_db.orders
  add column reporter_type tinyint not null default 0 comment '0 宿管 1 学生 2 教师 3 其他',
  add column reporter_contact varchar(32) null comment '报修人联系方式（选填）';

alter table order_db.orders
  modify column reporter_id bigint not null default 0 comment '宿管用户ID；公共报修为 0',
  modify column source varchar(20) default 'dormitory' comment '来源：dormitory/admin/public';

alter table order_db.orders add key idx_room_created (building_id, room, created_at);

create table if not exists order_db.sensitive_words (
  id bigint not null auto_increment primary key,
  word varchar(32) not null unique,
  created_at datetime default current_timestamp
) engine=InnoDB default charset=utf8mb4;

insert ignore into order_db.sensitive_words(word) values
  ('广告'), ('代刷'), ('加微信'), ('博彩');
