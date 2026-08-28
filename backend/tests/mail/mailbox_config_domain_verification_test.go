package mail_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestParseMailboxEngineConfigDisabledDoesNotReadDomainVerificationMode(t *testing.T) {
	config, err := mail.ParseMailboxEngineConfig(func(key string) (string, bool) {
		if key == "MAILBOX_DOMAIN_VERIFICATION_MODE" {
			t.Fatal("disabled engine must not read domain verification mode")
		}
		if key != "MAILBOX_ENGINE_ENABLED" {
			t.Fatalf("disabled engine read unexpected key %q", key)
		}
		return "false", true
	})
	if err != nil {
		t.Fatalf("ParseMailboxEngineConfig() error = %v", err)
	}
	want := mail.MailboxEngineConfig{
		Status:                 mail.MailboxEngineConfigStatusNotConfigured,
		MailAuthMode:           mail.MailboxEngineMailAuthModeDisabled,
		DomainVerificationMode: mail.MailboxEngineDomainVerificationModeDisabled,
	}
	if config != want {
		t.Fatalf("ParseMailboxEngineConfig() = %#v, want %#v", config, want)
	}
}

func TestParseMailboxEngineConfigAcceptsPublicDNSDomainVerification(t *testing.T) {
	config, err := mail.ParseMailboxEngineConfig(mailboxDomainVerificationLookup(mailboxDomainVerificationValues()))
	if err != nil {
		t.Fatalf("ParseMailboxEngineConfig() error = %v", err)
	}
	want := mail.MailboxEngineConfig{
		Enabled:                 true,
		Status:                  mail.MailboxEngineConfigStatusConfigured,
		Origin:                  "https://mail.example.test",
		Domain:                  "example.test",
		ExternalDeliveryEnabled: true,
		ExclusiveFolderWriter:   true,
		MailAuthMode:            mail.MailboxEngineMailAuthModeOAuthAuthorizationCodePKCE,
		DomainVerificationMode:  mail.MailboxEngineDomainVerificationModePublicDNS,
	}
	if config != want {
		t.Fatalf("ParseMailboxEngineConfig() = %#v, want %#v", config, want)
	}
}

func TestParseMailboxEngineConfigRejectsInvalidDomainVerificationMode(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		present    bool
		value      string
		wantStatus mail.MailboxEngineConfigStatus
		wantKind   mail.MailboxEngineConfigErrorKind
	}{
		{
			name:       "missing",
			wantStatus: mail.MailboxEngineConfigStatusDomainVerificationModeRequired,
			wantKind:   mail.MailboxEngineConfigErrorDomainVerificationModeRequired,
		},
		{
			name:       "empty",
			present:    true,
			value:      "",
			wantStatus: mail.MailboxEngineConfigStatusDomainVerificationModeRequired,
			wantKind:   mail.MailboxEngineConfigErrorDomainVerificationModeRequired,
		},
		{
			name:       "disabled",
			present:    true,
			value:      "disabled",
			wantStatus: mail.MailboxEngineConfigStatusDomainVerificationModeRequired,
			wantKind:   mail.MailboxEngineConfigErrorDomainVerificationModeRequired,
		},
		{
			name:       "uppercase",
			present:    true,
			value:      "PUBLIC_DNS",
			wantStatus: mail.MailboxEngineConfigStatusDomainVerificationModeUnsupported,
			wantKind:   mail.MailboxEngineConfigErrorDomainVerificationModeUnsupported,
		},
		{
			name:       "unknown",
			present:    true,
			value:      "untrusted-mode",
			wantStatus: mail.MailboxEngineConfigStatusDomainVerificationModeUnsupported,
			wantKind:   mail.MailboxEngineConfigErrorDomainVerificationModeUnsupported,
		},
		{
			name:       "oversize",
			present:    true,
			value:      strings.Repeat("x", 65),
			wantStatus: mail.MailboxEngineConfigStatusInvalid,
			wantKind:   mail.MailboxEngineConfigErrorValueTooLarge,
		},
		{
			name:       "invalid utf8",
			present:    true,
			value:      string([]byte{0xff}),
			wantStatus: mail.MailboxEngineConfigStatusInvalid,
			wantKind:   mail.MailboxEngineConfigErrorInvalidValue,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			values := mailboxDomainVerificationValues()
			if testCase.present {
				values["MAILBOX_DOMAIN_VERIFICATION_MODE"] = testCase.value
			} else {
				delete(values, "MAILBOX_DOMAIN_VERIFICATION_MODE")
			}

			config, err := mail.ParseMailboxEngineConfig(mailboxDomainVerificationLookup(values))
			if !errors.Is(err, mail.ErrMailboxEngineConfigInvalid) || err == nil || err.Error() != mail.ErrMailboxEngineConfigInvalid.Error() || strings.Contains(err.Error(), "untrusted-mode") {
				t.Fatalf("ParseMailboxEngineConfig() error = %v, want redacted invalid configuration", err)
			}
			if config.Status != testCase.wantStatus || config.DomainVerificationMode != mail.MailboxEngineDomainVerificationModeDisabled {
				t.Fatalf("ParseMailboxEngineConfig() config = %#v, want status %q and disabled verification mode", config, testCase.wantStatus)
			}
			kind, ok := mail.MailboxEngineConfigErrorKindOf(err)
			if !ok || kind != testCase.wantKind {
				t.Fatalf("MailboxEngineConfigErrorKindOf() = %q, %t, want %q, true", kind, ok, testCase.wantKind)
			}
		})
	}
}

func mailboxDomainVerificationValues() map[string]string {
	return map[string]string{
		"MAILBOX_ENGINE_ENABLED":           "true",
		"MAILBOX_ENGINE_ORIGIN":            "https://MAIL.example.test/",
		"MAILBOX_EXTERNAL_ENABLED":         "true",
		"MAILBOX_EXCLUSIVE_FOLDER_WRITER":  "true",
		"MAILBOX_INSTANCE_DOMAIN":          "Example.TEST",
		"MAILBOX_DOMAIN_VERIFICATION_MODE": "public_dns",
		"MAILBOX_ENGINE_MAIL_AUTH_MODE":    "oauth_authorization_code_pkce",
	}
}

func mailboxDomainVerificationLookup(values map[string]string) mail.MailboxConfigLookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
