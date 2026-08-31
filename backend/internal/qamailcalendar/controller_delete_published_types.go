//go:build qa && !release

package qamailcalendar

import "github.com/google/uuid"

// inboundRequestFixtureRootProof is the exact deterministic nine-node Mail
// root graph plus the fixed envelope relation and message-folder relation.
// It is a private comparable value so Controller CAS operations may compare it
// without retaining Ent rows or pointers.
type inboundRequestFixtureRootProof struct {
	organizationID      uuid.UUID
	ownerUserID         uuid.UUID
	mailDomainID        uuid.UUID
	mailAccountID       uuid.UUID
	mailIdentityID      uuid.UUID
	mailFolderID        uuid.UUID
	mailThreadID        uuid.UUID
	mailMessageID       uuid.UUID
	mailMessageFolderID uuid.UUID
	envelopeEventID     uuid.UUID
}

// inboundRequestFixtureFalseSentProof records the only permitted Mail shape
// for inbound-request-v1 after C2 has proved zero drafts, zero submissions,
// and the single original inbound message.
type inboundRequestFixtureFalseSentProof struct {
	mailAccountID               uuid.UUID
	mailMessageID               uuid.UUID
	originalInboundMessageCount int
	draftCount                  int
	submissionCount             int
}

// inboundRequestFixtureCalendarProof contains only the six derived Calendar
// domain IDs admitted by the published cleanup contract.
type inboundRequestFixtureCalendarProof struct {
	inboundMessageReceiptID uuid.UUID
	inboundObjectStateID    uuid.UUID
	eventID                 uuid.UUID
	attendeeID              uuid.UUID
	replyDeliveryID         uuid.UUID
	mutationReceiptID       uuid.UUID
}

// inboundRequestFixtureCompleteGraphProof joins the exact deterministic Mail
// roots, proven false-sent shape, and six-node Calendar closure. It is a value
// authority only; database proof remains C1-C6 responsibility.
type inboundRequestFixtureCompleteGraphProof struct {
	roots     inboundRequestFixtureRootProof
	falseSent inboundRequestFixtureFalseSentProof
	calendar  inboundRequestFixtureCalendarProof
}

// cleanupDomainProof is retained only during and after a claimed published
// cleanup. It intentionally contains no physical OutboxEvent ID, envelope,
// Ent row, pointer, map, or slice.
type cleanupDomainProof struct {
	present                 bool
	inboundMessageReceiptID uuid.UUID
	inboundObjectStateID    uuid.UUID
	eventID                 uuid.UUID
	attendeeID              uuid.UUID
	replyDeliveryID         uuid.UUID
	mutationReceiptID       uuid.UUID
}

func newInboundRequestFixtureRootProof(
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
) (inboundRequestFixtureRootProof, bool) {
	canonicalPlan, err := newInboundRequestFixturePlan(plan.nonce)
	if err != nil || plan != canonicalPlan || !inboundRequestFixturePersistedMatchesPlan(persisted, plan) {
		return inboundRequestFixtureRootProof{}, false
	}

	proof := inboundRequestFixtureRootProof{
		organizationID:      plan.organizationID,
		ownerUserID:         plan.ownerUserID,
		mailDomainID:        plan.mailDomainID,
		mailAccountID:       plan.mailAccountID,
		mailIdentityID:      plan.mailIdentityID,
		mailFolderID:        plan.mailFolderID,
		mailThreadID:        plan.mailThreadID,
		mailMessageID:       plan.mailMessageID,
		mailMessageFolderID: inboundRequestFixtureUUID("mail-message-folder", plan.nonce),
		envelopeEventID:     plan.envelopeEventID,
	}
	if !proof.valid() {
		return inboundRequestFixtureRootProof{}, false
	}
	return proof, true
}

func (proof inboundRequestFixtureRootProof) valid() bool {
	return proof.organizationID != uuid.Nil &&
		proof.ownerUserID != uuid.Nil &&
		proof.mailDomainID != uuid.Nil &&
		proof.mailAccountID != uuid.Nil &&
		proof.mailIdentityID != uuid.Nil &&
		proof.mailFolderID != uuid.Nil &&
		proof.mailThreadID != uuid.Nil &&
		proof.mailMessageID != uuid.Nil &&
		proof.mailMessageFolderID != uuid.Nil &&
		proof.envelopeEventID != uuid.Nil
}

func inboundRequestFixtureRootProofMatchesPlan(
	proof inboundRequestFixtureRootProof,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
) bool {
	expected, valid := newInboundRequestFixtureRootProof(plan, persisted)
	return valid && proof == expected
}

