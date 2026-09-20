-- V9.8 方案A：密码版本号（改密码后旧 JWT 立即失效）
-- 目标库：worker_db（users 表由 worker-rpc 管理，order-api 只经 RPC 访问）
-- 背景：当前为无状态 JWT，改密码不会让已签发 token 失效；加版本号后，
--       鉴权中间件比对 token 内 pwd_version 与库中值，不一致即视为登录态失效。
-- 执行方式：按 db/migrations 现有约定手工按序执行（迁移脚本不自动运行）。

ALTER TABLE users
  ADD COLUMN pwd_version INT NOT NULL DEFAULT 0 COMMENT '密码版本：修改密码时 +1，用于失效旧 JWT';