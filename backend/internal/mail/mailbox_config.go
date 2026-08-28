package mail

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxMailboxEngineConfigBooleanBytes                = 16
	maxMailboxEngineConfigOriginBytes                 = 2048
	maxMailboxEngineConfigDomainBytes                 = 253
	maxMailboxEngineConfigAuthModeBytes               = 64
	maxMailboxEngineConfigDomainVerificationModeBytes = 64
)

// MailboxConfigLookup supplies process configuration without coupling the mail domain
// to os.Getenv. Composition owns environment and secret-store access.
type MailboxConfigLookup func(string) (string, bool)

// MailboxEngineConfigStatus is the redacted readiness state used before a concrete
// mailbox engine is composed.
type MailboxEngineConfigStatus string

const (
	MailboxEngineConfigStatusNotConfigured                     MailboxEngineConfigStatus = "not_configured"
	MailboxEngineConfigStatusConfigured                        MailboxEngineConfigStatus = "configured"
	MailboxEngineConfigStatusInvalid                           MailboxEngineConfigStatus = "invalid"
	MailboxEngineConfigStatusDomainVerificationModeRequired    MailboxEngineConfigStatus = "domain_verification_mode_required"
	MailboxEngineConfigStatusDomainVerificationModeUnsupported MailboxEngineConfigStatus = "domain_verification_mode_unsupported"
	MailboxEngineConfigStatusAuthModeRequired                  MailboxEngineConfigStatus = "auth_mode_required"
	MailboxEngineConfigStatusAuthModeUnsupported               MailboxEngineConfigStatus = "auth_mode_unsupported"
)

// MailboxEngineMailAuthMode identifies the credential class for mailbox JMAP calls.
// No enabled auth mode is accepted until its least-privilege lifecycle is frozen.
type MailboxEngineMailAuthMode string

const (
	MailboxEngineMailAuthModeDisabled                   MailboxEngineMailAuthMode = "disabled"
	MailboxEngineMailAuthModeOAuthAuthorizationCodePKCE MailboxEngineMailAuthMode = "oauth_authorization_code_pkce"
)

// MailboxEngineDomainVerificationMode identifies the accepted public DNS
// verification strategy for an enabled mailbox engine.
type MailboxEngineDomainVerificationMode string

const (
	MailboxEngineDomainVerificationModeDisabled  MailboxEngineDomainVerificationMode = "disabled"
	MailboxEngineDomainVerificationModePublicDNS MailboxEngineDomainVerificationMode = "public_dns"
)

// MailboxEngineConfig contains only non-secret, process-owned engine settings. Origin
// is canonicalized to an origin without a trailing slash. Credentials, management API
// keys and provider-specific settings intentionally belong to later composition.
type MailboxEngineConfig struct {
	Enabled                 bool
	Status                  MailboxEngineConfigStatus
	Origin                  string
	Domain                  string
	ExternalDeliveryEnabled bool
	ExclusiveFolderWriter   bool
	MailAuthMode            MailboxEngineMailAuthMode
	DomainVerificationMode  MailboxEngineDomainVerificationMode
}

var ErrMailboxEngineConfigInvalid = errors.New("mailbox engine configuration invalid")

// MailboxEngineConfigErrorKind classifies configuration failures without retaining
// input values, credentials, URLs or provider responses.
type MailboxEngineConfigErrorKind string

const (
	MailboxEngineConfigErrorInvalidLookup                     MailboxEngineConfigErrorKind = "invalid_lookup"
	MailboxEngineConfigErrorValueTooLarge                     MailboxEngineConfigErrorKind = "value_too_large"
	MailboxEngineConfigErrorInvalidValue                      MailboxEngineConfigErrorKind = "invalid_value"
	MailboxEngineConfigErrorOriginRequired                    MailboxEngineConfigErrorKind = "origin_required"
	MailboxEngineConfigErrorInvalidOrigin                     MailboxEngineConfigErrorKind = "invalid_origin"
	MailboxEngineConfigErrorDomainRequired                    MailboxEngineConfigErrorKind = "domain_required"
	MailboxEngineConfigErrorInvalidDomain                     MailboxEngineConfigErrorKind = "invalid_domain"
	MailboxEngineConfigErrorDomainVerificationModeRequired    MailboxEngineConfigErrorKind = "domain_verification_mode_required"
	MailboxEngineConfigErrorDomainVerificationModeUnsupported MailboxEngineConfigErrorKind = "domain_verification_mode_unsupported"
	MailboxEngineConfigErrorAuthModeRequired                  MailboxEngineConfigErrorKind = "auth_mode_required"
	MailboxEngineConfigErrorAuthModeUnsupported               MailboxEngineConfigErrorKind = "auth_mode_unsupported"
)

