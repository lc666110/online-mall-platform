# online-mall-platform · MORI 商城

可运行的用户商城、管理工作台和 REST API。固定技术栈：用户端 **Vue 3 + TypeScript + Vite**；后台 **Vue 3 + Naive UI + ECharts**；后端 **Go + Gin + GORM**；持久化 **MySQL**；真实缓存/会话 **Redis**。

## 已实现

- 用户注册、登录、退出、个人中心和资料管理；JWT + Redis 可撤销会话。
- 商品首页、分类、搜索、详情、多图、SKU 选择、实时库存；响应式中文 UI。
- 购物车添加/改数量/移除、服务端价格校验、事务库存扣减、幂等下单。
- 订单列表/详情、确认/发货/收货/取消、物流记录、完成订单评价与评分。
- 后台商品和分类 CRUD、上下架、库存并发检查；用户启停及角色管理；整单售后申请/审批。
- ECharts 销售额、订单量、用户增长、销量排行和热销分析，使用 MySQL 实际数据。
- GORM 自动迁移、初始化 SQL、环境示例、多阶段 Docker 构建、Compose、CI 配置与真实依赖集成测试。

当前交易模型使用**线下付款人工确认**。售后批准记录退货/退款结果并恢复库存，需要运营先核实实物退回和线下退款。本项目没有接入支付网关、自动退款、短信、邮箱验证、物流查询平台或自动关单任务；不会模拟已完成资金支付。

## 目录

```text
frontend/                 用户商城，Vue 3 + TS + Vite
admin/                    管理后台，Vue 3 + Naive UI + ECharts
shared/                   类型、HTTP 客户端、显示函数与单元测试
server/
  main.go                 HTTP 服务、优雅退出
  config/                 环境配置
  controller/             REST 输入输出、管理接口
  service/                认证、商品、购物车、订单、评价、售后事务
  model/                  GORM 实体、索引与外键
  middleware/             JWT、角色、限流、安全头、访问缓存
  router/                 路由和真实 MySQL/Redis 集成测试
  database/               连接池、自动迁移、初始化 SQL
  cmd/seed/               显式加载演示商品
deploy/                   Nginx 与 Web 镜像
scripts/                  密钥初始化、Windows 本地启动/停止、构建预览
tests/                    Playwright 浏览器业务流程
docs/                     API、架构、部署、验证报告
compose.yaml
```

## 已配置环境（Windows）

本次交付目录：`store\online-mall-platform`。Go、MySQL、Redis、Playwright 浏览器与下载工具位于同一 store 下的 `tools`，依赖缓存位于 `cache`，数据库和日志位于 `runtime`，备份位于 `backups`（`scripts/backup-local.ps1` 可生成新的源码与数据库备份）。现有 Node/Git 使用本机已有安装。

