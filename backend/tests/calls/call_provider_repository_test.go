package calls_test

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/call"
	calls "github.com/valerii/hub/backend/internal/calls"
)

type phase150CallProviderOperation struct {
	query string
	args  []any
}

type phase150CallProviderRecord struct {
	id       uuid.UUID
	provider calls.PersistedCallProvider
	roomName string
	status   call.Status
}

type phase150CallProviderDriver struct {
	operations []phase150CallProviderOperation
	records    []phase150CallProviderRecord
	queryErr   error
}

func (d *phase150CallProviderDriver) Exec(_ context.Context, _ string, _ any, _ any) error {
	return errors.New("unexpected exec")
}

func (d *phase150CallProviderDriver) Query(_ context.Context, query string, args, v any) error {
	values, ok := args.([]any)
	if !ok {
		return fmt.Errorf("unexpected query args type %T", args)
	}
	d.operations = append(d.operations, phase150CallProviderOperation{
		query: query,
		args:  append([]any(nil), values...),
	})
	if d.queryErr != nil {
		return d.queryErr
	}
	rows, ok := v.(*entsql.Rows)
	if !ok {
		return fmt.Errorf("unexpected query rows type %T", v)
	}
	columns, err := phase150CallProviderColumns(query)
	if err != nil {
		return err
	}
	*rows = entsql.Rows{ColumnScanner: &phase150CallProviderRows{
		columns: columns,
		records: append([]phase150CallProviderRecord(nil), d.records...),
	}}
	return nil
}

func (d *phase150CallProviderDriver) Tx(context.Context) (dialect.Tx, error) {
	return dialect.NopTx(d), nil
}

func (*phase150CallProviderDriver) Close() error {
	return nil
}

func (*phase150CallProviderDriver) Dialect() string {
	return dialect.Postgres
}

type phase150CallProviderRows struct {
	columns []string
	records []phase150CallProviderRecord
	index   int
}

func (*phase150CallProviderRows) Close() error {
	return nil
}

func (*phase150CallProviderRows) ColumnTypes() ([]*stdsql.ColumnType, error) {
	return nil, nil
}

func (r *phase150CallProviderRows) Columns() ([]string, error) {
	return append([]string(nil), r.columns...), nil
}

func (*phase150CallProviderRows) Err() error {
	return nil
}

func (r *phase150CallProviderRows) Next() bool {
	if r.index >= len(r.records) {
		return false
	}
	r.index++
	return true
}

func (*phase150CallProviderRows) NextResultSet() bool {
	return false
}

func (r *phase150CallProviderRows) Scan(dest ...any) error {
	if r.index == 0 || r.index > len(r.records) {
		return errors.New("scan without a current row")
	}
	if len(dest) != len(r.columns) {
		return fmt.Errorf("scan destinations: got %d, want %d", len(dest), len(r.columns))
	}
	for i, target := range dest {
		if err := phase150AssignCallProviderValue(target, r.columns[i], r.records[r.index-1]); err != nil {
			return fmt.Errorf("scan column %q: %w", r.columns[i], err)
		}
	}
	return nil
}

func phase150CallProviderColumns(query string) ([]string, error) {
	selectMatch := regexp.MustCompile(`(?is)\bselect\s+(.+?)\s+from\s+`).FindStringSubmatch(query)
	if len(selectMatch) != 2 {
		return nil, fmt.Errorf("cannot locate SELECT fields in %q", query)
	}
	fieldMatches := regexp.MustCompile(`"[^"]+"\."([^"]+)"`).FindAllStringSubmatch(selectMatch[1], -1)
	if len(fieldMatches) == 0 {
		return nil, fmt.Errorf("cannot locate quoted selected fields in %q", selectMatch[1])
	}
	columns := make([]string, 0, len(fieldMatches))
	for _, match := range fieldMatches {
		columns = append(columns, match[1])
	}
	return columns, nil
}

