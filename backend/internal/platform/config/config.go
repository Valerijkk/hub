// Package config загружает конфигурацию приложения из переменных окружения.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// minJWTSecretLen — минимальная длина JWT_SECRET в production. 32 байта
// дают ~256 бит энтропии и соответствуют рекомендации RFC 7518 / NIST.
const minJWTSecretLen = 32

// Известные dev-дефолты, которые недопустимы в production. Проверяются
// case-insensitive против присвоенных значений.
var devDefaults = map[string]string{
	"JWT_SECRET":         "dev-secret-change-in-production",
	"S3_ACCESS_KEY":      "hub",
	"S3_SECRET_KEY":      "hubsecret",
	"LIVEKIT_API_KEY":    "devkey",
	"LIVEKIT_API_SECRET": "devsecret",
}

// Config — главная конфигурация приложения, загружаемая из переменных окружения.
type Config struct {
	Server      ServerConfig
	Postgres    PostgresConfig
	MongoDB     MongoDBConfig
	Redis       RedisConfig
	NATS        NATSConfig
	S3          S3Config
	JWT         JWTConfig
	Resend      ResendConfig
	Telegram    TelegramConfig
	LiveKit     LiveKitConfig
	WebPush     WebPushConfig
	Integration IntegrationConfig
	AI            AIConfig
}

// ServerConfig — параметры HTTP-сервера (порт, таймауты, режим разработки).
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Immutable    bool   // true в dev для отлова повторного использования Fiber RequestCtx (Pitfall 1)
	DevMode             bool     // DevMode включает dev-эндпоинты (seed, login без Magic Link)
	Environment         string   // ENVIRONMENT: production, staging, development (по умолчанию)
	MetricsAllowedCIDRs []string // METRICS_ALLOWED_CIDRS — пусто non-prod → "*"; production validate fail-closed
	AlertWebhookSecret  string   // ALERT_WEBHOOK_SECRET — Bearer для Alertmanager webhook
}

