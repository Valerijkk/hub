//go:build qa && !release

package qamailcalendar

import (
	"bytes"

	"github.com/google/uuid"
)

type inboundRequestFixtureExternalPolicySlot uint8

const (
	inboundRequestFixtureExternalPolicySlotInvalid inboundRequestFixtureExternalPolicySlot = iota
	inboundRequestFixtureExternalPolicySlotOrganizationFirst
	inboundRequestFixtureExternalPolicySlotOrganizationLast inboundRequestFixtureExternalPolicySlot = 14
	inboundRequestFixtureExternalPolicySlotUserFirst         inboundRequestFixtureExternalPolicySlot = 15
	inboundRequestFixtureExternalPolicySlotLast              inboundRequestFixtureExternalPolicySlot = 82
)

type inboundRequestFixtureExternalPolicyRootKind uint8

const (
	inboundRequestFixtureExternalPolicyRootKindInvalid inboundRequestFixtureExternalPolicyRootKind = iota
	inboundRequestFixtureExternalPolicyRootKindOrganization
	inboundRequestFixtureExternalPolicyRootKindUser
)

type inboundRequestFixtureExternalPolicyMode uint8

const (
	inboundRequestFixtureExternalPolicyModeInvalid inboundRequestFixtureExternalPolicyMode = iota
	inboundRequestFixtureExternalPolicyModeForbidden
	inboundRequestFixtureExternalPolicyModeAllowlisted
)

type inboundRequestFixtureExternalPolicySource uint8

const (
	inboundRequestFixtureExternalPolicySourceInvalid inboundRequestFixtureExternalPolicySource = iota
	inboundRequestFixtureExternalPolicySourceNone
	inboundRequestFixtureExternalPolicySourceRootProofOwnerUserID
	inboundRequestFixtureExternalPolicySourceRootProofMailDomainID
	inboundRequestFixtureExternalPolicySourceRootProofMailAccountID
	inboundRequestFixtureExternalPolicySourceCalendarProofEventID
	inboundRequestFixtureExternalPolicySourceCalendarProofInboundObjectStateID
	inboundRequestFixtureExternalPolicySourceCalendarProofMutationReceiptID
	inboundRequestFixtureExternalPolicySourceCalendarProofReplyDeliveryID
	inboundRequestFixtureExternalPolicySourceCalendarProofAttendeeID
)

// inboundRequestFixtureExternalPolicy is the immutable expected result for one
// external-closure slot. Its fields intentionally remain comparable so callers
// can retain a closed proof without Ent rows or dynamic lookup state.
type inboundRequestFixtureExternalPolicy struct {
	slot          inboundRequestFixtureExternalPolicySlot
	rootKind      inboundRequestFixtureExternalPolicyRootKind
	mode          inboundRequestFixtureExternalPolicyMode
	source        inboundRequestFixtureExternalPolicySource
	expectedCount int
	expectedID    uuid.UUID
}

// inboundRequestFixtureExternalObservation retains the ordered Limit(2)
// result for one external-closure slot without storing slices or query rows.
type inboundRequestFixtureExternalObservation struct {
	slot          inboundRequestFixtureExternalPolicySlot
	observedCount int
	rowIDs        [2]uuid.UUID
}

func (slot inboundRequestFixtureExternalPolicySlot) valid() bool {
	return slot >= inboundRequestFixtureExternalPolicySlotOrganizationFirst &&
		slot <= inboundRequestFixtureExternalPolicySlotLast
}

func (rootKind inboundRequestFixtureExternalPolicyRootKind) acceptsSlot(slot inboundRequestFixtureExternalPolicySlot) bool {
	switch rootKind {
	case inboundRequestFixtureExternalPolicyRootKindOrganization:
		return slot >= inboundRequestFixtureExternalPolicySlotOrganizationFirst &&
			slot <= inboundRequestFixtureExternalPolicySlotOrganizationLast
	case inboundRequestFixtureExternalPolicyRootKindUser:
		return slot >= inboundRequestFixtureExternalPolicySlotUserFirst &&
			slot <= inboundRequestFixtureExternalPolicySlotLast
	default:
		return false
	}
}

func (source inboundRequestFixtureExternalPolicySource) validAllowlisted() bool {
	switch source {
	case inboundRequestFixtureExternalPolicySourceRootProofOwnerUserID,
		inboundRequestFixtureExternalPolicySourceRootProofMailDomainID,
		inboundRequestFixtureExternalPolicySourceRootProofMailAccountID,
		inboundRequestFixtureExternalPolicySourceCalendarProofEventID,
		inboundRequestFixtureExternalPolicySourceCalendarProofInboundObjectStateID,
		inboundRequestFixtureExternalPolicySourceCalendarProofMutationReceiptID,
		inboundRequestFixtureExternalPolicySourceCalendarProofReplyDeliveryID,
		inboundRequestFixtureExternalPolicySourceCalendarProofAttendeeID:
		return true
	default:
		return false
	}
}

