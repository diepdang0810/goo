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

type KafkaProducerConfig struct {
    RequiredAcks string `mapstructure:"requiredAcks"` // "all", "local", "none"
    RetryMax     int    `mapstructure:"retryMax"`
    Compression  string `mapstructure:"compression"` // "none", "gzip", "snappy", "lz4", "zstd"
    MaxMessageBytes int `mapstructure:"maxMessageBytes"`
}

type KafkaConsumerConfig struct {
    SessionTimeoutMs  int `mapstructure:"sessionTimeoutMs"`
    HeartbeatIntervalMs int `mapstructure:"heartbeatIntervalMs"`
    MaxProcessingTimeMs int `mapstructure:"maxProcessingTimeMs"`
}

type KafkaConfig struct {
    Brokers  string              `mapstructure:"brokers"`
    GroupID  string              `mapstructure:"groupId"`
    Producer KafkaProducerConfig `mapstructure:"producer"`
    Consumer KafkaConsumerConfig `mapstructure:"consumer"`
    Retry    KafkaRetryConfig    `mapstructure:"retry"`
}
