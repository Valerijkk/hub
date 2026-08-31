package mail

// IsEmpty reports whether this fake holds no persisted, scripted, cached,
// pending, fault, or issued-ID state.
func (f *FakeMailEngine) IsEmpty() bool {
	if f == nil {
		return false
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	return f.serial == 0 &&
		len(f.domains) == 0 &&
		len(f.domainKeys) == 0 &&
		len(f.accounts) == 0 &&
		len(f.accountKeys) == 0 &&
		len(f.identityEmails) == 0 &&
		len(f.folders) == 0 &&
		len(f.messages) == 0 &&
		len(f.blobs) == 0 &&
		len(f.calendarParts) == 0 &&
		len(f.drafts) == 0 &&
		len(f.calendarDraftCreates) == 0 &&
		len(f.submissions) == 0 &&
		len(f.sieves) == 0 &&
		len(f.spamResults) == 0 &&
		len(f.faults) == 0 &&
		len(f.changes) == 0 &&
		len(f.changeScript) == 0 &&
		len(f.enumScript) == 0 &&
		len(f.overrides) == 0 &&
		len(f.snapshots) == 0 &&
		len(f.checkpoints) == 0 &&
		len(f.issuedIDs) == 0
}