func newInboundRequestFixtureExternalPolicyForbidden(
	slot inboundRequestFixtureExternalPolicySlot,
	rootKind inboundRequestFixtureExternalPolicyRootKind,
) (inboundRequestFixtureExternalPolicy, bool) {
	policy := inboundRequestFixtureExternalPolicy{
		slot:          slot,
		rootKind:      rootKind,
		mode:          inboundRequestFixtureExternalPolicyModeForbidden,
		source:        inboundRequestFixtureExternalPolicySourceNone,
		expectedCount: 0,
	}
	if !policy.valid() {
		return inboundRequestFixtureExternalPolicy{}, false
	}
	return policy, true
}

func newInboundRequestFixtureExternalPolicyAllowlisted(
	slot inboundRequestFixtureExternalPolicySlot,
	rootKind inboundRequestFixtureExternalPolicyRootKind,
	source inboundRequestFixtureExternalPolicySource,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixtureExternalPolicy, bool) {
	expectedID, valid := resolveInboundRequestFixtureExternalPolicyExpectedID(source, roots, calendar)
	if !valid {
		return inboundRequestFixtureExternalPolicy{}, false
	}

	policy := inboundRequestFixtureExternalPolicy{
		slot:          slot,
		rootKind:      rootKind,
		mode:          inboundRequestFixtureExternalPolicyModeAllowlisted,
		source:        source,
		expectedCount: 1,
		expectedID:    expectedID,
	}
	if !policy.valid() {
		return inboundRequestFixtureExternalPolicy{}, false
	}
	return policy, true
}

func resolveInboundRequestFixtureExternalPolicyExpectedID(
	source inboundRequestFixtureExternalPolicySource,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (uuid.UUID, bool) {
	if !roots.valid() || !calendar.valid() {
		return uuid.Nil, false
	}

	switch source {
	case inboundRequestFixtureExternalPolicySourceRootProofOwnerUserID:
		return roots.ownerUserID, true
	case inboundRequestFixtureExternalPolicySourceRootProofMailDomainID:
		return roots.mailDomainID, true
	case inboundRequestFixtureExternalPolicySourceRootProofMailAccountID:
		return roots.mailAccountID, true
	case inboundRequestFixtureExternalPolicySourceCalendarProofEventID:
		return calendar.eventID, true
	case inboundRequestFixtureExternalPolicySourceCalendarProofInboundObjectStateID:
		return calendar.inboundObjectStateID, true
	case inboundRequestFixtureExternalPolicySourceCalendarProofMutationReceiptID:
		return calendar.mutationReceiptID, true
	case inboundRequestFixtureExternalPolicySourceCalendarProofReplyDeliveryID:
		return calendar.replyDeliveryID, true
	case inboundRequestFixtureExternalPolicySourceCalendarProofAttendeeID:
		return calendar.attendeeID, true
	default:
		return uuid.Nil, false
	}
}

func (policy inboundRequestFixtureExternalPolicy) valid() bool {
	if !policy.slot.valid() || !policy.rootKind.acceptsSlot(policy.slot) {
		return false
	}

	switch policy.mode {
	case inboundRequestFixtureExternalPolicyModeForbidden:
		return policy.source == inboundRequestFixtureExternalPolicySourceNone &&
			policy.expectedCount == 0 &&
			policy.expectedID == uuid.Nil
	case inboundRequestFixtureExternalPolicyModeAllowlisted:
		return policy.source.validAllowlisted() &&
			policy.expectedCount == 1 &&
			policy.expectedID != uuid.Nil
	default:
		return false
	}
}

func newInboundRequestFixtureExternalObservation(
	slot inboundRequestFixtureExternalPolicySlot,
	observedCount int,
	firstID uuid.UUID,
	secondID uuid.UUID,
) (inboundRequestFixtureExternalObservation, bool) {
	observation := inboundRequestFixtureExternalObservation{
		slot:          slot,
		observedCount: observedCount,
		rowIDs:        [2]uuid.UUID{firstID, secondID},
	}
	if !observation.valid() {
		return inboundRequestFixtureExternalObservation{}, false
	}
	return observation, true
}

func (observation inboundRequestFixtureExternalObservation) valid() bool {
	if !observation.slot.valid() {
		return false
	}

	firstID := observation.rowIDs[0]
	secondID := observation.rowIDs[1]
	switch observation.observedCount {
	case 0:
		return firstID == uuid.Nil && secondID == uuid.Nil
	case 1:
		return firstID != uuid.Nil && secondID == uuid.Nil
	case 2:
		return firstID != uuid.Nil &&
			secondID != uuid.Nil &&
			firstID != secondID &&
			bytes.Compare(firstID[:], secondID[:]) < 0
	default:
		return false
	}
}

func (policy inboundRequestFixtureExternalPolicy) matches(observation inboundRequestFixtureExternalObservation) bool {
	if !policy.valid() || !observation.valid() || policy.slot != observation.slot {
		return false
	}

	switch policy.mode {
	case inboundRequestFixtureExternalPolicyModeForbidden:
		return observation.observedCount == 0
	case inboundRequestFixtureExternalPolicyModeAllowlisted:
		return observation.observedCount == 1 && observation.rowIDs[0] == policy.expectedID
	default:
		return false
	}
}
