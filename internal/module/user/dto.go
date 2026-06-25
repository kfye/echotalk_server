package user

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"max=64"`
	Code     string `json:"code" binding:"required"`
}

// SendCodeRequest 发送验证码请求。
type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest 刷新令牌请求。
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
