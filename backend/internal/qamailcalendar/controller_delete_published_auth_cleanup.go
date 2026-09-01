//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
)

func prepareInboundRequestFixturePublishedAuthCleanup(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	persisted inboundRequestFixturePersisted,
	authority inboundRequestFixturePublishedEngineAuthority,
) bool {
	lockedRoots, locked := lockInboundRequestFixturePublishedRoots(ctx, tx, plan, persisted, authority)
	if !locked ||
		!lockedRoots.roots.valid() ||
		lockedRoots.owner.owner == nil ||
		lockedRoots.owner.owner.ID != lockedRoots.roots.ownerUserID {
		return false
	}

	return deleteFixtureAuthSideEffectsLocked(ctx, tx, lockedRoots.roots.ownerUserID) == nil
}
