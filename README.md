# echotalk_server

EchoTalk「影子跟读」后台服务（Go + gin 模块化单体）。

## 技术栈

gin · GORM(MySQL 8) · Redis · viper · zap · validator · JWT(access+refresh)

## 目录结构

```
cmd/server          程序入口
internal/
  bootstrap         DB / Redis / 日志初始化
  config            viper 配置加载
  router            gin 引擎与路由装配
  middleware        鉴权 / 异常 / 日志 / CORS / 限流
  module/           业务模块（垂直分层：handler/service/repository/model/dto/router）
    user            账号体系（参考模块，已实现）
    content         内容管理（骨架）
    training        训练记录（骨架）
    payment         支付/会员 + channel 适配器（PaymentChannel / MockChannel）
    ops             运营配置（骨架）
  speech            语音能力网关（公共能力，业务模块禁止直连厂商 SDK）
    iflytek         科大讯飞 ISE Provider 实现
  pkg/              仓内共享：response / errcode / jwt
configs             配置文件
deploy              Dockerfile / docker-compose / nginx
scripts             迁移 / 备份脚本
api/openapi.yaml    三端对齐的接口契约
```

## 本地启动

```bash
go mod tidy
cp configs/config.example.yaml configs/config.yaml   # 按需修改
cp .env.example .env                                 # 填入密钥（不入仓）
go run ./cmd/server -config configs/config.yaml
curl http://localhost:8080/healthz
```

## 文档

- [docs/architecture.md](docs/architecture.md) — 架构流程图（请求流转 / 分层 / 依赖方向）
- [docs/docs-echotalk-md-gin-cheeky-yao.md](docs/docs-echotalk-md-gin-cheeky-yao.md) — 目录结构方案
- [docs/EchoTalk-技术方案.md](docs/EchoTalk-技术方案.md) — 完整技术方案