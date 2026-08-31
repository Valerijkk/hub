//go:build qa && !release

package qamailcalendar

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const inboundRequestFixtureNamespace = "hub.qa.mail-calendar.fixture.v1"

var errInvalidFixtureNonce = errors.New("qa mail calendar fixture nonce is invalid")

type inboundRequestFixturePlan struct {
	nonce                  string
	fixedAt                time.Time
	meetingStartsAt        time.Time
	meetingEndsAt          time.Time
	organizationID         uuid.UUID
	ownerUserID            uuid.UUID
	mailDomainID           uuid.UUID
	mailAccountID          uuid.UUID
	mailIdentityID         uuid.UUID
	mailFolderID           uuid.UUID
	mailThreadID           uuid.UUID
	mailMessageID          uuid.UUID
	envelopeEventID        uuid.UUID
	organizationSlug       string
	domain                 string
	ownerAddress           string
	organizerAddress       string
	domainProvisioningKey  string
	accountProvisioningKey string
	folderName             string
	normalizedFolderName   string
	subject                string
	normalizedSubject      string
	snippet                string
	itipUID                string
	correlationID          string
	threadKey              [sha256.Size]byte
	calendarLocatorHash    [sha256.Size]byte
	domainControlHash      [sha256.Size]byte
	accountControlHash     [sha256.Size]byte
	calendarText           string
}

func newInboundRequestFixturePlan(nonce string) (inboundRequestFixturePlan, error) {
	if !ValidNonce(nonce) {
		return inboundRequestFixturePlan{}, errInvalidFixtureNonce
	}

	organizationSlug := fmt.Sprintf("qa-%s", nonce)
	domain := organizationSlug + ".example.test"
	ownerAddress := "owner@" + domain
	organizerAddress := fmt.Sprintf("organizer-%s@example.test", nonce)
	fixedAt := time.Date(2026, time.August, 26, 0, 0, 0, 0, time.UTC)
	meetingStartsAt := time.Date(2026, time.August, 27, 10, 0, 0, 0, time.UTC)
	meetingEndsAt := time.Date(2026, time.August, 27, 10, 30, 0, 0, time.UTC)
	itipUID := "qa-inbound-request-" + nonce

	plan := inboundRequestFixturePlan{
		nonce:                  nonce,
		fixedAt:                fixedAt,
		meetingStartsAt:        meetingStartsAt,
		meetingEndsAt:          meetingEndsAt,
		organizationID:         inboundRequestFixtureUUID("organization", nonce),
		ownerUserID:            inboundRequestFixtureUUID("owner-user", nonce),
		mailDomainID:           inboundRequestFixtureUUID("mail-domain", nonce),
		mailAccountID:          inboundRequestFixtureUUID("mail-account", nonce),
		mailIdentityID:         inboundRequestFixtureUUID("mail-identity", nonce),
		mailFolderID:           inboundRequestFixtureUUID("mail-folder", nonce),
		mailThreadID:           inboundRequestFixtureUUID("mail-thread", nonce),
		mailMessageID:          inboundRequestFixtureUUID("mail-message", nonce),
		envelopeEventID:        inboundRequestFixtureUUID("envelope-event", nonce),
		organizationSlug:       organizationSlug,
		domain:                 domain,
		ownerAddress:           ownerAddress,
		organizerAddress:       organizerAddress,
		domainProvisioningKey:  "qa-domain-" + nonce,
		accountProvisioningKey: "qa-account-" + nonce,
		folderName:             "QA Inbox " + nonce,
		normalizedFolderName:   "qa inbox " + nonce,
		subject:                "QA Mail Calendar Meeting",
		normalizedSubject:      "qa mail calendar meeting",
		snippet:                "QA inbound calendar invitation",
		itipUID:                itipUID,
		correlationID:          "qa-mail-calendar-" + nonce,
		threadKey:              inboundRequestFixtureHash("thread-key", nonce),
		calendarLocatorHash:    inboundRequestFixtureHash("calendar-locator", nonce),
		domainControlHash:      inboundRequestFixtureHash("domain-control", nonce),
		accountControlHash:     inboundRequestFixtureHash("account-control", nonce),
		calendarText: fmt.Sprintf(
			"BEGIN:VCALENDAR\r\nVERSION:2.0\r\nCALSCALE:GREGORIAN\r\nMETHOD:REQUEST\r\nBEGIN:VEVENT\r\nUID:%s\r\nDTSTAMP:20260826T000000Z\r\nDTSTART:20260827T100000Z\r\nDTEND:20260827T103000Z\r\nSEQUENCE:0\r\nSTATUS:CONFIRMED\r\nSUMMARY:QA Mail Calendar Meeting\r\nORGANIZER:mailto:%s\r\nATTENDEE;PARTSTAT=NEEDS-ACTION;RSVP=TRUE:\r\n mailto:%s\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n",
			itipUID,
			organizerAddress,
			ownerAddress,
		),
	}

	return plan, nil
}

func inboundRequestFixtureUUID(label, nonce string) uuid.UUID {
	digest := inboundRequestFixtureHash(label, nonce)
	value := uuid.UUID{}
	copy(value[:], digest[:16])
	value[6] = value[6]&0x0f | 0x80
	value[8] = value[8]&0x3f | 0x80
	return value
}

func inboundRequestFixtureHash(label, nonce string) [sha256.Size]byte {
	return sha256.Sum256([]byte(inboundRequestFixtureNamespace + "\x00" + label + "\x00" + nonce))
}
