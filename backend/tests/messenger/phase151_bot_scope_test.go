package messenger_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/messenger"
)

var phase151BotMongoPaths = []messenger.BotMongoPath{
	messenger.BotMongoPathSend,
	messenger.BotMongoPathUpdate,
	messenger.BotMongoPathFind,
	messenger.BotMongoPathList,
	messenger.BotMongoPathGetChannelEntity,
	messenger.BotMongoPathGetChannelMemberIDsInternal,
}

type phase151BotMongoCall struct {
	path    messenger.BotMongoPath
	scope   *messenger.AuthorizedBotChannelScope
	command messenger.BotMongoCommand
}

type phase151BotMongoRecorder struct {
	calls []phase151BotMongoCall
}

func (r *phase151BotMongoRecorder) ExecuteAuthorizedBotMongoPath(
	_ context.Context,
	path messenger.BotMongoPath,
	scope *messenger.AuthorizedBotChannelScope,
	command messenger.BotMongoCommand,
) error {
	r.calls = append(r.calls, phase151BotMongoCall{path: path, scope: scope, command: command})
	return nil
}

func TestPhase151BotMongoPaths_RequireExactAuthorizedBotChannelScope(t *testing.T) {
	ctx := context.Background()
	for _, path := range phase151BotMongoPaths {
		t.Run(string(path), func(t *testing.T) {
			recorder := &phase151BotMongoRecorder{}
			command := phase151BotMongoCommand(path)
			scope := phase151AuthorizedBotScope(command)

			err := messenger.ExportPhase151ExecuteBotMongoPath(
				ctx, path, scope, command, []messenger.BotKind{messenger.BotKindHubAssistant}, recorder,
			)

			require.NoError(t, err)
			require.Len(t, recorder.calls, 1)
			assert.Equal(t, phase151BotMongoCall{path: path, scope: scope, command: command}, recorder.calls[0])
		})
	}
}

func TestPhase151BotMongoPaths_DeniedProofNeverTouchesMongo(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name   string
		mutate func(*messenger.AuthorizedBotChannelScope, *messenger.BotMongoCommand)
		nil    bool
	}{
		{name: "nil scope", nil: true},
		{name: "disabled bot", mutate: func(scope *messenger.AuthorizedBotChannelScope, _ *messenger.BotMongoCommand) { scope.BotEnabled = false }},
		{name: "lost project", mutate: func(scope *messenger.AuthorizedBotChannelScope, _ *messenger.BotMongoCommand) { scope.ProjectEligible = false }},
		{name: "integration mismatch", mutate: func(_ *messenger.AuthorizedBotChannelScope, command *messenger.BotMongoCommand) { command.IntegrationID = uuid.New() }},
		{name: "non system actor", mutate: func(scope *messenger.AuthorizedBotChannelScope, _ *messenger.BotMongoCommand) { scope.ActorKind = messenger.BotActorUser }},
		{name: "bot kind not allowlisted", mutate: func(scope *messenger.AuthorizedBotChannelScope, _ *messenger.BotMongoCommand) { scope.BotKind = messenger.BotKind("untrusted") }},
		{name: "tenant mismatch", mutate: func(_ *messenger.AuthorizedBotChannelScope, command *messenger.BotMongoCommand) { command.TenantID = uuid.New() }},
		{name: "channel mismatch", mutate: func(_ *messenger.AuthorizedBotChannelScope, command *messenger.BotMongoCommand) { command.ChannelID = uuid.New() }},
		{name: "project mismatch", mutate: func(_ *messenger.AuthorizedBotChannelScope, command *messenger.BotMongoCommand) { command.ProjectID = uuid.New() }},
		{name: "action mismatch", mutate: func(_ *messenger.AuthorizedBotChannelScope, command *messenger.BotMongoCommand) { command.Action = messenger.BotChannelAction("wrong") }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, path := range phase151BotMongoPaths {
				recorder := &phase151BotMongoRecorder{}
				command := phase151BotMongoCommand(path)
				scope := phase151AuthorizedBotScope(command)
				if tt.nil {
					scope = nil
				} else {
					tt.mutate(scope, &command)
				}

				err := messenger.ExportPhase151ExecuteBotMongoPath(
					ctx, path, scope, command, []messenger.BotKind{messenger.BotKindHubAssistant}, recorder,
				)

				assert.ErrorIs(t, err, messenger.ErrAuthorizedBotScopeDenied)
				assert.Empty(t, recorder.calls, "%s must be denied before Mongo", path)
			}
		})
	}
}

func phase151AuthorizedBotScope(command messenger.BotMongoCommand) *messenger.AuthorizedBotChannelScope {
	return &messenger.AuthorizedBotChannelScope{
		TenantID: command.TenantID, ActorID: uuid.New(), ActorKind: messenger.BotActorSystem,
		ChannelID: command.ChannelID, ProjectID: command.ProjectID, IntegrationID: command.IntegrationID,
		BotKind: messenger.BotKindHubAssistant, BotEnabled: true, ProjectEligible: true, Action: command.Action,
	}
}

func phase151BotMongoCommand(path messenger.BotMongoPath) messenger.BotMongoCommand {
	return messenger.BotMongoCommand{
		Path: path, TenantID: uuid.New(), ChannelID: uuid.New(), ProjectID: uuid.New(), IntegrationID: uuid.New(),
		Action: messenger.BotChannelAction(path),
	}
}
