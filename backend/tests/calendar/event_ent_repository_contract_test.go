package calendar_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/calendar"
)

var _ dialect.Driver = (*eventEntRepositoryDriver)(nil)
var _ dialect.Tx = (*eventEntRepositoryTx)(nil)

func TestEventEntRepositoryWithEventTransactionRejectsNilDependencies(t *testing.T) {
	callbackCalls := 0
	err := calendar.NewEventEntRepository(nil).WithEventTransaction(context.Background(), func(*calendar.EventTransaction) error {
		callbackCalls++
		return nil
	})
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("nil client error = %v, want invalid", err)
	}
	if callbackCalls != 0 {
		t.Fatalf("nil client callback calls = %d, want 0", callbackCalls)
	}

	driver := &eventEntRepositoryDriver{}
	repository := eventEntRepository(driver)
	err = repository.WithEventTransaction(context.Background(), nil)
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("nil callback error = %v, want invalid", err)
	}
	if driver.beginCalls != 0 {
		t.Fatalf("nil callback begin calls = %d, want 0", driver.beginCalls)
	}
}

func TestEventEntRepositoryWithEventTransactionOwnsOneClosedTransactionPerCallback(t *testing.T) {
	driver := &eventEntRepositoryDriver{}
	repository := eventEntRepository(driver)
	var retained *calendar.EventTransaction
	transactions := make([]*calendar.EventTransaction, 0, 2)

	for call := 0; call < 2; call++ {
		err := repository.WithEventTransaction(context.Background(), func(transaction *calendar.EventTransaction) error {
			transactions = append(transactions, transaction)
			if retained == nil {
				retained = transaction
			}
			return nil
		})
		if err != nil {
			t.Fatalf("WithEventTransaction() call %d error = %v", call+1, err)
		}
	}

	if driver.beginCalls != 2 || len(driver.transactions) != 2 {
		t.Fatalf("transactions begin:%d stored:%d, want 2/2", driver.beginCalls, len(driver.transactions))
	}
	if len(transactions) != 2 || transactions[0] == transactions[1] {
		t.Fatal("callback transaction was reused across calls")
	}
	for index, transaction := range driver.transactions {
		if transaction.commits != 1 || transaction.rollbacks != 0 {
			t.Fatalf("transaction %d commit:%d rollback:%d, want 1/0", index, transaction.commits, transaction.rollbacks)
		}
	}
	if _, err := retained.ResolveMutationScope(context.Background(), eventEntRepositoryActorID, calendar.MutationReceiptEventCreate); !errors.Is(err, calendar.ErrEventTransactionClosed) {
		t.Fatalf("retained transaction error = %v, want closed", err)
	}
}

func TestEventEntRepositoryWithEventTransactionLocksOrganizationBeforeUser(t *testing.T) {
	actorID := eventEntRepositoryActorID
	attendeeID := eventEntRepositoryAttendeeID
	driver := &eventEntRepositoryDriver{
		organizationID:     eventEntRepositoryOrganizationID,
		queryRows:          true,
		expectedQueryOrder: []string{"organization", "user", "organization", "user", "organization", "user"},
		userIDs:            []uuid.UUID{actorID, actorID, attendeeID},
	}
	repository := eventEntRepository(driver)

	err := repository.WithEventTransaction(context.Background(), func(transaction *calendar.EventTransaction) error {
		scope, err := transaction.ResolveMutationScope(context.Background(), actorID, calendar.MutationReceiptEventCreate)
		if err != nil {
			return err
		}
		if scope.OrganizationID() != eventEntRepositoryOrganizationID {
			t.Fatalf("organization ID = %s, want %s", scope.OrganizationID(), eventEntRepositoryOrganizationID)
		}
		_, err = transaction.ResolveCalendarEventCreateAuthorization(context.Background(), scope, calendar.CalendarEventCreateAuthorizationRequest{
			Organizer: calendar.CalendarEventCreateAttendeeCandidate{UserID: &actorID},
			Attendees: []calendar.CalendarEventCreateAttendeeCandidate{
				{UserID: &attendeeID},
			},
		})
		return err
	})
	if err != nil {
		t.Fatalf("authorization transaction error = %v", err)
	}
	driver.assertQueryOrder(t)
}

