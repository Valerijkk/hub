package mail_test

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
	mail "github.com/valerii/hub/backend/internal/mail"
)

var _ dialect.Driver = (*mailboxEntCalendarRSVPTargetDriver)(nil)
var _ dialect.Tx = (*mailboxEntCalendarRSVPTargetTx)(nil)

func TestMailboxEntRepositoryResolveCalendarEventForMessageSuppressesTargetForMissingSeriesAttendee(t *testing.T) {
	fixture := mailboxEntCalendarRSVPTargetFixtureValues()
	request, err := mail.NewMailboxCalendarCorrelationRequest(
		fixture.actorID,
		fixture.organizationID,
		fixture.accountID,
		fixture.messageID,
	)
	if err != nil {
		t.Fatalf("NewMailboxCalendarCorrelationRequest() error = %v", err)
	}

	for _, testCase := range []struct {
		name               string
		attendeeQueryError bool
	}{
		{name: "missing active series attendee"},
		{name: "attendee query error", attendeeQueryError: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			driver := &mailboxEntCalendarRSVPTargetDriver{
				fixture:            fixture,
				attendeeQueryError: testCase.attendeeQueryError,
			}
			repository, err := mail.NewMailboxEntRepository(
				ent.NewClient(ent.Driver(driver)),
				mailboxEntCalendarRSVPTargetClock{now: time.Date(2032, time.February, 4, 4, 5, 6, 0, time.UTC)},
			)
			if err != nil {
				t.Fatalf("NewMailboxEntRepository() error = %v", err)
			}

			result, err := repository.ResolveCalendarEventForMessage(context.Background(), request)
			if testCase.attendeeQueryError {
				requireMailboxCalendarCorrelationErrorKind(t, err, mail.MailboxCalendarCorrelationErrorKindUnavailable)
				driver.assertQueryOrder(t)
				return
			}
			if err != nil {
				t.Fatalf("ResolveCalendarEventForMessage() error = %v", err)
			}
			if !result.Matches(request) {
				t.Fatalf("correlation result does not match request: %#v", result)
			}
			if result.Path() != mail.MailboxCalendarCorrelationPathInboundReceipt {
				t.Fatalf("correlation path = %q, want inbound receipt", result.Path())
			}
			if result.CalendarEventID() != fixture.calendarEventID {
				t.Fatalf("calendar event ID = %s, want %s", result.CalendarEventID(), fixture.calendarEventID)
			}
			if result.CalendarRSVPTarget() != nil {
				t.Fatalf("calendar RSVP target = %#v, want nil", result.CalendarRSVPTarget())
			}
			driver.assertQueryOrder(t)
		})
	}
}

type mailboxEntCalendarRSVPTargetFixture struct {
	actorID         uuid.UUID
	organizationID  uuid.UUID
	accountID       uuid.UUID
	messageID       uuid.UUID
	calendarEventID uuid.UUID
	eventEnvelopeID uuid.UUID
	receiptID       uuid.UUID
	partLocatorHash []byte
}

func mailboxEntCalendarRSVPTargetFixtureValues() mailboxEntCalendarRSVPTargetFixture {
	partLocatorHash := make([]byte, 32)
	for index := range partLocatorHash {
		partLocatorHash[index] = byte(index + 1)
	}
	return mailboxEntCalendarRSVPTargetFixture{
		actorID:         uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		organizationID:  uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		accountID:       uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		messageID:       uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		calendarEventID: uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		eventEnvelopeID: uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		receiptID:       uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		partLocatorHash: partLocatorHash,
	}
}

type mailboxEntCalendarRSVPTargetClock struct {
	now time.Time
}

func (clock mailboxEntCalendarRSVPTargetClock) Now() time.Time {
	return clock.now
}

type mailboxEntCalendarRSVPTargetDriver struct {
	fixture            mailboxEntCalendarRSVPTargetFixture
	attendeeQueryError bool
	queryOrder         []string
}

func (*mailboxEntCalendarRSVPTargetDriver) Exec(context.Context, string, any, any) error {
	return errors.New("unexpected query execution")
}

func (driver *mailboxEntCalendarRSVPTargetDriver) Query(_ context.Context, query string, _ any, destination any) error {
	queryKind := mailboxEntCalendarRSVPTargetQueryKind(query)
	if queryKind == "" {
		return fmt.Errorf("unexpected resolver query %q", query)
	}
	driver.queryOrder = append(driver.queryOrder, queryKind)
	if queryKind == "calendar_attendees" && driver.attendeeQueryError {
		return errors.New("calendar attendee query failure")
	}

	rows, ok := destination.(*entsql.Rows)
	if !ok {
		return fmt.Errorf("resolver query destination = %T", destination)
	}
	*rows = entsql.Rows{ColumnScanner: driver.rows(queryKind)}
	return nil
}

