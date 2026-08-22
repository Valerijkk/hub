package mail_test

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/mail"
)

var _ mail.MailboxOwnerRepository = (*mail.MailboxEntRepository)(nil)

func TestMailboxEntRepositoryConstructorRejectsNilDependencies(t *testing.T) {
	clock := mailboxEntRepositoryClock{now: time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)}

	repository, err := mail.NewMailboxEntRepository(nil, clock)
	if repository != nil {
		t.Fatal("nil client repository must be nil")
	}
	if !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fatalf("nil client error = %v, want invalid", err)
	}

	repository, err = mail.NewMailboxEntRepository(ent.NewClient(), nil)
	if repository != nil || !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fail()
	}

	var typedNilClock *mailboxEntRepositoryClock
	repository, err = mail.NewMailboxEntRepository(ent.NewClient(), typedNilClock)
	if repository != nil {
		t.Fatal("typed-nil clock repository must be nil")
	}
	if !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fatalf("typed-nil clock error = %v, want invalid", err)
	}
}

func TestMailboxEntRepositoryHasFrozenPublicContract(t *testing.T) {
	constructorType := reflect.TypeOf(mail.NewMailboxEntRepository)
	wantConstructorType := reflect.TypeOf((func(*ent.Client, mail.MailboxRepositoryClock) (*mail.MailboxEntRepository, error))(nil))
	if constructorType != wantConstructorType {
		t.Fatalf("NewMailboxEntRepository type = %v, want %v", constructorType, wantConstructorType)
	}

	clockType := reflect.TypeOf((*mail.MailboxRepositoryClock)(nil)).Elem()
	method, exists := clockType.MethodByName("Now")
	if !exists {
		t.Fatal("MailboxRepositoryClock.Now is missing")
	}
	wantMethodType := reflect.TypeOf((func() time.Time)(nil))
	if method.Type != wantMethodType {
		t.Fatalf("MailboxRepositoryClock.Now type = %v, want %v", method.Type, wantMethodType)
	}
	if clockType.NumMethod() != 1 {
		t.Fatalf("MailboxRepositoryClock method count = %d, want 1", clockType.NumMethod())
	}
}

func TestMailboxEntRepositoryOwnerResolutionStaysLockedAndProviderFree(t *testing.T) {
	source, err := os.ReadFile("../../internal/mail/mailbox_ent_owner.go")
	if err != nil {
		t.Fatalf("read owner repository source: %v", err)
	}

	text := string(source)
	requireOrderedSourceFragments(t, text,
		"Organization.Query()",
		"User.Query()",
		"MailAccount.Query()",
		"MailIdentity.Query()",
	)
	if got := strings.Count(text, "ForUpdate()"); got != 4 {
		t.Fatalf("ForUpdate count = %d, want 4", got)
	}
	for _, forbidden := range []string{"MailEngine", "stalwart", "http.", "https://", "net/http"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("owner repository must not perform provider or network I/O: found %q", forbidden)
		}
	}
}

func TestMailboxEntRepositoryTimeContractRequiresExactSecondUTC(t *testing.T) {
	source, err := os.ReadFile("../../internal/mail/mailbox_ent_repository.go")
	if err != nil {
		t.Fatalf("read repository source: %v", err)
	}

	text := string(source)
	for _, required := range []string{"value.IsZero()", "value.Location() != time.UTC", "value.Nanosecond() != 0", "clock.Now()"} {
		if !strings.Contains(text, required) {
			t.Fatalf("repository time validation is missing %q", required)
		}
	}
}

func requireOrderedSourceFragments(t *testing.T, source string, fragments ...string) {
	t.Helper()
	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position == -1 {
			t.Fatalf("source is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("source lock order puts %q before its predecessor", fragment)
		}
		previous = position
	}
}

type mailboxEntRepositoryClock struct {
	now time.Time
}

func (clock mailboxEntRepositoryClock) Now() time.Time {
	return clock.now
}
