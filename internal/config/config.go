// Package config 基于 viper 加载配置；敏感项由环境变量覆盖（前缀 ECHOTALK_）。
package config

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用总配置。
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	MySQL   MySQLConfig   `mapstructure:"mysql"`
	Redis   RedisConfig   `mapstructure:"redis"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Iflytek IflytekConfig `mapstructure:"iflytek"`
	COS     COSConfig     `mapstructure:"cos"`
	Log     LogConfig     `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug / release
}

type MySQLConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Issuer     string        `mapstructure:"issuer"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type IflytekConfig struct {
	AppID     string `mapstructure:"app_id"`
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
	ISEHost   string `mapstructure:"ise_host"`
}

type COSConfig struct {
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // console / json
}

// Load 读取配置文件并应用环境变量覆盖。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	setDefaults(v)

	// 环境变量覆盖，如 mysql.dsn -> ECHOTALK_MYSQL_DSN
	v.SetEnvPrefix("ECHOTALK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindEnvs(v) // 关键：显式绑定，修复 Unmarshal 读不到 env 的坑

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// bindEnvs 显式绑定可被环境变量覆盖的项。
// viper 的 AutomaticEnv 配合 Unmarshal 时，对未在配置文件出现的嵌套 key 不读取 env，
// 必须 BindEnv 才能让 .env 中的密钥真正注入。
func bindEnvs(v *viper.Viper) {
	for _, key := range []string{
		"mysql.dsn",
		"redis.password",
		"jwt.secret",
		"iflytek.app_id", "iflytek.api_key", "iflytek.api_secret",
		"cos.bucket", "cos.region", "cos.secret_id", "cos.secret_key",
	} {
		_ = v.BindEnv(key)
	}
}

// Validate 校验关键配置，缺失或仍为默认占位即报错。
func (c *Config) Validate() error {
	if c.MySQL.DSN == "" {
		return errors.New("config: mysql.dsn 为空（设置 ECHOTALK_MYSQL_DSN）")
	}
	switch c.JWT.Secret {
	case "", "change-me", "change-me-in-env":
		return errors.New("config: jwt.secret 未配置或仍为默认值（设置 ECHOTALK_JWT_SECRET）")
	}
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("mysql.max_open_conns", 20)
	v.SetDefault("mysql.max_idle_conns", 5)
	v.SetDefault("mysql.conn_max_lifetime", time.Hour)
	v.SetDefault("jwt.issuer", "echotalk")
	v.SetDefault("jwt.access_ttl", 2*time.Hour)
	v.SetDefault("jwt.refresh_ttl", 720*time.Hour)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
}
