// Command server is the campus-market backend entrypoint.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lp/campus-market/internal/config"
	"github.com/lp/campus-market/internal/model"
	"github.com/lp/campus-market/internal/router"
	"github.com/lp/campus-market/internal/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	logger := util.NewLogger()
	cfg := config.Load()

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		logger.Error("db connect failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("db pool failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(
		&model.User{}, &model.Product{}, &model.Conversation{}, &model.Message{},
		&model.TradeOrder{}, &model.Review{}, &model.BookExchange{},
	); err != nil {
		logger.Error("auto migrate failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if cfg.SeedingEnabled {
		if err := seed(context.Background(), db, logger); err != nil {
			logger.Error("seeding failed", slog.String("error", err.Error()))
		}
	}

	engine := router.New(cfg, db, logger)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Info("campus-market server listening", slog.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
