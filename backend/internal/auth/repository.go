package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/session"
	"github.com/valerii/hub/backend/ent/systemsetting"
	"github.com/valerii/hub/backend/ent/user"
)

// firstUserAdvisoryKey — pg_advisory_xact_lock для атомарной первой регистрации (SEC-06).
const firstUserAdvisoryKey int64 = 0x48554221

// SessionRecordInput — данные строки refresh-сессии, сохраняемой в БД.
type SessionRecordInput struct {
	ID               uuid.UUID
	RefreshTokenHash string
	Device           string
	IPAddress        string
	ExpiresAt        time.Time
	UserID           uuid.UUID
}

// Repository — контракт persistence модуля auth (ARCH-01).
type Repository interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*ent.User, error)
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*ent.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (*ent.User, error)
	CountUsers(ctx context.Context) (int, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*ent.User, error)
	MarkOnboardingComplete(ctx context.Context, userID uuid.UUID) (*ent.User, error)
	ListPendingUsers(ctx context.Context) ([]*ent.User, error)
	ApprovePendingUser(ctx context.Context, userID uuid.UUID) error
	RejectPendingUser(ctx context.Context, userID uuid.UUID) error
	PurgeExpiredPending(ctx context.Context, cutoff time.Time) (int, error)

	ListUserSessionRefreshTokens(ctx context.Context, userID uuid.UUID) ([]string, error)
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
	CreateSessionRecord(ctx context.Context, input SessionRecordInput) (*ent.Session, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*ent.Session, error)
	UpdateSessionRefreshHash(ctx context.Context, sessionID uuid.UUID, hash string, lastActive time.Time) error
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*ent.Session, error)
	GetUserSession(ctx context.Context, userID, sessionID uuid.UUID) (*ent.Session, error)

	CountOrganizations(ctx context.Context) (int, error)
	GetOrganization(ctx context.Context) (*ent.Organization, error)
	OrganizationExistsByInviteToken(ctx context.Context, token string) (bool, error)
	CreateOrganizationWithFounder(ctx context.Context, email string) (*ent.User, error)
	CreatePendingMember(ctx context.Context, email string, orgID uuid.UUID) (*ent.User, error)
	CreateActiveMember(ctx context.Context, email string, orgID uuid.UUID) (*ent.User, error)

	GetSystemSettingValue(ctx context.Context, key string) (string, error)
	UpsertSystemSetting(ctx context.Context, key, value, description string) error

	CreateFirstOrMemberUser(ctx context.Context, apply func(*ent.UserCreate, user.SystemRole) *ent.UserCreate) (*ent.User, error)
	RegisterNewUserByEmail(ctx context.Context, email, inviteToken string) (*ent.User, error)

	WithEntTx(ctx context.Context, fn func(ctx context.Context, txRepo Repository) (*ent.User, error)) (*ent.User, error)
	WithPostgresAdvisoryTx(ctx context.Context, sqlDB *sql.DB, fn func(ctx context.Context, txRepo Repository) (*ent.User, error)) (*ent.User, error)
}

// EntRepository — реализация Repository через Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий auth.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) GetUser(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return r.client.User.Get(ctx, userID)
}

func (r *EntRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.client.User.Query().Where(user.EmailEQ(email)).Exist(ctx)
}

func (r *EntRepository) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
}

func (r *EntRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (*ent.User, error) {
	return r.client.User.Query().Where(user.TelegramIDEQ(telegramID)).Only(ctx)
}

func (r *EntRepository) CountUsers(ctx context.Context) (int, error) {
	return r.client.User.Query().Count(ctx)
}

func (r *EntRepository) UpdateUserProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*ent.User, error) {
	update := r.client.User.UpdateOneID(userID)
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.DisplayName != nil {
		update.SetDisplayName(*req.DisplayName)
	}
	if req.AvatarURL != nil {
		update.SetAvatarURL(*req.AvatarURL)
	}
	if req.Phone != nil {
		update.SetPhone(*req.Phone)
	}
	if req.Position != nil {
		update.SetPosition(*req.Position)
	}
	if req.Department != nil {
		update.SetDepartment(*req.Department)
	}
	if req.Timezone != nil {
		update.SetTimezone(*req.Timezone)
	}
	if req.Language != nil {
		update.SetLanguage(*req.Language)
	}
	if req.Birthday != nil {
		update.SetBirthday(*req.Birthday)
	}
	if req.QuietHoursStart != nil {
		update.SetQuietHoursStart(*req.QuietHoursStart)
	}
	if req.QuietHoursEnd != nil {
		update.SetQuietHoursEnd(*req.QuietHoursEnd)
	}
	if req.CustomFields != nil {
		update.SetCustomFields(req.CustomFields)
	}
	return update.Save(ctx)
}

