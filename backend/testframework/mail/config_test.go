package mail_test
import (
	"github.com/valerii/hub/backend/internal/mail"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("HUB_MAIL_PORT", "")
	t.Setenv("MAIL_DAILY_LIMIT", "")
	t.Setenv("RESEND_FROM_ALERTS", "")
	t.Setenv("RESEND_FROM_EMAIL", "")

	cfg := mail.LoadConfig()
	assert.Equal(t, 3001, cfg.Port)
	assert.Equal(t, 20, cfg.DailyLimit)
	assert.Equal(t, "alerts@hub.local", cfg.FromAlerts)
	assert.True(t, cfg.AlertsEnabled)
}

func TestLoadConfig_CustomEnv(t *testing.T) {
	t.Setenv("HUB_MAIL_PORT", "4000")
	t.Setenv("MAIL_DAILY_LIMIT", "5")
	t.Setenv("RESEND_FROM_ALERTS", "custom@hub.local")
	t.Setenv("HUB_MAIL_INTERNAL_TOKEN", "tok")

	cfg := mail.LoadConfig()
	assert.Equal(t, 4000, cfg.Port)
	assert.Equal(t, 5, cfg.DailyLimit)
	assert.Equal(t, "custom@hub.local", cfg.FromAlerts)
	assert.Equal(t, "tok", cfg.InternalToken)
}

func TestLoadConfig_DailyLimitMinimum(t *testing.T) {
	t.Setenv("MAIL_DAILY_LIMIT", "0")
	cfg := mail.LoadConfig()
	assert.Equal(t, 20, cfg.DailyLimit)
}

func TestGetEnv(t *testing.T) {
	t.Setenv("MAIL_TEST_KEY", "value")
	assert.Equal(t, "value", mail.ExportGetEnv("MAIL_TEST_KEY", "fallback"))
	assert.Equal(t, "fallback", mail.ExportGetEnv("MAIL_MISSING_KEY_XYZ", "fallback"))
}
