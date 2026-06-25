package bootstrap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/echotalk/echotalk_server/internal/config"
)

// InitLogger 初始化 zap 日志。format=console 时输出彩色易读文本，否则输出 JSON。
func InitLogger(cfg config.LogConfig) (*zap.Logger, error) {
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	var zc zap.Config
	if cfg.Format == "console" {
		zc = zap.NewDevelopmentConfig() // console encoder，易读
		zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zc = zap.NewProductionConfig() // json encoder，给采集
	}
	zc.Level = level
	zc.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // epoch → 人类可读
	return zc.Build()
}
