package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"automatictools/backend/internal/config"
	"automatictools/backend/internal/payment"
	"automatictools/backend/internal/router"
	"automatictools/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	db, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database failed", "error", err)
		os.Exit(1)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("get database connection failed", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := store.Migrate(db); err != nil {
		logger.Error("migrate database failed", "error", err)
		os.Exit(1)
	}
	if err := store.EnsureDefaultAdmin(db, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		logger.Error("initialize default admin failed", "error", err)
		os.Exit(1)
	}
	paymentProvider, err := payment.NewAlipay(cfg)
	if err != nil {
		logger.Error("initialize alipay failed", "error", err)
		os.Exit(1)
	}

	handler := router.New(router.Dependencies{
		Config:  cfg,
		DB:      db,
		Logger:  logger,
		Payment: paymentProvider,
	})

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("tools backend listening", "addr", cfg.Addr, "database", "postgresql")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
}
