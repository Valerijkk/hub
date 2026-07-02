package bases_test

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
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/bases"
)

// setupBasesRBAC поднимает два независимых приложения над одной БД:
// ownerApp авторизован как владелец базы, intruderApp — как другой пользователь.
func setupBasesRBAC(t *testing.T) (ownerApp, intruderApp *fiber.App, svc *Service, ownerID, intruderID uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:bases_rbac_"+uuid.NewString()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("rbac-owner@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	intruder, err := client.User.Create().
		SetEmail("rbac-intruder@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc = NewService(NewEntRepository(client))
	h := bases.NewHandler(svc)

	newApp := func(userID uuid.UUID) *fiber.App {
		app := fiber.New()
		api := app.Group("/api")
		authMw := func(c fiber.Ctx) error {
			c.Locals("userID", userID.String())
			return c.Next()
		}
		h.RegisterRoutes(api, authMw)
		return app
	}
	return newApp(owner.ID), newApp(intruder.ID), svc, owner.ID, intruder.ID
}

func createOwnerBase(t *testing.T, svc *Service, ownerID uuid.UUID, name string) *BaseDetailResponse {
	t.Helper()
	detail, err := svc.CreateBase(context.Background(), ownerID, CreateBaseRequest{Name: name})
	require.NoError(t, err)
	require.NotEmpty(t, detail.Tables)
	return detail
}

func TestRBAC_ListBases_HidesForeignBases(t *testing.T) {
	ownerApp, intruderApp, svc, ownerID, _ := setupBasesRBAC(t)
	createOwnerBase(t, svc, ownerID, "Owner Secret DB")

	req := httptest.NewRequest(http.MethodGet, "/api/bases", nil)
	resp, err := intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var list struct {
		Bases []BaseResponse `json:"bases"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	assert.Empty(t, list.Bases, "intruder must not see foreign bases")

	req = httptest.NewRequest(http.MethodGet, "/api/bases", nil)
	resp, err = ownerApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	assert.Len(t, list.Bases, 1, "owner still sees own base")
}

func TestRBAC_GetForeignBase_Forbidden(t *testing.T) {
	_, intruderApp, svc, ownerID, _ := setupBasesRBAC(t)
	detail := createOwnerBase(t, svc, ownerID, "Private")

	req := httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String(), nil)
	resp, err := intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRBAC_UpdateForeignBase_Forbidden(t *testing.T) {
	_, intruderApp, svc, ownerID, _ := setupBasesRBAC(t)
	detail := createOwnerBase(t, svc, ownerID, "Private")

	body, _ := json.Marshal(map[string]any{"name": "Hacked", "version": detail.Version})
	req := httptest.NewRequest(http.MethodPut, "/api/bases/"+detail.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRBAC_DeleteForeignBase_Forbidden(t *testing.T) {
	ownerApp, intruderApp, svc, ownerID, _ := setupBasesRBAC(t)
	detail := createOwnerBase(t, svc, ownerID, "Private")

	req := httptest.NewRequest(http.MethodDelete, "/api/bases/"+detail.ID.String(), nil)
	resp, err := intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// База не тронута: владелец продолжает её видеть.
	req = httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String(), nil)
	resp, err = ownerApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRBAC_CreateTableInForeignBase_Forbidden(t *testing.T) {
	_, intruderApp, svc, ownerID, _ := setupBasesRBAC(t)
	detail := createOwnerBase(t, svc, ownerID, "Private")

	body, _ := json.Marshal(map[string]string{"name": "Injected"})
	req := httptest.NewRequest(http.MethodPost, "/api/bases/"+detail.ID.String()+"/tables", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRBAC_ForeignTableViewColumnsRecords_Forbidden(t *testing.T) {
	_, intruderApp, svc, ownerID, _ := setupBasesRBAC(t)
	detail := createOwnerBase(t, svc, ownerID, "Private")
	tableID := detail.Tables[0].ID
	basePath := "/api/bases/" + detail.ID.String() + "/tables/" + tableID.String()

	req := httptest.NewRequest(http.MethodGet, basePath, nil)
	resp, err := intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "table view")

	colBody, _ := json.Marshal(map[string]string{"name": "Leak", "col_type": "text"})
	req = httptest.NewRequest(http.MethodPost, basePath+"/columns", bytes.NewReader(colBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "create column")

	recBody, _ := json.Marshal(map[string]any{"cells": map[string]any{}})
	req = httptest.NewRequest(http.MethodPost, basePath+"/records", bytes.NewReader(recBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "create record")

	req = httptest.NewRequest(http.MethodPut, basePath+"/records/"+uuid.NewString(), bytes.NewReader(recBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "update record")

	req = httptest.NewRequest(http.MethodDelete, basePath+"/records/"+uuid.NewString(), nil)
	resp, err = intruderApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "delete record")
}

func TestRBAC_OwnerFullAccess_Positive(t *testing.T) {
	ownerApp, _, svc, ownerID, _ := setupBasesRBAC(t)
	detail := createOwnerBase(t, svc, ownerID, "Mine")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String(), nil)
	resp, err := ownerApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String(), nil)
	resp, err = ownerApp.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var view TableViewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	resp.Body.Close()
	require.NotEmpty(t, view.Columns)

	colID := view.Columns[0].ID.String()
	recBody, _ := json.Marshal(map[string]any{"cells": map[string]any{colID: "row"}})
	req = httptest.NewRequest(http.MethodPost,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records",
		bytes.NewReader(recBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = ownerApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestRBAC_ServiceLevel_ForbiddenVsNotFound(t *testing.T) {
	_, _, svc, ownerID, intruderID := setupBasesRBAC(t)
	detail := createOwnerBase(t, svc, ownerID, "Svc")
	ctx := context.Background()

	_, err := svc.GetBase(ctx, intruderID, detail.ID)
	assert.ErrorIs(t, err, bases.ErrForbidden)

	err = svc.DeleteBase(ctx, intruderID, detail.ID)
	assert.ErrorIs(t, err, bases.ErrForbidden)

	// Несуществующая база остаётся 404, а не 403 — не раскрываем существование.
	_, err = svc.GetBase(ctx, intruderID, uuid.New())
	assert.ErrorIs(t, err, bases.ErrNotFound)

	assert.Equal(t, http.StatusForbidden, bases.ExportFiberStatusFromError(mapServiceError(bases.ErrForbidden)))
}
