package config

type TopicRetryConfig struct {
    EnableRetry bool   `mapstructure:"enableRetry"`
    MaxAttempts int    `mapstructure:"maxAttempts"`
    BackoffMs   int    `mapstructure:"backoffMs"`
}

type KafkaRetryConfig struct {
    RetrySuffix string                      `mapstructure:"retrySuffix"`
    DLQSuffix   string                      `mapstructure:"dlqSuffix"`
    Topics      map[string]TopicRetryConfig `mapstructure:"topics"`
}

type KafkaConfig struct {
    Brokers string           `mapstructure:"brokers"`
    GroupID string           `mapstructure:"groupId"`
    Retry   KafkaRetryConfig `mapstructure:"retry"`
}
