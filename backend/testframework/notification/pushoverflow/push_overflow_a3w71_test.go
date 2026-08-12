package pushoverflow_test

import (
	"testing"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/notification"
	"github.com/valerii/hub/backend/internal/platform/config"
)

func TestPushSender_QueueOverflow_A3W7_1_002(t *testing.T) {
	priv, pub, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)
	push := notification.NewPushSender(nil, config.WebPushConfig{
		VAPIDPublicKey: pub, VAPIDPrivateKey: priv, VAPIDSubject: "mailto:test@hub.local",
	})
	require.NotNil(t, push)
	uid := uuid.New()
	for i := 0; i < 201; i++ {
		push.Send(uid, "title", "body", "/")
	}
	depth := notification.ExportPushQueueLen(push)
	assert.LessOrEqual(t, depth, 200)
	assert.Equal(t, 200, depth)
}
