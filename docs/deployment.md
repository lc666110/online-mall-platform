# 部署、备份与运营

## 目标环境

Linux 主机上的 Docker Engine/Compose 运行 MySQL、Redis、Go API 和两个 Nginx SPA。部署前设置 DNS：例如 `shop.example.com`、`admin.example.com`，由宿主 TLS 代理分别转发到 localhost:8088、localhost:8089。默认 Compose 端口绑定 localhost；保持 API/SQL/Redis 在内部网络。

可使用宿主已有的 Nginx/Caddy 管理证书和 TLS。在外层 Nginx 配置中将 `/` 代理到商城/后台端口，设置 Host、X-Real-IP、X-Forwarded-For、X-Forwarded-Proto。内层 Nginx 默认以自己的直接客户端地址为限流依据；在宿主代理场景须为内层配置 **真实且限定的宿主代理地址**：`set_real_ip_from <实际宿主代理 IP>; real_ip_header X-Forwarded-For; real_ip_recursive on;`，之后由内层覆盖转发头。不要填 0.0.0.0/0，也不要将任意用户提供的 X-Forwarded-For 直接信任。

Compose 专用子网默认为 172.30.88.0/24。若与目标主机现有网络冲突，应同时修改 compose.yaml 的 subnet 和 server.TRUSTED_PROXIES。确认 HTTPS、正确代理 IP 后，登录限流才会按真实用户地址区分。上线前分别用两个来源验证限流桶，检查外部无法连接 MySQL/Redis/API。

`.env` 为唯一生产环境配置入口，使用 `scripts/init-env.py` 生成随机密钥，不会覆写已有文件。部署系统应保护 .env、数据库备份及 runtime 私密文件的访问。JWT_SECRET 轮换使已有 JWT 失效；Redis 重启且数据丢失也会使会话失效。正式环境按运行要求配置 AOF/持久卷与 Redis 内存监控。

## 初次部署

1. 将私有仓库克隆到目标主机；在根目录运行 `python scripts/init-env.py`。
2. 在服务器本机编辑 `.env` 的管理员邮箱及网络端口；不将该文件提交 Git。
3. 执行 `docker compose config --quiet`、`docker compose up -d --build`。
4. 检查 `docker compose ps` 全部 healthy，执行 server `/health` 检查。
5. 使用管理员登录后台，新增分类、真实商品和 SKU，确认价格单位为分。
6. 配置域名、HTTPS、代理信任链后完成一笔真实业务验收，包括注册、下单、确认、发货、收货、评价与售后。

演示商品只通过 `docker compose exec server seed` 显式添加；生产不自动插入。现有管理员不会因环境文件变更而重置密码。需要密码重置时应在受控运维流程下更新 bcrypt 哈希和会话，当前 UI 未提供忘记密码流程。

## 数据备份

MySQL 是业务数据事实来源。使用一致性逻辑备份并保存到受控目录，例如宿主 shell：

```sh
mkdir -p backups
docker compose exec -T mysql sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysqldump -u root --single-transaction --routines --triggers --no-tablespaces "$MYSQL_DATABASE"' > backups/mall-$(date +%Y%m%d-%H%M%S).sql
```

备份包含账号邮箱/联系方式、密码哈希、收货地址等业务数据，不上传 GitHub。源码可用 `git bundle create backups/source.bundle --all` 和 `git archive -o backups/source.zip HEAD` 备份，源码档案不包含 .env、node_modules、runtime 或浏览器 trace。生产密钥由自己的密钥存储独立备份。

恢复先在独立测试环境执行，验证订单数、明细总额、库存和管理员可登录；不要直接对运行中的生产库测试恢复：

```sh
cat backups/mall-backup.sql | docker compose exec -T mysql sh -c 'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysql -u root "$MYSQL_DATABASE"'
```

Redis 缓存可重建，恢复 MySQL 后应在停机维护窗口清除对应应用缓存/会话或使用新的 Redis 实例，并重新登录。不要复用备份时刻以后的购物车/商品版本缓存。定期演练恢复并按业务恢复点目标设置备份频率。

## 升级与回滚

升级前记录当前 Git commit 和镜像、备份 MySQL。运行测试和构建，再 `docker compose up -d --build`，检查健康与关键业务。GORM AutoMigrate 会在 API 启动时同步模型；本版本新增 cache_version/cart_version 为默认0的非空字段。后续删除/更名字段、改金额类型等应使用经过审查的显式迁移。

回滚代码先确认数据库兼容旧模型；结构不兼容时使用经过验证的备份恢复，不能只换旧镜像。代码和数据库恢复应作为同一维护操作，维护期间暂停新订单。

## 上线运行边界

- 订单确认、售后退款由运营人工核实；没有微信/支付宝资金接口。
- 待确认订单立即占库存，应运营取消长期未付款订单。当前无自动关单/后台任务队列。
- 当前售后为整单退货/退款，不支持部分明细售后、多次售后、换货。
- 商品搜索使用 MySQL LIKE 与分页；大规模目录可增加全文检索。当前无 CDN、对象存储上传和消息推送。
- 单实例/小规模部署已设计并验证交易一致性，但尚未进行目标宿主压力测试、灾备/多机容量验收。
- 日志记录状态与路径，不记录登录请求体。限制日志访问和保留周期，监控 5xx、连接池、Redis 会话故障和磁盘空间。
