package main

import (
	"github.com/prayosha/go-pos-backend/config"
	"github.com/prayosha/go-pos-backend/pkg/database"
	"github.com/prayosha/go-pos-backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.App.Env)
	defer logger.Sync()
	logger.Infof("Starting %s [%s]", cfg.App.Name, cfg.App.Env)

	// Databases
	db, err := database.ConnectPostgres(cfg)
	if err != nil {
		logger.Fatal("PostgreSQL connection failed", zap.Error(err))
	}
	redisClient, err := database.ConnectRedis(cfg)
	if err != nil {
		logger.Warn("Redis unavailable — sessions stored in-DB only", zap.Error(err))
	}
}