func phase150AssignCallProviderValue(
	target any,
	column string,
	record phase150CallProviderRecord,
) error {
	switch value := target.(type) {
	case *uuid.UUID:
		*value = phase150CallProviderUUID(column, record)
		return nil
	case *time.Time:
		*value = phase150CallProviderTime(column)
		return nil
	case *stdsql.NullString:
		value.String = phase150CallProviderString(column, record)
		value.Valid = true
		return nil
	case *stdsql.NullInt64:
		value.Int64 = 2
		value.Valid = true
		return nil
	case *stdsql.NullBool:
		value.Bool = true
		value.Valid = true
		return nil
	case *stdsql.NullTime:
		value.Time = phase150CallProviderTime(column)
		value.Valid = true
		return nil
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("unsupported scan destination %T", target)
	}
	element := value.Elem()
	switch element.Kind() {
	case reflect.String:
		element.SetString(phase150CallProviderString(column, record))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		element.SetInt(2)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		element.SetUint(2)
		return nil
	case reflect.Bool:
		element.SetBool(true)
		return nil
	case reflect.Float32, reflect.Float64:
		element.SetFloat(2)
		return nil
	case reflect.Slice:
		if element.Type().Elem().Kind() == reflect.Uint8 {
			element.SetBytes([]byte(phase150CallProviderString(column, record)))
			return nil
		}
	}
	if scanner, ok := target.(stdsql.Scanner); ok {
		return scanner.Scan(phase150CallProviderDatabaseValue(column, record))
	}
	return fmt.Errorf("unsupported scan destination %T", target)
}

func phase150CallProviderUUID(column string, record phase150CallProviderRecord) uuid.UUID {
	if column == "id" {
		return record.id
	}
	return uuid.MustParse("00000000-0000-4000-8000-000000000150")
}

func phase150CallProviderTime(string) time.Time {
	return time.Date(2026, time.September, 1, 8, 0, 0, 0, time.UTC)
}

func phase150CallProviderString(column string, record phase150CallProviderRecord) string {
	switch column {
	case "id":
		return record.id.String()
	case "channel_id", "created_by":
		return phase150CallProviderUUID(column, record).String()
	case "provider":
		return string(record.provider)
	case "room_name":
		return record.roomName
	case "status":
		return string(record.status)
	default:
		return "phase150-" + column
	}
}

func phase150CallProviderDatabaseValue(column string, record phase150CallProviderRecord) any {
	switch column {
	case "created_at", "updated_at", "started_at", "ended_at":
		return phase150CallProviderTime(column)
	case "participant_count":
		return int64(2)
	default:
		return phase150CallProviderString(column, record)
	}
}

func newPhase150CallProviderRepository(
	t *testing.T,
	driver *phase150CallProviderDriver,
) *calls.EntRepository {
	t.Helper()
	client := ent.NewClient(ent.Driver(driver))
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})
	return calls.NewEntRepository(client)
}

type phase150TypedNilContext struct{}

