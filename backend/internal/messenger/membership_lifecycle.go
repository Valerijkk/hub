package messenger

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// MembershipGrantEntryPoint identifies the service path requesting membership.
// Every supported path uses the same authorization and baseline coordinator.
type MembershipGrantEntryPoint string

const (
	MembershipGrantEntryPointCreateChannelOwner         MembershipGrantEntryPoint = "create_channel_owner"
	MembershipGrantEntryPointCreateChannelInitialMember MembershipGrantEntryPoint = "create_channel_initial_member"
	MembershipGrantEntryPointJoinChannel                MembershipGrantEntryPoint = "join_channel"
	MembershipGrantEntryPointAddMembers                 MembershipGrantEntryPoint = "add_members"
	MembershipGrantEntryPointAddMembersInternal         MembershipGrantEntryPoint = "add_members_internal"
	MembershipGrantEntryPointAcceptInvite               MembershipGrantEntryPoint = "accept_invite"
	MembershipGrantEntryPointCreateOrGetDM              MembershipGrantEntryPoint = "create_or_get_dm"
)

type MembershipTargetKind string

const (
	MembershipTargetKindChannel MembershipTargetKind = "channel"
	MembershipTargetKindDM      MembershipTargetKind = "dm"
)

type MembershipGrantRequest struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	TargetID   uuid.UUID
	TargetKind MembershipTargetKind
	EntryPoint MembershipGrantEntryPoint
}

// MembershipGrantAuthorization carries the exact PostgreSQL-authorized scope.
// The coordinator must never reconstruct this scope from caller input.
type MembershipGrantAuthorization struct {
	AlreadyMember bool
	Scope         AuthorizedGrantBaselineScope
}

type AuthorizedGrantBaselineScope struct {
	TenantID            uuid.UUID
	ActorID             uuid.UUID
	TargetID            uuid.UUID
	ChannelID           uuid.UUID
	Kind                string
	TargetKind          MembershipTargetKind
	IsProjectChannel    bool
	LifecycleGeneration int64
}

type MembershipBaselineKind string

const (
	MembershipBaselineChannelOrDM MembershipBaselineKind = "channel_or_dm"
	MembershipBaselineMention     MembershipBaselineKind = "mention"
)

type MembershipGrantCommitResult struct {
	Created bool
}

type Phase151MembershipLifecycleRepository interface {
	AuthorizeMembershipGrant(context.Context, MembershipGrantRequest) (MembershipGrantAuthorization, error)
	ResetStaleMembershipState(context.Context, AuthorizedGrantBaselineScope) error
	CaptureAuthorizedMembershipHead(context.Context, AuthorizedGrantBaselineScope) (MessagePosition, error)
	UpsertMembershipBaseline(context.Context, AuthorizedGrantBaselineScope, MembershipBaselineKind, MessagePosition) error
	CommitMembershipGrant(context.Context, AuthorizedGrantBaselineScope) (MembershipGrantCommitResult, error)
	CompensateUncommittedMembershipGrant(context.Context, AuthorizedGrantBaselineScope) error
	PublishMembershipGranted(context.Context, AuthorizedGrantBaselineScope) error
	CommitMembershipRevocation(context.Context, MembershipRevocationRequest) (MembershipLifecycleFence, error)
	CleanupRevokedMembershipState(context.Context, MembershipLifecycleFence) error
	PublishMembershipRevoked(context.Context, MembershipLifecycleFence) error
}

var (
	errPhase151MembershipLifecycleRepositoryRequired = errors.New("phase151 membership lifecycle repository is required")
	errPhase151MembershipLifecycleContextRequired    = errors.New("phase151 membership lifecycle context is required")
	errPhase151InvalidMembershipGrantRequest         = errors.New("invalid phase151 membership grant request")
	errPhase151InvalidAuthorizedGrantScope           = errors.New("invalid phase151 authorized grant scope")
	errPhase151InvalidMembershipRevocationRequest    = errors.New("invalid phase151 membership revocation request")
	errPhase151InvalidMembershipLifecycleFence       = errors.New("invalid phase151 membership lifecycle fence")
)

