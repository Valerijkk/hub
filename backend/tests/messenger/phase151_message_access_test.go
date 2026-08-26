package messenger_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

type phase151ChannelAccessKey struct {
	actorID   uuid.UUID
	channelID uuid.UUID
	action    messenger.ChannelAction
}

type phase151ChannelAccessResolver struct {
	grants      map[phase151ChannelAccessKey]messenger.ChannelAccess
	calls       []phase151ChannelAccessKey
	onAuthorize func(phase151ChannelAccessKey)
}

func (r *phase151ChannelAccessResolver) AuthorizeChannelAccess(
	_ context.Context,
	actorID uuid.UUID,
	channelID uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	key := phase151ChannelAccessKey{actorID: actorID, channelID: channelID, action: action}
	r.calls = append(r.calls, key)
	if r.onAuthorize != nil {
		r.onAuthorize(key)
	}
	access, ok := r.grants[key]
	if !ok {
		return messenger.ChannelAccess{}, errors.New("phase151 channel access denied")
	}
	return access, nil
}

func phase151Grant(
	resolver *phase151ChannelAccessResolver,
	actorID uuid.UUID,
	channel *ent.Channel,
	action messenger.ChannelAction,
	access messenger.ChannelAccess,
) {
	resolver.grants[phase151ChannelAccessKey{
		actorID: actorID, channelID: channel.ID, action: action,
	}] = access
}

func phase151ReadableAccess(channel *ent.Channel, isMember bool) messenger.ChannelAccess {
	return messenger.ChannelAccess{
		TenantID:  uuid.New(),
		Channel:   channel,
		IsMember:  isMember,
		CanRead:   true,
		CanFollow: isMember,
	}
}

type phase151MessageAccessRepository struct {
	*mockRepository
	timeline           []*messenger.MessageDocument
	listMessagesCalls  int
	findMessageCalls   map[string]int
	addBookmarkCalls   int
	operations         []string
}

func newPhase151MessageAccessRepository() *phase151MessageAccessRepository {
	return &phase151MessageAccessRepository{
		mockRepository:    newMockRepository(),
		findMessageCalls: make(map[string]int),
	}
}

