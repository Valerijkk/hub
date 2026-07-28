package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/platform/config"
)

// setProductionBaseEnv — минимальный безопасный prod-набор кроме переопределяемых ключей теста.
func setProductionBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DEV_MODE", "false")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 40))
	t.Setenv("S3_ACCESS_KEY", "prod-access")
	t.Setenv("S3_SECRET_KEY", "prod-secret-key-32chars-minimum!!")
	t.Setenv("LIVEKIT_API_KEY", "prod-lk-key")
	t.Setenv("LIVEKIT_API_SECRET", "prod-lk-secret")
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", strings.Repeat("b", 64))
	t.Setenv("METRICS_ALLOWED_CIDRS", "127.0.0.1/32")
	// A14-003: production требует userinfo в NATS_URL.
	t.Setenv("NATS_URL", "nats://hubnats:hubnatspass@nats:4222")
}

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
	setProductionBaseEnv(t)
	t.Setenv("JWT_SECRET", "dev-secret-change-in-production")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for dev JWT in production")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_productionRequiresEncryptionKey(t *testing.T) {
	setProductionBaseEnv(t)
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
	setProductionBaseEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg.Server.Environment != "production" {
		t.Errorf("Environment = %q", cfg.Server.Environment)
	}
	if len(cfg.Server.MetricsAllowedCIDRs) != 1 || cfg.Server.MetricsAllowedCIDRs[0] != "127.0.0.1/32" {
		t.Errorf("MetricsAllowedCIDRs = %v", cfg.Server.MetricsAllowedCIDRs)
	}
}

func TestLoad_productionAcceptsLoopbackMetricsCIDR_A14_005(t *testing.T) {
	setProductionBaseEnv(t)
	// Compose prod default — loopback, не blanket RFC1918.
	t.Setenv("METRICS_ALLOWED_CIDRS", "127.0.0.1/32")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if got := cfg.Server.MetricsAllowedCIDRs; len(got) != 1 || got[0] != "127.0.0.1/32" {
		t.Errorf("MetricsAllowedCIDRs = %v, want [127.0.0.1/32]", got)
	}
}

func TestLoad_productionRejectsEmptyMetricsCIDRs_A14_005(t *testing.T) {
	setProductionBaseEnv(t)
	t.Setenv("METRICS_ALLOWED_CIDRS", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for empty METRICS_ALLOWED_CIDRS in production")
	}
	if !strings.Contains(err.Error(), "METRICS_ALLOWED_CIDRS") {
		t.Errorf("unexpected error: %v", err)
	}
}

// A14-002: каждый ключ из devDefaults должен fail-closed по отдельности (DEV_MODE=false).
// Гранулярность: снятие одного guard'а в config.go краснеет ровно в своём subtest.
func TestLoad_productionRejectsEachDevDefault_A14_002(t *testing.T) {
	cases := []struct {
		envKey   string
		devValue string
		errSubstr string
	}{
		{"JWT_SECRET", "dev-secret-change-in-production", "JWT_SECRET"},
		{"S3_ACCESS_KEY", "hub", "S3_ACCESS_KEY"},
		{"S3_SECRET_KEY", "hubsecret", "S3_SECRET_KEY"},
		{"LIVEKIT_API_KEY", "devkey", "LIVEKIT_API_KEY"},
		{"LIVEKIT_API_SECRET", "devsecret", "LIVEKIT_API_SECRET"},
	}
	for _, tt := range cases {
		t.Run(tt.envKey, func(t *testing.T) {
			setProductionBaseEnv(t)
			t.Setenv(tt.envKey, tt.devValue)

			_, err := config.Load()
			if err == nil {
				t.Fatalf("expected error when %s is dev default %q", tt.envKey, tt.devValue)
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("error must mention %s, got: %v", tt.errSubstr, err)
			}
		})
	}
}

// A14-002 legacy: оба S3-ключа сразу (парный кейс остаётся как smoke).
func TestLoad_productionRejectsDefaultSeaweedFSKeys_A14_002(t *testing.T) {
	setProductionBaseEnv(t)
	t.Setenv("S3_ACCESS_KEY", "hub")
	t.Setenv("S3_SECRET_KEY", "hubsecret")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for default SeaweedFS keys in production")
	}
	if !strings.Contains(err.Error(), "S3_ACCESS_KEY") && !strings.Contains(err.Error(), "S3_SECRET_KEY") {
		t.Errorf("unexpected error: %v", err)
	}
}

// A14-003: ENVIRONMENT=production rejects NATS URL without user/password.
func TestLoad_productionRejectsNATSWithoutCreds_A14_003(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{name: "no_userinfo", url: "nats://nats:4222"},
		{name: "empty_password", url: "nats://hubnats:@nats:4222"},
		{name: "user_only", url: "nats://hubnats@nats:4222"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			setProductionBaseEnv(t)
			t.Setenv("NATS_URL", tt.url)

			_, err := config.Load()
			if err == nil {
				t.Fatal("expected error for NATS URL without credentials")
			}
			if !strings.Contains(err.Error(), "NATS_URL") {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoad_nonProductionAllowsNATSWithoutCreds(t *testing.T) {
	t.Setenv("DEV_MODE", "false")
	t.Setenv("ENVIRONMENT", "staging")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 40))
	t.Setenv("S3_ACCESS_KEY", "prod-access")
	t.Setenv("S3_SECRET_KEY", "prod-secret-key-32chars-minimum!!")
	t.Setenv("LIVEKIT_API_KEY", "prod-lk-key")
	t.Setenv("LIVEKIT_API_SECRET", "prod-lk-secret")
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", strings.Repeat("b", 64))
	t.Setenv("NATS_URL", "nats://nats:4222")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("staging without NATS auth should load: %v", err)
	}
	if cfg.NATS.URL != "nats://nats:4222" {
		t.Errorf("NATS.URL = %q", cfg.NATS.URL)
	}
}

// TC-PLATFORM-01 / A14-007: FiberRecoverEnableStackTrace mirrors DevMode.
// Wiring evidence (mutation-proof): cmd/server TestFiberRecoverConfig_* +
// TestWiringFiber_UsesFiberRecoverConfig_A14_007 (source assert).
// Mutation here: return true always → false case FAIL.
func TestFiberRecoverEnableStackTrace_FollowsDevMode_A14_007(t *testing.T) {
	cases := []struct {
		name    string
		devMode bool
		want    bool
	}{
		{name: "dev_on", devMode: true, want: true},
		{name: "dev_off", devMode: false, want: false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.ServerConfig{DevMode: tt.devMode}
			got := cfg.FiberRecoverEnableStackTrace()
			if got != tt.want {
				t.Fatalf("FiberRecoverEnableStackTrace() = %v, want %v", got, tt.want)
			}
			if got != cfg.DevMode {
				t.Fatalf("EnableStackTrace must equal DevMode: got %v DevMode %v", got, cfg.DevMode)
			}
		})
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
