//go:build hubtest

package config

// ExportIsDevDefault wraps isDevDefault for hubtest.
func ExportIsDevDefault(key, value string) bool {
	return isDevDefault(key, value)
}

// ExportGetEnvInt wraps getEnvInt for hubtest.
func ExportGetEnvInt(key string, fallback int) int {
	return getEnvInt(key, fallback)
}

// ExportGetEnvBool wraps getEnvBool for hubtest.
func ExportGetEnvBool(key string, fallback bool) bool {
	return getEnvBool(key, fallback)
}
