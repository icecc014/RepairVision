# 访问入口设计：宿管端 / 工人端 HTTP 直达（含多端口方案）

> 背景：2026-09-17 需求“把宿管端和工人端都改为 http 的方式访问”。
> 结论先说：这两个入口本来就是 http（Nginx 只监听 80，没有 https），当前用 `http://<host>:8080/m/dorm` 与 `/m/worker` 即可访问。
> 真正需要补的是“每个角色一个可直接打开、可贴二维码的短网址”，本文给出设计与实施方案。

## 1. 现状实测（2026-09-17，容器运行中）

| URL | 结果 | 说明 |
| --- | --- | --- |
| `http://localhost:8080/admin/` | 200 | PC 管理端 |
| `http://localhost:8080/m/` | 200 | 移动端入口（登录页选角色） |
| `http://localhost:8080/m/dorm` | 200 | 宿管端 |
| `http://localhost:8080/m/worker` | 200 | 工人端 |
| `http://10.1.97.149:8080/m/dorm` | 200 | 局域网 IP 直连（手机同一 WiFi 可开） |

仓库事实：`nginx.conf` 只有 `listen 80`（无 TLS 配置）；`docker-compose.yml` 的 web 服务映射 `${WEB_PORT:-8080}:80`；移动端 `createWebHistory('/m/')`，角色路由 `/m/dorm`、`/m/worker`，路由守卫已按角色校验。

## 2. 设计目标

1. 每个角色一个短网址，可直接打开、可贴二维码（宿管端 / 工人端 / 公共报修）；
2. 不改变现有 8080 入口与前端构建产物；
3. 手机（同 WiFi）与微信内置浏览器可直接访问；
4. 与 V7 公共报修页共用同一套入口体系；
5. 角色隔离不变：宿管访问 `/m/worker` 仍会被弹回登录页。

## 3. 方案对比

| 方案 | 入口 | 优点 | 缺点 |
| --- | --- | --- | --- |
| A（推荐） | `:8081/dorm`、`:8081/worker`、`:8081/report` | 只多开一个端口，防火墙与校园网限制少，三个短网址 | 三个入口同端口 |
| B（最直观） | `:8082/dorm`、`:8083/worker`、`:8081/report` | 一个角色一个端口，演示最直观 | 多开两个端口，防火墙需放行多个 |
| C（零改动） | `:8080/m/dorm`、`:8080/m/worker` | 无需改动 | 网址长，不像独立端 |

推荐 A，并在 8080 上也同时提供短路径（`/dorm`、`/worker`、`/report`），这样即使不改 compose 也能用短网址。

## 4. 实现设计

### 4.1 Nginx：短路径 302 到 SPA 路由

```nginx
# 角色直达短入口：302 到移动端 SPA 路由（Vue Router base 保持 /m/，不改前端构建）
location = /dorm   { return 302 /m/dorm; }
location = /worker { return 302 /m/worker; }
location = /report { return 302 /m/report; }   # V7 公共报修页
```

说明：移动端是 history 模式且 base 为 `/m/`，直接把 `/dorm/` 映射成静态目录会导致前端路由解析异常，因此采用跳转方式，既短又不改构建。

### 4.2 端口（方案 A）

```yaml
  web:
    ports:
      - "${WEB_PORT:-8080}:80"      # 管理端 + 移动端（保持）
      - "${PUBLIC_PORT:-8081}:80"   # 公共入口：/dorm、/worker、/report
```

方案 B 再增加 `${DORM_PORT:-8082}:80`、`${WORKER_PORT:-8083}:80`；Nginx 对这两个端口不区分（同一 server），网址形如 `http://<host>:8082/dorm`。

### 4.3 可选增强

- 登录页支持 `?role=dorm|worker` 预设：从 `/dorm` 进入只显示宿管登录，从 `/worker` 进入只显示工人登录（改动仅登录页一处）；
- 管理端增加入口二维码页：三个网址生成二维码，演示直接扫；
- `scripts/up.ps1` 启动后打印入口地址（含局域网 IP），现场不用找 IP。

## 5. 手机 / 微信访问要点

1. 用局域网 IP 而不是 localhost：本机当前为 `10.1.97.149`（以太网），例如 `http://10.1.97.149:8081/dorm`；
2. Windows 防火墙放行端口（示例）：`netsh advfirewall firewall add rule name="RepairVision 8081" dir=in action=allow protocol=TCP localport=8081`；
3. 手机与电脑需同一 WiFi/网段；若校园网开启 AP 隔离（客户端互访被禁），可改用手机热点让电脑连热点；
4. 若浏览器把 http 自动升级为 https 导致打不开：清站点缓存或显式输入 `http://`（Chrome 的“始终使用安全连接”可临时关闭）；本系统无 HTTPS，不存在证书问题；
5. 微信内置浏览器可直接打开 http 局域网地址；若提示“非官方网页”，点继续访问即可。

## 6. 验收标准

| 项 | 验收方式 |
| --- | --- |
| 短入口可用 | `curl -I http://localhost:8080/dorm` 返回 302 → `/m/dorm`，浏览器打开是宿管登录页 |
| 独立端口 | `http://<局域网IP>:8081/worker` 在手机上可打开并登录工人账号 |
| 角色隔离 | 宿管账号访问 worker 入口被弹回登录页 |
| 公共报修 | V7 完成后 `http://<局域网IP>:8081/report` 免登录可提交与查询 |
| 原入口不受影响 | `8080/admin/`、`8080/m/` 仍返回 200 |

## 7. 实施计划（小改动，1~2 次提交）

1. `nginx.conf` 增加 `/dorm`、`/worker`（V7 一并加 `/report`）三条 302；
2. `docker-compose.yml` 的 web 服务增加 `${PUBLIC_PORT:-8081}:80`（方案 A）；
3. 重建 web 容器：`docker compose up -d --build web`，异常时再加 `--force-recreate web`；
4. `scripts/up.ps1` 输出入口地址；
5. 可选：登录页角色预设 + 管理端入口二维码页。
