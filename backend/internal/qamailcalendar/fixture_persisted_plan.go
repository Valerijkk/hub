//go:build qa && !release

package qamailcalendar

func newInboundRequestFixturePersistedFromPlan(
	plan inboundRequestFixturePlan,
) (inboundRequestFixturePersisted, bool) {
	if !inboundRequestFixturePlanIsCanonical(plan) {
		return inboundRequestFixturePersisted{}, false
	}

	persisted := inboundRequestFixturePersisted{
		organizationID:      plan.organizationID,
		ownerUserID:         plan.ownerUserID,
		mailAccountID:       plan.mailAccountID,
		mailIdentityID:      plan.mailIdentityID,
		mailFolderID:        plan.mailFolderID,
		mailThreadID:        plan.mailThreadID,
		mailMessageID:       plan.mailMessageID,
		envelopeEventID:     plan.envelopeEventID,
		calendarLocatorHash: plan.calendarLocatorHash,
	}
	if !inboundRequestFixturePersistedMatchesPlan(persisted, plan) {
		return inboundRequestFixturePersisted{}, false
	}
	return persisted, true
}
