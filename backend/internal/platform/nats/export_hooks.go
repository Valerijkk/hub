//go:build hubtest

package nats

import "time"

// ExportRedactNATSURL wraps redactNATSURL for hubtest (A14-003).
func ExportRedactNATSURL(raw string) string {
	return redactNATSURL(raw)
}

// ExportSetAfterCorePublishHook sets a hook between Publish and Flush
// (P1-BE-24 mutation: drop before Flush → PublishMsg returns error).
func ExportSetAfterCorePublishHook(fn func()) {
	afterCorePublishHook = fn
}

// ExportNATSDrainTimeout — W4-A2-006: Close ждёт Drain не дольше этого.
func ExportNATSDrainTimeout() time.Duration {
	return natsDrainTimeout
}
