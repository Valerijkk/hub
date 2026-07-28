package onec_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	hrmod "github.com/valerii/hub/backend/internal/hr"
	"github.com/valerii/hub/backend/internal/integrations/onec"
)

func TestConfig_validateAndDecode(t *testing.T) {
	assert.NoError(t, onec.ExportConfigValidate(onec.Config{Enabled: false}))
	assert.ErrorIs(t, onec.ExportConfigValidate(onec.Config{Enabled: true, WebhookURL: ""}), onec.ErrInvalidConfig)
	assert.ErrorIs(t, onec.ExportConfigValidate(onec.Config{Enabled: true, WebhookURL: "https://x.com", EmployeeIDField: "bad"}), onec.ErrInvalidConfig)

	cfg, err := onec.ExportDecodeConfig("")
	require.NoError(t, err)
	assert.False(t, cfg.Enabled)
	assert.Equal(t, onec.EmployeeIDEmail, cfg.EmployeeIDField)

	raw, err := onec.ExportEncodeConfig(onec.Config{Enabled: true, WebhookURL: "https://1c.example/hook"})
	require.NoError(t, err)
	cfg, err = onec.ExportDecodeConfig(raw)
	require.NoError(t, err)
	assert.True(t, cfg.Enabled)
}

func TestService_UpdateSettings(t *testing.T) {
	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{}, nil)

	enabled := true
	url := "https://1c.example.local/webhook"
	field := onec.EmployeeIDPhone
	token := "secret-token"
	resp, err := svc.UpdateSettings(context.Background(), onec.UpdateSettingsRequest{
		Enabled:         &enabled,
		WebhookURL:      &url,
		BearerToken:     &token,
		EmployeeIDField: &field,
	})
	require.NoError(t, err)
	assert.True(t, resp.Enabled)
	assert.Equal(t, url, resp.WebhookURL)
	assert.True(t, resp.BearerTokenSet)
	assert.Equal(t, onec.EmployeeIDPhone, resp.EmployeeIDField)

	badField := onec.EmployeeIDField("nope")
	_, err = svc.UpdateSettings(context.Background(), onec.UpdateSettingsRequest{EmployeeIDField: &badField})
	assert.ErrorIs(t, err, onec.ErrInvalidConfig)

	enabledOnly := true
	fresh := onec.NewService(&memStore{data: map[string]string{}}, stubUsers{}, nil)
	_, err = fresh.UpdateSettings(context.Background(), onec.UpdateSettingsRequest{Enabled: &enabledOnly})
	assert.ErrorIs(t, err, onec.ErrInvalidConfig)
}

func TestService_TestWebhook(t *testing.T) {
	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{}, nil)
	onec.ExportServiceSetClient(svc, &http.Client{Timeout: 5 * time.Second})

	_, err := svc.TestWebhook(context.Background())
	assert.ErrorIs(t, err, onec.ErrNotConfigured)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	cfg := onec.Config{Enabled: true, WebhookURL: srv.URL, EmployeeIDField: onec.EmployeeIDEmail}
	raw, _ := onec.ExportEncodeConfig(cfg)
	store.data[onec.ExportSettingsKey()] = raw

	resp, err := svc.TestWebhook(context.Background())
	require.NoError(t, err)
	assert.True(t, resp.OK)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(failSrv.Close)

	cfg.WebhookURL = failSrv.URL
	raw, _ = onec.ExportEncodeConfig(cfg)
	store.data[onec.ExportSettingsKey()] = raw
	resp, err = svc.TestWebhook(context.Background())
	require.NoError(t, err)
	assert.False(t, resp.OK)
}

func TestService_ExportApprovedLeave_disabledSkips(t *testing.T) {
	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{}, nil)
	err := svc.ExportApprovedLeave(context.Background(), hrmod.LeaveRequestResponse{ID: uuid.New()}, uuid.New())
	require.NoError(t, err)
}

func TestService_ExportApprovedLeave_badStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{}, nil)
	onec.ExportServiceSetClient(svc, &http.Client{Timeout: 5 * time.Second})
	cfg := onec.Config{Enabled: true, WebhookURL: srv.URL, EmployeeIDField: onec.EmployeeIDHubUUID}
	raw, _ := onec.ExportEncodeConfig(cfg)
	store.data[onec.ExportSettingsKey()] = raw

	err := svc.ExportApprovedLeave(context.Background(), hrmod.LeaveRequestResponse{
		ID:          uuid.New(),
		RequesterID: uuid.New(),
		LeaveType:   hrmod.LeaveSick,
		StartDate:   "2026-08-01",
		EndDate:     "2026-08-03",
	}, uuid.New())
	assert.ErrorIs(t, err, onec.ErrWebhookFailed)
}

func TestService_ExportApprovedLeave_phoneExternalID(t *testing.T) {
	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{}, nil)
	onec.ExportServiceSetClient(svc, &http.Client{Timeout: 5 * time.Second})
	cfg := onec.Config{
		Enabled:         true,
		WebhookURL:      srv.URL,
		BearerToken:     "tok123",
		EmployeeIDField: onec.EmployeeIDPhone,
	}
	raw, _ := onec.ExportEncodeConfig(cfg)
	store.data[onec.ExportSettingsKey()] = raw

	reviewed := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	req := hrmod.LeaveRequestResponse{
		ID:            uuid.New(),
		RequesterID:   uuid.New(),
		RequesterName: "Dev",
		LeaveType:     hrmod.LeaveVacation,
		StartDate:     "2026-07-01",
		EndDate:       "2026-07-14",
		ReviewedAt:    &reviewed,
	}
	err := svc.ExportApprovedLeave(context.Background(), req, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Bearer tok123", authHeader)
}

func TestBuildLeavePayload_fields(t *testing.T) {
	reviewed := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	req := hrmod.LeaveRequestResponse{
		ID:          uuid.New(),
		RequesterID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		LeaveType:   hrmod.LeaveVacation,
		StartDate:   "2026-07-01",
		EndDate:     "2026-07-14",
		Reason:      "rest",
		ReviewedAt:  &reviewed,
	}
	payload := onec.ExportBuildLeavePayload(req, uuid.New(), "", "", "", onec.EmployeeIDHubUUID)
	assert.Equal(t, req.RequesterID.String(), payload.Employee.ExternalID)
	assert.NotNil(t, payload.ApprovedAt)
	assert.Equal(t, req.RequesterName, payload.Employee.DisplayName)
}

func TestAdminHandler_UpdateSettingsAndTest(t *testing.T) {
	store := &memStore{data: map[string]string{}}
	svc := onec.NewService(store, stubUsers{}, nil)
	onec.ExportServiceSetClient(svc, &http.Client{Timeout: 5 * time.Second})

	h := onec.NewAdminHandler(svc, func(c fiber.Ctx, _ uuid.UUID) (bool, error) { return true, nil })
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var ping map[string]string
		_ = json.Unmarshal(body, &ping)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	body, _ := json.Marshal(map[string]any{
		"enabled":           true,
		"webhook_url":       srv.URL,
		"employee_id_field": "email",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/integrations/onec/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/admin/integrations/onec/test", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAdminHandler_ForbiddenForNonAdmin(t *testing.T) {
	svc := onec.NewService(&memStore{data: map[string]string{}}, stubUsers{}, nil)
	h := onec.NewAdminHandler(svc, func(c fiber.Ctx, _ uuid.UUID) (bool, error) { return false, nil })
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/integrations/onec/settings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
