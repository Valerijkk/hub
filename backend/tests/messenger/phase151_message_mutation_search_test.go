package messenger_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

var _ messenger.AuthorizedMessageSearchScopeResolver = (*phase151MessageSearchScopeFixture)(nil)
var _ messenger.AuthorizedUnreadScopeResolver = (*phase151MessageSearchScopeFixture)(nil)

type phase151MutationRepository struct {
	*mockRepository
	operations  []string
	findCalls   int
	updateCalls int
	insertCalls int
}

func newPhase151MutationRepository() *phase151MutationRepository {
	return &phase151MutationRepository{mockRepository: newMockRepository()}
}

func (r *phase151MutationRepository) FindMessageByID(
	ctx context.Context,
	messageID string,
) (*messenger.MessageDocument, error) {
	r.findCalls++
	r.operations = append(r.operations, "mongo:find:"+messageID)
	return r.mockRepository.FindMessageByID(ctx, messageID)
}

func (r *phase151MutationRepository) UpdateMessage(
	ctx context.Context,
	messageID string,
	contentJSON map[string]interface{},
	contentText string,
) (*messenger.MessageDocument, error) {
	r.updateCalls++
	r.operations = append(r.operations, "mongo:update:"+messageID)
	return r.mockRepository.UpdateMessage(ctx, messageID, contentJSON, contentText)
}

func (r *phase151MutationRepository) InsertMessage(
	ctx context.Context,
	doc *messenger.MessageDocument,
) (string, error) {
	r.insertCalls++
	r.operations = append(r.operations, "mongo:insert:"+doc.ChannelID.String())
	return r.mockRepository.InsertMessage(ctx, doc)
}

type phase151MessageSearchRepository struct {
	*mockRepository
	searchCalls  int
	searchScopes [][]uuid.UUID
}

func newPhase151MessageSearchRepository() *phase151MessageSearchRepository {
	return &phase151MessageSearchRepository{mockRepository: newMockRepository()}
}

func (r *phase151MessageSearchRepository) SearchMessages(
	_ context.Context,
	_ string,
	channelIDs []uuid.UUID,
	_ int,
) ([]*messenger.MessageDocument, error) {
	r.searchCalls++
	r.searchScopes = append(r.searchScopes, append([]uuid.UUID(nil), channelIDs...))
	return nil, nil
}

type phase151MessageSearchScopeFixture struct {
	scope            messenger.AuthorizedMessageSearchScope
	err              error
	calls            []uuid.UUID
	unreadScopeCalls int
}

func (f *phase151MessageSearchScopeFixture) ResolveAuthorizedMessageSearchScope(
	_ context.Context,
	userID uuid.UUID,
) (messenger.AuthorizedMessageSearchScope, error) {
	f.calls = append(f.calls, userID)
	if f.err != nil {
		return messenger.AuthorizedMessageSearchScope{}, f.err
	}
	return f.scope, nil
}

func (f *phase151MessageSearchScopeFixture) ResolveAuthorizedUnreadScope(
	_ context.Context,
	_ uuid.UUID,
) (messenger.AuthorizedUnreadScope, error) {
	f.unreadScopeCalls++
	return messenger.AuthorizedUnreadScope{}, errors.New("message search must not resolve AuthorizedUnreadScope")
}

func TestPhase151EditMessageRequiresFreshAccessBeforeMongoLookupOrMutation(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	channelID := uuid.New()
	messageID := "phase151-edit-fresh-proof"
	repo := newPhase151MutationRepository()
	channel := &ent.Channel{ID: channelID, ChannelType: "private"}
	repo.channels[channelID] = channel
	repo.members[channelID] = map[uuid.UUID]*ent.ChannelMember{
		actorID: {UserID: actorID, ChannelID: channelID},
	}
	repo.messages[messageID] = &messenger.MessageDocument{
		ID:        messageID,
		ChannelID: channelID,
		SenderID:  actorID,
	}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	resolver.onAuthorize = func(key phase151ChannelAccessKey) {
		repo.operations = append(repo.operations, "pg:authorize:"+key.channelID.String())
	}
	editAccess := phase151ReadableAccess(channel, true)
	editAccess.CanEdit = true
	phase151Grant(
		resolver,
		actorID,
		channel,
		messenger.ChannelActionEditMessage,
		editAccess,
	)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	updated, err := service.EditMessage(ctx, channelID, messageID, actorID, messenger.EditMessageRequest{
		ContentText: "freshly authorised edit",
	})

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "freshly authorised edit", updated.ContentText)
	assert.Equal(t, []phase151ChannelAccessKey{{
		actorID: actorID, channelID: channelID, action: messenger.ChannelActionEditMessage,
	}}, resolver.calls)
	assert.Equal(t, []string{
		"pg:authorize:" + channelID.String(),
		"mongo:find:" + messageID,
		"mongo:update:" + messageID,
	}, repo.operations)
	assert.Equal(t, 1, repo.findCalls)
	assert.Equal(t, 1, repo.updateCalls)
}

