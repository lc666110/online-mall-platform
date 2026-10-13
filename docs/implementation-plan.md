# online-mall-platform Implementation Plan

**Goal:** 完成用户商城、后台和真实 MySQL/Redis 后端，测试、备份并上传私有 GitHub。
**Architecture:** 独立 Vue 应用 + 分层 Go REST API，事务库存和 Redis 会话。
**Tech Stack:** Vue3/TypeScript/Vite, Naive UI/ECharts, Go/Gin/GORM, MySQL, Redis。
**Spec:** 2026-10-13-online-mall-platform-design.md

## Global Constraints

新增软件、缓存和源码全部位于 D:/bbbbbbbbbbb/store；仓库私有；保留已有文件；不上传密钥。
用户已授权持续执行、安装及同步，按阶段执行并记录证据。

## Review Focus

越权读取订单；并发库存超卖；重试重复扣库存；注销会话继续使用；售后重复返库存。

### Task 1: 环境、数据库和 API 契约

- [ ] 便携工具安装，验证版本和服务连接。
- [ ] model 定义九张表，迁移并验证索引与外键；SQL 与模型一致。
- [ ] docs/api.md 定义请求、返回、权限及状态。

### Task 2: 鉴权与目录

Files: server/config, database, model, middleware, service/auth.go, controller/auth.go, router.
Interfaces: /api/v1/auth/{register,login,logout}, /me, /products, /categories。

- [ ] 首先编写真实 HTTP 集成测试：注册、重复邮箱、错误登录、注销、角色拒绝。
- [ ] 实现 JWT + Redis 会话、bcrypt、输入验证和 RBAC。
- [ ] go test、go vet 通过后提交。

### Task 3: 购物车、订单、评价与后台

Files: server/service、controller 下 cart, order, review, admin。
Interfaces: /cart, /orders, /reviews 和 /admin/*，订单使用 Idempotency-Key。

- [ ] 编写失败集成测试：SKU 加购、订单幂等、并发库存、越权、非法状态、评价资格。
- [ ] 实现持久购物车、事务下单、取消、物流、售后、CRUD 和统计。
- [ ] 全套后端测试通过并记录 Redis 键证据。

### Task 4: 用户商城

Files: frontend/src/{components,views,api,types,router}。
Consumes: Task 2/3 API；金额统一整数分。

- [ ] 编写关键 UI/接口契约测试，验证缺失行为。
- [ ] 实现响应式页面和所有用户流程，完整 TS 类型与路由懒加载。
- [ ] 类型检查、构建和浏览器流程验证。

### Task 5: 管理后台

Files: admin/src/{components,views,api,types,router}。
Consumes: /admin/*，统计基于真实业务。

- [ ] 编写后台权限、表单和物流流程测试。
- [ ] 实现 Naive UI CRUD、用户权限、售后和 ECharts。
- [ ] 类型检查、构建和浏览器后台流程验证。

### Task 6: 部署、审计与交付

- [ ] Dockerfiles、Compose、反向代理、环境示例、启动/停止/备份脚本、CI。
- [ ] 全流程注册到收货评价，并审计全部权限和边界。
- [ ] 独立代码审查，修复重要发现，重新全量测试。
- [ ] README/SQL/API 文档、测试记录、Git bundle 和源码备份。
- [ ] 上传 GitHub 私有仓库并验证远端提交；公网部署受主机配置限制，如实记录。
