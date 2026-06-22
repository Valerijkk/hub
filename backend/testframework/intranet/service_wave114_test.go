package intranet_test
import (
	"github.com/valerii/hub/backend/internal/intranet"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_DeleteNews_NotFound(t *testing.T) {
	svc, adminID, _ := setupIntranetService(t)
	err := svc.DeleteNews(context.Background(), adminID, uuid.New())
	assert.ErrorIs(t, err, intranet.ErrNotFound)
}

func TestService_DeleteNews_Forbidden(t *testing.T) {
	svc, adminID, memberID := setupIntranetService(t)
	ctx := context.Background()

	created, err := svc.CreateNews(ctx, adminID, intranet.CreateNewsRequest{
		Title: "To delete", Body: "x", Published: true,
	})
	require.NoError(t, err)

	err = svc.DeleteNews(ctx, memberID, created.ID)
	assert.ErrorIs(t, err, intranet.ErrForbidden)
}
