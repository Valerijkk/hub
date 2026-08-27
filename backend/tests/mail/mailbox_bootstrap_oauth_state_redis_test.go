package mail_test

import (
	`context`
	`crypto/sha256`
	`encoding/hex`
	`encoding/json`
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
	bootstrapOAuthStatePrefix = `hub:mail:oauth:bootstrap:state:`
	normalOAuthStatePrefix    = `hub:mail:oauth:state:`
)

func TestMailboxBootstrapOAuthRedisStateStore_RoundTripConsumesExactlyOnce(t *testing.T) {
	server, store := newMailboxBootstrapOAuthRedisStateStore(t)
	state := bootstrapOAuthState()
	want := validMailboxBootstrapOAuthPending()
	if err := store.SaveMailboxBootstrapOAuthState(context.Background(), state, want); err != nil {
		t.Fatalf(`SaveMailboxBootstrapOAuthState() error = %v`, err)
	}

	bootstrapKey := bootstrapOAuthRedisStateKey(state)
	if !server.Exists(bootstrapKey) || server.Exists(state) || server.Exists(normalOAuthRedisStateKey(state)) {
		t.Fatalf(`SaveMailboxBootstrapOAuthState() stored an unexpected Redis key set: %v`, server.Keys())
	}

	got, err := store.ConsumeMailboxBootstrapOAuthState(context.Background(), state, want.BrowserBindingHash)
	if err != nil {
		t.Fatalf(`ConsumeMailboxBootstrapOAuthState() error = %v`, err)
	}
	assertMailboxBootstrapOAuthPendingEqual(t, got, want)
	if server.Exists(bootstrapKey) {
		t.Fatalf(`ConsumeMailboxBootstrapOAuthState() retained consumed key %q`, bootstrapKey)
	}

	_, err = store.ConsumeMailboxBootstrapOAuthState(context.Background(), state, want.BrowserBindingHash)
	if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`second ConsumeMailboxBootstrapOAuthState() error = %v; want state invalid`, err)
	}
	assertBootstrapOAuthStateErrorRedacted(t, err)
}

func TestMailboxBootstrapOAuthRedisStateStore_WrongBindingPreservesState(t *testing.T) {
	server, store := newMailboxBootstrapOAuthRedisStateStore(t)
	state := bootstrapOAuthState()
	want := validMailboxBootstrapOAuthPending()
	if err := store.SaveMailboxBootstrapOAuthState(context.Background(), state, want); err != nil {
		t.Fatalf(`SaveMailboxBootstrapOAuthState() error = %v`, err)
	}

	wrongBinding := sha256.Sum256([]byte(`other-browser-binding`))
	_, err := store.ConsumeMailboxBootstrapOAuthState(context.Background(), state, wrongBinding)
	if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`ConsumeMailboxBootstrapOAuthState() wrong-binding error = %v; want state invalid`, err)
	}
	assertBootstrapOAuthStateErrorRedacted(t, err)
	if !server.Exists(bootstrapOAuthRedisStateKey(state)) {
		t.Fatal(`wrong binding consumed bootstrap state`)
	}

	got, err := store.ConsumeMailboxBootstrapOAuthState(context.Background(), state, want.BrowserBindingHash)
	if err != nil {
		t.Fatalf(`ConsumeMailboxBootstrapOAuthState() correct-binding error = %v`, err)
	}
	assertMailboxBootstrapOAuthPendingEqual(t, got, want)
}

