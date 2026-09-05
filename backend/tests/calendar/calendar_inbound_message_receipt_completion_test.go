package calendar_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
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
	mail "github.com/valerii/hub/backend/internal/mail"
)

var _ dialect.Driver = (*calendarInboundReceiptCompletionDriver)(nil)
var _ dialect.Tx = (*calendarInboundReceiptCompletionTx)(nil)
var _ calendar.CalendarInboundMessageRepository = calendarInboundReceiptCompletionRepository{}
var _ mail.MailboxCalendarPartReadEngine = (*calendarInboundReceiptCompletionEngine)(nil)

func TestCalendarInboundMessageReceiptCompletionPersistsOccurrenceSnapshotWithCAS(t *testing.T) {
	fixture := newCalendarInboundReceiptCompletionFixture(t)
	for _, testCase := range []struct {
		name            string
		scope           string
		recurrenceID    string
		wantPublicError bool
	}{
		{name: "unknown snapshot becomes occurrence"},
		{name: "same occurrence replay", scope: "occurrence", recurrenceID: fixture.recurrenceID},
		{name: "series snapshot conflicts", scope: "series", wantPublicError: true},
		{name: "other occurrence snapshot conflicts", scope: "occurrence", recurrenceID: "2026-04-03", wantPublicError: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			driver := &calendarInboundReceiptCompletionDriver{
				fixture:           fixture,
				receiptState:      "processing",
				receiptScope:      testCase.scope,
				receiptRecurrence: testCase.recurrenceID,
			}
			service, work := calendarInboundReceiptCompletionService(t, driver, fixture)

			err := service.ProcessCalendarInboundMessage(context.Background(), work)
			if testCase.wantPublicError {
				if !errors.Is(err, calendar.ErrEventInvalid) {
					t.Fatalf("ProcessCalendarInboundMessage() error = %v, want public invalid", err)
				}
				if driver.receiptUpdates != 0 || driver.eventCommits != 0 || driver.eventRollbacks != 1 {
					t.Fatalf("conflict completion update:%d commits:%d rollbacks:%d, want 0/0/1", driver.receiptUpdates, driver.eventCommits, driver.eventRollbacks)
				}
				if driver.receiptState != "processing" || driver.receiptScope != testCase.scope || driver.receiptRecurrence != testCase.recurrenceID {
					t.Fatalf("conflict mutated stored receipt state:%q scope:%q recurrence:%q", driver.receiptState, driver.receiptScope, driver.receiptRecurrence)
				}
			} else {
				if err != nil {
					t.Fatalf("ProcessCalendarInboundMessage() error = %v", err)
				}
				if driver.receiptUpdates != 1 || driver.eventCommits != 1 || driver.eventRollbacks != 0 {
					t.Fatalf("completion update:%d commits:%d rollbacks:%d, want 1/1/0", driver.receiptUpdates, driver.eventCommits, driver.eventRollbacks)
				}
				if driver.receiptState != "duplicate" || driver.receiptScope != "occurrence" || driver.receiptRecurrence != fixture.recurrenceID {
					t.Fatalf("stored receipt state:%q scope:%q recurrence:%q, want duplicate/occurrence/%q", driver.receiptState, driver.receiptScope, driver.receiptRecurrence, fixture.recurrenceID)
				}
				driver.assertCompletionCAS(t)
			}
			driver.assertQueryOrder(t)
		})
	}
}

type calendarInboundReceiptCompletionFixture struct {
	organizationID  uuid.UUID
	ownerID         uuid.UUID
	mailboxID       uuid.UUID
	messageID       uuid.UUID
	receiptID       uuid.UUID
	leaseToken      uuid.UUID
	envelopeID      uuid.UUID
	objectStateID   uuid.UUID
	calendarEventID uuid.UUID
	identity        calendar.ITIPIdentity
	wire            []byte
	recurrenceID    string
	locatorHash     [sha256.Size]byte
	clock           time.Time
}

