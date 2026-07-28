package audit_test

// Тесты сервиса аудита: интеграционные и unit-сценарии testframework.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
)

type mockAuditRepo struct {
	logs                   []*ent.AuditLog
	lastListFilter         audit.AuditLogFilters
	createCalls            []createAuditLogCall
	createErr              error
	deactivateErr          error
	deactivateIDs          []uuid.UUID
	deleteSessIDs          []uuid.UUID
	reassignAssignedCalls  [][2]uuid.UUID
	reassignReportedCalls  [][2]uuid.UUID
	reassignAssignedN      int
	reassignErr            error
}

type createAuditLogCall struct {
	action, entityType, entityID string
	userID                       *uuid.UUID
	details                      map[string]interface{}
	ipAddress, userAgent         string
}

func (m *mockAuditRepo) CreateAuditLog(_ context.Context, action, entityType, entityID string, userID, _ *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) (*ent.AuditLog, error) {
	m.createCalls = append(m.createCalls, createAuditLogCall{
		action: action, entityType: entityType, entityID: entityID,
		userID: userID, details: details, ipAddress: ipAddress, userAgent: userAgent,
	})
	if m.createErr != nil {
		return nil, m.createErr
	}
	log := &ent.AuditLog{
		ID: uuid.New(), Action: action, EntityType: entityType, EntityID: entityID,
		Details: details, IPAddress: ipAddress, UserAgent: userAgent, CreatedAt: time.Now().UTC(),
	}
	if userID != nil {
		uid := *userID
		log.UserID = &uid
	}
	m.logs = append(m.logs, log)
	return log, nil
}

func (m *mockAuditRepo) ListAuditLogs(_ context.Context, filters audit.AuditLogFilters) ([]*ent.AuditLog, int, error) {
	m.lastListFilter = filters
	var matched []*ent.AuditLog
	for _, log := range m.logs {
		if filters.EntityID != "" && log.EntityID != filters.EntityID {
			continue
		}
		if filters.IPAddress != "" && !strings.Contains(log.IPAddress, filters.IPAddress) {
			continue
		}
		if filters.Action != "" && log.Action != filters.Action {
			continue
		}
		if filters.SlowOnly {
			if log.EntityType != "http" {
				continue
			}
			dur, ok := log.Details["duration_ms"]
			if !ok {
				continue
			}
			var ms float64
			switch v := dur.(type) {
			case int64:
				ms = float64(v)
			case int:
				ms = float64(v)
			case float64:
				ms = v
			default:
				continue
			}
			if ms <= float64(audit.ExportSlowRequestThresholdMS()) {
				continue
			}
		}
		matched = append(matched, log)
	}
	page := filters.Page
	if page < 1 {
		page = 1
	}
	perPage := filters.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 50
	}
	offset := (page - 1) * perPage
	total := len(matched)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + perPage
	if end > total {
		end = total
	}
	return matched[offset:end], total, nil
}

func (m *mockAuditRepo) DeactivateUser(_ context.Context, userID uuid.UUID) error {
	m.deactivateIDs = append(m.deactivateIDs, userID)
	return m.deactivateErr
}

func (m *mockAuditRepo) DeleteUserSessions(_ context.Context, userID uuid.UUID) (int, error) {
	m.deleteSessIDs = append(m.deleteSessIDs, userID)
	return 1, nil
}