func ExportPhase151ExecuteMembershipGrant(
	ctx context.Context,
	request MembershipGrantRequest,
	repository Phase151MembershipLifecycleRepository,
) (MembershipGrantCommitResult, error) {
	if err := validatePhase151MembershipGrantRequest(ctx, request, repository); err != nil {
		return MembershipGrantCommitResult{}, err
	}

	authorization, err := repository.AuthorizeMembershipGrant(ctx, request)
	if err != nil {
		return MembershipGrantCommitResult{}, err
	}
	if err := validatePhase151AuthorizedGrantScope(request, authorization.Scope); err != nil {
		return MembershipGrantCommitResult{}, err
	}
	if authorization.AlreadyMember {
		return MembershipGrantCommitResult{Created: false}, nil
	}

	scope := authorization.Scope
	compensate := func(operationErr error) (MembershipGrantCommitResult, error) {
		if compensationErr := repository.CompensateUncommittedMembershipGrant(ctx, scope); compensationErr != nil {
			return MembershipGrantCommitResult{}, errors.Join(operationErr, compensationErr)
		}
		return MembershipGrantCommitResult{}, operationErr
	}

	if err := repository.ResetStaleMembershipState(ctx, scope); err != nil {
		return compensate(err)
	}
	head, err := repository.CaptureAuthorizedMembershipHead(ctx, scope)
	if err != nil {
		return compensate(err)
	}
	if err := repository.UpsertMembershipBaseline(ctx, scope, MembershipBaselineChannelOrDM, head); err != nil {
		return compensate(err)
	}
	if err := repository.UpsertMembershipBaseline(ctx, scope, MembershipBaselineMention, head); err != nil {
		return compensate(err)
	}

	result, err := repository.CommitMembershipGrant(ctx, scope)
	if err != nil {
		return compensate(err)
	}
	if !result.Created {
		if compensationErr := repository.CompensateUncommittedMembershipGrant(ctx, scope); compensationErr != nil {
			return result, compensationErr
		}
		return result, nil
	}
	if err := repository.PublishMembershipGranted(ctx, scope); err != nil {
		return result, err
	}
	return result, nil
}

func ExportPhase151MessageIsAfterBaseline(baseline, candidate MessagePosition) bool {
	return messagePositionAfter(candidate, baseline)
}

type MembershipRevocationReason string

const (
	MembershipRevocationReasonLeave              MembershipRevocationReason = "leave"
	MembershipRevocationReasonProjectDeactivated MembershipRevocationReason = "project_deactivated"
	MembershipRevocationReasonUserDisabled       MembershipRevocationReason = "user_disabled"
)

type MembershipRevocationRequest struct {
	TenantID            uuid.UUID
	UserID              uuid.UUID
	TargetID            uuid.UUID
	TargetKind          MembershipTargetKind
	Reason              MembershipRevocationReason
	LifecycleGeneration int64
}

type MembershipRevocationResult struct {
	Fence         MembershipLifecycleFence
	CleanupFailed bool
	PublishFailed bool
}

type MembershipLifecycleFence struct {
	TenantID            uuid.UUID
	ActorID             uuid.UUID
	TargetID            uuid.UUID
	ChannelID           uuid.UUID
	TargetKind          MembershipTargetKind
	LifecycleGeneration int64
}

// ExportPhase151ExecuteMembershipRevoke commits the PG revocation first, then
// attempts cleanup and always publishes the content-free removal notification.
func ExportPhase151ExecuteMembershipRevoke(
	ctx context.Context,
	request MembershipRevocationRequest,
	repository Phase151MembershipLifecycleRepository,
) (MembershipRevocationResult, error) {
	if err := validatePhase151MembershipRevocationRequest(ctx, request, repository); err != nil {
		return MembershipRevocationResult{}, err
	}

	fence, err := repository.CommitMembershipRevocation(ctx, request)
	if err != nil {
		return MembershipRevocationResult{}, err
	}
	if err := validatePhase151MembershipLifecycleFence(request, fence); err != nil {
		return MembershipRevocationResult{Fence: fence}, err
	}

	cleanupErr := repository.CleanupRevokedMembershipState(ctx, fence)
	publishErr := repository.PublishMembershipRevoked(ctx, fence)
	result := MembershipRevocationResult{
		Fence:         fence,
		CleanupFailed: cleanupErr != nil,
		PublishFailed: publishErr != nil,
	}
	if publishErr != nil {
		return result, publishErr
	}
	return result, nil
}

func ExportPhase151MembershipCleanupFenceMatches(expected, actual MembershipLifecycleFence) bool {
	return expected.TenantID == actual.TenantID &&
		expected.ActorID == actual.ActorID &&
		expected.TargetID == actual.TargetID &&
		expected.ChannelID == actual.ChannelID &&
		expected.TargetKind == actual.TargetKind &&
		expected.LifecycleGeneration == actual.LifecycleGeneration
}

