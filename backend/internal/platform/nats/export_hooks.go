//go:build hubtest

package nats

// ExportRedactNATSURL wraps redactNATSURL for hubtest (A14-003).
func ExportRedactNATSURL(raw string) string {
	return redactNATSURL(raw)
}