func (r *phase151MessageAccessRepository) ListMessages(
	_ context.Context,
	channelID uuid.UUID,
	_ *time.Time,
	limit int,
) ([]*messenger.MessageDocument, error) {
	r.listMessagesCalls++
	r.operations = append(r.operations, "mongo:list_messages:"+channelID.String())
	result := make([]*messenger.MessageDocument, 0, len(r.timeline))
	for _, document := range r.timeline {
		if document.ChannelID == channelID {
			result = append(result, document)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *phase151MessageAccessRepository) ListBookmarks(ctx context.Context, userID uuid.UUID) ([]*ent.Bookmark, error) {
	r.operations = append(r.operations, "pg:list_bookmarks")
	return r.mockRepository.ListBookmarks(ctx, userID)
}

func (r *phase151MessageAccessRepository) FindMessageByID(
	ctx context.Context,
	messageID string,
) (*messenger.MessageDocument, error) {
	r.findMessageCalls[messageID]++
	r.operations = append(r.operations, "mongo:find_message:"+messageID)
	return r.mockRepository.FindMessageByID(ctx, messageID)
}

func (r *phase151MessageAccessRepository) AddBookmark(
	ctx context.Context,
	userID uuid.UUID,
	messageID string,
	channelID uuid.UUID,
) (*ent.Bookmark, error) {
	r.addBookmarkCalls++
	return r.mockRepository.AddBookmark(ctx, userID, messageID, channelID)
}

func TestPhase151ListMessagesAllowsSameTenantPublicPrejoin(t *testing.T) {
	ctx := context.Background()
	repo := newPhase151MessageAccessRepository()
	ownerID := uuid.New()
	readerID := uuid.New()
	channelID := setupChannelWithMembers(repo.mockRepository, "public", false, ownerID)
	channel := repo.channels[channelID]
	root := &messenger.MessageDocument{
		ID:          "phase151-public-root",
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "public history",
		CreatedAt:   time.Now().UTC(),
	}
	repo.timeline = []*messenger.MessageDocument{root}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	phase151Grant(resolver, readerID, channel, messenger.ChannelActionReadTimeline, phase151ReadableAccess(channel, false))
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	messages, err := service.ListMessages(ctx, channelID, readerID, nil, 50)

	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, root.ID, messages[0].ID)
	assert.Equal(t, []phase151ChannelAccessKey{{
		actorID: readerID, channelID: channelID, action: messenger.ChannelActionReadTimeline,
	}}, resolver.calls)
	assert.Equal(t, 1, repo.listMessagesCalls)
}

func TestPhase151ListMessagesIncludesRootAndCanonicalAlsoSendReplyOnly(t *testing.T) {
	ctx := context.Background()
	repo := newPhase151MessageAccessRepository()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo.mockRepository, "public", false, memberID)
	channel := repo.channels[channelID]
	root := &messenger.MessageDocument{
		ID:          "phase151-root",
		ChannelID:   channelID,
		SenderID:    memberID,
		ContentText: "root",
		CreatedAt:   time.Now().UTC(),
	}
	alsoSend := &messenger.MessageDocument{
		ID:                "phase151-also-send",
		ChannelID:         channelID,
		SenderID:          memberID,
		ContentText:       "visible in both places",
		ThreadID:          root.ID,
		AlsoSendToChannel: true,
		CreatedAt:         time.Now().UTC(),
	}
	threadOnly := &messenger.MessageDocument{
		ID:                "phase151-thread-only",
		ChannelID:         channelID,
		SenderID:          memberID,
		ContentText:       "thread only",
		ThreadID:          root.ID,
		AlsoSendToChannel: false,
		CreatedAt:         time.Now().UTC(),
	}
	repo.timeline = []*messenger.MessageDocument{root, alsoSend, threadOnly}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	phase151Grant(resolver, memberID, channel, messenger.ChannelActionReadTimeline, phase151ReadableAccess(channel, true))
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	messages, err := service.ListMessages(ctx, channelID, memberID, nil, 50)

	require.NoError(t, err)
	require.Len(t, messages, 2)
	assert.Equal(t, []string{root.ID, alsoSend.ID}, []string{messages[0].ID, messages[1].ID})
	assert.Equal(t, root.ID, messages[1].ThreadID)
	assert.Equal(t, 1, repo.listMessagesCalls)
}

