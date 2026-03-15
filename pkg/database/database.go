package database

import (
	"context"
	"fmt"
	"time"

	"github.com/prayosha/go-pos-backend/config"
	"github.com/prayosha/go-pos-backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	DB    *gorm.DB
	Redis *redis.Client
)

// opens a GORM connection pool to PostgreSQL.
func ConnectPostgres(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Kolkata",
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Name, cfg.Database.SSLMode,
	)

	logLevel := gormlogger.Silent
	if cfg.App.Env == "development" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:  gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	logger.Info("PostgreSQL connected")
	DB = db
	return db, nil
}

// ConnectRedis creates a Redis client and verifies connectivity.
func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	logger.Info("Redis connected")
	Redis = client
	return client, nil
}
