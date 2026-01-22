package server

import (
	"net/http"

	apiMiddleware "go1/internal/api/middleware"
	"go1/internal/shared/order"

	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/middleware"
	"go1/pkg/postgres"
	"go1/pkg/redis"
	"go1/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) initLogger() error {
	log := logger.NewZapLogger(s.config.App.Env)
	logger.SetLogger(log)
	return nil
}

func (s *Server) initTelemetry() error {
	tp, err := telemetry.InitTracer(s.config.App.Name, s.config.Jaeger.Endpoint)
	if err != nil {
		logger.Log.Warn("Failed to initialize OpenTelemetry", logger.Field{Key: "error", Value: err})
		return nil // Don't fail app startup if tracing fails
	}
	s.tracerProvider = tp
	logger.Log.Info("OpenTelemetry initialized", logger.Field{Key: "endpoint", Value: s.config.Jaeger.Endpoint})
	return nil
}

func (s *Server) initPostgres() error {
	pg, err := postgres.New(
		s.config.Postgres.Host,
		s.config.Postgres.Port,
		s.config.Postgres.User,
		s.config.Postgres.Password,
		s.config.Postgres.DBName,
		s.config.Postgres.MaxConns,
		s.config.Postgres.MinConns,
	)
	if err != nil {
		return err
	}
	s.postgres = pg
	return nil
}

func (s *Server) initRedis() error {
	rd, err := redis.New(s.config.Redis.Addr)
	if err != nil {
		return err
	}
	s.redis = rd
	return nil
}

func (s *Server) initKafka() error {
	kf, err := kafka.NewProducer(s.config.Kafka.Brokers, s.config.Kafka.Producer)
	if err != nil {
		return err
	}
	s.kafka = kf
	return nil
}

func (s *Server) initHTTPServer() {
	if s.config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New() // Use gin.New() instead of gin.Default() to control middleware order

	// Register middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.TracingMiddleware(s.config.App.Name))
	router.Use(apiMiddleware.AuthMiddleware(false)) // Auth enabled

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize Modules
	order.Init(router, s.postgres.Pool)

	s.httpServer = &http.Server{
		Addr:    ":" + s.config.App.Port,
		Handler: router,
	}
}
