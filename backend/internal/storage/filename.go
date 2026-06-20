package storage

import (
	"path/filepath"
	"strings"
	"unicode"
)

// sanitizeFilename strips path components and unsafe characters from upload names (NF-09).
// Original filename is kept in FileMetadata; only the S3 key segment is sanitized.
func sanitizeFilename(name string) string {
	base := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	if base == "." || base == ".." {
		base = "file"
	}

	var b strings.Builder
	b.Grow(len(base))
	for _, r := range base {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '.', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), ".")
	if out == "" {
		return "file"
	}
	if len(out) > 200 {
		return out[:200]
	}
	return out
}
