//go:build qa && !release

package qamailcalendar

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const accountMaterializationFixtureNamespace = "hub.qa.mail-calendar.account-materialization.v1"

type accountMaterializationFixturePlan struct {
	nonce                         string
	fixedAt                       time.Time
	organizationID                uuid.UUID
	ownerUserID                   uuid.UUID
	mailDomainID                  uuid.UUID
	organizationSlug              string
	domain                        string
	ownerLocalPart                string
	ownerAddress                  string
	domainProvisioningKey         string
	engineDomainID                string
	domainControlHash             [sha256.Size]byte
	expectedAccountIntentID       uuid.UUID
	expectedAccountProvisioningKey string
	expectedAccountControlHash    [sha256.Size]byte
	expectedEngineAccountID       string
}

func newAccountMaterializationFixturePlan(nonce string) (accountMaterializationFixturePlan, error) {
	if !ValidNonce(nonce) {
		return accountMaterializationFixturePlan{}, errInvalidFixtureNonce
	}

	organizationID := accountMaterializationFixtureUUID("organization", nonce)
	ownerUserID := accountMaterializationFixtureUUID("owner-user", nonce)
	mailDomainID := accountMaterializationFixtureUUID("mail-domain", nonce)
	organizationSlug := "qa-materialization-" + nonce
	domain := organizationSlug + ".example.test"
	ownerLocalPart := "owner"
	ownerAddress := ownerLocalPart + "@" + domain
	domainProvisioningKey := "qa-account-materialization-domain-" + nonce
	expectedAccountIntentID := uuid.NewSHA1(
		uuid.NameSpaceURL,
		[]byte("hub.mail.account.intent.v1\x00"+organizationID.String()+"\x00"+ownerUserID.String()+"\x00"+mailDomainID.String()),
	)
	expectedAccountProvisioningKey := "hub-mail-account-v1:" + organizationID.String() + ":" + ownerUserID.String() + ":" + mailDomainID.String() + ":" + expectedAccountIntentID.String()
	expectedAccountControlHash := sha256.Sum256([]byte(
		"hub.mail.account.control.v1\x00" + organizationID.String() + "\x00" + ownerUserID.String() + "\x00" + mailDomainID.String() + "\x00" + expectedAccountIntentID.String() + "\x00" + ownerLocalPart + "\x00enabled=1",
	))
	expectedEngineAccountDigest := sha256.Sum256([]byte(
		"hub.qa.mail.account.provisioner.v1\x00" + expectedAccountProvisioningKey,
	))

	return accountMaterializationFixturePlan{
		nonce:                          nonce,
		fixedAt:                        time.Date(2026, time.August, 26, 0, 0, 0, 0, time.UTC),
		organizationID:                 organizationID,
		ownerUserID:                    ownerUserID,
		mailDomainID:                   mailDomainID,
		organizationSlug:               organizationSlug,
		domain:                         domain,
		ownerLocalPart:                 ownerLocalPart,
		ownerAddress:                   ownerAddress,
		domainProvisioningKey:          domainProvisioningKey,
		engineDomainID:                 "qa-materialization-domain-v1:" + nonce,
		domainControlHash:              accountMaterializationFixtureHash("domain-control", nonce),
		expectedAccountIntentID:        expectedAccountIntentID,
		expectedAccountProvisioningKey: expectedAccountProvisioningKey,
		expectedAccountControlHash:     expectedAccountControlHash,
		expectedEngineAccountID:        "qa-mail-account-v1:" + hex.EncodeToString(expectedEngineAccountDigest[:]),
	}, nil
}

func accountMaterializationFixturePlanMatchesCanonical(plan accountMaterializationFixturePlan) bool {
	canonicalPlan, err := newAccountMaterializationFixturePlan(plan.nonce)
	return err == nil && plan == canonicalPlan
}

func accountMaterializationFixtureUUID(label, nonce string) uuid.UUID {
	digest := accountMaterializationFixtureHash(label, nonce)
	value := uuid.UUID{}
	copy(value[:], digest[:16])
	value[6] = value[6]&0x0f | 0x80
	value[8] = value[8]&0x3f | 0x80
	return value
}

func accountMaterializationFixtureHash(label, nonce string) [sha256.Size]byte {
	return sha256.Sum256([]byte(accountMaterializationFixtureNamespace + "\x00" + label + "\x00" + nonce))
}
