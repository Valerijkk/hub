package messenger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UnreadCategory string

const (
	UnreadCategoryChannel        UnreadCategory = "channel"
	UnreadCategoryDM             UnreadCategory = "dm"
	UnreadCategoryMention        UnreadCategory = "mention"
	UnreadCategoryFollowedThread UnreadCategory = "followed_thread"
)

var unreadCategories = [...]UnreadCategory{
	UnreadCategoryChannel,
	UnreadCategoryDM,
	UnreadCategoryMention,
	UnreadCategoryFollowedThread,
}

type AuthorizedUnreadTarget struct {
	ChannelID uuid.UUID
	Kind      string
}

type AuthorizedUnreadScope struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Targets  []AuthorizedUnreadTarget
}

type AuthorizedUnreadScopeResolver interface {
	ResolveAuthorizedUnreadScope(context.Context, uuid.UUID) (AuthorizedUnreadScope, error)
}

type UnreadCandidate struct {
	TargetID  string
	MessageID string
	SenderID  uuid.UUID
}

type Phase151UnreadCoordinatorRepository interface {
	ListUnreadCandidates(context.Context, AuthorizedUnreadScope, UnreadCategory) ([]UnreadCandidate, error)
}

type UnreadResponse struct {
	Channels  map[string]int64 `json:"channels"`
	DMs       map[string]int64 `json:"dms"`
	Mentions  map[string]int64 `json:"mentions"`
	Threads   map[string]int64 `json:"threads"`
	AnyUnread bool             `json:"any_unread"`
}

func ExportPhase151BuildUnreadResponse(
	ctx context.Context,
	scope AuthorizedUnreadScope,
	repository Phase151UnreadCoordinatorRepository,
) (UnreadResponse, error) {
	response := newPhase151UnreadResponse()
	if repository == nil || scope.TenantID == uuid.Nil || scope.UserID == uuid.Nil {
		return response, errors.New("authorized unread scope is required")
	}

	for _, category := range unreadCategories {
		candidates, err := repository.ListUnreadCandidates(ctx, scope, category)
		if err != nil {
			return response, fmt.Errorf("list %s unread candidates: %w", category, err)
		}
		counts := unreadCategoryMap(&response, category)
		for _, candidate := range candidates {
			if candidate.TargetID == "" || candidate.MessageID == "" || candidate.SenderID == scope.UserID {
				continue
			}
			counts[candidate.TargetID]++
			response.AnyUnread = true
		}
	}
	return response, nil
}

func newPhase151UnreadResponse() UnreadResponse {
	return UnreadResponse{
		Channels: make(map[string]int64),
		DMs:      make(map[string]int64),
		Mentions: make(map[string]int64),
		Threads:  make(map[string]int64),
	}
}

func unreadCategoryMap(response *UnreadResponse, category UnreadCategory) map[string]int64 {
	switch category {
	case UnreadCategoryChannel:
		return response.Channels
	case UnreadCategoryDM:
		return response.DMs
	case UnreadCategoryMention:
		return response.Mentions
	case UnreadCategoryFollowedThread:
		return response.Threads
	default:
		return nil
	}
}

type Phase151PendingRecoveryTarget struct {
	TenantID  uuid.UUID
	ChannelID uuid.UUID
	Kind      string
}

type Phase151RecoveryTargetCursor struct {
	TenantID  uuid.UUID
	ChannelID uuid.UUID
	Kind      string
}

type Phase151PendingRecoveryTargetPage struct {
	Targets []Phase151PendingRecoveryTarget
	Next    *Phase151RecoveryTargetCursor
}

type Phase151PendingRecoveryConfig struct {
	MaxPages         int
	MaxTargets       int
	PendingScanLimit int
	MinimumInterval  time.Duration
}

type Phase151PendingRecoveryRepository interface {
	TryAcquirePendingRecoveryInterval(context.Context, time.Duration) (bool, error)
	ListPendingRecoveryTargets(context.Context, *Phase151RecoveryTargetCursor, int) (Phase151PendingRecoveryTargetPage, error)
	ReauthorizePendingRecoveryTarget(context.Context, Phase151PendingRecoveryTarget) (AuthorizedMessageProducerScope, error)
	ListPendingPostCreateDocuments(context.Context, AuthorizedMessageProducerScope, Phase151PendingRecoveryTarget, int) ([]MessageDocument, error)
	ReconcilePendingPostCreateEffects(context.Context, AuthorizedMessageProducerScope, MessageDocument) error
}

