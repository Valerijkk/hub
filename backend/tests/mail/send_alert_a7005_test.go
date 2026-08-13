package mail_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/mail"
)

// TestService_SendAlert_WithKeyDoesNotStub_A7_005 — with RESEND_API_KEY set,
// SendAlert must attempt real delivery (error), not silently return nil.
func TestService_SendAlert_WithKeyDoesNotStub_A7_005(t *testing.T) {
	svc := mail.NewService(mail.Config{
		ResendAPIKey: "re_test_key_not_real",
		FromAlerts:   "alerts@hub.local",
	})
	err := svc.SendAlert(context.Background(), mail.SendRequest{
		To: "user@test.local", Subject: "Alert", HTMLBody: "<p>x</p>", Template: "run_failed",
	})
	require.Error(t, err, "A7-005: stub must not succeed when key is set")
	assert.Contains(t, err.Error(), "hub-mail")
}
