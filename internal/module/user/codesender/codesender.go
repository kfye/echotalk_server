// Package codesender 定义验证码发送/校验适配器。
// 内测用 MockSender（固定码）；日后接邮件/短信只新增实现，注册逻辑不动。
package codesender

import "context"

// CodeSender 验证码发送/校验适配器。
type CodeSender interface {
	Send(ctx context.Context, target string) error        // 发码（target=邮箱）
	Verify(ctx context.Context, target, code string) bool // 校验
}
