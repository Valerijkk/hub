package mail

import (
	"context"
	stdsql "database/sql"
	"errors"
	"reflect"
	"time"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/valerii/hub/backend/ent"
)

func (repository *MailboxEntRepository) withMailboxEntReadSnapshot(ctx context.Context, callback func(*ent.Tx) error) (err error) {
	if ctx == nil || ctx.Err() != nil || repository == nil || repository.client == nil || callback == nil {
		return mailboxEntSafeError(errors.New("mailbox ent read snapshot is unavailable"))
	}

	transaction, err := repository.client.BeginTx(ctx, &stdsql.TxOptions{
		Isolation: stdsql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return mailboxEntSafeError(err)
	}

	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback()
			panic(recovered)
		}
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := callback(transaction); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return mailboxEntSafeError(err)
	}
	committed = true
	return nil
}

// MailboxRepositoryClock is the sole time authority for durable mailbox work.
type MailboxRepositoryClock interface {
	Now() time.Time
}

// MailboxEntRepository owns private Ent state shared by Mail repositories.
type MailboxEntRepository struct {
	client *ent.Client
	clock  MailboxRepositoryClock
}

var _ MailboxOwnerRepository = (*MailboxEntRepository)(nil)

// NewMailboxEntRepository creates the production Mail Ent repository.
func NewMailboxEntRepository(client *ent.Client, clock MailboxRepositoryClock) (*MailboxEntRepository, error) {
	if client == nil || mailboxEntNil(clock) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	return &MailboxEntRepository{client: client, clock: clock}, nil
}

func (repository *MailboxEntRepository) withMailboxEntTransaction(ctx context.Context, callback func(*ent.Tx) error) (err error) {
	if repository == nil || repository.client == nil || ctx == nil || callback == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}

	transaction, err := repository.client.Tx(ctx)
	if err != nil {
		return mailboxEntSafeError(err)
	}
	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback()
			panic(recovered)
		}
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	if err := callback(transaction); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return mailboxEntSafeError(err)
	}
	committed = true
	return nil
}

func (repository *MailboxEntRepository) mailboxEntNow() (time.Time, error) {
	if repository == nil || mailboxEntNil(repository.clock) {
		return time.Time{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	now := repository.clock.Now()
	if !mailboxEntValidTime(now) {
		return time.Time{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return now, nil
}

func mailboxEntDatabaseWallClock(ctx context.Context, transaction *ent.Tx) (time.Time, error) {
	if ctx == nil || transaction == nil {
		return time.Time{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	var rows entsql.Rows
	if err := transaction.Driver().Query(ctx, "SELECT date_trunc('second', clock_timestamp())", []any{}, &rows); err != nil {
		return time.Time{}, mailboxEntSafeError(err)
	}

	var databaseWallClock time.Time
	if err := entsql.ScanOne(&rows, &databaseWallClock); err != nil {
		_ = rows.Close()
		return time.Time{}, mailboxEntSafeError(err)
	}
	if err := rows.Close(); err != nil {
		return time.Time{}, mailboxEntSafeError(err)
	}

	databaseWallClock = databaseWallClock.UTC()
	if !mailboxEntValidTime(databaseWallClock) {
		return time.Time{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return databaseWallClock, nil
}

func mailboxEntCheckedTimeAdd(now time.Time, duration time.Duration) (time.Time, bool) {
	if !mailboxEntValidTime(now) || duration <= 0 || duration%time.Second != 0 {
		return time.Time{}, false
	}

	next := now.Add(duration)
	return next, mailboxEntValidTime(next) && next.After(now)
}

func mailboxEntValidTime(value time.Time) bool {
	if value.IsZero() || value.Location() != time.UTC || value.Nanosecond() != 0 {
		return false
	}
	return true
}

func mailboxEntSafeError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrMailboxServiceInvalid):
		return mailboxServiceError(MailboxServiceErrorInvalid)
	case errors.Is(err, ErrMailboxServiceNotFound):
		return mailboxServiceError(MailboxServiceErrorNotFound)
	case errors.Is(err, ErrMailboxServiceDeleting):
		return mailboxServiceError(MailboxServiceErrorDeleting)
	case errors.Is(err, ErrMailboxServiceUnavailable):
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	case ent.IsNotFound(err):
		return mailboxServiceError(MailboxServiceErrorNotFound)
	case ent.IsConstraintError(err):
		return mailboxServiceError(MailboxServiceErrorInvalid)
	default:
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
}

func mailboxEntNil(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func mailboxEntStringPointer(value *string) *string {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}

func mailboxEntRecoveryCodePointer(value *MailboxOwnerRecoveryCode) *MailboxOwnerRecoveryCode {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}

func mailboxEntOwnerRecordCopy(record MailboxOwnerRecord) MailboxOwnerRecord {
	return MailboxOwnerRecord{
		ActorID:         record.ActorID,
		OrganizationID:  record.OrganizationID,
		AccountID:       record.AccountID,
		EngineAccountID: mailboxEntStringPointer(record.EngineAccountID),
		State:           record.State,
		ProjectionState: record.ProjectionState,
		PrimaryAddress:  mailboxEntStringPointer(record.PrimaryAddress),
		RecoveryCode:    mailboxEntRecoveryCodePointer(record.RecoveryCode),
	}
}
