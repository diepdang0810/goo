package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go1/config"
	kafkaTestHandler "go1/internal/api/handlers/kafka_test"
	shipmentHandler "go1/internal/api/handlers/shipment"
	"go1/internal/modules/order"
	"go1/internal/modules/user"
	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/middleware"
	"go1/pkg/postgres"
	"go1/pkg/redis"
	"go1/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.temporal.io/sdk/client"
)

type App struct {
	httpServer     *http.Server
	postgres       *postgres.Postgres
	redis          *redis.RedisClient
	kafka          *kafka.KafkaProducer
	config         *config.Config
	tracerProvider *sdktrace.TracerProvider
	temporalClient client.Client
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{config: cfg}

	if err := app.initLogger(); err != nil {
		return nil, err
	}
	if err := app.initTelemetry(); err != nil {
		return nil, err
	}
	if err := app.initPostgres(); err != nil {
		return nil, err
	}
	if err := app.initRedis(); err != nil {
		return nil, err
	}
	if err := app.initKafka(); err != nil {
		return nil, err
	}
	if err := app.initTemporal(); err != nil {
		return nil, err
	}

	app.initHTTPServer()

	return app, nil
}

func (a *App) initLogger() error {
	log := logger.NewZapLogger(a.config.App.Env)
	logger.SetLogger(log)
	return nil
}

func (a *App) initTelemetry() error {
	tp, err := telemetry.InitTracer(a.config.App.Name, a.config.Jaeger.Endpoint)
	if err != nil {
		logger.Log.Warn("Failed to initialize OpenTelemetry", logger.Field{Key: "error", Value: err})
		return nil // Don't fail app startup if tracing fails
	}
	a.tracerProvider = tp
	logger.Log.Info("OpenTelemetry initialized", logger.Field{Key: "endpoint", Value: a.config.Jaeger.Endpoint})
	return nil
}

func (a *App) initPostgres() error {
	pg, err := postgres.New(
		a.config.Postgres.Host,
		a.config.Postgres.Port,
		a.config.Postgres.User,
		a.config.Postgres.Password,
		a.config.Postgres.DBName,
		a.config.Postgres.MaxConns,
		a.config.Postgres.MinConns,
	)
	if err != nil {
		return err
	}
	a.postgres = pg
	return nil
}

func (a *App) initRedis() error {
	rd, err := redis.New(a.config.Redis.Addr)
	if err != nil {
		return err
	}
	a.redis = rd
	return nil
}

func (a *App) initKafka() error {
	kf, err := kafka.NewProducer(a.config.Kafka.Brokers, a.config.Kafka.Producer)
	if err != nil {
		return err
	}
	a.kafka = kf
	return nil
}

func (a *App) initTemporal() error {
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233", // Should be from config
	})
	if err != nil {
		return err
	}
	a.temporalClient = c
	return nil
}

func (a *App) initHTTPServer() {
	if a.config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New() // Use gin.New() instead of gin.Default() to control middleware order

	// Register middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.TracingMiddleware(a.config.App.Name))
	router.Use(middleware.AuthMiddleware(true)) // Bypass auth by default

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize Modules
	user.Init(router, a.postgres.Pool, a.redis, a.kafka)
	order.Init(router, a.postgres.Pool, a.temporalClient)

	shipmentH := shipmentHandler.NewShipmentHandler(a.kafka)
	shipmentHandler.RegisterRoutes(router, shipmentH)

	// Initialize Kafka Test Handler
	kafkaTest := kafkaTestHandler.NewKafkaTestHandler(a.kafka)
	kafkaTestHandler.RegisterRoutes(router, kafkaTest)

	a.httpServer = &http.Server{
		Addr:    ":" + a.config.App.Port,
		Handler: router,
	}
}

func (a *App) Run() error {
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Failed to listen and serve", logger.Field{Key: "error", Value: err})
		}
	}()

	logger.Log.Info("Server is running", logger.Field{Key: "port", Value: a.config.App.Port})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	a.postgres.Close()
	a.redis.Close()
	a.kafka.Close()

	// Shutdown tracer provider
	if a.tracerProvider != nil {
		if err := a.tracerProvider.Shutdown(context.Background()); err != nil {
			logger.Log.Warn("Failed to shutdown tracer provider", logger.Field{Key: "error", Value: err})
		}
	}

	logger.Log.Info("Server exited properly")
	return nil
}
