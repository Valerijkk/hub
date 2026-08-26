package messenger

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
)

// ChannelAction is the closed vocabulary of Messenger channel operations that
// require a fresh PostgreSQL-backed authorization decision.
type ChannelAction string

const (
	ChannelActionReadTimeline  ChannelAction = "read_timeline"
	ChannelActionReadThread    ChannelAction = "read_thread"
	ChannelActionReadParent    ChannelAction = "read_parent"
	ChannelActionSend          ChannelAction = "send"
	ChannelActionReply         ChannelAction = "reply"
	ChannelActionTyping        ChannelAction = "typing"
	ChannelActionEditMessage   ChannelAction = "edit_message"
	ChannelActionForwardSource ChannelAction = "forward_source"
	ChannelActionForwardTarget ChannelAction = "forward_target"
	ChannelActionDeleteMessage ChannelAction = "delete_message"
	ChannelActionPin           ChannelAction = "pin"
	ChannelActionListPins      ChannelAction = "list_pins"
	ChannelActionReact         ChannelAction = "react"
	ChannelActionBookmark      ChannelAction = "bookmark"
	ChannelActionPromote       ChannelAction = "promote"
	ChannelActionReceiveEvent  ChannelAction = "receive_event"
)

var (
	ErrChannelAccessDenied          = errors.New("messenger channel access denied")
	ErrInstanceOrganizationInvalid = errors.New("messenger requires exactly one instance organization")
)

// ChannelAccess is the complete result of one fresh channel authorization.
// Callers must consume capability flags and must not infer authority from Role.
type ChannelAccess struct {
	TenantID                   uuid.UUID
	Channel                    *ent.Channel
	Role                       string
	IsMember                   bool
	HasActiveProjectMembership bool

	CanRead          bool
	CanPost          bool
	CanReply         bool
	CanJoin          bool
	CanFollow        bool
	CanUnfollow      bool
	CanViewUnread    bool
	CanDeleteMessage bool
	CanPin           bool
	CanListPins      bool
	CanType          bool
	CanEdit          bool
	CanForward       bool
	CanReact         bool
	CanBookmark      bool
	CanPromote       bool
	CanReceiveEvent  bool
}

// ChannelAccessResolver is the narrow authority port used by Messenger
// application paths. Implementations resolve the organization server-side.
type ChannelAccessResolver interface {
	AuthorizeChannelAccess(context.Context, uuid.UUID, uuid.UUID, ChannelAction) (ChannelAccess, error)
}

// InstanceOrganizationIDSource exposes only the singleton organization lookup
// needed by the Phase 151 authority adapter.
type InstanceOrganizationIDSource interface {
	ListInstanceOrganizationIDs(context.Context) ([]uuid.UUID, error)
}

// ChannelAuthorityFacts are already-loaded PostgreSQL authority facts. Mongo
// message data is intentionally absent and cannot grant channel access.
type ChannelAuthorityFacts struct {
	TenantID                   uuid.UUID
	ActorID                    uuid.UUID
	ActorOrganizationID        uuid.UUID
	ActorActive                bool
	Channel                    *ent.Channel
	IsMember                   bool
	Role                       string
	HasActiveProjectMembership bool
}

// RecipientChannelProjection is safe to attach to an already-authorized
// recipient delivery. It copies capabilities without role-based inference.
type RecipientChannelProjection struct {
	ID                          string
	Role                        string
	IsMember                    bool
	HasActiveProjectMembership  bool
	CanRead                     bool
	CanPost                     bool
	CanReply                    bool
	CanJoin                     bool
	CanFollow                   bool
	CanUnfollow                 bool
	CanViewUnread               bool
	CanDeleteMessage            bool
	CanPin                      bool
	CanListPins                 bool
	CanType                     bool
	CanEdit                     bool
	CanForward                  bool
	CanReact                    bool
	CanBookmark                 bool
	CanPromote                  bool
	CanReceiveEvent             bool
}

// ExportPhase151ResolveInstanceOrganizationID rejects ambiguous deployments;
// caller-provided organization IDs never select the authorization tenant.
func ExportPhase151ResolveInstanceOrganizationID(
	ctx context.Context,
	source InstanceOrganizationIDSource,
) (uuid.UUID, error) {
	if source == nil {
		return uuid.Nil, ErrInstanceOrganizationInvalid
	}

	ids, err := source.ListInstanceOrganizationIDs(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("list instance organizations: %w", err)
	}
	if len(ids) != 1 || ids[0] == uuid.Nil {
		return uuid.Nil, ErrInstanceOrganizationInvalid
	}
	return ids[0], nil
}