func TestPhase151DeniedEditStopsBeforeAnyMongoLookupOrMutation(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	channelID := uuid.New()
	messageID := "phase151-denied-edit"
	repo := newPhase151MutationRepository()
	repo.messages[messageID] = &messenger.MessageDocument{ID: messageID, ChannelID: channelID, SenderID: actorID}
	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	updated, err := service.EditMessage(ctx, channelID, messageID, actorID, messenger.EditMessageRequest{ContentText: "must not write"})

	require.Error(t, err)
	assert.Nil(t, updated)
	assert.Equal(t, []phase151ChannelAccessKey{{
		actorID: actorID, channelID: channelID, action: messenger.ChannelActionEditMessage,
	}}, resolver.calls)
	assert.Zero(t, repo.findCalls, "denied edit must not discover a Mongo message")
	assert.Zero(t, repo.updateCalls, "denied edit must not mutate Mongo")
	assert.Empty(t, repo.operations)
}

func TestPhase151ForwardRequiresFreshSourceAndTargetProofBeforeMongo(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	sourceChannelID := uuid.New()
	targetChannelID := uuid.New()
	sourceMessageID := "phase151-forward-source"
	repo := newPhase151MutationRepository()
	sourceChannel := &ent.Channel{ID: sourceChannelID, ChannelType: "private"}
	targetChannel := &ent.Channel{ID: targetChannelID, ChannelType: "private"}
	repo.channels[sourceChannelID] = sourceChannel
	repo.channels[targetChannelID] = targetChannel
	repo.members[sourceChannelID] = map[uuid.UUID]*ent.ChannelMember{actorID: {UserID: actorID, ChannelID: sourceChannelID}}
	repo.members[targetChannelID] = map[uuid.UUID]*ent.ChannelMember{actorID: {UserID: actorID, ChannelID: targetChannelID}}
	repo.messages[sourceMessageID] = &messenger.MessageDocument{
		ID:          sourceMessageID,
		ChannelID:   sourceChannelID,
		SenderID:    uuid.New(),
		ContentText: "authorised source",
	}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	resolver.onAuthorize = func(key phase151ChannelAccessKey) {
		repo.operations = append(repo.operations, "pg:authorize:"+key.channelID.String())
	}
	sourceAccess := phase151ReadableAccess(sourceChannel, true)
	sourceAccess.CanForward = true
	phase151Grant(resolver, actorID, sourceChannel, messenger.ChannelActionForwardSource, sourceAccess)
	targetAccess := phase151ReadableAccess(targetChannel, true)
	targetAccess.CanPost = true
	phase151Grant(resolver, actorID, targetChannel, messenger.ChannelActionForwardTarget, targetAccess)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	forwarded, err := service.ForwardMessage(ctx, sourceChannelID, actorID, sourceMessageID, targetChannelID.String())

	require.NoError(t, err)
	require.NotNil(t, forwarded)
	assert.Equal(t, targetChannelID.String(), forwarded.ChannelID)
	assert.ElementsMatch(t, []phase151ChannelAccessKey{
		{actorID: actorID, channelID: sourceChannelID, action: messenger.ChannelActionForwardSource},
		{actorID: actorID, channelID: targetChannelID, action: messenger.ChannelActionForwardTarget},
	}, resolver.calls)
	require.Len(t, repo.operations, 4)
	assert.Contains(t, repo.operations[:2], "pg:authorize:"+sourceChannelID.String())
	assert.Contains(t, repo.operations[:2], "pg:authorize:"+targetChannelID.String())
	assert.Equal(t, "mongo:find:"+sourceMessageID, repo.operations[2])
	assert.Equal(t, "mongo:insert:"+targetChannelID.String(), repo.operations[3])
	assert.Equal(t, 1, repo.findCalls)
	assert.Equal(t, 1, repo.insertCalls)
}

