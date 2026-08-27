//go:build hubtest

package calls_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

var fixedJitsiIssueTime = time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)

type jitsiCredentialIssuer interface {
	Issue(*ent.Call, uuid.UUID, string) (string, error)
}

type jitsiTokenClaims struct {
	jwt.RegisteredClaims
	Room    string             `json:"room"`
	Context jitsiContextClaims `json:"context"`
}

type jitsiContextClaims struct {
	User jitsiUserClaims `json:"user"`
}

type jitsiUserClaims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestJitsiCredentialIssuerConstructorRejectsUnsafeOrIncompleteConfig(t *testing.T) {
	tests := []struct {
		name     string
		errorKey string
		mutate   func(*config.JitsiConfig)
	}{
		{
			name:     "disabled",
			errorKey: "JITSI_ENABLED",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.Enabled = false
			},
		},
		{
			name:     "missing host",
			errorKey: "JITSI_HOST",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.Host = ""
			},
		},
		{
			name:     "missing public URL",
			errorKey: "JITSI_PUBLIC_URL",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.PublicURL = ""
			},
		},
		{
			name:     "missing app ID",
			errorKey: "JITSI_APP_ID",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.AppID = ""
			},
		},
		{
			name:     "missing app secret",
			errorKey: "JITSI_APP_SECRET",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.AppSecret = ""
			},
		},
		{
			name:     "missing XMPP domain",
			errorKey: "JITSI_XMPP_DOMAIN",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.XMPPDomain = ""
			},
		},
		{
			name:     "HTTP public URL",
			errorKey: "JITSI_PUBLIC_URL",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.PublicURL = "http://meet.hub.test"
			},
		},
		{
			name:     "host and URL drift",
			errorKey: "JITSI_PUBLIC_URL",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.PublicURL = "https://other.hub.test"
			},
		},
		{
			name:     "external meet.jit.si",
			errorKey: "JITSI_HOST",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.Host = "meet.jit.si"
				cfg.PublicURL = "https://meet.jit.si"
			},
		},
		{
			name:     "short secret",
			errorKey: "JITSI_APP_SECRET",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.AppSecret = strings.Repeat("s", 31)
			},
		},
		{
			name:     "required marker",
			errorKey: "JITSI_APP_SECRET",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.AppSecret = "__REQUIRED_UNUSABLE_NO_DEFAULT__"
			},
		},
		{
			name:     "known dev secret",
			errorKey: "JITSI_APP_SECRET",
			mutate: func(cfg *config.JitsiConfig) {
				cfg.AppSecret = "devsecret"
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := validJitsiIssuerConfig()
			testCase.mutate(&cfg)

			_, err := calls.ExportNewJitsiCredentialIssuerForTest(cfg, func() time.Time {
				return fixedJitsiIssueTime
			})

			requireCredentialSafeError(t, err, cfg.AppSecret)
			assert.Contains(t, err.Error(), testCase.errorKey)
		})
	}
}

func TestJitsiCredentialIssuerRejectsInvalidServerInputsWithoutReturningToken(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	issuer := newJitsiCredentialIssuer(t, cfg)
	validCall := activeJitsiCall()
	validUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")

	tests := []struct {
		name     string
		call     *ent.Call
		userID   uuid.UUID
		userName string
	}{
		{name: "nil call", call: nil, userID: validUserID, userName: "Ada Lovelace"},
		{
			name: "inactive call",
			call: &ent.Call{
				ID:       validCall.ID,
				RoomName: validCall.RoomName,
				Status:   entcall.StatusEnded,
			},
			userID:   validUserID,
			userName: "Ada Lovelace",
		},
		{
			name: "zero call ID",
			call: &ent.Call{
				ID:       uuid.Nil,
				RoomName: validCall.RoomName,
				Status:   entcall.StatusActive,
			},
			userID:   validUserID,
			userName: "Ada Lovelace",
		},
		{
			name: "blank server room name",
			call: &ent.Call{
				ID:       validCall.ID,
				RoomName: "",
				Status:   entcall.StatusActive,
			},
			userID:   validUserID,
			userName: "Ada Lovelace",
		},
		{
			name: "whitespace-only server room name",
			call: &ent.Call{
				ID:       validCall.ID,
				RoomName: " \t\r\n ",
				Status:   entcall.StatusActive,
			},
			userID:   validUserID,
			userName: "Ada Lovelace",
		},
		{
			name: "padded server room name",
			call: &ent.Call{
				ID:       validCall.ID,
				RoomName: " hub-room-42 ",
				Status:   entcall.StatusActive,
			},
			userID:   validUserID,
			userName: "Ada Lovelace",
		},
		{name: "zero authenticated user", call: validCall, userID: uuid.Nil, userName: "Ada Lovelace"},
		{name: "blank server-known name", call: validCall, userID: validUserID, userName: " \t\r\n "},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rawToken, err := issuer.Issue(testCase.call, testCase.userID, testCase.userName)

			requireCredentialSafeError(t, err, cfg.AppSecret)
			if rawToken != "" {
				t.Fatal("issuer returned a credential for invalid server input")
			}
		})
	}
}

