//go:build qa && !release

package qamailcalendar

// inboundRequestFixturePublishedEngineMode states whether published-cleanup
// authority still has a same-boot Fake engine fixture or is reconstructed in a
// fresh boot. It is private and closed so a caller cannot relax mapless rules.
type inboundRequestFixturePublishedEngineMode uint8

const (
	inboundRequestFixturePublishedEngineModeInvalid inboundRequestFixturePublishedEngineMode = iota
	inboundRequestFixturePublishedEngineModeMapped
	inboundRequestFixturePublishedEngineModeMapless
)

// inboundRequestFixturePublishedEngineAuthority is a comparable, pure value
// authority for the engine-facing portion of published cleanup. Mapped mode
// owns the exact same-boot opaque identifiers. Mapless mode intentionally
// retains none of those identifiers and derives admissible database values
// solely from the canonical fixture nonce.
type inboundRequestFixturePublishedEngineAuthority struct {
	mode      inboundRequestFixturePublishedEngineMode
	nonce     string
	messageID string
	version   string
	mapped    inboundRequestEngineFixture
}

func newInboundRequestFixturePublishedMappedEngineAuthority(
	plan inboundRequestFixturePlan,
	mapped inboundRequestEngineFixture,
) (inboundRequestFixturePublishedEngineAuthority, bool) {
	authority := inboundRequestFixturePublishedEngineAuthority{
		mode:      inboundRequestFixturePublishedEngineModeMapped,
		nonce:     plan.nonce,
		messageID: inboundRequestFixturePublishedEngineMessageID(plan),
		version:   inboundRequestFixturePublishedEngineVersion(plan),
		mapped:    mapped,
	}
	if !inboundRequestFixturePlanIsCanonical(plan) || !authority.valid() {
		return inboundRequestFixturePublishedEngineAuthority{}, false
	}
	return authority, true
}

func newInboundRequestFixturePublishedMaplessEngineAuthority(
	plan inboundRequestFixturePlan,
) (inboundRequestFixturePublishedEngineAuthority, bool) {
	authority := inboundRequestFixturePublishedEngineAuthority{
		mode:      inboundRequestFixturePublishedEngineModeMapless,
		nonce:     plan.nonce,
		messageID: inboundRequestFixturePublishedEngineMessageID(plan),
		version:   inboundRequestFixturePublishedEngineVersion(plan),
	}
	if !inboundRequestFixturePlanIsCanonical(plan) || !authority.valid() {
		return inboundRequestFixturePublishedEngineAuthority{}, false
	}
	return authority, true
}

func inboundRequestFixturePlanIsCanonical(plan inboundRequestFixturePlan) bool {
	canonicalPlan, err := newInboundRequestFixturePlan(plan.nonce)
	return err == nil && plan == canonicalPlan
}

func (authority inboundRequestFixturePublishedEngineAuthority) valid() bool {
	plan, err := newInboundRequestFixturePlan(authority.nonce)
	if err != nil || authority.nonce != plan.nonce {
		return false
	}
	if authority.messageID != inboundRequestFixturePublishedEngineMessageID(plan) ||
		authority.version != inboundRequestFixturePublishedEngineVersion(plan) {
		return false
	}

	switch authority.mode {
	case inboundRequestFixturePublishedEngineModeMapped:
		return authority.mapped.accountID != "" &&
			authority.mapped.primaryIdentityID != "" &&
			authority.mapped.folderID != "" &&
			authority.mapped.messageID == inboundRequestFixturePublishedEngineMessageID(plan) &&
			authority.mapped.version == inboundRequestFixturePublishedEngineVersion(plan)
	case inboundRequestFixturePublishedEngineModeMapless:
		return authority.mapped == (inboundRequestEngineFixture{})
	default:
		return false
	}
}

func (authority inboundRequestFixturePublishedEngineAuthority) mappedMode() bool {
	return authority.valid() && authority.mode == inboundRequestFixturePublishedEngineModeMapped
}

func (authority inboundRequestFixturePublishedEngineAuthority) maplessMode() bool {
	return authority.valid() && authority.mode == inboundRequestFixturePublishedEngineModeMapless
}

func (authority inboundRequestFixturePublishedEngineAuthority) matchesAccountFolderIdentity(
	engineAccountID string,
	engineFolderID string,
	engineIdentityID string,
) bool {
	if !authority.valid() {
		return false
	}

	switch authority.mode {
	case inboundRequestFixturePublishedEngineModeMapped:
		return engineAccountID == authority.mapped.accountID &&
			engineFolderID == authority.mapped.folderID &&
			engineIdentityID == authority.mapped.primaryIdentityID
	case inboundRequestFixturePublishedEngineModeMapless:
		return engineAccountID != "" &&
			engineFolderID != "" &&
			engineIdentityID == engineAccountID+"-primary"
	default:
		return false
	}
}

func (authority inboundRequestFixturePublishedEngineAuthority) matchesMessageVersion(
	engineMessageID string,
	engineVersion string,
) bool {
	if !authority.valid() {
		return false
	}
	plan, err := newInboundRequestFixturePlan(authority.nonce)
	return err == nil &&
		engineMessageID == inboundRequestFixturePublishedEngineMessageID(plan) &&
		engineVersion == inboundRequestFixturePublishedEngineVersion(plan)
}

func inboundRequestFixturePublishedEngineMessageID(plan inboundRequestFixturePlan) string {
	return "qa-message-" + plan.nonce
}

func inboundRequestFixturePublishedEngineVersion(plan inboundRequestFixturePlan) string {
	return "qa-version-" + plan.nonce
}
