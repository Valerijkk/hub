package bases_test
import (
	"github.com/valerii/hub/backend/internal/bases"
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
)

func setupBasesTest(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:bases_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("base@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo)
	h := bases.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, owner.ID
}

func TestHandler_CreateBaseAndRecord(t *testing.T) {
	app, _ := setupBasesTest(t)

	body := `{"name":"Team Inventory"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bases", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var detail BaseDetailResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	require.NotEmpty(t, detail.Tables)
	assert.Equal(t, "Team Inventory", detail.Name)
	assert.Equal(t, 1, detail.TableCount)

	tableID := detail.Tables[0].ID

	req = httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var view TableViewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	require.NotEmpty(t, view.Columns)

	colID := view.Columns[0].ID.String()
	recBody, err := json.Marshal(map[string]interface{}{
		"cells": map[string]interface{}{colID: "MacBook Pro"},
	})
	require.NoError(t, err)

	req = httptest.NewRequest(
		http.MethodPost,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records",
		bytes.NewReader(recBody),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var record RecordResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&record))
	assert.Equal(t, "MacBook Pro", record.Cells[colID])
}

func TestHandler_ListBases(t *testing.T) {
	app, _ := setupBasesTest(t)

	body := `{"name":"Wiki DB"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bases", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	_, err := app.Test(req)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/api/bases", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var list struct {
		Bases []BaseResponse `json:"bases"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	assert.Len(t, list.Bases, 1)
}

func TestHandler_UpdateBaseDeleteRecord(t *testing.T) {
	app, _ := setupBasesTest(t)

	body := `{"name":"Ops DB"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bases", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var detail BaseDetailResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	tableID := detail.Tables[0].ID

	updateBody, err := json.Marshal(map[string]any{
		"name":    "Ops Database",
		"version": detail.Version,
	})
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodPut, "/api/bases/"+detail.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	var view TableViewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	colID := view.Columns[0].ID.String()

	recBody, _ := json.Marshal(map[string]any{"cells": map[string]any{colID: "Row1"}})
	req = httptest.NewRequest(
		http.MethodPost,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records",
		bytes.NewReader(recBody),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var record RecordResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&record))

	req = httptest.NewRequest(
		http.MethodDelete,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records/"+record.ID.String(),
		nil,
	)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/bases/"+detail.ID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