func newCalendarInboundReceiptCompletionFixture(t *testing.T) calendarInboundReceiptCompletionFixture {
	t.Helper()
	const wire = "BEGIN:VCALENDAR\r\nPRODID:-//Hub//Test//EN\r\nVERSION:2.0\r\nMETHOD:REQUEST\r\nBEGIN:VEVENT\r\nUID:receipt-cas@example.test\r\nDTSTAMP:20260301T120000Z\r\nDTSTART;VALUE=DATE:20260402\r\nDTEND;VALUE=DATE:20260403\r\nRECURRENCE-ID;VALUE=DATE:20260402\r\nSEQUENCE:7\r\nSTATUS:CONFIRMED\r\nSUMMARY:Receipt CAS\r\nORGANIZER:mailto:owner@example.test\r\nATTENDEE;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;RSVP=FALSE:mailto:owner@example.test\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"
	message, err := calendar.ParseITIP([]byte(wire), calendar.DefaultITIPParseLimits())
	if err != nil {
		t.Fatalf("ParseITIP() error = %v", err)
	}
	identity, err := calendar.CanonicalITIPIdentity(message)
	if err != nil || !identity.HasRecurrenceID {
		t.Fatalf("CanonicalITIPIdentity() identity:%#v error:%v", identity, err)
	}
	var locatorHash [sha256.Size]byte
	for index := range locatorHash {
		locatorHash[index] = byte(index + 1)
	}
	return calendarInboundReceiptCompletionFixture{
		organizationID:  uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		ownerID:         uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		mailboxID:       uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		messageID:       uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		receiptID:       uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		leaseToken:      uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		envelopeID:      uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		objectStateID:   uuid.MustParse("88888888-8888-4888-8888-888888888888"),
		calendarEventID: uuid.MustParse("99999999-9999-4999-8999-999999999999"),
		identity:        identity,
		wire:            []byte(wire),
		recurrenceID:    identity.RecurrenceID,
		locatorHash:     locatorHash,
		clock:           time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC),
	}
}

func calendarInboundReceiptCompletionService(t *testing.T, driver *calendarInboundReceiptCompletionDriver, fixture calendarInboundReceiptCompletionFixture) (*calendar.CalendarInboundMessageService, calendar.CalendarInboundMessageWork) {
	t.Helper()
	mailRepository, err := mail.NewMailboxEntRepository(ent.NewClient(ent.Driver(driver)), calendarInboundReceiptCompletionClock{fixture.clock})
	if err != nil {
		t.Fatalf("NewMailboxEntRepository() error = %v", err)
	}
	source := mail.NewMailboxEntCalendarPartSource(mailRepository, &calendarInboundReceiptCompletionEngine{fixture: fixture})
	eventRepository := calendar.NewEventEntRepository(ent.NewClient(ent.Driver(driver)))
	eventService := calendar.NewEventService(eventRepository, nil, nil, nil)
	service, err := calendar.NewCalendarInboundMessageService(calendarInboundReceiptCompletionRepository{}, source, eventService)
	if err != nil {
		t.Fatalf("NewCalendarInboundMessageService() error = %v", err)
	}
	registration, err := calendar.NewCalendarInboundMessageRegistration(fixture.envelopeID, fixture.organizationID, fixture.mailboxID, fixture.messageID, fixture.locatorHash)
	if err != nil {
		t.Fatalf("NewCalendarInboundMessageRegistration() error = %v", err)
	}
	work, err := calendar.NewCalendarInboundMessageWork(registration, calendar.CalendarInboundMessageClaim{ReceiptID: fixture.receiptID, LeaseToken: fixture.leaseToken, LeaseGeneration: 1})
	if err != nil {
		t.Fatalf("NewCalendarInboundMessageWork() error = %v", err)
	}
	return service, work
}

type calendarInboundReceiptCompletionRepository struct{}

func (calendarInboundReceiptCompletionRepository) RegisterAndClaimCalendarInboundMessage(context.Context, calendar.CalendarInboundMessageRegistration) (calendar.CalendarInboundMessageClaimResult, error) {
	return calendar.CalendarInboundMessageClaimResult{}, calendar.ErrEventInvalid
}

func (calendarInboundReceiptCompletionRepository) ClaimDueCalendarInboundMessages(context.Context, int) ([]calendar.CalendarInboundMessageWork, error) {
	return nil, calendar.ErrEventInvalid
}

func (calendarInboundReceiptCompletionRepository) CleanupFinalizedCalendarInboundMessages(context.Context, int) (int, error) {
	return 0, calendar.ErrEventInvalid
}

type calendarInboundReceiptCompletionClock struct {
	now time.Time
}

func (clock calendarInboundReceiptCompletionClock) Now() time.Time {
	return clock.now
}

type calendarInboundReceiptCompletionEngine struct {
	fixture calendarInboundReceiptCompletionFixture
}

func (engine *calendarInboundReceiptCompletionEngine) ReadCalendarPart(_ context.Context, request mail.EngineCalendarPartReadRequest) (mail.EngineCalendarPartReadResult, error) {
	if request.AccountID != "engine-account" || request.MessageID != "engine-message" || request.Version != "engine-version" || request.LocatorHash != engine.fixture.locatorHash || request.Method != mail.MailCalendarMethodRequest {
		return mail.EngineCalendarPartReadResult{}, errors.New("unexpected calendar part engine request")
	}
	return mail.EngineCalendarPartReadResult{LocatorHash: engine.fixture.locatorHash, ITIP: append([]byte(nil), engine.fixture.wire...)}, nil
}

