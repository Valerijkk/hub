//go:build hubtest

package observability

import "net"

// ExportRecordOperation wraps recordOperation for hubtest.
func ExportRecordOperation(operation string, durationMs int64, failed bool) {
	recordOperation(operation, durationMs, failed)
}

// ExportParseCIDRs wraps parseCIDRs for hubtest.
func ExportParseCIDRs(cidrs []string) []*net.IPNet {
	return parseCIDRs(cidrs)
}

// ExportIpAllowed wraps ipAllowed for hubtest.
func ExportIpAllowed(ip net.IP, nets []*net.IPNet) bool {
	return ipAllowed(ip, nets)
}

// ExportSetTracingActiveForTest sets tracingActive for hubtest.
func ExportSetTracingActiveForTest(active bool) {
	tracingActive.Store(active)
}
