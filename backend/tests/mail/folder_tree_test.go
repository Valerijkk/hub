package mail_test

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/mail"
)

var (
	fixtureAccountID        = uuid.MustParse("00000000-0000-4000-8000-000000000001")
	fixtureForeignAccountID = uuid.MustParse("00000000-0000-4000-8000-000000000002")
	fixtureMissingParentID  = uuid.MustParse("00000000-0000-4000-8000-000000000003")
	fixtureMissingTargetID  = uuid.MustParse("00000000-0000-4000-8000-000000000004")

	_ func(mail.FolderTreeSnapshot) ([]mail.FolderTreeNode, error) = mail.NormalizeFolderTree
	_ func(mail.FolderTreeSnapshot, mail.FolderTreeChange) (mail.FolderTreeDecision, error) = mail.EvaluateFolderTreeChange
)

var fixedFolderIDs = []uuid.UUID{
	uuid.MustParse("00000000-0000-4000-8000-000000000101"),
	uuid.MustParse("00000000-0000-4000-8000-000000000102"),
	uuid.MustParse("00000000-0000-4000-8000-000000000103"),
	uuid.MustParse("00000000-0000-4000-8000-000000000104"),
	uuid.MustParse("00000000-0000-4000-8000-000000000105"),
	uuid.MustParse("00000000-0000-4000-8000-000000000106"),
	uuid.MustParse("00000000-0000-4000-8000-000000000107"),
	uuid.MustParse("00000000-0000-4000-8000-000000000108"),
	uuid.MustParse("00000000-0000-4000-8000-000000000109"),
	uuid.MustParse("00000000-0000-4000-8000-000000000110"),
	uuid.MustParse("00000000-0000-4000-8000-000000000111"),
	uuid.MustParse("00000000-0000-4000-8000-000000000112"),
	uuid.MustParse("00000000-0000-4000-8000-000000000113"),
	uuid.MustParse("00000000-0000-4000-8000-000000000114"),
	uuid.MustParse("00000000-0000-4000-8000-000000000115"),
	uuid.MustParse("00000000-0000-4000-8000-000000000116"),
	uuid.MustParse("00000000-0000-4000-8000-000000000117"),
	uuid.MustParse("00000000-0000-4000-8000-000000000118"),
	uuid.MustParse("00000000-0000-4000-8000-000000000119"),
	uuid.MustParse("00000000-0000-4000-8000-000000000120"),
	uuid.MustParse("00000000-0000-4000-8000-000000000121"),
	uuid.MustParse("00000000-0000-4000-8000-000000000122"),
	uuid.MustParse("00000000-0000-4000-8000-000000000123"),
	uuid.MustParse("00000000-0000-4000-8000-000000000124"),
	uuid.MustParse("00000000-0000-4000-8000-000000000125"),
	uuid.MustParse("00000000-0000-4000-8000-000000000126"),
	uuid.MustParse("00000000-0000-4000-8000-000000000127"),
	uuid.MustParse("00000000-0000-4000-8000-000000000128"),
	uuid.MustParse("00000000-0000-4000-8000-000000000129"),
	uuid.MustParse("00000000-0000-4000-8000-000000000130"),
	uuid.MustParse("00000000-0000-4000-8000-000000000131"),
	uuid.MustParse("00000000-0000-4000-8000-000000000132"),
	uuid.MustParse("00000000-0000-4000-8000-000000000133"),
	uuid.MustParse("00000000-0000-4000-8000-000000000134"),
	uuid.MustParse("00000000-0000-4000-8000-000000000135"),
	uuid.MustParse("00000000-0000-4000-8000-000000000136"),
	uuid.MustParse("00000000-0000-4000-8000-000000000137"),
	uuid.MustParse("00000000-0000-4000-8000-000000000138"),
	uuid.MustParse("00000000-0000-4000-8000-000000000139"),
	uuid.MustParse("00000000-0000-4000-8000-000000000140"),
	uuid.MustParse("00000000-0000-4000-8000-000000000141"),
	uuid.MustParse("00000000-0000-4000-8000-000000000142"),
	uuid.MustParse("00000000-0000-4000-8000-000000000143"),
	uuid.MustParse("00000000-0000-4000-8000-000000000144"),
	uuid.MustParse("00000000-0000-4000-8000-000000000145"),
	uuid.MustParse("00000000-0000-4000-8000-000000000146"),
	uuid.MustParse("00000000-0000-4000-8000-000000000147"),
	uuid.MustParse("00000000-0000-4000-8000-000000000148"),
	uuid.MustParse("00000000-0000-4000-8000-000000000149"),
	uuid.MustParse("00000000-0000-4000-8000-000000000150"),
	uuid.MustParse("00000000-0000-4000-8000-000000000151"),
	uuid.MustParse("00000000-0000-4000-8000-000000000152"),
	uuid.MustParse("00000000-0000-4000-8000-000000000153"),
	uuid.MustParse("00000000-0000-4000-8000-000000000154"),
	uuid.MustParse("00000000-0000-4000-8000-000000000155"),
	uuid.MustParse("00000000-0000-4000-8000-000000000156"),
	uuid.MustParse("00000000-0000-4000-8000-000000000157"),
	uuid.MustParse("00000000-0000-4000-8000-000000000158"),
	uuid.MustParse("00000000-0000-4000-8000-000000000159"),
	uuid.MustParse("00000000-0000-4000-8000-000000000160"),
	uuid.MustParse("00000000-0000-4000-8000-000000000161"),
	uuid.MustParse("00000000-0000-4000-8000-000000000162"),
	uuid.MustParse("00000000-0000-4000-8000-000000000163"),
	uuid.MustParse("00000000-0000-4000-8000-000000000164"),
	uuid.MustParse("00000000-0000-4000-8000-000000000165"),
	uuid.MustParse("00000000-0000-4000-8000-000000000166"),
	uuid.MustParse("00000000-0000-4000-8000-000000000167"),
	uuid.MustParse("00000000-0000-4000-8000-000000000168"),
	uuid.MustParse("00000000-0000-4000-8000-000000000169"),
	uuid.MustParse("00000000-0000-4000-8000-000000000170"),
	uuid.MustParse("00000000-0000-4000-8000-000000000171"),
	uuid.MustParse("00000000-0000-4000-8000-000000000172"),
	uuid.MustParse("00000000-0000-4000-8000-000000000173"),
	uuid.MustParse("00000000-0000-4000-8000-000000000174"),
	uuid.MustParse("00000000-0000-4000-8000-000000000175"),
	uuid.MustParse("00000000-0000-4000-8000-000000000176"),
	uuid.MustParse("00000000-0000-4000-8000-000000000177"),
	uuid.MustParse("00000000-0000-4000-8000-000000000178"),
	uuid.MustParse("00000000-0000-4000-8000-000000000179"),
	uuid.MustParse("00000000-0000-4000-8000-000000000180"),
	uuid.MustParse("00000000-0000-4000-8000-000000000181"),
	uuid.MustParse("00000000-0000-4000-8000-000000000182"),
	uuid.MustParse("00000000-0000-4000-8000-000000000183"),
	uuid.MustParse("00000000-0000-4000-8000-000000000184"),
	uuid.MustParse("00000000-0000-4000-8000-000000000185"),
	uuid.MustParse("00000000-0000-4000-8000-000000000186"),
	uuid.MustParse("00000000-0000-4000-8000-000000000187"),
	uuid.MustParse("00000000-0000-4000-8000-000000000188"),
	uuid.MustParse("00000000-0000-4000-8000-000000000189"),
	uuid.MustParse("00000000-0000-4000-8000-000000000190"),
	uuid.MustParse("00000000-0000-4000-8000-000000000191"),
	uuid.MustParse("00000000-0000-4000-8000-000000000192"),
	uuid.MustParse("00000000-0000-4000-8000-000000000193"),
	uuid.MustParse("00000000-0000-4000-8000-000000000194"),
	uuid.MustParse("00000000-0000-4000-8000-000000000195"),
	uuid.MustParse("00000000-0000-4000-8000-000000000196"),
	uuid.MustParse("00000000-0000-4000-8000-000000000197"),
	uuid.MustParse("00000000-0000-4000-8000-000000000198"),
	uuid.MustParse("00000000-0000-4000-8000-000000000199"),
	uuid.MustParse("00000000-0000-4000-8000-000000000200"),
	uuid.MustParse("00000000-0000-4000-8000-000000000201"),
}

