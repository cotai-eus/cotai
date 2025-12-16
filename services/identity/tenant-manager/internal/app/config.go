package app

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Kafka       KafkaConfig
	JWT         JWTConfig
	Observability ObservabilityConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port     int    `mapstructure:"PORT"`
	GRPCPort int    `mapstructure:"GRPC_PORT"`
	Env      string `mapstructure:"ENV"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"DATABASE_HOST"`
	Port            int           `mapstructure:"DATABASE_PORT"`
	Name            string        `mapstructure:"DATABASE_NAME"`
	User            string        `mapstructure:"DATABASE_USER"`
	Password        string        `mapstructure:"DATABASE_PASSWORD"`
	MaxConns        int           `mapstructure:"DATABASE_MAX_CONNS"`
	MaxIdleConns    int           `mapstructure:"DATABASE_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DATABASE_CONN_MAX_LIFETIME"`
	SSLMode         string        `mapstructure:"DATABASE_SSL_MODE"`
	MigrationsPath  string        `mapstructure:"MIGRATIONS_PATH"`
}

// ConnectionString returns the PostgreSQL connection string
func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers               []string `mapstructure:"KAFKA_BROKERS"`
	TopicTenantLifecycle  string   `mapstructure:"KAFKA_TOPIC_TENANT_LIFECYCLE"`
	ClientID              string   `mapstructure:"KAFKA_CLIENT_ID"`
	Acks                  int      `mapstructure:"KAFKA_ACKS"`
	Compression           string   `mapstructure:"KAFKA_COMPRESSION"`
	MaxRetry              int      `mapstructure:"KAFKA_MAX_RETRY"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PublicKeyURL string `mapstructure:"JWT_PUBLIC_KEY_URL"`
	Issuer       string `mapstructure:"JWT_ISSUER"`
	Audience     string `mapstructure:"JWT_AUDIENCE"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	JaegerAgentHost   string  `mapstructure:"JAEGER_AGENT_HOST"`
	JaegerAgentPort   int     `mapstructure:"JAEGER_AGENT_PORT"`
	JaegerServiceName string  `mapstructure:"JAEGER_SERVICE_NAME"`
	JaegerSamplerType string  `mapstructure:"JAEGER_SAMPLER_TYPE"`
	JaegerSamplerParam float64 `mapstructure:"JAEGER_SAMPLER_PARAM"`
	PrometheusEnabled bool    `mapstructure:"PROMETHEUS_ENABLED"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("PORT", 8082)
	viper.SetDefault("GRPC_PORT", 9082)
	viper.SetDefault("ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")

	viper.SetDefault("DATABASE_MAX_CONNS", 25)
	viper.SetDefault("DATABASE_MAX_IDLE_CONNS", 5)
	viper.SetDefault("DATABASE_CONN_MAX_LIFETIME", "300s")
	viper.SetDefault("DATABASE_SSL_MODE", "disable")
	viper.SetDefault("MIGRATIONS_PATH", "file://migrations/tenant_schema")

	viper.SetDefault("KAFKA_ACKS", 1)
	viper.SetDefault("KAFKA_COMPRESSION", "snappy")
	viper.SetDefault("KAFKA_MAX_RETRY", 3)
	viper.SetDefault("KAFKA_TOPIC_TENANT_LIFECYCLE", "tenant.lifecycle")

	viper.SetDefault("JAEGER_SAMPLER_TYPE", "probabilistic")
	viper.SetDefault("JAEGER_SAMPLER_PARAM", 0.1)
	viper.SetDefault("PROMETHEUS_ENABLED", true)

	config := &Config{}

	config.Server.Port = viper.GetInt("PORT")
	config.Server.GRPCPort = viper.GetInt("GRPC_PORT")
	config.Server.Env = viper.GetString("ENV")
	config.Server.LogLevel = viper.GetString("LOG_LEVEL")

	config.Database.Host = viper.GetString("DATABASE_HOST")
	config.Database.Port = viper.GetInt("DATABASE_PORT")
	config.Database.Name = viper.GetString("DATABASE_NAME")
	config.Database.User = viper.GetString("DATABASE_USER")
	config.Database.Password = viper.GetString("DATABASE_PASSWORD")
	config.Database.MaxConns = viper.GetInt("DATABASE_MAX_CONNS")
	config.Database.MaxIdleConns = viper.GetInt("DATABASE_MAX_IDLE_CONNS")
	config.Database.ConnMaxLifetime = viper.GetDuration("DATABASE_CONN_MAX_LIFETIME")
	config.Database.SSLMode = viper.GetString("DATABASE_SSL_MODE")
	config.Database.MigrationsPath = viper.GetString("MIGRATIONS_PATH")

	// Parse Kafka brokers (comma-separated)
	brokers := viper.GetString("KAFKA_BROKERS")
	if brokers != "" {
		config.Kafka.Brokers = []string{brokers}
	}
	config.Kafka.TopicTenantLifecycle = viper.GetString("KAFKA_TOPIC_TENANT_LIFECYCLE")
	config.Kafka.ClientID = viper.GetString("KAFKA_CLIENT_ID")
	config.Kafka.Acks = viper.GetInt("KAFKA_ACKS")
	config.Kafka.Compression = viper.GetString("KAFKA_COMPRESSION")
	config.Kafka.MaxRetry = viper.GetInt("KAFKA_MAX_RETRY")

	config.JWT.PublicKeyURL = viper.GetString("JWT_PUBLIC_KEY_URL")
	config.JWT.Issuer = viper.GetString("JWT_ISSUER")
	config.JWT.Audience = viper.GetString("JWT_AUDIENCE")

	config.Observability.JaegerAgentHost = viper.GetString("JAEGER_AGENT_HOST")
	config.Observability.JaegerAgentPort = viper.GetInt("JAEGER_AGENT_PORT")
	config.Observability.JaegerServiceName = viper.GetString("JAEGER_SERVICE_NAME")
	config.Observability.JaegerSamplerType = viper.GetString("JAEGER_SAMPLER_TYPE")
	config.Observability.JaegerSamplerParam = viper.GetFloat64("JAEGER_SAMPLER_PARAM")
	config.Observability.PrometheusEnabled = viper.GetBool("PROMETHEUS_ENABLED")

	return config, nil
}