func ExportPhase151ExecutePendingPostCreateRecovery(
	ctx context.Context,
	repository Phase151PendingRecoveryRepository,
	config Phase151PendingRecoveryConfig,
) error {
	if repository == nil || config.MaxPages <= 0 || config.MaxTargets <= 0 ||
		config.PendingScanLimit <= 0 || config.MinimumInterval <= 0 {
		return errors.New("valid pending recovery configuration is required")
	}
	acquired, err := repository.TryAcquirePendingRecoveryInterval(ctx, config.MinimumInterval)
	if err != nil {
		return fmt.Errorf("acquire pending recovery interval: %w", err)
	}
	if !acquired {
		return nil
	}

	seen := make(map[Phase151PendingRecoveryTarget]struct{})
	var after *Phase151RecoveryTargetCursor
	for pageNumber := 0; pageNumber < config.MaxPages && len(seen) < config.MaxTargets; pageNumber++ {
		page, err := repository.ListPendingRecoveryTargets(ctx, after, config.MaxTargets-len(seen))
		if err != nil {
			return fmt.Errorf("list pending recovery targets: %w", err)
		}
		for _, target := range page.Targets {
			if len(seen) >= config.MaxTargets {
				break
			}
			if target.TenantID == uuid.Nil || target.ChannelID == uuid.Nil || target.Kind == "" {
				continue
			}
			if _, duplicate := seen[target]; duplicate {
				continue
			}
			seen[target] = struct{}{}

			scope, err := repository.ReauthorizePendingRecoveryTarget(ctx, target)
			if err != nil {
				continue
			}
			if scope.TenantID != target.TenantID || scope.ChannelID != target.ChannelID ||
				scope.ActorID == uuid.Nil {
				continue
			}
			documents, err := repository.ListPendingPostCreateDocuments(ctx, scope, target, config.PendingScanLimit)
			if err != nil {
				return fmt.Errorf("list pending post-create documents: %w", err)
			}
			seenDocuments := make(map[string]struct{}, config.PendingScanLimit)
			for documentIndex, document := range documents {
				if documentIndex >= config.PendingScanLimit {
					break
				}
				if document.ID == "" || document.PostCreateEffectsState != PostCreateEffectsStatePending {
					continue
				}
				if _, duplicate := seenDocuments[document.ID]; duplicate {
					continue
				}
				seenDocuments[document.ID] = struct{}{}
				if err := repository.ReconcilePendingPostCreateEffects(ctx, scope, document); err != nil {
					return fmt.Errorf("reconcile pending post-create effects: %w", err)
				}
			}
		}
		if page.Next == nil {
			break
		}
		next := *page.Next
		after = &next
	}
	return nil
}

type Phase151UnreadCutoverCursor struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
}

type Phase151AuthorizedUnreadScopePage struct {
	Scopes []AuthorizedUnreadScope
	Next   *Phase151UnreadCutoverCursor
}

type Phase151UnreadCutoverConfig struct {
	MaxPages  int
	MaxScopes int
}

type Phase151UnreadCutoverScopePager interface {
	ListFreshAuthorizedUnreadScopes(context.Context, *Phase151UnreadCutoverCursor, int) (Phase151AuthorizedUnreadScopePage, error)
}

type Phase151UnreadCutoverRepository interface {
	CurrentAuthorizedTargetHead(context.Context, AuthorizedUnreadScope, AuthorizedUnreadTarget, UnreadCategory) (MessagePosition, error)
	ListEligiblePersistedFollowedRoots(context.Context, AuthorizedUnreadScope, AuthorizedUnreadTarget) ([]string, error)
	CurrentAuthorizedFollowedRootHead(context.Context, AuthorizedUnreadScope, AuthorizedUnreadTarget, string) (MessagePosition, error)
	PersistUnreadCutoverBaseline(context.Context, AuthorizedUnreadScope, UnreadCategory, string, MessagePosition) error
}

func ExportPhase151ExecuteUnreadCutover(
	ctx context.Context,
	scopes Phase151UnreadCutoverScopePager,
	repository Phase151UnreadCutoverRepository,
	config Phase151UnreadCutoverConfig,
) error {
	if scopes == nil || repository == nil || config.MaxPages <= 0 || config.MaxScopes <= 0 {
		return errors.New("valid unread cutover configuration is required")
	}

	processed := 0
	var after *Phase151UnreadCutoverCursor
	for pageNumber := 0; pageNumber < config.MaxPages && processed < config.MaxScopes; pageNumber++ {
		page, err := scopes.ListFreshAuthorizedUnreadScopes(ctx, after, config.MaxScopes-processed)
		if err != nil {
			return fmt.Errorf("list authorized unread scopes: %w", err)
		}
		for _, scope := range page.Scopes {
			if processed >= config.MaxScopes {
				break
			}
			processed++
			if scope.TenantID == uuid.Nil || scope.UserID == uuid.Nil || len(scope.Targets) == 0 {
				continue
			}
			for _, target := range scope.Targets {
				if target.ChannelID == uuid.Nil {
					continue
				}
				if err := baselineAuthorizedUnreadTarget(ctx, repository, scope, target); err != nil {
					return err
				}
			}
		}
		if page.Next == nil {
			break
		}
		next := *page.Next
		after = &next
	}
	return nil
}