func fields(t *testing.T, value any, names []string, types []reflect.Type) {
	t.Helper()
	typ := reflect.TypeOf(value)
	require.Equal(t, len(names), typ.NumField())
	for i := range names {
		f := typ.Field(i)
		require.Equal(t, names[i], f.Name)
		require.Equal(t, types[i], f.Type)
	}
}

func TestFolderTreeABIFields(t *testing.T) {
	uuidType := reflect.TypeOf(uuid.UUID{})
	roleType := reflect.TypeOf(mail.FolderTreeRole(""))
	refType := reflect.TypeOf(mail.FolderReferenceState(""))
	fields(t, mail.FolderTreeFolder{}, []string{"AccountID", "ID", "ParentID", "Role", "Name", "SortOrder"}, []reflect.Type{uuidType, uuidType, uuidType, roleType, reflect.TypeOf(""), reflect.TypeOf(int32(0))})
	fields(t, mail.FolderTreeSnapshot{}, []string{"AccountID", "Folders"}, []reflect.Type{uuidType, reflect.TypeOf([]mail.FolderTreeFolder{})})
	fields(t, mail.FolderTreeNode{}, []string{"AccountID", "ID", "ParentID", "Role", "Name", "NormalizedName", "SortOrder"}, []reflect.Type{uuidType, uuidType, uuidType, roleType, reflect.TypeOf(""), reflect.TypeOf(""), reflect.TypeOf(int32(0))})
	fields(t, mail.FolderTreeCreate{}, []string{"ID", "ParentID", "Name", "SortOrder"}, []reflect.Type{uuidType, uuidType, reflect.TypeOf(""), reflect.TypeOf(int32(0))})
	fields(t, mail.FolderTreeRename{}, []string{"FolderID", "Name"}, []reflect.Type{uuidType, reflect.TypeOf("")})
	fields(t, mail.FolderTreeReparent{}, []string{"FolderID", "ParentID"}, []reflect.Type{uuidType, uuidType})
	fields(t, mail.FolderTreeDelete{}, []string{"FolderID", "ActiveMoveRuleReference"}, []reflect.Type{uuidType, refType})
	fields(t, mail.FolderTreeChange{}, []string{"Create", "Rename", "Reparent", "Delete"}, []reflect.Type{reflect.TypeOf((*mail.FolderTreeCreate)(nil)), reflect.TypeOf((*mail.FolderTreeRename)(nil)), reflect.TypeOf((*mail.FolderTreeReparent)(nil)), reflect.TypeOf((*mail.FolderTreeDelete)(nil))})
	fields(t, mail.FolderTreeDecision{}, []string{"NormalizedName"}, []reflect.Type{reflect.TypeOf("")})
	fields(t, mail.FolderPolicyError{}, []string{"Kind"}, []reflect.Type{reflect.TypeOf(mail.FolderErrorKind(""))})
}

