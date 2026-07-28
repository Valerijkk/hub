//go:build hubtest

package tms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/valerii/hub/backend/internal/tms"
)

// Mutation: HasSuffix без ToLower → .CSV → "".
func TestDetectImportFormat_CaseInsensitive_P3BE09(t *testing.T) {
	assert.Equal(t, "csv", tms.ExportDetectImportFormat("cases.CSV"))
	assert.Equal(t, "json", tms.ExportDetectImportFormat("data.Json"))
	assert.Equal(t, "xml", tms.ExportDetectImportFormat("suite.XML"))
	assert.Equal(t, "excel", tms.ExportDetectImportFormat("book.XLSX"))
	assert.Equal(t, "", tms.ExportDetectImportFormat("notes.txt"))
}
