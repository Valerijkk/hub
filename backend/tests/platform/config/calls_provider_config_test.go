//go:build hubtest

package config_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/config"
)

const validJitsiAppSecret = "0123456789abcdef0123456789abcdef"

func TestLoadDefaultsCallsProviderToLiveKit(t *testing.T) {
	setCallsConfigTestEnv(t, "development")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "livekit", cfg.Calls.ActiveProvider)
	assert.False(t, cfg.Calls.Jitsi.Enabled)
}

func TestLoadRejectsUnknownCallsProvider(t *testing.T) {
	setCallsConfigTestEnv(t, "development")
	t.Setenv("CALLS_ACTIVE_PROVIDER", "unknown-sfu")

	_, err := config.Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown calls provider")
	assert.Contains(t, err.Error(), "unknown-sfu")
}

func TestLoadRejectsJitsiAsActiveProviderBeforePhase152(t *testing.T) {
	setCallsConfigTestEnv(t, "development")
	setValidJitsiEnv(t)
	t.Setenv("CALLS_ACTIVE_PROVIDER", "jitsi")

	_, err := config.Load()
	require.EqualError(t, err, "jitsi provider is not available before phase 152")
}

func TestLoadAllowsDisabledJitsiWithEmptyFieldsInEveryMode(t *testing.T) {
	for _, environment := range []string{"development", "production"} {
		t.Run(environment, func(t *testing.T) {
			setCallsConfigTestEnv(t, environment)

			cfg, err := config.Load()
			require.NoError(t, err)
			assert.False(t, cfg.Calls.Jitsi.Enabled)
			assert.Empty(t, cfg.Calls.Jitsi.Host)
			assert.Empty(t, cfg.Calls.Jitsi.PublicURL)
			assert.Empty(t, cfg.Calls.Jitsi.AppID)
			assert.Empty(t, cfg.Calls.Jitsi.AppSecret)
			assert.Empty(t, cfg.Calls.Jitsi.XMPPDomain)
		})
	}
}

func TestLoadEnabledJitsiRequiresEveryFieldInEveryMode(t *testing.T) {
	fields := []struct {
		name string
		env  string
	}{
		{name: "host", env: "JITSI_HOST"},
		{name: "public URL", env: "JITSI_PUBLIC_URL"},
		{name: "app ID", env: "JITSI_APP_ID"},
		{name: "app secret", env: "JITSI_APP_SECRET"},
		{name: "XMPP domain", env: "JITSI_XMPP_DOMAIN"},
	}

	for _, environment := range []string{"development", "production"} {
		for _, field := range fields {
			t.Run(environment+"/missing "+field.name, func(t *testing.T) {
				setCallsConfigTestEnv(t, environment)
				setValidJitsiEnv(t)
				t.Setenv(field.env, "")

				_, err := config.Load()
				require.Error(t, err)
				assert.Contains(t, err.Error(), field.env)
				assert.NotContains(t, err.Error(), validJitsiAppSecret)
			})
		}
	}
}