func TestPhase151ListBookmarksOmitsDeniedChannelBeforeMongoEnrichment(t *testing.T) {
	ctx := context.Background()
	repo := newPhase151MessageAccessRepository()
	ownerID := uuid.New()
	viewerID := uuid.New()
	publicChannelID := setupChannelWithMembers(repo.mockRepository, "public", false, ownerID)
	privateChannelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID)
	publicChannel := repo.channels[publicChannelID]

	publicMessage := &messenger.MessageDocument{ID: "phase151-public-bookmark", ChannelID: publicChannelID, SenderID: ownerID, ContentText: "visible", CreatedAt: time.Now().UTC()}
	privateMessage := &messenger.MessageDocument{ID: "phase151-private-bookmark", ChannelID: privateChannelID, SenderID: ownerID, ContentText: "secret", CreatedAt: time.Now().UTC()}
	repo.messages[publicMessage.ID] = publicMessage
	repo.messages[privateMessage.ID] = privateMessage
	publicBookmark := &ent.Bookmark{ID: uuid.New(), UserID: viewerID, ChannelID: publicChannelID, MessageID: publicMessage.ID, CreatedAt: time.Now().UTC()}
	privateBookmark := &ent.Bookmark{ID: uuid.New(), UserID: viewerID, ChannelID: privateChannelID, MessageID: privateMessage.ID, CreatedAt: time.Now().UTC()}
	repo.bookmarks[viewerID] = []*ent.Bookmark{publicBookmark, privateBookmark}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	resolver.onAuthorize = func(key phase151ChannelAccessKey) {
		repo.operations = append(repo.operations, "pg:authorize:"+key.channelID.String())
	}
	phase151Grant(resolver, viewerID, publicChannel, messenger.ChannelActionReadTimeline, phase151ReadableAccess(publicChannel, false))
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	bookmarks, err := service.ListBookmarks(ctx, viewerID)

	require.NoError(t, err)
	require.Len(t, bookmarks, 1)
	assert.Equal(t, publicBookmark.ID.String(), bookmarks[0].ID)
	assert.Equal(t, publicChannelID.String(), bookmarks[0].ChannelID)
	assert.Equal(t, publicMessage.ID, bookmarks[0].MessageID)
	assert.NotEqual(t, privateBookmark.ID.String(), bookmarks[0].ID)
	assert.NotEqual(t, privateChannelID.String(), bookmarks[0].ChannelID)
	assert.NotEqual(t, privateMessage.ID, bookmarks[0].MessageID)
	assert.Zero(t, repo.findMessageCalls[privateMessage.ID], "denied bookmark target must not reach Mongo enrichment")
	assert.Equal(t, 1, repo.findMessageCalls[publicMessage.ID])
	require.NotEmpty(t, repo.operations)
	assert.Equal(t, "pg:list_bookmarks", repo.operations[0], "bookmark candidates must come from PostgreSQL before per-channel authorization or Mongo enrichment")
	assert.Contains(t, repo.operations, "pg:authorize:"+publicChannelID.String())
	assert.Contains(t, repo.operations, "pg:authorize:"+privateChannelID.String())
	assert.NotContains(t, repo.operations, "mongo:find_message:"+privateMessage.ID)
	assert.ElementsMatch(t, []phase151ChannelAccessKey{
		{actorID: viewerID, channelID: publicChannelID, action: messenger.ChannelActionReadTimeline},
		{actorID: viewerID, channelID: privateChannelID, action: messenger.ChannelActionReadTimeline},
	}, resolver.calls)
}

