package mail

import (
	`bytes`
	`context`
	`crypto/sha256`
	`crypto/subtle`
	`encoding/base64`
	`encoding/hex`
	`encoding/json`
	`errors`
	`io`
	`time`

	`github.com/google/uuid`
	`github.com/redis/go-redis/v9`
)

const (
	mailboxBootstrapOAuthStateRedisPrefix = `hub:mail:oauth:bootstrap:state:`
	mailboxBootstrapOAuthStateJSONMaximum = 4096
	mailboxBootstrapOAuthStateJSONVersion = 1
	mailboxBootstrapOAuthStateTTL         = mailboxOAuthStateTTL
)

// MailboxBootstrapOAuthPending is the exact short-lived callback state for a
// password-enrollment bootstrap flow. It has no provider token or password.
type MailboxBootstrapOAuthPending struct {
	Scope                        MailboxOAuthScope
	EngineAccountID              string
	PasswordEnrollmentGeneration int64
	CodeVerifier                 string
	BrowserBindingHash           [sha256.Size]byte
	ExpiresAt                    time.Time
}

// MailboxBootstrapOAuthStateStore stores and atomically consumes only a
// bootstrap OAuth callback state. It is deliberately not interchangeable with
// the normal mailbox OAuth state store.
type MailboxBootstrapOAuthStateStore interface {
	SaveMailboxBootstrapOAuthState(context.Context, string, MailboxBootstrapOAuthPending) error
	ConsumeMailboxBootstrapOAuthState(context.Context, string, [sha256.Size]byte) (MailboxBootstrapOAuthPending, error)
}

// This tiny CAS delete never interprets the stored JSON. Go validates the
// complete bounded payload before it chooses whether to preserve or remove it.
var deleteMailboxBootstrapOAuthStatePayloadScript = redis.NewScript(`
local payload = redis.call('GET', KEYS[1])
if payload and payload == ARGV[1] then
  redis.call('DEL', KEYS[1])
  return 1
end
return 0
`)

// MailboxBootstrapOAuthRedisStateStore is the Redis implementation of the
// bootstrap-only state store. The client is intentionally private so callers
// cannot use it to cross into normal OAuth state keys.
type MailboxBootstrapOAuthRedisStateStore struct {
	client *redis.Client
}

type mailboxBootstrapOAuthRedisStateRecord struct {
	Version                        int
	ActorID                        string
	OrganizationID                 string
	AccountID                      string
	EngineAccountID                string
	PasswordEnrollmentGeneration   int64
	CodeVerifier                   string
	BindingHash                    string
	ExpiresAtUnix                  int64
}

