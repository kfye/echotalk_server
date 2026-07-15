// Package iflytek 科大讯飞 ISE 云端 API 的 Provider 实现。
package iflytek

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/echotalk/echotalk_server/internal/config"
	"github.com/echotalk/echotalk_server/internal/speech"
)

// 音频分帧参数：16K/16bit/单声道 → 1280 字节 = 40ms，一帧一发按实时节奏。
const (
	wavHeaderSize = 44
	audioChunk    = 1280
	frameInterval = 40 * time.Millisecond

	audioStatusFirst  = 1 // aus 首帧
	audioStatusMiddle = 2 // aus 中间帧
	audioStatusLast   = 4 // aus 末帧

	dataStatusBegin = 0 // data.status 参数帧
	dataStatusMid   = 1 // data.status 音频中
	dataStatusEnd   = 2 // data.status 结束/结果帧

	// textBOM 讯飞要求朗读文本前置 UTF-8 BOM()。
	textBOM = "\uFEFF"
)

// ISEProvider 实现 speech.Provider，对接讯飞 ISE 云端评测 API（WebSocket 流式）。
type ISEProvider struct {
	cfg config.IflytekConfig
}

// NewISEProvider 创建讯飞 ISE Provider。
func NewISEProvider(cfg config.IflytekConfig) *ISEProvider {
	return &ISEProvider{cfg: cfg}
}

// 编译期断言：确保实现了 speech.Provider 接口。
var _ speech.Provider = (*ISEProvider)(nil)

// iseRequest 讯飞 ISE 请求帧。
type iseRequest struct {
	Common   *iseCommon   `json:"common,omitempty"`
	Business *iseBusiness `json:"business,omitempty"`
	Data     iseData      `json:"data"`
}

type iseCommon struct {
	AppID string `json:"app_id"`
}

type iseBusiness struct {
	Category string `json:"category,omitempty"` // 题型 read_sentence
	Rstcd    string `json:"rstcd,omitempty"`    // 结果编码 utf8
	Sub      string `json:"sub,omitempty"`      // 服务 ise
	Group    string `json:"group,omitempty"`    // 群体 pupil/adult
	Ent      string `json:"ent,omitempty"`      // 语种 en_vip/cn_vip
	Tte      string `json:"tte,omitempty"`      // 文本编码 utf-8
	Cmd      string `json:"cmd,omitempty"`      // ssb 参数帧 / auw 音频帧
	Auf      string `json:"auf,omitempty"`      // 音频采样 audio/L16;rate=16000
	Aue      string `json:"aue,omitempty"`      // 音频编码 raw
	Text     string `json:"text,omitempty"`     // 朗读文本(带 BOM)
	Aus      int    `json:"aus,omitempty"`      // 音频帧标识 1首/2中/4末
	TtpSkip  bool   `json:"ttp_skip,omitempty"` // 跳过文本合成
}

type iseData struct {
	Status   int    `json:"status"`
	Data     string `json:"data,omitempty"`      // base64 PCM
	DataType int    `json:"data_type,omitempty"` // 1=音频
	Encoding string `json:"encoding,omitempty"`  // raw
}

// iseResponse 讯飞 ISE 响应帧。
type iseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
	Data    struct {
		Status int    `json:"status"`
		Data   string `json:"data"` // base64 XML(结束帧)
	} `json:"data"`
}