func TestLoadEnabledJitsiRequiresCanonicalSelfHostedHTTPSOrigin(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		publicURL string
		errorKey  string
	}{
		{name: "HTTP", host: "meet.hub.test", publicURL: "http://meet.hub.test", errorKey: "JITSI_PUBLIC_URL"},
		{name: "userinfo", host: "meet.hub.test", publicURL: "https://user:password@meet.hub.test", errorKey: "JITSI_PUBLIC_URL"},
		{name: "root slash", host: "meet.hub.test", publicURL: "https://meet.hub.test/", errorKey: "JITSI_PUBLIC_URL"},
		{name: "path", host: "meet.hub.test", publicURL: "https://meet.hub.test/conference", errorKey: "JITSI_PUBLIC_URL"},
		{name: "query", host: "meet.hub.test", publicURL: "https://meet.hub.test?tenant=hub", errorKey: "JITSI_PUBLIC_URL"},
		{name: "fragment", host: "meet.hub.test", publicURL: "https://meet.hub.test#room", errorKey: "JITSI_PUBLIC_URL"},
		{name: "explicit port", host: "meet.hub.test", publicURL: "https://meet.hub.test:443", errorKey: "JITSI_PUBLIC_URL"},
		{name: "different host", host: "meet.hub.test", publicURL: "https://calls.hub.test", errorKey: "JITSI_PUBLIC_URL"},
		{name: "non-dedicated matching host", host: "calls.hub.test", publicURL: "https://calls.hub.test", errorKey: "JITSI_HOST"},
		{name: "scheme in host", host: "https://meet.hub.test", publicURL: "https://meet.hub.test", errorKey: "JITSI_HOST"},
		{name: "uppercase host", host: "Meet.hub.test", publicURL: "https://Meet.hub.test", errorKey: "JITSI_HOST"},
		{name: "trailing dot host", host: "meet.hub.test.", publicURL: "https://meet.hub.test.", errorKey: "JITSI_HOST"},
		{name: "bare meet with empty suffix", host: "meet.", publicURL: "https://meet.", errorKey: "JITSI_HOST"},
		{name: "empty DNS label", host: "meet..hub.test", publicURL: "https://meet..hub.test", errorKey: "JITSI_HOST"},
		{name: "underscore label", host: "meet.team_name.hub.test", publicURL: "https://meet.team_name.hub.test", errorKey: "JITSI_HOST"},
		{name: "wildcard label", host: "meet.*.hub.test", publicURL: "https://meet.*.hub.test", errorKey: "JITSI_HOST"},
		{name: "leading whitespace", host: " meet.hub.test", publicURL: "https:// meet.hub.test", errorKey: "JITSI_HOST"},
		{name: "trailing whitespace", host: "meet.hub.test ", publicURL: "https://meet.hub.test ", errorKey: "JITSI_HOST"},
		{name: "leading control character", host: "\tmeet.hub.test", publicURL: "https://\tmeet.hub.test", errorKey: "JITSI_HOST"},
		{name: "trailing control character", host: "meet.hub.test\n", publicURL: "https://meet.hub.test\n", errorKey: "JITSI_HOST"},
		{name: "leading hyphen label", host: "meet.-team.hub.test", publicURL: "https://meet.-team.hub.test", errorKey: "JITSI_HOST"},
		{name: "trailing hyphen label", host: "meet.team-.hub.test", publicURL: "https://meet.team-.hub.test", errorKey: "JITSI_HOST"},
		{name: "IP literal host", host: "192.0.2.10", publicURL: "https://192.0.2.10", errorKey: "JITSI_HOST"},
		{name: "port in host", host: "meet.hub.test:443", publicURL: "https://meet.hub.test:443", errorKey: "JITSI_HOST"},
		{name: "path in host", host: "meet.hub.test/conference", publicURL: "https://meet.hub.test/conference", errorKey: "JITSI_HOST"},
		{name: "external meet.jit.si", host: "meet.jit.si", publicURL: "https://meet.jit.si", errorKey: "JITSI_HOST"},
		{name: "external meet.jit.si subdomain", host: "team.meet.jit.si", publicURL: "https://team.meet.jit.si", errorKey: "JITSI_HOST"},
		{name: "meet-prefixed external meet.jit.si suffix", host: "meet.team.meet.jit.si", publicURL: "https://meet.team.meet.jit.si", errorKey: "JITSI_HOST"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setCallsConfigTestEnv(t, "development")
			setValidJitsiEnv(t)
			t.Setenv("JITSI_HOST", testCase.host)
			t.Setenv("JITSI_PUBLIC_URL", testCase.publicURL)

			_, err := config.Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.errorKey)
			assert.NotContains(t, err.Error(), validJitsiAppSecret)
		})
	}
}