func TestJitsiCredentialIssuerBindsClaimsToExactCallAndAuthenticatedUser(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	issuer := newJitsiCredentialIssuer(t, cfg)
	callEntity := activeJitsiCall()
	callEntity.RoomName = "hub-call-room-42"
	callEntity.CreatedBy = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	authenticatedUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")

	rawToken, err := issuer.Issue(callEntity, authenticatedUserID, "Ada Lovelace")
	require.NoError(t, err)
	if rawToken == "" {
		t.Fatal("issuer returned an empty credential")
	}

	parsedToken, claims, err := inspectJitsiIssuerToken(
		rawToken,
		cfg.AppSecret,
		fixedJitsiIssueTime.Add(time.Minute),
		cfg,
		callEntity.RoomName,
		authenticatedUserID,
		"Ada Lovelace",
	)
	require.NoError(t, err)
	require.True(t, parsedToken.Valid)
	assert.Equal(t, jwt.SigningMethodHS256.Alg(), parsedToken.Method.Alg())
	assert.Equal(t, "hub-jitsi", claims.Issuer)
	assert.Equal(t, jwt.ClaimStrings{"jitsi"}, claims.Audience)
	assert.Equal(t, "xmpp.hub.internal", claims.Subject)
	assert.Equal(t, "hub-call-room-42", claims.Room)
	assert.Equal(t, callEntity.RoomName, claims.Room)
	assert.NotEqual(t, callEntity.ID.String(), claims.Room)
	assert.Equal(t, authenticatedUserID.String(), claims.Context.User.ID)
	assert.NotEqual(t, callEntity.CreatedBy.String(), claims.Context.User.ID)
	assert.Equal(t, "Ada Lovelace", claims.Context.User.Name)
	require.NotNil(t, claims.IssuedAt)
	require.NotNil(t, claims.ExpiresAt)
	assert.True(t, claims.IssuedAt.Time.Equal(fixedJitsiIssueTime))
	assert.True(t, claims.ExpiresAt.Time.Equal(fixedJitsiIssueTime.Add(5*time.Minute)))
	assert.Equal(t, 5*time.Minute, claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time))
}

func TestJitsiCredentialIssuerOmitsPrivilegeGuestWildcardAndCredentialClaims(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	issuer := newJitsiCredentialIssuer(t, cfg)

	rawToken, err := issuer.Issue(
		activeJitsiCall(),
		uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		"Ada Lovelace",
	)
	require.NoError(t, err)
	if rawToken == "" {
		t.Fatal("issuer returned an empty credential")
	}

	claims, err := inspectJitsiIssuerMapClaims(rawToken, cfg.AppSecret, fixedJitsiIssueTime.Add(time.Minute), cfg)
	require.NoError(t, err)
	requireExactClaimKeys(t, map[string]interface{}(claims), "iss", "aud", "sub", "iat", "exp", "room", "context")

	audience, ok := claims["aud"].([]interface{})
	require.True(t, ok, "aud must be a JSON array")
	assert.Equal(t, []interface{}{"jitsi"}, audience)
	assert.Equal(t, activeJitsiCall().RoomName, claims["room"])
	assert.NotEqual(t, "*", claims["room"])

	for _, forbiddenClaim := range []string{"moderator", "features", "guest", "wildcard", "token", "secret", "*"} {
		if _, exists := claims[forbiddenClaim]; exists {
			t.Fatalf("forbidden claim %q is present", forbiddenClaim)
		}
	}

	contextClaims, ok := claims["context"].(map[string]interface{})
	require.True(t, ok)
	requireExactClaimKeys(t, contextClaims, "user")
	userClaims, ok := contextClaims["user"].(map[string]interface{})
	require.True(t, ok)
	requireExactClaimKeys(t, userClaims, "id", "name")
	assert.Equal(t, "22222222-2222-4222-8222-222222222222", userClaims["id"])
	assert.Equal(t, "Ada Lovelace", userClaims["name"])
}