func validatePhase151MembershipGrantRequest(
	ctx context.Context,
	request MembershipGrantRequest,
	repository Phase151MembershipLifecycleRepository,
) error {
	if repository == nil {
		return errPhase151MembershipLifecycleRepositoryRequired
	}
	if ctx == nil {
		return errPhase151MembershipLifecycleContextRequired
	}
	if request.TenantID == uuid.Nil || request.UserID == uuid.Nil || request.TargetID == uuid.Nil {
		return errPhase151InvalidMembershipGrantRequest
	}
	if !isKnownMembershipTargetKind(request.TargetKind) || !isKnownMembershipGrantEntryPoint(request.EntryPoint) {
		return errPhase151InvalidMembershipGrantRequest
	}
	if request.EntryPoint == MembershipGrantEntryPointCreateOrGetDM {
		if request.TargetKind != MembershipTargetKindDM {
			return errPhase151InvalidMembershipGrantRequest
		}
		return nil
	}
	if request.TargetKind != MembershipTargetKindChannel {
		return errPhase151InvalidMembershipGrantRequest
	}
	return nil
}

func validatePhase151AuthorizedGrantScope(
	request MembershipGrantRequest,
	scope AuthorizedGrantBaselineScope,
) error {
	if scope.TenantID == uuid.Nil ||
		scope.ActorID == uuid.Nil ||
		scope.TargetID == uuid.Nil ||
		scope.ChannelID == uuid.Nil ||
		strings.TrimSpace(scope.Kind) == "" ||
		scope.LifecycleGeneration <= 0 ||
		!isKnownMembershipTargetKind(scope.TargetKind) {
		return errPhase151InvalidAuthorizedGrantScope
	}
	if scope.TenantID != request.TenantID ||
		scope.ActorID != request.UserID ||
		scope.TargetID != request.TargetID ||
		scope.TargetKind != request.TargetKind {
		return errPhase151InvalidAuthorizedGrantScope
	}
	return nil
}

func validatePhase151MembershipRevocationRequest(
	ctx context.Context,
	request MembershipRevocationRequest,
	repository Phase151MembershipLifecycleRepository,
) error {
	if repository == nil {
		return errPhase151MembershipLifecycleRepositoryRequired
	}
	if ctx == nil {
		return errPhase151MembershipLifecycleContextRequired
	}
	if request.TenantID == uuid.Nil || request.UserID == uuid.Nil || request.TargetID == uuid.Nil {
		return errPhase151InvalidMembershipRevocationRequest
	}
	if !isKnownMembershipTargetKind(request.TargetKind) || !isKnownMembershipRevocationReason(request.Reason) {
		return errPhase151InvalidMembershipRevocationRequest
	}
	return nil
}

func validatePhase151MembershipLifecycleFence(
	request MembershipRevocationRequest,
	fence MembershipLifecycleFence,
) error {
	if fence.TenantID == uuid.Nil ||
		fence.ActorID == uuid.Nil ||
		fence.TargetID == uuid.Nil ||
		fence.ChannelID == uuid.Nil ||
		fence.LifecycleGeneration <= 0 ||
		!isKnownMembershipTargetKind(fence.TargetKind) {
		return errPhase151InvalidMembershipLifecycleFence
	}
	if fence.TenantID != request.TenantID ||
		fence.ActorID != request.UserID ||
		fence.TargetID != request.TargetID ||
		fence.TargetKind != request.TargetKind {
		return errPhase151InvalidMembershipLifecycleFence
	}
	return nil
}

func isKnownMembershipGrantEntryPoint(entryPoint MembershipGrantEntryPoint) bool {
	switch entryPoint {
	case MembershipGrantEntryPointCreateChannelOwner,
		MembershipGrantEntryPointCreateChannelInitialMember,
		MembershipGrantEntryPointJoinChannel,
		MembershipGrantEntryPointAddMembers,
		MembershipGrantEntryPointAddMembersInternal,
		MembershipGrantEntryPointAcceptInvite,
		MembershipGrantEntryPointCreateOrGetDM:
		return true
	default:
		return false
	}
}

func isKnownMembershipTargetKind(targetKind MembershipTargetKind) bool {
	switch targetKind {
	case MembershipTargetKindChannel, MembershipTargetKindDM:
		return true
	default:
		return false
	}
}

func isKnownMembershipRevocationReason(reason MembershipRevocationReason) bool {
	switch reason {
	case MembershipRevocationReasonLeave,
		MembershipRevocationReasonProjectDeactivated,
		MembershipRevocationReasonUserDisabled:
		return true
	default:
		return false
	}
}
