package audit_test

// Тесты HTTP-обработчика аудита: интеграционные и unit-сценарии testframework.
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

	"github.com/valerii/hub/backend/internal/audit"
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

func (m *userMgmtMockRepo) WithTx(_ context.Context, fn func(audit.Repository) error) error {
	return fn(m)
}

// TestHandler_DeactivateUser_AdminOK проверяет HTTP-слой: HTTP-обработчик deactivate user админ ok.
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

// TestHandler_DeactivateUser_InvalidID проверяет отклонение при HTTP-обработчик deactivate user некорректный id.
func TestHandler_DeactivateUser_InvalidID(t *testing.T) {
	adminID := uuid.New()
	repo := &userMgmtMockRepo{adminMockRepo: adminMockRepo{adminID: adminID}}
	app := setupAuditHandlerTest(t, repo, adminID)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/not-uuid/deactivate", nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_RestoreItem_AdminOK проверяет HTTP-слой: HTTP-обработчик restore item админ ok.
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

// TestHandler_ReassignTasks_AdminOK проверяет HTTP-слой: HTTP-обработчик reassign tasks админ ok.
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

// TestHandler_ReassignTasks_InvalidBody проверяет отклонение при HTTP-обработчик reassign tasks некорректный body.
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

// TestHandler_DeleteUserGDPR_AdminOK проверяет HTTP-слой: HTTP-обработчик удаление user gdpr админ ok.
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

// TestHandler_DeleteUserGDPR_InvalidReassignTo проверяет отклонение при HTTP-обработчик удаление user gdpr некорректный reassign to.
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
