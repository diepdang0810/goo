package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go1/config"
	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Server struct {
	httpServer     *http.Server
	postgres       *postgres.Postgres
	redis          *redis.RedisClient
	kafka          *kafka.KafkaProducer
	config         *config.Config
	tracerProvider *sdktrace.TracerProvider
}

func New(cfg *config.Config) (*Server, error) {
	s := &Server{config: cfg}

	if err := s.initLogger(); err != nil {
		return nil, err
	}
	if err := s.initTelemetry(); err != nil {
		return nil, err
	}
	if err := s.initPostgres(); err != nil {
		return nil, err
	}
	if err := s.initRedis(); err != nil {
		return nil, err
	}
	if err := s.initKafka(); err != nil {
		return nil, err
	}

	s.initHTTPServer()

	return s, nil
}

func (s *Server) Run() error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Failed to listen and serve", logger.Field{Key: "error", Value: err})
		}
	}()

	logger.Log.Info("Server is running", logger.Field{Key: "port", Value: s.config.App.Port})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	if s.kafka != nil {
		s.kafka.Close()
	}
	if s.postgres != nil {
		s.postgres.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}

	// Shutdown tracer provider
	if s.tracerProvider != nil {
		if err := s.tracerProvider.Shutdown(context.Background()); err != nil {
			logger.Log.Warn("Failed to shutdown tracer provider", logger.Field{Key: "error", Value: err})
		}
	}

	logger.Log.Info("Server exited properly")
	return nil
}