func TestFolderPolicyErrors(t *testing.T) {
	roleType := reflect.TypeOf(mail.FolderTreeRole(""))
	for _, value := range []any{
		mail.FolderTreeRoleInbox,
		mail.FolderTreeRoleSent,
		mail.FolderTreeRoleDrafts,
		mail.FolderTreeRoleArchive,
		mail.FolderTreeRoleTrash,
		mail.FolderTreeRoleJunk,
	} {
		require.Equal(t, roleType, reflect.TypeOf(value))
	}
	refType := reflect.TypeOf(mail.FolderReferenceState(""))
	for _, value := range []any{mail.FolderReferenceAbsent, mail.FolderReferencePresent} {
		require.Equal(t, refType, reflect.TypeOf(value))
	}
	errorKindType := reflect.TypeOf(mail.FolderErrorKind(""))
	for _, value := range []any{
		mail.FolderErrorInvalid,
		mail.FolderErrorTooLarge,
		mail.FolderErrorDuplicate,
		mail.FolderErrorNameConflict,
		mail.FolderErrorAccountMismatch,
		mail.FolderErrorNotFound,
		mail.FolderErrorParentUnavailable,
		mail.FolderErrorSelfParent,
		mail.FolderErrorCycle,
		mail.FolderErrorDepthExceeded,
		mail.FolderErrorSystemRole,
		mail.FolderErrorInUse,
		mail.FolderErrorLimitExceeded,
	} {
		require.Equal(t, errorKindType, reflect.TypeOf(value))
	}
	cases := []struct {
		kind     mail.FolderErrorKind
		literal  string
		text     string
		sentinel error
	}{
		{mail.FolderErrorInvalid, "invalid", "mail folder invalid", mail.ErrFolderInvalid},
		{mail.FolderErrorTooLarge, "too_large", "mail folder exceeds bounds", mail.ErrFolderTooLarge},
		{mail.FolderErrorDuplicate, "duplicate", "mail folder duplicate", mail.ErrFolderDuplicate},
		{mail.FolderErrorNameConflict, "name_conflict", "mail folder name conflict", mail.ErrFolderNameConflict},
		{mail.FolderErrorAccountMismatch, "account_mismatch", "mail folder account mismatch", mail.ErrFolderAccountMismatch},
		{mail.FolderErrorNotFound, "not_found", "mail folder not found", mail.ErrFolderNotFound},
		{mail.FolderErrorParentUnavailable, "parent_unavailable", "mail folder parent unavailable", mail.ErrFolderParentUnavailable},
		{mail.FolderErrorSelfParent, "self_parent", "mail folder self parent", mail.ErrFolderSelfParent},
		{mail.FolderErrorCycle, "cycle", "mail folder cycle", mail.ErrFolderCycle},
		{mail.FolderErrorDepthExceeded, "depth_exceeded", "mail folder depth exceeds bounds", mail.ErrFolderDepthExceeded},
		{mail.FolderErrorSystemRole, "system_role", "mail folder system role protected", mail.ErrFolderSystemRole},
		{mail.FolderErrorInUse, "in_use", "mail folder in use", mail.ErrFolderInUse},
		{mail.FolderErrorLimitExceeded, "limit_exceeded", "mail folder limit exceeded", mail.ErrFolderLimitExceeded},
	}
	require.Equal(t, "absent", string(mail.FolderReferenceAbsent))
	require.Equal(t, "present", string(mail.FolderReferencePresent))
	require.Equal(t, "", string(mail.FolderTreeRole("")))
	require.Equal(t, "inbox", string(mail.FolderTreeRoleInbox))
	require.Equal(t, "sent", string(mail.FolderTreeRoleSent))
	require.Equal(t, "drafts", string(mail.FolderTreeRoleDrafts))
	require.Equal(t, "archive", string(mail.FolderTreeRoleArchive))
	require.Equal(t, "trash", string(mail.FolderTreeRoleTrash))
	require.Equal(t, "junk", string(mail.FolderTreeRoleJunk))
	for _, tc := range cases {
		require.Equal(t, tc.literal, string(tc.kind))
		err := &mail.FolderPolicyError{Kind: tc.kind}
		require.Equal(t, tc.text, err.Error())
		require.True(t, errors.Is(err, tc.sentinel))
		var typed *mail.FolderPolicyError
		require.True(t, errors.As(err, &typed))
		require.Equal(t, tc.kind, typed.Kind)
		for _, other := range cases {
			if other.sentinel != tc.sentinel {
				require.False(t, errors.Is(err, other.sentinel))
			}
		}
	}
	var nilPolicy *mail.FolderPolicyError
	require.Equal(t, "mail folder policy error", nilPolicy.Error())
	require.Nil(t, nilPolicy.Unwrap())
	unknown := &mail.FolderPolicyError{Kind: mail.FolderErrorKind("unknown")}
	require.Equal(t, "mail folder policy error", unknown.Error())
	require.Nil(t, unknown.Unwrap())
}

func id(n int) uuid.UUID { return fixedFolderIDs[n] }

func snapshot100() mail.FolderTreeSnapshot {
	account := fixtureAccountID
	rows := make([]mail.FolderTreeFolder, 0, 100)
	roles := []mail.FolderTreeRole{mail.FolderTreeRoleInbox, mail.FolderTreeRoleSent, mail.FolderTreeRoleDrafts, mail.FolderTreeRoleArchive, mail.FolderTreeRoleTrash, mail.FolderTreeRoleJunk}
	for i := 0; i < 100; i++ {
		role := mail.FolderTreeRole("")
		name := "custom-" + string(rune('a'+(i%26))) + strings.Repeat("x", i/26)
		if i < len(roles) {
			role = roles[i]
			name = []string{"Inbox", "Sent", "Drafts", "Archive", "Trash", "Junk"}[i]
		}
		rows = append(rows, mail.FolderTreeFolder{AccountID: account, ID: id(i), Role: role, Name: name, SortOrder: int32(i)})
	}
	return mail.FolderTreeSnapshot{AccountID: account, Folders: rows}
}

func folderRow(folderID, parentID uuid.UUID, role mail.FolderTreeRole, name string) mail.FolderTreeFolder {
	return mail.FolderTreeFolder{AccountID: fixtureAccountID, ID: folderID, ParentID: parentID, Role: role, Name: name}
}

func snapshot(rows ...mail.FolderTreeFolder) mail.FolderTreeSnapshot {
	return mail.FolderTreeSnapshot{AccountID: fixtureAccountID, Folders: rows}
}

func requireNormalizeFailure(t *testing.T, input mail.FolderTreeSnapshot, sentinel error, text string) {
	t.Helper()
	before := cloneSnapshot(input)
	nodes, err := mail.NormalizeFolderTree(input)
	require.Nil(t, nodes)
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, text, err.Error())
	requireSnapshotEqual(t, before, input)
}

func requireChangeFailure(t *testing.T, input mail.FolderTreeSnapshot, change mail.FolderTreeChange, sentinel error, text string) {
	t.Helper()
	before := cloneSnapshot(input)
	beforeChange := cloneChange(change)
	decision, err := mail.EvaluateFolderTreeChange(input, change)
	require.Equal(t, mail.FolderTreeDecision{}, decision)
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, text, err.Error())
	requireSnapshotEqual(t, before, input)
	require.Equal(t, beforeChange, change)
}

func normalize(t *testing.T, input mail.FolderTreeSnapshot) []mail.FolderTreeNode {
	t.Helper()
	before := cloneSnapshot(input)
	nodes, err := mail.NormalizeFolderTree(input)
	require.NoError(t, err)
	requireSnapshotEqual(t, before, input)
	return nodes
}

func evaluate(t *testing.T, input mail.FolderTreeSnapshot, change mail.FolderTreeChange) mail.FolderTreeDecision {
	t.Helper()
	before := cloneSnapshot(input)
	beforeChange := cloneChange(change)
	decision, err := mail.EvaluateFolderTreeChange(input, change)
	require.NoError(t, err)
	requireSnapshotEqual(t, before, input)
	require.Equal(t, beforeChange, change)
	return decision
}

