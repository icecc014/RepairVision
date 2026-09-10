-- V4-S3 楼层布局设计器
-- 为楼栋增加“标准层布局”字段：管理员在 PC 端用 2D 设计器画出的一层布局，
-- 2D 平面图与 3D 楼宇都按这份布局渲染；为空时回退到内置的 16 间标准层模板。
-- 网格编码：0=空白 / 1=房间 / 2=过道 / 3=楼梯 / 4=公共区，房间号自动生成（楼层×100+序号）。

USE map_db;

ALTER TABLE buildings
  ADD COLUMN layout_json TEXT NULL COMMENT '标准层布局(JSON)：管理员在布局设计器中绘制' AFTER rooms_per_floor;
