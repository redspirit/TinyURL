package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"tinyurl/internal/app"
	"tinyurl/internal/config"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appCtx := context.Background()

	a, err := app.Build(appCtx, cfg)
	if err != nil {
		panic(err)
	}
	a.Log.Info("server starting", "addr", cfg.HTTP.Address)

	serverErr := make(chan error, 1)

	go func() {
		if err := a.Server.Start(); err != nil {
			a.Log.Error("server error", "err", err)
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.Log.Info("received shutdown signal")
	case err := <-serverErr:
		a.Log.Error("server failed", "err", err)
	}

	a.Log.Info("shutting down...")

	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(shCtx); err != nil {
		a.Log.Error("shutdown error", "err", err)
	}
	if err := a.DBClose(); err != nil {
		a.Log.Error("db close error", "err", err)
	}
	a.Log.Info("bye")
}