func TestNormalizeFolderTreeNamedBoundaryRows(t *testing.T) {
	postFoldRaw := strings.Repeat("\u1F88", 120)
	postFoldExpected := strings.Repeat("\u1F00\u03B9", 120)
	require.Equal(t, 120, utf8.RuneCountInString(postFoldRaw))
	require.Equal(t, 360, len(postFoldRaw))
	require.Equal(t, 600, len(postFoldExpected))

	zeroRowAccount := folderRow(id(0), uuid.Nil, "", "zero row account")
	zeroRowAccount.AccountID = uuid.Nil
	foreignRowAccount := folderRow(id(0), uuid.Nil, "", "foreign row account")
	foreignRowAccount.AccountID = fixtureForeignAccountID
	zeroRowID := folderRow(uuid.Nil, uuid.Nil, "", "zero row id")
	rows101 := snapshot100()
	rows101.Folders = append(rows101.Folders, folderRow(id(100), uuid.Nil, "", "one too many"))
	cases := []struct {
		name     string
		input    mail.FolderTreeSnapshot
		sentinel error
		text     string
		emptyOK  bool
	}{
		{name: "valid-empty-snapshot", input: mail.FolderTreeSnapshot{AccountID: fixtureAccountID, Folders: []mail.FolderTreeFolder{}}, emptyOK: true},
		{name: "zero-snapshot-account", input: mail.FolderTreeSnapshot{Folders: []mail.FolderTreeFolder{}}, sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
		{name: "zero-row-account", input: snapshot(zeroRowAccount), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
		{name: "foreign-row-account", input: snapshot(foreignRowAccount), sentinel: mail.ErrFolderAccountMismatch, text: "mail folder account mismatch"},
		{name: "zero-row-id", input: snapshot(zeroRowID), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
		{name: "101-rows", input: rows101, sentinel: mail.ErrFolderTooLarge, text: "mail folder exceeds bounds"},
		{name: "nul-name", input: snapshot(folderRow(id(0), uuid.Nil, "", "bad\x00name")), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
		{name: "c1-control-name", input: snapshot(folderRow(id(0), uuid.Nil, "", "bad\u0085name")), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
		{name: "invalid-utf8-name", input: snapshot(folderRow(id(0), uuid.Nil, "", string([]byte{0xff}))), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
		{name: "121-rune-name", input: snapshot(folderRow(id(0), uuid.Nil, "", strings.Repeat("a", 121))), sentinel: mail.ErrFolderTooLarge, text: "mail folder exceeds bounds"},
		{name: "post-fold-key-overflow", input: snapshot(folderRow(id(0), uuid.Nil, "", postFoldRaw)), sentinel: mail.ErrFolderTooLarge, text: "mail folder exceeds bounds"},
		{name: "duplicate-snapshot-id", input: snapshot(folderRow(id(0), uuid.Nil, "", "first"), folderRow(id(0), uuid.Nil, "", "second")), sentinel: mail.ErrFolderDuplicate, text: "mail folder duplicate"},
		{name: "duplicate-inbox-role", input: snapshot(folderRow(id(0), uuid.Nil, mail.FolderTreeRoleInbox, "Inbox"), folderRow(id(1), uuid.Nil, mail.FolderTreeRoleInbox, "Inbox second")), sentinel: mail.ErrFolderDuplicate, text: "mail folder duplicate"},
		{name: "missing-parent", input: snapshot(folderRow(id(0), fixtureMissingParentID, "", "child")), sentinel: mail.ErrFolderParentUnavailable, text: "mail folder parent unavailable"},
		{name: "whitespace-key", input: snapshot(folderRow(id(0), uuid.Nil, "", "\u00a0\t")), sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.emptyOK {
				nodes := normalize(t, tc.input)
				require.NotNil(t, nodes)
				require.Len(t, nodes, 0)
				return
			}
			requireNormalizeFailure(t, tc.input, tc.sentinel, tc.text)
		})
	}
}

func TestNormalizeFolderTreeNormalizesNamesAndKeepsRoot(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{name: "nfc", raw: "A\u030a", want: "\u00e5"},
		{name: "collapse-trim", raw: "\tCafe\u0301\u00a0", want: "caf\u00e9"},
		{name: "fold-expansion", raw: "Stra\u00dfe", want: "strasse"},
		{name: "sigma", raw: "\u03a3", want: "\u03c3"},
		{name: "dotted-i-expansion", raw: "\u0130", want: "i\u0307"},
	}
	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := snapshot(folderRow(id(index), uuid.Nil, "", tc.raw))
			nodes := normalize(t, input)
			require.Equal(t, []mail.FolderTreeNode{{
				AccountID: fixtureAccountID,
				ID:        id(index),
				Name:      tc.raw,
				NormalizedName: tc.want,
			}}, nodes)
		})
	}
}

func TestNormalizeFolderTreeRejectsSelfCycleAndDepth(t *testing.T) {
	requireNormalizeFailure(t, snapshot(folderRow(id(0), id(0), "", "self")), mail.ErrFolderSelfParent, "mail folder self parent")
	requireNormalizeFailure(t, snapshot(
		folderRow(id(0), id(1), "", "one"),
		folderRow(id(1), id(0), "", "two"),
	), mail.ErrFolderCycle, "mail folder cycle")
	requireNormalizeFailure(t, snapshot(
		folderRow(id(0), uuid.Nil, "", "root"),
		folderRow(id(1), id(0), "", "one"),
		folderRow(id(2), id(1), "", "two"),
		folderRow(id(3), id(2), "", "three"),
	), mail.ErrFolderDepthExceeded, "mail folder depth exceeds bounds")
}

func TestNormalizeFolderTreeSeparatesSystemAndCustomNameNamespaces(t *testing.T) {
	nodes := normalize(t, snapshot(
		folderRow(id(0), uuid.Nil, mail.FolderTreeRoleInbox, "Inbox"),
		folderRow(id(1), uuid.Nil, "", "inbox"),
	))
	require.Len(t, nodes, 2)
	require.Equal(t, mail.FolderTreeRoleInbox, nodes[0].Role)
	require.Equal(t, "inbox", nodes[0].NormalizedName)
	require.Equal(t, mail.FolderTreeRole(""), nodes[1].Role)
	require.Equal(t, "inbox", nodes[1].NormalizedName)
}

func TestNormalizeFolderTreeRejectsCustomSiblingNameCollision(t *testing.T) {
	requireNormalizeFailure(t, snapshot(
		folderRow(id(0), uuid.Nil, "", "parent"),
		folderRow(id(1), id(0), "", "\u00a0Stra\u00dfe"),
		folderRow(id(2), id(0), "", "STRASSE\t"),
	), mail.ErrFolderNameConflict, "mail folder name conflict")
}

func TestNormalizeFolderTreeOrdersSystemThenCustomBySortKeyAndRawUUID(t *testing.T) {
	roles := []mail.FolderTreeRole{
		mail.FolderTreeRoleJunk,
		mail.FolderTreeRoleArchive,
		mail.FolderTreeRoleInbox,
		mail.FolderTreeRoleTrash,
		mail.FolderTreeRoleSent,
		mail.FolderTreeRoleDrafts,
	}
	names := map[mail.FolderTreeRole]string{
		mail.FolderTreeRoleInbox:   "Inbox",
		mail.FolderTreeRoleSent:    "Sent",
		mail.FolderTreeRoleDrafts:  "Drafts",
		mail.FolderTreeRoleArchive: "Archive",
		mail.FolderTreeRoleTrash:   "Trash",
		mail.FolderTreeRoleJunk:    "Junk",
	}
	rows := make([]mail.FolderTreeFolder, 0, 10)
	for i, role := range roles {
		rows = append(rows, mail.FolderTreeFolder{AccountID: fixtureAccountID, ID: id(i), Role: role, Name: names[role], SortOrder: int32(99 - i)})
	}
	rows = append(rows,
		mail.FolderTreeFolder{AccountID: fixtureAccountID, ID: id(7), Name: "zeta", SortOrder: 0},
		mail.FolderTreeFolder{AccountID: fixtureAccountID, ID: id(9), Name: "alpha", SortOrder: 1},
		mail.FolderTreeFolder{AccountID: fixtureAccountID, ID: id(8), Name: "alpha", SortOrder: 1},
		mail.FolderTreeFolder{AccountID: fixtureAccountID, ID: id(6), Name: "beta", SortOrder: 1},
	)
	nodes := normalize(t, snapshot(rows...))
	require.Equal(t, []uuid.UUID{id(2), id(4), id(5), id(1), id(3), id(0), id(7), id(8), id(9), id(6)}, nodeIDs(nodes))
}

func TestNormalizeFolderTreeDoesNotAliasInputOrOutput(t *testing.T) {
	original := snapshot(folderRow(id(0), uuid.Nil, "", "Original"))
	input := cloneSnapshot(original)
	nodes := normalize(t, input)
	input.Folders[0].Name = "input mutation"
	require.Equal(t, "Original", nodes[0].Name)
	require.Equal(t, "original", nodes[0].NormalizedName)
	nodes[0].Name = "returned mutation"
	nodes[0].NormalizedName = "returned mutation"
	next := normalize(t, original)
	require.Equal(t, "Original", next[0].Name)
	require.Equal(t, "original", next[0].NormalizedName)
}

func TestFolderTreeNamedMutationFailureRows(t *testing.T) {
	base := snapshot(folderRow(id(0), uuid.Nil, "", "existing"))
	cases := []struct {
		name     string
		input    mail.FolderTreeSnapshot
		change   mail.FolderTreeChange
		sentinel error
		text     string
	}{
		{name: "rename-missing-target", input: snapshot(), change: mail.FolderTreeChange{Rename: &mail.FolderTreeRename{FolderID: fixtureMissingTargetID, Name: "renamed"}}, sentinel: mail.ErrFolderNotFound, text: "mail folder not found"},
		{name: "delete-missing-target", input: snapshot(), change: mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: fixtureMissingTargetID, ActiveMoveRuleReference: mail.FolderReferenceAbsent}}, sentinel: mail.ErrFolderNotFound, text: "mail folder not found"},
		{name: "create-duplicate-id-below-capacity", input: base, change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(0), Name: "new custom"}}, sentinel: mail.ErrFolderDuplicate, text: "mail folder duplicate"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireChangeFailure(t, tc.input, tc.change, tc.sentinel, tc.text)
		})
	}
}

