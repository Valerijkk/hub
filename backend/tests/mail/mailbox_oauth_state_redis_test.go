package mail_test

import (
	`context`
	`crypto/sha256`
	`encoding/hex`
	`errors`
	`strings`
	`testing`
	`time`

	`github.com/alicebob/miniredis/v2`
	`github.com/google/uuid`
	`github.com/redis/go-redis/v9`

	`github.com/valerii/hub/backend/internal/mail`
)

const (
	mailboxOAuthStateRedisPrefixForTest          = `hub:mail:oauth:state:`
	mailboxBootstrapOAuthStateRedisPrefixForTest = `hub:mail:oauth:bootstrap:state:`
)

func TestMailboxOAuthRedisStateStore_HashedKeyRoundTripTTLAndReplay(t *testing.T) {
	server, store := newMailboxOAuthRedisStateStoreForTest(t)
	state := mailboxOAuthStateForTest()
	want := validMailboxOAuthPendingStateForTest()
	if err := store.SaveMailboxOAuthState(context.Background(), state, want); err != nil {
		t.Fatalf(`SaveMailboxOAuthState() error = %v`, err)
	}

	key := mailboxOAuthRedisStateKeyForTest(state)
	bootstrapKey := mailboxBootstrapOAuthRedisStateKeyForTest(state)
	if !server.Exists(key) || server.Exists(state) || server.Exists(bootstrapKey) {
		t.Fatalf(`SaveMailboxOAuthState() stored an unexpected Redis key set: %v`, server.Keys())
	}
	payload, err := server.Get(key)
	if err != nil {
		t.Fatalf(`read OAuth Redis payload: %v`, err)
	}
	if strings.Contains(payload, state) {
		t.Fatal(`SaveMailboxOAuthState() stored raw state in the Redis payload`)
	}
	if ttl := server.TTL(key); ttl != 10*time.Minute {
		t.Fatalf(`SaveMailboxOAuthState() TTL = %s, want %s`, ttl, 10*time.Minute)
	}

	got, err := store.ConsumeMailboxOAuthState(context.Background(), state, want.BrowserBindingHash)
	if err != nil {
		t.Fatalf(`ConsumeMailboxOAuthState() error = %v`, err)
	}
	assertMailboxOAuthPendingStateForTest(t, got, want)
	if server.Exists(key) {
		t.Fatalf(`ConsumeMailboxOAuthState() retained consumed key %q`, key)
	}

	_, err = store.ConsumeMailboxOAuthState(context.Background(), state, want.BrowserBindingHash)
	if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`second ConsumeMailboxOAuthState() error = %v, want state invalid`, err)
	}
}

func TestMailboxOAuthRedisStateStore_ExpiredStateIsRejectedAndRemoved(t *testing.T) {
	server, store := newMailboxOAuthRedisStateStoreForTest(t)
	state := mailboxOAuthStateForTest()
	want := validMailboxOAuthPendingStateForTest()
	if err := store.SaveMailboxOAuthState(context.Background(), state, want); err != nil {
		t.Fatalf(`SaveMailboxOAuthState() error = %v`, err)
	}

	key := mailboxOAuthRedisStateKeyForTest(state)
	ttl := server.TTL(key)
	if ttl != 10*time.Minute {
		t.Fatalf(`SaveMailboxOAuthState() TTL = %s, want %s`, ttl, 10*time.Minute)
	}
	if ttl <= 0 {
		t.Fatalf(`SaveMailboxOAuthState() TTL = %s, want positive`, ttl)
	}
	server.FastForward(ttl + time.Second)

	_, err := store.ConsumeMailboxOAuthState(context.Background(), state, want.BrowserBindingHash)
	if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`ConsumeMailboxOAuthState() expired error = %v, want state invalid`, err)
	}
	if server.Exists(key) {
		t.Fatalf(`ConsumeMailboxOAuthState() retained expired key %q`, key)
	}
}

