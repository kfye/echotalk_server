package content

import "context"

// MembershipChecker 判断用户是否为有效会员，用于付费内容门禁。
// 内测期用 noMembershipChecker 桩（一律非会员）；Day 5 会员模块就绪后换真实现，门禁逻辑不变。
type MembershipChecker interface {
	IsActiveMember(ctx context.Context, userID uint) (bool, error)
}

// noMembershipChecker 占位实现：任何人都视为非会员。
type noMembershipChecker struct{}

// NewNoMembershipChecker 创建非会员桩。
func NewNoMembershipChecker() MembershipChecker { return noMembershipChecker{} }

func (noMembershipChecker) IsActiveMember(ctx context.Context, userID uint) (bool, error) {
	return false, nil
}