// NewMailboxBootstrapOAuthRedisStateStore constructs a bootstrap-only store
// without performing Redis I/O.
func NewMailboxBootstrapOAuthRedisStateStore(client *redis.Client) (*MailboxBootstrapOAuthRedisStateStore, error) {
	if client == nil {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return &MailboxBootstrapOAuthRedisStateStore{client: client}, nil
}

// SaveMailboxBootstrapOAuthState writes one bounded callback state exactly
// once. Existing state is never replaced and its prior TTL is retained.
func (store *MailboxBootstrapOAuthRedisStateStore) SaveMailboxBootstrapOAuthState(ctx context.Context, state string, pending MailboxBootstrapOAuthPending) error {
	if store == nil || store.client == nil || ctx == nil || !validMailboxOAuthBase64URL(state, 43) || !validMailboxBootstrapOAuthPending(pending) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	record := map[string]any{
		`version`:                        mailboxBootstrapOAuthStateJSONVersion,
		`actor_id`:                       pending.Scope.ActorID.String(),
		`organization_id`:                pending.Scope.OrganizationID.String(),
		`account_id`:                     pending.Scope.AccountID.String(),
		`engine_account_id`:              pending.EngineAccountID,
		`password_enrollment_generation`: pending.PasswordEnrollmentGeneration,
		`code_verifier`:                  pending.CodeVerifier,
		`binding_hash`:                   base64.RawURLEncoding.EncodeToString(pending.BrowserBindingHash[:]),
		`expires_at_unix`:                pending.ExpiresAt.Unix(),
	}
	payload, err := json.Marshal(record)
	if err != nil || len(payload) == 0 || len(payload) > mailboxBootstrapOAuthStateJSONMaximum {
		clear(payload)
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	defer clear(payload)
	stored, err := store.client.SetNX(ctx, mailboxBootstrapOAuthRedisStateKey(state), payload, mailboxBootstrapOAuthStateTTL).Result()
	if err != nil {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if !stored {
		return mailboxOAuthError(MailboxOAuthErrorConflict)
	}
	return nil
}

// ConsumeMailboxBootstrapOAuthState verifies the browser binding and removes
// a matching state in the same Redis operation, preventing callback replay.
func (store *MailboxBootstrapOAuthRedisStateStore) ConsumeMailboxBootstrapOAuthState(ctx context.Context, state string, expectedBindingHash [sha256.Size]byte) (MailboxBootstrapOAuthPending, error) {
	if store == nil || store.client == nil || ctx == nil || !validMailboxOAuthBase64URL(state, 43) || expectedBindingHash == [sha256.Size]byte{} {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	key := mailboxBootstrapOAuthRedisStateKey(state)
	payload, err := store.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	if err != nil {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	defer clear(payload)
	if len(payload) == 0 || len(payload) > mailboxBootstrapOAuthStateJSONMaximum {
		store.deleteMailboxBootstrapOAuthStatePayload(ctx, key, payload)
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	pending, valid := decodeMailboxBootstrapOAuthRedisState(payload)
	if !valid {
		store.deleteMailboxBootstrapOAuthStatePayload(ctx, key, payload)
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if subtle.ConstantTimeCompare(pending.BrowserBindingHash[:], expectedBindingHash[:]) != 1 {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	if !store.deleteMailboxBootstrapOAuthStatePayload(ctx, key, payload) {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	return pending, nil
}

func (store *MailboxBootstrapOAuthRedisStateStore) deleteMailboxBootstrapOAuthStatePayload(ctx context.Context, key string, payload []byte) bool {
	result, err := deleteMailboxBootstrapOAuthStatePayloadScript.Run(ctx, store.client, []string{key}, string(payload)).Result()
	if err != nil {
		return false
	}
	deleted, ok := mailboxOAuthRedisInteger(result)
	return ok && deleted == 1
}

func validMailboxBootstrapOAuthPending(pending MailboxBootstrapOAuthPending) bool {
	return validMailboxOAuthScope(pending.Scope) && validMailboxOAuthOpaque(pending.EngineAccountID, MaxEngineOpaqueIDBytes) && pending.PasswordEnrollmentGeneration >= 1 && validMailboxOAuthBase64URL(pending.CodeVerifier, 86) && pending.BrowserBindingHash != [sha256.Size]byte{} && validMailboxOAuthTime(pending.ExpiresAt)
}

func mailboxBootstrapOAuthRedisStateKey(state string) string {
	rawState := []byte(state)
	defer clear(rawState)
	digest := sha256.Sum256(rawState)
	return mailboxBootstrapOAuthStateRedisPrefix + hex.EncodeToString(digest[:])
}

func decodeMailboxBootstrapOAuthRedisState(payload []byte) (MailboxBootstrapOAuthPending, bool) {
	if len(payload) == 0 || len(payload) > mailboxBootstrapOAuthStateJSONMaximum {
		return MailboxBootstrapOAuthPending{}, false
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var fields map[string]json.RawMessage
	if err := decoder.Decode(&fields); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF || len(fields) != 9 {
		return MailboxBootstrapOAuthPending{}, false
	}
	for _, field := range []string{`version`, `actor_id`, `organization_id`, `account_id`, `engine_account_id`, `password_enrollment_generation`, `code_verifier`, `binding_hash`, `expires_at_unix`} {
		if _, ok := fields[field]; !ok {
			return MailboxBootstrapOAuthPending{}, false
		}
	}
	if !mailboxBootstrapOAuthRedisFieldsExact(fields) {
		return MailboxBootstrapOAuthPending{}, false
	}
	var record mailboxBootstrapOAuthRedisStateRecord
	if err := json.Unmarshal(fields[`version`], &record.Version); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`actor_id`], &record.ActorID); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`organization_id`], &record.OrganizationID); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`account_id`], &record.AccountID); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`engine_account_id`], &record.EngineAccountID); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`password_enrollment_generation`], &record.PasswordEnrollmentGeneration); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`code_verifier`], &record.CodeVerifier); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`binding_hash`], &record.BindingHash); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if err := json.Unmarshal(fields[`expires_at_unix`], &record.ExpiresAtUnix); err != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	if record.Version != mailboxBootstrapOAuthStateJSONVersion || record.PasswordEnrollmentGeneration < 1 || !validMailboxOAuthBase64URL(record.CodeVerifier, 86) || !validMailboxOAuthBase64URL(record.BindingHash, 43) || record.ExpiresAtUnix < 1 {
		return MailboxBootstrapOAuthPending{}, false
	}
	binding, err := base64.RawURLEncoding.DecodeString(record.BindingHash)
	if err != nil || len(binding) != sha256.Size {
		clear(binding)
		return MailboxBootstrapOAuthPending{}, false
	}
	var bindingHash [sha256.Size]byte
	copy(bindingHash[:], binding)
	clear(binding)
	actorID, actorErr := uuid.Parse(record.ActorID)
	organizationID, organizationErr := uuid.Parse(record.OrganizationID)
	accountID, accountErr := uuid.Parse(record.AccountID)
	if actorErr != nil || organizationErr != nil || accountErr != nil {
		return MailboxBootstrapOAuthPending{}, false
	}
	pending := MailboxBootstrapOAuthPending{
		Scope:                        MailboxOAuthScope{ActorID: actorID, OrganizationID: organizationID, AccountID: accountID},
		EngineAccountID:              record.EngineAccountID,
		PasswordEnrollmentGeneration: record.PasswordEnrollmentGeneration,
		CodeVerifier:                 record.CodeVerifier,
		BrowserBindingHash:           bindingHash,
		ExpiresAt:                    time.Unix(record.ExpiresAtUnix, 0).UTC(),
	}
	if !validMailboxBootstrapOAuthPending(pending) {
		return MailboxBootstrapOAuthPending{}, false
	}
	return pending, true
}

func mailboxBootstrapOAuthRedisFieldsExact(fields map[string]json.RawMessage) bool {
	for field := range fields {
		switch field {
		case `version`, `actor_id`, `organization_id`, `account_id`, `engine_account_id`, `password_enrollment_generation`, `code_verifier`, `binding_hash`, `expires_at_unix`:
		default:
			return false
		}
	}
	return true
}

var _ MailboxBootstrapOAuthStateStore = (*MailboxBootstrapOAuthRedisStateStore)(nil)
