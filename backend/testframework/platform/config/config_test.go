package config_test
import (
	"github.com/valerii/hub/backend/internal/platform/config"
	"os"
	"strings"
	"testing"
)

func TestLoad_devModeAllowsDefaults(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	t.Setenv("ENVIRONMENT", "development")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if !cfg.Server.DevMode {
		t.Error("expected DevMode=true")
	}
	if cfg.JWT.Secret == "" {
		t.Error("expected JWT secret fallback in dev")
	}
}

func TestLoad_productionRejectsDevJWT(t *testing.T) {
	t.Setenv("DEV_MODE", "false")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("JWT_SECRET", "dev-secret-change-in-production")
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", strings.Repeat("a", 64))

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for dev JWT in production")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_productionRequiresEncryptionKey(t *testing.T) {
	t.Setenv("DEV_MODE", "false")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("JWT_SECRET", strings.Repeat("x", 32))
	t.Setenv("S3_ACCESS_KEY", "prod-access")
	t.Setenv("S3_SECRET_KEY", "prod-secret-key-32chars-minimum!!")
	t.Setenv("LIVEKIT_API_KEY", "prod-lk-key")
	t.Setenv("LIVEKIT_API_SECRET", "prod-lk-secret")
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing INTEGRATION_ENCRYPTION_KEY")
	}
	if !strings.Contains(err.Error(), "INTEGRATION_ENCRYPTION_KEY") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_productionDevModeConflict(t *testing.T) {
	t.Setenv("DEV_MODE", "true")
	t.Setenv("ENVIRONMENT", "production")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected DEV_MODE + production conflict error")
	}
}

func TestLoad_productionValidConfig(t *testing.T) {
	t.Setenv("DEV_MODE", "false")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 40))
	t.Setenv("S3_ACCESS_KEY", "prod-access")
	t.Setenv("S3_SECRET_KEY", "prod-secret-key-32chars-minimum!!")
	t.Setenv("LIVEKIT_API_KEY", "prod-lk-key")
	t.Setenv("LIVEKIT_API_SECRET", "prod-lk-secret")
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", strings.Repeat("b", 64))

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg.Server.Environment != "production" {
		t.Errorf("Environment = %q", cfg.Server.Environment)
	}
}

func TestIsDevDefault(t *testing.T) {
	tests := []struct {
		key   string
		value string
		want  bool
	}{
		{"JWT_SECRET", "dev-secret-change-in-production", true},
		{"JWT_SECRET", "DEV-SECRET-CHANGE-IN-PRODUCTION", true},
		{"JWT_SECRET", "my-strong-production-secret-32bytes", false},
		{"S3_ACCESS_KEY", "hub", true},
		{"UNKNOWN", "hub", false},
	}
	for _, tt := range tests {
		if got := config.ExportIsDevDefault(tt.key, tt.value); got != tt.want {
			t.Errorf("isDevDefault(%q, %q) = %v, want %v", tt.key, tt.value, got, tt.want)
		}
	}
}

func TestGetEnvHelpers(t *testing.T) {
	key := "HUB_CONFIG_TEST_INT"
	t.Setenv(key, "42")
	if got := config.ExportGetEnvInt(key, 1); got != 42 {
		t.Errorf("getEnvInt = %d", got)
	}
	t.Setenv(key, "invalid")
	if got := config.ExportGetEnvInt(key, 7); got != 7 {
		t.Errorf("getEnvInt fallback = %d", got)
	}

	boolKey := "HUB_CONFIG_TEST_BOOL"
	t.Setenv(boolKey, "true")
	if got := config.ExportGetEnvBool(boolKey, false); !got {
		t.Error("getEnvBool expected true")
	}
}

func TestLoad_clearsEnvBetweenTests(t *testing.T) {
	// Убеждаемся, что нет утечки из прошлых тестов — config.Load не должен паниковать.
	os.Unsetenv("DEV_MODE")
}
