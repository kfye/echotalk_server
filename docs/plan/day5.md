# Day 5 · 会员与支付（后端 · 模拟通道）— 任务分解

## Context
Day 5 目标（M1 里程碑）：**下单 → 模拟支付 → 开会员 → 付费内容放行** 全链路在服务端跑通，且付费边界由后端裁决（非仅前端隐藏）。
两条贯穿约束：①支付适配器模式（`PaymentChannel` + `MockChannel`，日后接微信/支付宝/IAP 不动会员逻辑）；②付费边界服务端校验。

### 现状盘点（已有骨架 vs 待建）
**已就绪**
- 模型已建并已登记 AutoMigrate：`payment.Product`（价格/时长/类型/上下架/排序）、`payment.Order`（状态机 0待支付/1已支付/2退款/3关闭）、`payment.Membership`（一用户一条，status/start_at/expire_at/source/last_order_id）。见 [model.go](internal/module/payment/model.go) / [product.go](internal/module/payment/product.go) / [membership.go](internal/module/payment/membership.go)。
- 适配器：[channel.go](internal/module/payment/channel/channel.go) 定义 `PaymentChannel`（Name/Pay/Query/Refund/VerifyCallback）；[mock.go](internal/module/payment/channel/mock.go) 实现 `MockChannel`（恒成功）。已在 main.go 注入。
- content 付费门禁已接好接口：[service.go:26-39](internal/module/content/service.go#L26) `locked()` 调 `MembershipChecker.IsActiveMember`，List/Detail 已用；Detail 锁定时隐藏媒体地址（[dto.go toDetail](internal/module/content/dto.go#L68)）。
- 错误码：`ErrOrderNotFound=14001`、`ErrPayFailed=14002`。

**待建（Day 5 真正的活）**
- `CreateOrder` 价格写死 `Amount:100`（[service.go:29](internal/module/payment/service.go#L29)），未从 Product 读价/时长、未校验上架；且**调完 Pay 后没落库支付结果、没置订单状态、没写会员**——目前下单即调 Pay 但只在内存改 TradeNo 就返回。
- Repository 仅 `Create(order)`，缺 Product 读取、Order 查询/更新、Membership 读写。
- 仅 `POST /payment/orders` 一个端点；缺 SKU 列表、支付确认、会员状态查询、手动发卡。
- content 仍用 `noMembershipChecker` 桩（一律非会员，[main.go:56](cmd/server/main.go#L56)），**缺真实 checker**（查 memberships 表）。
- 无 SKU 数据；无 OpenAPI `/payment/*` 契约。

---

## 任务分解（A–F，可逐个确认执行）

### Task A — SKU 读取与播种
- Product 仓储：`GetByID`、`ListOnline`（status=上架、按 sort）。
- App 端 `GET /api/v1/payment/products`（付费墙取 SKU 列表）。
- 播种 1 个 SKU 供联调（扩展 `cmd/import` 或一条 SQL；管理端增删改留 Day 7）。
- 新增错误码：`ErrProductNotFound`、`ErrProductOffline`（14003/14004）。

### Task B — 下单：真实价格 + 待支付订单
- `CreateOrder` 改为：按 `product_id` 读 Product → 校验存在且上架 → `Amount=Product.Price`、记下 `DurationDays` 语义 → 落库 `OrderStatusPending` → 返回 `order_no`（+ `MockChannel.Pay` 的 pay_payload 占位，供客户端"唤起支付"）。**下单阶段不开会员**（开会员在确认步）。
- Repository 增：`GetOrderByNo`、`UpdateOrder`。

### Task C — 模拟支付确认 + 会员开通/续期
- `POST /api/v1/payment/orders/:order_no/confirm`（模拟"支付回调/确认"）：
  - 查订单 → 校验属当前用户、状态为 pending（已支付则幂等返回，不重复开会员）。
  - 经 `channel.Query`/`VerifyCallback` 确认成功 → 订单置 `OrderStatusPaid` + `PaidAt` + `TradeNo`。
  - **会员开通/续期**：查该用户 membership；有效未过期则 `ExpireAt += DurationDays`（续费顺延），否则 `StartAt=now`、`ExpireAt=now+DurationDays`；置 `Status=Active`、`Source=Order`、`LastOrderID`。
- Membership 仓储：`GetByUserID`、`Upsert`（事务内与订单更新一起，保证一致）。å

### Task D — 真实 MembershipChecker + 接入 content 门禁
- 在 payment 内实现满足 `content.MembershipChecker` 的真实 checker：`IsActiveMember = 存在 && Status=Active && ExpireAt>now`（结构化鸭子类型，payment 暴露、content 消费，不互相 import）。
- main.go 用它替换 `content.NewNoMembershipChecker()`（[main.go:56](cmd/server/main.go#L56)）。
- 会员状态查询 `GET /api/v1/payment/membership`（个人中心展示有效期/状态）。

### Task E — 手动发卡（运营）
- `POST /api/v1/admin/memberships/grant`（给指定 user_id 开/续会员，`Source=Manual`），复用 content 既有 `/admin/*` 约定：**仅校验登录、角色校验留 Day 6/7**（与 [content/router.go:25-27](internal/module/content/router.go#L26) 一致）。
- 复用 Task C 的开通/续期逻辑（抽成 service 内共用函数）。

### Task F — 付费门禁口径 + OpenAPI 契约 + 端到端验证
- 确认付费详情口径（见下方决策）。
- OpenAPI 补 `/payment/products`、`/payment/orders`、`/payment/orders/{order_no}/confirm`、`/payment/membership`、`/admin/memberships/grant` + schemas（Product/Order/Membership）。
- 端到端 curl（DoD = M1）：
  1. 注册登录拿 token；`GET /payment/products` 看到 SKU。
  2. 非会员访问付费内容 → content Detail `locked=true`、媒体地址隐藏。
  3. `POST /payment/orders` 下单（pending）→ `POST .../confirm` 确认 → 订单 paid、membership active。
  4. 再访问付费内容 → `locked=false`、媒体地址放出。
  5. 续费：再买一单确认 → `expire_at` 顺延。
  6. 手动发卡给另一 user → 该用户直接变会员、付费解锁。
  7. DB 核对 orders/memberships 落库正确。

---

## 已确认的设计决策
1. **订单流程**：✅ 拆分 —— `POST /orders` 建 pending 订单；独立 `POST /orders/:no/confirm` 模拟支付回调，确认成功才开会员。
2. **付费详情口径**：✅ 沿用现有 locked 标志（HTTP 200 + `locked=true` + 隐藏媒体地址），不加硬 403。content 现有逻辑不动，只换真实 checker。

## 不在 Day 5 范围
- 真实微信/支付宝/IAP（适配器留口，Day 内测后）。
- 运营角色权限体系、SKU 管理 UI（Day 6/7 管理端）。
- 退款/对账完整流程（Refund 接口已留桩，内测不做闭环）。