func (*phase150TypedNilContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (*phase150TypedNilContext) Done() <-chan struct{} {
	return nil
}

func (*phase150TypedNilContext) Err() error {
	return nil
}

func (*phase150TypedNilContext) Value(any) any {
	return nil
}

func TestPersistedCallProviderConstantsAreExactAndClosed(t *testing.T) {
	providers := []calls.PersistedCallProvider{
		calls.PersistedCallProviderLiveKit,
		calls.PersistedCallProviderJitsi,
	}

	require.Equal(t, []calls.PersistedCallProvider{
		"livekit",
		"jitsi",
	}, providers)
	require.NotEqual(t, calls.PersistedCallProviderLiveKit, calls.PersistedCallProviderJitsi)
}

func TestEntRepositoryGetCallByProviderAndRoomNameRejectsInvalidInputBeforeQuery(t *testing.T) {
	validProvider := calls.PersistedCallProviderLiveKit
	validRoom := "phase150-room"
	var nilContext context.Context
	var typedNilContext *phase150TypedNilContext

	cases := []struct {
		name     string
		repo     func(*testing.T, *phase150CallProviderDriver) *calls.EntRepository
		ctx      context.Context
		provider calls.PersistedCallProvider
		room     string
	}{
		{
			name:     "nil context",
			repo:     newPhase150CallProviderRepository,
			ctx:      nilContext,
			provider: validProvider,
			room:     validRoom,
		},
		{
			name:     "typed nil context",
			repo:     newPhase150CallProviderRepository,
			ctx:      typedNilContext,
			provider: validProvider,
			room:     validRoom,
		},
		{
			name: "nil repository",
			repo: func(*testing.T, *phase150CallProviderDriver) *calls.EntRepository {
				return nil
			},
			ctx:      context.Background(),
			provider: validProvider,
			room:     validRoom,
		},
		{
			name: "nil client",
			repo: func(*testing.T, *phase150CallProviderDriver) *calls.EntRepository {
				return calls.NewEntRepository(nil)
			},
			ctx:      context.Background(),
			provider: validProvider,
			room:     validRoom,
		},
		{
			name:     "blank provider",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: "",
			room:     validRoom,
		},
		{
			name:     "unknown provider",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: "custom",
			room:     validRoom,
		},
		{
			name:     "case drift provider",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: "LiveKit",
			room:     validRoom,
		},
		{
			name:     "leading padded provider",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: calls.PersistedCallProvider(" livekit"),
			room:     validRoom,
		},
		{
			name:     "trailing padded provider",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: calls.PersistedCallProvider("jitsi "),
			room:     validRoom,
		},
		{
			name:     "blank room",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: validProvider,
			room:     "",
		},
		{
			name:     "whitespace room",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: validProvider,
			room:     " ",
		},
		{
			name:     "leading padded room",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: validProvider,
			room:     " phase150-room",
		},
		{
			name:     "trailing padded room",
			repo:     newPhase150CallProviderRepository,
			ctx:      context.Background(),
			provider: validProvider,
			room:     "phase150-room ",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			driver := &phase150CallProviderDriver{}
			repo := testCase.repo(t, driver)

			got, err := repo.GetCallByProviderAndRoomName(testCase.ctx, testCase.provider, testCase.room)

			require.Nil(t, got)
			require.Same(t, calls.ErrCallProviderLookupInvalid, err)
			require.Empty(t, driver.operations)
		})
	}
}

func TestEntRepositoryGetCallByProviderAndRoomNameMapsExactProviderRoomRows(t *testing.T) {
	for _, provider := range []calls.PersistedCallProvider{
		calls.PersistedCallProviderLiveKit,
		calls.PersistedCallProviderJitsi,
	} {
		for _, status := range []call.Status{call.StatusActive, call.StatusEnded} {
			t.Run(string(provider)+"/"+string(status), func(t *testing.T) {
				roomName := "phase150-" + string(provider) + "-" + string(status)
				record := phase150CallProviderRecord{
					id:       uuid.New(),
					provider: provider,
					roomName: roomName,
					status:   status,
				}
				driver := &phase150CallProviderDriver{records: []phase150CallProviderRecord{record}}
				repo := newPhase150CallProviderRepository(t, driver)

				got, err := repo.GetCallByProviderAndRoomName(context.Background(), provider, roomName)

				require.NoError(t, err)
				require.NotNil(t, got)
				require.Equal(t, record.id, got.ID)
				require.Equal(t, record.roomName, got.RoomName)
				require.Equal(t, call.Provider(record.provider), got.Provider)
				require.Equal(t, record.status, got.Status)
				phase150RequireExactProviderRoomLookup(t, driver, provider, roomName)
			})
		}
	}
}

