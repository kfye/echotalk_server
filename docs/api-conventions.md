# EchoTalk 后端接口契约（统一响应 / 错误码 / 异常）

> 全后端遵循。三端（App / 管理端 / 后端）以此对齐。代码落点见
> [response](../internal/pkg/response/response.go)、[errcode](../internal/pkg/errcode/errcode.go)、[middleware](../internal/middleware/)。

## 1. 统一响应体

所有接口返回同一结构：

```json
{ "code": 0, "msg": "ok", "data": { }, "request_id": "a1b2c3d4" }
```

| 字段 | 说明 |
|---|---|
| `code` | 业务码，`0` 表示成功，非 0 见错误码表 |
| `msg` | 提示信息（成功为 `ok`，失败为错误提示，可含字段级校验详情） |
| `data` | 业务数据，失败时省略 |
| `request_id` | 本次请求唯一 ID，同时在响应头 `X-Request-ID`；排查问题时提供给后端 |

列表接口的 `data` 统一用分页结构：

```json
{ "code": 0, "msg": "ok",
  "data": { "list": [ ], "total": 120, "page": 1, "page_size": 20 } }
```

## 2. HTTP 状态码规则

- **框架/传输级错误 → 语义 HTTP 状态码**：客户端/网关/监控据此判断。
- **纯业务拒绝 → HTTP 200 + 业务码**：当作正常响应渲染提示。

| 场景 | HTTP | 示例 body |
|---|---|---|
| 成功 | 200 | `{"code":0,"msg":"ok",...}` |
| 参数错误 | 400 | `{"code":10400,"msg":"email 必须是有效邮箱地址"}` |
| 未授权/登录失效 | 401 | `{"code":10401,"msg":"未授权或登录失效"}` |
| 无权限 | 403 | `{"code":10403,"msg":"无访问权限"}` |
| 资源不存在 | 404 | `{"code":10404,"msg":"资源不存在"}` |
| 限流 | 429 | `{"code":10429,"msg":"请求过于频繁"}` |
| 服务器异常/panic | 500 | `{"code":10500,"msg":"服务器内部错误"}` |
| 业务拒绝（如密码错误） | 200 | `{"code":11003,"msg":"密码错误"}` |

## 3. 错误码分段

按模块千号段划分，每模块预留 1000 号：

| 段 | 模块 | 已定义（示例） |
|---|---|---|
| `10xxx` | 通用/框架 | 10400 参数错误、10401 未授权、10403 无权限、10404 不存在、10429 限流、10500 服务器错误 |
| `11xxx` | user | 11001 用户不存在、11002 用户已存在、11003 密码错误、11004 令牌无效 |
| `12xxx` | content | （待定义） |
| `13xxx` | training | （待定义） |
| `14xxx` | payment | 14001 订单不存在、14002 支付失败 |
| `15xxx` | speech | 15001 语音不可用、15002 音频格式不符 |
| `16xxx` | ops | （待定义） |

## 4. 开发约定（写新接口照此）

- 定义错误：业务类用 `errcode.New(code, msg)`（默认 HTTP 200）；框架类用 `errcode.NewHTTP(code, http, msg)`。
- handler 出错统一 `response.Error(c, err)`：`*errcode.Error` 按其 HTTP 状态返回，未知 error 归一为 500 且不暴露内部细节。
- 参数校验：`response.Error(c, errcode.ErrParam.WithMsg(validatorx.Message(err)))`。
- 成功：`response.Success(c, data)`；列表：`response.SuccessPage(c, list, total, page, pageSize)`。
- 禁止 handler 直接 `c.JSON` 拼裸响应，必须经 `response` 包，保证结构统一。

## 5. 全局异常

[middleware.Recovery](../internal/middleware/recovery.go) 捕获所有 panic：记 zap 堆栈日志（带 request_id）后返回 `10500 / HTTP 500`，绝不裸抛或泄漏堆栈给客户端。
