package messenger

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

type Phase151MentionAudience interface {
	ResolveChannelEventAudience(context.Context, uuid.UUID, string) ([]uuid.UUID, error)
}

type Phase151MentionPresence interface {
	GetOnlineUserIDs(context.Context) ([]string, error)
}

type Phase151MentionRoleSource interface {
	ListProjectMembersByRole(context.Context, uuid.UUID, string) ([]uuid.UUID, error)
}

// Phase151EligibleMentionResolver freezes the effective recipient set before
// Mongo persistence. Every candidate comes from current PostgreSQL-backed
// channel audience and is never discovered from message storage.
type Phase151EligibleMentionResolver struct {
	authority ChannelAccessResolver
	audience  Phase151MentionAudience
	presence  Phase151MentionPresence
	roles     Phase151MentionRoleSource
}

func NewPhase151EligibleMentionResolver(
	authority ChannelAccessResolver,
	audience Phase151MentionAudience,
	presence Phase151MentionPresence,
	roles Phase151MentionRoleSource,
) *Phase151EligibleMentionResolver {
	if authority == nil || audience == nil || presence == nil || roles == nil {
		panic("post-create mention resolver dependencies are required")
	}
	return &Phase151EligibleMentionResolver{
		authority: authority,
		audience:  audience,
		presence:  presence,
		roles:     roles,
	}
}

func (r *Phase151EligibleMentionResolver) ResolveEligibleMentionUserIDs(
	ctx context.Context,
	scope AuthorizedMessageProducerScope,
	request PostCreatePipelineRequest,
) ([]uuid.UUID, error) {
	if ctx == nil || ctx.Err() != nil || r == nil || r.authority == nil ||
		r.audience == nil || r.presence == nil || r.roles == nil {
		return nil, ErrPostCreatePipelineInvalid
	}
	if err := validatePostCreatePipelineRequest(scope, request); err != nil {
		return nil, err
	}
	if len(request.Message.Mentions) == 0 && len(request.Message.MentionGroups) == 0 {
		return []uuid.UUID{}, nil
	}

	access, err := r.authority.AuthorizeChannelAccess(
		ctx,
		scope.ActorID,
		scope.ChannelID,
		ChannelActionReceiveEvent,
	)
	if err != nil {
		return nil, fmt.Errorf("authorize mention resolution: %w", err)
	}
	if access.TenantID != scope.TenantID || access.Channel == nil || access.Channel.ID != scope.ChannelID {
		return nil, ErrPostCreatePipelineInvalid
	}

	recipients, err := r.audience.ResolveChannelEventAudience(
		ctx,
		scope.ChannelID,
		UserNewMessageNotificationEvent,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve mention audience: %w", err)
	}
	eligible := make(map[uuid.UUID]struct{}, len(recipients))
	for _, recipientID := range recipients {
		if recipientID != uuid.Nil {
			eligible[recipientID] = struct{}{}
		}
	}
	if _, senderEligible := eligible[scope.ActorID]; !senderEligible {
		return nil, ErrPostCreatePipelineInvalid
	}

	resolved := make(map[uuid.UUID]struct{})
	addEligible := func(candidateID uuid.UUID) {
		if candidateID == uuid.Nil || candidateID == scope.ActorID {
			return
		}
		if _, ok := eligible[candidateID]; ok {
			resolved[candidateID] = struct{}{}
		}
	}
	for _, rawUserID := range request.Message.Mentions {
		userID, parseErr := uuid.Parse(strings.TrimSpace(rawUserID))
		if parseErr == nil {
			addEligible(userID)
		}
	}

	seenGroups := make(map[string]struct{}, len(request.Message.MentionGroups))
	for _, rawGroup := range request.Message.MentionGroups {
		group := strings.ToLower(strings.TrimSpace(rawGroup))
		if _, duplicate := seenGroups[group]; duplicate {
			continue
		}
		seenGroups[group] = struct{}{}

		switch group {
		case "all":
			for candidateID := range eligible {
				addEligible(candidateID)
			}
		case "online":
			onlineUserIDs, presenceErr := r.presence.GetOnlineUserIDs(ctx)
			if presenceErr != nil {
				return nil, fmt.Errorf("resolve online mentions: %w", presenceErr)
			}
			for _, rawUserID := range onlineUserIDs {
				userID, parseErr := uuid.Parse(strings.TrimSpace(rawUserID))
				if parseErr == nil {
					addEligible(userID)
				}
			}
		case "owner", "developer", "tester", "viewer":
			if access.Channel.ProjectID == nil || *access.Channel.ProjectID == uuid.Nil {
				continue
			}
			roleUserIDs, roleErr := r.roles.ListProjectMembersByRole(ctx, *access.Channel.ProjectID, group)
			if roleErr != nil {
				return nil, fmt.Errorf("resolve project role mentions: %w", roleErr)
			}
			for _, userID := range roleUserIDs {
				addEligible(userID)
			}
		}
	}

	result := make([]uuid.UUID, 0, len(resolved))
	for userID := range resolved {
		result = append(result, userID)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result, nil
}