func TestCreateCapacityPrecedence(t *testing.T) {
	full := snapshot100()
	cases := []struct {
		name     string
		change   mail.FolderTreeChange
		sentinel error
		text     string
	}{
		{name: "zero-id-before-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: uuid.Nil, Name: "valid"}}, sentinel: mail.ErrFolderInvalid, text: "mail folder invalid"},
		{name: "valid-raw-name-at-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(100), Name: strings.Repeat("x", 480)}}, sentinel: mail.ErrFolderLimitExceeded, text: "mail folder limit exceeded"},
		{name: "raw-byte-overflow-before-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(100), Name: strings.Repeat("x", 481)}}, sentinel: mail.ErrFolderTooLarge, text: "mail folder exceeds bounds"},
		{name: "malformed-within-raw-bound-at-capacity", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(100), Name: "\xff"}}, sentinel: mail.ErrFolderLimitExceeded, text: "mail folder limit exceeded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireChangeFailure(t, full, tc.change, tc.sentinel, tc.text)
		})
	}
}

func TestFolderTreeChangeRequiresExactlyOneActionAndNonzeroTargets(t *testing.T) {
	base := snapshot(folderRow(id(0), uuid.Nil, "", "existing"))
	cases := []struct {
		name   string
		change mail.FolderTreeChange
	}{
		{name: "no-action", change: mail.FolderTreeChange{}},
		{name: "two-actions", change: mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(1), Name: "new"}, Rename: &mail.FolderTreeRename{FolderID: id(0), Name: "rename"}}},
		{name: "zero-rename-target", change: mail.FolderTreeChange{Rename: &mail.FolderTreeRename{Name: "rename"}}},
		{name: "zero-reparent-target", change: mail.FolderTreeChange{Reparent: &mail.FolderTreeReparent{ParentID: id(0)}}},
		{name: "zero-delete-target", change: mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{ActiveMoveRuleReference: mail.FolderReferenceAbsent}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			requireChangeFailure(t, base, tc.change, mail.ErrFolderInvalid, "mail folder invalid")
		})
	}
}