func newInboundRequestFixtureFalseSentProof(
	roots inboundRequestFixtureRootProof,
	mailAccountID uuid.UUID,
	mailMessageID uuid.UUID,
	originalInboundMessageCount int,
	draftCount int,
	submissionCount int,
) (inboundRequestFixtureFalseSentProof, bool) {
	if !roots.valid() ||
		mailAccountID != roots.mailAccountID ||
		mailMessageID != roots.mailMessageID ||
		originalInboundMessageCount != 1 ||
		draftCount != 0 ||
		submissionCount != 0 {
		return inboundRequestFixtureFalseSentProof{}, false
	}

	return inboundRequestFixtureFalseSentProof{
		mailAccountID:               mailAccountID,
		mailMessageID:               mailMessageID,
		originalInboundMessageCount: originalInboundMessageCount,
		draftCount:                   draftCount,
		submissionCount:              submissionCount,
	}, true
}

func (proof inboundRequestFixtureFalseSentProof) matchesRoots(roots inboundRequestFixtureRootProof) bool {
	return roots.valid() &&
		proof.mailAccountID == roots.mailAccountID &&
		proof.mailMessageID == roots.mailMessageID &&
		proof.originalInboundMessageCount == 1 &&
		proof.draftCount == 0 &&
		proof.submissionCount == 0
}

func newInboundRequestFixtureCalendarProof(
	inboundMessageReceiptID uuid.UUID,
	inboundObjectStateID uuid.UUID,
	eventID uuid.UUID,
	attendeeID uuid.UUID,
	replyDeliveryID uuid.UUID,
	mutationReceiptID uuid.UUID,
) (inboundRequestFixtureCalendarProof, bool) {
	proof := inboundRequestFixtureCalendarProof{
		inboundMessageReceiptID: inboundMessageReceiptID,
		inboundObjectStateID:    inboundObjectStateID,
		eventID:                 eventID,
		attendeeID:              attendeeID,
		replyDeliveryID:         replyDeliveryID,
		mutationReceiptID:       mutationReceiptID,
	}
	if !proof.valid() {
		return inboundRequestFixtureCalendarProof{}, false
	}
	return proof, true
}

func (proof inboundRequestFixtureCalendarProof) valid() bool {
	return proof.inboundMessageReceiptID != uuid.Nil &&
		proof.inboundObjectStateID != uuid.Nil &&
		proof.eventID != uuid.Nil &&
		proof.attendeeID != uuid.Nil &&
		proof.replyDeliveryID != uuid.Nil &&
		proof.mutationReceiptID != uuid.Nil
}

func newInboundRequestFixtureCompleteGraphProof(
	roots inboundRequestFixtureRootProof,
	falseSent inboundRequestFixtureFalseSentProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixtureCompleteGraphProof, bool) {
	if !roots.valid() || !falseSent.matchesRoots(roots) || !calendar.valid() {
		return inboundRequestFixtureCompleteGraphProof{}, false
	}
	return inboundRequestFixtureCompleteGraphProof{
		roots:     roots,
		falseSent: falseSent,
		calendar:  calendar,
	}, true
}

func newCleanupDomainProof(calendar inboundRequestFixtureCalendarProof) (cleanupDomainProof, bool) {
	if !calendar.valid() {
		return cleanupDomainProof{}, false
	}
	return cleanupDomainProof{
		present:                 true,
		inboundMessageReceiptID: calendar.inboundMessageReceiptID,
		inboundObjectStateID:    calendar.inboundObjectStateID,
		eventID:                 calendar.eventID,
		attendeeID:              calendar.attendeeID,
		replyDeliveryID:         calendar.replyDeliveryID,
		mutationReceiptID:       calendar.mutationReceiptID,
	}, true
}

func (proof cleanupDomainProof) valid() bool {
	if !proof.present {
		return proof == (cleanupDomainProof{})
	}
	_, valid := newInboundRequestFixtureCalendarProof(
		proof.inboundMessageReceiptID,
		proof.inboundObjectStateID,
		proof.eventID,
		proof.attendeeID,
		proof.replyDeliveryID,
		proof.mutationReceiptID,
	)
	return valid
}

func inboundRequestFixtureStateAcceptsCleanupDomainProof(
	state inboundRequestFixtureState,
	proof cleanupDomainProof,
) bool {
	if !proof.valid() {
		return false
	}

	switch state {
	case inboundRequestFixtureStateCreating,
		inboundRequestFixtureStateCreated,
		inboundRequestFixtureStateDeleting,
		inboundRequestFixtureStatePublishing,
		inboundRequestFixtureStatePublished,
		inboundRequestFixtureStatePublishUnknown,
		inboundRequestFixtureStateCleanupFailed:
		return !proof.present
	case inboundRequestFixtureStateCleanupClaimed,
		inboundRequestFixtureStateCleanupCommitUnknown,
		inboundRequestFixtureStateCleanupEngineFailed,
		inboundRequestFixtureStateCleaned:
		return proof.present
	default:
		return false
	}
}