func TestEventEntRepositoryWithEventTransactionRollsBackCallbackFailures(t *testing.T) {
	callbackErr := errors.New("callback failure")
	driver := &eventEntRepositoryDriver{}
	repository := eventEntRepository(driver)

	err := repository.WithEventTransaction(context.Background(), func(*calendar.EventTransaction) error {
		return callbackErr
	})
	if err != callbackErr {
		t.Fatalf("callback error = %v, want original %v", err, callbackErr)
	}
	if driver.beginCalls != 1 || len(driver.transactions) != 1 || driver.transactions[0].commits != 0 || driver.transactions[0].rollbacks != 1 {
		t.Fatalf("callback error transaction = begin:%d tx:%d commit:%d rollback:%d, want 1/1/0/1", driver.beginCalls, len(driver.transactions), driver.transactions[0].commits, driver.transactions[0].rollbacks)
	}

	panicValue := errors.New("callback panic")
	driver = &eventEntRepositoryDriver{}
	repository = eventEntRepository(driver)
	func() {
		defer func() {
			if got := recover(); got != panicValue {
				t.Fatalf("callback panic = %#v, want %#v", got, panicValue)
			}
		}()
		_ = repository.WithEventTransaction(context.Background(), func(*calendar.EventTransaction) error {
			panic(panicValue)
		})
	}()
	if driver.beginCalls != 1 || len(driver.transactions) != 1 || driver.transactions[0].commits != 0 || driver.transactions[0].rollbacks != 1 {
		t.Fatalf("callback panic transaction = begin:%d tx:%d commit:%d rollback:%d, want 1/1/0/1", driver.beginCalls, len(driver.transactions), driver.transactions[0].commits, driver.transactions[0].rollbacks)
	}
}

func TestEventEntRepositoryWithEventTransactionMapsBeginAndCommitFailures(t *testing.T) {
	driver := &eventEntRepositoryDriver{beginErr: errors.New("begin failure")}
	repository := eventEntRepository(driver)
	callbackCalls := 0
	err := repository.WithEventTransaction(context.Background(), func(*calendar.EventTransaction) error {
		callbackCalls++
		return nil
	})
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("begin error = %v, want invalid", err)
	}
	if driver.beginCalls != 1 || callbackCalls != 0 || len(driver.transactions) != 0 {
		t.Fatalf("begin failure = begin:%d callbacks:%d transactions:%d, want 1/0/0", driver.beginCalls, callbackCalls, len(driver.transactions))
	}

	driver = &eventEntRepositoryDriver{nextTransaction: func() *eventEntRepositoryTx {
		return &eventEntRepositoryTx{commitErr: errors.New("commit failure")}
	}}
	repository = eventEntRepository(driver)
	err = repository.WithEventTransaction(context.Background(), func(*calendar.EventTransaction) error { return nil })
	if !errors.Is(err, calendar.ErrEventInvalid) {
		t.Fatalf("commit error = %v, want invalid", err)
	}
	if driver.beginCalls != 1 || len(driver.transactions) != 1 || driver.transactions[0].commits != 1 || driver.transactions[0].rollbacks != 1 {
		t.Fatalf("commit failure transaction = begin:%d tx:%d commit:%d rollback:%d, want 1/1/1/1", driver.beginCalls, len(driver.transactions), driver.transactions[0].commits, driver.transactions[0].rollbacks)
	}
}

func eventEntRepository(driver *eventEntRepositoryDriver) *calendar.EventEntRepository {
	return calendar.NewEventEntRepository(ent.NewClient(ent.Driver(driver)))
}

type eventEntRepositoryDriver struct {
	beginCalls         int
	beginErr           error
	transactions       []*eventEntRepositoryTx
	nextTransaction    func() *eventEntRepositoryTx
	organizationID     uuid.UUID
	queryOrder         []string
	expectedQueryOrder []string
	queryRows          bool
	userIDs            []uuid.UUID
	userQueryIndex     int
}

func (driver *eventEntRepositoryDriver) Exec(context.Context, string, any, any) error {
	return errors.New("unexpected query execution")
}

func (driver *eventEntRepositoryDriver) Query(_ context.Context, query string, value any, destination any) error {
	if !driver.queryRows {
		return errors.New("unexpected query execution")
	}
	queryKind := eventEntRepositoryQueryKind(query)
	if queryKind == "" {
		return errors.New("unexpected authorization query")
	}
	driver.queryOrder = append(driver.queryOrder, queryKind)
	queryIndex := len(driver.queryOrder) - 1
	if queryIndex >= len(driver.expectedQueryOrder) || driver.expectedQueryOrder[queryIndex] != queryKind {
		return fmt.Errorf("authorization lock query %d = %s", queryIndex, queryKind)
	}
	rows, ok := destination.(*entsql.Rows)
	if !ok {
		return fmt.Errorf("authorization query destination = %T", destination)
	}
	rowID := driver.organizationID
	if queryKind == "user" {
		if driver.userQueryIndex >= len(driver.userIDs) {
			return errors.New("unexpected user query")
		}
		rowID = driver.userIDs[driver.userQueryIndex]
		driver.userQueryIndex++
	}
	_ = value
	*rows = entsql.Rows{ColumnScanner: &eventEntRepositoryRows{id: rowID}}
	return nil
}

