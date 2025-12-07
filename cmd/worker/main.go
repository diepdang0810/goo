package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go1/config"
	"go1/internal/worker"
	"go1/pkg/logger"
)

func main() {
	// Initialize logger with default settings until config is loaded
	logger.SetLogger(logger.NewZapLogger("development"))

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal("Failed to load config", logger.Field{Key: "error", Value: err})
	}

	// Initialize worker application
	app, err := worker.NewApp(cfg)
	if err != nil {
		logger.Log.Fatal("Failed to initialize worker app", logger.Field{Key: "error", Value: err})
	}
	defer app.Close()

	// Create context that listens for the interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run worker
	if err := app.Run(ctx); err != nil && err != context.Canceled {
		logger.Log.Fatal("Worker failed", logger.Field{Key: "error", Value: err})
	}

	logger.Log.Info("Worker shut down gracefully")
}
