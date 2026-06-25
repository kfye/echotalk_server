// Package codesender 验证码子系统：送达适配器(CodeSender) + 存储(CodeStore)。
//
// 设计：把「产/送达码」与「存/校验码」拆开——
//   - CodeSender：可替换的送达渠道。内测用 MockSender（固定码、不真发）；
//     日后接邮件/短信只新增实现，主流程不动。
//   - CodeStore：通用存储（Redis），带 TTL + 一次性消费，所有渠道共用，不随渠道变。
package codesender

import (
	"context"
	"time"
)

// CodeSender 验证码送达适配器：生成并送达验证码，返回所生成的码以便落库。
type CodeSender interface {
	Send(ctx context.Context, target string) (code string, err error)
}

// CodeStore 验证码存储：带 TTL 保存 + 一次性消费校验。
type CodeStore interface {
	// Save 保存待验证码，TTL 后自动过期。
	Save(ctx context.Context, scene, target, code string, ttl time.Duration) error
	// Consume 校验码：命中则删除（一次性），返回是否通过。
	Consume(ctx context.Context, scene, target, code string) (bool, error)
}
