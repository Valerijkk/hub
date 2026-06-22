package gitlab_test
import (
	"github.com/valerii/hub/backend/internal/gitlab"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func TestToIntegrationResponse(t *testing.T) {
	now := time.Now().UTC()
	ig := &ent.GitLabIntegration{
		ID:          uuid.New(),
		GitlabURL:   "https://gitlab.com",
		ProjectPath: "group/project",
		CreatedAt:   now,
	}

	resp := gitlab.ExportToIntegrationResponse(ig)
	require.NotNil(t, resp)
	assert.Equal(t, ig.ID.String(), resp.ID)
	assert.Equal(t, "https://gitlab.com", resp.GitLabURL)
	assert.Equal(t, "group/project", resp.ProjectPath)
	assert.Equal(t, now, resp.CreatedAt)
}
