# Day 4 · 语音评测接入（后端）— 事项与方案

## 目标
App 上传一段录音（16K/16bit/单声道 WAV）+ 句子文本/句ID → 经 speech 网关送讯飞 ISE 评测 → 返回总分及发音/流利度/完整度/词级，并落 `training_records`；提供训练历史接口；讯飞异常时降级兜底不抛 500。

## 现状（已存在的骨架 — 不用重建）
- ✅ `speech.Gateway.Evaluate`：校验 WAV → 调 provider → **失败即降级**返回 `Degraded:true`（[gateway.go](internal/speech/gateway.go)）。
- ✅ `speech.Provider` 接口 + `EvaluateRequest`/`EvaluateResult`/`WordScore`（[provider.go](internal/speech/provider.go)）。
- ✅ `ValidateWAV` 16K/16bit/单声道校验（[audio.go](internal/speech/audio.go)）。
- ✅ errcode `15001 ErrSpeechUnavailable`、`15002 ErrAudioFormat`。
- ✅ `training.TrainingRecord` 模型（user/video/sentence_index/text/4项分/word_details JSON/degraded），AutoMigrate 已含（[model.go](internal/module/training/model.go)）。
- ✅ 网关已装配进 server + `router.Deps.Speech`，但**尚未使用**（[router.go:55-57](internal/router/router.go#L55) `_ = d.Speech`）。

## 缺口（Day 4 要补）
- ❌ 讯飞 `ISEProvider.Evaluate` 是**空桩**（返回空结果，[ise.go:25-28](internal/speech/iflytek/ise.go#L25)）—— 真实 HMAC 鉴权 + 音频上传 + XML 结果解析未做。
- ❌ `training` 模块只有 `model.go`+`doc.go`，**无 repository/service/handler/router/dto**。
- ❌ 无评测接口、无历史接口；training 路由未注册。
- ❌ 录音未存 COS（`AudioURL` 字段无人填）。
- ❌ OpenAPI 无 `/training/*` 接口与 schema。

## ⚠️ 与计划/约束的对齐（先看）
1. **讯飞 ISE 是唯一外部审批卡点**（Day1 即申请）。`iflytek.app_id/api_key/api_secret` 当前为空 → 真实密钥**只进 .env**（`config.yaml` 留空）。
2. **降级预案（推荐执行顺序据此定）**：网关已支持 `Degraded` 兜底。**先用空桩/降级结果把 B→C→E 链路跑通**，讯飞就绪后再落地 Task A 真实实现，主流程不被审批阻塞。
3. **音频规格**由 `gateway.ValidateWAV` 把关（16K/16bit/单声道 WAV）；端上 `record` 负责产出，后端只校验不转码。
4. **`sentence_index` 与字幕文件 `seq` 对齐**（Day3 已定）；记录冗余存 `sentence_text` 作评测试题文本，使记录自包含。
5. **付费门禁**：内测先只校验登录；"付费内容才可评测"的会员判断留 Day5（与 content 一致可走 `MembershipChecker`，本日不强求）。

---

## 任务分解（建议执行序：B → C → E → A → D → F）

### A. 讯飞 ISE Provider 真实实现（核心 · 外部依赖）
落地 [ise.go](internal/speech/iflytek/ise.go) 的 `Evaluate` 空桩：
- HMAC-SHA256 鉴权签名（按讯飞 ISE 云端 API 规范，host/date/digest）。
- 音频上传：讯飞 ISE WebSocket 流式（或 HTTP），题型 `read_sentence`、语种英文、试题文本=句子文本、音频 16K/16bit/单声道 WAV。
- 解析返回 XML → 填 `EvaluateResult`（overall/accuracy/fluency/integrity/words）。
- 调用加 `context` 超时；厂商错误码归一为 error（由网关转降级）。
- **依赖讯飞 ISE 开通 + 密钥进 .env**；未就绪则本任务延后，先用降级跑通其余。

### B. training 模块分层骨架（仿 content/payment）
- `repository.go`：`Create(record)`、`ListByUser(userID, page, size)`（倒序）、（可选 `GetByID`）。
- `dto.go`：`EvaluateInput`（multipart 绑定）、`EvaluateResponse`、`RecordItem`。
- `service.go`：评测编排（调 `speech.Gateway` → 落库）。
- `handler.go` / `router.go`：`RegisterRoutes(api, db, jwt, speechGW)`，在 [router.go](internal/router/router.go) 注册（替换 `_ = d.Speech` 占位）。

### C. 评测接口 `POST /api/v1/training/evaluate`（鉴权）
- `Auth` 中间件（必须登录，取 `user_id`）。
- 入参：multipart `audio`(WAV) + `video_id` + `sentence_index` + `text`(句子文本)。
- 流程：读音频 → `speechGW.Evaluate({audio,text,en_us})` → 落 `training_records`（含 user/video/sentence_index/text/四项分/word_details JSON/degraded）→ 返回结构化评分。
- 降级结果（`Degraded:true`）也落库并明确返回口径（见 F），不报 500。

### D. 录音存 COS（可选增强）
- 把上传的 WAV 传 COS（复用 [cos.Uploader](internal/pkg/cos/cos.go)）→ 回填 `audio_url`，供 Day10「原声 vs 录音对照回放」。
- 内测可先留空（DoD 不强求）；做则需 COS 密钥（**先轮换泄露密钥**，只进 .env）。建议**可选/后置**。

### E. 训练历史接口 `GET /api/v1/training/records`（鉴权 · 分页）
- 按 `user_id` 倒序分页（`response.SuccessPage`），返回 `RecordItem[]`。
- （可选）`GET /training/records/{id}` 单条详情（含 word_details）。

### F. 降级/错误处理 + OpenAPI 契约 + 验证
- 降级口径：handler 对 `Degraded` 给明确响应（如分数置 0 + `degraded:true` + 提示语），ISE 调用 context 超时；复用 errcode 15001/15002，按需补。
- OpenAPI：[api/openapi.yaml](api/openapi.yaml) 增 `POST /training/evaluate`(multipart) + `GET /training/records` + schema（EvaluateResult/TrainingRecordItem）。
- 验证：curl 传 16K WAV → 评分 JSON + 落库；history 列出该用户记录；模拟讯飞异常 → 降级兜底不 500。

---

## 完成标准（DoD，对齐冲刺计划）
- curl 上传一段 16K WAV → 返回结构化评分 JSON 并落 `training_records`。
- 训练历史接口能列出该用户记录。
- 讯飞异常/未就绪时接口有兜底返回（degraded），不抛 500。

## 验证
- `go build ./... && go vet ./... && gofmt -l`。
- curl：evaluate（带 token，multipart WAV）→ 评分；records（分页）→ 历史。
- 降级路径：未配讯飞密钥时走空桩/降级，链路仍通。

## 备注 / 顺延
- 讯飞 ISE 审批就绪前，A 可延后；B/C/E 先跑降级链路（与冲刺计划「降级预案 1」一致）。
- iflytek 密钥、（若做 D）COS 密钥只进 .env；泄露的腾讯云密钥**待轮换**。
- 之前 Day1–3 改动多未提交，仍在 `feat-v1`。