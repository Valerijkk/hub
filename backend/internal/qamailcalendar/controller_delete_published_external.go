//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
)

// inboundRequestFixturePublishedExternalProof retains the complete frozen
// external-closure result without dynamic lookup state or Ent rows.
type inboundRequestFixturePublishedExternalProof struct {
	organization inboundRequestFixturePublishedExternalOrganizationShard
	userA        inboundRequestFixturePublishedExternalUserAShard
	userB        inboundRequestFixturePublishedExternalUserBShard
	userC        inboundRequestFixturePublishedExternalUserCShard
	userD        inboundRequestFixturePublishedExternalUserDShard
	userE        inboundRequestFixturePublishedExternalUserEShard
}

func observeInboundRequestFixturePublishedExternal(
	ctx context.Context,
	tx *ent.Tx,
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) (inboundRequestFixturePublishedExternalProof, bool) {
	if ctx == nil || tx == nil || !roots.valid() || !calendar.valid() {
		return inboundRequestFixturePublishedExternalProof{}, false
	}

	organization, observed := observeInboundRequestFixturePublishedExternalOrganization(ctx, tx, roots, calendar)
	if !observed {
		return inboundRequestFixturePublishedExternalProof{}, false
	}

	userA, observed := observeInboundRequestFixturePublishedExternalUserA(ctx, tx, roots, calendar)
	if !observed {
		return inboundRequestFixturePublishedExternalProof{}, false
	}

	userB, observed := observeInboundRequestFixturePublishedExternalUserB(ctx, tx, roots, calendar)
	if !observed {
		return inboundRequestFixturePublishedExternalProof{}, false
	}

	userC, observed := observeInboundRequestFixturePublishedExternalUserC(ctx, tx, roots, calendar)
	if !observed {
		return inboundRequestFixturePublishedExternalProof{}, false
	}

	userD, observed := observeInboundRequestFixturePublishedExternalUserD(ctx, tx, roots, calendar)
	if !observed {
		return inboundRequestFixturePublishedExternalProof{}, false
	}

	userE, observed := observeInboundRequestFixturePublishedExternalUserE(ctx, tx, roots, calendar)
	if !observed {
		return inboundRequestFixturePublishedExternalProof{}, false
	}

	proof := inboundRequestFixturePublishedExternalProof{
		organization: organization,
		userA:        userA,
		userB:        userB,
		userC:        userC,
		userD:        userD,
		userE:        userE,
	}
	if !proof.valid(roots, calendar) {
		return inboundRequestFixturePublishedExternalProof{}, false
	}
	return proof, true
}

func (proof inboundRequestFixturePublishedExternalProof) valid(
	roots inboundRequestFixtureRootProof,
	calendar inboundRequestFixtureCalendarProof,
) bool {
	if !roots.valid() || !calendar.valid() {
		return false
	}

	organizationValid := proof.organization.valid(roots, calendar)
	userAValid := proof.userA.valid(roots, calendar)
	userBValid := proof.userB.valid(roots, calendar)
	userCValid := proof.userC.valid(roots, calendar)
	userDValid := proof.userD.valid(roots, calendar)
	userEValid := proof.userE.valid(roots, calendar)
	return organizationValid && userAValid && userBValid && userCValid && userDValid && userEValid
}
