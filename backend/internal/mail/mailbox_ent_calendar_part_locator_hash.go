package mail

// mailboxEntCalendarPartLocatorHashValue treats the Ent zero-length carrier on read as logical absence;
// the database CHECK makes a physically persisted empty value invalid.
func mailboxEntCalendarPartLocatorHashValue(value *[]byte) []byte {
	if value == nil || len(*value) == 0 {
		return nil
	}
	return *value
}
