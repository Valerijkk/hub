package config

import (
	"os"
	"strconv"
	"time"
)

// Config — главная конфигурация приложения, загружаемая из переменных окружения.
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
	LiveKit  LiveKitConfig
}

// ServerConfig — параметры HTTP-сервера (порт, таймауты, режим разработки).
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Immutable    bool // true in dev to catch Fiber RequestCtx reuse (Pitfall 1)
	DevMode      bool // DevMode включает dev-эндпоинты (seed, login без Magic Link)
}

// PostgresConfig — параметры подключения к PostgreSQL.
type PostgresConfig struct {
	URL string
}

// MongoDBConfig — параметры подключения к MongoDB.
type MongoDBConfig struct {
	URL      string
	Database string
}

// RedisConfig — параметры подключения к Redis.
type RedisConfig struct {
	URL string
}

// NATSConfig — параметры подключения к NATS.
type NATSConfig struct {
	URL string
}

// S3Config — параметры подключения к S3-совместимому хранилищу (SeaweedFS).
type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

// JWTConfig — параметры JWT-токенов (секрет, TTL access/refresh).
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// ResendConfig — параметры сервиса отправки email через Resend API.
type ResendConfig struct {
	APIKey    string
	FromEmail string
}

// TelegramConfig — параметры Telegram-бота для аутентификации и уведомлений.
type TelegramConfig struct {
	BotToken string
}

// LiveKitConfig содержит параметры подключения к серверу LiveKit.
type LiveKitConfig struct {
	URL       string // URL сервера LiveKit (например, "http://localhost:7880")
	APIKey    string // Ключ API для аутентификации
	APISecret string // Секрет API для подписи токенов
}

// Load загружает конфигурацию из переменных окружения с значениями по умолчанию.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 3000),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Immutable:    getEnvBool("FIBER_IMMUTABLE", true),
			DevMode:      getEnvBool("DEV_MODE", false),
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
		LiveKit: LiveKitConfig{
			URL:       getEnv("LIVEKIT_URL", "http://localhost:7880"),
			APIKey:    getEnv("LIVEKIT_API_KEY", "devkey"),
			APISecret: getEnv("LIVEKIT_API_SECRET", "devsecret"),
		},
	}
}

// getEnv возвращает значение переменной окружения или fallback, если не задана.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt возвращает целочисленное значение переменной окружения или fallback.
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

// getEnvBool возвращает булево значение переменной окружения или fallback.
func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