// ExportPhase151ResolveChannelAccessFromFacts applies the closed Phase 151
// capability matrix. Every denied result is zero-valued to avoid metadata
// disclosure to a caller that has no current authority.
func ExportPhase151ResolveChannelAccessFromFacts(
	facts ChannelAuthorityFacts,
	action ChannelAction,
) (ChannelAccess, error) {
	if !isKnownChannelAction(action) ||
		facts.TenantID == uuid.Nil ||
		facts.ActorID == uuid.Nil ||
		facts.ActorOrganizationID != facts.TenantID ||
		!facts.ActorActive ||
		facts.Channel == nil ||
		facts.Channel.ID == uuid.Nil {
		return ChannelAccess{}, ErrChannelAccessDenied
	}
	if facts.Channel.ProjectID != nil && !facts.HasActiveProjectMembership {
		return ChannelAccess{}, ErrChannelAccessDenied
	}

	access := ChannelAccess{
		TenantID:                   facts.TenantID,
		Channel:                    facts.Channel,
		Role:                       facts.Role,
		IsMember:                   facts.IsMember,
		HasActiveProjectMembership: facts.HasActiveProjectMembership,
	}

	if !facts.IsMember {
		if string(facts.Channel.ChannelType) != "public" || action != ChannelActionReadTimeline {
			return ChannelAccess{}, ErrChannelAccessDenied
		}
		access.CanRead = true
		access.CanJoin = true
		return access, nil
	}

	privileged := isPrivilegedChannelRole(facts.Role)
	canWrite := !facts.Channel.IsReadonly || privileged
	access.CanRead = true
	access.CanPost = canWrite
	access.CanReply = canWrite
	access.CanFollow = true
	access.CanUnfollow = true
	access.CanViewUnread = true
	access.CanDeleteMessage = privileged
	access.CanPin = privileged
	access.CanListPins = true
	access.CanType = true
	access.CanEdit = canWrite
	access.CanForward = canWrite
	access.CanReact = true
	access.CanBookmark = true
	access.CanPromote = privileged
	access.CanReceiveEvent = true

	if !channelActionAllowed(access, action) {
		return ChannelAccess{}, ErrChannelAccessDenied
	}
	return access, nil
}

func ExportPhase151RecipientChannelProjection(access ChannelAccess) RecipientChannelProjection {
	projection := RecipientChannelProjection{
		Role:                       access.Role,
		IsMember:                   access.IsMember,
		HasActiveProjectMembership: access.HasActiveProjectMembership,
		CanRead:                    access.CanRead,
		CanPost:                    access.CanPost,
		CanReply:                   access.CanReply,
		CanJoin:                    access.CanJoin,
		CanFollow:                  access.CanFollow,
		CanUnfollow:                access.CanUnfollow,
		CanViewUnread:              access.CanViewUnread,
		CanDeleteMessage:           access.CanDeleteMessage,
		CanPin:                     access.CanPin,
		CanListPins:                access.CanListPins,
		CanType:                    access.CanType,
		CanEdit:                    access.CanEdit,
		CanForward:                 access.CanForward,
		CanReact:                   access.CanReact,
		CanBookmark:                access.CanBookmark,
		CanPromote:                 access.CanPromote,
		CanReceiveEvent:            access.CanReceiveEvent,
	}
	if access.Channel != nil {
		projection.ID = access.Channel.ID.String()
	}
	return projection
}

func isKnownChannelAction(action ChannelAction) bool {
	switch action {
	case ChannelActionReadTimeline,
		ChannelActionReadThread,
		ChannelActionReadParent,
		ChannelActionSend,
		ChannelActionReply,
		ChannelActionTyping,
		ChannelActionEditMessage,
		ChannelActionForwardSource,
		ChannelActionForwardTarget,
		ChannelActionDeleteMessage,
		ChannelActionPin,
		ChannelActionListPins,
		ChannelActionReact,
		ChannelActionBookmark,
		ChannelActionPromote,
		ChannelActionReceiveEvent:
		return true
	default:
		return false
	}
}

func channelActionAllowed(access ChannelAccess, action ChannelAction) bool {
	switch action {
	case ChannelActionReadTimeline, ChannelActionReadThread, ChannelActionReadParent:
		return access.CanRead
	case ChannelActionSend:
		return access.CanPost
	case ChannelActionReply:
		return access.CanReply
	case ChannelActionTyping:
		return access.CanType
	case ChannelActionEditMessage:
		return access.CanEdit
	case ChannelActionForwardSource:
		return access.CanRead && access.CanForward
	case ChannelActionForwardTarget:
		return access.CanPost
	case ChannelActionDeleteMessage:
		// Stage 1 proves live member/project/channel access before the single
		// Mongo ownership lookup. CanDeleteMessage remains the delete-any
		// moderator capability projected to clients.
		return access.IsMember && access.CanRead
	case ChannelActionPin:
		return access.CanPin
	case ChannelActionListPins:
		return access.CanListPins
	case ChannelActionReact:
		return access.CanReact
	case ChannelActionBookmark:
		return access.CanBookmark
	case ChannelActionPromote:
		return access.CanPromote
	case ChannelActionReceiveEvent:
		return access.CanReceiveEvent
	default:
		return false
	}
}

func isPrivilegedChannelRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "admin", "moderator":
		return true
	default:
		return false
	}
}