func TestFolderTreeChangeReturnsOnlyDocumentedDecisions(t *testing.T) {
	base := snapshot(
		folderRow(id(0), uuid.Nil, "", "parent"),
		folderRow(id(1), id(0), "", "child"),
		folderRow(id(2), uuid.Nil, "", "rename me"),
	)
	decision := evaluate(t, base, mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(3), Name: "\tCafe\u0301\u00a0", SortOrder: 7}})
	require.Equal(t, mail.FolderTreeDecision{NormalizedName: "caf\u00e9"}, decision)
	decision = evaluate(t, base, mail.FolderTreeChange{Rename: &mail.FolderTreeRename{FolderID: id(2), Name: "rename me"}})
	require.Equal(t, mail.FolderTreeDecision{NormalizedName: "rename me"}, decision)
	decision = evaluate(t, base, mail.FolderTreeChange{Reparent: &mail.FolderTreeReparent{FolderID: id(1), ParentID: uuid.Nil}})
	require.Equal(t, mail.FolderTreeDecision{}, decision)
	decision = evaluate(t, base, mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: id(2), ActiveMoveRuleReference: mail.FolderReferenceAbsent}})
	require.Equal(t, mail.FolderTreeDecision{}, decision)
}

func TestFolderTreeChangeRevalidatesGraphAndDeleteReferences(t *testing.T) {
	graph := snapshot(
		folderRow(id(0), uuid.Nil, "", "root"),
		folderRow(id(1), id(0), "", "child"),
		folderRow(id(2), id(1), "", "grandchild"),
	)
	requireChangeFailure(t, graph, mail.FolderTreeChange{Reparent: &mail.FolderTreeReparent{FolderID: id(0), ParentID: id(2)}}, mail.ErrFolderCycle, "mail folder cycle")
	requireChangeFailure(t, graph, mail.FolderTreeChange{Reparent: &mail.FolderTreeReparent{FolderID: id(1), ParentID: fixtureMissingParentID}}, mail.ErrFolderParentUnavailable, "mail folder parent unavailable")
	requireChangeFailure(t, graph, mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: id(0), ActiveMoveRuleReference: mail.FolderReferenceAbsent}}, mail.ErrFolderInUse, "mail folder in use")
	requireChangeFailure(t, graph, mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: id(2), ActiveMoveRuleReference: mail.FolderReferencePresent}}, mail.ErrFolderInUse, "mail folder in use")
	requireChangeFailure(t, graph, mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: id(2), ActiveMoveRuleReference: "unknown"}}, mail.ErrFolderInvalid, "mail folder invalid")
	requireChangeFailure(t, graph, mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: id(2)}}, mail.ErrFolderInvalid, "mail folder invalid")
}

func TestFolderTreeSystemActionsTakePrecedence(t *testing.T) {
	system := snapshot(folderRow(id(0), uuid.Nil, mail.FolderTreeRoleInbox, "Inbox"))
	requireChangeFailure(t, system, mail.FolderTreeChange{Rename: &mail.FolderTreeRename{FolderID: id(0), Name: "\xff"}}, mail.ErrFolderSystemRole, "mail folder system role protected")
	requireChangeFailure(t, system, mail.FolderTreeChange{Reparent: &mail.FolderTreeReparent{FolderID: id(0), ParentID: fixtureMissingParentID}}, mail.ErrFolderSystemRole, "mail folder system role protected")
	requireChangeFailure(t, system, mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: id(0), ActiveMoveRuleReference: "unknown"}}, mail.ErrFolderSystemRole, "mail folder system role protected")
}

func TestFolderTreeDeleteDoesNotChangeFullSnapshot(t *testing.T) {
	full := snapshot100()
	decision := evaluate(t, full, mail.FolderTreeChange{Delete: &mail.FolderTreeDelete{FolderID: id(99), ActiveMoveRuleReference: mail.FolderReferenceAbsent}})
	require.Equal(t, mail.FolderTreeDecision{}, decision)
	requireChangeFailure(t, full, mail.FolderTreeChange{Create: &mail.FolderTreeCreate{ID: id(100), Name: "still full"}}, mail.ErrFolderLimitExceeded, "mail folder limit exceeded")
}

func cloneSnapshot(value mail.FolderTreeSnapshot) mail.FolderTreeSnapshot {
	clone := value
	clone.Folders = make([]mail.FolderTreeFolder, len(value.Folders))
	copy(clone.Folders, value.Folders)
	return clone
}

func cloneChange(value mail.FolderTreeChange) mail.FolderTreeChange {
	clone := value
	if value.Create != nil {
		create := *value.Create
		clone.Create = &create
	}
	if value.Rename != nil {
		rename := *value.Rename
		clone.Rename = &rename
	}
	if value.Reparent != nil {
		reparent := *value.Reparent
		clone.Reparent = &reparent
	}
	if value.Delete != nil {
		delete := *value.Delete
		clone.Delete = &delete
	}
	return clone
}

func requireSnapshotEqual(t *testing.T, want, got mail.FolderTreeSnapshot) {
	t.Helper()
	require.True(t, bytes.Equal(snapshotBytes(want), snapshotBytes(got)))
}

func snapshotBytes(value mail.FolderTreeSnapshot) []byte {
	var buffer bytes.Buffer
	writeUUID(&buffer, value.AccountID)
	writeLength(&buffer, len(value.Folders))
	for _, row := range value.Folders {
		writeUUID(&buffer, row.AccountID)
		writeUUID(&buffer, row.ID)
		writeUUID(&buffer, row.ParentID)
		writeString(&buffer, string(row.Role))
		writeString(&buffer, row.Name)
		order := uint32(row.SortOrder)
		buffer.WriteByte(byte(order))
		buffer.WriteByte(byte(order >> 8))
		buffer.WriteByte(byte(order >> 16))
		buffer.WriteByte(byte(order >> 24))
	}
	return buffer.Bytes()
}

func writeUUID(buffer *bytes.Buffer, value uuid.UUID) {
	buffer.Write(value[:])
}

func writeString(buffer *bytes.Buffer, value string) {
	writeLength(buffer, len(value))
	buffer.WriteString(value)
}

func writeLength(buffer *bytes.Buffer, value int) {
	length := uint64(value)
	for shift := 0; shift < 64; shift += 8 {
		buffer.WriteByte(byte(length >> shift))
	}
}

func nodeIDs(nodes []mail.FolderTreeNode) []uuid.UUID {
	ids := make([]uuid.UUID, len(nodes))
	for i := range nodes {
		ids[i] = nodes[i].ID
	}
	return ids
}

func sourcePath(t *testing.T) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "mail", "folder_tree.go"))
}

