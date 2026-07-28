//go:build hubtest

package nats

// ExportRedactNATSURL wraps redactNATSURL for hubtest (A14-003).
func ExportRedactNATSURL(raw string) string {
	return redactNATSURL(raw)
}

// ExportSetAfterCorePublishHook sets a hook between Publish and Flush
// (P1-BE-24 mutation: drop before Flush → PublishMsg returns error).
func ExportSetAfterCorePublishHook(fn func()) {
	afterCorePublishHook = fn
}
