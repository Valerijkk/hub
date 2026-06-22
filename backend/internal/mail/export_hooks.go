//go:build hubtest

package mail

// ExportInternalTokenHeader возвращает имя внутреннего auth-заголовка для hubtest.
func ExportInternalTokenHeader() string {
	return internalTokenHeader
}

// ExportGetEnv оборачивает getEnv для hubtest.
func ExportGetEnv(key, fallback string) string {
	return getEnv(key, fallback)
}
