package mail

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/valerii/hub/backend/ent"
)

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
