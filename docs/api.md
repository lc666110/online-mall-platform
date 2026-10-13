# REST API v1

基础路径 `/api/v1`。请求和响应均为 JSON；认证 `Authorization: Bearer <token>`。成功响应 `{"data": ...}`，错误响应 `{"error": "可读信息"}`。列表返回 `data: {items:[],total,page,page_size}`，默认 page=1、page_size=12，最大 100。400 输入/状态无效，401 未登录/失效，403 无权限，404 不存在或非本人资源，409 冲突，429 限流，503 依赖暂不可用。

`GET /health` 位于基础路径之外，MySQL 和 Redis 健康返回 200 `{"status":"ok"}`，异常 503。

## 公开接口

| 方法 | 路径                    | 参数/返回                                                               |
| ---- | ----------------------- | ----------------------------------------------------------------------- |
| POST | `/auth/register`        | `{email,password,name}` -> User，201；密码 10–72 字符且 UTF-8 <=72 字节 |
| POST | `/auth/login`           | `{email,password}` -> `{token,user}`，24 小时会话                       |
| GET  | `/categories`           | Category[]                                                              |
| GET  | `/products`             | `q,category_id,page,page_size` -> Page<Product>；仅上架                 |
| GET  | `/products/:id`         | Product，含 category、variants                                          |
| GET  | `/products/:id/reviews` | `page,page_size` -> Page<Review>                                        |

注册和登录合计按客户端 IP 30 次/分钟。密码、缓存版本字段永不返回。Product.images 是存储 JSON 数组的字符串，例如 `["/images/product-1.svg"]`；variant.price 单位分，stock 为整数。

## 登录用户接口

| 方法   | 路径                      | 参数/返回                                                                   |
| ------ | ------------------------- | --------------------------------------------------------------------------- |
| POST   | `/auth/logout`            | 无请求体 -> true；当前会话立即失效                                          |
| GET    | `/me`                     | User                                                                        |
| PATCH  | `/me`                     | `{name,phone,avatar}` -> User；头像为 HTTPS 地址或空                        |
| GET    | `/cart`                   | CartItem[]，含实时 Variant                                                  |
| POST   | `/cart`                   | `{variant_id,quantity}` -> true；追加数量，最终数量 1–99 且不超库存         |
| PATCH  | `/cart/:id`               | `{quantity}` -> true；设置绝对数量                                          |
| DELETE | `/cart/:id`               | true                                                                        |
| POST   | `/orders`                 | 请求头 Idempotency-Key 8–80 字符；`{recipient,phone,address}` -> Order，201 |
| GET    | `/orders`                 | `page,page_size` -> Page<Order>                                             |
| GET    | `/orders/:id`             | Order，含 items                                                             |
| PATCH  | `/orders/:id/status`      | `{status}` -> Order；pending→cancelled / shipped→completed                  |
| POST   | `/reviews`                | `{order_item_id,rating,content}` -> Review，201；rating 1–5，content <=2000 |
| POST   | `/orders/:id/after-sales` | `{reason}` -> AfterSale，201；reason <=1000                                 |
| GET    | `/after-sales`            | `page,page_size` -> Page<AfterSale>                                         |

下单时结算整个购物车；前端无需发送金额或商品价格。数量/库存/上架状态均重新校验。同一用户重复 key 返回已生成订单；新订单请使用新 key。

## 管理员/员工接口

需要角色 admin 或 staff，另标明 admin-only 的接口仅管理员可用。

| 方法   | 路径                         | 参数/返回                                                                             |
| ------ | ---------------------------- | ------------------------------------------------------------------------------------- |
| GET    | `/admin/dashboard`           | Dashboard：sales 分、orders/users/products 总数，30 天 trend/growth，销量前10 ranking |
| GET    | `/admin/products`            | 与商品列表参数相同；包含下架商品                                                      |
| POST   | `/admin/products`            | ProductInput -> Product                                                               |
| PUT    | `/admin/products/:id`        | ProductInput -> Product；保留未包含的已有 SKU，不删除历史规格                         |
| PATCH  | `/admin/products/:id/status` | `{active:boolean}` -> true；独立上下架，不写库存                                      |
| DELETE | `/admin/products/:id`        | true；有订单/购物车引用返回409，应下架                                                |
| GET    | `/admin/categories`          | Category[]                                                                            |
| POST   | `/admin/categories`          | `{name}` -> Category                                                                  |
| PUT    | `/admin/categories/:id`      | `{name}` -> Category                                                                  |
| DELETE | `/admin/categories/:id`      | true；分类仍有商品409                                                                 |
| GET    | `/admin/orders`              | Page<Order>                                                                           |
| GET    | `/admin/orders/:id`          | Order                                                                                 |
| PATCH  | `/admin/orders/:id/status`   | `{status,carrier,tracking}` -> Order；发货需要后两字段                                |
| GET    | `/admin/after-sales`         | Page<AfterSale>                                                                       |
| PATCH  | `/admin/after-sales/:id`     | `{status:"approved"                                                                   | "rejected",note}` -> true；仅pending |
| GET    | `/admin/users`               | Page<User>，admin-only                                                                |
| PATCH  | `/admin/users/:id`           | `{name,phone,role:"user"                                                              | "staff"                              | "admin",enabled}` -> User，admin-only |

ProductInput 示例：

```json
{
  "category_id": 1,
  "name": "无线耳机",
  "description": "商品详情",
  "images": ["/images/product-1.svg"],
  "active": true,
  "variants": [
    {
      "id": 0,
      "sku": "EAR-A",
      "specification": "黑色",
      "price": 29900,
      "stock": 30
    }
  ]
}
```

新规格 id=0，价格1–100000000分，库存0–1000000，规格最多50个、图片最多10张。图片允许 HTTPS 或站内 `/images/` URL。编辑已有规格必须传 `expected_stock` 为读取表单时的库存，`stock` 为希望设置的库存。stock 与 expected_stock 相同表示不调整库存；不同则做并发检查，库存已变化返回409且整个事务回滚。

订单状态转换及退款边界见 [architecture.md](architecture.md)。统计 sales 为已确认/发货/完成订单额，排除取消/售后完成；代表订单金额，不是支付平台实收。增长按注册日期 UTC 统计，商品销量按当前有效订单明细累计。

## 调用示例

```sh
# TOKEN 为登录响应 token，请不要提交到版本库
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/cart
curl -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: unique-order-key-2026' \
  -d '{"recipient":"张三","phone":"13800138000","address":"测试地址"}' \
  http://127.0.0.1:8080/api/v1/orders
```
