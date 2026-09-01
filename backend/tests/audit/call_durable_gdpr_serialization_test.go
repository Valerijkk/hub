package audit_test

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/callproviderdelivery"
	"github.com/valerii/hub/backend/ent/user"
	audit "github.com/valerii/hub/backend/internal/audit"
)

type phase150GDPRRepo struct {
	gdprStubRepo

	steps         []string
	preRole       user.SystemRole
	lockedRole    user.SystemRole
	preErr        error
	lockErr       error
	eraseErr      error
	clearErr      error
	auditErr      error
	hardDeleteErr error
	commitErr     error
	lockUserIDs   []uuid.UUID
	eraseUserIDs  []uuid.UUID
	terminalAts   []time.Time
}

func newPhase150GDPRRepo() *phase150GDPRRepo {
	return &phase150GDPRRepo{
		preRole:    user.SystemRole("member"),
		lockedRole: user.SystemRole("member"),
	}
}

func (r *phase150GDPRRepo) mark(step string) {
	r.steps = append(r.steps, step)
}

func (r *phase150GDPRRepo) GetUserSystemRole(context.Context, uuid.UUID) (user.SystemRole, error) {
	r.mark("precheck")
	if r.preErr != nil {
		return "", r.preErr
	}
	return r.preRole, nil
}

func (r *phase150GDPRRepo) WithTx(_ context.Context, fn func(audit.Repository) error) error {
	r.mark("tx-begin")
	if err := fn(r); err != nil {
		r.mark("tx-rollback")
		return err
	}
	r.mark("tx-commit")
	return r.commitErr
}

func (r *phase150GDPRRepo) LockUserForGDPR(_ context.Context, userID uuid.UUID) (user.SystemRole, error) {
	r.mark("lock")
	r.lockUserIDs = append(r.lockUserIDs, userID)
	if r.lockErr != nil {
		return "", r.lockErr
	}
	return r.lockedRole, nil
}

func (r *phase150GDPRRepo) EraseUserDurableCallData(
	_ context.Context,
	userID uuid.UUID,
	terminalAt time.Time,
) error {
	r.mark("erase")
	r.eraseUserIDs = append(r.eraseUserIDs, userID)
	r.terminalAts = append(r.terminalAts, terminalAt)
	return r.eraseErr
}

func (r *phase150GDPRRepo) ReassignAssignedIssues(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	r.mark("reassign-assigned")
	return 0, nil
}

func (r *phase150GDPRRepo) ReassignReportedIssues(context.Context, uuid.UUID, uuid.UUID) (int, error) {
	r.mark("reassign-reported")
	return 0, nil
}

func (r *phase150GDPRRepo) ClearUserAssignments(context.Context, uuid.UUID) (int, error) {
	r.mark("clear-assignments")
	return 0, r.clearErr
}

func (r *phase150GDPRRepo) DeleteUserComments(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-comments")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserReactions(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-reactions")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserSessions(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-sessions")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserActivityLogs(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-activity")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserTimeLogs(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-time")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserFiles(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-files")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserProjectMemberships(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-project-memberships")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserChannelMemberships(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-channel-memberships")
	return 0, nil
}

func (r *phase150GDPRRepo) DeleteUserTeamMemberships(context.Context, uuid.UUID) (int, error) {
	r.mark("delete-team-memberships")
	return 0, nil
}

func (r *phase150GDPRRepo) ScrubUserAuditPII(context.Context, uuid.UUID) (int, error) {
	r.mark("scrub-audit-pii")
	return 0, nil
}

func (r *phase150GDPRRepo) CreateAuditLog(
	context.Context,
	string,
	string,
	string,
	*uuid.UUID,
	*uuid.UUID,
	map[string]interface{},
	string,
	string,
) (*ent.AuditLog, error) {
	r.mark("audit")
	if r.auditErr != nil {
		return nil, r.auditErr
	}
	return &ent.AuditLog{}, nil
}

func (r *phase150GDPRRepo) HardDeleteUser(context.Context, uuid.UUID) error {
	r.mark("hard-delete")
	return r.hardDeleteErr
}

