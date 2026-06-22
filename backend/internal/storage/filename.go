package storage

import (
	"path/filepath"
	"strings"
	"unicode"
)

// sanitizeFilename убирает path-компоненты и небезопасные символы (NF-09).
// Оригинальное имя в FileMetadata; санитизируется только сегмент S3-ключа.
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
