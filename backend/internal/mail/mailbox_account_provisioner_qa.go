//go:build (qa || hubtest) && !release

package mail

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

const (
	qaMailboxAccountProvisionerNamespace = "hub.qa.mail.account.provisioner.v1"
	qaMailboxAccountProvisionerIDPrefix  = "qa-mail-account-v1:"
)

// QAMailboxAccountProvisioner deterministically projects accepted account
// provisioning input without engine, Registry, or other external I/O.
type QAMailboxAccountProvisioner struct{}

var _ MailboxAccountProvisioner = (*QAMailboxAccountProvisioner)(nil)

// NewQAMailboxAccountProvisioner constructs the stateless QA provisioner.
func NewQAMailboxAccountProvisioner() *QAMailboxAccountProvisioner {
	return &QAMailboxAccountProvisioner{}
}

// EnsureAccount validates one exact Registry-shaped account request and
// returns its deterministic QA projection.
func (provisioner *QAMailboxAccountProvisioner) EnsureAccount(ctx context.Context, spec AccountSpec) (EngineAccount, error) {
	if provisioner == nil || ctx == nil || !validQAMailboxAccountProvisionerSpec(spec) {
		return EngineAccount{}, &EngineError{Kind: EngineErrorRejectedPolicy}
	}
	if ctx.Err() != nil {
		return EngineAccount{}, &EngineError{Kind: EngineErrorUnavailable}
	}

	accountID := qaMailboxAccountProvisionerID(spec.ProvisioningKey)
	if ctx.Err() != nil {
		return EngineAccount{}, &EngineError{Kind: EngineErrorUnavailable}
	}
	return EngineAccount{
		AccountID:         accountID,
		State:             EngineAccountProvisioning,
		AppliedGeneration: spec.DesiredGeneration,
		AppliedHash:       spec.DesiredHash,
		AliasIdentityIDs:  []string{},
	}, nil
}

func validQAMailboxAccountProvisionerSpec(spec AccountSpec) bool {
	return validRegistryProvisioningAccountSpec(spec) && validQAMailboxAccountProvisionerHash(spec.DesiredHash)
}

func validQAMailboxAccountProvisionerHash(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for index := range value {
		character := value[index]
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func qaMailboxAccountProvisionerID(provisioningKey string) string {
	digest := sha256.Sum256([]byte(qaMailboxAccountProvisionerNamespace + "\x00" + provisioningKey))
	return qaMailboxAccountProvisionerIDPrefix + hex.EncodeToString(digest[:])
}
