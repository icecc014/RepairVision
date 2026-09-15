-- V5.1 区域概览（校园 / 建筑群总平面图）
-- 说明：为降低跨服务改动成本，区域概览数据先由 order 服务托管在 order_db，
--      对外仍提供 /api/admin/campus-layout 与 /api/campus-map 接口；
--      布局 JSON 内用 buildingId 关联楼栋，并冗余 label 名称，后续可平滑迁移到 map 服务。

USE order_db;

CREATE TABLE IF NOT EXISTS campus_layouts (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL DEFAULT '默认区域概览',
  is_default TINYINT NOT NULL DEFAULT 1 COMMENT '1=默认展示的概览图',
  grid_cols INT NOT NULL DEFAULT 40 COMMENT '网格列数',
  grid_rows INT NOT NULL DEFAULT 30 COMMENT '网格行数',
  layout_json LONGTEXT NULL COMMENT '图元区块 JSON',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_default (is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO campus_layouts(id, name, is_default, grid_cols, grid_rows, layout_json)
SELECT 1, '默认区域概览', 1, 40, 30, NULL
WHERE NOT EXISTS (SELECT 1 FROM campus_layouts WHERE id = 1);
