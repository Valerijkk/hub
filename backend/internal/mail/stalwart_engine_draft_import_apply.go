package mail

import (
	"context"
	"time"
)

// StalwartDraftMIMEImportApplyRequest identifies one persisted, already
// uploaded MIME blob and the exact initial-draft mutation that owns it.
// ReceivedAt must be the durable timestamp chosen before provider I/O.
type StalwartDraftMIMEImportApplyRequest struct {
	AccountID  string
	DraftKey   string
	MutationID string
	BlobID     string
	SizeBytes  int64
	ReceivedAt time.Time
}

// ApplyDraftMIMEImport imports one uploaded MIME blob and returns it only after
// an exact provider readback proves the resulting owner draft. It never retries
// Email/import; any post-mutation ambiguity requires external marker
// reconciliation.
func (engine *StalwartEngine) ApplyDraftMIMEImport(ctx context.Context, request StalwartDraftMIMEImportApplyRequest) (EngineMessage, error) {
	if !stalwartValidDraftMIMEImportApplyRequest(ctx, request) {
		return EngineMessage{}, stalwartEngineError(EngineErrorRejectedPolicy)
	}
	draftMarker, err := stalwartOpaqueMarker("draft", request.AccountID, request.DraftKey)
	if err != nil {
		return EngineMessage{}, err
	}
	mutationMarker, err := stalwartDraftMutationMarkerV2(request.AccountID, request.DraftKey, request.MutationID)
	if err != nil {
		return EngineMessage{}, err
	}
	emailState, err := engine.stalwartReadEmailState(ctx, request.AccountID)
	if err != nil {
		return EngineMessage{}, err
	}
	folders, _, err := engine.listFolderSnapshot(ctx, request.AccountID)
	if err != nil {
		return EngineMessage{}, err
	}
	draftsID, err := stalwartExactFolderRole(folders, EngineFolderRoleDrafts)
	if err != nil {
		return EngineMessage{}, err
	}
	imported, err := engine.ImportDraftMIME(ctx, StalwartDraftMIMEImportRequest{
		AccountID:       request.AccountID,
		DraftsMailboxID: draftsID,
		BlobID:          request.BlobID,
		SizeBytes:       request.SizeBytes,
		IfInState:       emailState,
		DraftKey:        request.DraftKey,
		MutationID:      request.MutationID,
		ReceivedAt:      request.ReceivedAt,
	})
	if err != nil {
		return EngineMessage{}, err
	}
	if imported.AccountID != request.AccountID ||
		imported.BlobID != request.BlobID ||
		imported.SizeBytes != request.SizeBytes ||
		!validOpaqueID(imported.MessageID) ||
		!validOpaqueID(imported.ThreadID) ||
		!validOpaqueID(imported.OldState) ||
		!validOpaqueID(imported.NewState) ||
		imported.OldState != emailState ||
		imported.NewState == imported.OldState {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	message, err := engine.stalwartReadDraftMessage(ctx, request.AccountID, imported.MessageID, request.SizeBytes)
	if err != nil {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	if !stalwartExactImportedDraftMessage(message, request, imported, draftsID, draftMarker, mutationMarker) {
		return EngineMessage{}, stalwartEngineError(EngineErrorIndeterminate)
	}
	return message, nil
}

func stalwartValidDraftMIMEImportApplyRequest(ctx context.Context, request StalwartDraftMIMEImportApplyRequest) bool {
	return ctx != nil &&
		validAccountID(request.AccountID) &&
		validSafeText(request.DraftKey, MaxEngineOpaqueIDBytes, true) &&
		validSafeText(request.MutationID, MaxEngineOpaqueIDBytes, true) &&
		validOpaqueID(request.BlobID) &&
		request.SizeBytes >= 1 &&
		request.SizeBytes <= MaxEngineMessageBytes &&
		stalwartValidImportReceivedAt(request.ReceivedAt)
}

func stalwartExactImportedDraftMessage(message EngineMessage, request StalwartDraftMIMEImportApplyRequest, imported StalwartDraftMIMEImportResult, draftsID, draftMarker, mutationMarker string) bool {
	projection := message.Projection
	return projection.AccountID == request.AccountID &&
		projection.MessageID == imported.MessageID &&
		validOpaqueID(projection.Version) &&
		projection.SizeBytes == request.SizeBytes &&
		projection.ReceivedAt != nil &&
		projection.ReceivedAt.Equal(request.ReceivedAt) &&
		projection.Origin == EngineProjectionOriginConfirmedOwnerDraft &&
		stalwartExactDraftOperation(message, request.AccountID, draftsID, draftMarker, mutationMarker)
}