func (driver *mailboxEntCalendarRSVPTargetDriver) Tx(context.Context) (dialect.Tx, error) {
	return &mailboxEntCalendarRSVPTargetTx{driver: driver}, nil
}

func (driver *mailboxEntCalendarRSVPTargetDriver) BeginTx(ctx context.Context, _ *sql.TxOptions) (dialect.Tx, error) {
	return driver.Tx(ctx)
}

func (*mailboxEntCalendarRSVPTargetDriver) Close() error {
	return nil
}

func (*mailboxEntCalendarRSVPTargetDriver) Dialect() string {
	return dialect.Postgres
}

func (driver *mailboxEntCalendarRSVPTargetDriver) rows(queryKind string) *mailboxEntCalendarRSVPTargetRows {
	switch queryKind {
	case "mail_accounts":
		return &mailboxEntCalendarRSVPTargetRows{
			columns: []string{"id", "organization_id", "user_id", "desired_enabled", "status", "projection_state"},
			values:  [][]any{{driver.fixture.accountID, driver.fixture.organizationID, driver.fixture.actorID, true, "active", "ready"}},
		}
	case "mail_messages":
		return &mailboxEntCalendarRSVPTargetRows{
			columns: []string{"id", "mail_account_id", "is_draft", "projection_origin", "calendar_method", "calendar_part_locator_hash", "deleted_at"},
			values:  [][]any{{driver.fixture.messageID, driver.fixture.accountID, false, "inbound_or_unknown", "request", driver.fixture.partLocatorHash, nil}},
		}
	case "calendar_inbound_message_receipts":
		receiptTime := time.Date(2032, time.February, 4, 4, 5, 6, 0, time.UTC)
		return &mailboxEntCalendarRSVPTargetRows{
			columns: []string{"id", "organization_id", "mailbox_id", "mail_message_id", "part_locator_hash", "event_envelope_id", "state", "attempts", "lease_token", "lease_generation", "lease_until", "next_attempt_at", "error_code", "calendar_event_id", "deferred_object_state_id", "deferred_expires_at", "calendar_component_scope", "recurrence_id", "created_at", "updated_at"},
			values:  [][]any{{driver.fixture.receiptID, driver.fixture.organizationID, driver.fixture.accountID, driver.fixture.messageID, driver.fixture.partLocatorHash, driver.fixture.eventEnvelopeID, "applied", 1, nil, int64(1), nil, nil, nil, driver.fixture.calendarEventID, nil, nil, "series", nil, receiptTime, receiptTime}},
		}
	case "calendar_events":
		return &mailboxEntCalendarRSVPTargetRows{
			columns: []string{"id", "organization_id", "calendar_owner_id", "organizer_id", "deleted_at"},
			values:  [][]any{{driver.fixture.calendarEventID, driver.fixture.organizationID, driver.fixture.actorID, driver.fixture.actorID, nil}},
		}
	case "calendar_attendees":
		return &mailboxEntCalendarRSVPTargetRows{columns: []string{"id"}}
	default:
		panic(fmt.Sprintf("unsupported query kind %q", queryKind))
	}
}

func (driver *mailboxEntCalendarRSVPTargetDriver) assertQueryOrder(t *testing.T) {
	t.Helper()
	want := []string{
		"mail_accounts",
		"mail_messages",
		"calendar_inbound_message_receipts",
		"calendar_events",
		"calendar_attendees",
	}
	if got := strings.Join(driver.queryOrder, ","); got != strings.Join(want, ",") {
		t.Fatalf("query order = %s, want %s", got, strings.Join(want, ","))
	}
}

func mailboxEntCalendarRSVPTargetQueryKind(query string) string {
	switch {
	case strings.Contains(query, `FROM "mail_accounts"`):
		return "mail_accounts"
	case strings.Contains(query, `FROM "mail_messages"`):
		return "mail_messages"
	case strings.Contains(query, `FROM "calendar_inbound_message_receipts"`):
		return "calendar_inbound_message_receipts"
	case strings.Contains(query, `FROM "calendar_events"`):
		return "calendar_events"
	case strings.Contains(query, `FROM "calendar_attendees"`):
		return "calendar_attendees"
	default:
		return ""
	}
}