func (m *mockAuditRepo) GetSystemSetting(context.Context, string) (string, error) { return "", nil }
func (m *mockAuditRepo) GetSystemSettingsBatch(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (m *mockAuditRepo) UpsertSystemSetting(context.Context, string, string, string) error {
	return nil
}
func (m *mockAuditRepo) ListDeletedIssues(context.Context, *uuid.UUID, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (m *mockAuditRepo) ListDeletedTestCases(context.Context, *uuid.UUID, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (m *mockAuditRepo) ListDeletedFiles(context.Context, *uuid.UUID, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (m *mockAuditRepo) ListDeletedProjects(context.Context, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (m *mockAuditRepo) ListDeletedNewsPosts(context.Context, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (m *mockAuditRepo) ListDeletedScheduledCalls(context.Context, *uuid.UUID, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (m *mockAuditRepo) RestoreIssue(context.Context, uuid.UUID) error   { return nil }
func (m *mockAuditRepo) RestoreTestCase(context.Context, uuid.UUID) error { return nil }
func (m *mockAuditRepo) RestoreFile(context.Context, uuid.UUID) error    { return nil }
func (m *mockAuditRepo) RestoreProject(context.Context, uuid.UUID) error { return nil }
func (m *mockAuditRepo) RestoreNewsPost(context.Context, uuid.UUID) error { return nil }
func (m *mockAuditRepo) RestoreScheduledCall(context.Context, uuid.UUID) error { return nil }
func (m *mockAuditRepo) SoftDeleteIssue(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (m *mockAuditRepo) SoftDeleteTestCase(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (m *mockAuditRepo) SoftDeleteFile(context.Context, uuid.UUID, time.Time) error { return nil }
func (m *mockAuditRepo) SoftDeleteProject(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (m *mockAuditRepo) SoftDeleteNewsPost(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (m *mockAuditRepo) SoftDeleteScheduledCall(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (m *mockAuditRepo) PurgeExpiredIssues(context.Context, time.Time) (int, error)    { return 0, nil }
func (m *mockAuditRepo) PurgeExpiredTestCases(context.Context, time.Time) (int, error) { return 0, nil }
func (m *mockAuditRepo) PurgeExpiredFiles(context.Context, time.Time) (int, error)     { return 0, nil }
func (m *mockAuditRepo) PurgeExpiredProjects(context.Context, time.Time) (int, error)  { return 0, nil }
func (m *mockAuditRepo) PurgeExpiredNewsPosts(context.Context, time.Time) (int, error) { return 0, nil }
func (m *mockAuditRepo) PurgeExpiredScheduledCalls(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuditRepo) ReassignAssignedIssues(_ context.Context, from, to uuid.UUID) (int, error) {
	m.reassignAssignedCalls = append(m.reassignAssignedCalls, [2]uuid.UUID{from, to})
	if m.reassignAssignedN != 0 {
		return m.reassignAssignedN, m.reassignErr
	}
	return 0, m.reassignErr
}
func (m *mockAuditRepo) ReassignReportedIssues(_ context.Context, from, to uuid.UUID) (int, error) {
	m.reassignReportedCalls = append(m.reassignReportedCalls, [2]uuid.UUID{from, to})
	return 0, nil
}
func (m *mockAuditRepo) ClearUserAssignments(context.Context, uuid.UUID) (int, error) { return 0, nil }
func (m *mockAuditRepo) DeleteUserComments(context.Context, uuid.UUID) (int, error)   { return 0, nil }
func (m *mockAuditRepo) DeleteUserReactions(context.Context, uuid.UUID) (int, error)  { return 0, nil }
func (m *mockAuditRepo) DeleteUserActivityLogs(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockAuditRepo) DeleteUserTimeLogs(context.Context, uuid.UUID) (int, error) { return 0, nil }
func (m *mockAuditRepo) DeleteUserFiles(context.Context, uuid.UUID) (int, error)    { return 0, nil }
func (m *mockAuditRepo) DeleteUserProjectMemberships(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockAuditRepo) DeleteUserChannelMemberships(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockAuditRepo) DeleteUserTeamMemberships(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockAuditRepo) ScrubUserAuditPII(context.Context, uuid.UUID) (int, error) { return 0, nil }
func (m *mockAuditRepo) HardDeleteUser(context.Context, uuid.UUID) error            { return nil }
func (m *mockAuditRepo) IsSystemAdmin(context.Context, uuid.UUID) (bool, error)     { return false, nil }
func (m *mockAuditRepo) GetUserSystemRole(context.Context, uuid.UUID) (user.SystemRole, error) {
	return user.SystemRoleMember, nil
}

// TestBanUser проверяет сценарий: ban user.
func TestBanUser(t *testing.T) {
	adminID := uuid.New()
	targetID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name           string
		reason         string
		target         uuid.UUID
		admin          uuid.UUID
		deactErr       error
		createErr      error
		wantErr        string
		wantDeact      bool
		wantSessRevoke bool
		wantAudit      bool
	}{
		{
			name:    "empty reason rejected",
			reason:  "",
			target:  targetID,
			admin:   adminID,
			wantErr: "укажите причину блокировки",
		},
		{
			name:    "self-ban rejected",
			reason:  "abuse",
			target:  adminID,
			admin:   adminID,
			wantErr: "нельзя заблокировать самого себя",
		},
		{
			name:      "deactivate failure propagates",
			reason:    "spam",
			target:    targetID,
			admin:     adminID,
			deactErr:  errors.New("db down"),
			wantErr:   "деактивация пользователя",
			wantDeact: true,
		},
		{
			name:           "success deactivates and audits",
			reason:         "harassment",
			target:         targetID,
			admin:          adminID,
			wantDeact:      true,
			wantSessRevoke: true,
			wantAudit:      true,
		},
		{
			// W6-A4-015 / P3-BE-01: CreateAuditLog fail после deactivate → err caller.
			name:      "audit log failure after ban",
			reason:    "spam",
			target:    targetID,
			admin:     adminID,
			createErr: errors.New("audit db down"),
			wantErr:   "audit ban log",
			wantDeact: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuditRepo{deactivateErr: tt.deactErr, createErr: tt.createErr}
			svc := audit.NewAuditService(repo, nil)

			err := svc.BanUser(ctx, tt.target, tt.admin, tt.reason, "10.0.0.1", "test-agent")
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantDeact {
				require.Contains(t, repo.deactivateIDs, tt.target)
			} else {
				assert.Empty(t, repo.deactivateIDs)
			}
			if tt.wantSessRevoke {
				require.Contains(t, repo.deleteSessIDs, tt.target)
			}

			if tt.wantAudit {
				require.Len(t, repo.createCalls, 1)
				call := repo.createCalls[0]
				assert.Equal(t, "ban", call.action)
				assert.Equal(t, "user", call.entityType)
				assert.Equal(t, tt.target.String(), call.entityID)
				require.NotNil(t, call.userID)
				assert.Equal(t, tt.admin, *call.userID)
				assert.Equal(t, tt.reason, call.details["reason"])
				assert.Equal(t, tt.admin.String(), call.details["banned_by"])
				assert.Equal(t, "10.0.0.1", call.ipAddress)
				assert.Equal(t, "test-agent", call.userAgent)
			}
		})
	}
}

// TestRecordHTTPRequest проверяет сценарий: record http request.
func TestRecordHTTPRequest(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name       string
		status     int
		wantAction string
	}{
		{"success maps to request", 200, "request"},
		{"redirect maps to request", 302, "request"},
		{"client error maps to request.error", 404, "request.error"},
		{"server error maps to request.error", 500, "request.error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuditRepo{}
			svc := audit.NewAuditService(repo, nil)

			svc.RecordHTTPRequest(ctx, audit.HTTPAuditEntry{
				Method: "GET", Path: "/api/issues", Status: tt.status,
				DurationMS: 42, RequestID: "req-1", UserID: &userID,
				IPAddress: "192.168.1.5", UserAgent: "curl/8.0",
			})

			require.Len(t, repo.createCalls, 1)
			call := repo.createCalls[0]
			assert.Equal(t, tt.wantAction, call.action)
			assert.Equal(t, "http", call.entityType)
			assert.Equal(t, "/api/issues", call.entityID)
			assert.Equal(t, "GET", call.details["method"])
			assert.Equal(t, "/api/issues", call.details["path"])
			assert.Equal(t, tt.status, call.details["status"])
			assert.Equal(t, int64(42), call.details["duration_ms"])
			assert.Equal(t, "req-1", call.details["request_id"])
			require.NotNil(t, call.userID)
			assert.Equal(t, userID, *call.userID)
		})
	}
}

// TestListAuditLogs_FiltersIPAndEntityID проверяет сценарий: список audit logs filters ip and entity id.
func TestListAuditLogs_FiltersIPAndEntityID(t *testing.T) {
	ctx := context.Background()
	entityA := uuid.New().String()
	entityB := uuid.New().String()

	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)

	_, _ = repo.CreateAuditLog(ctx, "login", "session", entityA, nil, nil, nil, "10.1.2.3", "")
	_, _ = repo.CreateAuditLog(ctx, "login", "session", entityB, nil, nil, nil, "10.9.9.9", "")
	_, _ = repo.CreateAuditLog(ctx, "logout", "session", entityA, nil, nil, nil, "10.1.2.99", "")

	tests := []struct {
		name      string
		filters   audit.AuditLogFilters
		wantTotal int
		wantIDs   []string
	}{
		{
			name:      "filter by entity_id",
			filters:   audit.AuditLogFilters{EntityID: entityA, Page: 1, PerPage: 50},
			wantTotal: 2,
			wantIDs:   []string{entityA, entityA},
		},
		{
			name:      "filter by ip substring",
			filters:   audit.AuditLogFilters{IPAddress: "10.1.2", Page: 1, PerPage: 50},
			wantTotal: 2,
		},
		{
			name:      "combined entity_id and ip",
			filters:   audit.AuditLogFilters{EntityID: entityA, IPAddress: "10.1.2.3", Page: 1, PerPage: 50},
			wantTotal: 1,
			wantIDs:   []string{entityA},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.ListAuditLogs(ctx, tt.filters)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, resp.Total)
			assert.Equal(t, tt.filters.EntityID, repo.lastListFilter.EntityID)
			assert.Equal(t, tt.filters.IPAddress, repo.lastListFilter.IPAddress)
			if tt.wantIDs != nil {
				got := make([]string, 0, len(resp.Items))
				for _, item := range resp.Items {
					got = append(got, item.EntityID)
				}
				assert.ElementsMatch(t, tt.wantIDs, got)
			}
		})
	}
}

// TestListAuditLogs_PaginationDefaults проверяет сценарий: список audit logs pagination defaults.
func TestListAuditLogs_PaginationDefaults(t *testing.T) {
	ctx := context.Background()
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)

	for i := 0; i < 3; i++ {
		_, _ = repo.CreateAuditLog(ctx, "tick", "http", fmt.Sprintf("/api/%d", i), nil, nil, nil, "", "")
	}

	resp, err := svc.ListAuditLogs(ctx, audit.AuditLogFilters{Page: 0, PerPage: 0})
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 50, resp.PerPage)
	assert.Len(t, resp.Items, 3)
}

// TestListAuditLogs_SlowOnly проверяет сценарий: список audit logs slow only.
func TestListAuditLogs_SlowOnly(t *testing.T) {
	ctx := context.Background()
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)

	_, _ = repo.CreateAuditLog(ctx, "request", "http", "/slow", nil, nil,
		map[string]interface{}{"duration_ms": int64(900)}, "", "")
	_, _ = repo.CreateAuditLog(ctx, "request", "http", "/fast", nil, nil,
		map[string]interface{}{"duration_ms": int64(50)}, "", "")
	_, _ = repo.CreateAuditLog(ctx, "login", "session", "sess-1", nil, nil,
		map[string]interface{}{"duration_ms": int64(900)}, "", "")

	resp, err := svc.ListAuditLogs(ctx, audit.AuditLogFilters{SlowOnly: true, Page: 1, PerPage: 50})
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "/slow", resp.Items[0].EntityID)
	assert.True(t, repo.lastListFilter.SlowOnly)
}

// TestRecordHTTPRequest_SkipsOnCreateError — W6-A5-009 / P3-BE-02: inject createErr,
// create вызван ровно 1 раз, no panic (ошибка глотается после slog).
func TestRecordHTTPRequest_SkipsOnCreateError(t *testing.T) {
	repo := &mockAuditRepo{createErr: errors.New("db down")}
	svc := audit.NewAuditService(repo, nil)

	assert.NotPanics(t, func() {
		svc.RecordHTTPRequest(context.Background(), audit.HTTPAuditEntry{
			Method: "POST", Path: "/api/x", Status: 201, DurationMS: 10,
		})
	})
	require.Len(t, repo.createCalls, 1)
	assert.Empty(t, repo.logs, "failed Create must not append to in-memory logs")
}

// TestRecordHTTPRequest_NilUserID проверяет сценарий: record http request nil user id.
func TestRecordHTTPRequest_NilUserID(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := audit.NewAuditService(repo, nil)

	svc.RecordHTTPRequest(context.Background(), audit.HTTPAuditEntry{
		Method: "GET", Path: "/api/health", Status: 200, DurationMS: 1,
	})
	require.Len(t, repo.createCalls, 1)
	assert.Nil(t, repo.createCalls[0].userID)
}
