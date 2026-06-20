package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userMgmtMockRepo struct {
	adminMockRepo
	restoreType string
	restoreID   uuid.UUID
	reassignFrom uuid.UUID
	reassignTo   uuid.UUID
	reassignCount int
	gdprUserID    uuid.UUID
	gdprReassign  *uuid.UUID
}

func (m *userMgmtMockRepo) RestoreIssue(_ context.Context, id uuid.UUID) error {
	m.restoreType = "issue"
	m.restoreID = id
	return nil
}

func (m *userMgmtMockRepo) ReassignAssignedIssues(_ context.Context, from, to uuid.UUID) (int, error) {
	m.reassignFrom = from
	m.reassignTo = to
	return m.reassignCount, nil
}

func (m *userMgmtMockRepo) HardDeleteUser(_ context.Context, userID uuid.UUID) error {
	m.gdprUserID = userID
	return nil
}

func (m *userMgmtMockRepo) DeleteUserComments(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *userMgmtMockRepo) DeleteUserReactions(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *userMgmtMockRepo) DeleteUserActivityLogs(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *userMgmtMockRepo) DeleteUserTimeLogs(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *userMgmtMockRepo) ReassignReportedIssues(_ context.Context, _, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *userMgmtMockRepo) ClearUserAssignments(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}

func TestHandler_DeactivateUser_AdminOK(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+targetID.String()+"/deactivate", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, repo.deactivateIDs, targetID)
}

func TestHandler_DeactivateUser_InvalidID(t *testing.T) {
	adminID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/not-uuid/deactivate", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_RestoreItem_AdminOK(t *testing.T) {
	adminID := uuid.New()
	itemID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/trash/issue/"+itemID.String()+"/restore", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, itemID, repo.restoreID)
}

func TestHandler_ReassignTasks_AdminOK(t *testing.T) {
	adminID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	repo := &userMgmtMockRepo{
		adminMockRepo: adminMockRepo{adminID: adminID},
		reassignCount: 3,
	}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users/"+fromID.String()+"/reassign",
		strings.NewReader(`{"to_user_id":"`+toID.String()+`"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, fromID, repo.reassignFrom)
	assert.Equal(t, toID, repo.reassignTo)
}

func TestHandler_ReassignTasks_InvalidBody(t *testing.T) {
	adminID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users/"+uuid.New().String()+"/reassign",
		strings.NewReader(`{`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_DeleteUserGDPR_AdminOK(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	reassignID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	url := "/api/admin/users/" + targetID.String() + "/gdpr?reassign_to=" + reassignID.String()
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, targetID, repo.gdprUserID)
}

func TestHandler_DeleteUserGDPR_InvalidReassignTo(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+targetID.String()+"/gdpr?reassign_to=bad", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
