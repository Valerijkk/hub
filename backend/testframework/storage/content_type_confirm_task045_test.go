//go:build hubtest

package storage_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valerii/hub/backend/internal/storage"
)

func TestValidateUploadContentType_HTMLRejected(t *testing.T) {
	assert.ErrorIs(t, storage.ExportValidateUploadContentType("text/html"), storage.ErrContentTypeNotAllowed)
	assert.ErrorIs(t, storage.ExportValidateUploadContentType("image/svg+xml"), storage.ErrContentTypeNotAllowed)
	assert.NoError(t, storage.ExportValidateUploadContentType("text/plain"))
	assert.NoError(t, storage.ExportValidateUploadContentType("application/pdf"))
}

func TestIsS3NotFound(t *testing.T) {
	assert.True(t, storage.ExportIsS3NotFound(errors.New("api error NotFound")))
	assert.True(t, storage.ExportIsS3NotFound(errors.New("NoSuchKey")))
	assert.False(t, storage.ExportIsS3NotFound(errors.New("timeout")))
}