func TestPhase151ListBookmarksGroupsReadAuthorityBeforeMongoAndPreservesSourceOrder(t *testing.T) {
	ctx := context.Background()
	repo := newPhase151MessageAccessRepository()
	ownerID := uuid.New()
	viewerID := uuid.New()
	channelAID := setupChannelWithMembers(repo.mockRepository, "public", false, ownerID)
	channelBID := setupChannelWithMembers(repo.mockRepository, "public", false, ownerID)
	channelA := repo.channels[channelAID]
	channelB := repo.channels[channelBID]

	firstBookmark := &ent.Bookmark{ID: uuid.New(), UserID: viewerID, ChannelID: channelAID, MessageID: "phase151-bookmark-a-first", CreatedAt: time.Now().UTC()}
	secondBookmark := &ent.Bookmark{ID: uuid.New(), UserID: viewerID, ChannelID: channelBID, MessageID: "phase151-bookmark-b", CreatedAt: time.Now().UTC()}
	thirdBookmark := &ent.Bookmark{ID: uuid.New(), UserID: viewerID, ChannelID: channelAID, MessageID: "phase151-bookmark-a-second", CreatedAt: time.Now().UTC()}
	repo.bookmarks[viewerID] = []*ent.Bookmark{firstBookmark, secondBookmark, thirdBookmark}
	repo.messages[firstBookmark.MessageID] = &messenger.MessageDocument{ID: firstBookmark.MessageID, ChannelID: channelAID, SenderID: ownerID, ContentText: "first A", CreatedAt: time.Now().UTC()}
	repo.messages[secondBookmark.MessageID] = &messenger.MessageDocument{ID: secondBookmark.MessageID, ChannelID: channelBID, SenderID: ownerID, ContentText: "only B", CreatedAt: time.Now().UTC()}
	repo.messages[thirdBookmark.MessageID] = &messenger.MessageDocument{ID: thirdBookmark.MessageID, ChannelID: channelAID, SenderID: ownerID, ContentText: "second A", CreatedAt: time.Now().UTC()}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	resolver.onAuthorize = func(key phase151ChannelAccessKey) {
		repo.operations = append(repo.operations, "pg:authorize:"+key.channelID.String())
	}
	phase151Grant(resolver, viewerID, channelA, messenger.ChannelActionReadTimeline, phase151ReadableAccess(channelA, false))
	phase151Grant(resolver, viewerID, channelB, messenger.ChannelActionReadTimeline, phase151ReadableAccess(channelB, false))
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	bookmarks, err := service.ListBookmarks(ctx, viewerID)

	require.NoError(t, err)
	require.Len(t, bookmarks, 3)
	assert.Equal(t, []string{firstBookmark.ID.String(), secondBookmark.ID.String(), thirdBookmark.ID.String()}, []string{bookmarks[0].ID, bookmarks[1].ID, bookmarks[2].ID})
	assert.Equal(t, []string{channelAID.String(), channelBID.String(), channelAID.String()}, []string{bookmarks[0].ChannelID, bookmarks[1].ChannelID, bookmarks[2].ChannelID})
	assert.Equal(t, []string{"first A", "only B", "second A"}, []string{bookmarks[0].ContentPreview, bookmarks[1].ContentPreview, bookmarks[2].ContentPreview})
	assert.Equal(t, []phase151ChannelAccessKey{
		{actorID: viewerID, channelID: channelAID, action: messenger.ChannelActionReadTimeline},
		{actorID: viewerID, channelID: channelBID, action: messenger.ChannelActionReadTimeline},
	}, resolver.calls)

	require.NotEmpty(t, repo.operations)
	assert.Equal(t, "pg:list_bookmarks", repo.operations[0])
	firstMongoIndex := -1
	channelAAuthorizationIndex := -1
	channelBAuthorizationIndex := -1
	for index, operation := range repo.operations {
		switch operation {
		case "mongo:find_message:" + firstBookmark.MessageID, "mongo:find_message:" + secondBookmark.MessageID, "mongo:find_message:" + thirdBookmark.MessageID:
			if firstMongoIndex == -1 {
				firstMongoIndex = index
			}
		case "pg:authorize:" + channelAID.String():
			channelAAuthorizationIndex = index
		case "pg:authorize:" + channelBID.String():
			channelBAuthorizationIndex = index
		}
	}
	require.GreaterOrEqual(t, firstMongoIndex, 0)
	require.GreaterOrEqual(t, channelAAuthorizationIndex, 0)
	require.GreaterOrEqual(t, channelBAuthorizationIndex, 0)
	assert.Less(t, channelAAuthorizationIndex, firstMongoIndex)
	assert.Less(t, channelBAuthorizationIndex, firstMongoIndex)
}
func TestPhase151ListBookmarksKeepsCrossChannelBookmarkButOmitsPreview(t *testing.T) {
	ctx := context.Background()
	repo := newPhase151MessageAccessRepository()
	ownerID := uuid.New()
	viewerID := uuid.New()
	readableChannelID := setupChannelWithMembers(repo.mockRepository, "public", false, ownerID)
	foreignChannelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID)
	readableChannel := repo.channels[readableChannelID]

	bookmark := &ent.Bookmark{
		ID:        uuid.New(),
		UserID:    viewerID,
		ChannelID: readableChannelID,
		MessageID: "phase151-cross-channel-bookmark",
		CreatedAt: time.Now().UTC(),
	}
	repo.bookmarks[viewerID] = []*ent.Bookmark{bookmark}
	repo.messages[bookmark.MessageID] = &messenger.MessageDocument{
		ID:          bookmark.MessageID,
		ChannelID:   foreignChannelID,
		SenderID:    ownerID,
		ContentText: "secret preview from another channel",
		CreatedAt:   time.Now().UTC(),
	}

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	phase151Grant(resolver, viewerID, readableChannel, messenger.ChannelActionReadTimeline, phase151ReadableAccess(readableChannel, false))
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	bookmarks, err := service.ListBookmarks(ctx, viewerID)

	require.NoError(t, err)
	require.Len(t, bookmarks, 1, "the bookmark row remains usable for its authorized channel")
	assert.Equal(t, bookmark.ID.String(), bookmarks[0].ID)
	assert.Equal(t, readableChannelID.String(), bookmarks[0].ChannelID)
	assert.Equal(t, bookmark.MessageID, bookmarks[0].MessageID)
	assert.Empty(t, bookmarks[0].ContentPreview, "a message from another channel must not enrich this bookmark")
	assert.NotContains(t, bookmarks[0].ContentPreview, "secret preview from another channel")
	assert.Equal(t, 1, repo.findMessageCalls[bookmark.MessageID])
}
func TestPhase151AddBookmarkProjectIneligibleUserIsDeniedBeforeMongoAndEntWrite(t *testing.T) {
	ctx := context.Background()
	repo := newPhase151MessageAccessRepository()
	ownerID := uuid.New()
	viewerID := uuid.New()
	projectID := uuid.New()
	channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID, viewerID)
	channel := repo.channels[channelID]
	channel.ProjectID = &projectID
	message := &messenger.MessageDocument{ID: "phase151-project-bookmark", ChannelID: channelID, SenderID: ownerID, ContentText: "project restricted", CreatedAt: time.Now().UTC()}
	repo.messages[message.ID] = message

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	err := service.AddBookmark(ctx, viewerID, message.ID, channelID)

	require.Error(t, err)
	assert.Zero(t, repo.findMessageCalls[message.ID], "Bookmark denial must precede Mongo message lookup")
	assert.Zero(t, repo.addBookmarkCalls, "Bookmark denial must precede Ent bookmark write")
	assert.Equal(t, []phase151ChannelAccessKey{{
		actorID: viewerID, channelID: channelID, action: messenger.ChannelActionBookmark,
	}}, resolver.calls)
}

