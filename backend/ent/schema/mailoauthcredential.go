package schema

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

const (
	mailOAuthIssuerOriginMaxBytes = 2048
	mailOAuthCiphertextMinBytes   = 32
	mailOAuthCiphertextMaxBytes   = 16384
	mailOAuthGrantedScopeMaxBytes = 1024
)

// MailOAuthCredential stores only encrypted, account-bound delegated access
// material and its terminal lifecycle state.
type MailOAuthCredential struct {
	ent.Schema
}

// Fields defines the encrypted credential envelope. No plaintext OAuth token
// is persisted by this schema.
func (MailOAuthCredential) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).
			Unique().
			Immutable(),
		field.Enum("provider").
			Values("stalwart").
			Default("stalwart").
			Immutable(),
		field.String("issuer_origin").
			MinLen(1).
			MaxLen(mailOAuthIssuerOriginMaxBytes).
			Validate(validateMailOAuthIssuerOrigin).
			Immutable(),
		field.Enum("status").
			Values("active", "reconnect_required", "disconnected").
			Default("active"),
		field.Int64("credential_generation").
			Default(1).
			Min(1),
		field.Int32("encryption_format_version").
			Default(1).
			Min(1).
			Max(1),
		field.Bytes("access_token_ciphertext").
			Optional().
			Nillable().
			MinLen(mailOAuthCiphertextMinBytes).
			MaxLen(mailOAuthCiphertextMaxBytes).
			Sensitive(),
		field.Bytes("refresh_token_ciphertext").
			Optional().
			Nillable().
			MinLen(mailOAuthCiphertextMinBytes).
			MaxLen(mailOAuthCiphertextMaxBytes).
			Sensitive(),
		field.Enum("token_type").
			Values("bearer").
			Optional().
			Nillable(),
		field.String("granted_scope").
			Optional().
			Nillable().
			MinLen(1).
			MaxLen(mailOAuthGrantedScopeMaxBytes).
			Validate(validateMailOAuthGrantedScope),
		field.Time("access_token_expires_at").
			Optional().
			Nillable(),
		field.Time("refresh_token_expires_at").
			Optional().
			Nillable(),
		field.Time("connected_at").
			Default(nowUTC).
			Immutable(),
		field.Time("refreshed_at").
			Optional().
			Nillable(),
		field.Time("disconnected_at").
			Optional().
			Nillable(),
		field.Enum("last_error_code").
			Values("invalid_grant", "unauthorized_account", "account_mismatch").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Mixin adds optimistic locking to credential rotation and lifecycle changes.
func (MailOAuthCredential) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

// Annotations keeps every persisted state closed over a single safe shape.
func (MailOAuthCredential) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_oauth_credentials_provider": "provider = 'stalwart'",
			"mail_oauth_credentials_generation_version": "credential_generation >= 1 AND encryption_format_version = 1",
			"mail_oauth_credentials_issuer_origin_bounds": "octet_length(issuer_origin) BETWEEN 1 AND 2048",
			"mail_oauth_credentials_ciphertext_bounds": "(access_token_ciphertext IS NULL OR octet_length(access_token_ciphertext) BETWEEN 32 AND 16384) AND (refresh_token_ciphertext IS NULL OR octet_length(refresh_token_ciphertext) BETWEEN 32 AND 16384)",
			"mail_oauth_credentials_granted_scope_bounds": "granted_scope IS NULL OR octet_length(granted_scope) BETWEEN 1 AND 1024",
			"mail_oauth_credentials_access_pair": "(access_token_ciphertext IS NULL) = (access_token_expires_at IS NULL)",
			"mail_oauth_credentials_refresh_pair": "(refresh_token_ciphertext IS NULL) = (refresh_token_expires_at IS NULL)",
			"mail_oauth_credentials_bearer_only": "token_type IS NULL OR token_type = 'bearer'",
			"mail_oauth_credentials_error_allowlist": "last_error_code IS NULL OR last_error_code IN ('invalid_grant','unauthorized_account','account_mismatch')",
			"mail_oauth_credentials_status_state": "(status = 'active' AND access_token_ciphertext IS NOT NULL AND access_token_expires_at IS NOT NULL AND token_type IS NOT NULL AND token_type = 'bearer' AND granted_scope IS NOT NULL AND disconnected_at IS NULL AND last_error_code IS NULL) OR (status = 'reconnect_required' AND access_token_ciphertext IS NULL AND refresh_token_ciphertext IS NULL AND token_type IS NULL AND granted_scope IS NULL AND access_token_expires_at IS NULL AND refresh_token_expires_at IS NULL AND disconnected_at IS NULL AND last_error_code IS NOT NULL) OR (status = 'disconnected' AND access_token_ciphertext IS NULL AND refresh_token_ciphertext IS NULL AND token_type IS NULL AND granted_scope IS NULL AND access_token_expires_at IS NULL AND refresh_token_expires_at IS NULL AND disconnected_at IS NOT NULL AND last_error_code IS NULL)",
			"mail_oauth_credentials_refreshed_active": "refreshed_at IS NULL OR status = 'active'",
			"mail_oauth_credentials_refresh_expiry_order": "refresh_token_expires_at IS NULL OR (access_token_expires_at IS NOT NULL AND refresh_token_expires_at > access_token_expires_at)",
		}),
	}
}

// Edges binds one credential row immutably to its owning mail account. A hard
// account purge cascades only to this owned secret envelope.
func (MailOAuthCredential) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).
			Ref("oauth_credential").
			Field("mail_account_id").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func validateMailOAuthIssuerOrigin(value string) error {
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > mailOAuthIssuerOriginMaxBytes {
		return errors.New("invalid mail OAuth issuer origin")
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Opaque != "" || parsed.User != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawPath != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return errors.New("invalid mail OAuth issuer origin")
	}
	hostname := parsed.Hostname()
	if hostname == "" || hostname != strings.ToLower(hostname) || strings.HasSuffix(hostname, ".") {
		return errors.New("invalid mail OAuth issuer origin")
	}
	canonicalHost := hostname
	if address := net.ParseIP(hostname); address != nil {
		if address.String() != hostname {
			return errors.New("invalid mail OAuth issuer origin")
		}
		if strings.Contains(hostname, ":") {
			canonicalHost = "[" + hostname + "]"
		}
	} else if validateNormalizedMailDomain(hostname) != nil {
		return errors.New("invalid mail OAuth issuer origin")
	}
	port := parsed.Port()
	if port != "" {
		portNumber, portErr := strconv.Atoi(port)
		if portErr != nil || portNumber < 1 || portNumber > 65535 || strconv.Itoa(portNumber) != port || portNumber == 443 {
			return errors.New("invalid mail OAuth issuer origin")
		}
		canonicalHost += ":" + port
	}
	if value != "https://"+canonicalHost {
		return errors.New("invalid mail OAuth issuer origin")
	}
	return nil
}

func validateMailOAuthGrantedScope(value string) error {
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > mailOAuthGrantedScopeMaxBytes {
		return errors.New("invalid canonical mail OAuth scope")
	}
	var previous string
	for index, token := range strings.Split(value, " ") {
		if token == "" || (index > 0 && previous >= token) {
			return errors.New("invalid canonical mail OAuth scope")
		}
		for _, character := range []byte(token) {
			if character != 0x21 && (character < 0x23 || character > 0x5b) && (character < 0x5d || character > 0x7e) {
				return errors.New("invalid canonical mail OAuth scope")
			}
		}
		previous = token
	}
	return nil
}
