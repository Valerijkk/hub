package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "report.pdf", "report.pdf"},
		{"unix traversal", "../../etc/passwd", "passwd"},
		{"windows traversal", "..\\secret.txt", "secret.txt"},
		{"spaces and parens", "my file (1).png", "my_file_1.png"},
		{"empty", "", "file"},
		{"only dots", "../../../", "file"},
		{"max length trim", strings.Repeat("a", 250) + ".bin", strings.Repeat("a", 200)},
		{"null byte injection", "safe\x00.exe", "safe.exe"},
		{"control chars stripped", "doc\x1f\x7f.pdf", "doc.pdf"},
		{"hidden dotfile", ".htaccess", "htaccess"},
		{"absolute unix path", "/etc/shadow", "shadow"},
		{"absolute windows path", `C:\Users\secret.doc`, "secret.doc"},
		{"mixed separators", `foo/bar\baz.txt`, "baz.txt"},
		{"only unsafe chars", "???", "file"},
		{"leading trailing dots trimmed", "...report...", "report"},
		{"unicode letters kept", "отчёт.pdf", "отчёт.pdf"},
		{"emoji stripped", "file🚀.txt", "file.txt"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeFilename(tc.in)
			assert.Equal(t, tc.want, got)
			assert.NotContains(t, got, "..")
			assert.NotContains(t, got, "/")
			assert.NotContains(t, got, "\\")
			assert.LessOrEqual(t, len(got), 200)
		})
	}
}

func TestSanitizeFilename_NoPathTraversalInOutput(t *testing.T) {
	payloads := []string{
		"../../../etc/passwd",
		"....//....//etc/passwd",
		"foo/../../../bar.txt",
		"%2e%2e%2f", // encoded traversal attempt stays literal after sanitize
	}
	for _, in := range payloads {
		out := sanitizeFilename(in)
		assert.False(t, strings.HasPrefix(out, "."), "output must not start with dot: %q -> %q", in, out)
		assert.NotContains(t, out, "..")
	}
}