type calendarInboundReceiptCompletionDriver struct {
	fixture           calendarInboundReceiptCompletionFixture
	receiptState      string
	receiptScope      string
	receiptRecurrence string
	receiptUpdates    int
	eventCommits      int
	eventRollbacks    int
	eventQueries      []string
	sourceQueries     []string
	objectStateReads  int
	updateArguments   []any
	updateSQL         string
}

func (*calendarInboundReceiptCompletionDriver) Exec(context.Context, string, any, any) error {
	return errors.New("unexpected direct execution")
}

func (driver *calendarInboundReceiptCompletionDriver) Query(_ context.Context, query string, value any, destination any) error {
	return driver.query(false, query, value, destination)
}

func (driver *calendarInboundReceiptCompletionDriver) query(event bool, query string, _ any, destination any) error {
	kind := calendarInboundReceiptCompletionQueryKind(query)
	if kind == "" {
		return fmt.Errorf("unexpected query %q", query)
	}
	if event {
		driver.eventQueries = append(driver.eventQueries, kind)
	} else {
		driver.sourceQueries = append(driver.sourceQueries, kind)
	}
	rows, ok := destination.(*entsql.Rows)
	if !ok {
		return fmt.Errorf("query destination = %T", destination)
	}
	*rows = entsql.Rows{ColumnScanner: driver.rows(kind, event)}
	return nil
}

func (driver *calendarInboundReceiptCompletionDriver) Tx(context.Context) (dialect.Tx, error) {
	return &calendarInboundReceiptCompletionTx{driver: driver, event: true}, nil
}

func (driver *calendarInboundReceiptCompletionDriver) BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error) {
	return &calendarInboundReceiptCompletionTx{driver: driver}, nil
}

func (*calendarInboundReceiptCompletionDriver) Close() error {
	return nil
}

func (*calendarInboundReceiptCompletionDriver) Dialect() string {
	return dialect.Postgres
}

func (driver *calendarInboundReceiptCompletionDriver) rows(kind string, event bool) *calendarInboundReceiptCompletionRows {
	fixture := driver.fixture
	switch kind {
	case "clock":
		return calendarInboundReceiptCompletionRowsOf([]string{"clock"}, []any{fixture.clock})
	case "mail_messages":
		from, _ := json.Marshal(map[string]string{"email": "owner@example.test", "name": ""})
		return calendarInboundReceiptCompletionRowsOf(
			[]string{"id", "mail_account_id", "engine_message_id", "engine_version", "from", "calendar_method", "calendar_part_locator_hash", "authentication_class", "deleted_at"},
			[]any{fixture.messageID, fixture.mailboxID, "engine-message", "engine-version", from, "request", fixture.locatorHash[:], "dmarc_aligned", nil},
		)
	case "mail_accounts":
		hash := make([]byte, sha256.Size)
		for index := range hash {
			hash[index] = byte(index + 1)
		}
		return calendarInboundReceiptCompletionRowsOf(
			[]string{"id", "organization_id", "user_id", "engine_account_id", "desired_enabled", "status", "control_generation", "deployed_generation", "deployed_hash_version", "deployed_hash", "projection_state", "sync_cursor", "last_synced_at", "rebuild_checkpoint", "rebuild_base_cursor", "sync_error_code"},
			[]any{fixture.mailboxID, fixture.organizationID, fixture.ownerID, "engine-account", true, "active", int64(1), int64(1), int32(1), hash, "ready", "cursor", fixture.clock, nil, nil, nil},
		)
	case "users":
		return calendarInboundReceiptCompletionRowsOf([]string{"id", "status"}, []any{fixture.ownerID, "active"})
	case "organizations":
		return calendarInboundReceiptCompletionRowsOf([]string{"id"}, []any{fixture.organizationID})
	case "mail_identities":
		return calendarInboundReceiptCompletionRowsOf([]string{"id", "mail_account_id", "normalized_address", "kind", "status"}, []any{fixture.ownerID, fixture.mailboxID, "owner@example.test", "primary", "active"})
	case "calendar_inbound_message_receipts":
		var scope, recurrence any
		if driver.receiptScope != "" {
			scope = driver.receiptScope
			if driver.receiptRecurrence != "" {
				recurrence = driver.receiptRecurrence
			}
		}
		return calendarInboundReceiptCompletionRowsOf(
			[]string{"id", "organization_id", "mailbox_id", "mail_message_id", "part_locator_hash", "event_envelope_id", "state", "attempts", "lease_token", "lease_generation", "lease_until", "next_attempt_at", "error_code", "calendar_event_id", "deferred_object_state_id", "deferred_expires_at", "calendar_component_scope", "recurrence_id", "created_at", "updated_at"},
			[]any{fixture.receiptID, fixture.organizationID, fixture.mailboxID, fixture.messageID, fixture.locatorHash[:], fixture.envelopeID, driver.receiptState, 1, fixture.leaseToken, int64(1), fixture.clock.Add(time.Hour), nil, nil, nil, nil, nil, scope, recurrence, fixture.clock, fixture.clock},
		)
	case "calendar_inbound_object_states":
		if !event {
			return &calendarInboundReceiptCompletionRows{columns: []string{"id"}}
		}
		driver.objectStateReads++
		if driver.objectStateReads == 1 {
			return &calendarInboundReceiptCompletionRows{columns: []string{"id"}}
		}
		return calendarInboundReceiptCompletionRowsOf(
			[]string{"id", "organization_id", "calendar_owner_id", "uid", "recurrence_id", "organizer_email", "sequence", "method", "payload_hash_version", "payload_hash", "state", "calendar_event_id", "expires_at"},
			[]any{fixture.objectStateID, fixture.organizationID, fixture.ownerID, fixture.identity.UID, fixture.identity.RecurrenceID, "owner@example.test", fixture.identity.Sequence, "request", fixture.identity.CanonicalVersion, fixture.identity.PayloadHash[:], "applied", fixture.calendarEventID, nil},
		)
	default:
		panic(fmt.Sprintf("unsupported query kind %q", kind))
	}
}

