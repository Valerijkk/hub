package messenger

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type BotMongoPath string

const (
	BotMongoPathSend                         BotMongoPath = "send"
	BotMongoPathUpdate                       BotMongoPath = "update"
	BotMongoPathFind                         BotMongoPath = "find"
	BotMongoPathList                         BotMongoPath = "list"
	BotMongoPathGetChannelEntity             BotMongoPath = "getChannelEntity"
	BotMongoPathGetChannelMemberIDsInternal  BotMongoPath = "getChannelMemberIDsInternal"
)

type BotActorKind string

const (
	BotActorKindSystem BotActorKind = "system"
	BotActorKindUser   BotActorKind = "user"
	BotActorSystem                    = BotActorKindSystem
	BotActorUser                      = BotActorKindUser
)

type BotKind string

const (
	BotKindHubAssistant BotKind = "hub_assistant"
	BotKindSystem       BotKind = "system"
)

type BotChannelAction string

type AuthorizedBotChannelScope struct {
	TenantID        uuid.UUID
	ActorID         uuid.UUID
	ChannelID       uuid.UUID
	ProjectID       uuid.UUID
	IntegrationID   uuid.UUID
	ActorKind       BotActorKind
	BotKind         BotKind
	BotEnabled      bool
	ProjectEligible bool
	Action          BotChannelAction
}

type BotMongoCommand struct {
	Path          BotMongoPath
	TenantID      uuid.UUID
	ChannelID     uuid.UUID
	ProjectID     uuid.UUID
	IntegrationID uuid.UUID
	Action        BotChannelAction
}

type AuthorizedBotMongoPathExecutor interface {
	ExecuteAuthorizedBotMongoPath(
		context.Context,
		BotMongoPath,
		*AuthorizedBotChannelScope,
		BotMongoCommand,
	) error
}

var ErrAuthorizedBotScopeDenied = errors.New("authorized bot scope denied")

func ExportPhase151ExecuteBotMongoPath(
	ctx context.Context,
	path BotMongoPath,
	scope *AuthorizedBotChannelScope,
	command BotMongoCommand,
	allowedKinds []BotKind,
	executor AuthorizedBotMongoPathExecutor,
) error {
	if !isKnownBotMongoPath(path) || command.Path != path || scope == nil || executor == nil {
		return ErrAuthorizedBotScopeDenied
	}

	if scope.TenantID == uuid.Nil ||
		scope.ActorID == uuid.Nil ||
		scope.ChannelID == uuid.Nil ||
		scope.ProjectID == uuid.Nil ||
		scope.IntegrationID == uuid.Nil ||
		command.TenantID == uuid.Nil ||
		command.ChannelID == uuid.Nil ||
		command.ProjectID == uuid.Nil ||
		command.IntegrationID == uuid.Nil {
		return ErrAuthorizedBotScopeDenied
	}

	if !scope.BotEnabled || !scope.ProjectEligible || scope.ActorKind != BotActorKindSystem {
		return ErrAuthorizedBotScopeDenied
	}

	if scope.TenantID != command.TenantID ||
		scope.ChannelID != command.ChannelID ||
		scope.ProjectID != command.ProjectID ||
		scope.IntegrationID != command.IntegrationID {
		return ErrAuthorizedBotScopeDenied
	}

	if scope.Action != command.Action || command.Action != BotChannelAction(path) {
		return ErrAuthorizedBotScopeDenied
	}

	if !isAllowedBotKind(scope.BotKind, allowedKinds) {
		return ErrAuthorizedBotScopeDenied
	}

	return executor.ExecuteAuthorizedBotMongoPath(ctx, path, scope, command)
}

func isKnownBotMongoPath(path BotMongoPath) bool {
	switch path {
	case BotMongoPathSend,
		BotMongoPathUpdate,
		BotMongoPathFind,
		BotMongoPathList,
		BotMongoPathGetChannelEntity,
		BotMongoPathGetChannelMemberIDsInternal:
		return true
	default:
		return false
	}
}

func isAllowedBotKind(kind BotKind, allowedKinds []BotKind) bool {
	for _, allowedKind := range allowedKinds {
		if kind == allowedKind {
			return true
		}
	}

	return false
}
