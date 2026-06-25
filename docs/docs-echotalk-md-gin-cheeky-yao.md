# EchoTalk 后台 — gin 模块化单体目录结构

## Context

`echotalk_server` 目前是空仓（仅 `CLAUDE.md`、`docs/`、`README.md`，无任何 Go 代码、无 `go.mod`）。
需要按技术方案（`docs/EchoTalk-技术方案.md` §3.2 / §7）与 `CLAUDE.md` 的关键约束，生成一套 **gin 模块化单体**骨架，作为后续业务开发的地基。

目标形态（已与用户确认）：
- **垂直按模块切分**：每个业务模块自带 `handler/service/repository/model/dto/router`，模块边界即数据表边界，为日后拆微服务留空间。
- **目录 + 关键起步文件**：搭出可 `go build` 通过的骨架——基础设施（配置/日志/DB/Redis/路由/中间件/统一响应/错误码/JWT）全部落地，`user` 作为参考模块写全，其余业务模块先建目录骨架。

须落实的 `CLAUDE.md` 硬约束：
- 语音一律经 `speech` 网关调讯飞 ISE，业务模块禁止直连 SDK；音频 16K/16bit/单声道 WAV；讯飞异常兜底降级不抛 500。
- 支付适配器模式：`PaymentChannel` 接口 + 内测 `MockChannel`，会员逻辑与渠道解耦。
- 付费边界服务端校验；统一响应 + 错误码 + 全局异常中间件；JWT（access + refresh）。
- 密钥不进仓库，走环境变量 / `.env`（`.env` 入 `.gitignore`）。

---

## 目录结构

```
echotalk_server/
├── cmd/
│   └── server/
│       └── main.go                 # 入口：load config → init logger/db/redis → 装配 router → 启动 gin
├── api/
│   └── openapi.yaml                # OpenAPI 契约（三端对齐，占位）
├── configs/
│   ├── config.yaml                 # 默认配置（非密钥：端口、日志级别、池大小…）
│   └── config.example.yaml         # 示例
├── internal/
│   ├── bootstrap/                  # 依赖初始化与装配
│   │   ├── db.go                   # GORM + MySQL（连接池/瘦身参数）
│   │   ├── redis.go                # go-redis
│   │   └── logger.go               # zap 初始化
│   ├── config/
│   │   └── config.go               # viper 加载 + 配置结构体（含 env 覆盖）
│   ├── router/
│   │   └── router.go               # gin 引擎 + 全局中间件 + 按模块注册路由
│   ├── middleware/
│   │   ├── recovery.go             # 全局异常 → 统一错误响应
│   │   ├── auth.go                 # JWT 鉴权（access）
│   │   ├── cors.go
│   │   ├── ratelimit.go            # 限流（语音/AI 调用保护预留）
│   │   └── access_log.go           # 请求日志（zap）
│   ├── module/                     # 应用服务层（业务模块，垂直切分）
│   │   ├── user/                   # ★ 参考模块，写全
│   │   │   ├── router.go           # RegisterRoutes(rg *gin.RouterGroup, ...)
│   │   │   ├── handler.go          # gin handler，调 service
│   │   │   ├── service.go          # 业务逻辑
│   │   │   ├── repository.go       # GORM 数据访问
│   │   │   ├── model.go            # GORM 实体（独立表边界）
│   │   │   └── dto.go              # 请求/响应结构 + validator tag
│   │   ├── content/                # 视频/文章/音标、句级数据、上下架（目录骨架）
│   │   ├── training/               # 跟读/发音/文章成绩、收藏、生词本、打卡（目录骨架）
│   │   ├── payment/                # 订单、会员、双通道、对账
│   │   │   ├── router.go / handler.go / service.go / repository.go / model.go / dto.go
│   │   │   └── channel/            # 支付适配器
│   │   │       ├── channel.go      # PaymentChannel 接口
│   │   │       └── mock.go         # MockChannel（内测实现）
│   │   └── ops/                    # 运营配置：分类标签/价格/功能开关/审核（目录骨架）
│   ├── speech/                     # ★ 公共能力层：语音能力网关（独立于 module）
│   │   ├── gateway.go              # 对上统一接口：Evaluate/ASR/TTS + 限流/缓存/降级
│   │   ├── provider.go             # Provider 接口（评测/识别/合成），多厂商抽象
│   │   ├── audio.go                # 16K/16bit/单声道 WAV 校验辅助
│   │   └── iflytek/
│   │       └── ise.go              # 讯飞 ISE 云端 API 实现（Provider 实现）
│   └── pkg/                        # 仓内共享（非对外）
│       ├── response/response.go    # 统一响应结构 Success/Fail
│       ├── errcode/errcode.go      # 错误码定义
│       └── jwt/jwt.go              # access + refresh 签发/校验
├── migrations/                     # SQL / GORM AutoMigrate 脚本（占位）
├── deploy/
│   ├── Dockerfile
│   ├── docker-compose.yml          # Nginx + Go + MySQL + Redis（含每容器内存上限注释）
│   └── nginx/echotalk.conf         # 反代 + HTTPS 占位
├── scripts/
│   ├── migrate.sh
│   └── backup.sh                   # mysqldump → COS（占位）
├── .env.example                    # DB 密码 / 讯飞 key / COS secret 占位
├── go.mod
├── go.sum
└── README.md（已存在，补充启动说明）
```