func TestEntRepositoryGetCallByProviderAndRoomNamePreservesEntCardinalityErrors(t *testing.T) {
	provider := calls.PersistedCallProviderJitsi
	roomName := "phase150-cardinality"
	record := phase150CallProviderRecord{
		id:       uuid.New(),
		provider: provider,
		roomName: roomName,
		status:   call.StatusEnded,
	}

	t.Run("no rows is Ent not found", func(t *testing.T) {
		driver := &phase150CallProviderDriver{}
		repo := newPhase150CallProviderRepository(t, driver)

		got, err := repo.GetCallByProviderAndRoomName(context.Background(), provider, roomName)

		require.Nil(t, got)
		require.True(t, ent.IsNotFound(err))
		phase150RequireExactProviderRoomLookup(t, driver, provider, roomName)
	})

	t.Run("multiple rows is Ent not singular", func(t *testing.T) {
		driver := &phase150CallProviderDriver{records: []phase150CallProviderRecord{
			record,
			{
				id:       uuid.New(),
				provider: provider,
				roomName: roomName,
				status:   call.StatusActive,
			},
		}}
		repo := newPhase150CallProviderRepository(t, driver)

		got, err := repo.GetCallByProviderAndRoomName(context.Background(), provider, roomName)

		require.Nil(t, got)
		require.True(t, ent.IsNotSingular(err))
		phase150RequireExactProviderRoomLookup(t, driver, provider, roomName)
	})
}

func TestEntRepositoryGetCallByProviderAndRoomNamePreservesDriverErrorIdentity(t *testing.T) {
	provider := calls.PersistedCallProviderLiveKit
	roomName := "phase150-driver-error"
	driverErr := errors.New("phase150 driver failure")
	driver := &phase150CallProviderDriver{queryErr: driverErr}
	repo := newPhase150CallProviderRepository(t, driver)

	got, err := repo.GetCallByProviderAndRoomName(context.Background(), provider, roomName)

	require.Nil(t, got)
	require.Same(t, driverErr, err)
	phase150RequireExactProviderRoomLookup(t, driver, provider, roomName)
}

func phase150RequireExactProviderRoomLookup(
	t *testing.T,
	driver *phase150CallProviderDriver,
	provider calls.PersistedCallProvider,
	roomName string,
) {
	t.Helper()
	require.Len(t, driver.operations, 1)
	operation := driver.operations[0]
	query := strings.ToLower(strings.Join(strings.Fields(operation.query), " "))
	require.Contains(t, query, "select")
	require.Contains(t, query, "from \"calls\"")
	where := phase150CallProviderWhereClause(t, query)
	require.Regexp(t, `"provider"\s*=\s*\$(\d+)`, where)
	require.Regexp(t, `"room_name"\s*=\s*\$(\d+)`, where)
	require.Equal(t, 1, strings.Count(where, " and "))
	require.NotContains(t, where, " or ")
	for _, forbidden := range []string{"status", "tenant", "channel", "deleted"} {
		require.NotContains(t, where, forbidden)
	}
	require.Len(t, operation.args, 2)
	phase150RequireCallProviderBinding(t, where, `"provider"\s*=\s*\$(\d+)`, operation.args, call.Provider(provider))
	phase150RequireCallProviderBinding(t, where, `"room_name"\s*=\s*\$(\d+)`, operation.args, roomName)
}

func phase150CallProviderWhereClause(t *testing.T, query string) string {
	t.Helper()
	whereStart := strings.Index(query, " where ")
	require.NotEqual(t, -1, whereStart, "missing WHERE clause: %s", query)
	where := query[whereStart+7:]
	if limitStart := strings.Index(where, " limit "); limitStart != -1 {
		where = where[:limitStart]
	}
	return where
}

func phase150RequireCallProviderBinding(
	t *testing.T,
	where string,
	expression string,
	args []any,
	want any,
) {
	t.Helper()
	match := regexp.MustCompile(expression).FindStringSubmatch(where)
	require.Len(t, match, 2, "missing binding %s in %s", expression, where)
	var argIndex int
	_, err := fmt.Sscanf(match[1], "%d", &argIndex)
	require.NoError(t, err)
	require.GreaterOrEqual(t, argIndex, 1)
	require.LessOrEqual(t, argIndex, len(args))
	require.Truef(
		t,
		reflect.DeepEqual(args[argIndex-1], want),
		"binding $%d is %#v, want %#v",
		argIndex,
		args[argIndex-1],
		want,
	)
}
