//go:build hubtest

package middleware

// ExportValidRequestID wraps validRequestID for hubtest (W4-A1-002).
func ExportValidRequestID(id string) bool {
	return validRequestID(id)
}
