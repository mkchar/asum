package db

import (
	"asum/pkg/logx"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func New(conf Config) *DB {
	db, err := gorm.Open(postgres.Open(conf.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		logx.Errorf("连接数据库失败:%v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logx.Errorf("获取 sql.DB 失败:%v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return &DB{db}
}
