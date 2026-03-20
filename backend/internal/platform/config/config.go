package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
	NATS     NATSConfig
	S3       S3Config
	JWT      JWTConfig
	Resend   ResendConfig
	Telegram TelegramConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Immutable    bool // true in dev to catch Fiber RequestCtx reuse (Pitfall 1)
}

type PostgresConfig struct {
	URL string
}

type MongoDBConfig struct {
	URL      string
	Database string
}

type RedisConfig struct {
	URL string
}

type NATSConfig struct {
	URL string
}

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type ResendConfig struct {
	APIKey    string
	FromEmail string
}

type TelegramConfig struct {
	BotToken string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 3000),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Immutable:    getEnvBool("FIBER_IMMUTABLE", true),
		},
		Postgres: PostgresConfig{
			URL: getEnv("DATABASE_URL", "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"),
		},
		MongoDB: MongoDBConfig{
			URL:      getEnv("MONGODB_URL", "mongodb://hub:hub@localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "hub"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		},
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
		},
		S3: S3Config{
			Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:8333"),
			AccessKey: getEnv("S3_ACCESS_KEY", "hub"),
			SecretKey: getEnv("S3_SECRET_KEY", "hubsecret"),
			Bucket:    getEnv("S3_BUCKET", "hub"),
			Region:    getEnv("S3_REGION", "us-east-1"),
			UseSSL:    getEnvBool("S3_USE_SSL", false),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 30 * 24 * time.Hour,
		},
		Resend: ResendConfig{
			APIKey:    getEnv("RESEND_API_KEY", ""),
			FromEmail: getEnv("RESEND_FROM_EMAIL", "noreply@hub.local"),
		},
		Telegram: TelegramConfig{
			BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
