# CLAUDE.md — echotalk_server（后台 / Go）

> EchoTalk「影子跟读」项目的**后台服务**（轻量多仓之一）。完整技术方案与 14 天内测计划见 Claude Project 知识库。前后端接口以一份 **OpenAPI 契约**对齐。

## 项目背景

EchoTalk 是面向全年龄段英语学习者的移动端 App，核心路径「跟读—发音—交流」。本仓库是三端唯一的数据与鉴权来源。

- **内测范围**：影子跟读 + 账号体系 + 支付与会员（**模拟支付**），仅 Android。
- **阶段**：单人开发，纯开发测试，无真实用户。
- **同项目其它仓**：`echotalk_app`（Flutter 客户端）、`echotalk_manage`（Vue 管理端）。

## 本仓库角色

Golang **模块化单体**后台，承载账号、内容、训练、会员、运营等业务，以及统一的语音能力网关。模块边界：`user` / `content` / `training` / `payment` / `ops` / `speech`（公共能力），不过早微服务。

## 技术栈

- **语言/框架**：Go + gin（清晰分层）。
- **数据**：MySQL 8、Redis。
- **配套**：viper（配置）、GORM（ORM/迁移）、zap（日志）、validator（校验）。
- **媒体**：腾讯云 COS + CDN，视频 HLS。
- **语音**：科大讯飞 **ISE 云端 API**，封装在 `speech` 网关。

## 关键约束

- **语音评测一律经 `speech` 网关调讯飞 ISE 云端 API**，禁止业务模块直接耦合 SDK；音频 **16K/16bit/单声道 WAV**；讯飞异常要兜底降级，不抛 500。
- **支付适配器模式**：定义 `PaymentChannel` 接口，内测实现 `MockChannel`；日后接微信/支付宝/IAP 只新增实现，**会员逻辑不动**。
- **付费边界服务端校验**：免费/付费解锁判断必须在后端，客户端只做展示。
- 统一响应结构 + 错误码 + 全局异常中间件；鉴权用 JWT（access + refresh）。
- **密钥不进仓库**：DB 密码、讯飞 key、COS secret 走环境变量 / `.env`（`.env` 入 `.gitignore`）。

## 部署与运维预期

- **现状**：单台腾讯云 **2 核 2G**（偏小）+ Portainer + 赠送 50G COS；纯开发测试。
- **容器**：Nginx（反代/HTTPS）+ Go 应用 + MySQL + Redis，Portainer 管理；管理端静态文件由 Nginx 托管。
- **小机器优化（防 OOM）**：加 Swap；MySQL 调瘦（buffer pool 128–256M、关 performance_schema、调小 max_connections）；Redis maxmemory ~128M + 淘汰策略；每容器设内存上限。
- **安全**：MySQL/Redis 绝不对公网开放，安全组仅放 80/443。
- **备份**：定时 `mysqldump → COS`。
- **媒体**：COS + CDN，视频不从服务器直出；COS 不可当 MySQL 数据盘。
- **备案**：内测仅 Android，可用 network security config 暂缓，正式上线前办理。
- **升级触发点**：转「有真实用户」时再升配 / DB 拆第二台轻量服务器 / 迁云数据库。

## 待确认
语音厂商最终选型（讯飞 vs 腾讯，POC 后定）；内容来源（自制/采购）。