func TestMailboxBootstrapOAuthRedisStateStore_ExpiryAndConflict(t *testing.T) {
	t.Run(`expiry`, func(t *testing.T) {
		server, store := newMailboxBootstrapOAuthRedisStateStore(t)
		state := bootstrapOAuthState()
		pending := validMailboxBootstrapOAuthPending()
		if err := store.SaveMailboxBootstrapOAuthState(context.Background(), state, pending); err != nil {
			t.Fatalf(`SaveMailboxBootstrapOAuthState() error = %v`, err)
		}
		key := bootstrapOAuthRedisStateKey(state)
		ttl := server.TTL(key)
		if ttl <= 0 {
			t.Fatalf(`SaveMailboxBootstrapOAuthState() TTL = %s; want positive`, ttl)
		}
		server.FastForward(ttl + time.Second)
		_, err := store.ConsumeMailboxBootstrapOAuthState(context.Background(), state, pending.BrowserBindingHash)
		if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
			t.Fatalf(`ConsumeMailboxBootstrapOAuthState() expired error = %v; want state invalid`, err)
		}
		assertBootstrapOAuthStateErrorRedacted(t, err)
	})

	t.Run(`conflict`, func(t *testing.T) {
		_, store := newMailboxBootstrapOAuthRedisStateStore(t)
		state := bootstrapOAuthState()
		first := validMailboxBootstrapOAuthPending()
		if err := store.SaveMailboxBootstrapOAuthState(context.Background(), state, first); err != nil {
			t.Fatalf(`first SaveMailboxBootstrapOAuthState() error = %v`, err)
		}
		second := first
		second.EngineAccountID = `engine-account-bootstrap-other`
		err := store.SaveMailboxBootstrapOAuthState(context.Background(), state, second)
		if !errors.Is(err, mail.ErrMailboxOAuthConflict) {
			t.Fatalf(`second SaveMailboxBootstrapOAuthState() error = %v; want conflict`, err)
		}
		assertBootstrapOAuthStateErrorRedacted(t, err)
		got, err := store.ConsumeMailboxBootstrapOAuthState(context.Background(), state, first.BrowserBindingHash)
		if err != nil {
			t.Fatalf(`ConsumeMailboxBootstrapOAuthState() error = %v`, err)
		}
		assertMailboxBootstrapOAuthPendingEqual(t, got, first)
	})
}

func TestMailboxBootstrapOAuthRedisStateStore_DeletesMalformedVersionAndOversizePayloads(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*testing.T, *miniredis.Miniredis, string)
	}{
		{name: `malformed`, mutate: func(t *testing.T, server *miniredis.Miniredis, key string) {
			t.Helper()
			if err := server.Set(key, `{`); err != nil {
				t.Fatalf(`set malformed record: %v`, err)
			}
		}},
		{name: `trailing data`, mutate: func(t *testing.T, server *miniredis.Miniredis, key string) {
			t.Helper()
			payload, err := server.Get(key)
			if err != nil {
				t.Fatalf(`read bootstrap Redis record: %v`, err)
			}
			if err := server.Set(key, payload+`x`); err != nil {
				t.Fatalf(`set trailing record: %v`, err)
			}
		}},
		{name: `unknown member`, mutate: func(t *testing.T, server *miniredis.Miniredis, key string) {
			t.Helper()
			record := bootstrapOAuthRedisRecord(t, server, key)
			record[`unexpected`] = true
			setBootstrapOAuthRedisRecord(t, server, key, record)
		}},
		{name: `version`, mutate: func(t *testing.T, server *miniredis.Miniredis, key string) {
			t.Helper()
			record := bootstrapOAuthRedisRecord(t, server, key)
			record[`version`] = 2
			setBootstrapOAuthRedisRecord(t, server, key, record)
		}},
		{name: `oversize`, mutate: func(t *testing.T, server *miniredis.Miniredis, key string) {
			t.Helper()
			if err := server.Set(key, strings.Repeat(`x`, 4097)); err != nil {
				t.Fatalf(`set oversize record: %v`, err)
			}
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			server, store := newMailboxBootstrapOAuthRedisStateStore(t)
			state := bootstrapOAuthState()
			pending := validMailboxBootstrapOAuthPending()
			if err := store.SaveMailboxBootstrapOAuthState(context.Background(), state, pending); err != nil {
				t.Fatalf(`SaveMailboxBootstrapOAuthState() error = %v`, err)
			}
			key := bootstrapOAuthRedisStateKey(state)
			testCase.mutate(t, server, key)

			_, err := store.ConsumeMailboxBootstrapOAuthState(context.Background(), state, pending.BrowserBindingHash)
			if !errors.Is(err, mail.ErrMailboxOAuthUnavailable) {
				t.Fatalf(`ConsumeMailboxBootstrapOAuthState() error = %v; want unavailable`, err)
			}
			assertBootstrapOAuthStateErrorRedacted(t, err)
			if server.Exists(key) {
				t.Fatalf(`ConsumeMailboxBootstrapOAuthState() retained invalid key %q`, key)
			}
		})
	}
}