// FiberRecoverEnableStackTrace — A14-007: Fiber recover пишет stack только в DevMode
// (иначе шум + утечка локальных путей в централизованные логи / Loki).
// Wiring: cmd/server/wiring_fiber.go → fiberRecoverConfig → EnableStackTrace.
func (c ServerConfig) FiberRecoverEnableStackTrace() bool {
	return c.DevMode
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

// TelegramConfig — параметры Telegram-бота для уведомлений (bind/delivery).
// Login через Telegram запрещён (ADR-AUTH-01); Magic Link only.
type TelegramConfig struct {
	BotToken    string
	BotUsername string
}

// LiveKitConfig содержит параметры подключения к серверу LiveKit.
type LiveKitConfig struct {
	URL       string // URL сервера LiveKit (например, "http://localhost:7880")
	APIKey    string // Ключ API для аутентификации
	APISecret string // Секрет API для подписи токенов
}

// WebPushConfig — параметры Web Push (VAPID) для браузерных push-уведомлений.
type WebPushConfig struct {
	VAPIDPublicKey  string // Публичный VAPID-ключ (base64url)
	VAPIDPrivateKey string // Приватный VAPID-ckлюч (base64url)
	VAPIDSubject    string // Контактный email (mailto:admin@example.com)
}

// IntegrationConfig — параметры шифрования секретов внешних интеграций
// (GitHub/GitLab access tokens, webhook secrets). См. HI-05.
type IntegrationConfig struct {
	// EncryptionKey — 32-байтовый ключ AES-256-GCM в hex или base64. Пустое
	// значение допустимо только в DevMode (используется детерминированный
	// dev-ключ). В production пустое значение приводит к fail-start.
	EncryptionKey string
}

// AIConfig — флаги AI Provider Hub (AI-SPEC S8).
type AIConfig struct {
	OfflineOnly bool
}

// Load загружает конфигурацию из переменных окружения с дефолтами и
// валидирует, что в production не остались dev-значения (CR-05, CR-06).
// Возвращает ошибку, чтобы вызывающий мог завершить процесс с понятным
// сообщением, а не падать позже на странных HTTP-ошибках.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 3000),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Immutable:    getEnvBool("FIBER_IMMUTABLE", true),
			DevMode:      getEnvBool("DEV_MODE", false),
			Environment:         getEnv("ENVIRONMENT", "development"),
			MetricsAllowedCIDRs: splitCSV(getEnv("METRICS_ALLOWED_CIDRS", "")),
			AlertWebhookSecret:  getEnv("ALERT_WEBHOOK_SECRET", ""),
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
			BotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
			BotUsername: getEnv("TELEGRAM_BOT_USERNAME", ""),
		},
		LiveKit: LiveKitConfig{
			URL:       getEnv("LIVEKIT_URL", "http://localhost:7880"),
			APIKey:    getEnv("LIVEKIT_API_KEY", "devkey"),
			APISecret: getEnv("LIVEKIT_API_SECRET", "devsecret"),
		},
		WebPush: WebPushConfig{
			VAPIDPublicKey:  getEnv("VAPID_PUBLIC_KEY", ""),
			VAPIDPrivateKey: getEnv("VAPID_PRIVATE_KEY", ""),
			VAPIDSubject:    getEnv("VAPID_SUBJECT", "mailto:admin@hub.local"),
		},
		Integration: IntegrationConfig{
			EncryptionKey: getEnv("INTEGRATION_ENCRYPTION_KEY", ""),
		},
		AI: AIConfig{
			OfflineOnly: getEnvBool("AI_OFFLINE_ONLY", false),
		},
	}

	// W4-A3-004: MetricsEndpoint fail-closed на пустом CIDR. Non-prod пустой
	// METRICS_ALLOWED_CIDRS → явный sentinel "*", иначе /metrics сразу 403.
	if len(cfg.Server.MetricsAllowedCIDRs) == 0 &&
		!strings.EqualFold(cfg.Server.Environment, "production") {
		cfg.Server.MetricsAllowedCIDRs = []string{"*"}
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate проверяет конфигурацию на безопасные значения для production:
// в production (DevMode=false) запрещены dev-дефолты JWT/S3/LiveKit и
// JWT_SECRET короче minJWTSecretLen. Также запрещена комбинация
// ENVIRONMENT=production + DEV_MODE=true — запрещено в config.validate.
// A14-002/A14-003: SeaweedFS hub/hubsecret и NATS без user/pass —
// fail-closed при ENVIRONMENT=production (и S3-дефолты при любом DEV_MODE=false).
func (c *Config) validate() error {
	if c.Server.DevMode && strings.EqualFold(c.Server.Environment, "production") {
		return errors.New("DEV_MODE=true несовместим с ENVIRONMENT=production — отключите DEV_MODE")
	}
	if c.Server.DevMode {
		// В dev-режиме дефолтные значения разрешены — они для того и нужны.
		return nil
	}

	var problems []string

	// JWT secret: запрещаем пустой, dev-дефолт и слишком короткий.
	if c.JWT.Secret == "" {
		problems = append(problems, "JWT_SECRET не задан")
	} else if isDevDefault("JWT_SECRET", c.JWT.Secret) ||
		strings.Contains(strings.ToLower(c.JWT.Secret), "change-in-production") {
		problems = append(problems, "JWT_SECRET использует dev-дефолт — задайте сильное значение")
	} else if len(c.JWT.Secret) < minJWTSecretLen {
		problems = append(problems, fmt.Sprintf("JWT_SECRET короче %d байт", minJWTSecretLen))
	}

	// A14-002: S3 credentials — запрещаем SeaweedFS dev-дефолты (hub/hubsecret).
	if isDevDefault("S3_ACCESS_KEY", c.S3.AccessKey) {
		problems = append(problems, "S3_ACCESS_KEY использует dev-дефолт (hub)")
	}
	if isDevDefault("S3_SECRET_KEY", c.S3.SecretKey) {
		problems = append(problems, "S3_SECRET_KEY использует dev-дефолт (hubsecret)")
	}

	// LiveKit credentials: запрещаем dev-дефолты (devkey/devsecret).
	if isDevDefault("LIVEKIT_API_KEY", c.LiveKit.APIKey) {
		problems = append(problems, "LIVEKIT_API_KEY использует dev-дефолт")
	}
	if isDevDefault("LIVEKIT_API_SECRET", c.LiveKit.APISecret) {
		problems = append(problems, "LIVEKIT_API_SECRET использует dev-дефолт")
	}

	// Ключ шифрования интеграций: обязателен в production (HI-05). Для
	// шифрования GitHub/GitLab tokens и webhook secrets at rest.
	if c.Integration.EncryptionKey == "" {
		problems = append(problems, "INTEGRATION_ENCRYPTION_KEY не задан (требуется 32 байта в hex или base64)")
	}

	// A14-005: в production /metrics обязан иметь CIDR allowlist (fail-closed).
	// Compose prod default — 127.0.0.1/32; blanket RFC1918 небезопасен (любой Docker peer).
	if strings.EqualFold(c.Server.Environment, "production") && len(c.Server.MetricsAllowedCIDRs) == 0 {
		problems = append(problems, "METRICS_ALLOWED_CIDRS пуст в production — задайте CIDR scrape-источника (напр. 127.0.0.1/32 или CIDR Prometheus)")
	}

	// A14-003: в ENVIRONMENT=production NATS URL обязан содержать non-empty user и password.
	if strings.EqualFold(c.Server.Environment, "production") {
		if msg := natsProductionCredsProblem(c.NATS.URL); msg != "" {
			problems = append(problems, msg)
		}
	}

	// P1-BE-26: weak DB/Mongo/Redis defaults (hub:hub@) запрещены при !DevMode —
	// симметрия с A14-002 (S3 hub/hubsecret).
	if msg := weakDataStoreURLProblem("DATABASE_URL", c.Postgres.URL); msg != "" {
		problems = append(problems, msg)
	}
	if msg := weakDataStoreURLProblem("MONGODB_URL", c.MongoDB.URL); msg != "" {
		problems = append(problems, msg)
	}
	if msg := weakDataStoreURLProblem("REDIS_URL", c.Redis.URL); msg != "" {
		problems = append(problems, msg)
	}

	if len(problems) > 0 {
		return fmt.Errorf("небезопасная конфигурация (DEV_MODE=false): %s", strings.Join(problems, "; "))
	}
	return nil
}

// weakDataStoreURLProblem — непустая строка, если URL содержит известные
// weak credentials hub:hub@ (compose/dev defaults).
func weakDataStoreURLProblem(name, raw string) string {
	if strings.Contains(strings.ToLower(raw), "hub:hub@") {
		return fmt.Sprintf("%s использует weak credentials hub:hub — задайте сильные учётные данные", name)
	}
	return ""
}

// natsProductionCredsProblem — пустая строка если NATS_URL содержит user+password;
// иначе человекочитаемая причина fail-start (A14-003).
func natsProductionCredsProblem(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return "NATS_URL некорректен в production"
	}
	if u.User == nil {
		return "NATS_URL без user/password в production — задайте nats://USER:PASS@host:4222"
	}
	user := strings.TrimSpace(u.User.Username())
	pass, hasPass := u.User.Password()
	if user == "" || !hasPass || pass == "" {
		return "NATS_URL: пустой user или password в production"
	}
	return ""
}

// isDevDefault сравнивает значение со списком известных небезопасных
// дефолтов. Сравнение регистронезависимое для устойчивости к опечаткам
// в конфигах деплоя.
func isDevDefault(key, value string) bool {
	dv, ok := devDefaults[key]
	if !ok {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(value), dv)
}

// splitCSV — CSV из env, пустые части пропускаем.
func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
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
