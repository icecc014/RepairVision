# RepairVision — 校园维修工单智能调度与 2D/3D 混合可视化系统

毕业设计项目。实现“宿管报修 → 自动派单 → 工人维修 → 实时跟踪 → 数据看板”的完整闭环，并在此基础上提供楼宇 2D/3D 可视化。

## 技术栈

| 类别 | 技术 |
| --- | --- |
| 后端微服务 | Go 1.22 + go-zero v1.7.2 |
| RPC | gRPC / Etcd 服务发现 |
| 数据库 | MySQL 8.0.37（单实例 3 schema） |
| 缓存 / MQ | Redis 7.2 / Kafka（KRaft 单节点） |
| 前端 | Vue3 + TypeScript + Vite |
| PC 管理端 | Element Plus |
| 移动端 H5 | Vant |
| 可视化 | Konva.js 2D + Three.js 3D |
| 部署 | Docker Compose + Nginx |

## 演示账号

| 角色 | 账号 | 密码 |
| --- | --- | --- |
| 超级管理员 | admin | admin123 |
| 宿管（1号楼） | dorm1 | admin123 |
| 宿管（2号楼） | dorm2 | admin123 |
| 宿管（3号楼） | dorm3 | admin123 |
| 宿管（4号楼） | dorm4 | admin123 |
| 工人 | worker1 | admin123 |
| 工人 | worker2 | admin123 |
| 工人 | worker3 | admin123 |
| 工人 | worker4 | admin123 |

## 快速启动

```bash
# 1. 复制环境变量（首次）
copy .env.example .env   # Windows
cp .env.example .env     # Linux/macOS

# 2. 修改 .env 中的 MYSQL_ROOT_PASSWORD / JWT_SECRET

# 3. 一键启动（自动构建 8 个容器并初始化种子数据）
./scripts/up.ps1         # Windows PowerShell
docker compose up -d --build
```

启动后访问：

- PC 管理端：http://localhost:8080/admin/
- 移动端登录：http://localhost:8080/m/
  - 宿管入口：http://localhost:8080/m/dorm
  - 工人入口：http://localhost:8080/m/worker

> 登录态使用 sessionStorage，不同角色可分别开标签页互不影响。

## 页面入口

### PC 管理端 `/admin/`

- 工单总览（含派单评分）
- 数据统计看板
- 人员账号管理
- 建筑信息管理
- 维修类型字典
- 派单规则配置
- 操作日志

### 移动端 H5 `/m/`

- 宿管：本栋工单列表 / 极简报修（只填房间号自动推导楼层）/ 取消未开工工单
- 工人：我的工单（开工/完工）/ 报修地图（2D 楼栋 + 红点徽标 + 3D 楼宇 + 按类型批量完工）

## 功能与阶段进度

| 阶段 | 内容 | 状态 |
| --- | --- | --- |
| P0 | 工程骨架 / 三服务 / 双前端 / 一键启动 | ✅ |
| P1 | MVP 业务闭环：报修→派单→维修→完工 | ✅ |
| P2 | 管理端 CRUD / 去重 / 加权派单 / 日志 / WebSocket | ✅ |
| P3 | 2D 地图 / 3D 楼宇 / 批量完工 / fault_markers 同步 | ✅ |
| P4 | 统计看板 / Kafka 异步日志 / 论文与演示材料 | 部分完成 |


## V3（feature/v3 分支 · 开发中）

- V3.1 派单升级：独立派单引擎（当班/最大并发约束、分钟级负载分）、管理员人工改派、开工即标记已接受派单。
- V3.1 派单工作台：批量派单（容量槽位 + 最小代价指派）、PC 手动派单/改派、工人最大并发配置。
- V3.2 轮休排班：一键生成周排班（每人每周休 1 天）、PC 排班日历可改、休息日不派单、工人端“我的班次”。
- V3.0 领域数据：新增第 4 栋测试楼与 dorm4/worker3/worker4；
  工单增加优先级、预计时长与派单/开工/完成时间点；
  服务端房间号合法性校验；worker_db 预留 max_concurrent 与 worker_schedules 排班表。

## V2（feature/v2 分支）

- 3D 楼宇按房间单元化建模：户型平面、半透明外墙、房间号、楼层筛选、透视开关。
- 3D 支持拖拽/缩放/自动旋转，点击房间或红点查看工单，并可直接开工/完工联动地图。
- 移动端 `/m/dorm` 与 `/m/worker` 路由隔离；登录态按标签页隔离。
- 宿管报修只填房间号，自动推导楼层并实时预览。
- PC 操作日志支持时间范围筛选与 CSV 导出；新增“权限说明”页。
## 架构

```text
PC Admin ─┐
          ├─ Nginx :8080 ── /api ── order-api :8888 (HTTP)
Mobile H5 ┘                └─ /ws ── order-api :8890 (WebSocket)

order-api ── gRPC ── worker-rpc :9001 ── worker_db
         └─ gRPC ── map-rpc :9002 ── map_db
         └─ SQL ── order_db
         └─ Kafka topic operation-logs（异步操作日志）
```

数据隔离：worker_db（账号/技能/楼栋管辖）、map_db（楼宇/故障标记）、order_db（工单/派单记录/字典/规则/日志）。禁止跨库直连，跨库数据一律通过 RPC 获取。

## 目录结构

```text
AGENTS/          项目规约、技术数据库、路线图、工作日志
db/init/         三库 DDL 初始化
deploy/docker/   服务镜像
services/
  order/         order-api HTTP + WebSocket + Kafka 日志
  worker/        worker-rpc 人员/技能/管辖
  map/           map-rpc 楼宇/故障标记
frontend/
  pc-admin/      PC 管理端
  mobile-web/    H5 移动端
scripts/         一键启动/停止/日志
```

## 注意事项

- `.env` 已被 gitignore，请勿提交真实密码与密钥。
- 依赖版本全部锁定，禁止随意升级/更换第三方库。
- 楼栋删除会做工单引用保护；账号“删除”实际为停用，保留历史数据。