type phase150MessageEraser struct {
	repo  *phase150GDPRRepo
	calls []uuid.UUID
}

func (e *phase150MessageEraser) DeleteMessagesBySender(
	_ context.Context,
	userID uuid.UUID,
) (int64, error) {
	e.repo.mark("mongo")
	e.calls = append(e.calls, userID)
	return 0, nil
}

func newPhase150GDPRService(repo *phase150GDPRRepo) (*audit.AuditService, *phase150MessageEraser) {
	svc := audit.NewAuditService(repo, nil)
	eraser := &phase150MessageEraser{repo: repo}
	svc.SetMessageEraser(eraser)
	return svc, eraser
}

func phase150LegacySteps(prefix string) []string {
	return []string{
		"precheck",
		"tx-begin",
		"lock",
		"erase",
		prefix,
		"delete-comments",
		"delete-reactions",
		"delete-sessions",
		"delete-activity",
		"delete-time",
		"delete-files",
		"delete-project-memberships",
		"delete-channel-memberships",
		"delete-team-memberships",
		"scrub-audit-pii",
		"audit",
		"hard-delete",
		"tx-commit",
		"mongo",
	}
}

func TestCallDurableGDPRSerialization(t *testing.T) {
	t.Run("precheck error is fail fast", func(t *testing.T) {
		preErr := errors.New("precheck failed")
		repo := newPhase150GDPRRepo()
		repo.preErr = preErr
		svc, eraser := newPhase150GDPRService(repo)

		err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)

		require.ErrorIs(t, err, preErr)
		require.Equal(t, []string{"precheck"}, repo.steps)
		require.Empty(t, repo.lockUserIDs)
		require.Empty(t, repo.eraseUserIDs)
		require.Empty(t, eraser.calls)
	})

	t.Run("pre transaction founder remains fail fast", func(t *testing.T) {
		repo := newPhase150GDPRRepo()
		repo.preRole = user.SystemRole("founder")
		svc, eraser := newPhase150GDPRService(repo)

		err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)

		require.ErrorIs(t, err, audit.ErrCannotDeleteFounder)
		require.Equal(t, []string{"precheck"}, repo.steps)
		require.Empty(t, repo.lockUserIDs)
		require.Empty(t, repo.eraseUserIDs)
		require.Empty(t, eraser.calls)
	})

	t.Run("nil reassignment locks and erases before every legacy mutation", func(t *testing.T) {
		repo := newPhase150GDPRRepo()
		svc, eraser := newPhase150GDPRService(repo)
		userID := uuid.New()
		before := time.Now().UTC()

		err := svc.DeleteUserGDPR(context.Background(), uuid.New(), userID, nil)

		after := time.Now().UTC()
		require.NoError(t, err)
		require.Equal(t, phase150LegacySteps("clear-assignments"), repo.steps)
		require.Equal(t, []uuid.UUID{userID}, repo.lockUserIDs)
		require.Equal(t, []uuid.UUID{userID}, repo.eraseUserIDs)
		require.Equal(t, []uuid.UUID{userID}, eraser.calls)
		require.Len(t, repo.terminalAts, 1)
		require.False(t, repo.terminalAts[0].IsZero())
		require.Equal(t, time.UTC, repo.terminalAts[0].Location())
		require.False(t, repo.terminalAts[0].Before(before))
		require.False(t, repo.terminalAts[0].After(after))
	})

	t.Run("nonnull reassignment locks and erases before both reassign mutations", func(t *testing.T) {
		repo := newPhase150GDPRRepo()
		svc, eraser := newPhase150GDPRService(repo)
		userID := uuid.New()
		reassignTo := uuid.New()

		err := svc.DeleteUserGDPR(context.Background(), uuid.New(), userID, &reassignTo)

		require.NoError(t, err)
		expected := phase150LegacySteps("reassign-assigned")
		expected = append(expected[:5], append([]string{"reassign-reported"}, expected[5:]...)...)
		require.Equal(t, expected, repo.steps)
		require.Equal(t, []uuid.UUID{userID}, repo.lockUserIDs)
		require.Equal(t, []uuid.UUID{userID}, repo.eraseUserIDs)
		require.Equal(t, []uuid.UUID{userID}, eraser.calls)
		require.Len(t, repo.terminalAts, 1)
	})

	failureErr := errors.New("phase 150 failure")
	cases := []struct {
		name    string
		arrange func(*phase150GDPRRepo)
		want    error
		steps   []string
	}{
		{
			name: "locked founder stops before mutation",
			arrange: func(r *phase150GDPRRepo) {
				r.lockedRole = user.SystemRole("founder")
			},
			want:  audit.ErrCannotDeleteFounder,
			steps: []string{"precheck", "tx-begin", "lock", "tx-rollback"},
		},
		{
			name: "lock failure rolls back",
			arrange: func(r *phase150GDPRRepo) {
				r.lockErr = failureErr
			},
			want:  failureErr,
			steps: []string{"precheck", "tx-begin", "lock", "tx-rollback"},
		},
		{
			name: "erase failure rolls back",
			arrange: func(r *phase150GDPRRepo) {
				r.eraseErr = failureErr
			},
			want:  failureErr,
			steps: []string{"precheck", "tx-begin", "lock", "erase", "tx-rollback"},
		},
		{
			name: "callback failure rolls back",
			arrange: func(r *phase150GDPRRepo) {
				r.clearErr = failureErr
			},
			want: failureErr,
			steps: []string{
				"precheck", "tx-begin", "lock", "erase", "clear-assignments", "tx-rollback",
			},
		},
		{
			name: "audit failure rolls back",
			arrange: func(r *phase150GDPRRepo) {
				r.auditErr = failureErr
			},
			want: failureErr,
			steps: []string{
				"precheck",
				"tx-begin",
				"lock",
				"erase",
				"clear-assignments",
				"delete-comments",
				"delete-reactions",
				"delete-sessions",
				"delete-activity",
				"delete-time",
				"delete-files",
				"delete-project-memberships",
				"delete-channel-memberships",
				"delete-team-memberships",
				"scrub-audit-pii",
				"audit",
				"tx-rollback",
			},
		},
		{
			name: "hard delete failure rolls back",
			arrange: func(r *phase150GDPRRepo) {
				r.hardDeleteErr = failureErr
			},
			want: failureErr,
			steps: []string{
				"precheck",
				"tx-begin",
				"lock",
				"erase",
				"clear-assignments",
				"delete-comments",
				"delete-reactions",
				"delete-sessions",
				"delete-activity",
				"delete-time",
				"delete-files",
				"delete-project-memberships",
				"delete-channel-memberships",
				"delete-team-memberships",
				"scrub-audit-pii",
				"audit",
				"hard-delete",
				"tx-rollback",
			},
		},
		{
			name: "commit failure does not fake rollback and never dispatches mongo",
			arrange: func(r *phase150GDPRRepo) {
				r.commitErr = failureErr
			},
			want: failureErr,
			steps: []string{
				"precheck",
				"tx-begin",
				"lock",
				"erase",
				"clear-assignments",
				"delete-comments",
				"delete-reactions",
				"delete-sessions",
				"delete-activity",
				"delete-time",
				"delete-files",
				"delete-project-memberships",
				"delete-channel-memberships",
				"delete-team-memberships",
				"scrub-audit-pii",
				"audit",
				"hard-delete",
				"tx-commit",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			repo := newPhase150GDPRRepo()
			tt.arrange(repo)
			svc, eraser := newPhase150GDPRService(repo)

			err := svc.DeleteUserGDPR(context.Background(), uuid.New(), uuid.New(), nil)

			require.ErrorIs(t, err, tt.want)
			require.Equal(t, tt.steps, repo.steps)
			require.Empty(t, eraser.calls)
			require.NotContains(t, repo.steps, "mongo")
		})
	}
}

