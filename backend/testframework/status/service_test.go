package status_test
import (
	"github.com/valerii/hub/backend/internal/status"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetStatus_allOperational(t *testing.T) {
	ok := func(ctx context.Context) error { return nil }
	svc := status.NewService(ok, ok, ok, ok)

	resp := svc.GetStatus(context.Background())
	assert.Equal(t, "operational", resp.Status)
	require.Len(t, resp.Components, 5)
	for _, c := range resp.Components {
		assert.Equal(t, "operational", c.Status)
	}
}

func TestService_GetStatus_degraded(t *testing.T) {
	fail := func(ctx context.Context) error { return errors.New("down") }
	svc := status.NewService(fail, nil, fail, nil)

	resp := svc.GetStatus(context.Background())
	assert.Equal(t, "degraded", resp.Status)

	byID := map[string]string{}
	for _, c := range resp.Components {
		byID[c.ID] = c.Status
	}
	assert.Equal(t, "outage", byID["postgres"])
	assert.Equal(t, "unknown", byID["redis"])
	assert.Equal(t, "outage", byID["mongo"])
	assert.Equal(t, "unknown", byID["nats"])
	assert.Equal(t, "operational", byID["api"])
}