func TestPhase151DeniedForwardTargetStopsBeforeSourceMongoLookup(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	sourceChannelID := uuid.New()
	targetChannelID := uuid.New()
	sourceMessageID := "phase151-denied-forward-target"
	repo := newPhase151MutationRepository()
	sourceChannel := &ent.Channel{ID: sourceChannelID, ChannelType: "private"}
	targetChannel := &ent.Channel{ID: targetChannelID, ChannelType: "private"}
	repo.channels[sourceChannelID] = sourceChannel
	repo.channels[targetChannelID] = targetChannel
	repo.messages[sourceMessageID] = &messenger.MessageDocument{ID: sourceMessageID, ChannelID: sourceChannelID, SenderID: uuid.New()}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	sourceAccess := phase151ReadableAccess(sourceChannel, true)
	sourceAccess.CanForward = true
	phase151Grant(resolver, actorID, sourceChannel, messenger.ChannelActionForwardSource, sourceAccess)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	forwarded, err := service.ForwardMessage(ctx, sourceChannelID, actorID, sourceMessageID, targetChannelID.String())

	require.Error(t, err)
	assert.Nil(t, forwarded)
	assert.ElementsMatch(t, []phase151ChannelAccessKey{
		{actorID: actorID, channelID: sourceChannelID, action: messenger.ChannelActionForwardSource},
		{actorID: actorID, channelID: targetChannelID, action: messenger.ChannelActionForwardTarget},
	}, resolver.calls)
	assert.Zero(t, repo.findCalls, "a denied target must not disclose the source message")
	assert.Zero(t, repo.insertCalls, "a denied target must not create a Mongo message")
	assert.Empty(t, repo.operations)
}

func TestPhase151MessengerSearchUsesAuthorizedMessageSearchScopeOnly(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	publicChannelID := uuid.New()
	memberChannelID := uuid.New()
	activeProjectChannelID := uuid.New()
	tenantID := uuid.New()
	repo := newPhase151MessageSearchRepository()
	scopeResolver := &phase151MessageSearchScopeFixture{scope: messenger.AuthorizedMessageSearchScope{
		TenantID: tenantID,
		UserID:   userID,
		Targets: []messenger.AuthorizedMessageSearchTarget{
			{ChannelID: publicChannelID, Kind: "channel"},
			{ChannelID: memberChannelID, Kind: "channel"},
			// The resolver returns this target only after fresh active ProjectMember authorization.
			{ChannelID: activeProjectChannelID, Kind: "channel", IsProjectChannel: true},
		},
	}}
	service := messenger.ExportNewMessengerServiceWithChannelAccessAndSearchScopeTest(
		repo,
		phase151Deny(messenger.ChannelActionReadTimeline),
		scopeResolver,
	)

	results, err := service.SearchMessages(ctx, userID, "authorised", "", 20)

	require.NoError(t, err)
	assert.Empty(t, results)
	assert.Equal(t, []uuid.UUID{userID}, scopeResolver.calls)
	assert.Zero(t, scopeResolver.unreadScopeCalls, "message search must not derive candidates from AuthorizedUnreadScope")
	require.Equal(t, 1, repo.searchCalls)
	assert.Equal(t, [][]uuid.UUID{{publicChannelID, memberChannelID, activeProjectChannelID}}, repo.searchScopes)
}

func TestPhase151MessengerSearchEmptyOrDeniedScopeNeverReachesMongo(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	for name, fixture := range map[string]*phase151MessageSearchScopeFixture{
		"empty": {
			scope: messenger.AuthorizedMessageSearchScope{TenantID: uuid.New(), UserID: userID},
		},
		"denied": {
			err: errors.New("scope denied"),
		},
		"revoked": {
			err: errors.New("membership revoked"),
		},
		"disabled": {
			err: errors.New("account disabled"),
		},
		"cross-tenant": {
			err: errors.New("cross-tenant scope denied"),
		},
		"private-nonmember": {
			err: errors.New("private non-member scope denied"),
		},
		"project-ineligible": {
			err: errors.New("project-ineligible scope denied"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			repo := newPhase151MessageSearchRepository()
			service := messenger.ExportNewMessengerServiceWithChannelAccessAndSearchScopeTest(
				repo,
				phase151Deny(messenger.ChannelActionReadTimeline),
				fixture,
			)

			results, _ := service.SearchMessages(ctx, userID, "must not query Mongo", "", 20)

			assert.Empty(t, results)
			assert.Equal(t, []uuid.UUID{userID}, fixture.calls)
			assert.Zero(t, fixture.unreadScopeCalls, "message search must not fall back to AuthorizedUnreadScope")
			assert.Zero(t, repo.searchCalls, "empty or denied scope must not discover message candidates in Mongo")
			assert.Empty(t, repo.searchScopes)
		})
	}
}
func TestPhase151MessengerSearchNilScopeResolverFailsClosedBeforeMongo(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newPhase151MessageSearchRepository()
	service := messenger.ExportNewMessengerServiceWithChannelAccessAndSearchScopeTest(
		repo,
		phase151Deny(messenger.ChannelActionReadTimeline),
		nil,
	)

	results, _ := service.SearchMessages(ctx, userID, "must not query Mongo", "", 20)

	assert.Empty(t, results)
	assert.Zero(t, repo.searchCalls, "a nil message-search scope resolver must fail closed before Mongo")
	assert.Empty(t, repo.searchScopes)
}