func TestMailboxBootstrapOAuthRedisStateStore_SeparatesNormalPrefixAndRawState(t *testing.T) {
	server, store := newMailboxBootstrapOAuthRedisStateStore(t)
	state := bootstrapOAuthState()
	pending := validMailboxBootstrapOAuthPending()
	if err := store.SaveMailboxBootstrapOAuthState(context.Background(), state, pending); err != nil {
		t.Fatalf(`SaveMailboxBootstrapOAuthState() error = %v`, err)
	}
	bootstrapKey := bootstrapOAuthRedisStateKey(state)
	payload, err := server.Get(bootstrapKey)
	if err != nil {
		t.Fatalf(`read bootstrap payload: %v`, err)
	}
	if strings.Contains(payload, state) || server.Exists(state) {
		t.Fatal(`bootstrap state was stored raw`)
	}

	normalState := strings.Repeat(`U`, 43)
	normalKey := normalOAuthRedisStateKey(normalState)
	if err := server.Set(normalKey, payload); err != nil {
		t.Fatalf(`seed normal OAuth prefix: %v`, err)
	}
	_, err = store.ConsumeMailboxBootstrapOAuthState(context.Background(), normalState, pending.BrowserBindingHash)
	if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`ConsumeMailboxBootstrapOAuthState() normal-prefix error = %v; want state invalid`, err)
	}
	assertBootstrapOAuthStateErrorRedacted(t, err)
	if !server.Exists(normalKey) {
		t.Fatal(`bootstrap store consumed normal OAuth prefix`)
	}
}

func TestMailboxBootstrapOAuthRedisStateStore_InvalidInputDoesNotTouchRedis(t *testing.T) {
	server, store := newMailboxBootstrapOAuthRedisStateStore(t)
	pending := validMailboxBootstrapOAuthPending()
	before := server.CommandCount()
	if err := store.SaveMailboxBootstrapOAuthState(context.Background(), `invalid`, pending); !errors.Is(err, mail.ErrMailboxOAuthInvalid) {
		t.Fatalf(`SaveMailboxBootstrapOAuthState() invalid-state error = %v; want invalid`, err)
	} else {
		assertBootstrapOAuthStateErrorRedacted(t, err)
	}

	invalidPending := pending
	invalidPending.PasswordEnrollmentGeneration = 0
	if err := store.SaveMailboxBootstrapOAuthState(context.Background(), bootstrapOAuthState(), invalidPending); !errors.Is(err, mail.ErrMailboxOAuthInvalid) {
		t.Fatalf(`SaveMailboxBootstrapOAuthState() invalid-pending error = %v; want invalid`, err)
	} else {
		assertBootstrapOAuthStateErrorRedacted(t, err)
	}

	_, err := store.ConsumeMailboxBootstrapOAuthState(context.Background(), `invalid`, pending.BrowserBindingHash)
	if !errors.Is(err, mail.ErrMailboxOAuthStateInvalid) {
		t.Fatalf(`ConsumeMailboxBootstrapOAuthState() invalid-state error = %v; want state invalid`, err)
	}
	assertBootstrapOAuthStateErrorRedacted(t, err)
	if got := server.CommandCount(); got != before {
		t.Fatalf(`invalid calls issued %d Redis commands; want %d`, got, before)
	}
}