管理员邮箱和随机密码在 **store\runtime\local-secrets.json`** 的 `admin_email` / `admin_password`，请在本机查看，。新建普通用户可直接在商城注册。没有硬编码管理员密码。

```powershell
Set-Location D:\bbbbbbbbbbb\store\online-mall-platform
$env:npm_config_cache='D:\bbbbbbbbbbb\store\cache\npm'
npm ci
npm run build
powershell -ExecutionPolicy Bypass -File .\scripts\start-local.ps1
```

- 商城：[http://127.0.0.1:8088](http://127.0.0.1:8088)
- 后台：[http://127.0.0.1:8089](http://127.0.0.1:8089)
- 健康检查：[http://127.0.0.1:8080/health](http://127.0.0.1:8080/health)

`start-local.ps1 -Development` 启动 Vite 5173/5174；默认启动压缩构建的本地同源预览。`stop-local.ps1` 只停止已记录且确认属于 store 目录的进程；数据库数据保留。预览服务仅绑定 localhost，不用于公网承载。修改后端后应停止 API 再启动，脚本不会替换正在监听的旧实例。

这组 Windows 脚本使用本机已初始化的 runtime 服务，不会替其他机器自动初始化数据库或安装工具。其他机器建议使用下方 Docker 流程，或手动设置服务与环境变量。

## Docker 快速部署

需要 Linux Docker Engine + Compose（或支持 Linux 容器的 Docker Desktop），Node/Go 不需要单独安装。[GitHub CI 运行 #2](https://github.com/lc666110/online-mall-platform/actions/runs/38435468092) 已验证镜像构建、Compose 启动及 Nginx 代理后的7个浏览器业务场景。本次 Windows 环境没有 Docker daemon，本机联调使用 MySQL/Redis 原生进程；正式部署仍需在目标宿主验收自己的配置。

```sh
# 在项目根目录
python scripts/init-env.py
# 确认 .env 中管理员邮箱；随机密码/密钥已生成
docker compose config --quiet
docker compose up -d --build
docker compose ps
docker compose exec server wget -qO- http://127.0.0.1:8080/health
# 可选：仅演示环境加载 8 个示例商品；正式环境通过后台添加真实商品
docker compose exec server seed
```

默认 storefront 8088、admin 8089 仅对宿主 localhost 开放。MySQL/Redis/API 不暴露宿主端口，数据使用命名卷。`.env` 中改密码不会自动修改已初始化的 MySQL 账号；必须先正确更新数据库账号，再更新应用配置。不要使用 `docker compose down -v`，该命令删除数据库卷。

公网域名、HTTPS、备份恢复、日志和升级说明见 [docs/deployment.md](docs/deployment.md)。此源码仓库不是运行中的公网商城；完成公网部署需要目标服务器、域名和证书。

## 手动开发启动

建议 Node 22、Go 1.27、MySQL 8.4、Redis 8。先创建空数据库和专用账号，然后设置环境变量。`server/.env.example` 是配置参考，后端读取进程环境，**不会自动加载 .env 文件**。

```sh
export MYSQL_DSN='mall:<password>@tcp(127.0.0.1:3306)/mall?charset=utf8mb4&parseTime=true&loc=UTC'
export REDIS_ADDR='127.0.0.1:6379'
export REDIS_PASSWORD='<password>'
export JWT_SECRET='<random-32-or-more-characters>'
export ADMIN_EMAIL='admin@example.com'
export ADMIN_PASSWORD='<random-12-or-more-characters>'
cd server
go mod download
go run .
# 可选，在另一终端并沿用相同环境：go run ./cmd/seed
```

在项目根目录另外运行 `npm ci`，`npm run dev`、`npm run dev:admin`。Vite 代理 API 到 localhost:8080。后端首次启动自动迁移九张表，必要时也可先导入 [server/database/init.sql](server/database/init.sql)。首次配置 ADMIN_EMAIL/ADMIN_PASSWORD 创建管理员；已有账号不会被启动配置覆写。

## 验证

```sh
npm run test
npm run typecheck
npm run build
cd server
go vet ./...
go test ./... -count=1
```

后端集成测试必须设置 MYSQL_DSN 指向名称以 `_test` 结尾的**独立测试数据库**，并连接真实 Redis（测试固定 DB15，应用默认 DB0）；未设置 DSN 会跳过集成测试。测试插入独立标识的测试记录，勿指向正式业务数据库。浏览器端完整业务测试需要加载演示商品，并设置：

```sh
export MALL_TEST_CREDENTIALS='/absolute/path/private-test-credentials.json'
# JSON 仅需 admin_email、admin_password，不要提交
export MALL_STOREFRONT_URL='http://127.0.0.1:8088'
export MALL_ADMIN_URL='http://127.0.0.1:8089'
npx playwright install chromium
npm run test:e2e
```

没有凭据时后台/完整订单浏览器测试跳过，注册和移动布局测试仍运行。Playwright 失败 trace 可能包含登录凭据，默认不提交、不打入源码备份。CI 采用一次性测试环境、随机凭据，不上传这些 trace。

验证范围和已修复审计项见 [docs/verification.md](docs/verification.md)。API 使用说明见 [docs/api.md](docs/api.md)，数据一致性设计见 [docs/architecture.md](docs/architecture.md)。
