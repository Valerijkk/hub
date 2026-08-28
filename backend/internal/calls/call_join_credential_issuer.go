package calls

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	entcall "github.com/valerii/hub/backend/ent/call"
)

// ErrCallJoinCredentialInvalid marks malformed trusted inputs at the issuer boundary.
var ErrCallJoinCredentialInvalid = errors.New("invalid call join credential input")

// CallJoinCredentialIssuer issues a credential from trusted server-loaded call
// and identity facts. It is not an authorization boundary: a future service
// owns loading, membership, and authorization before invoking an issuer.
type CallJoinCredentialIssuer interface {
	IssueCallJoinCredential(call *ent.Call, authenticatedUserID uuid.UUID, serverKnownUserName string) (JoinCredential, error)
}

var _ CallJoinCredentialIssuer = (*LiveKitConferenceProvider)(nil)
var _ CallJoinCredentialIssuer = (*jitsiCredentialIssuer)(nil)

// IssueCallJoinCredential validates trusted inputs then delegates credential
// construction to the legacy LiveKit issuer exactly once.
func (p *LiveKitConferenceProvider) IssueCallJoinCredential(
	call *ent.Call,
	authenticatedUserID uuid.UUID,
	serverKnownUserName string,
) (JoinCredential, error) {
	if err := validateCallJoinCredentialInput(call, authenticatedUserID, serverKnownUserName); err != nil {
		return JoinCredential{}, err
	}

	return p.IssueJoinCredential(call.RoomName, authenticatedUserID.String(), serverKnownUserName)
}

// IssueCallJoinCredential validates trusted inputs then delegates credential
// construction to the legacy Jitsi issuer exactly once.
func (i *jitsiCredentialIssuer) IssueCallJoinCredential(
	call *ent.Call,
	authenticatedUserID uuid.UUID,
	serverKnownUserName string,
) (JoinCredential, error) {
	if err := validateCallJoinCredentialInput(call, authenticatedUserID, serverKnownUserName); err != nil {
		return JoinCredential{}, err
	}

	return i.IssueJoinCredential(call, authenticatedUserID, serverKnownUserName)
}

func validateCallJoinCredentialInput(
	call *ent.Call,
	authenticatedUserID uuid.UUID,
	serverKnownUserName string,
) error {
	if call == nil ||
		call.ID == uuid.Nil ||
		call.Status != entcall.StatusActive ||
		strings.TrimSpace(call.RoomName) == "" ||
		call.RoomName != strings.TrimSpace(call.RoomName) ||
		authenticatedUserID == uuid.Nil ||
		strings.TrimSpace(serverKnownUserName) == "" {
		return ErrCallJoinCredentialInvalid
	}

	return nil
}
