package database

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"online-mall-platform/server/config"
	"online-mall-platform/server/model"
	"time"
)

func Open(c config.Config) (*gorm.DB, *redis.Client, error) {
	db, e := gorm.Open(mysql.Open(c.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if e != nil {
		return nil, nil, e
	}
	sql, e := db.DB()
	if e != nil {
		return nil, nil, e
	}
	sql.SetMaxOpenConns(25)
	sql.SetMaxIdleConns(10)
	sql.SetConnMaxLifetime(5 * time.Minute)
	if e = db.AutoMigrate(&model.User{}, &model.Category{}, &model.Product{}, &model.Variant{}, &model.CartItem{}, &model.Order{}, &model.OrderItem{}, &model.Review{}, &model.AfterSale{}); e != nil {
		return nil, nil, e
	}
	r := redis.NewClient(&redis.Options{Addr: c.RedisAddr, Password: c.RedisPassword, DB: c.RedisDB, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if e = r.Ping(ctx).Err(); e != nil {
		return nil, nil, fmt.Errorf("redis: %w", e)
	}
	return db, r, nil
}