// MailboxEngineConfigError is safe for logs because it includes only a fixed kind.
type MailboxEngineConfigError struct {
	Kind MailboxEngineConfigErrorKind
}

func (errorValue *MailboxEngineConfigError) Error() string {
	return ErrMailboxEngineConfigInvalid.Error()
}

func (errorValue *MailboxEngineConfigError) Unwrap() error {
	return ErrMailboxEngineConfigInvalid
}

// MailboxEngineConfigErrorKindOf extracts the stable, redacted configuration error.
func MailboxEngineConfigErrorKindOf(err error) (MailboxEngineConfigErrorKind, bool) {
	var configError *MailboxEngineConfigError
	if !errors.As(err, &configError) || configError == nil {
		return "", false
	}
	return configError.Kind, true
}

// ParseMailboxEngineConfig parses a bounded non-secret configuration through an injected
// lookup. It never reads credentials, performs I/O, resolves DNS, or enables a mailbox
// JMAP client by falling back to Resend, passwords, management API keys, or client creds.
func ParseMailboxEngineConfig(lookup MailboxConfigLookup) (MailboxEngineConfig, error) {
	enabled, err := mailboxEngineConfigBool(lookup, "MAILBOX_ENGINE_ENABLED", false)
	if err != nil {
		return mailboxEngineConfigFailure(MailboxEngineConfig{}, MailboxEngineConfigStatusInvalid, mailboxEngineConfigErrorKind(err))
	}
	if !enabled {
		return MailboxEngineConfig{
			Status:                 MailboxEngineConfigStatusNotConfigured,
			MailAuthMode:           MailboxEngineMailAuthModeDisabled,
			DomainVerificationMode: MailboxEngineDomainVerificationModeDisabled,
		}, nil
	}

	config := MailboxEngineConfig{
		Enabled:                true,
		Status:                 MailboxEngineConfigStatusInvalid,
		MailAuthMode:           MailboxEngineMailAuthModeDisabled,
		DomainVerificationMode: MailboxEngineDomainVerificationModeDisabled,
	}

	rawOrigin, present, err := mailboxEngineConfigValue(lookup, "MAILBOX_ENGINE_ORIGIN", maxMailboxEngineConfigOriginBytes)
	if err != nil {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, mailboxEngineConfigErrorKind(err))
	}
	if !present || rawOrigin == "" {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, MailboxEngineConfigErrorOriginRequired)
	}
	origin, valid := normalizeMailboxEngineOrigin(rawOrigin)
	if !valid {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, MailboxEngineConfigErrorInvalidOrigin)
	}
	config.Origin = origin

	externalDeliveryEnabled, err := mailboxEngineConfigBool(lookup, "MAILBOX_EXTERNAL_ENABLED", false)
	if err != nil {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, mailboxEngineConfigErrorKind(err))
	}
	config.ExternalDeliveryEnabled = externalDeliveryEnabled

	exclusiveFolderWriter, err := mailboxEngineConfigBool(lookup, "MAILBOX_EXCLUSIVE_FOLDER_WRITER", false)
	if err != nil {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, mailboxEngineConfigErrorKind(err))
	}
	config.ExclusiveFolderWriter = exclusiveFolderWriter

	rawDomain, domainPresent, err := mailboxEngineConfigValue(lookup, "MAILBOX_INSTANCE_DOMAIN", maxMailboxEngineConfigDomainBytes)
	if err != nil {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, mailboxEngineConfigErrorKind(err))
	}
	if domainPresent && rawDomain != "" {
		domain, valid := normalizeMailboxEngineDomain(rawDomain)
		if !valid {
			return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, MailboxEngineConfigErrorInvalidDomain)
		}
		config.Domain = domain
	}
	if config.ExternalDeliveryEnabled && config.Domain == "" {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, MailboxEngineConfigErrorDomainRequired)
	}

	rawDomainVerificationMode, domainVerificationModePresent, err := mailboxEngineConfigValue(lookup, "MAILBOX_DOMAIN_VERIFICATION_MODE", maxMailboxEngineConfigDomainVerificationModeBytes)
	if err != nil {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, mailboxEngineConfigErrorKind(err))
	}
	if !domainVerificationModePresent || rawDomainVerificationMode == "" || rawDomainVerificationMode == string(MailboxEngineDomainVerificationModeDisabled) {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusDomainVerificationModeRequired, MailboxEngineConfigErrorDomainVerificationModeRequired)
	}
	if rawDomainVerificationMode != string(MailboxEngineDomainVerificationModePublicDNS) {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusDomainVerificationModeUnsupported, MailboxEngineConfigErrorDomainVerificationModeUnsupported)
	}
	config.DomainVerificationMode = MailboxEngineDomainVerificationModePublicDNS

	rawAuthMode, authModePresent, err := mailboxEngineConfigValue(lookup, "MAILBOX_ENGINE_MAIL_AUTH_MODE", maxMailboxEngineConfigAuthModeBytes)
	if err != nil {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusInvalid, mailboxEngineConfigErrorKind(err))
	}
	if !authModePresent || rawAuthMode == "" || strings.EqualFold(rawAuthMode, string(MailboxEngineMailAuthModeDisabled)) {
		return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusAuthModeRequired, MailboxEngineConfigErrorAuthModeRequired)
	}
	if strings.EqualFold(rawAuthMode, string(MailboxEngineMailAuthModeOAuthAuthorizationCodePKCE)) {
		config.MailAuthMode = MailboxEngineMailAuthModeOAuthAuthorizationCodePKCE
		config.Status = MailboxEngineConfigStatusConfigured
		return config, nil
	}

	return mailboxEngineConfigFailure(config, MailboxEngineConfigStatusAuthModeUnsupported, MailboxEngineConfigErrorAuthModeUnsupported)
}