func baselineAuthorizedUnreadTarget(
	ctx context.Context,
	repository Phase151UnreadCutoverRepository,
	scope AuthorizedUnreadScope,
	target AuthorizedUnreadTarget,
) error {
	switch target.Kind {
	case "channel":
		if err := baselineAuthorizedTargetCategory(ctx, repository, scope, target, UnreadCategoryChannel, target.ChannelID.String()); err != nil {
			return err
		}
		if err := baselineAuthorizedTargetCategory(ctx, repository, scope, target, UnreadCategoryMention, target.ChannelID.String()); err != nil {
			return err
		}
		roots, err := repository.ListEligiblePersistedFollowedRoots(ctx, scope, target)
		if err != nil {
			return fmt.Errorf("list eligible followed roots: %w", err)
		}
		for _, rootMessageID := range roots {
			if rootMessageID == "" {
				continue
			}
			head, err := repository.CurrentAuthorizedFollowedRootHead(ctx, scope, target, rootMessageID)
			if err != nil {
				return fmt.Errorf("resolve followed root head: %w", err)
			}
			if err := repository.PersistUnreadCutoverBaseline(ctx, scope, UnreadCategoryFollowedThread, rootMessageID, head); err != nil {
				return fmt.Errorf("persist followed root baseline: %w", err)
			}
		}
	case "dm":
		return baselineAuthorizedTargetCategory(ctx, repository, scope, target, UnreadCategoryDM, target.ChannelID.String())
	default:
		return nil
	}
	return nil
}

func baselineAuthorizedTargetCategory(
	ctx context.Context,
	repository Phase151UnreadCutoverRepository,
	scope AuthorizedUnreadScope,
	target AuthorizedUnreadTarget,
	category UnreadCategory,
	targetID string,
) error {
	head, err := repository.CurrentAuthorizedTargetHead(ctx, scope, target, category)
	if err != nil {
		return fmt.Errorf("resolve %s head: %w", category, err)
	}
	if err := repository.PersistUnreadCutoverBaseline(ctx, scope, category, targetID, head); err != nil {
		return fmt.Errorf("persist %s baseline: %w", category, err)
	}
	return nil
}

type DashboardUnreadSummary struct {
	UniqueMessageCount int64 `json:"unique_message_count"`
	HasUnread          bool  `json:"has_unread"`
}

type Phase151DashboardUnreadRepository interface {
	ListAuthorizedUnreadMessageIDs(context.Context, AuthorizedUnreadScope, UnreadCategory) ([]string, error)
}

func ExportPhase151BuildDashboardUnreadSummary(
	ctx context.Context,
	userID uuid.UUID,
	scopeResolver AuthorizedUnreadScopeResolver,
	repository Phase151DashboardUnreadRepository,
) (DashboardUnreadSummary, error) {
	if userID == uuid.Nil || scopeResolver == nil || repository == nil {
		return DashboardUnreadSummary{}, errors.New("authorized dashboard unread dependencies are required")
	}
	scope, err := scopeResolver.ResolveAuthorizedUnreadScope(ctx, userID)
	if err != nil {
		return DashboardUnreadSummary{}, fmt.Errorf("resolve authorized unread scope: %w", err)
	}
	if scope.TenantID == uuid.Nil || scope.UserID != userID {
		return DashboardUnreadSummary{}, errors.New("authorized unread scope denied")
	}

	unique := make(map[string]struct{})
	for _, category := range unreadCategories {
		messageIDs, err := repository.ListAuthorizedUnreadMessageIDs(ctx, scope, category)
		if err != nil {
			return DashboardUnreadSummary{}, fmt.Errorf("list %s dashboard unread messages: %w", category, err)
		}
		for _, messageID := range messageIDs {
			if messageID != "" {
				unique[messageID] = struct{}{}
			}
	}
	}
	count := int64(len(unique))
	return DashboardUnreadSummary{UniqueMessageCount: count, HasUnread: count > 0}, nil
}

type Phase151UnreadDependencyContract struct {
	PresenceRedisAllowed            bool
	UnreadRedisAccepted             bool
	DashboardUnreadRedisAccepted    bool
	DurableUnreadRepositoryRequired bool
	AuthorizedUnreadScopeRequired   bool
}

func ExportPhase151UnreadDependencyContract() Phase151UnreadDependencyContract {
	return Phase151UnreadDependencyContract{
		PresenceRedisAllowed:            true,
		DurableUnreadRepositoryRequired: true,
		AuthorizedUnreadScopeRequired:   true,
	}
}

type Phase151UnreadCutoverResult struct {
	BaselinedAuthorizedTargets bool
	ImportedRedisCounters      bool
}

// ExportPhase151RunUnreadCutover is the service-level cutover seam. The
// detailed paged cutover is executed through ExportPhase151ExecuteUnreadCutover;
// this entry point intentionally never consults the legacy Redis dependency.
func ExportPhase151RunUnreadCutover(
	_ context.Context,
	service *MessengerService,
	userID uuid.UUID,
) (Phase151UnreadCutoverResult, error) {
	if service == nil || userID == uuid.Nil {
		return Phase151UnreadCutoverResult{}, errors.New("messenger service and user are required")
	}
	return Phase151UnreadCutoverResult{}, errors.New("service-level unread cutover is not bound; use the authorized paged cutover")
}

type Phase151DurableUnreadRuntime interface {
	GetDurableUnreadCounts(context.Context, uuid.UUID) (map[string]int64, error)
	MarkDurableChannelRead(context.Context, uuid.UUID, uuid.UUID) error
}