func newMailboxBootstrapOAuthRedisStateStore(t *testing.T) (*miniredis.Miniredis, *mail.MailboxBootstrapOAuthRedisStateStore) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	store, err := mail.NewMailboxBootstrapOAuthRedisStateStore(client)
	if err != nil {
		t.Fatalf(`NewMailboxBootstrapOAuthRedisStateStore() error = %v`, err)
	}
	return server, store
}

func validMailboxBootstrapOAuthPending() mail.MailboxBootstrapOAuthPending {
	binding := sha256.Sum256([]byte(`bootstrap-browser-binding`))
	return mail.MailboxBootstrapOAuthPending{
		Scope: mail.MailboxOAuthScope{
			ActorID:        uuid.MustParse(`00000000-0000-0000-0000-000000000001`),
			OrganizationID: uuid.MustParse(`00000000-0000-0000-0000-000000000002`),
			AccountID:      uuid.MustParse(`00000000-0000-0000-0000-000000000003`),
		},
		EngineAccountID:             `engine-account-bootstrap`,
		PasswordEnrollmentGeneration: 1,
		CodeVerifier:                 strings.Repeat(`A`, 86),
		BrowserBindingHash:           binding,
		ExpiresAt:                    time.Date(2026, time.August, 27, 12, 0, 0, 0, time.UTC),
	}
}

func bootstrapOAuthState() string { return strings.Repeat(`Q`, 43) }

func bootstrapOAuthRedisStateKey(state string) string {
	digest := sha256.Sum256([]byte(state))
	return bootstrapOAuthStatePrefix + hex.EncodeToString(digest[:])
}

func normalOAuthRedisStateKey(state string) string {
	digest := sha256.Sum256([]byte(state))
	return normalOAuthStatePrefix + hex.EncodeToString(digest[:])
}

func bootstrapOAuthRedisRecord(t *testing.T, server *miniredis.Miniredis, key string) map[string]any {
	t.Helper()
	payload, err := server.Get(key)
	if err != nil {
		t.Fatalf(`read bootstrap Redis record: %v`, err)
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		t.Fatalf(`decode bootstrap Redis record: %v`, err)
	}
	return record
}

func setBootstrapOAuthRedisRecord(t *testing.T, server *miniredis.Miniredis, key string, record map[string]any) {
	t.Helper()
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatalf(`encode bootstrap Redis record: %v`, err)
	}
	if err := server.Set(key, string(payload)); err != nil {
		t.Fatalf(`set bootstrap Redis record: %v`, err)
	}
}

func assertMailboxBootstrapOAuthPendingEqual(t *testing.T, got, want mail.MailboxBootstrapOAuthPending) {
	t.Helper()
	if got.Scope != want.Scope || got.EngineAccountID != want.EngineAccountID || got.PasswordEnrollmentGeneration != want.PasswordEnrollmentGeneration || got.CodeVerifier != want.CodeVerifier || got.BrowserBindingHash != want.BrowserBindingHash || !got.ExpiresAt.Equal(want.ExpiresAt) || got.ExpiresAt.Location() != time.UTC || got.ExpiresAt.Nanosecond() != 0 {
		t.Fatalf(`pending = %#v; want exact scope, engine, generation, verifier, binding and UTC expiry`, got)
	}
}

func assertBootstrapOAuthStateErrorRedacted(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal(`missing bootstrap OAuth state error`)
	}
	for _, forbidden := range []string{`engine-account-bootstrap`, `bootstrap-browser-binding`, strings.Repeat(`Q`, 43), strings.Repeat(`A`, 86), `token`, `password`, `provider`} {
		if strings.Contains(strings.ToLower(err.Error()), strings.ToLower(forbidden)) {
			t.Fatalf(`bootstrap OAuth state error exposed sensitive detail: %q`, err)
		}
	}
}