func mailboxEngineConfigFailure(config MailboxEngineConfig, status MailboxEngineConfigStatus, kind MailboxEngineConfigErrorKind) (MailboxEngineConfig, error) {
	config.Status = status
	return config, &MailboxEngineConfigError{Kind: kind}
}

func mailboxEngineConfigBool(lookup MailboxConfigLookup, key string, fallback bool) (bool, error) {
	raw, present, err := mailboxEngineConfigValue(lookup, key, maxMailboxEngineConfigBooleanBytes)
	if err != nil {
		return false, err
	}
	if !present {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, &MailboxEngineConfigError{Kind: MailboxEngineConfigErrorInvalidValue}
	}
	return value, nil
}

func mailboxEngineConfigValue(lookup MailboxConfigLookup, key string, maximum int) (string, bool, error) {
	if lookup == nil || maximum < 1 {
		return "", false, &MailboxEngineConfigError{Kind: MailboxEngineConfigErrorInvalidLookup}
	}
	raw, present := lookup(key)
	if !present {
		return "", false, nil
	}
	if len(raw) > maximum {
		return "", false, &MailboxEngineConfigError{Kind: MailboxEngineConfigErrorValueTooLarge}
	}
	if !utf8.ValidString(raw) {
		return "", false, &MailboxEngineConfigError{Kind: MailboxEngineConfigErrorInvalidValue}
	}
	return strings.TrimSpace(raw), true, nil
}

func mailboxEngineConfigErrorKind(err error) MailboxEngineConfigErrorKind {
	var configError *MailboxEngineConfigError
	if errors.As(err, &configError) && configError != nil {
		return configError.Kind
	}
	return MailboxEngineConfigErrorInvalidValue
}

func normalizeMailboxEngineOrigin(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > maxMailboxEngineConfigOriginBytes || !utf8.ValidString(value) {
		return "", false
	}

	if strings.Contains(value, "#") {
		return "", false
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed == nil || parsed.Opaque != "" || parsed.User != nil || parsed.Host == "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return "", false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", false
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", false
	}
	if parsed.RawPath != "" && parsed.RawPath != "/" {
		return "", false
	}
	if !validMailboxEngineOriginHost(parsed.Hostname(), parsed.Port()) {
		return "", false
	}
	return scheme + "://" + strings.ToLower(parsed.Host), true
}

func validMailboxEngineOriginHost(host, port string) bool {
	if host == "" || !utf8.ValidString(host) {
		return false
	}
	for _, character := range host {
		if unicode.IsControl(character) || unicode.IsSpace(character) || character > unicode.MaxASCII {
			return false
		}
	}
	if port == "" {
		return true
	}
	value, err := strconv.Atoi(port)
	return err == nil && value > 0 && value <= 65535
}

func normalizeMailboxEngineDomain(raw string) (string, bool) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" || len(value) > maxMailboxEngineConfigDomainBytes || !utf8.ValidString(value) || strings.HasSuffix(value, ".") {
		return "", false
	}
	labels := strings.Split(value, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", false
		}
		for _, character := range label {
			if character != '-' && (character < 'a' || character > 'z') && (character < '0' || character > '9') {
				return "", false
			}
		}
	}
	return value, true
}