func (r *EntRepository) MarkOnboardingComplete(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return r.client.User.UpdateOneID(userID).
		SetOnboardingCompleted(true).
		Save(ctx)
}

func (r *EntRepository) ListPendingUsers(ctx context.Context) ([]*ent.User, error) {
	return r.client.User.Query().
		Where(user.StatusEQ(user.StatusPending)).
		Order(ent.Desc(user.FieldPendingSince)).
		All(ctx)
}

func (r *EntRepository) ApprovePendingUser(ctx context.Context, userID uuid.UUID) error {
	n, err := r.client.User.Update().
		Where(user.IDEQ(userID), user.StatusEQ(user.StatusPending)).
		SetStatus(user.StatusActive).
		ClearPendingSince().
		Save(ctx)
	if err != nil {
		return fmt.Errorf("approve pending user: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *EntRepository) RejectPendingUser(ctx context.Context, userID uuid.UUID) error {
	n, err := r.client.User.Delete().
		Where(user.IDEQ(userID), user.StatusEQ(user.StatusPending)).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("reject pending user: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *EntRepository) PurgeExpiredPending(ctx context.Context, cutoff time.Time) (int, error) {
	n, err := r.client.User.Delete().
		Where(user.StatusEQ(user.StatusPending), user.PendingSinceLT(cutoff)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("purge expired pending: %w", err)
	}
	return n, nil
}

func (r *EntRepository) ListUserSessionRefreshTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return r.client.Session.Query().
		Where(session.HasUserWith(user.ID(userID))).
		Select(session.FieldRefreshToken).
		Strings(ctx)
}

func (r *EntRepository) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := r.client.Session.Delete().
		Where(session.HasUserWith(user.ID(userID))).
		Exec(ctx)
	return err
}

func (r *EntRepository) CreateSessionRecord(ctx context.Context, input SessionRecordInput) (*ent.Session, error) {
	return r.client.Session.Create().
		SetID(input.ID).
		SetRefreshToken(input.RefreshTokenHash).
		SetDevice(input.Device).
		SetIPAddress(input.IPAddress).
		SetExpiresAt(input.ExpiresAt).
		SetUserID(input.UserID).
		Save(ctx)
}

func (r *EntRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*ent.Session, error) {
	return r.client.Session.Get(ctx, sessionID)
}

func (r *EntRepository) UpdateSessionRefreshHash(ctx context.Context, sessionID uuid.UUID, hash string, lastActive time.Time) error {
	_, err := r.client.Session.UpdateOneID(sessionID).
		SetRefreshToken(hash).
		SetLastActiveAt(lastActive).
		Save(ctx)
	return err
}

func (r *EntRepository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	return r.client.Session.DeleteOneID(sessionID).Exec(ctx)
}

func (r *EntRepository) ListUserSessions(ctx context.Context, userID uuid.UUID) ([]*ent.Session, error) {
	return r.client.Session.Query().
		Where(session.HasUserWith(user.ID(userID))).
		Order(ent.Desc(session.FieldCreatedAt)).
		All(ctx)
}

func (r *EntRepository) GetUserSession(ctx context.Context, userID, sessionID uuid.UUID) (*ent.Session, error) {
	return r.client.Session.Query().
		Where(session.ID(sessionID), session.HasUserWith(user.ID(userID))).
		Only(ctx)
}

func (r *EntRepository) CountOrganizations(ctx context.Context) (int, error) {
	return r.client.Organization.Query().Count(ctx)
}

func (r *EntRepository) GetOrganization(ctx context.Context) (*ent.Organization, error) {
	return r.client.Organization.Query().Only(ctx)
}

func (r *EntRepository) OrganizationExistsByInviteToken(ctx context.Context, token string) (bool, error) {
	return r.client.Organization.Query().
		Where(organization.InviteTokenEQ(token)).
		Exist(ctx)
}

func (r *EntRepository) CreateOrganizationWithFounder(ctx context.Context, email string) (*ent.User, error) {
	// W6-A1-009 / P3-BE-07: inviteToken-параметр убран — org всегда получает generateInviteToken().
	token, err := generateInviteToken()
	if err != nil {
		return nil, err
	}
	org, err := r.client.Organization.Create().
		SetName("Hub").
		SetSlug("hub").
		SetInviteToken(token).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}
	return r.client.User.Create().
		SetEmail(email).
		SetSystemRole(user.SystemRoleFounder).
		SetStatus(user.StatusActive).
		SetOrganization(org).
		Save(ctx)
}

