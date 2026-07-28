//go:build hubtest

package http

import "net"

// ExportParseOpenGraphHTML wraps parseOpenGraphHTML for hubtest.
func ExportParseOpenGraphHTML(doc string) *OpenGraphMeta {
	return parseOpenGraphHTML(doc)
}

// ExportIsBlockedIP wraps isBlockedIP for hubtest.
func ExportIsBlockedIP(ip net.IP) bool {
	return isBlockedIP(ip)
}

// ExportNormalizeIPForBlocklist — P1-BE-23: mapped IPv6 → 4 байта (мутация drop To4 → len≠4).
func ExportNormalizeIPForBlocklist(ip net.IP) net.IP {
	return normalizeIPForBlocklist(ip)
}