func TestFolderTreeSourceASTContract(t *testing.T) {
	path := sourcePath(t)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, data, parser.ParseComments)
	require.NoError(t, err)
	require.Equal(t, "mail", file.Name.Name)

	allowedImports := map[string]bool{
		"bytes": true, "errors": true, "sort": true, "strings": true,
		"unicode": true, "unicode/utf8": true, "github.com/google/uuid": true,
		"golang.org/x/text/cases": true, "golang.org/x/text/unicode/norm": true,
	}
	for _, imp := range file.Imports {
		require.True(t, allowedImports[strings.Trim(imp.Path.Value, "\"")])
	}
	bindings := importPathBindings(file)
	require.False(t, bindings["github.com/google/uuid"]["."])
	require.False(t, bindings["errors"]["."])

	allowedTypes := map[string]bool{"FolderTreeRole": true, "FolderTreeFolder": true, "FolderTreeSnapshot": true, "FolderTreeNode": true, "FolderTreeCreate": true, "FolderTreeRename": true, "FolderTreeReparent": true, "FolderReferenceState": true, "FolderTreeDelete": true, "FolderTreeChange": true, "FolderTreeDecision": true, "FolderErrorKind": true, "FolderPolicyError": true}
	allowedFunctions := map[string]bool{"NormalizeFolderTree": true, "EvaluateFolderTreeChange": true}
	allowedMethods := map[string]bool{"Error": true, "Unwrap": true}
	allowedConstants := map[string]bool{"FolderTreeRoleInbox": true, "FolderTreeRoleSent": true, "FolderTreeRoleDrafts": true, "FolderTreeRoleArchive": true, "FolderTreeRoleTrash": true, "FolderTreeRoleJunk": true, "FolderReferenceAbsent": true, "FolderReferencePresent": true, "FolderErrorInvalid": true, "FolderErrorTooLarge": true, "FolderErrorDuplicate": true, "FolderErrorNameConflict": true, "FolderErrorAccountMismatch": true, "FolderErrorNotFound": true, "FolderErrorParentUnavailable": true, "FolderErrorSelfParent": true, "FolderErrorCycle": true, "FolderErrorDepthExceeded": true, "FolderErrorSystemRole": true, "FolderErrorInUse": true, "FolderErrorLimitExceeded": true}
	allowedSentinels := map[string]string{"ErrFolderInvalid": "mail folder invalid", "ErrFolderTooLarge": "mail folder exceeds bounds", "ErrFolderDuplicate": "mail folder duplicate", "ErrFolderNameConflict": "mail folder name conflict", "ErrFolderAccountMismatch": "mail folder account mismatch", "ErrFolderNotFound": "mail folder not found", "ErrFolderParentUnavailable": "mail folder parent unavailable", "ErrFolderSelfParent": "mail folder self parent", "ErrFolderCycle": "mail folder cycle", "ErrFolderDepthExceeded": "mail folder depth exceeds bounds", "ErrFolderSystemRole": "mail folder system role protected", "ErrFolderInUse": "mail folder in use", "ErrFolderLimitExceeded": "mail folder limit exceeded"}
	verifyPublicDeclarations(t, file, allowedTypes, allowedFunctions, allowedMethods, allowedConstants, allowedSentinels, bindings["errors"])
	verifyPurityAST(t, file, allowedSentinels, bindings["github.com/google/uuid"])
}

func importPathBindings(file *ast.File) map[string]map[string]bool {
	bindings := make(map[string]map[string]bool)
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		binding := path[strings.LastIndex(path, "/")+1:]
		if imp.Name != nil {
			binding = imp.Name.Name
		}
		if bindings[path] == nil {
			bindings[path] = make(map[string]bool)
		}
		bindings[path][binding] = true
	}
	return bindings
}

func verifyPublicDeclarations(t *testing.T, file *ast.File, allowedTypes, allowedFunctions, allowedMethods, allowedConstants map[string]bool, allowedSentinels map[string]string, errorsBindings map[string]bool) {
	t.Helper()
	remainingTypes := copySet(allowedTypes)
	remainingFunctions := copySet(allowedFunctions)
	remainingMethods := copySet(allowedMethods)
	remainingConstants := copySet(allowedConstants)
	remainingSentinels := copyStrings(allowedSentinels)
	for _, declaration := range file.Decls {
		switch declared := declaration.(type) {
		case *ast.GenDecl:
			switch declared.Tok {
			case token.TYPE:
				for _, spec := range declared.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					if typeSpec.Name.IsExported() {
						require.True(t, remainingTypes[typeSpec.Name.Name])
						delete(remainingTypes, typeSpec.Name.Name)
					}
				}
			case token.CONST:
				for _, spec := range declared.Specs {
					valueSpec := spec.(*ast.ValueSpec)
					for _, name := range valueSpec.Names {
						if name.IsExported() {
							require.True(t, remainingConstants[name.Name])
							delete(remainingConstants, name.Name)
						}
					}
				}
			case token.VAR:
				for _, spec := range declared.Specs {
					valueSpec := spec.(*ast.ValueSpec)
					require.Len(t, valueSpec.Names, 1)
					require.Len(t, valueSpec.Values, 1)
					name := valueSpec.Names[0].Name
					want, ok := remainingSentinels[name]
					require.True(t, ok)
					call, ok := valueSpec.Values[0].(*ast.CallExpr)
					require.True(t, ok)
					selector, ok := call.Fun.(*ast.SelectorExpr)
					require.True(t, ok)
					packageName, ok := selector.X.(*ast.Ident)
					require.True(t, ok)
					require.True(t, errorsBindings[packageName.Name])
					require.Equal(t, "New", selector.Sel.Name)
					require.Len(t, call.Args, 1)
					literal, ok := call.Args[0].(*ast.BasicLit)
					require.True(t, ok)
					require.Equal(t, token.STRING, literal.Kind)
					require.Equal(t, "\""+want+"\"", literal.Value)
					delete(remainingSentinels, name)
				}
			}
		case *ast.FuncDecl:
			if !declared.Name.IsExported() {
				continue
			}
			if declared.Recv == nil {
				require.True(t, remainingFunctions[declared.Name.Name])
				delete(remainingFunctions, declared.Name.Name)
				continue
			}
			require.True(t, isFolderPolicyErrorReceiver(declared.Recv))
			require.True(t, remainingMethods[declared.Name.Name])
			verifyFolderPolicyMethodSignature(t, declared)
			delete(remainingMethods, declared.Name.Name)
		}
	}
	require.Empty(t, remainingTypes)
	require.Empty(t, remainingFunctions)
	require.Empty(t, remainingMethods)
	require.Empty(t, remainingConstants)
	require.Empty(t, remainingSentinels)
}

func verifyFolderPolicyMethodSignature(t *testing.T, function *ast.FuncDecl) {
	t.Helper()
	require.True(t, emptyFieldList(function.Type.Params))
	switch function.Name.Name {
	case "Error":
		require.True(t, singleIdentResult(function.Type.Results, "string"))
	case "Unwrap":
		require.True(t, singleIdentResult(function.Type.Results, "error"))
	default:
		t.Fatalf("unexpected FolderPolicyError method %q", function.Name.Name)
	}
}

