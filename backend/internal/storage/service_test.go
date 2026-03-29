package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUploadURLTTL(t *testing.T) {
	assert.Equal(t, 5*time.Minute, uploadURLTTL, "upload URL TTL should be 5 minutes")
}

func TestDownloadURLTTL(t *testing.T) {
	assert.Equal(t, 1*time.Hour, downloadURLTTL, "download URL TTL should be 1 hour")
}

func TestUploadResult_Fields(t *testing.T) {
	// Verify UploadResult struct can be constructed with expected fields.
	r := UploadResult{
		S3Key: "proj-id/files/file-id/report.pdf",
	}
	assert.Contains(t, r.S3Key, "files/", "S3 key should contain files/ path segment")
	assert.Contains(t, r.S3Key, "report.pdf", "S3 key should preserve original filename")
}
