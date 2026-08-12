package status_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/status"
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

func TestService_GetPublicStatus_NoComponents_W7A5002(t *testing.T) {
	fail := func(ctx context.Context) error { return errors.New("down") }
	svc := status.NewService(fail, nil, nil, nil)

	pub := svc.GetPublicStatus(context.Background())
	assert.Equal(t, "degraded", pub.Status)
	require.Len(t, pub.Components, 1)
	assert.Equal(t, "hub", pub.Components[0].ID)
	assert.Equal(t, "degraded", pub.Components[0].Status)

	det := svc.GetDetailedStatus(context.Background())
	require.GreaterOrEqual(t, len(det.Components), 4)
	ids := map[string]bool{}
	for _, c := range det.Components {
		ids[c.ID] = true
	}
	assert.True(t, ids["postgres"])
}