func (driver *calendarInboundReceiptCompletionDriver) assertCompletionCAS(t *testing.T) {
	t.Helper()
	for _, fragment := range []string{`UPDATE "calendar_inbound_message_receipts"`, `"id"`, `"state"`, `"lease_token"`, `"lease_generation"`, `"lease_until"`, `"calendar_component_scope"`, `"recurrence_id"`} {
		if !strings.Contains(driver.updateSQL, fragment) {
			t.Fatalf("receipt completion SQL is missing %q: %s", fragment, driver.updateSQL)
		}
	}
	for _, value := range []any{driver.fixture.receiptID, "processing", driver.fixture.leaseToken, int64(1), driver.fixture.clock, "duplicate", "occurrence", driver.fixture.recurrenceID} {
		if !calendarInboundReceiptCompletionContainsValue(driver.updateArguments, value) {
			t.Fatalf("receipt completion CAS is missing %#v: %#v", value, driver.updateArguments)
		}
	}
}

func (driver *calendarInboundReceiptCompletionDriver) assertQueryOrder(t *testing.T) {
	t.Helper()
	wantEvent := []string{"calendar_inbound_message_receipts", "clock", "calendar_inbound_message_receipts", "mail_accounts", "organizations", "users", "mail_accounts", "calendar_inbound_object_states", "calendar_inbound_object_states", "calendar_inbound_message_receipts"}
	if !reflect.DeepEqual(driver.eventQueries, wantEvent) {
		t.Fatalf("event query order = %#v, want %#v", driver.eventQueries, wantEvent)
	}
	wantSource := []string{"mail_messages", "mail_accounts", "users", "mail_messages", "mail_accounts", "users"}
	if !reflect.DeepEqual(driver.sourceQueries, wantSource) {
		t.Fatalf("source query order = %#v, want %#v", driver.sourceQueries, wantSource)
	}
	if driver.objectStateReads != 2 {
		t.Fatalf("object-state reads = %d, want 2", driver.objectStateReads)
	}
}

func calendarInboundReceiptCompletionQueryKind(query string) string {
	switch {
	case strings.Contains(query, "date_trunc('second', clock_timestamp())"):
		return "clock"
	case strings.Contains(query, `FROM "mail_messages"`):
		return "mail_messages"
	case strings.Contains(query, `FROM "mail_accounts"`):
		return "mail_accounts"
	case strings.Contains(query, `FROM "users"`):
		return "users"
	case strings.Contains(query, `FROM "organizations"`):
		return "organizations"
	case strings.Contains(query, `FROM "mail_identities"`):
		return "mail_identities"
	case strings.Contains(query, `FROM "calendar_inbound_message_receipts"`):
		return "calendar_inbound_message_receipts"
	case strings.Contains(query, `FROM "calendar_inbound_object_states"`):
		return "calendar_inbound_object_states"
	default:
		return ""
	}
}

type calendarInboundReceiptCompletionTx struct {
	driver *calendarInboundReceiptCompletionDriver
	event  bool
}