// TestJitsiIssuerInspectionRejectsWrongSecretAlgorithmExpiryAndLifetime uses
// synthetic JWTs only to self-check test inspection of issuer output. Phase 150
// product code does not accept or verify incoming Jitsi JWTs.
func TestJitsiIssuerInspectionRejectsWrongSecretAlgorithmExpiryAndLifetime(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	issuer := newJitsiCredentialIssuer(t, cfg)
	issuedToken, err := issuer.Issue(
		activeJitsiCall(),
		uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		"Ada Lovelace",
	)
	require.NoError(t, err)

	h384Token := signJitsiClaimsForTest(
		t,
		jwt.SigningMethodHS384,
		jitsiMapClaims(cfg, fixedJitsiIssueTime, 5*time.Minute),
		[]byte(cfg.AppSecret),
	)
	noneToken := signJitsiClaimsForTest(
		t,
		jwt.SigningMethodNone,
		jitsiMapClaims(cfg, fixedJitsiIssueTime, 5*time.Minute),
		jwt.UnsafeAllowNoneSignatureType,
	)
	nonFiveMinuteToken := signJitsiClaimsForTest(
		t,
		jwt.SigningMethodHS256,
		jitsiMapClaims(cfg, fixedJitsiIssueTime, 10*time.Minute),
		[]byte(cfg.AppSecret),
	)

	tests := []struct {
		name   string
		raw    string
		secret string
		now    time.Time
	}{
		{
			name:   "wrong secret",
			raw:    issuedToken,
			secret: "wrong-verifier-secret-0123456789",
			now:    fixedJitsiIssueTime.Add(time.Minute),
		},
		{
			name:   "HS384",
			raw:    h384Token,
			secret: cfg.AppSecret,
			now:    fixedJitsiIssueTime.Add(time.Minute),
		},
		{
			name:   "none",
			raw:    noneToken,
			secret: cfg.AppSecret,
			now:    fixedJitsiIssueTime.Add(time.Minute),
		},
		{
			name:   "expired",
			raw:    issuedToken,
			secret: cfg.AppSecret,
			now:    fixedJitsiIssueTime.Add(5*time.Minute + time.Second),
		},
		{
			name:   "non-five-minute lifetime",
			raw:    nonFiveMinuteToken,
			secret: cfg.AppSecret,
			now:    fixedJitsiIssueTime.Add(time.Minute),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := inspectJitsiIssuerToken(
				testCase.raw,
				testCase.secret,
				testCase.now,
				cfg,
				activeJitsiCall().RoomName,
				uuid.MustParse("22222222-2222-4222-8222-222222222222"),
				"Ada Lovelace",
			)
			requireCredentialSafeError(t, err, testCase.raw, testCase.secret, cfg.AppSecret)
		})
	}
}

