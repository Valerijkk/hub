//go:build hubtest

package nats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func TestRedactNATSURL_A14_003(t *testing.T) {
	got := natsclient.ExportRedactNATSURL("nats://hub:s3cret@nats:4222")
	assert.Equal(t, "nats://***:***@nats:4222", got)

	plain := "nats://nats:4222"
	assert.Equal(t, plain, natsclient.ExportRedactNATSURL(plain))
}
