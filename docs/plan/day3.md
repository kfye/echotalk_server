# Day 3 · 内容与媒体（后端）— 事项与方案

## 目标
App 能取到内容列表/详情与句级数据，媒体经 COS+CDN 播放；产出导入命令 + 1 条可播放样例。

## 现状
- `content` 模块只有 `model.go`(Video) + `doc.go`，**无 service/handler/router**（需整建）。
- Video 模型已就绪：`title/description/cover_url/hls_url/subtitle_en_url/subtitle_cn_url/duration/difficulty/category/is_free/status/sort/软删`，AutoMigrate 已含。
- **无 COS/上传代码**；`cmd/` 仅 `server`；OpenAPI 已定义 content/admin 接口与 schema（VideoListItem/VideoDetail/VideoInput）；统一响应/分页(`response.SuccessPage`)/错误码已就绪。

## ⚠️ 与原计划的三处对齐（先看）
1. **不建 `sentences` 表**（早先已定）：句级数据走**字幕文件**（`subtitle_*_url`），不是 DB 表。→ 详情接口返回字幕文件 URL，**App 拉取+解析**，后端不拼"句级数组"。
2. **付费锁 `locked` 依赖会员判断**，会员模块 **Day 5** 才做。→ Day 3 用一个 `MembershipChecker` 接口 + 临时桩（一律非会员）占位，Day 5 换真实现，门禁逻辑不动（适配器思路）。
3. **HLS 转码不自动化**：内测 1 条样例用 ffmpeg 手动转码上传，后端不做转码流水线。

---

## 任务分解

### A. content 模块分层（仿 user/payment）
- `repository.go`：`List`(分页+category/difficulty 过滤+仅上架)、`GetByID`、admin 的 `Create/Update/UpdateStatus/Delete`(软删)
- `service.go`：业务 + **付费门禁**（算 `locked`，locked 时隐藏 media URL）
- `handler.go` / `router.go`：App 接口 + admin 接口
- `errcode.go`：加 content 段 `ErrContentNotFound = New(12001, "内容不存在")`

### B. App 接口（对齐 OpenAPI）
- `GET /content/videos`：仅 `status=上架`；分页(page/page_size，page_size≤100)、category/difficulty 过滤；返回 `VideoListItem`(含 is_free + locked)；用 `response.SuccessPage`。
- `GET /content/videos/{id}`：返回 `VideoDetail`，**付费门禁**：
  - 免费或已解锁 → 返回 `hls_url` + 字幕 URL
  - 付费且未解锁 → `locked=true`，**不返回 hls_url/字幕 URL**（服务端隐藏，端上只显示锁）
- 列表/详情走**可选鉴权**（匿名也能浏览；匿名对付费内容即锁）。

### C. 字幕 / 句级数据
- 定义字幕格式（见"待确认1"）：推荐**单个 JSON**，元素 `{seq,start_ms,end_ms,text_en,text_cn,phonetic?}`。
- 存 COS → CDN，URL 落 `subtitle_*_url`；与 `training_records.sentence_index` 对齐。

### D. COS / 媒体（内测最小路径）
- DoD 关键是"1 条样例经 CDN 可播放"。最省事：**手动**把转码后的 HLS + 字幕传 COS（控制台/coscmd），导入命令只写 CDN URL 入库。
- 后端 COS 直传签名接口（对齐 OpenAPI `/system/upload/token`）→ **留到 Day 6 管理端上传时做**，Day 3 不强求。
- ⚠️ 用 COS 前先**轮换之前泄露的腾讯云密钥**，新密钥只进 `.env`。

### E. 导入命令 `cmd/import`
- 新增 `cmd/import/main.go`（独立二进制，复用 config + bootstrap.InitDB）。
- 输入：一个 JSON（视频元数据 + 字幕文件本地路径或已传 CDN url）。
- 行为：插入 Video 行；（可选）把本地字幕上传 COS 并回填 url。
- 运行：`go run ./cmd/import -config configs/config.yaml -file sample.json`。

### F. 录入样例 + 验证（内容源：已有 mp4）
- 用现有 mp4 手动转 HLS（单码率即可）：
  `ffmpeg -i lesson.mp4 -codec copy -hls_time 6 -hls_list_size 0 -f hls lesson.m3u8`
  → 得到 `lesson.m3u8 + lesson_*.ts`，整文件夹传 COS。
- 做 1 份字幕 JSON（句级时间戳）→ 传 COS。
- 跑导入命令把元数据 + `hls_url`(m3u8 的 CDN 地址) + 字幕 url 入库。
- 验证：列表/详情/CDN 播放全链路。

---

## 补充建议
- **列表只暴露上架**；草稿/下架仅 admin 列表可见（OpenAPI 已分 `/content` vs `/admin`）。
- **付费门禁集中在 service**，handler 只取 userID（可能匿名）。
- **admin 接口鉴权**：先复用现有 JWT 中间件；运营/普通用户的角色校验留 Day 6/7（先 TODO，不放开给普通用户）。
- **媒体防盗链**：CDN 防盗链/Referer 校验属上线前事项，内测可暂缓。

## 待确认（动手前定）
1. **字幕格式**：单个合并 JSON（en+cn+时间戳，**推荐**，最适合逐句跟读）vs 两个独立文件(en/cn)。选合并 JSON 则 `subtitle_cn_url` 字段空置或后续精简。
2. **COS 范围**：Day 3 只做"手动传 COS + 导入元数据"（**推荐**，保 DoD）vs 顺带做后端直传签名接口。
3. **locked 占位**：用 `MembershipChecker` 桩（一律非会员），Day 5 接真实会员 —— 可以吗？

## 完成标准（按实情微调）
- 导入命令入库 1 条带字幕的内容
- 列表标免费/付费；详情解锁时返回完整 media+字幕 URL，未解锁隐藏
- CDN 的 HLS 在浏览器/VLC 可播；字幕文件可拉取解析出句级数组

## 验证
- `go build/vet/gofmt`
- curl：list（分页/过滤/锁标识）+ detail（解锁 vs 锁）
- 导入命令跑通 + DB 校验
- 浏览器/VLC 播 CDN m3u8

## 备注（顺延项）
- 之前几轮（验证码 Redis 化、注册增强、登录防枚举、登出）已实现但**未提交**；Day 3 前可先 commit 到 `feat-v1`。
- 腾讯云泄露密钥**待轮换**。