// Synthetic claim mutations self-check only the test inspector; Phase 150
// product code does not accept or verify incoming Jitsi JWTs.
func TestJitsiIssuerInspectionRejectsWrongOrMissingRequiredClaims(t *testing.T) {
	cfg := validJitsiIssuerConfig()
	expectedRoom := activeJitsiCall().RoomName
	expectedUserID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	expectedUserName := "Ada Lovelace"
	tests := []struct {
		name   string
		mutate func(jwt.MapClaims)
	}{
		{
			name: "wrong issuer",
			mutate: func(claims jwt.MapClaims) {
				claims["iss"] = "wrong-issuer"
			},
		},
		{
			name: "missing issuer",
			mutate: func(claims jwt.MapClaims) {
				delete(claims, "iss")
			},
		},
		{
			name: "wrong audience",
			mutate: func(claims jwt.MapClaims) {
				claims["aud"] = []string{"wrong-audience"}
			},
		},
		{
			name: "missing audience",
			mutate: func(claims jwt.MapClaims) {
				delete(claims, "aud")
			},
		},
		{
			name: "wrong subject",
			mutate: func(claims jwt.MapClaims) {
				claims["sub"] = "wrong-xmpp-domain"
			},
		},
		{
			name: "missing subject",
			mutate: func(claims jwt.MapClaims) {
				delete(claims, "sub")
			},
		},
		{
			name: "wrong room",
			mutate: func(claims jwt.MapClaims) {
				claims["room"] = "99999999-9999-4999-8999-999999999999"
			},
		},
		{
			name: "missing room",
			mutate: func(claims jwt.MapClaims) {
				delete(claims, "room")
			},
		},
		{
			name: "missing iat",
			mutate: func(claims jwt.MapClaims) {
				delete(claims, "iat")
			},
		},
		{
			name: "missing exp",
			mutate: func(claims jwt.MapClaims) {
				delete(claims, "exp")
			},
		},
		{
			name: "missing context",
			mutate: func(claims jwt.MapClaims) {
				delete(claims, "context")
			},
		},
		{
			name: "wrong context",
			mutate: func(claims jwt.MapClaims) {
				claims["context"] = map[string]interface{}{"unexpected": "value"}
			},
		},
		{
			name: "nonobject context",
			mutate: func(claims jwt.MapClaims) {
				claims["context"] = "not-an-object"
			},
		},
		{
			name: "missing context user",
			mutate: func(claims jwt.MapClaims) {
				claims["context"] = map[string]interface{}{}
			},
		},
		{
			name: "wrong context user",
			mutate: func(claims jwt.MapClaims) {
				claims["context"] = map[string]interface{}{
					"user": map[string]interface{}{"unexpected": "value"},
				}
			},
		},
		{
			name: "nonobject context user",
			mutate: func(claims jwt.MapClaims) {
				claims["context"] = map[string]interface{}{"user": "not-an-object"}
			},
		},
		{
			name: "missing user id",
			mutate: func(claims jwt.MapClaims) {
				delete(jitsiMapUserClaims(claims), "id")
			},
		},
		{
			name: "wrong user id",
			mutate: func(claims jwt.MapClaims) {
				jitsiMapUserClaims(claims)["id"] = "33333333-3333-4333-8333-333333333333"
			},
		},
		{
			name: "blank user id",
			mutate: func(claims jwt.MapClaims) {
				jitsiMapUserClaims(claims)["id"] = "   "
			},
		},
		{
			name: "non-string user id",
			mutate: func(claims jwt.MapClaims) {
				jitsiMapUserClaims(claims)["id"] = 42
			},
		},
		{
			name: "missing user name",
			mutate: func(claims jwt.MapClaims) {
				delete(jitsiMapUserClaims(claims), "name")
			},
		},
		{
			name: "wrong user name",
			mutate: func(claims jwt.MapClaims) {
				jitsiMapUserClaims(claims)["name"] = "Mallory"
			},
		},
		{
			name: "blank user name",
			mutate: func(claims jwt.MapClaims) {
				jitsiMapUserClaims(claims)["name"] = "   "
			},
		},
		{
			name: "non-string user name",
			mutate: func(claims jwt.MapClaims) {
				jitsiMapUserClaims(claims)["name"] = true
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			claims := jitsiMapClaims(cfg, fixedJitsiIssueTime, 5*time.Minute)
			testCase.mutate(claims)
			rawToken := signJitsiClaimsForTest(t, jwt.SigningMethodHS256, claims, []byte(cfg.AppSecret))

			_, _, err := inspectJitsiIssuerToken(
				rawToken,
				cfg.AppSecret,
				fixedJitsiIssueTime.Add(time.Minute),
				cfg,
				expectedRoom,
				expectedUserID,
				expectedUserName,
			)
			requireCredentialSafeError(t, err, rawToken, cfg.AppSecret)
		})
	}
}

func newJitsiCredentialIssuer(t *testing.T, cfg config.JitsiConfig) jitsiCredentialIssuer {
	t.Helper()
	issuer, err := calls.ExportNewJitsiCredentialIssuerForTest(cfg, func() time.Time {
		return fixedJitsiIssueTime
	})
	require.NoError(t, err)
	return issuer
}

