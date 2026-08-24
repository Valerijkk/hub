package mail

import (
	"context"
	"reflect"

	"github.com/google/uuid"
)

// mailboxDraftAttachmentMaterializer converts one already-fenced durable draft
// attachment graph into the sealed source capability consumed by the MIME
// writer. It exposes neither object locators nor attachment bytes.
type mailboxDraftAttachmentMaterializer interface {
	MaterializeMailboxDraftAttachments(context.Context, mailboxDraftAttachmentMaterializationRequest) (mailboxDraftAttachmentMaterializationResult, error)
}

type mailboxDraftAttachmentMaterializationRequest struct {
	ActorID         uuid.UUID
	OrganizationID  uuid.UUID
	AccountID       uuid.UUID
	EngineAccountID string
	Fence           MailboxEngineCommandFence
	DraftID         uuid.UUID
	DraftClientID   uuid.UUID
	Revision        int64
	MutationID      uuid.UUID
	WorkKind        MailboxDraftWorkKind
	EffectiveState  MailboxDraftEffectiveState
}

// mailboxDraftAttachmentMaterializationResult retains the package-private
// stream capability only. A nil capability is the canonical attachmentless
// result.
type mailboxDraftAttachmentMaterializationResult struct {
	capability *mailboxDraftAttachmentCapability
}

func mailboxDraftAttachmentMaterializationRequestValid(request mailboxDraftAttachmentMaterializationRequest) bool {
	if request.ActorID == uuid.Nil || request.OrganizationID == uuid.Nil || request.AccountID == uuid.Nil || request.DraftID == uuid.Nil || request.DraftClientID == uuid.Nil || request.MutationID == uuid.Nil || request.Revision < 1 || request.Fence.AccountID != request.AccountID || !mailboxEntEngineCommandFenceValid(request.Fence) || !syncValidOpaque(request.EngineAccountID, MaxEngineOpaqueIDBytes) || !mailboxDraftAttachmentMaterializationEffectiveStateValid(request.EffectiveState) {
		return false
	}
	switch request.WorkKind {
	case MailboxDraftWorkCreate, MailboxDraftWorkSave:
		return true
	default:
		return false
	}
}

func mailboxDraftAttachmentMaterializationEffectiveStateValid(state MailboxDraftEffectiveState) bool {
	if state.FromIdentityID == uuid.Nil || !mailboxEntDraftReadUUIDPointerValid(state.ReplyToMessageID) || !mailboxEntDraftReadUUIDPointerValid(state.ForwardMessageID) || len(state.AttachmentIDs) > MaxEngineAttachmentItems {
		return false
	}
	return mailboxEntDraftReadUUIDSliceValid(state.AttachmentIDs)
}

func mailboxDraftAttachmentMaterializerDependencyNil(dependency any) bool {
	if dependency == nil {
		return true
	}

	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
