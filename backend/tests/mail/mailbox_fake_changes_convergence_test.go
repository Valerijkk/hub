package mail_test

import (
	"context"
	"testing"

	mail "github.com/valerii/hub/backend/internal/mail"
)

func TestFakeMailEngineNaturalChangesCollapseLatestMessageMutationAndAdvanceRawCursor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFake("message-a", "version-a")

	draft, err := fake.SaveDraft(ctx, testDraft("account-a", "draft-a", "", "", "mutation-a"))
	if err != nil {
		t.Fatalf("SaveDraft: %v", err)
	}
	if err := fake.SetFlags(ctx, "account-a", []string{draft.Projection.MessageID}, mail.FlagPatch{Seen: mail.FlagValue{Present: true, Value: true}}); err != nil {
		t.Fatalf("SetFlags: %v", err)
	}
	page, err := fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "sync:0", Page: mail.EnginePageRequest{Limit: 2}})
	if err != nil {
		t.Fatalf("Changes: %v", err)
	}
	if page.NextCursor != "sync:2" || page.Bounds != (mail.EnginePageBounds{Requested: 2, Maximum: mail.MaxEnginePageItems, Returned: 1}) {
		t.Fatalf("Changes page cursor/bounds = %#v, want one logical row after two consumed raw mutations", page)
	}
	if page.CreatedOrUpdated == nil || page.DestroyedIDs == nil || len(page.CreatedOrUpdated) != 1 || len(page.DestroyedIDs) != 0 {
		t.Fatalf("Changes page rows = %#v, want concrete slices with one collapsed projection", page)
	}
	if got := page.CreatedOrUpdated[0]; got.MessageID != draft.Projection.MessageID || !got.Flags.Seen {
		t.Fatalf("Changes latest projection = %#v, want the flagged latest revision of %q", got, draft.Projection.MessageID)
	}
	terminal, err := fake.Changes(ctx, mail.EngineChangeRequest{AccountID: "account-a", FromCursor: "sync:2", Page: mail.EnginePageRequest{Limit: 2}})
	if err != nil {
		t.Fatalf("Changes after terminal raw cursor: %v", err)
	}
	if terminal.NextCursor != "" || terminal.Bounds != (mail.EnginePageBounds{Requested: 2, Maximum: mail.MaxEnginePageItems, Returned: 0}) {
		t.Fatalf("Changes after terminal raw cursor = %#v, want terminal empty page", terminal)
	}
	if terminal.CreatedOrUpdated == nil || terminal.DestroyedIDs == nil || len(terminal.CreatedOrUpdated) != 0 || len(terminal.DestroyedIDs) != 0 {
		t.Fatalf("Changes after terminal raw cursor rows = %#v, want concrete empty slices", terminal)
	}
}
