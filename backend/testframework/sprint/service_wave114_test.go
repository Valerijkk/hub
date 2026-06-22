package sprint_test
import (
	"github.com/valerii/hub/backend/internal/sprint"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateSprint_NotFound(t *testing.T) {
	svc := sprint.NewSprintService(newMockRepository(), nil, nil)
	name := "Missing"
	_, err := svc.UpdateSprint(context.Background(), uuid.New(), sprint.UpdateSprintRequest{Name: &name})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}