func (driver *eventEntRepositoryDriver) Tx(context.Context) (dialect.Tx, error) {
	driver.beginCalls++
	if driver.beginErr != nil {
		return nil, driver.beginErr
	}
	transaction := &eventEntRepositoryTx{}
	if driver.nextTransaction != nil {
		transaction = driver.nextTransaction()
	}
	transaction.driver = driver
	driver.transactions = append(driver.transactions, transaction)
	return transaction, nil
}

func (driver *eventEntRepositoryDriver) BeginTx(ctx context.Context, _ *sql.TxOptions) (dialect.Tx, error) {
	return driver.Tx(ctx)
}

func (*eventEntRepositoryDriver) Close() error { return nil }

func (*eventEntRepositoryDriver) Dialect() string { return dialect.Postgres }

type eventEntRepositoryTx struct {
	driver    *eventEntRepositoryDriver
	commitErr error
	commits   int
	rollbacks int
}

func (transaction *eventEntRepositoryTx) Exec(context.Context, string, any, any) error {
	return errors.New("unexpected transaction execution")
}

func (transaction *eventEntRepositoryTx) Query(ctx context.Context, query string, value any, destination any) error {
	if transaction.driver == nil {
		return errors.New("unexpected transaction query")
	}
	return transaction.driver.Query(ctx, query, value, destination)
}

func (transaction *eventEntRepositoryTx) Commit() error {
	transaction.commits++
	return transaction.commitErr
}

func (transaction *eventEntRepositoryTx) Rollback() error {
	transaction.rollbacks++
	return nil
}

func (driver *eventEntRepositoryDriver) assertQueryOrder(t *testing.T) {
	t.Helper()
	if len(driver.queryOrder) != len(driver.expectedQueryOrder) {
		t.Fatalf("authorization query count = %d, want %d", len(driver.queryOrder), len(driver.expectedQueryOrder))
	}
	if got, want := strings.Join(driver.queryOrder, ","), strings.Join(driver.expectedQueryOrder, ","); got != want {
		t.Fatalf("authorization lock query order = %s, want %s", got, want)
	}
}

func eventEntRepositoryQueryKind(query string) string {
	switch {
	case strings.Contains(query, "organizations"):
		return "organization"
	case strings.Contains(query, "users"):
		return "user"
	default:
		return ""
	}
}

type eventEntRepositoryRows struct {
	id      uuid.UUID
	visited bool
}

func (*eventEntRepositoryRows) Close() error { return nil }

func (*eventEntRepositoryRows) ColumnTypes() ([]*sql.ColumnType, error) { return nil, nil }

func (*eventEntRepositoryRows) Columns() ([]string, error) { return nil, nil }

func (*eventEntRepositoryRows) Err() error { return nil }

func (rows *eventEntRepositoryRows) Next() bool {
	if rows.visited {
		return false
	}
	rows.visited = true
	return true
}

func (*eventEntRepositoryRows) NextResultSet() bool { return false }

func (rows *eventEntRepositoryRows) Scan(destinations ...any) error {
	for _, destination := range destinations {
		eventEntRepositorySetScanValue(destination, rows.id)
	}
	return nil
}

func eventEntRepositorySetScanValue(destination any, id uuid.UUID) {
	value := reflect.ValueOf(destination)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return
	}
	target := value.Elem()
	switch {
	case target.Type() == reflect.TypeOf(uuid.UUID{}):
		target.Set(reflect.ValueOf(id))
	case target.Type() == reflect.TypeOf(time.Time{}):
		target.Set(reflect.ValueOf(time.Unix(0, 0).UTC()))
	case target.Kind() == reflect.String:
		target.SetString("active")
	case target.Kind() == reflect.Bool:
		target.SetBool(false)
	case target.Kind() >= reflect.Int && target.Kind() <= reflect.Int64:
		target.SetInt(1)
	case target.Kind() >= reflect.Uint && target.Kind() <= reflect.Uint64:
		target.SetUint(1)
	case target.Kind() == reflect.Slice:
		target.Set(reflect.MakeSlice(target.Type(), 0, 0))
	}
}

var eventEntRepositoryActorID = uuid.UUID{1}
var eventEntRepositoryOrganizationID = uuid.UUID{2}
var eventEntRepositoryAttendeeID = uuid.UUID{3}