// Evaluate 调用讯飞 ISE 评测。任一异常返回 error，由网关判定降级（不抛 500）。
func (p *ISEProvider) Evaluate(ctx context.Context, req speech.EvaluateRequest) (speech.EvaluateResult, error) {
	if p.cfg.AppID == "" || p.cfg.APIKey == "" || p.cfg.APISecret == "" {
		return speech.EvaluateResult{}, errors.New("讯飞 ISE 未配置密钥(APPID/APIKey/APISecret)")
	}
	pcm := req.Audio
	if len(pcm) > wavHeaderSize { // 剥掉 WAV 头得裸 PCM(讯飞 aue:raw)
		pcm = pcm[wavHeaderSize:]
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, buildAuthURL(p.cfg.ISEHost, p.cfg.APIKey, p.cfg.APISecret), nil)
	if err != nil {
		return speech.EvaluateResult{}, fmt.Errorf("讯飞 ISE 连接失败: %w", err)
	}
	defer func() { _ = conn.Close() }()
	if d, ok := ctx.Deadline(); ok {
		_ = conn.SetWriteDeadline(d)
		_ = conn.SetReadDeadline(d)
	}

	// 发送在独立 goroutine，主协程读结果，避免 TCP 背压死锁。
	sendErr := make(chan error, 1)
	go func() { sendErr <- p.sendFrames(conn, req, pcm) }()

	// 读循环：直到结束帧(status=2)拿到 base64 XML。
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return speech.EvaluateResult{}, fmt.Errorf("讯飞 ISE 读取失败: %w", err)
		}
		var resp iseResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			return speech.EvaluateResult{}, fmt.Errorf("讯飞 ISE 响应解析失败: %w", err)
		}
		if resp.Code != 0 {
			return speech.EvaluateResult{}, fmt.Errorf("讯飞 ISE 返回错误 code=%d msg=%s sid=%s", resp.Code, resp.Message, resp.Sid)
		}
		if resp.Data.Status == dataStatusEnd {
			if serr := <-sendErr; serr != nil {
				return speech.EvaluateResult{}, serr
			}
			xmlBytes, err := base64.StdEncoding.DecodeString(resp.Data.Data)
			if err != nil {
				return speech.EvaluateResult{}, fmt.Errorf("讯飞 ISE 结果 base64 解码失败: %w", err)
			}
			return parseResult(xmlBytes)
		}
	}
}

// sendFrames 依次发送：参数帧(ssb) → 音频帧(auw, 首/中) → 末帧(auw, aus=4, status=2)。
func (p *ISEProvider) sendFrames(conn *websocket.Conn, req speech.EvaluateRequest, pcm []byte) error {
	// 参数帧
	first := iseRequest{
		Common: &iseCommon{AppID: p.cfg.AppID},
		Business: &iseBusiness{
			Category: "read_sentence",
			Rstcd:    "utf8",
			Sub:      "ise",
			Group:    "pupil",
			Ent:      entOf(req.Language),
			Tte:      "utf-8",
			Cmd:      "ssb",
			Auf:      "audio/L16;rate=16000",
			Aue:      "raw",
			Text:     textBOM + req.Text,
			TtpSkip:  true,
		},
		Data: iseData{Status: dataStatusBegin},
	}
	if err := conn.WriteJSON(first); err != nil {
		return fmt.Errorf("讯飞 ISE 发送参数帧失败: %w", err)
	}

	// 音频帧（按 1280B 切块，一帧一发，40ms 节奏模拟实时）
	total := len(pcm)
	for off := 0; off < total; off += audioChunk {
		end := off + audioChunk
		if end > total {
			end = total
		}
		aus := audioStatusMiddle
		if off == 0 {
			aus = audioStatusFirst
		}
		frame := iseRequest{
			Business: &iseBusiness{Cmd: "auw", Aus: aus, Aue: "raw"},
			Data: iseData{
				Status:   dataStatusMid,
				Data:     base64.StdEncoding.EncodeToString(pcm[off:end]),
				DataType: 1,
				Encoding: "raw",
			},
		}
		if err := conn.WriteJSON(frame); err != nil {
			return fmt.Errorf("讯飞 ISE 发送音频帧失败: %w", err)
		}
		time.Sleep(frameInterval)
	}

	// 末帧：aus=4, status=2，通知评测结束。
	last := iseRequest{
		Business: &iseBusiness{Cmd: "auw", Aus: audioStatusLast, Aue: "raw"},
		Data:     iseData{Status: dataStatusEnd, Data: "", DataType: 1, Encoding: "raw"},
	}
	if err := conn.WriteJSON(last); err != nil {
		return fmt.Errorf("讯飞 ISE 发送末帧失败: %w", err)
	}
	return nil
}

// entOf 语种映射：en 开头 → 英文评测 en_vip；其余暂按中文 cn_vip。
func entOf(language string) string {
	if strings.HasPrefix(strings.ToLower(language), "en") {
		return "en_vip"
	}
	return "cn_vip"
}
