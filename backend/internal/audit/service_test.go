package audit

import (
	"testing"
	"time"
)

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
			got := daysRemaining(tt.deletedAt)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("daysRemaining() = %d, want between %d and %d", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestMapSubjectToAction проверяет маршрутизацию NATS subject в пару (действие, тип сущности).
func TestMapSubjectToAction(t *testing.T) {
	consumer := &AuditConsumer{}

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
			action, entity := consumer.mapSubjectToAction(tt.subject)
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
			got := mapCRUDSuffix(tt.suffix)
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
	got := extractStringField(payload, "missing", "id")
	if got != "abc-123" {
		t.Errorf("extractStringField = %q, want %q", got, "abc-123")
	}

	// Пустой ключ — пропускается
	got = extractStringField(payload, "empty", "user_id")
	if got != "user-456" {
		t.Errorf("extractStringField = %q, want %q", got, "user-456")
	}

	// Нет подходящих — пустая строка
	got = extractStringField(payload, "nonexistent")
	if got != "" {
		t.Errorf("extractStringField = %q, want empty", got)
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
	got := extractUUIDField(payload, "user_id")
	if got == nil {
		t.Fatal("extractUUIDField returned nil for valid UUID")
	}
	if got.String() != validUUID {
		t.Errorf("extractUUIDField = %q, want %q", got.String(), validUUID)
	}

	// Невалидный UUID — пропускается
	got = extractUUIDField(payload, "project_id")
	if got != nil {
		t.Errorf("extractUUIDField should return nil for invalid UUID, got %v", got)
	}

	// Несуществующий ключ
	got = extractUUIDField(payload, "missing")
	if got != nil {
		t.Errorf("extractUUIDField should return nil for missing key, got %v", got)
	}
}

// TestAuditLogFiltersDefaults проверяет значения по умолчанию для фильтров.
func TestAuditLogFiltersDefaults(t *testing.T) {
	filters := AuditLogFilters{}

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
	item := TrashItemResponse{
		ID:            "test-id",
		Name:          "Test Item",
		EntityType:    "issue",
		DeletedAt:     time.Now().Add(-15 * 24 * time.Hour),
		DaysRemaining: daysRemaining(time.Now().Add(-15 * 24 * time.Hour)),
	}

	if item.DaysRemaining < 14 || item.DaysRemaining > 15 {
		t.Errorf("DaysRemaining = %d, want between 14 and 15", item.DaysRemaining)
	}
}