func (r *EntRepository) CreatePendingMember(ctx context.Context, email string, orgID uuid.UUID) (*ent.User, error) {
	now := time.Now().UTC()
	return r.client.User.Create().
		SetEmail(email).
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusPending).
		SetPendingSince(now).
		SetOrganizationID(orgID).
		Save(ctx)
}

func (r *EntRepository) CreateActiveMember(ctx context.Context, email string, orgID uuid.UUID) (*ent.User, error) {
	return r.client.User.Create().
		SetEmail(email).
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetOrganizationID(orgID).
		Save(ctx)
}

func (r *EntRepository) GetSystemSettingValue(ctx context.Context, key string) (string, error) {
	setting, err := r.client.SystemSetting.Query().
		Where(systemsetting.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return setting.Value, nil
}

func (r *EntRepository) UpsertSystemSetting(ctx context.Context, key, value, description string) error {
	n, err := r.client.SystemSetting.Update().
		Where(systemsetting.KeyEQ(key)).
		SetValue(value).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("upsert system setting %q (update): %w", key, err)
	}
	if n > 0 {
		return nil
	}
	_, err = r.client.SystemSetting.Create().
		SetKey(key).
		SetValue(value).
		SetDescription(description).
		Save(ctx)
	if err == nil {
		return nil
	}
	if ent.IsConstraintError(err) {
		if _, retryErr := r.client.SystemSetting.Update().
			Where(systemsetting.KeyEQ(key)).
			SetValue(value).
			Save(ctx); retryErr != nil {
			return fmt.Errorf("upsert system setting %q (retry update): %w", key, retryErr)
		}
		return nil
	}
	return fmt.Errorf("upsert system setting %q (create): %w", key, err)
}

func (r *EntRepository) CreateFirstOrMemberUser(ctx context.Context, apply func(*ent.UserCreate, user.SystemRole) *ent.UserCreate) (*ent.User, error) {
	count, err := r.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	role := user.SystemRoleMember
	if count == 0 {
		role = user.SystemRoleFounder
	}
	u, err := apply(r.client.User.Create(), role).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *EntRepository) RegisterNewUserByEmail(ctx context.Context, email, inviteToken string) (*ent.User, error) {
	count, err := r.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	if count == 0 {
		return r.CreateOrganizationWithFounder(ctx, email)
	}

	org, err := r.GetOrganization(ctx)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}

	mode, err := r.GetSystemSettingValue(ctx, "registration_mode")
	if err != nil {
		return nil, err
	}
	if mode == "" {
		mode = "invite_only"
	}

	if mode == "invite_only" {
		if inviteToken == "" {
			return nil, ErrInvalidInviteToken
		}
		ok, err := r.OrganizationExistsByInviteToken(ctx, inviteToken)
		if err != nil {
			return nil, fmt.Errorf("validate invite token: %w", err)
		}
		if !ok {
			return nil, ErrInvalidInviteToken
		}
		return r.CreatePendingMember(ctx, email, org.ID)
	}

	return r.CreateActiveMember(ctx, email, org.ID)
}

// WithEntTx выполняет fn в Ent-транзакции с rollback при ошибке.
func (r *EntRepository) WithEntTx(ctx context.Context, fn func(ctx context.Context, txRepo Repository) (*ent.User, error)) (*ent.User, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin ent tx: %w", err)
	}
	rolledBack := true
	defer func() {
		if rolledBack {
			_ = tx.Rollback()
		}
	}()

	u, err := fn(ctx, NewEntRepository(tx.Client()))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	rolledBack = false
	return u, nil
}

// WithPostgresAdvisoryTx выполняет fn в SQL-транзакции с pg_advisory_xact_lock —
// защита от гонки двух одновременных «первых пользователей» (SEC-06).
func (r *EntRepository) WithPostgresAdvisoryTx(ctx context.Context, sqlDB *sql.DB, fn func(ctx context.Context, txRepo Repository) (*ent.User, error)) (*ent.User, error) {
	sqlTx, err := sqlDB.BeginTx(ctx, nil)
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

	u, err := fn(ctx, NewEntRepository(txClient))
	if err != nil {
		return nil, err
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	rolledBack = false
	return u, nil
}

// IsNotFound сообщает об ошибках ent not-found.
func IsNotFound(err error) bool {
	return ent.IsNotFound(err)
}
