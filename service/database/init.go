package database

import (
	"fmt"
	"os"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	log := logger.Default.LogMode(logger.Info)
	if os.Getenv("RUN_ENV") == "prod" {
		log = logger.Default.LogMode(logger.Silent)
	}
	endpoint := config.AppConfig.Tracker.Database
	db, err := gorm.Open(mysql.Open(fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=Local&parseTime=True",
		endpoint.User, endpoint.Pass, endpoint.Host, endpoint.Port, endpoint.Database,
	)), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 log,
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(150)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}
