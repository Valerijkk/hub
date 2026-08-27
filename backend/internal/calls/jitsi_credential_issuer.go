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
	publicURL  string
	xmppDomain string
}

type jitsiIssuedCredential struct {
	Token     string
	ExpiresAt time.Time
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
		publicURL:  cfg.PublicURL,
		xmppDomain: cfg.XMPPDomain,
	}, nil
}

// Issue creates a short-lived Jitsi credential for a server-loaded active call.
func (i *jitsiCredentialIssuer) Issue(call *ent.Call, authenticatedUserID uuid.UUID, serverKnownUserName string) (string, error) {
	credential, err := i.issueCredential(call, authenticatedUserID, serverKnownUserName)
	if err != nil {
		return "", err
	}
	return credential.Token, nil
}

// IssueJoinCredential creates provider-neutral Jitsi join data for a server-loaded active call.
func (i *jitsiCredentialIssuer) IssueJoinCredential(call *ent.Call, authenticatedUserID uuid.UUID, serverKnownUserName string) (JoinCredential, error) {
	issued, err := i.issueCredential(call, authenticatedUserID, serverKnownUserName)
	if err != nil {
		return JoinCredential{}, err
	}
	return JoinCredential{
		Provider:   "jitsi",
		RoomName:   call.RoomName,
		ServiceURL: i.publicURL,
		Token:      issued.Token,
	}, nil
}
func (i *jitsiCredentialIssuer) issueCredential(call *ent.Call, authenticatedUserID uuid.UUID, serverKnownUserName string) (jitsiIssuedCredential, error) {
	if call == nil {
		return jitsiIssuedCredential{}, errors.New("cannot issue Jitsi credential without a call")
	}
	if call.ID == uuid.Nil {
		return jitsiIssuedCredential{}, errors.New("cannot issue Jitsi credential for a call without an ID")
	}
	if call.Status != entcall.StatusActive {
		return jitsiIssuedCredential{}, errors.New("cannot issue Jitsi credential for an inactive call")
	}
	if strings.TrimSpace(call.RoomName) == "" || call.RoomName != strings.TrimSpace(call.RoomName) {
		return jitsiIssuedCredential{}, errors.New("cannot issue Jitsi credential without a canonical server room name")
	}
	if authenticatedUserID == uuid.Nil {
		return jitsiIssuedCredential{}, errors.New("cannot issue Jitsi credential without an authenticated user")
	}
	if strings.TrimSpace(serverKnownUserName) == "" {
		return jitsiIssuedCredential{}, errors.New("cannot issue Jitsi credential without a server-known user name")
	}

	issuedAt := i.now().UTC().Truncate(time.Second)
	expiresAt := issuedAt.Add(jitsiCredentialTTL)
	claims := jitsiCredentialClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"jitsi"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
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
		Room: call.RoomName,
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.appSecret)
	if err != nil {
		return jitsiIssuedCredential{}, errors.New("could not issue Jitsi credential")
	}
	return jitsiIssuedCredential{Token: token, ExpiresAt: expiresAt}, nil
}