func TestPhase151PromotionUsesChannelBoundRouteAndRejectsOldOrUnauthorizedRouteBeforeMongo(t *testing.T) {
	ownerID := uuid.New()
	readerID := uuid.New()
	repo := newPhase151MessageAccessRepository()
	channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID)
	foreignChannelID := setupChannelWithMembers(repo.mockRepository, "private", false, readerID)
	channel := repo.channels[channelID]
	foreignChannel := repo.channels[foreignChannelID]
	source := &messenger.MessageDocument{ID: "phase151-promote-source", ChannelID: channelID, SenderID: ownerID, ContentText: "promote", CreatedAt: time.Now().UTC()}
	repo.messages[source.ID] = source

	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	promoteAccess := phase151ReadableAccess(channel, true)
	promoteAccess.CanPromote = true
	phase151Grant(resolver, ownerID, channel, messenger.ChannelActionPromote, promoteAccess)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
	service.SetIssuePromoter(fixedIssuePromoter{issueKey: "HUB-151", issueID: uuid.NewString()})
	service.SetPromotePermissionChecker(allowPromotionChecker{})
	app := phase151MessengerRouteApp(service, ownerID)

	t.Run("channel-bound route is available", func(t *testing.T) {
		repo.findMessageCalls = make(map[string]int)
		req := httptest.NewRequest(http.MethodPost, "/api/messenger/channels/"+channelID.String()+"/messages/"+source.ID+"/promote-to-issue", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "route must reach the channel-bound handler before body validation")
		assert.Zero(t, repo.findMessageCalls[source.ID], "invalid request body must not force Mongo lookup")
	})

	t.Run("message-id-only top-level route is unavailable", func(t *testing.T) {
		repo.findMessageCalls = make(map[string]int)
		req := httptest.NewRequest(http.MethodPost, "/api/messages/"+source.ID+"/promote-to-issue", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Zero(t, repo.findMessageCalls[source.ID])
	})

	t.Run("unreadable mismatched route is denied before Mongo", func(t *testing.T) {
		repo.findMessageCalls = make(map[string]int)
		mismatchedApp := phase151MessengerRouteApp(service, readerID)
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/messenger/channels/"+foreignChannel.ID.String()+"/messages/"+source.ID+"/promote-to-issue",
			strings.NewReader(`{"project_id":"`+uuid.NewString()+`","title":"must not disclose"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		resp, err := mismatchedApp.Test(req)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusNotFound, resp.StatusCode, "the route itself must exist")
		assert.NotEqual(t, http.StatusCreated, resp.StatusCode)
		assert.Zero(t, repo.findMessageCalls[source.ID], "denied route channel must not disclose through Mongo")
		assert.Equal(t, []phase151ChannelAccessKey{{
			actorID: readerID, channelID: foreignChannelID, action: messenger.ChannelActionPromote,
		}}, resolver.calls[len(resolver.calls)-1:])
	})
}

func phase151MessengerRouteApp(service *messenger.MessengerService, userID uuid.UUID) *fiber.App {
	app := fiber.New()
	api := app.Group("/api")
	auth := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	messenger.NewHandler(service).RegisterRoutes(api, auth, nil)
	return app
}

func TestPhase151ListMessagesAllowsActivePrivateMemberAfterFreshProof(t *testing.T) {
	ctx := context.Background()
	repo := newPhase151MessageAccessRepository()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID, memberID)
	channel := repo.channels[channelID]
	root := &messenger.MessageDocument{
		ID: "phase151-private-member-root", ChannelID: channelID, SenderID: ownerID,
		ContentText: "member history", CreatedAt: time.Now().UTC(),
	}
	repo.timeline = []*messenger.MessageDocument{root}
	resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
	resolver.onAuthorize = func(key phase151ChannelAccessKey) {
		repo.operations = append(repo.operations, "pg:authorize:"+key.channelID.String())
	}
	phase151Grant(resolver, memberID, channel, messenger.ChannelActionReadTimeline, phase151ReadableAccess(channel, true))
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

	messages, err := service.ListMessages(ctx, channelID, memberID, nil, 50)

	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, root.ID, messages[0].ID)
	assert.Equal(t, []phase151ChannelAccessKey{{
		actorID: memberID, channelID: channelID, action: messenger.ChannelActionReadTimeline,
	}}, resolver.calls)
	assert.Equal(t, []string{
		"pg:authorize:" + channelID.String(),
		"mongo:list_messages:" + channelID.String(),
	}, repo.operations)
	assert.Equal(t, 1, repo.listMessagesCalls)
}

func TestPhase151ListMessagesAuthorityDenialsStopBeforeMongo(t *testing.T) {
	for _, reason := range []string{
		"private nonmember",
		"revoked member",
		"disabled account",
		"cross-tenant actor",
		"inactive or missing project member",
	} {
		t.Run(reason, func(t *testing.T) {
			ctx := context.Background()
			repo := newPhase151MessageAccessRepository()
			ownerID := uuid.New()
			actorID := uuid.New()
			channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID)
			channel := repo.channels[channelID]
			if reason == "inactive or missing project member" {
				projectID := uuid.New()
				channel.ProjectID = &projectID
			}
			secret := &messenger.MessageDocument{
				ID: "phase151-denied-history-" + strings.ReplaceAll(reason, " ", "-"),
				ChannelID: channelID, SenderID: ownerID, ContentText: "must remain undisclosed",
				CreatedAt: time.Now().UTC(),
			}
			repo.timeline = []*messenger.MessageDocument{secret}
			resolver := &phase151ChannelAccessResolver{grants: make(map[phase151ChannelAccessKey]messenger.ChannelAccess)}
			service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)

			messages, err := service.ListMessages(ctx, channelID, actorID, nil, 50)

			require.Error(t, err)
			assert.Empty(t, messages)
			assert.Equal(t, []phase151ChannelAccessKey{{
				actorID: actorID, channelID: channelID, action: messenger.ChannelActionReadTimeline,
			}}, resolver.calls)
			assert.Zero(t, repo.listMessagesCalls, "fresh PG denial must precede every Mongo timeline lookup")
			assert.Empty(t, repo.operations)
			assert.NotContains(t, messages, secret)
		})
	}
}