type mailboxEntCalendarRSVPTargetTx struct {
	driver *mailboxEntCalendarRSVPTargetDriver
}

func (*mailboxEntCalendarRSVPTargetTx) Exec(context.Context, string, any, any) error {
	return errors.New("unexpected transaction execution")
}

func (transaction *mailboxEntCalendarRSVPTargetTx) Query(ctx context.Context, query string, value any, destination any) error {
	if transaction.driver == nil {
		return errors.New("missing resolver driver")
	}
	return transaction.driver.Query(ctx, query, value, destination)
}

func (*mailboxEntCalendarRSVPTargetTx) Commit() error {
	return nil
}

func (*mailboxEntCalendarRSVPTargetTx) Rollback() error {
	return nil
}

type mailboxEntCalendarRSVPTargetRows struct {
	columns []string
	values  [][]any
	index   int
}

func (*mailboxEntCalendarRSVPTargetRows) Close() error {
	return nil
}

func (*mailboxEntCalendarRSVPTargetRows) ColumnTypes() ([]*sql.ColumnType, error) {
	return nil, nil
}

func (rows *mailboxEntCalendarRSVPTargetRows) Columns() ([]string, error) {
	return rows.columns, nil
}

func (*mailboxEntCalendarRSVPTargetRows) Err() error {
	return nil
}

func (rows *mailboxEntCalendarRSVPTargetRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++
	return true
}

func (*mailboxEntCalendarRSVPTargetRows) NextResultSet() bool {
	return false
}

func (rows *mailboxEntCalendarRSVPTargetRows) Scan(destinations ...any) error {
	if rows.index == 0 || rows.index > len(rows.values) {
		return errors.New("scan without a row")
	}
	if len(destinations) != len(rows.columns) {
		return fmt.Errorf("scan destinations = %d, want %d", len(destinations), len(rows.columns))
	}
	for index, value := range rows.values[rows.index-1] {
		if err := mailboxEntCalendarRSVPTargetSetScanValue(destinations[index], value); err != nil {
			return err
		}
	}
	return nil
}

func mailboxEntCalendarRSVPTargetSetScanValue(destination, value any) error {
	target := reflect.ValueOf(destination)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return fmt.Errorf("scan destination = %T", destination)
	}
	target = target.Elem()
	if value == nil {
		return nil
	}
	if scanner, ok := destination.(*entsql.NullScanner); ok {
		id, ok := value.(uuid.UUID)
		if !ok {
			return fmt.Errorf("scan UUID value = %T", value)
		}
		output, ok := scanner.S.(*uuid.UUID)
		if !ok || output == nil {
			return errors.New("invalid UUID null scanner")
		}
		*output = id
		scanner.Valid = true
		return nil
	}
	if target.Type() == reflect.TypeOf(sql.NullString{}) {
		var text string
		switch typed := value.(type) {
		case string:
			text = typed
		case uuid.UUID:
			text = typed.String()
		default:
			return fmt.Errorf("scan string value = %T", value)
		}
		target.Set(reflect.ValueOf(sql.NullString{String: text, Valid: true}))
		return nil
	}
	if target.Type() == reflect.TypeOf(sql.NullBool{}) {
		enabled, ok := value.(bool)
		if !ok {
			return errors.New(reflect.TypeOf(value).String())
		}
		target.Set(reflect.ValueOf(sql.NullBool{Bool: enabled, Valid: true}))
		return nil
	}
	if target.Type() == reflect.TypeOf(sql.NullInt64{}) {
		var number int64
		switch typed := value.(type) {
		case int:
			number = int64(typed)
		case int64:
			number = typed
		default:
			return fmt.Errorf("scan null int value = %T", value)
		}
		target.Set(reflect.ValueOf(sql.NullInt64{Int64: number, Valid: true}))
		return nil
	}
	if target.Type() == reflect.TypeOf(sql.NullTime{}) {
		instant, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("scan null time value = %T", value)
		}
		target.Set(reflect.ValueOf(sql.NullTime{Time: instant, Valid: true}))
		return nil
	}
	source := reflect.ValueOf(value)
	if source.Type().AssignableTo(target.Type()) {
		target.Set(source)
		return nil
	}
	if source.Type().ConvertibleTo(target.Type()) {
		target.Set(source.Convert(target.Type()))
		return nil
	}
	return fmt.Errorf("scan value %T does not fit %s", value, target.Type())
}
