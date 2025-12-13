package data

import (
	"fmt"
	"forgetting-curve/backend/internal/biz"
	"time"

	"forgetting-curve/backend/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Data 数据访问层
type Data struct {
	db *gorm.DB
}

// NewData 创建数据访问层
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	// 配置 GORM logger
	newLogger := log.NewHelper(logger)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&biz.Student{}, &biz.Word{}, &biz.ConfusedWord{}, &biz.Plan{}); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	cleanup := func() {
		newLogger.Info("closing the data resources")
		sqlDB.Close()
	}

	return &Data{db: db}, cleanup, nil
}

// GetDB 获取数据库连接
func (d *Data) GetDB() *gorm.DB {
	return d.db
}
