package mail

import (
	"os"
	"strconv"
	"time"
)

// Config — конфигурация микросервиса hub-mail.
type Config struct {
	Port          int
	ResendAPIKey  string
	FromAlerts    string
	DailyLimit    int
	AlertsEnabled bool
	InternalToken string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration

	// alertsSet — true если AlertsEnabled задан явно (LoadConfig / SetAlertsEnabled).
	// Нужен, чтобы литерал Config{} в тестах получал default AlertsEnabled=true.
	alertsSet bool
}

// LoadConfig читает env для hub-mail (Phase 15.2 skeleton).
func LoadConfig() Config {
	port, _ := strconv.Atoi(getEnv("HUB_MAIL_PORT", "3001"))
	limit, _ := strconv.Atoi(getEnv("MAIL_DAILY_LIMIT", "20"))
	if limit < 1 {
		limit = 20
	}
	enabled := getEnv("MAIL_ALERTS_ENABLED", "true") != "false"
	return Config{
		Port:          port,
		ResendAPIKey:  os.Getenv("RESEND_API_KEY"),
		FromAlerts:    getEnv("RESEND_FROM_ALERTS", getEnv("RESEND_FROM_EMAIL", "alerts@hub.local")),
		DailyLimit:    limit,
		AlertsEnabled: enabled,
		alertsSet:     true,
		InternalToken: os.Getenv("HUB_MAIL_INTERNAL_TOKEN"),
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  30 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