func TestLoadEnabledJitsiEnforcesDNSHostLengthBoundaries(t *testing.T) {
	maxLabelHost := "meet." + strings.Repeat("a", 63) + ".test"
	maxTotalHost := "meet." + strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 56)
	tooLongLabelHost := "meet." + strings.Repeat("a", 64) + ".test"
	tooLongTotalHost := "meet." + strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 57)

	require.Len(t, maxTotalHost, 253)
	require.Len(t, tooLongTotalHost, 254)

	for _, host := range []string{maxLabelHost, maxTotalHost} {
		t.Run("accept "+host, func(t *testing.T) {
			setCallsConfigTestEnv(t, "development")
			setValidJitsiEnv(t)
			t.Setenv("JITSI_HOST", host)
			t.Setenv("JITSI_PUBLIC_URL", "https://"+host)

			cfg, err := config.Load()
			require.NoError(t, err)
			assert.Equal(t, host, cfg.Calls.Jitsi.Host)
			assert.Equal(t, "https://"+host, cfg.Calls.Jitsi.PublicURL)
		})
	}

	for name, host := range map[string]string{
		"64 byte label": tooLongLabelHost,
		"254 byte host": tooLongTotalHost,
	} {
		t.Run("reject "+name, func(t *testing.T) {
			setCallsConfigTestEnv(t, "development")
			setValidJitsiEnv(t)
			t.Setenv("JITSI_HOST", host)
			t.Setenv("JITSI_PUBLIC_URL", "https://"+host)

			_, err := config.Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "JITSI_HOST")
			assert.NotContains(t, err.Error(), validJitsiAppSecret)
		})
	}
}
func TestLoadEnabledJitsiKeepsHostAndXMPPSubjectSeparateAndExact(t *testing.T) {
	setCallsConfigTestEnv(t, "development")
	setValidJitsiEnv(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.True(t, cfg.Calls.Jitsi.Enabled)
	assert.Equal(t, "meet.hub.test", cfg.Calls.Jitsi.Host)
	assert.Equal(t, "https://meet.hub.test", cfg.Calls.Jitsi.PublicURL)
	assert.Equal(t, "hub-jitsi", cfg.Calls.Jitsi.AppID)
	assert.Equal(t, validJitsiAppSecret, cfg.Calls.Jitsi.AppSecret)
	assert.Equal(t, "xmpp.hub.internal", cfg.Calls.Jitsi.XMPPDomain)
}

func TestLoadEnabledJitsiRejectsWeakPlaceholderOrReusedAppSecret(t *testing.T) {
	strongReusedSecret := strings.Repeat("L", 32)
	tests := []struct {
		name          string
		secret        string
		liveKitSecret string
		hubJWTSecret  string
	}{
		{name: "short", secret: strings.Repeat("s", 31)},
		{name: "required marker", secret: "__REQUIRED_UNUSABLE_NO_DEFAULT__"},
		{name: "known dev secret", secret: "devsecret"},
		{name: "change me placeholder", secret: "change-me-in-production-0123456789"},
		{name: "change in production placeholder", secret: "jitsi-secret-change-in-production-0123456789"},
		{name: "reused LiveKit secret", secret: strongReusedSecret, liveKitSecret: strongReusedSecret},
		{name: "reused Hub JWT secret", secret: strongReusedSecret, hubJWTSecret: strongReusedSecret},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			setCallsConfigTestEnv(t, "development")
			setValidJitsiEnv(t)
			t.Setenv("JITSI_APP_SECRET", testCase.secret)
			if testCase.liveKitSecret != "" {
				t.Setenv("LIVEKIT_API_SECRET", testCase.liveKitSecret)
			}
			if testCase.hubJWTSecret != "" {
				t.Setenv("JWT_SECRET", testCase.hubJWTSecret)
			}

			_, err := config.Load()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "JITSI_APP_SECRET")
			assert.NotContains(t, err.Error(), testCase.secret)
		})
	}
}

func TestLoadRejectsIndependentJitsiIssuerOrAudienceInputs(t *testing.T) {
	for _, envName := range []string{"JITSI_ISSUER", "JITSI_AUDIENCE"} {
		for _, jitsiEnabled := range []bool{false, true} {
			mode := "disabled"
			if jitsiEnabled {
				mode = "enabled"
			}

			t.Run(envName+"/"+mode, func(t *testing.T) {
				setCallsConfigTestEnv(t, "development")
				if jitsiEnabled {
					setValidJitsiEnv(t)
				}
				t.Setenv(envName, "forbidden-shadow-value")

				_, err := config.Load()
				require.Error(t, err)
				assert.Contains(t, err.Error(), envName)
				assert.NotContains(t, err.Error(), validJitsiAppSecret)
				assert.NotContains(t, err.Error(), "forbidden-shadow-value")
			})
		}
	}
}

func setCallsConfigTestEnv(t *testing.T, environment string) {
	t.Helper()
	if environment == "production" {
		setProductionBaseEnv(t)
	} else {
		t.Setenv("DEV_MODE", "true")
		t.Setenv("ENVIRONMENT", "development")
	}

	for _, key := range []string{
		"CALLS_ACTIVE_PROVIDER",
		"JITSI_ENABLED",
		"JITSI_HOST",
		"JITSI_PUBLIC_URL",
		"JITSI_APP_ID",
		"JITSI_APP_SECRET",
		"JITSI_XMPP_DOMAIN",
		"JITSI_ISSUER",
		"JITSI_AUDIENCE",
	} {
		t.Setenv(key, "")
	}
}

func setValidJitsiEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JITSI_ENABLED", "true")
	t.Setenv("JITSI_HOST", "meet.hub.test")
	t.Setenv("JITSI_PUBLIC_URL", "https://meet.hub.test")
	t.Setenv("JITSI_APP_ID", "hub-jitsi")
	t.Setenv("JITSI_APP_SECRET", validJitsiAppSecret)
	t.Setenv("JITSI_XMPP_DOMAIN", "xmpp.hub.internal")
}
