package payment

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// membershipChecker 真实会员门禁判定，供 content 等模块消费。
// 通过方法签名结构化满足 content.MembershipChecker，payment 不反向 import content。
type membershipChecker struct {
	repo *Repository
}

// NewMembershipChecker 创建真实会员门禁判定器（自带仓库）。
func NewMembershipChecker(db *gorm.DB) *membershipChecker {
	return &membershipChecker{repo: NewRepository(db)}
}

// IsActiveMember 判断用户当前是否有效会员：存在 && 状态有效 && 未过期(实时按到期时间)。
func (c *membershipChecker) IsActiveMember(ctx context.Context, userID uint) (bool, error) {
	m, err := c.repo.GetMembershipByUserID(userID)
	if err != nil {
		return false, err
	}
	if m == nil {
		return false, nil
	}
	return m.Status == MembershipStatusActive && m.ExpireAt.After(time.Now()), nil
}
