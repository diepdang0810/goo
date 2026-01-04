package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Jaeger   JaegerConfig   `mapstructure:"jaeger"`
	Temporal TemporalConfig `mapstructure:"temporal"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("app.name", "go1")
	v.SetDefault("app.port", "8080")
	v.SetDefault("app.env", "development")

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", "5432")
	v.SetDefault("postgres.max_conns", 10)
	v.SetDefault("postgres.min_conns", 2)

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("kafka.brokers", "localhost:9092")
	v.SetDefault("jaeger.endpoint", "localhost:4318")

	v.SetDefault("kafka.topics.shipment_events", "shipment-events")
	v.SetDefault("kafka.topics.dispatch_events", "dispatch-events")
	v.SetDefault("kafka.topics.order_events", "dbserver1.public.orders")

	v.SetDefault("temporal.hostPort", "localhost:7233")
	v.SetDefault("temporal.namespace", "default")
	v.SetDefault("temporal.taskQueue", "ORDER_TASK_QUEUE")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
