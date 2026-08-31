//go:build qa && !release

package qamailcalendar

import (
	"bytes"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

const (
	threadSearchFixtureNamespace = "hub.qa.mail-calendar.thread-search.v1"

	threadSearchFixtureRolePinnedHigh        = "pinned-high"
	threadSearchFixtureRolePinnedLow         = "pinned-low"
	threadSearchFixtureRoleRecent            = "recent"
	threadSearchFixtureRoleOlder             = "older"
	threadSearchFixtureRoleQAOnlyDecoy       = "qa-only-decoy"
	threadSearchFixtureRoleCalendarOnlyDecoy = "calendar-only-decoy"
)

func newThreadSearchFixturePlan(nonce string) (threadSearchFixturePlan, error) {
	if !ValidNonce(nonce) {
		return threadSearchFixturePlan{}, errInvalidFixtureNonce
	}

	fixedAt := time.Date(2026, time.August, 29, 0, 0, 0, 0, time.UTC)
	organizationSlug := "qa-thread-search-" + nonce
	domain := organizationSlug + ".example.test"
	ownerAddress := "owner@" + domain

	pinnedHighIDs := threadSearchFixtureChildIDs("pinned-first", nonce)
	pinnedLowIDs := threadSearchFixtureChildIDs("pinned-second", nonce)
	if bytes.Compare(pinnedHighIDs.threadID[:], pinnedLowIDs.threadID[:]) < 0 {
		pinnedHighIDs, pinnedLowIDs = pinnedLowIDs, pinnedHighIDs
	}

	rows := threadSearchFixtureRows{
		newThreadSearchFixtureRow(
			nonce,
			pinnedHighIDs,
			threadSearchFixtureRolePinnedHigh,
			"qa "+nonce+" pinned-high",
			"calendar alpha schedule token",
			true,
			true,
			fixedAt.Add(2*time.Minute),
		),
		newThreadSearchFixtureRow(
			nonce,
			pinnedLowIDs,
			threadSearchFixtureRolePinnedLow,
			"calendar "+nonce+" pinned-low",
			"qa beta schedule token",
			true,
			true,
			fixedAt.Add(2*time.Minute),
		),
		newThreadSearchFixtureRow(
			nonce,
			threadSearchFixtureChildIDs("recent", nonce),
			threadSearchFixtureRoleRecent,
			"qa "+nonce+" recent calendar",
			"gamma schedule token",
			false,
			false,
			fixedAt.Add(3*time.Minute),
		),
		newThreadSearchFixtureRow(
			nonce,
			threadSearchFixtureChildIDs("older", nonce),
			threadSearchFixtureRoleOlder,
			"qa "+nonce+" older calendar",
			"delta schedule token",
			false,
			false,
			fixedAt,
		),
		newThreadSearchFixtureRow(
			nonce,
			threadSearchFixtureChildIDs("qa-only-decoy", nonce),
			threadSearchFixtureRoleQAOnlyDecoy,
			"qa "+nonce+" only",
			"epsilon schedule token",
			false,
			false,
			fixedAt.Add(5*time.Minute),
		),
		newThreadSearchFixtureRow(
			nonce,
			threadSearchFixtureChildIDs("calendar-only-decoy", nonce),
			threadSearchFixtureRoleCalendarOnlyDecoy,
			"calendar "+nonce+" only",
			"zeta schedule token",
			false,
			false,
			fixedAt.Add(4*time.Minute),
		),
	}

	return threadSearchFixturePlan{
		nonce:                  nonce,
		fixedAt:                fixedAt,
		organizationID:         threadSearchFixtureUUID("organization", nonce),
		ownerUserID:            threadSearchFixtureUUID("owner-user", nonce),
		mailDomainID:           threadSearchFixtureUUID("mail-domain", nonce),
		mailAccountID:          threadSearchFixtureUUID("mail-account", nonce),
		mailIdentityID:         threadSearchFixtureUUID("mail-identity", nonce),
		mailFolderID:           threadSearchFixtureUUID("mail-folder", nonce),
		organizationSlug:       organizationSlug,
		domain:                 domain,
		ownerAddress:           ownerAddress,
		domainProvisioningKey:  "qa-thread-search-domain-" + nonce,
		accountProvisioningKey: "qa-thread-search-account-" + nonce,
		engineDomainID:         "qa-thread-search-domain-v1:" + nonce,
		engineAccountID:        "qa-thread-search-account-v1:" + nonce,
		engineIdentityID:       "qa-thread-search-identity-v1:" + nonce,
		engineFolderID:         "qa-thread-search-folder-v1:" + nonce,
		folderName:             "QA Thread Search " + nonce,
		normalizedFolderName:   "qa thread search " + nonce,
		correlationID:          "qa-thread-search-" + nonce,
		domainControlHash:      threadSearchFixtureHash("domain-control", nonce),
		accountControlHash:     threadSearchFixtureHash("account-control", nonce),
		rows:                   rows,
		matchingThreadIDs: [4]uuid.UUID{
			rows[0].threadID,
			rows[1].threadID,
			rows[2].threadID,
			rows[3].threadID,
		},
		decoyThreadIDs: [2]uuid.UUID{
			rows[4].threadID,
			rows[5].threadID,
		},
	}, nil
}

func newThreadSearchFixtureRow(
	nonce string,
	ids threadSearchFixtureRowIDs,
	role string,
	normalizedSubject string,
	snippet string,
	isPinned bool,
	isFlagged bool,
	lastMessageAt time.Time,
) threadSearchFixtureRow {
	return threadSearchFixtureRow{
		role:              role,
		threadID:          ids.threadID,
		messageID:         ids.messageID,
		messageFolderID:   ids.messageFolderID,
		normalizedSubject: normalizedSubject,
		snippet:           snippet,
		isPinned:          isPinned,
		isFlagged:         isFlagged,
		lastMessageAt:     lastMessageAt,
		threadKey:         threadSearchFixtureHash("thread-key-"+ids.label, nonce),
		engineMessageID:   "qa-thread-search-message-v1:" + ids.label + ":" + nonce,
		engineVersion:     "qa-thread-search-version-v1:" + ids.label + ":" + nonce,
	}
}

func threadSearchFixturePlanMatchesCanonical(plan threadSearchFixturePlan) bool {
	canonicalPlan, err := newThreadSearchFixturePlan(plan.nonce)
	return err == nil && plan == canonicalPlan
}

func threadSearchFixtureChildIDs(label, nonce string) threadSearchFixtureRowIDs {
	return threadSearchFixtureRowIDs{
		label:           label,
		threadID:        threadSearchFixtureUUID("mail-thread-"+label, nonce),
		messageID:       threadSearchFixtureUUID("mail-message-"+label, nonce),
		messageFolderID: threadSearchFixtureUUID("mail-message-folder-"+label, nonce),
	}
}

func threadSearchFixtureUUID(label, nonce string) uuid.UUID {
	digest := threadSearchFixtureHash(label, nonce)
	value := uuid.UUID{}
	copy(value[:], digest[:16])
	value[6] = value[6]&0x0f | 0x80
	value[8] = value[8]&0x3f | 0x80
	return value
}

func threadSearchFixtureHash(label, nonce string) [sha256.Size]byte {
	return sha256.Sum256([]byte(threadSearchFixtureNamespace + "\x00" + label + "\x00" + nonce))
}