func validJitsiIssuerConfig() config.JitsiConfig {
	return config.JitsiConfig{
		Enabled:    true,
		Host:       "meet.hub.test",
		PublicURL:  "https://meet.hub.test",
		AppID:      "hub-jitsi",
		AppSecret:  "0123456789abcdef0123456789abcdef",
		XMPPDomain: "xmpp.hub.internal",
	}
}

func activeJitsiCall() *ent.Call {
	return &ent.Call{
		ID:        uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		RoomName:  "legacy-livekit-room",
		Status:    entcall.StatusActive,
		CreatedBy: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
}

// inspectJitsiIssuerToken validates test inspection of issuer output only.
// Phase 150 product code does not accept or verify incoming Jitsi JWTs.
func inspectJitsiIssuerToken(
	rawToken string,
	secret string,
	now time.Time,
	cfg config.JitsiConfig,
	expectedRoom string,
	expectedUserID uuid.UUID,
	expectedUserName string,
) (*jwt.Token, *jitsiTokenClaims, error) {
	claims := &jitsiTokenClaims{}
	parser := jitsiIssuerInspectionParser(now, cfg)
	parsedToken, err := parser.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected Jitsi signing algorithm")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return parsedToken, nil, err
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return parsedToken, nil, errors.New("Jitsi credential requires iat and exp")
	}
	if claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) != 5*time.Minute {
		return parsedToken, nil, errors.New("Jitsi credential lifetime must equal five minutes")
	}
	if claims.Subject != cfg.XMPPDomain {
		return parsedToken, nil, errors.New("Jitsi credential subject does not match configured XMPP domain")
	}
	if claims.Room != expectedRoom {
		return parsedToken, nil, errors.New("Jitsi credential room does not match server-loaded call")
	}
	if claims.Context.User.ID != expectedUserID.String() {
		return parsedToken, nil, errors.New("Jitsi credential user id does not match authenticated user")
	}
	if claims.Context.User.Name != expectedUserName {
		return parsedToken, nil, errors.New("Jitsi credential user name does not match server-known name")
	}
	return parsedToken, claims, nil
}

func inspectJitsiIssuerMapClaims(
	rawToken string,
	secret string,
	now time.Time,
	cfg config.JitsiConfig,
) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	parsedToken, err := jitsiIssuerInspectionParser(now, cfg).ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected Jitsi signing algorithm")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, errors.New("invalid Jitsi credential")
	}
	return claims, nil
}

func jitsiIssuerInspectionParser(now time.Time, cfg config.JitsiConfig) *jwt.Parser {
	return jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(cfg.AppID),
		jwt.WithAudience("jitsi"),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
}

func jitsiMapClaims(cfg config.JitsiConfig, issuedAt time.Time, lifetime time.Duration) jwt.MapClaims {
	return jwt.MapClaims{
		"iss":  cfg.AppID,
		"aud":  []string{"jitsi"},
		"sub":  cfg.XMPPDomain,
		"iat":  issuedAt.Unix(),
		"exp":  issuedAt.Add(lifetime).Unix(),
		"room": activeJitsiCall().RoomName,
		"context": map[string]interface{}{
			"user": map[string]interface{}{
				"id":   "22222222-2222-4222-8222-222222222222",
				"name": "Ada Lovelace",
			},
		},
	}
}

func signJitsiClaimsForTest(
	t *testing.T,
	method jwt.SigningMethod,
	claims jwt.Claims,
	key interface{},
) string {
	t.Helper()
	rawToken, err := jwt.NewWithClaims(method, claims).SignedString(key)
	require.NoError(t, err)
	return rawToken
}

func jitsiMapUserClaims(claims jwt.MapClaims) map[string]interface{} {
	contextClaims := claims["context"].(map[string]interface{})
	return contextClaims["user"].(map[string]interface{})
}

func requireExactClaimKeys(t *testing.T, claims map[string]interface{}, expected ...string) {
	t.Helper()
	if len(claims) != len(expected) {
		t.Fatal("claim object has an unexpected key set")
	}
	for _, key := range expected {
		if _, exists := claims[key]; !exists {
			t.Fatalf("required claim key %q is missing", key)
		}
	}
}

func requireCredentialSafeError(t *testing.T, err error, credentials ...string) {
	t.Helper()
	require.Error(t, err)
	for _, credential := range credentials {
		if credential != "" && strings.Contains(err.Error(), credential) {
			t.Fatal("error leaked a credential")
		}
	}
}
