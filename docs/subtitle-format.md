# 字幕 / 逐句跟读脚本格式（跨仓契约）

> EchoTalk「影子跟读」的句级数据格式。字幕文件经 **COS → CDN** 托管，URL 落在 `videos.subtitle_en_url`；**App 拉取该 URL 自行解析**，后端不在接口里拼装句级数组。
>
> 本格式是 **App / 后端 / 管理端三方共用的契约**，与 OpenAPI 接口契约同等地位。改字段先改本文件并升 `version`。

## 设计背景（为什么是这个格式）

影子跟读是**逐句**进行的：`播放某句原声 → 用户跟读 → 录音送讯飞 ISE 评分 → 看分 → 下一句`。
所以 App 需要把一段视频拆成"带时间戳的句子数组"，每句提供：**播哪一段、屏幕显示什么、拿什么文本去评分**。
句级数据是**静态只读内容**，因此不进 MySQL（不设 `sentences` 表），而是做成一个 JSON 文件托管在 COS，App 一次拉取本地使用。

## 范例

见 [docs/samples/subtitle.sample.json](samples/subtitle.sample.json)：

```json
{
  "version": 1,
  "language": "en",
  "sentences": [
    { "seq": 0, "start_ms": 0,    "end_ms": 2600, "text_en": "Hello and welcome.",        "text_cn": "你好，欢迎。",       "phonetic": "" },
    { "seq": 1, "start_ms": 2600, "end_ms": 6200, "text_en": "Today we learn shadowing.", "text_cn": "今天我们学影子跟读。" },
    { "seq": 2, "start_ms": 6200, "end_ms": 9800, "text_en": "Repeat after me, please.",  "text_cn": "请跟我读。" }
  ]
}
```

## 字段定义

### 顶层

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `version` | int | 是 | 格式版本号，当前为 `1`。字段演进时递增，App 据此选解析规则 |
| `language` | string | 是 | 主语种，如 `en`。标明 `text_en` 的语言，为多语种留口子 |
| `sentences` | array | 是 | 句子数组，按播放顺序排列；元素见下 |

### `sentences[]` 元素

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `seq` | int | 是 | 句子序号，**从 0 开始、连续递增**。句子的身份号（见下方对齐说明） |
| `start_ms` | int | 是 | 该句在视频中的起始毫秒。App 据此定位/播放该句片段 |
| `end_ms` | int | 是 | 该句在视频中的结束毫秒（`end_ms > start_ms`） |
| `text_en` | string | 是 | 英文原文。用于显示、跟读对照，**并作为讯飞 ISE 评测的标准答案文本** |
| `text_cn` | string | 合并 JSON 时必填 | 中文翻译，纯展示。en/cn 合并在同一文件时必填 |
| `phonetic` | string | 否 | 音标/发音提示。无则留空或省略 |

## 合并 JSON 约定（en + cn 同文件）

内测采用**单个合并 JSON**：`text_en` 与 `text_cn` 放在同一文件的同一句里（逐句跟读时一句的英/中/时间一起用，最省事）。因此：

- `videos.subtitle_en_url` → 指向这个合并 JSON 文件。
- `videos.subtitle_cn_url` → **空置**（中文已在合并文件内，不再单独出一个中文文件）。

> 若日后改为 en/cn 拆分两个文件，再启用 `subtitle_cn_url` 并升 `version`。

## ⚠️ 跨仓对齐：`seq` === `training_records.sentence_index`

字幕里的句子序号字段叫 `seq`，跟读评测记录表 `training_records` 里叫 `sentence_index`——**两者是同一个数，指向同一句**：

```
字幕文件                训练记录表                     讯飞评测
sentences[2].seq = 2 → training_records.sentence_index = 2 → 用 sentences[2].text_en 作为评测文本
```

- 用户跟读第 `N` 句，录音送评分，后端把结果写入 `training_records`，`sentence_index = N`（即字幕的 `seq`）。
- 反查"用户某句读得如何"：用 `video_id + sentence_index` 定位记录，再回字幕文件取同 `seq` 的 `text_en`。

**三方必须共识**：编号一律 **从 0 开始、连续**。任一端从 1 开始或跳号，都会导致跟读记录与句子错位。对应模型：[internal/module/training/model.go](../internal/module/training/model.go)（`SentenceIndex` 字段）。

## 托管与消费

- 文件传 COS → CDN，URL 入库 `videos.subtitle_en_url`（详情接口在内容**已解锁**时返回该 URL，未解锁隐藏）。
- App 拿到 URL 后自行拉取、解析 `sentences[]` 驱动逐句跟读。后端不解析、不拼装句级数组。
