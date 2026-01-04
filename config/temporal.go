package config

type TemporalConfig struct {
	HostPort  string `mapstructure:"hostPort"`
	Namespace string `mapstructure:"namespace"`
	TaskQueue string `mapstructure:"taskQueue"`
}
