package calendar_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func meetingUpdatePersistenceSource(t *testing.T, file string) string {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	sourcePath := filepath.Clean(filepath.Join(filepath.Dir(testFile), "..", "..", "internal", "calendar", file))
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read Calendar source %q: %v", file, err)
	}
	return string(source)
}

func meetingUpdatePersistenceFunctionSlice(t *testing.T, source, start, nextDeclaration string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing function start %q", start)
	}
	nextIndex := strings.Index(source[startIndex+len(start):], nextDeclaration)
	if nextIndex < 0 {
		t.Fatalf("source is missing next declaration %q", nextDeclaration)
	}
	return source[startIndex : startIndex+len(start)+nextIndex]
}

func requireMeetingUpdatePersistenceSource(t *testing.T, source string, terms ...string) {
	t.Helper()

	for _, term := range terms {
		if !strings.Contains(source, term) {
			t.Fatalf("source is missing %q", term)
		}
	}
}

func TestNewMeetingUpdatePersistenceCommandAuthorizationSourceContract(t *testing.T) {
	command := meetingUpdatePersistenceFunctionSlice(
		t,
		meetingUpdatePersistenceSource(t, "meeting.go"),
		"func NewMeetingUpdatePersistenceCommand(",
		"func (command MeetingUpdatePersistenceCommand) valid() bool {",
	)
	if strings.Contains(command, "authorization.Current") {
		t.Fatal("persistence command must not compare authorization.Current before the Ent re-lock")
	}
	requireMeetingUpdatePersistenceSource(t, command,
		"current.EventVersion != expectedVersion",
		"authorization.Organizer.UserID != current.OwnerID",
		"authorization.ProjectID != canonical.input.ProjectID || authorization.ChannelID != canonical.input.ChannelID",
		"reservations.EventID() != current.EventID || reservations.OrganizationID() != current.OrganizationID || reservations.OrganizerID() != current.OwnerID",
	)
}

func TestPersistMeetingUpdateRelockSourceContract(t *testing.T) {
	persist := meetingUpdatePersistenceFunctionSlice(
		t,
		meetingUpdatePersistenceSource(t, "event_ent_transaction.go"),
		"func (backend *eventEntTransactionBackend) PersistMeetingUpdate(",
		"func (backend *eventEntTransactionBackend) PersistMeetingCancel(",
	)
	requireMeetingUpdatePersistenceSource(t, persist,
		"backend.lockMeetingMutationRows(ctx, scope, command.current.EventID)",
		"meetingMutationSnapshotsEqual(rows.snapshot, command.current)",
		"calendarevent.VersionEQ(expectedEventVersion)",
	)
}