func (transaction *calendarInboundReceiptCompletionTx) Exec(ctx context.Context, query string, value, destination any) error {
	if !transaction.event || !strings.Contains(query, `UPDATE "calendar_inbound_message_receipts"`) {
		return fmt.Errorf("unexpected execution %q", query)
	}
	arguments, ok := value.([]any)
	if !ok {
		return fmt.Errorf("update arguments = %T", value)
	}
	transaction.driver.receiptUpdates++
	transaction.driver.updateArguments = append([]any(nil), arguments...)
	transaction.driver.updateSQL = query
	transaction.driver.receiptState = "duplicate"
	transaction.driver.receiptScope = "occurrence"
	transaction.driver.receiptRecurrence = transaction.driver.fixture.recurrenceID
	result, ok := destination.(*sql.Result)
	if !ok {
		return fmt.Errorf("update result destination = %T", destination)
	}
	*result = calendarInboundReceiptCompletionResult(1)
	return nil
}

func (transaction *calendarInboundReceiptCompletionTx) Query(ctx context.Context, query string, value, destination any) error {
	return transaction.driver.query(transaction.event, query, value, destination)
}

func (transaction *calendarInboundReceiptCompletionTx) Commit() error {
	if transaction.event {
		transaction.driver.eventCommits++
	}
	return nil
}

func (transaction *calendarInboundReceiptCompletionTx) Rollback() error {
	if transaction.event {
		transaction.driver.eventRollbacks++
	}
	return nil
}

type calendarInboundReceiptCompletionResult int64

func (result calendarInboundReceiptCompletionResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (result calendarInboundReceiptCompletionResult) RowsAffected() (int64, error) {
	return int64(result), nil
}

type calendarInboundReceiptCompletionRows struct {
	columns []string
	values  [][]any
	index   int
}

func calendarInboundReceiptCompletionRowsOf(columns []string, values []any) *calendarInboundReceiptCompletionRows {
	return &calendarInboundReceiptCompletionRows{columns: columns, values: [][]any{values}}
}

func (*calendarInboundReceiptCompletionRows) Close() error {
	return nil
}

func (*calendarInboundReceiptCompletionRows) ColumnTypes() ([]*sql.ColumnType, error) {
	return nil, nil
}

func (rows *calendarInboundReceiptCompletionRows) Columns() ([]string, error) {
	return rows.columns, nil
}

func (*calendarInboundReceiptCompletionRows) Err() error {
	return nil
}

func (rows *calendarInboundReceiptCompletionRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++
	return true
}

func (*calendarInboundReceiptCompletionRows) NextResultSet() bool {
	return false
}

func (rows *calendarInboundReceiptCompletionRows) Scan(destinations ...any) error {
	if rows.index == 0 || rows.index > len(rows.values) || len(destinations) != len(rows.columns) {
		return errors.New("invalid row scan")
	}
	for index, value := range rows.values[rows.index-1] {
		if err := calendarInboundReceiptCompletionSetScanValue(destinations[index], value); err != nil {
			return err
		}
	}
	return nil
}

func calendarInboundReceiptCompletionSetScanValue(destination, value any) error {
	target := reflect.ValueOf(destination)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return fmt.Errorf("scan destination = %T", destination)
	}
	target = target.Elem()
	if value == nil {
		return nil
	}
	switch target.Type() {
	case reflect.TypeOf(sql.NullString{}):
		switch typed := value.(type) {
		case string:
			target.Set(reflect.ValueOf(sql.NullString{String: typed, Valid: true}))
		case uuid.UUID:
			target.Set(reflect.ValueOf(sql.NullString{String: typed.String(), Valid: true}))
		default:
			return fmt.Errorf("scan null string value = %T", value)
		}
		return nil
	case reflect.TypeOf(sql.NullTime{}):
		instant, ok := value.(time.Time)
		if !ok {
			return fmt.Errorf("scan null time value = %T", value)
		}
		target.Set(reflect.ValueOf(sql.NullTime{Time: instant, Valid: true}))
		return nil
	case reflect.TypeOf(sql.NullInt64{}):
		switch typed := value.(type) {
		case int:
			target.Set(reflect.ValueOf(sql.NullInt64{Int64: int64(typed), Valid: true}))
		case int64:
			target.Set(reflect.ValueOf(sql.NullInt64{Int64: typed, Valid: true}))
		default:
			return fmt.Errorf("scan null int value = %T", value)
		}
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

func calendarInboundReceiptCompletionContainsValue(values []any, want any) bool {
	for _, value := range values {
		if reflect.DeepEqual(value, want) {
			return true
		}
	}
	return false
}
