package user

// RegisterRequest 注册请求（手机号 + 验证码 + 密码）。
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`                 // 手机号(格式在 service 正则校验)
	Password string `json:"password" binding:"required,min=6,max=32"` // 登录密码
	Nickname string `json:"nickname" binding:"max=64"`                // 昵称
	Code     string `json:"code" binding:"required"`                  // 短信验证码(注册时验手机号归属)
}

// SendCodeRequest 发送验证码请求。
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required"` // 手机号(格式在 service 正则校验)
}

// LoginRequest 登录请求（手机号 + 密码）。
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`    // 手机号
	Password string `json:"password" binding:"required"` // 登录密码
}

// RefreshRequest 刷新令牌请求。
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
