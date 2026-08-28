package messenger

import "context"

type phase151PendingRecoveryTargetPager interface {
	ListPendingRecoveryTargets(
		context.Context,
		*Phase151RecoveryTargetCursor,
		int,
	) (Phase151PendingRecoveryTargetPage, error)
}

type phase151PendingRecoveryTargetReauthorizer interface {
	ReauthorizePendingRecoveryTarget(
		context.Context,
		Phase151PendingRecoveryTarget,
	) (Phase151RecoveryTargetAuthority, error)
}

type phase151PendingPostCreateDocumentLister interface {
	ListPendingPostCreateDocuments(
		context.Context,
		Phase151RecoveryTargetAuthority,
		int,
	) ([]MessageDocument, error)
}

// Phase151PostCreateRecoveryScannerAdapter composes independent recovery scan
// ports without adding validation or recovery lifecycle behavior.
type Phase151PostCreateRecoveryScannerAdapter struct {
	pager        phase151PendingRecoveryTargetPager
	reauthorizer phase151PendingRecoveryTargetReauthorizer
	documents    phase151PendingPostCreateDocumentLister
}

var _ Phase151PendingRecoveryScanRepository = (*Phase151PostCreateRecoveryScannerAdapter)(nil)

func NewPhase151PostCreateRecoveryScannerAdapter(
	pager phase151PendingRecoveryTargetPager,
	reauthorizer phase151PendingRecoveryTargetReauthorizer,
	documents phase151PendingPostCreateDocumentLister,
) *Phase151PostCreateRecoveryScannerAdapter {
	if phase151PostCreateDependencyNil(pager) || phase151PostCreateDependencyNil(reauthorizer) ||
		phase151PostCreateDependencyNil(documents) {
		return nil
	}
	return &Phase151PostCreateRecoveryScannerAdapter{
		pager: pager, reauthorizer: reauthorizer, documents: documents,
	}
}

func (a *Phase151PostCreateRecoveryScannerAdapter) ListPendingRecoveryTargets(
	ctx context.Context,
	after *Phase151RecoveryTargetCursor,
	limit int,
) (Phase151PendingRecoveryTargetPage, error) {
	if err := a.validate(ctx); err != nil {
		return Phase151PendingRecoveryTargetPage{}, err
	}
	return a.pager.ListPendingRecoveryTargets(ctx, after, limit)
}

func (a *Phase151PostCreateRecoveryScannerAdapter) ReauthorizePendingRecoveryTarget(
	ctx context.Context,
	target Phase151PendingRecoveryTarget,
) (Phase151RecoveryTargetAuthority, error) {
	if err := a.validate(ctx); err != nil {
		return Phase151RecoveryTargetAuthority{}, err
	}
	return a.reauthorizer.ReauthorizePendingRecoveryTarget(ctx, target)
}

func (a *Phase151PostCreateRecoveryScannerAdapter) ListPendingPostCreateDocuments(
	ctx context.Context,
	authority Phase151RecoveryTargetAuthority,
	limit int,
) ([]MessageDocument, error) {
	if err := a.validate(ctx); err != nil {
		return nil, err
	}
	return a.documents.ListPendingPostCreateDocuments(ctx, authority, limit)
}

func (a *Phase151PostCreateRecoveryScannerAdapter) validate(ctx context.Context) error {
	if a == nil || phase151PostCreateDependencyNil(ctx) ||
		phase151PostCreateDependencyNil(a.pager) ||
		phase151PostCreateDependencyNil(a.reauthorizer) ||
		phase151PostCreateDependencyNil(a.documents) {
		return ErrPostCreatePipelineInvalid
	}
	return ctx.Err()
}
