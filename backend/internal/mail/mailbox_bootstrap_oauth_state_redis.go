package mail

import (
	`bytes`
	`context`
	`crypto/sha256`
	`crypto/subtle`
	`encoding/base64`
	`encoding/hex`
	`encoding/json`
	`io`
	`strconv`
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

// This compare-and-delete operation preserves an otherwise valid state for a
// wrong browser binding, while any malformed, unknown-version, or oversize
// payload is deleted atomically before it can be retried.
var consumeMailboxBootstrapOAuthStateScript = redis.NewScript(`
local payload = redis.call('GET', KEYS[1])
if not payload then
  return {0}
end
local maximum = tonumber(ARGV[2])
if not maximum or maximum < 1 or maximum > 4096 or string.len(payload) > maximum then
  redis.call('DEL', KEYS[1])
  return {-1}
end
local decoded_ok, decoded = pcall(cjson.decode, payload)
if not decoded_ok or type(decoded) ~= 'table' then
  redis.call('DEL', KEYS[1])
  return {-1}
end
local allowed = {
  ['version'] = true,
  ['actor_id'] = true,
  ['organization_id'] = true,
  ['account_id'] = true,
  ['engine_account_id'] = true,
  ['password_enrollment_generation'] = true,
  ['code_verifier'] = true,
  ['binding_hash'] = true,
  ['expires_at_unix'] = true
}
local count = 0
for key, _ in pairs(decoded) do
  if not allowed[key] then
    redis.call('DEL', KEYS[1])
    return {-1}
  end
  count = count + 1
end
if count ~= 9 or type(decoded['version']) ~= 'number' or decoded['version'] ~= 1 or type(decoded['actor_id']) ~= 'string' or type(decoded['organization_id']) ~= 'string' or type(decoded['account_id']) ~= 'string' or type(decoded['engine_account_id']) ~= 'string' or type(decoded['password_enrollment_generation']) ~= 'number' or type(decoded['code_verifier']) ~= 'string' or type(decoded['binding_hash']) ~= 'string' or type(decoded['expires_at_unix']) ~= 'number' then
  redis.call('DEL', KEYS[1])
  return {-1}
end
if decoded['password_enrollment_generation'] < 1 or decoded['password_enrollment_generation'] ~= math.floor(decoded['password_enrollment_generation']) or decoded['expires_at_unix'] < 1 or decoded['expires_at_unix'] ~= math.floor(decoded['expires_at_unix']) or string.len(decoded['engine_account_id']) < 1 or string.len(decoded['code_verifier']) ~= 86 or string.len(decoded['binding_hash']) ~= 43 or not string.match(decoded['code_verifier'], '^[A-Za-z0-9_-]+$') or not string.match(decoded['binding_hash'], '^[A-Za-z0-9_-]+$') then
  redis.call('DEL', KEYS[1])
  return {-1}
end
if decoded['binding_hash'] ~= ARGV[1] then
  return {0}
end
redis.call('DEL', KEYS[1])
return {1, payload}
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
	expected := base64.RawURLEncoding.EncodeToString(expectedBindingHash[:])
	result, err := consumeMailboxBootstrapOAuthStateScript.Run(ctx, store.client, []string{mailboxBootstrapOAuthRedisStateKey(state)}, expected, strconv.Itoa(mailboxBootstrapOAuthStateJSONMaximum)).Result()
	if err != nil {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	values, ok := result.([]interface{})
	if !ok || len(values) < 1 {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	status, ok := mailboxOAuthRedisInteger(values[0])
	if !ok {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if status == 0 {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	if status != 1 || len(values) != 2 {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	payload, ok := values[1].(string)
	if !ok || len(payload) == 0 || len(payload) > mailboxBootstrapOAuthStateJSONMaximum {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	ownedPayload := []byte(payload)
	defer clear(ownedPayload)
	record, ok := decodeMailboxBootstrapOAuthRedisState(ownedPayload)
	if !ok {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	binding, err := base64.RawURLEncoding.DecodeString(record.BindingHash)
	if err != nil || len(binding) != sha256.Size || subtle.ConstantTimeCompare(binding, expectedBindingHash[:]) != 1 {
		clear(binding)
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	clear(binding)
	actorID, actorErr := uuid.Parse(record.ActorID)
	organizationID, organizationErr := uuid.Parse(record.OrganizationID)
	accountID, accountErr := uuid.Parse(record.AccountID)
	if actorErr != nil || organizationErr != nil || accountErr != nil {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	pending := MailboxBootstrapOAuthPending{
		Scope:                        MailboxOAuthScope{ActorID: actorID, OrganizationID: organizationID, AccountID: accountID},
		EngineAccountID:              record.EngineAccountID,
		PasswordEnrollmentGeneration: record.PasswordEnrollmentGeneration,
		CodeVerifier:                 record.CodeVerifier,
		BrowserBindingHash:           expectedBindingHash,
		ExpiresAt:                    time.Unix(record.ExpiresAtUnix, 0).UTC(),
	}
	if !validMailboxBootstrapOAuthPending(pending) {
		return MailboxBootstrapOAuthPending{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	return pending, nil
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

func decodeMailboxBootstrapOAuthRedisState(payload []byte) (mailboxBootstrapOAuthRedisStateRecord, bool) {
	if len(payload) == 0 || len(payload) > mailboxBootstrapOAuthStateJSONMaximum {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var fields map[string]json.RawMessage
	if err := decoder.Decode(&fields); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF || len(fields) != 9 {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	for _, field := range []string{`version`, `actor_id`, `organization_id`, `account_id`, `engine_account_id`, `password_enrollment_generation`, `code_verifier`, `binding_hash`, `expires_at_unix`} {
		if _, ok := fields[field]; !ok {
			return mailboxBootstrapOAuthRedisStateRecord{}, false
		}
	}
	if !mailboxBootstrapOAuthRedisFieldsExact(fields) {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	var record mailboxBootstrapOAuthRedisStateRecord
	if err := json.Unmarshal(fields[`version`], &record.Version); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`actor_id`], &record.ActorID); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`organization_id`], &record.OrganizationID); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`account_id`], &record.AccountID); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`engine_account_id`], &record.EngineAccountID); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`password_enrollment_generation`], &record.PasswordEnrollmentGeneration); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`code_verifier`], &record.CodeVerifier); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`binding_hash`], &record.BindingHash); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if err := json.Unmarshal(fields[`expires_at_unix`], &record.ExpiresAtUnix); err != nil {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	if record.Version != mailboxBootstrapOAuthStateJSONVersion || record.PasswordEnrollmentGeneration < 1 || !validMailboxOAuthBase64URL(record.CodeVerifier, 86) || !validMailboxOAuthBase64URL(record.BindingHash, 43) || record.ExpiresAtUnix < 1 {
		return mailboxBootstrapOAuthRedisStateRecord{}, false
	}
	return record, true
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
