package audit_test

// Тесты сервиса аудита: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/audit"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

// gdprStubRepo — минимальный stub audit.Repository для проверки BE-H6:
// ошибки каскадного удаления должны накапливаться и возвращаться через errors.Join.
type gdprStubRepo struct {
	failReactions bool
	failTimeLogs  bool
}

func (s *gdprStubRepo) CreateAuditLog(context.Context, string, string, string, *uuid.UUID, *uuid.UUID, map[string]interface{}, string, string) (*ent.AuditLog, error) {
	return &ent.AuditLog{}, nil
}
func (s *gdprStubRepo) ListAuditLogs(context.Context, audit.AuditLogFilters) ([]*ent.AuditLog, int, error) {
	return nil, 0, nil
}
func (s *gdprStubRepo) GetSystemSetting(context.Context, string) (string, error) {
	return "", nil
}
func (s *gdprStubRepo) GetSystemSettingsBatch(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (s *gdprStubRepo) UpsertSystemSetting(context.Context, string, string, string) error {
	return nil
}
func (s *gdprStubRepo) ListDeletedIssues(context.Context, *uuid.UUID, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (s *gdprStubRepo) ListDeletedTestCases(context.Context, *uuid.UUID, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (s *gdprStubRepo) ListDeletedFiles(context.Context, *uuid.UUID, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (s *gdprStubRepo) ListDeletedProjects(context.Context, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (s *gdprStubRepo) ListDeletedNewsPosts(context.Context, time.Time, int, int) ([]audit.TrashItem, int, error) {
	return nil, 0, nil
}
func (s *gdprStubRepo) RestoreIssue(context.Context, uuid.UUID) error   { return nil }
func (s *gdprStubRepo) RestoreTestCase(context.Context, uuid.UUID) error { return nil }
func (s *gdprStubRepo) RestoreFile(context.Context, uuid.UUID) error    { return nil }
func (s *gdprStubRepo) RestoreProject(context.Context, uuid.UUID) error { return nil }
func (s *gdprStubRepo) RestoreNewsPost(context.Context, uuid.UUID) error { return nil }
func (s *gdprStubRepo) SoftDeleteIssue(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (s *gdprStubRepo) SoftDeleteTestCase(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (s *gdprStubRepo) SoftDeleteFile(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (s *gdprStubRepo) SoftDeleteProject(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (s *gdprStubRepo) SoftDeleteNewsPost(context.Context, uuid.UUID, time.Time) error {
	return nil
}
func (s *gdprStubRepo) PurgeExpiredIssues(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) PurgeExpiredTestCases(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) PurgeExpiredFiles(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) PurgeExpiredProjects(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) PurgeExpiredNewsPosts(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) DeactivateUser(context.Context, uuid.UUID) error { return nil }
func (s *gdprStubRepo) DeleteUserSessions(context.Context, uuid.UUID) (int, error) {
	return 1, nil
}
func (s *gdprStubRepo) ReassignAssignedIssues(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) ReassignReportedIssues(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) ClearUserAssignments(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) DeleteUserComments(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) DeleteUserReactions(context.Context, uuid.UUID) (int, error) {
	if s.failReactions {
		return 0, fmt.Errorf("reactions db error")
	}
	return 2, nil
}
func (s *gdprStubRepo) DeleteUserActivityLogs(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (s *gdprStubRepo) DeleteUserTimeLogs(context.Context, uuid.UUID) (int, error) {
	if s.failTimeLogs {
		return 0, fmt.Errorf("time_logs db error")
	}
	return 0, nil
}
func (s *gdprStubRepo) HardDeleteUser(context.Context, uuid.UUID) error { return nil }
func (s *gdprStubRepo) IsSystemAdmin(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (s *gdprStubRepo) GetUserSystemRole(context.Context, uuid.UUID) (user.SystemRole, error) {
	return user.SystemRoleMember, nil
}

// TestDeleteUserGDPR_JoinsCascadeErrors — BE-H6: частичные ошибки каскада
// не должны проглатываться; handler должен получить combined error.
func TestDeleteUserGDPR_JoinsCascadeErrors(t *testing.T) {
	userID := uuid.New()
	svc := audit.NewAuditService(&gdprStubRepo{
		failReactions: true,
		failTimeLogs:  true,
	}, nil)

	err := svc.DeleteUserGDPR(context.Background(), uuid.New(), userID, nil)
	if err == nil {
		t.Fatal("expected joined cascade errors, got nil")
	}

	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatalf("expected errors.Join aggregate, got %T: %v", err, err)
	}
	unwrapped := joined.Unwrap()
	if len(unwrapped) != 2 {
		t.Fatalf("expected 2 cascade errors, got %d: %v", len(unwrapped), unwrapped)
	}
}

// TestDaysRemaining проверяет вычисление оставшихся дней до окончательного удаления.
func TestDaysRemaining(t *testing.T) {
	tests := []struct {
		name      string
		deletedAt time.Time
		wantMin   int
		wantMax   int
	}{
		{
			name:      "только что удалено — 30 дней",
			deletedAt: time.Now(),
			wantMin:   29,
			wantMax:   30,
		},
		{
			name:      "удалено 10 дней назад — 20 дней",
			deletedAt: time.Now().Add(-10 * 24 * time.Hour),
			wantMin:   19,
			wantMax:   20,
		},
		{
			name:      "удалено 29 дней назад — 1 день",
			deletedAt: time.Now().Add(-29 * 24 * time.Hour),
			wantMin:   0,
			wantMax:   1,
		},
		{
			name:      "удалено 31 день назад — 0 дней",
			deletedAt: time.Now().Add(-31 * 24 * time.Hour),
			wantMin:   0,
			wantMax:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := audit.ExportDaysRemaining(tt.deletedAt)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("daysRemaining() = %d, want between %d and %d", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestMapSubjectToAction проверяет маршрутизацию NATS subject в пару (действие, тип сущности).
func TestMapSubjectToAction(t *testing.T) {
	consumer := audit.ExportNewAuditConsumerForTest(nil)

	tests := []struct {
		subject    string
		wantAction string
		wantEntity string
	}{
		{"auth.session.created", "login", "session"},
		{"auth.session.revoked", "logout", "session"},
		{"tracker.proj-id.issue.created", "create", "issue"},
		{"tracker.proj-id.issue.updated", "update", "issue"},
		{"tracker.proj-id.issue.deleted", "delete", "issue"},
		{"project.created", "create", "project"},
		{"project.updated", "update", "project"},
		{"tms.proj-id.testcase.created", "create", "testcase"},
		{"messenger.channel.ch-id.message.sent", "create", "message"},
		{"calls.channel.ch-id.started", "create", "call"},
		{"calls.channel.ch-id.ended", "update", "call"},
		{"unknown.subject", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			action, entity := audit.ExportConsumerMapSubjectToAction(consumer, tt.subject)
			if action != tt.wantAction {
				t.Errorf("action = %q, want %q", action, tt.wantAction)
			}
			if entity != tt.wantEntity {
				t.Errorf("entity = %q, want %q", entity, tt.wantEntity)
			}
		})
	}
}

// TestMapCRUDSuffix проверяет конвертацию суффикса события в действие аудита.
func TestMapCRUDSuffix(t *testing.T) {
	tests := []struct {
		suffix string
		want   string
	}{
		{"created", "create"},
		{"sent", "create"},
		{"updated", "update"},
		{"edited", "update"},
		{"deleted", "delete"},
		{"removed", "delete"},
		{"started", "create"},
		{"ended", "update"},
		{"finished", "update"},
		{"custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.suffix, func(t *testing.T) {
			got := audit.ExportMapCRUDSuffix(tt.suffix)
			if got != tt.want {
				t.Errorf("mapCRUDSuffix(%q) = %q, want %q", tt.suffix, got, tt.want)
			}
		})
	}
}

// TestExtractStringField проверяет извлечение строкового поля из payload.
func TestExtractStringField(t *testing.T) {
	payload := map[string]interface{}{
		"id":      "abc-123",
		"user_id": "user-456",
		"empty":   "",
	}

	// Должен вернуть первый непустой ключ
	got := audit.ExportExtractStringField(payload, "missing", "id")
	if got != "abc-123" {
		t.Errorf("audit.ExportExtractStringField = %q, want %q", got, "abc-123")
	}

	// Пустой ключ — пропускается
	got = audit.ExportExtractStringField(payload, "empty", "user_id")
	if got != "user-456" {
		t.Errorf("audit.ExportExtractStringField = %q, want %q", got, "user-456")
	}

	// Нет подходящих — пустая строка
	got = audit.ExportExtractStringField(payload, "nonexistent")
	if got != "" {
		t.Errorf("audit.ExportExtractStringField = %q, want empty", got)
	}
}

// TestExtractUUIDField проверяет извлечение UUID из payload.
func TestExtractUUIDField(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	payload := map[string]interface{}{
		"user_id":    validUUID,
		"project_id": "not-a-uuid",
	}

	// Валидный UUID
	got := audit.ExportExtractUUIDField(payload, "user_id")
	if got == nil {
		t.Fatal("extractUUIDField returned nil for valid UUID")
	}
	if got.String() != validUUID {
		t.Errorf("extractUUIDField = %q, want %q", got.String(), validUUID)
	}

	// Невалидный UUID — пропускается
	got = audit.ExportExtractUUIDField(payload, "project_id")
	if got != nil {
		t.Errorf("extractUUIDField should return nil for invalid UUID, got %v", got)
	}

	// Несуществующий ключ
	got = audit.ExportExtractUUIDField(payload, "missing")
	if got != nil {
		t.Errorf("extractUUIDField should return nil for missing key, got %v", got)
	}
}

// TestAuditLogFiltersDefaults проверяет значения по умолчанию для фильтров.
func TestAuditLogFiltersDefaults(t *testing.T) {
	filters := audit.AuditLogFilters{}

	if filters.Page != 0 {
		t.Errorf("default page = %d, want 0", filters.Page)
	}
	if filters.PerPage != 0 {
		t.Errorf("default per_page = %d, want 0", filters.PerPage)
	}
	if filters.Action != "" {
		t.Errorf("default action = %q, want empty", filters.Action)
	}
}

// TestTrashItemResponseDaysRemaining проверяет корректность вычисления оставшихся дней.
func TestTrashItemResponseDaysRemaining(t *testing.T) {
	item := audit.TrashItemResponse{
		ID:            "test-id",
		Name:          "Test Item",
		EntityType:    "issue",
		DeletedAt:     time.Now().Add(-15 * 24 * time.Hour),
		DaysRemaining: audit.ExportDaysRemaining(time.Now().Add(-15 * 24 * time.Hour)),
	}

	if item.DaysRemaining < 14 || item.DaysRemaining > 15 {
		t.Errorf("DaysRemaining = %d, want between 14 and 15", item.DaysRemaining)
	}
}