type phase150CallOperation struct {
	kind  string
	query string
	args  []any
}

type phase150CallDriver struct {
	operations []phase150CallOperation
	queryRole  user.SystemRole
	queryErr   error
	execErr    error
	failExecAt int
	execCount  int
}

func (d *phase150CallDriver) Exec(_ context.Context, query string, args, v any) error {
	values, ok := args.([]any)
	if !ok {
		return fmt.Errorf("unexpected exec args type %T", args)
	}
	d.operations = append(d.operations, phase150CallOperation{
		kind:  "exec",
		query: query,
		args:  append([]any(nil), values...),
	})
	execAt := d.execCount
	d.execCount++
	if d.execErr != nil && d.failExecAt == execAt {
		return d.execErr
	}
	result, ok := v.(*stdsql.Result)
	if !ok {
		return fmt.Errorf("unexpected exec result type %T", v)
	}
	*result = phase150Result(1)
	return nil
}

func (d *phase150CallDriver) Query(_ context.Context, query string, args, v any) error {
	values, ok := args.([]any)
	if !ok {
		return fmt.Errorf("unexpected query args type %T", args)
	}
	d.operations = append(d.operations, phase150CallOperation{
		kind:  "query",
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
	*rows = entsql.Rows{ColumnScanner: &phase150RoleRows{role: d.queryRole}}
	return nil
}

func (d *phase150CallDriver) Tx(context.Context) (dialect.Tx, error) {
	return dialect.NopTx(d), nil
}

func (*phase150CallDriver) Close() error {
	return nil
}

func (*phase150CallDriver) Dialect() string {
	return dialect.Postgres
}

type phase150Result int64

func (r phase150Result) LastInsertId() (int64, error) {
	return int64(r), nil
}

func (r phase150Result) RowsAffected() (int64, error) {
	return int64(r), nil
}

type phase150RoleRows struct {
	role    user.SystemRole
	yielded bool
}

func (*phase150RoleRows) Close() error {
	return nil
}

func (*phase150RoleRows) ColumnTypes() ([]*stdsql.ColumnType, error) {
	return nil, nil
}

func (*phase150RoleRows) Columns() ([]string, error) {
	return []string{"system_role"}, nil
}

func (*phase150RoleRows) Err() error {
	return nil
}

func (r *phase150RoleRows) Next() bool {
	if r.yielded {
		return false
	}
	r.yielded = true
	return true
}

func (*phase150RoleRows) NextResultSet() bool {
	return false
}

func (r *phase150RoleRows) Scan(dest ...any) error {
	if len(dest) != 1 {
		return fmt.Errorf("expected one scan destination, got %d", len(dest))
	}
	switch target := dest[0].(type) {
	case *string:
		*target = string(r.role)
		return nil
	case *user.SystemRole:
		*target = r.role
		return nil
	case *any:
		*target = string(r.role)
		return nil
	case stdsql.Scanner:
		return target.Scan(string(r.role))
	default:
		return fmt.Errorf("unsupported scan destination %T", dest[0])
	}
}

func newPhase150Repository(t *testing.T, driver *phase150CallDriver) audit.Repository {
	t.Helper()
	client := ent.NewClient(ent.Driver(driver))
	t.Cleanup(func() {
		require.NoError(t, client.Close())
	})
	return audit.NewRepository(client)
}

func TestCallDurableGDPRRepositoryLockAndEraseQueryShape(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	for _, role := range []user.SystemRole{
		user.SystemRole("member"),
		user.SystemRole("founder"),
	} {
		t.Run("lock returns "+string(role), func(t *testing.T) {
			driver := &phase150CallDriver{queryRole: role, failExecAt: -1}
			repo := newPhase150Repository(t, driver)

			gotRole, err := repo.LockUserForGDPR(ctx, userID)

			require.NoError(t, err)
			require.Equal(t, role, gotRole)
			require.Len(t, driver.operations, 1)
			require.Equal(t, "query", driver.operations[0].kind)
			query := phase150NormalizedSQL(driver.operations[0].query)
			require.Contains(t, query, "select")
			require.Contains(t, query, "from \"users\"")
			require.Contains(t, query, "for update")
			fromStart := strings.Index(query, " from ")
			require.Greater(t, fromStart, 0)
			selectClause := query[:fromStart]
			require.Contains(t, selectClause, "\"system_role\"")
			require.NotContains(t, selectClause, ",")
			phase150RequireBinding(
				t,
				query,
				"\"id\"\\s*=\\s*\\$(\\d+)",
				driver.operations[0].args,
				userID,
			)
			phase150RequireExactArgs(t, driver.operations[0].args, []any{userID})
		})
	}

	t.Run("lock query failure propagates", func(t *testing.T) {
		stopErr := errors.New("lock query failure")
		driver := &phase150CallDriver{queryErr: stopErr, failExecAt: -1}
		repo := newPhase150Repository(t, driver)

		_, err := repo.LockUserForGDPR(ctx, userID)

		require.ErrorIs(t, err, stopErr)
		require.Len(t, driver.operations, 1)
	})

	t.Run("erase has four exact statements and bindings", func(t *testing.T) {
		driver := &phase150CallDriver{failExecAt: -1}
		repo := newPhase150Repository(t, driver)
		terminalAt := time.Date(2026, time.September, 1, 4, 0, 0, 0, time.UTC)

		err := repo.EraseUserDurableCallData(ctx, userID, terminalAt)

		require.NoError(t, err)
		require.Len(t, driver.operations, 4)
		for _, operation := range driver.operations {
			require.Equal(t, "exec", operation.kind)
		}

		connectionDelete := phase150NormalizedSQL(driver.operations[0].query)
		require.Contains(t, connectionDelete, "delete from \"call_participant_connections\"")
		require.Contains(t, connectionDelete, "\"user_id\" = $1")
		phase150RequireExactArgs(t, driver.operations[0].args, []any{userID})

		presenceDelete := phase150NormalizedSQL(driver.operations[1].query)
		require.Contains(t, presenceDelete, "delete from \"call_participant_presences\"")
		require.Contains(t, presenceDelete, "\"user_id\" = $1")
		phase150RequireExactArgs(t, driver.operations[1].args, []any{userID})

		reconcileQuery := phase150NormalizedSQL(driver.operations[2].query)
		require.Contains(t, reconcileQuery, "update \"call_provider_deliveries\" set")
		reconcileSet, reconcileWhere := phase150SplitSetWhere(t, reconcileQuery)
		require.ElementsMatch(
			t,
			[]string{"result", "terminal_at", "user_id"},
			phase150SetColumns(reconcileSet),
		)
		require.Regexp(t, "\"user_id\"\\s*=\\s*null", reconcileSet)
		phase150RequireBinding(
			t,
			reconcileSet,
			"\"result\"\\s*=\\s*\\$(\\d+)",
			driver.operations[2].args,
			callproviderdelivery.ResultIgnored,
		)
		phase150RequireBinding(
			t,
			reconcileSet,
			"\"terminal_at\"\\s*=\\s*\\$(\\d+)",
			driver.operations[2].args,
			terminalAt,
		)
		phase150RequireBinding(
			t,
			reconcileWhere,
			"\"user_id\"\\s*=\\s*\\$(\\d+)",
			driver.operations[2].args,
			userID,
		)
		phase150RequireInBindings(
			t,
			reconcileWhere,
			"result",
			driver.operations[2].args,
			[]any{
				callproviderdelivery.ResultPending,
				callproviderdelivery.ResultReconcile,
			},
		)
		require.Len(t, driver.operations[2].args, 5)
		phase150RequireImmutableReplayColumnsAbsent(t, reconcileSet)

		terminalQuery := phase150NormalizedSQL(driver.operations[3].query)
		require.Contains(t, terminalQuery, "update \"call_provider_deliveries\" set")
		terminalSet, terminalWhere := phase150SplitSetWhere(t, terminalQuery)
		require.Equal(t, []string{"user_id"}, phase150SetColumns(terminalSet))
		require.Regexp(t, "\"user_id\"\\s*=\\s*null", terminalSet)
		phase150RequireBinding(
			t,
			terminalWhere,
			"\"user_id\"\\s*=\\s*\\$(\\d+)",
			driver.operations[3].args,
			userID,
		)
		phase150RequireInBindings(
			t,
			terminalWhere,
			"result",
			driver.operations[3].args,
			[]any{
				callproviderdelivery.ResultApplied,
				callproviderdelivery.ResultIgnored,
			},
		)
		require.Len(t, driver.operations[3].args, 3)
		phase150RequireImmutableReplayColumnsAbsent(t, terminalSet)
		require.False(t, phase150ContainsExactArg(driver.operations[3].args, terminalAt))
	})

	t.Run("stops on every failed durable statement", func(t *testing.T) {
		stopErr := errors.New("durable statement failure")
		for failAt := 0; failAt < 4; failAt++ {
			t.Run(fmt.Sprintf("statement-%d", failAt+1), func(t *testing.T) {
				driver := &phase150CallDriver{execErr: stopErr, failExecAt: failAt}
				repo := newPhase150Repository(t, driver)

				err := repo.EraseUserDurableCallData(
					ctx,
					userID,
					time.Date(2026, time.September, 1, 4, 0, 0, 0, time.UTC),
				)

				require.ErrorIs(t, err, stopErr)
				require.Len(t, driver.operations, failAt+1)
			})
		}
	})
}

func phase150NormalizedSQL(query string) string {
	return strings.ToLower(strings.Join(strings.Fields(query), " "))
}

func phase150SplitSetWhere(t *testing.T, query string) (string, string) {
	t.Helper()
	setStart := strings.Index(query, " set ")
	require.NotEqual(t, -1, setStart, "missing SET clause: %s", query)
	whereStart := strings.Index(query[setStart+5:], " where ")
	require.NotEqual(t, -1, whereStart, "missing WHERE clause: %s", query)
	whereStart += setStart + 5
	return query[setStart+5 : whereStart], query[whereStart+7:]
}

func phase150SetColumns(setClause string) []string {
	matches := regexp.MustCompile("\"([^\"]+)\"\\s*=").FindAllStringSubmatch(setClause, -1)
	columns := make([]string, 0, len(matches))
	for _, match := range matches {
		columns = append(columns, match[1])
	}
	return columns
}

func phase150RequireBinding(
	t *testing.T,
	clause string,
	expression string,
	args []any,
	want any,
) {
	t.Helper()
	match := regexp.MustCompile(expression).FindStringSubmatch(clause)
	require.Len(t, match, 2, "missing binding %s in %s", expression, clause)
	argIndex, err := strconv.Atoi(match[1])
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

func phase150RequireInBindings(
	t *testing.T,
	clause string,
	column string,
	args []any,
	want []any,
) {
	t.Helper()
	expression := regexp.MustCompile(
		regexp.QuoteMeta("\""+column+"\"") + "\\s+in\\s*\\(([^)]*)\\)",
	)
	match := expression.FindStringSubmatch(clause)
	require.Len(t, match, 2, "missing IN binding for %s in %s", column, clause)
	placeholders := regexp.MustCompile("\\$(\\d+)").FindAllStringSubmatch(match[1], -1)
	require.Len(t, placeholders, len(want))
	got := make([]any, 0, len(placeholders))
	for _, placeholder := range placeholders {
		argIndex, err := strconv.Atoi(placeholder[1])
		require.NoError(t, err)
		require.GreaterOrEqual(t, argIndex, 1)
		require.LessOrEqual(t, argIndex, len(args))
		got = append(got, args[argIndex-1])
	}
	require.Truef(
		t,
		reflect.DeepEqual(got, want),
		"IN bindings are %#v, want %#v",
		got,
		want,
	)
}

func phase150RequireExactArgs(t *testing.T, got []any, want []any) {
	t.Helper()
	require.Truef(
		t,
		reflect.DeepEqual(got, want),
		"arguments are %#v, want %#v",
		got,
		want,
	)
}

func phase150ContainsExactArg(args []any, want any) bool {
	for _, got := range args {
		if reflect.DeepEqual(got, want) {
			return true
		}
	}
	return false
}

func phase150RequireImmutableReplayColumnsAbsent(t *testing.T, setClause string) {
	t.Helper()
	for _, column := range []string{
		"provider",
		"delivery_id",
		"kind",
		"call_id",
		"epoch",
		"received_at",
	} {
		require.NotContains(t, setClause, "\""+column+"\"")
	}
}
