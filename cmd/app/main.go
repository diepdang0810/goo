package main

import (
	"log"

	"go1/config"
	"go1/internal/server"
	"go1/pkg/logger"
)

func main() {
	// Initialize logger with default settings (development) until config is loaded
	logger.SetLogger(logger.NewZapLogger("development"))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Log.Fatal("Failed to load config", logger.Field{Key: "error", Value: err})
	}
	logger.Log.Info("Loaded Config", logger.Field{Key: "config", Value: cfg})

	app, err := server.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
