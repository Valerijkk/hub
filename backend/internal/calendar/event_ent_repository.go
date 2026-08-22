package calendar

import (
	"context"
	"time"

	"github.com/valerii/hub/backend/ent"
)

const calendarMutationReceiptTTL = 24 * time.Hour

// EventEntRepository owns Calendar transactions backed by Ent.
type EventEntRepository struct {
	client *ent.Client
	now    func() time.Time
}

var _ EventRepository = (*EventEntRepository)(nil)

// NewEventEntRepository creates the production Calendar Ent repository.
func NewEventEntRepository(client *ent.Client) *EventEntRepository {
	return newEventEntRepository(client, func() time.Time {
		return time.Now().UTC()
	})
}

func newEventEntRepository(client *ent.Client, now func() time.Time) *EventEntRepository {
	if now == nil {
		now = func() time.Time {
			return time.Now().UTC()
		}
	}
	return &EventEntRepository{client: client, now: now}
}

// WithEventTransaction owns exactly one Ent transaction for one callback.
func (repository *EventEntRepository) WithEventTransaction(ctx context.Context, callback func(*EventTransaction) error) error {
	if repository == nil || repository.client == nil || callback == nil {
		return ErrEventInvalid
	}

	transaction, err := repository.client.Tx(ctx)
	if err != nil {
		return calendarEntSafeError(err)
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

	backend := newEventEntTransactionBackend(transaction, repository.now)
	if err := RunEventTransaction(backend, callback); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return calendarEntSafeError(err)
	}
	committed = true
	return nil
}
