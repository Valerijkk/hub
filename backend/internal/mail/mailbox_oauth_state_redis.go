package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	mailboxOAuthStateRedisPrefix = "hub:mail:oauth:state:"
	mailboxOAuthStateJSONMaximum = 4096
	mailboxOAuthStateJSONVersion = 1
)

// This compare-and-delete operation is intentionally one atomic script. A
// callback with the wrong browser binding cannot consume the legitimate flow.
var consumeMailboxOAuthStateScript = redis.NewScript(`
local payload = redis.call('GET', KEYS[1])
if not payload then
  return {0}
end
local decoded_ok, decoded = pcall(cjson.decode, payload)
if not decoded_ok or type(decoded) ~= 'table' or decoded['binding_hash'] ~= ARGV[1] then
  if not decoded_ok or type(decoded) ~= 'table' then
    redis.call('DEL', KEYS[1])
    return {-1}
  end
  return {0}
end
redis.call('DEL', KEYS[1])
return {1, payload}
`)

type MailboxOAuthRedisStateStore struct {
	client *redis.Client
}

type mailboxOAuthRedisStateRecord struct {
	Version              int    `json:"version"`
	ActorID              string `json:"actor_id"`
	OrganizationID       string `json:"organization_id"`
	AccountID            string `json:"account_id"`
	CredentialGeneration int64  `json:"credential_generation"`
	CodeVerifier         string `json:"code_verifier"`
	BindingHash          string `json:"binding_hash"`
	ExpiresAtUnix        int64  `json:"expires_at_unix"`
}

func NewMailboxOAuthRedisStateStore(client *redis.Client) (*MailboxOAuthRedisStateStore, error) {
	if client == nil {
		return nil, mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	return &MailboxOAuthRedisStateStore{client: client}, nil
}

func (store *MailboxOAuthRedisStateStore) SaveMailboxOAuthState(ctx context.Context, state string, pending MailboxOAuthPendingState) error {
	if store == nil || store.client == nil || ctx == nil || !validMailboxOAuthBase64URL(state, 43) || !validMailboxOAuthPendingState(pending) {
		return mailboxOAuthError(MailboxOAuthErrorInvalid)
	}
	record := mailboxOAuthRedisStateRecord{
		Version: mailboxOAuthStateJSONVersion,
		ActorID: pending.Scope.ActorID.String(), OrganizationID: pending.Scope.OrganizationID.String(), AccountID: pending.Scope.AccountID.String(),
		CredentialGeneration: pending.CredentialGeneration,
		CodeVerifier: pending.CodeVerifier,
		BindingHash: base64.RawURLEncoding.EncodeToString(pending.BrowserBindingHash[:]),
		ExpiresAtUnix: pending.ExpiresAt.Unix(),
	}
	payload, err := json.Marshal(record)
	if err != nil || len(payload) == 0 || len(payload) > mailboxOAuthStateJSONMaximum {
		clear(payload)
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	defer clear(payload)
	stored, err := store.client.SetNX(ctx, mailboxOAuthStateRedisKey(state), payload, mailboxOAuthStateTTL).Result()
	if err != nil {
		return mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if !stored {
		return mailboxOAuthError(MailboxOAuthErrorConflict)
	}
	return nil
}

func (store *MailboxOAuthRedisStateStore) ConsumeMailboxOAuthState(ctx context.Context, state string, expectedBindingHash [sha256.Size]byte) (MailboxOAuthPendingState, error) {
	if store == nil || store.client == nil || ctx == nil || !validMailboxOAuthBase64URL(state, 43) || expectedBindingHash == [sha256.Size]byte{} {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	expected := base64.RawURLEncoding.EncodeToString(expectedBindingHash[:])
	result, err := consumeMailboxOAuthStateScript.Run(ctx, store.client, []string{mailboxOAuthStateRedisKey(state)}, expected).Result()
	if err != nil {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	values, ok := result.([]interface{})
	if !ok || len(values) < 1 {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	status, ok := mailboxOAuthRedisInteger(values[0])
	if !ok {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	if status == 0 {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	if status != 1 || len(values) != 2 {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	payload, ok := values[1].(string)
	if !ok || len(payload) == 0 || len(payload) > mailboxOAuthStateJSONMaximum {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	record, ok := decodeMailboxOAuthRedisState([]byte(payload))
	if !ok {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	binding, err := base64.RawURLEncoding.DecodeString(record.BindingHash)
	if err != nil || len(binding) != sha256.Size || subtle.ConstantTimeCompare(binding, expectedBindingHash[:]) != 1 {
		clear(binding)
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorStateInvalid)
	}
	clear(binding)
	actorID, actorErr := uuid.Parse(record.ActorID)
	organizationID, organizationErr := uuid.Parse(record.OrganizationID)
	accountID, accountErr := uuid.Parse(record.AccountID)
	if actorErr != nil || organizationErr != nil || accountErr != nil {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	pending := MailboxOAuthPendingState{
		Scope: MailboxOAuthScope{ActorID: actorID, OrganizationID: organizationID, AccountID: accountID},
		CredentialGeneration: record.CredentialGeneration,
		CodeVerifier: record.CodeVerifier,
		BrowserBindingHash: expectedBindingHash,
		ExpiresAt: time.Unix(record.ExpiresAtUnix, 0).UTC(),
	}
	if !validMailboxOAuthPendingState(pending) {
		return MailboxOAuthPendingState{}, mailboxOAuthError(MailboxOAuthErrorUnavailable)
	}
	return pending, nil
}

func mailboxOAuthRedisStateKey(state string) string {
	digest := sha256.Sum256([]byte(state))
	return mailboxOAuthStateRedisPrefix + hex.EncodeToString(digest[:])
}

func decodeMailboxOAuthRedisState(payload []byte) (mailboxOAuthRedisStateRecord, bool) {
	if len(payload) == 0 || len(payload) > mailboxOAuthStateJSONMaximum {
		return mailboxOAuthRedisStateRecord{}, false
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var record mailboxOAuthRedisStateRecord
	if err := decoder.Decode(&record); err != nil {
		return mailboxOAuthRedisStateRecord{}, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return mailboxOAuthRedisStateRecord{}, false
	}
	if record.Version != mailboxOAuthStateJSONVersion || record.CredentialGeneration < 1 || !validMailboxOAuthBase64URL(record.CodeVerifier, 86) || !validMailboxOAuthBase64URL(record.BindingHash, 43) || record.ExpiresAtUnix < 1 {
		return mailboxOAuthRedisStateRecord{}, false
	}
	return record, true
}

func mailboxOAuthRedisInteger(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		return parsed, err == nil
	case []byte:
		parsed, err := strconv.ParseInt(string(typed), 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

var _ MailboxOAuthStateStore = (*MailboxOAuthRedisStateStore)(nil)
