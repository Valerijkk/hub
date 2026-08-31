//go:build qa && !release

package qamailcalendar

// inboundRequestFixturePublishedFreshBootMapsEmpty is a point-in-time
// three-map snapshot only, not a reservation or restart/PID/generation/worker/
// Fake-engine attestation; C8e/C9d must prove those independently.
func (c *Controller) inboundRequestFixturePublishedFreshBootMapsEmpty() bool {
	if c == nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.fixtures != nil &&
		c.accountMaterializationFixtures != nil &&
		c.threadSearchFixtures != nil &&
		len(c.fixtures) == 0 &&
		len(c.accountMaterializationFixtures) == 0 &&
		len(c.threadSearchFixtures) == 0
}
