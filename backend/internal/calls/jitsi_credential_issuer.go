package calls

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
	"github.com/valerii/hub/backend/internal/platform/config"
)

const jitsiCredentialTTL = 5 * time.Minute

type jitsiCredentialIssuer struct {
	appID      string
	appSecret  []byte
	now        func() time.Time
	xmppDomain string
}

type jitsiCredentialClaims struct {
	jwt.RegisteredClaims
	Context jitsiCredentialContext `json:"context"`
	Room    string                 `json:"room"`
}

type jitsiCredentialContext struct {
	User jitsiCredentialUser `json:"user"`
}

type jitsiCredentialUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func newJitsiCredentialIssuer(cfg config.JitsiConfig, now func() time.Time) (*jitsiCredentialIssuer, error) {
	if err := config.ValidateJitsiConfig(cfg); err != nil {
		return nil, err
	}
	if now == nil {
		return nil, errors.New("Jitsi credential issuer clock is required")
	}

	return &jitsiCredentialIssuer{
		appID:      cfg.AppID,
		appSecret:  append([]byte(nil), []byte(cfg.AppSecret)...),
		now:        now,
		xmppDomain: cfg.XMPPDomain,
	}, nil
}

// Issue creates a short-lived Jitsi credential for a server-loaded active call.
func (i *jitsiCredentialIssuer) Issue(call *ent.Call, authenticatedUserID uuid.UUID, serverKnownUserName string) (string, error) {
	if call == nil {
		return "", errors.New("cannot issue Jitsi credential without a call")
	}
	if call.ID == uuid.Nil {
		return "", errors.New("cannot issue Jitsi credential for a call without an ID")
	}
	if call.Status != entcall.StatusActive {
		return "", errors.New("cannot issue Jitsi credential for an inactive call")
	}
	if authenticatedUserID == uuid.Nil {
		return "", errors.New("cannot issue Jitsi credential without an authenticated user")
	}
	if strings.TrimSpace(serverKnownUserName) == "" {
		return "", errors.New("cannot issue Jitsi credential without a server-known user name")
	}

	issuedAt := i.now().UTC().Truncate(time.Second)
	claims := jitsiCredentialClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"jitsi"},
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(jitsiCredentialTTL)),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			Issuer:    i.appID,
			Subject:   i.xmppDomain,
		},
		Context: jitsiCredentialContext{
			User: jitsiCredentialUser{
				ID:   authenticatedUserID.String(),
				Name: serverKnownUserName,
			},
		},
		Room: call.ID.String(),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.appSecret)
	if err != nil {
		return "", errors.New("could not issue Jitsi credential")
	}
	return token, nil
}
