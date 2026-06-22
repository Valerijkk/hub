package env_test

// Тесты модуля тестового окружения: интеграционные и unit-сценарии testframework.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/testframework/env"
	"github.com/valerii/hub/backend/testframework/factory"
)

// TestBuilder_Build_SeedsProjectAndUser проверяет сценарий: builder build seeds project and user.
func TestBuilder_Build_SeedsProjectAndUser(t *testing.T) {
	e := env.New(t, "builder").Build()
	require.NotEqual(t, e.UserID.String(), "00000000-0000-0000-0000-000000000000")
	require.NotEqual(t, e.ProjectID.String(), "00000000-0000-0000-0000-000000000000")

	proj, err := e.Client.Project.Get(e.Ctx, e.ProjectID)
	require.NoError(t, err)
	assert.Equal(t, "TST", proj.Key)

	uf := factory.NewUserFactory(t, e.Client, e.Ctx)
	extra := uf.Create("assignee")
	assert.NotEqual(t, e.UserID, extra.ID)
}
