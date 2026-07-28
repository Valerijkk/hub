package calls_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/internal/calls"
	"github.com/valerii/hub/backend/internal/platform/config"
)

// --- Service-layer denial (guard) ---

// TestScheduledCall_Create_RejectsStranger_P1_BE_22 — W3-A5-003.
// Mutation: убрать requireChannelMember в Create → stranger создаст запись.
func TestScheduledCall_Create_RejectsStranger_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("AuthzC").SetKey("AZC").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	_, err = scheduled.CreateScheduledCall(ctx, proj.ID, f.stranger.ID, calls.CreateScheduledCallRequest{
		Title: "Pwn", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.ErrorIs(t, err, calls.ErrCallForbidden)
}

// TestScheduledCall_Create_RejectsProjectMismatch_P1_BE_22 — channel не из path project.
func TestScheduledCall_Create_RejectsProjectMismatch_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	projA, err := f.client.Project.Create().SetName("AuthzA").SetKey("AZA").Save(ctx)
	require.NoError(t, err)
	projB, err := f.client.Project.Create().SetName("AuthzB").SetKey("AZB").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, projA.ID)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	_, err = scheduled.CreateScheduledCall(ctx, projB.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Wrong proj", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.ErrorIs(t, err, calls.ErrCallForbidden)
}

// TestScheduledCall_List_RejectsStranger_P1_BE_22 — W3-A5-004.
func TestScheduledCall_List_RejectsStranger_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("AuthzL").SetKey("AZL").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	_, err = scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Secret", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	_, err = scheduled.ListScheduledCalls(ctx, f.channel.ID, f.stranger.ID)
	require.ErrorIs(t, err, calls.ErrCallForbidden)
}

// TestScheduledCall_Update_RejectsStranger_P1_BE_22 — W3-A5-001.
// Mutation: убрать authorizeScheduledCallMutate → stranger обновит title.
func TestScheduledCall_Update_RejectsStranger_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("AuthzU").SetKey("AZU").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Mine", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	newTitle := "Hijacked"
	_, err = scheduled.UpdateScheduledCall(ctx, created.ID, f.stranger.ID, calls.UpdateScheduledCallRequest{Title: &newTitle})
	require.ErrorIs(t, err, calls.ErrCallForbidden)

	got, err := f.client.ScheduledCall.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Mine", got.Title)
}

// TestScheduledCall_Update_RejectsPlainMember_P1_BE_22 — member ≠ creator/owner/admin.
func TestScheduledCall_Update_RejectsPlainMember_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("AuthzM").SetKey("AZM").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "OwnerCall", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	newTitle := "MemberEdit"
	_, err = scheduled.UpdateScheduledCall(ctx, created.ID, f.member.ID, calls.UpdateScheduledCallRequest{Title: &newTitle})
	require.ErrorIs(t, err, calls.ErrCallForbidden)
}

// TestScheduledCall_Delete_RejectsStranger_P1_BE_22 — W3-A5-002; deleted_at остаётся NULL.
func TestScheduledCall_Delete_RejectsStranger_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := context.Background()
	proj, err := f.client.Project.Create().SetName("AuthzD").SetKey("AZD").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	scheduled := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduled.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Keep", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	err = scheduled.DeleteScheduledCall(ctx, created.ID, f.stranger.ID)
	require.ErrorIs(t, err, calls.ErrCallForbidden)

	got, err := f.client.ScheduledCall.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt)
	assert.True(t, got.IsActive)
}

// --- HTTP denial suite ---

func newScheduledAuthzApp(f *callsTestFixture) *fiber.App {
	scheduledSvc := calls.NewScheduledCallService(f.repo, nil)
	h := calls.ExportNewHandlerWithServiceAndScheduled(f.svc, scheduledSvc, config.LiveKitConfig{
		URL: "http://localhost:7880", APIKey: "testkey", APISecret: "testsecretthatshouldbelongenough",
	})
	return newCallsValidationApp(h)
}

// TestHandler_ScheduledCall_StrangerUpdate_403_P1_BE_22 — критерий мутации TASK-018.
func TestHandler_ScheduledCall_StrangerUpdate_403_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("HTTP Upd").SetKey("HUP").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	scheduledSvc := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduledSvc.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "Daily", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(2 * time.Hour),
	})
	require.NoError(t, err)

	app := newScheduledAuthzApp(f)
	body := []byte(`{"title":"Hijacked"}`)
	req := httptest.NewRequest(fiber.MethodPut, "/scheduled-calls/"+created.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_ScheduledCall_StrangerDelete_403_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("HTTP Del").SetKey("HDL").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	scheduledSvc := calls.NewScheduledCallService(f.repo, nil)
	created, err := scheduledSvc.CreateScheduledCall(ctx, proj.ID, f.owner.ID, calls.CreateScheduledCallRequest{
		Title: "KeepHTTP", ChannelID: f.channel.ID, ScheduledAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	app := newScheduledAuthzApp(f)
	req := httptest.NewRequest(fiber.MethodDelete, "/scheduled-calls/"+created.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	got, err := f.client.ScheduledCall.Query().Where(scheduledcall.ID(created.ID)).Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt)
}

func TestHandler_ScheduledCall_StrangerList_403_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	app := newScheduledAuthzApp(f)

	req := httptest.NewRequest(fiber.MethodGet, "/channels/"+f.channel.ID.String()+"/scheduled-calls", nil)
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_ScheduledCall_StrangerCreate_403_P1_BE_22(t *testing.T) {
	f := newCallsTestFixture(t)
	ctx := t.Context()
	proj, err := f.client.Project.Create().SetName("HTTP Crt").SetKey("HCR").Save(ctx)
	require.NoError(t, err)
	f.linkChannelToProject(t, proj.ID)

	app := newScheduledAuthzApp(f)
	when := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	payload, _ := json.Marshal(map[string]any{
		"title":        "Nope",
		"channel_id":   f.channel.ID.String(),
		"scheduled_at": when,
	})
	req := httptest.NewRequest(fiber.MethodPost, "/projects/"+proj.ID.String()+"/scheduled-calls", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", f.stranger.ID.String())
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
