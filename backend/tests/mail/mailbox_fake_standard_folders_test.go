package mail_test

import (
	"context"
	"reflect"
	"testing"

	mail "github.com/valerii/hub/backend/internal/mail"
)

func TestFakeMailEnginePutFolderSeedsStandardFoldersDeterministically(t *testing.T) {
	ctx := context.Background()
	fake, accountID := newFakeFolderAccount(t)
	want := []mail.EngineFolder{
		{AccountID: accountID, FolderID: "folder-01-inbox", Name: "Inbox", Role: mail.EngineFolderRoleInbox},
		{AccountID: accountID, FolderID: "folder-02-sent", Name: "Sent", Role: mail.EngineFolderRoleSent},
		{AccountID: accountID, FolderID: "folder-03-drafts", Name: "Drafts", Role: mail.EngineFolderRoleDrafts},
		{AccountID: accountID, FolderID: "folder-04-archive", Name: "Archive", Role: mail.EngineFolderRoleArchive},
		{AccountID: accountID, FolderID: "folder-05-trash", Name: "Trash", Role: mail.EngineFolderRoleTrash},
		{AccountID: accountID, FolderID: "folder-06-junk", Name: "Junk", Role: mail.EngineFolderRoleJunk},
	}
	for _, folder := range want {
		if !fake.PutFolder(folder) {
			t.Fatalf("PutFolder(%#v) = false, want true", folder)
		}
	}
	for _, folder := range want {
		if !fake.PutFolder(folder) {
			t.Fatalf("PutFolder exact replay %#v = false, want true", folder)
		}
	}

	got, err := fake.ListFolders(ctx, accountID)
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListFolders = %#v, want sorted standard folders %#v", got, want)
	}

	conflict := want[0]
	conflict.Name = "Incoming"
	if fake.PutFolder(conflict) {
		t.Fatal("PutFolder accepted a conflicting replay")
	}
	if fake.PutFolder(mail.EngineFolder{AccountID: accountID, FolderID: "folder-07-inbox", Name: "Another Inbox", Role: mail.EngineFolderRoleInbox}) {
		t.Fatal("PutFolder accepted a duplicate standard role")
	}
}

func TestFakeMailEnginePutFolderRejectsInvalidParentAndCrossAccountState(t *testing.T) {
	ctx := context.Background()
	fake, accountID := newFakeFolderAccount(t)
	if fake.PutFolder(mail.EngineFolder{AccountID: accountID, FolderID: "550e8400-e29b-41d4-a716-446655440000", Name: "Inbox", Role: mail.EngineFolderRoleInbox}) {
		t.Fatal("PutFolder accepted a UUID-shaped non-opaque folder ID")
	}
	if fake.PutFolder(mail.EngineFolder{AccountID: accountID, FolderID: accountID, Name: "Colliding", Role: mail.EngineFolderRoleCustom}) {
		t.Fatal("PutFolder accepted a folder ID already issued to an account")
	}
	if !fake.PutFolder(mail.EngineFolder{AccountID: accountID, FolderID: "folder-parent", Name: "Projects", Role: mail.EngineFolderRoleCustom}) {
		t.Fatal("PutFolder parent custom folder = false, want true")
	}
	if fake.PutFolder(mail.EngineFolder{AccountID: accountID, FolderID: "folder-nested-inbox", ParentID: "folder-parent", Name: "Inbox", Role: mail.EngineFolderRoleInbox}) {
		t.Fatal("PutFolder accepted a standard role with a parent")
	}
	if fake.PutFolder(mail.EngineFolder{AccountID: accountID, FolderID: "folder-missing-parent", ParentID: "missing-parent", Name: "Child", Role: mail.EngineFolderRoleCustom}) {
		t.Fatal("PutFolder accepted a custom folder with a missing parent")
	}
	if fake.PutFolder(mail.EngineFolder{AccountID: accountID, FolderID: "folder-duplicate-name", Name: " projects ", Role: mail.EngineFolderRoleCustom}) {
		t.Fatal("PutFolder accepted a duplicate canonical custom sibling name")
	}
	if fake.PutFolder(mail.EngineFolder{AccountID: "unknown-account", FolderID: "folder-unowned", Name: "Inbox", Role: mail.EngineFolderRoleInbox}) {
		t.Fatal("PutFolder accepted a folder for an account not ensured by the fake")
	}

	other, err := fake.EnsureAccount(ctx, mail.AccountSpec{ProvisioningKey: "other-key", LocalPart: "other", Domain: "example.test", Enabled: true})
	if err != nil {
		t.Fatalf("EnsureAccount other: %v", err)
	}
	if fake.PutFolder(mail.EngineFolder{AccountID: other.AccountID, FolderID: "folder-cross-account", ParentID: "folder-parent", Name: "Cross account", Role: mail.EngineFolderRoleCustom}) {
		t.Fatal("PutFolder accepted a parent from another account")
	}
}

func newFakeFolderAccount(t *testing.T) (*mail.FakeMailEngine, string) {
	t.Helper()
	fake := mail.NewFakeMailEngine(mail.FakeMailEngineConfig{})
	if _, err := fake.EnsureDomain(context.Background(), mail.DomainSpec{ProvisioningKey: "domain-key", Domain: "example.test", Present: true, Enabled: true}); err != nil {
		t.Fatalf("EnsureDomain: %v", err)
	}
	account, err := fake.EnsureAccount(context.Background(), mail.AccountSpec{ProvisioningKey: "account-key", LocalPart: "ada", Domain: "example.test", Enabled: true})
	if err != nil {
		t.Fatalf("EnsureAccount: %v", err)
	}
	return fake, account.AccountID
}
