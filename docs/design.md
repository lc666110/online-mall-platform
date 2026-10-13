# online-mall-platform 架构设计草案

日期：2026-10-13
目标：可运行、可部署、分模块测试并同步 GitHub 的商城。源码与备份保存在 D:/bbbbbbbbbbb/store。

## 架构选择

推荐模块化单体：frontend 用户商城、admin 管理端、server Go API，共同使用 /api/v1。相比微服务，事务更可靠且部署成本更低；相比将用户端与后台合并，独立构建便于权限和发布边界管理。
用户端 Vue 3 + TypeScript + Vite；后台 Vue 3 + TypeScript + Vite + Naive UI + ECharts；服务 Go + Gin + GORM；数据库 MySQL；缓存 Redis。不得替换为 SQLite 或内存缓存来声称完成联调。
server 采用 main.go、config、controller、service、model、middleware、router、database，controller 负责参数与响应，service 负责业务事务，model 负责关系和字段。

## 数据设计

核心表 users、products、categories、orders、order_items、cart_items、reviews。
增加 product_variants 保存真实 SKU 规格和库存；after_sales 保存售后申请及审批；必要时 role_permissions 保存权限映射。
users 包含唯一邮箱、密码哈希、昵称、头像、角色、启用状态、时间戳；返回数据不含密码。
products 包含分类、名称、描述、图片、上下架状态；SKU 包含唯一 sku、规格、价格分、库存。
orders 包含唯一订单号、用户、状态、金额分、收货信息快照、物流信息、时间戳；order_items 保存商品和 SKU 快照、数量、成交单价。
cart_items 对 user_id + variant_id 建唯一索引。reviews 绑定用户、商品和已完成的 order_item，评分 1–5，每个订单明细最多一次评价。
使用外键、搜索和列表索引；GORM 自动迁移与初始化 SQL 一致。金额使用整数分，时间统一 UTC。

## API 与业务

统一 /api/v1 JSON 响应、分页、参数验证、错误码。
/auth/register、login、logout；/me 资料；/categories、products、products/:id；/cart；/orders；/reviews。
/admin 下提供商品、SKU、分类、订单、用户、权限、售后和统计接口。
JWT 带会话标识，Redis 校验会话、注销失效；密码 bcrypt；后台 RBAC、账号禁用检查；所有用户资源强制按登录用户过滤。
订单状态 pending -> confirmed -> shipped -> completed；pending/confirmed 可按规则取消；售后独立记录并由后台审核。用户不能直接指定任意状态。
创建订单使用 MySQL 事务、库存条件更新或行锁、服务端重算价格、幂等键；取消仅恢复一次库存。禁止信任客户端金额。
没有支付服务凭据时，不实现虚假支付。默认支持人工确认订单；真实在线收款需接入指定渠道及验签回调。

## Redis 实际使用

登录会话带 TTL；热点商品读缓存并在编辑后失效；购物车使用 MySQL 持久化和 Redis 读缓存并在修改后失效；访问记录使用带 TTL 的 Redis 计数/有界记录。
缓存故障时商品和购物车回源；会话认证故障安全拒绝。限制登录尝试并返回合理错误。

## UI

用户端响应式首页、分类搜索、商品图片与 SKU 选择、购物车、结算、订单列表详情、个人资料和评价。分页与路由懒加载、加载/空/错误状态。
后台 Naive UI 布局、表单验证、商品分类库存、订单物流售后、用户角色页面。ECharts 展示实际数据库汇总销售额、订单、用户增长、销量排行和热销分析。

## 测试与审计

各模块执行测试后才能进入下一阶段，记录具体命令与结果。
后端覆盖鉴权、资源归属、权限、输入错误、订单幂等、库存并发、状态流转和评价资格。真实 MySQL + Redis 集成测试。
前端与后台 typecheck、构建、关键流程测试；浏览器验证注册登录、SKU 加购、下单、管理发货、收货评价、统计。
审计不以编译成功替代完整业务验证；失败先修复再推进。

## 部署与交付

Docker Compose 包含 MySQL、Redis、Go API 和静态前端/后台反向代理，健康检查、持久卷、环境配置示例、无硬编码生产凭据。
README、API 文档、初始化 SQL、部署/备份恢复说明、CI 检查与模块验收记录。
GitHub 仓库名 online-mall-platform，默认私有保护初期配置。密钥和真实 .env 不上传。按阶段提交，同步验证远端提交。
源码位于 store/online-mall-platform，阶段备份位于 store/backups；不覆盖 store 下已有项目。
公开在线运行需确定部署主机、域名与 TLS；GitHub 上传源码本身不等于公网部署。

## 开发顺序

1 架构设计；2 数据库；3 API 契约；4 后端；5 用户端；6 后台；7 联调；8 部署配置。每阶段保存验收记录与备份。

## 待确认

采用上述架构和人工确认订单流程；仓库默认私有。部署目标未提供，先交付可一键运行的 Docker 部署并完成可用环境下的实际联调。
