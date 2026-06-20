package auth

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/systemsetting"
)

// RegistrationModeService отвечает за системную настройку режима регистрации
// (open / invite_only) и связанные проверки (SC-14-1).
//
// SEC-06: первая регистрация (count == 0 → SystemRoleAdmin) сериализуется
// process-level мьютексом `firstUserMu`. Без этого две параллельные
// регистрации в "свежей" системе могут обе увидеть count == 0 и обе стать
// админами. Мьютекс закрывает race для single-instance deployment.
// Для multi-instance Postgres-кластера дополнительно нужен
// `pg_advisory_xact_lock(<const>)` — тот же ключ удерживается до commit'а
// транзакции и сериализует регистрации между процессами. SQLite (enttest)
// в advisory lock не нуждается: транзакции уже сериализуемы по умолчанию.
type RegistrationModeService struct {
	db *ent.Client
	// sqlDB enables pg_advisory_xact_lock for multi-instance first-admin (SEC-06).
	sqlDB *sql.DB

	// firstUserMu сериализует CreateFirstOrMemberUser: count → Create
	// должна выполняться атомарно, иначе два конкурентных регистратора
	// могут оба получить SystemRoleAdmin.
	firstUserMu sync.Mutex
}

// NewRegistrationModeService создаёт RegistrationModeService.
// sqlDB may be nil in unit tests (sqlite); production passes shared *sql.DB from main.
func NewRegistrationModeService(db *ent.Client, sqlDB *sql.DB) *RegistrationModeService {
	return &RegistrationModeService{db: db, sqlDB: sqlDB}
}

// GetRegistrationMode читает системную настройку "registration_mode". По умолчанию "open".
func (s *RegistrationModeService) GetRegistrationMode(ctx context.Context) (string, error) {
	setting, err := s.db.SystemSetting.Query().
		Where(systemsetting.KeyEQ("registration_mode")).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "open", nil
		}
		return "", fmt.Errorf("get registration mode: %w", err)
	}
	return setting.Value, nil
}

// SetRegistrationMode обновляет режим регистрации (только для администратора).
//
// ME-04: Ent в этой версии репозитория не экспонирует OnConflict/Upsert
// (см. ent/systemsetting_create.go — нет методов OnConflict*). Поэтому
// делаем "Update → Create on miss" с однократным retry на уникальное
// нарушение: если два администратора параллельно ставят режим, второй
// после AlreadyExists делает повторный Update и попадает в существующую
// строку.
func (s *RegistrationModeService) SetRegistrationMode(ctx context.Context, mode string) error {
	return s.upsertSystemSetting(ctx, "registration_mode", mode,
		"Controls whether new users can register: open or invite_only")
}

// CheckRegistrationAllowed проверяет, разрешена ли регистрация в текущем режиме.
// PUBLIC: вызывается LoginService при создании нового пользователя.
func (s *RegistrationModeService) CheckRegistrationAllowed(ctx context.Context) error {
	mode, _ := s.GetRegistrationMode(ctx)
	if mode == "invite_only" {
		return ErrRegistrationInviteOnly
	}
	return nil
}

// IsFirstUser проверяет, является ли текущая регистрация первой (для назначения admin).
// PUBLIC: вызывается LoginService при создании нового пользователя.
//
// DEPRECATED for first-admin path: см. SEC-06. Используйте CreateFirstOrMemberUser,
// который атомарно (под mu) определяет роль и создаёт пользователя. Метод
// оставлен для совместимости (вспомогательные проверки, не для назначения роли).
func (s *RegistrationModeService) IsFirstUser(ctx context.Context) bool {
	count, err := s.db.User.Query().Count(ctx)
	if err != nil {
		return false
	}
	return count == 0
}

// upsertSystemSetting реализует идемпотентный upsert для SystemSetting,
// устойчивый к параллельным вызовам. Алгоритм:
//
//  1. Попытаться Update — если запись существует, обновим её value.
//  2. Если n == 0 (нет такой записи) — Create.
//  3. Если Create вернул IsConstraintError (= unique violation = другой
//     процесс успел создать запись между шагами 1 и 2) — повторить
//     Update; на этот раз он гарантированно попадёт в созданную чужим
//     процессом строку.
func (s *RegistrationModeService) upsertSystemSetting(ctx context.Context, key, value, description string) error {
	n, err := s.db.SystemSetting.Update().
		Where(systemsetting.KeyEQ(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("upsert system setting %q (update): %w", key, err)
	}
	if n > 0 {
		return nil
	}

	_, err = s.db.SystemSetting.Create().
		SetKey(key).
		SetValue(value).
		SetDescription(description).
		Save(ctx)
	if err == nil {
		return nil
	}

	// Possible race: another caller created the row between our Update
	// and our Create. Retry the Update once — at this point the row
	// definitely exists.
	if ent.IsConstraintError(err) {
		if _, retryErr := s.db.SystemSetting.Update().
			Where(systemsetting.KeyEQ(key)).
			SetValue(value).
			Save(ctx); retryErr != nil {
			return fmt.Errorf("upsert system setting %q (retry update): %w", key, retryErr)
		}
		return nil
	}
	return fmt.Errorf("upsert system setting %q (create): %w", key, err)
}

// sentinel errors
var (
	ErrRegistrationInviteOnly = fmt.Errorf("registration is by invite only")
)
