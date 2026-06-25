package bootstrap

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/echotalk/echotalk_server/internal/config"
)

// InitDB 初始化 GORM/MySQL 连接池。
func InitDB(cfg config.MySQLConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	return db, nil
}