### 分层与依赖方向
- 调用方向：`handler → service → repository → DB`；`service → speech.Gateway`（语音）、`service → payment/channel.PaymentChannel`（支付）。
- 业务模块**不得** import `speech/iflytek` 或任何讯飞 SDK，只依赖 `speech.Gateway` 接口。
- `internal/pkg/*` 为无业务依赖的纯工具，可被任意层引用。

---

## 关键起步文件要点

| 文件 | 内容 |
|---|---|
| `cmd/server/main.go` | 串起 config→logger→db→redis→router→`engine.Run`，优雅退出 |
| `internal/config/config.go` | viper 读 `configs/config.yaml` + env 覆盖；结构体含 `Server/MySQL/Redis/JWT/Iflytek/COS` |
| `internal/pkg/response/response.go` | `Success(c, data)` / `Fail(c, code, msg)`，固定 `{code,msg,data}` 结构 |
| `internal/pkg/errcode/errcode.go` | 错误码常量分段（通用/user/content/payment/speech…） |
| `internal/pkg/jwt/jwt.go` | `GeneratePair`（access+refresh）/`ParseAccess`/`Refresh` |
| `internal/middleware/recovery.go` | recover panic → `errcode` + 统一响应，绝不裸 500 |
| `internal/module/payment/channel/channel.go` | `type PaymentChannel interface { Pay/Query/Refund/VerifyCallback }` |
| `internal/module/payment/channel/mock.go` | `MockChannel` 直接返回成功，供内测模拟支付 |
| `internal/speech/provider.go` | `Provider` 接口；`gateway.go` 内做降级（讯飞失败返回兜底结果而非 error 透传） |
| `internal/module/user/*` | 全套参考实现（注册/登录/refresh 接口走通编译） |

### 建议依赖（go.mod）
`gin` · `gorm.io/gorm` + `gorm.io/driver/mysql` · `redis/go-redis/v9` · `spf13/viper` · `uber-go/zap` · `go-playground/validator/v10` · `golang-jwt/jwt/v5`。

### 配置/安全
- `.env` 已在现有 `.gitignore` 中需确认覆盖；新增 `.env.example` 列出所有密钥占位。
- `configs/config.yaml` 只放非敏感默认值，敏感项用 `${ENV_VAR}` 由 viper 从环境注入。

---

## 验证方式

1. `go mod tidy` 拉齐依赖，`go build ./...` 全量编译通过（骨架可编译是本次交付的硬指标）。
2. `go vet ./...` 无异常。
3. `cp configs/config.example.yaml configs/config.yaml && cp .env.example .env`，填本地 MySQL/Redis 后 `go run ./cmd/server`，访问健康检查路由（如 `GET /healthz`）返回统一响应结构。
4. 确认 `user` 模块的注册/登录路由可命中（即便 service 为最小实现），验证分层与中间件装配链路通畅。
5. 人工核对：`grep -r iflytek internal/module` 应为空 —— 业务模块未直连语音 SDK。

> 说明：本阶段只交付可编译的结构骨架与参考模块，具体业务逻辑、表结构、OpenAPI 契约内容在各模块开发前再细化。
