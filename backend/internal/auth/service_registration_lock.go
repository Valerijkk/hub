package auth

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

// firstUserAdvisoryKey — ключ pg_advisory_xact_lock для атомарной первой регистрации (SEC-06).
const firstUserAdvisoryKey int64 = 0x48554221

// CreateFirstOrMemberUser атомарно выбирает SystemRole и создаёт пользователя.
// PostgreSQL: advisory xact lock + ent tx. SQLite/tests: mutex + ent tx без advisory.
func (s *RegistrationModeService) CreateFirstOrMemberUser(
	ctx context.Context,
	apply func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate,
) (*ent.User, error) {
	s.firstUserMu.Lock()
	defer s.firstUserMu.Unlock()

	if s.sqlDB != nil {
		return s.createFirstOrMemberUserPostgres(ctx, apply)
	}
	return s.createFirstOrMemberUserEntTx(ctx, apply)
}

func (s *RegistrationModeService) createFirstOrMemberUserPostgres(
	ctx context.Context,
	apply func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate,
) (*ent.User, error) {
	sqlTx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin sql tx: %w", err)
	}
	rolledBack := true
	defer func() {
		if rolledBack {
			_ = sqlTx.Rollback()
		}
	}()

	if _, err := sqlTx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, firstUserAdvisoryKey); err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	drv := entsql.NewDriver(dialect.Postgres, entsql.Conn{ExecQuerier: sqlTx})
	txClient := ent.NewClient(ent.Driver(drv))

	u, err := s.createUserInTx(ctx, txClient, apply)
	if err != nil {
		return nil, err
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	rolledBack = false
	return u, nil
}

func (s *RegistrationModeService) createFirstOrMemberUserEntTx(
	ctx context.Context,
	apply func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate,
) (*ent.User, error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ent tx: %w", err)
	}
	rolledBack := true
	defer func() {
		if rolledBack {
			_ = tx.Rollback()
		}
	}()

	u, err := s.createUserInTx(ctx, tx.Client(), apply)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	rolledBack = false
	return u, nil
}

func (s *RegistrationModeService) createUserInTx(
	ctx context.Context,
	client *ent.Client,
	apply func(b *ent.UserCreate, role user.SystemRole) *ent.UserCreate,
) (*ent.User, error) {
	count, err := client.User.Query().Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	role := user.SystemRoleMember
	if count == 0 {
		role = user.SystemRoleAdmin
	}
	u, err := apply(client.User.Create(), role).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}