func emptyFieldList(fields *ast.FieldList) bool {
	return fields == nil || len(fields.List) == 0
}

func singleIdentResult(fields *ast.FieldList, name string) bool {
	if fields == nil || len(fields.List) != 1 || len(fields.List[0].Names) > 1 {
		return false
	}
	identifier, ok := fields.List[0].Type.(*ast.Ident)
	return ok && identifier.Name == name
}

func verifyPurityAST(t *testing.T, file *ast.File, allowedSentinels map[string]string, uuidBindings map[string]bool) {
	t.Helper()
	functionTypes := functionTypeNames(file)
	packageVariables := packageVariableNames(file)
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			if function.Body == nil {
				t.Fatalf("folder tree source has declaration-only function %q", function.Name.Name)
			}
			if function.Recv == nil && function.Name.Name == "init" {
				t.Fatal("folder tree source must not declare init")
			}
		}
		if general, ok := declaration.(*ast.GenDecl); ok {
			switch general.Tok {
			case token.VAR:
				for _, spec := range general.Specs {
					valueSpec := spec.(*ast.ValueSpec)
					for _, name := range valueSpec.Names {
						if _, ok := allowedSentinels[name.Name]; !ok {
							t.Fatalf("folder tree source has non-sentinel package variable %q", name.Name)
						}
					}
				}
			case token.TYPE:
				for _, spec := range general.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					if functionTypes[typeSpec.Name.Name] {
						t.Fatalf("folder tree source must not declare function type %q", typeSpec.Name.Name)
					}
				}
			}
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.GoStmt:
			t.Fatal("folder tree source must not start goroutines")
		case *ast.AssignStmt:
			for _, target := range value.Lhs {
				if isSentinelTarget(target, allowedSentinels) {
					t.Fatal("folder tree source must not assign to an error sentinel")
				}
			}
		case *ast.IncDecStmt:
			if isSentinelTarget(value.X, allowedSentinels) {
				t.Fatal("folder tree source must not mutate an error sentinel")
			}
		case *ast.UnaryExpr:
			if value.Op == token.AND && isSentinelTarget(value.X, allowedSentinels) {
				t.Fatal("folder tree source must not take an error sentinel address")
			}
		case *ast.Ident:
			switch value.Name {
			case "context", "MailEngine", "Repository", "repository", "callback":
				t.Fatalf("folder tree source contains forbidden identifier %q", value.Name)
			}
			if packageVariables[value.Name] && value.Obj != nil && value.Obj.Kind == ast.Var {
				if _, ok := allowedSentinels[value.Name]; !ok {
					t.Fatalf("folder tree source uses non-sentinel package variable %q", value.Name)
				}
			}
		case *ast.Field:
			if containsFunctionType(value.Type, functionTypes) {
				t.Fatal("folder tree source must not declare callback fields or signatures")
			}
		case *ast.SelectorExpr:
			if isRandomUUIDSelector(value, uuidBindings) {
				t.Fatal("folder tree source must not generate UUIDs")
			}
		}
		return true
	})
}

func functionTypeNames(file *ast.File) map[string]bool {
	names := make(map[string]bool)
	changed := true
	for changed {
		changed = false
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				if !names[typeSpec.Name.Name] && containsFunctionType(typeSpec.Type, names) {
					names[typeSpec.Name.Name] = true
					changed = true
				}
			}
		}
	}
	return names
}

func packageVariableNames(file *ast.File) map[string]bool {
	names := make(map[string]bool)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				names[name.Name] = true
			}
		}
	}
	return names
}

func isRandomUUIDSelector(selector *ast.SelectorExpr, uuidBindings map[string]bool) bool {
	packageName, ok := selector.X.(*ast.Ident)
	if !ok || !uuidBindings[packageName.Name] {
		return false
	}
	name := selector.Sel.Name
	lower := strings.ToLower(name)
	return strings.HasPrefix(name, "New") || strings.Contains(lower, "random") || strings.Contains(lower, "rand")
}

func isSentinelTarget(value ast.Expr, allowedSentinels map[string]string) bool {
	switch target := value.(type) {
	case *ast.Ident:
		_, ok := allowedSentinels[target.Name]
		return ok
	case *ast.ParenExpr:
		return isSentinelTarget(target.X, allowedSentinels)
	case *ast.StarExpr:
		return isSentinelTarget(target.X, allowedSentinels)
	case *ast.UnaryExpr:
		return target.Op == token.AND && isSentinelTarget(target.X, allowedSentinels)
	default:
		return false
	}
}

func isFolderPolicyErrorReceiver(receiver *ast.FieldList) bool {
	if len(receiver.List) != 1 {
		return false
	}
	pointer, ok := receiver.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	identifier, ok := pointer.X.(*ast.Ident)
	return ok && identifier.Name == "FolderPolicyError"
}

func containsFunctionType(value ast.Expr, names map[string]bool) bool {
	switch typed := value.(type) {
	case *ast.FuncType:
		return true
	case *ast.Ident:
		return names[typed.Name]
	case *ast.ParenExpr:
		return containsFunctionType(typed.X, names)
	case *ast.StarExpr:
		return containsFunctionType(typed.X, names)
	case *ast.ArrayType:
		return containsFunctionType(typed.Elt, names)
	case *ast.MapType:
		return containsFunctionType(typed.Key, names) || containsFunctionType(typed.Value, names)
	case *ast.ChanType:
		return containsFunctionType(typed.Value, names)
	case *ast.Ellipsis:
		return containsFunctionType(typed.Elt, names)
	case *ast.StructType:
		return fieldListContainsFunctionType(typed.Fields, names)
	case *ast.InterfaceType:
		return fieldListContainsFunctionType(typed.Methods, names)
	default:
		return false
	}
}


func fieldListContainsFunctionType(fields *ast.FieldList, names map[string]bool) bool {
	if fields == nil {
		return false
	}
	for _, field := range fields.List {
		if containsFunctionType(field.Type, names) {
			return true
		}
	}
	return false
}

func copySet(input map[string]bool) map[string]bool {
	copy := make(map[string]bool, len(input))
	for value := range input {
		copy[value] = true
	}
	return copy
}

func copyStrings(input map[string]string) map[string]string {
	copy := make(map[string]string, len(input))
	for key, value := range input {
		copy[key] = value
	}
	return copy
}
