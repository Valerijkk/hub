//go:build hubtest

package tms_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/platform/middleware"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestUpdateSuite_RequiresVersion_StructValidator — A3W2-1-001: omit version → 400.
// Mutation: убрать Version из UpdateSuiteRequest (или validate tag) → omit проходит Bind.
func TestUpdateSuite_RequiresVersion_StructValidator(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	h := tms.NewHandler(svc)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New(fiber.Config{StructValidator: middleware.NewStructValidator()})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	grp := api.Group("/projects/:projectId/tms", authMw)
	grp.Post("/suites", h.CreateSuite)
	grp.Put("/suites/:suiteId", h.UpdateSuite)

	base := "/api/projects/" + projectID.String() + "/tms/suites"
	createBody := `{"name":"OL Suite"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var created tms.SuiteResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.Equal(t, 1, created.Version)

	req = httptest.NewRequest(http.MethodPut, base+"/"+created.ID, bytes.NewReader([]byte(`{"name":"no-ver"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPut, base+"/"+created.ID, bytes.NewReader([]byte(`{"name":"zero-ver","version":0}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestUpdateSuite_OptimisticLock_Service_A3W2_1_001 — сервисный контракт: stale → 409.
// Prod VersionEQ: TestEntRepository_UpdateSuite_OptimisticLock_A3W2_1_001.
func TestUpdateSuite_OptimisticLock_Service_A3W2_1_001(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	created, err := svc.CreateSuite(ctx, projectID, tms.CreateSuiteRequest{Name: "Lock me"})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	suiteID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	name1 := "v2"
	updated, err := svc.UpdateSuite(ctx, projectID, suiteID, tms.UpdateSuiteRequest{
		Name: &name1, Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "v2", updated.Name)
	assert.Equal(t, 2, updated.Version)

	stale := "should-not-apply"
	_, err = svc.UpdateSuite(ctx, projectID, suiteID, tms.UpdateSuiteRequest{
		Name: &stale, Version: created.Version, // stale: still 1, DB has 2
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)

	name3 := "v3"
	updated2, err := svc.UpdateSuite(ctx, projectID, suiteID, tms.UpdateSuiteRequest{
		Name: &name3, Version: updated.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "v3", updated2.Name)
	assert.Equal(t, 3, updated2.Version)
}

// TestEntRepository_UpdateSuite_OptimisticLock_A3W2_1_001 — stale Version на Postgres
// бьёт в EntRepository.UpdateSuite (VersionEQ), не в мок.
// Mutation: снять testsuite.VersionEQ(input.Version) в ent_repository.go → stale update
// проходит (affected>0), тест краснеет.
func TestEntRepository_UpdateSuite_OptimisticLock_A3W2_1_001(t *testing.T) {
	repo, ctx, projectID, _ := newTMSPostgresRepoEnv(t)

	created, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Lock me"})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	name := "v2"
	updated, err := repo.UpdateSuite(ctx, projectID, created.ID, tms.UpdateSuiteInput{
		Name: &name, Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "v2", updated.Name)
	assert.Equal(t, 2, updated.Version)

	stale := "should-not-apply"
	_, err = repo.UpdateSuite(ctx, projectID, created.ID, tms.UpdateSuiteInput{
		Name: &stale, Version: created.Version, // stale: still 1, DB has 2
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)

	got, err := repo.GetSuite(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "v2", got.Name, "stale write must not change name")
	assert.Equal(t, 2, got.Version)
}
