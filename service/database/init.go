package database

import (
	"fmt"

	"github.com/PBH-BTN/trunker/biz/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	endpoint := config.AppConfig.Tracker.Database
	db, err := gorm.Open(mysql.Open(fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=Local&parseTime=True",
		endpoint.User, endpoint.Pass, endpoint.Host, endpoint.Port, endpoint.Database,
	)), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}
	return db
}
