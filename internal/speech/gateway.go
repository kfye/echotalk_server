package speech

import (
	"context"

	"go.uber.org/zap"
)

// Gateway 语音能力网关：负责音频校验、限流、缓存、调用计费统计与厂商降级。
// 业务模块只依赖本网关，不直接耦合任何厂商 SDK。
type Gateway struct {
	provider Provider
	logger   *zap.Logger
}

// NewGateway 创建网关。
func NewGateway(provider Provider, logger *zap.Logger) *Gateway {
	return &Gateway{provider: provider, logger: logger}
}

// Evaluate 发音评测。厂商异常时降级兜底，不向上抛 500。
func (g *Gateway) Evaluate(ctx context.Context, req EvaluateRequest) (EvaluateResult, error) {
	if err := ValidateWAV(req.Audio); err != nil {
		return EvaluateResult{}, err
	}
	res, err := g.provider.Evaluate(ctx, req)
	if err != nil {
		g.logger.Warn("speech provider evaluate failed, degrading", zap.Error(err))
		return EvaluateResult{Degraded: true}, nil
	}
	return res, nil
}
