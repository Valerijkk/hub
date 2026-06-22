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
