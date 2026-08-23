package audit_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/audit"
)

type mockAuditRepo struct {
	audit.Repository
	logs           []*ent.AuditLog
	lastListFilter audit.AuditLogFilters
	settings       map[string]string
	deactivateIDs  []uuid.UUID
	deleteSessIDs  []uuid.UUID
	deleteSessErr  error
}

func (m *mockAuditRepo) CreateAuditLog(_ context.Context, action, entityType, entityID string, userID, projectID *uuid.UUID, details map[string]interface{}, ipAddress, userAgent string) (*ent.AuditLog, error) {
	log := &ent.AuditLog{
		ID:         uuid.New(),
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		UserID:     userID,
		ProjectID:  projectID,
		Details:    details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now().UTC(),
	}
	m.logs = append(m.logs, log)
	return log, nil
}

func (m *mockAuditRepo) ListAuditLogs(_ context.Context, filters audit.AuditLogFilters) ([]*ent.AuditLog, int, error) {
	m.lastListFilter = filters
	return m.logs, len(m.logs), nil
}

func (m *mockAuditRepo) GetSystemSetting(_ context.Context, key string) (string, error) {
	return m.settings[key], nil
}

func (m *mockAuditRepo) GetSystemSettingsBatch(_ context.Context, keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value := m.settings[key]; value != "" {
			values[key] = value
		}
	}
	return values, nil
}

func (m *mockAuditRepo) UpsertSystemSetting(_ context.Context, key, value, _ string) error {
	if m.settings == nil {
		m.settings = make(map[string]string)
	}
	m.settings[key] = value
	return nil
}

func (m *mockAuditRepo) ListDeletedIssues(_ context.Context, _ *uuid.UUID, _ time.Time, _, _ int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}

func (m *mockAuditRepo) DeactivateUser(_ context.Context, userID uuid.UUID) error {
	m.deactivateIDs = append(m.deactivateIDs, userID)
	return nil
}

func (m *mockAuditRepo) DeleteUserSessions(_ context.Context, userID uuid.UUID) (int, error) {
	m.deleteSessIDs = append(m.deleteSessIDs, userID)
	if m.deleteSessErr != nil {
		return 0, m.deleteSessErr
	}
	return 1, nil
}

func (*mockAuditRepo) GetUserSystemRole(context.Context, uuid.UUID) (user.SystemRole, error) {
	return user.SystemRole("member"), nil
}

func (*mockAuditRepo) IsSystemAdmin(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}

var _ audit.Repository = (*mockAuditRepo)(nil)

type gdprStubRepo struct {
	mockAuditRepo
}

func (r *gdprStubRepo) WithTx(_ context.Context, fn func(audit.Repository) error) error {
	return fn(r)
}

func (*gdprStubRepo) ClearUserAssignments(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserComments(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserReactions(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserActivityLogs(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserTimeLogs(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserFiles(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserProjectMemberships(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserChannelMemberships(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) DeleteUserTeamMemberships(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) ScrubUserAuditPII(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}

func (*gdprStubRepo) HardDeleteUser(context.Context, uuid.UUID) error {
	return nil
}