func TestMailboxOAuthRedisStateStore_WrongBindingAndConflictPreserveOriginal(t *testing.T) {
	server, store := newMailboxOAuthRedisStateStoreForTest(t)
	state := mailboxOAuthStateForTest()
	want := validMailboxOAuthPendingStateForTest()
	if err := store.SaveMailboxOAuthState(context.Background(), state, want); err != nil {
		t.Fatalf(`SaveMailboxOAuthState() error = %v`, err)
	}

	key := mailboxOAuthRedisStateKeyForTest(state)
	originalPayload, err := server.Get(key)
	if err != nil {
		t.Fatalf(`read original OAuth Redis payload: %v`, err)
	}
	wrongBinding := sha256.Sum256([]byte(`other-browser-binding`))
	_, err = store.ConsumeMailboxOAuthState(context.Background(), state, wrongBinding)
	if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`ConsumeMailboxOAuthState() wrong-binding error = %v, want state invalid`, err)
	}
	if !server.Exists(key) {
		t.Fatal(`wrong binding consumed OAuth state`)
	}

	replacement := want
	replacement.CredentialGeneration++
	err = store.SaveMailboxOAuthState(context.Background(), state, replacement)
	if !errors.Is(err, mail.ErrMailboxOAuthConflict) {
		t.Fatalf(`second SaveMailboxOAuthState() error = %v, want conflict`, err)
	}
	afterConflictPayload, err := server.Get(key)
	if err != nil {
		t.Fatalf(`read OAuth Redis payload after conflict: %v`, err)
	}
	if afterConflictPayload != originalPayload {
		t.Fatal(`duplicate SaveMailboxOAuthState() replaced original state`)
	}

	got, err := store.ConsumeMailboxOAuthState(context.Background(), state, want.BrowserBindingHash)
	if err != nil {
		t.Fatalf(`ConsumeMailboxOAuthState() correct-binding error = %v`, err)
	}
	assertMailboxOAuthPendingStateForTest(t, got, want)
	if server.Exists(key) {
		t.Fatalf(`ConsumeMailboxOAuthState() retained consumed key %q`, key)
	}
}

func newMailboxOAuthRedisStateStoreForTest(t *testing.T) (*miniredis.Miniredis, *mail.MailboxOAuthRedisStateStore) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store, err := mail.NewMailboxOAuthRedisStateStore(client)
	if err != nil {
		t.Fatalf(`NewMailboxOAuthRedisStateStore() error = %v`, err)
	}
	return server, store
}

func mailboxOAuthStateForTest() string {
	return strings.Repeat(`Q`, 43)
}

func validMailboxOAuthPendingStateForTest() mail.MailboxOAuthPendingState {
	binding := sha256.Sum256([]byte(`browser-binding`))
	return mail.MailboxOAuthPendingState{
		Scope: mail.MailboxOAuthScope{
			ActorID:        uuid.MustParse(`00000000-0000-0000-0000-000000000001`),
			OrganizationID: uuid.MustParse(`00000000-0000-0000-0000-000000000002`),
			AccountID:      uuid.MustParse(`00000000-0000-0000-0000-000000000003`),
		},
		CredentialGeneration: 1,
		CodeVerifier:         strings.Repeat(`A`, 86),
		BrowserBindingHash:   binding,
		ExpiresAt:            time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC),
	}
}

func mailboxOAuthRedisStateKeyForTest(state string) string {
	digest := sha256.Sum256([]byte(state))
	return mailboxOAuthStateRedisPrefixForTest + hex.EncodeToString(digest[:])
}

func mailboxBootstrapOAuthRedisStateKeyForTest(state string) string {
	digest := sha256.Sum256([]byte(state))
	return mailboxBootstrapOAuthStateRedisPrefixForTest + hex.EncodeToString(digest[:])
}

func assertMailboxOAuthPendingStateForTest(t *testing.T, got, want mail.MailboxOAuthPendingState) {
	t.Helper()
	if got.Scope != want.Scope || got.CredentialGeneration != want.CredentialGeneration || got.CodeVerifier != want.CodeVerifier || got.BrowserBindingHash != want.BrowserBindingHash || !got.ExpiresAt.Equal(want.ExpiresAt) || got.ExpiresAt.Location() != time.UTC || got.ExpiresAt.Nanosecond() != 0 {
		t.Fatalf(`pending = %#v, want exact scope, generation, verifier, binding and UTC expiry`, got)
	}
}
