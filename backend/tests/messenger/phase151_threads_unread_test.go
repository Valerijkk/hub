package messenger_test

import (
	"context"
	"encoding/json"
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
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// The Phase 151 thread reader is allowed to inspect a member follow record
// only after channel access succeeds.  Keeping the spy compile-time checked
// makes a pre-join follow lookup observable instead of relying on JSON alone.
type phase151ThreadFollowLookupSpy interface {
	FindThreadFollow(ctx context.Context, tenantID, userID, channelID uuid.UUID, rootMessageID string) (string, bool, error)
}

// A thread receipt can advance its durable cursor only through this narrow
// repository seam.  Deleted and missing resources must not call it.
type phase151ThreadReceiptCursorSpy interface {
	AdvanceThreadReadCursor(ctx context.Context, tenantID, userID, channelID uuid.UUID, rootMessageID, messageID string) (bool, error)
}

// The unread coordinator has to enumerate persisted eligible follows before
// returning a thread map, then omit follows whose root has disappeared.
type phase151EligibleThreadFollowLookupSpy interface {
	ListEligibleThreadFollowRoots(ctx context.Context, tenantID, userID uuid.UUID) ([]string, error)
}

type phase151ThreadUnreadRepository struct {
	*mockRepository
	findMessageCalls             map[string]int
	listThreadRepliesCalls       int
	threadFollowLookupCalls      int
	threadReceiptAdvanceCalls    int
	threadFollowRootListCalls    int
	persistedThreadFollowRootIDs []string
}

var _ phase151ThreadFollowLookupSpy = (*phase151ThreadUnreadRepository)(nil)
var _ phase151ThreadReceiptCursorSpy = (*phase151ThreadUnreadRepository)(nil)
var _ phase151EligibleThreadFollowLookupSpy = (*phase151ThreadUnreadRepository)(nil)

func newPhase151ThreadUnreadRepository() *phase151ThreadUnreadRepository {
	return &phase151ThreadUnreadRepository{
		mockRepository:   newMockRepository(),
		findMessageCalls: make(map[string]int),
	}
}

func (r *phase151ThreadUnreadRepository) FindMessageByID(
	ctx context.Context,
	messageID string,
) (*messenger.MessageDocument, error) {
	r.findMessageCalls[messageID]++
	return r.mockRepository.FindMessageByID(ctx, messageID)
}

func (r *phase151ThreadUnreadRepository) ListThreadReplies(
	ctx context.Context,
	threadID string,
	before *time.Time,
	limit int,
) ([]*messenger.MessageDocument, error) {
	r.listThreadRepliesCalls++
	return r.mockRepository.ListThreadReplies(ctx, threadID, before, limit)
}

func (r *phase151ThreadUnreadRepository) FindThreadFollow(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	_ string,
) (string, bool, error) {
	r.threadFollowLookupCalls++
	return "manual_follow", true, nil
}

func (r *phase151ThreadUnreadRepository) AdvanceThreadReadCursor(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	_ string,
	_ string,
) (bool, error) {
	r.threadReceiptAdvanceCalls++
	return true, nil
}

func (r *phase151ThreadUnreadRepository) ListEligibleThreadFollowRoots(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
) ([]string, error) {
	r.threadFollowRootListCalls++
	return append([]string(nil), r.persistedThreadFollowRootIDs...), nil
}

type phase151ThreadAccessKey struct {
	actorID   uuid.UUID
	channelID uuid.UUID
	action    messenger.ChannelAction
}

type phase151ThreadAccessResolver struct {
	grants map[phase151ThreadAccessKey]messenger.ChannelAccess
	calls  []phase151ThreadAccessKey
}

func (r *phase151ThreadAccessResolver) AuthorizeChannelAccess(
	_ context.Context,
	actorID uuid.UUID,
	channelID uuid.UUID,
	action messenger.ChannelAction,
) (messenger.ChannelAccess, error) {
	key := phase151ThreadAccessKey{actorID: actorID, channelID: channelID, action: action}
	r.calls = append(r.calls, key)
	access, ok := r.grants[key]
	if !ok {
		return messenger.ChannelAccess{}, apperrors.NewForbidden("phase151 thread access denied")
	}
	return access, nil
}

func phase151GrantThreadAccess(
	resolver *phase151ThreadAccessResolver,
	actorID uuid.UUID,
	channel *ent.Channel,
	action messenger.ChannelAction,
	isMember bool,
) {
	resolver.grants[phase151ThreadAccessKey{
		actorID: actorID, channelID: channel.ID, action: action,
	}] = messenger.ChannelAccess{
		TenantID:  uuid.New(),
		Channel:   channel,
		IsMember:  isMember,
		CanRead:   true,
		CanFollow: isMember,
	}
}

func phase151ThreadRouteApp(service *messenger.MessengerService, userID uuid.UUID) *fiber.App {
	app := fiber.New()
	api := app.Group("/api")
	auth := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	messenger.NewHandler(service).RegisterRoutes(api, auth, nil)
	return app
}

func TestPhase151LegacyRepliesProveRouteChannelBeforeParentMongo(t *testing.T) {
	repo := newPhase151ThreadUnreadRepository()
	ownerID := uuid.New()
	actorID := uuid.New()
	routeChannelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID)
	foreignChannelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID, actorID)
	root := &messenger.MessageDocument{
		ID:          "phase151-legacy-protected-root",
		ChannelID:   foreignChannelID,
		SenderID:    ownerID,
		ContentText: "must never disclose through a forged route channel",
		CreatedAt:   time.Now().UTC(),
	}
	repo.messages[root.ID] = root

	resolver := &phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
	app := phase151ThreadRouteApp(service, actorID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/messenger/channels/"+routeChannelID.String()+"/messages/"+root.ID+"/replies",
		http.NoBody,
	)
	response, err := app.Test(request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.Zero(t, repo.findMessageCalls[root.ID], "route-channel denial must precede parent Mongo lookup")
	assert.Zero(t, repo.listThreadRepliesCalls, "route-channel denial must precede reply Mongo lookup")
	assert.Equal(t, []phase151ThreadAccessKey{{
		actorID: actorID, channelID: routeChannelID, action: messenger.ChannelActionReadParent,
	}}, resolver.calls)
}

func TestPhase151ExactThreadRouteProvesRouteChannelBeforeParentMongo(t *testing.T) {
	repo := newPhase151ThreadUnreadRepository()
	ownerID := uuid.New()
	actorID := uuid.New()
	channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID)
	root := &messenger.MessageDocument{
		ID:          "phase151-exact-protected-root",
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "private thread root",
		CreatedAt:   time.Now().UTC(),
	}
	repo.messages[root.ID] = root

	resolver := &phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)}
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
	app := phase151ThreadRouteApp(service, actorID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/messenger/channels/"+channelID.String()+"/threads/"+root.ID,
		http.NoBody,
	)
	response, err := app.Test(request)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, response.StatusCode)
	assert.Zero(t, repo.findMessageCalls[root.ID], "exact thread route must authorize the route channel before loading its root")
	assert.Zero(t, repo.listThreadRepliesCalls)
	assert.Zero(t, repo.threadFollowLookupCalls)
	assert.Equal(t, []phase151ThreadAccessKey{{
		actorID: actorID, channelID: channelID, action: messenger.ChannelActionReadThread,
	}}, resolver.calls)
}

func TestPhase151PublicPrejoinThreadOmitsFollowModeWithoutFollowLookup(t *testing.T) {
	repo := newPhase151ThreadUnreadRepository()
	ownerID := uuid.New()
	publicReaderID := uuid.New()
	channelID := setupChannelWithMembers(repo.mockRepository, "public", false, ownerID)
	channel := repo.channels[channelID]
	root := &messenger.MessageDocument{
		ID:          "phase151-public-thread-root",
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "public root",
		CreatedAt:   time.Now().UTC(),
	}
	reply := &messenger.MessageDocument{
		ID:          "phase151-public-thread-reply",
		ChannelID:   channelID,
		SenderID:    ownerID,
		ThreadID:    root.ID,
		ContentText: "public reply",
		CreatedAt:   time.Now().UTC(),
	}
	repo.messages[root.ID] = root
	repo.messages[reply.ID] = reply

	resolver := &phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)}
	phase151GrantThreadAccess(resolver, publicReaderID, channel, messenger.ChannelActionReadThread, false)
	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
	app := phase151ThreadRouteApp(service, publicReaderID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/messenger/channels/"+channelID.String()+"/threads/"+root.ID,
		http.NoBody,
	)
	response, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	body := decodePhase151ThreadJSON(t, response)
	assert.Contains(t, body, "parent", "exact thread response must identify its parent root")
	assert.Contains(t, string(body["parent"]), root.ID)
	assert.NotContains(t, body, "follow_mode", "public pre-join readers have no follow state")
	assert.Zero(t, repo.threadFollowLookupCalls, "public pre-join read must not look up a follow record")
	assert.Equal(t, []phase151ThreadAccessKey{{
		actorID: publicReaderID, channelID: channelID, action: messenger.ChannelActionReadThread,
	}}, resolver.calls)
}

func TestPhase151MissingOrDeletedThreadRootIsNonDisclosive(t *testing.T) {
	for _, tt := range []struct {
		name      string
		root      *messenger.MessageDocument
		rootID    string
	}{
		{
			name:   "missing root",
			rootID: "phase151-missing-thread-root",
		},
		{
			name: "externally deleted root",
			root: &messenger.MessageDocument{
				ID:          "phase151-deleted-thread-root",
				ContentText: "deleted root must stay opaque",
				CreatedAt:   time.Now().UTC(),
			},
			rootID: "phase151-deleted-thread-root",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newPhase151ThreadUnreadRepository()
			ownerID := uuid.New()
			memberID := uuid.New()
			channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID, memberID)
			channel := repo.channels[channelID]
			if tt.root != nil {
				deletedAt := time.Now().UTC()
				tt.root.ChannelID = channelID
				tt.root.SenderID = ownerID
				tt.root.DeletedAt = &deletedAt
				repo.messages[tt.root.ID] = tt.root
			}

			resolver := &phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)}
			phase151GrantThreadAccess(resolver, memberID, channel, messenger.ChannelActionReadThread, true)
			service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
			app := phase151ThreadRouteApp(service, memberID)

			request := httptest.NewRequest(
				http.MethodGet,
				"/api/messenger/channels/"+channelID.String()+"/threads/"+tt.rootID,
				http.NoBody,
			)
			response, err := app.Test(request)

			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
			assert.Equal(t, 1, repo.findMessageCalls[tt.rootID], "authorized request may inspect only the requested root")
			assert.Zero(t, repo.listThreadRepliesCalls, "missing/deleted root must not list replies")
			assert.Zero(t, repo.threadFollowLookupCalls, "missing/deleted root must not disclose follow state")
			body, readErr := io.ReadAll(response.Body)
			require.NoError(t, readErr)
			assert.NotContains(t, string(body), tt.rootID)
			assert.NotContains(t, string(body), "created_at")
			assert.NotContains(t, string(body), "follow_mode")
		})
	}
}

func TestPhase151DeletedOrMissingThreadReceiptDoesNotAdvanceCursorOrDiscloseMetadata(t *testing.T) {
	for _, tt := range []struct {
		name        string
		rootDeleted bool
		candidateID string
		candidate   *messenger.MessageDocument
	}{
		{
			name:        "deleted root",
			rootDeleted: true,
			candidateID: "phase151-deleted-root-receipt-candidate",
		},
		{
			name:        "missing reply",
			candidateID: "phase151-missing-receipt-reply",
		},
		{
			name:        "externally deleted reply",
			candidateID: "phase151-deleted-receipt-reply",
			candidate: &messenger.MessageDocument{
				ID:          "phase151-deleted-receipt-reply",
				ContentText: "deleted reply must stay opaque",
				CreatedAt:   time.Now().UTC(),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := newPhase151ThreadUnreadRepository()
			ownerID := uuid.New()
			memberID := uuid.New()
			channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID, memberID)
			channel := repo.channels[channelID]
			root := &messenger.MessageDocument{
				ID:          "phase151-thread-receipt-root",
				ChannelID:   channelID,
				SenderID:    ownerID,
				ContentText: "receipt root",
				CreatedAt:   time.Now().UTC(),
			}
			if tt.rootDeleted {
				deletedAt := time.Now().UTC()
				root.DeletedAt = &deletedAt
			}
			repo.messages[root.ID] = root
			if tt.candidate != nil {
				deletedAt := time.Now().UTC()
				tt.candidate.ChannelID = channelID
				tt.candidate.SenderID = ownerID
				tt.candidate.ThreadID = root.ID
				tt.candidate.DeletedAt = &deletedAt
				repo.messages[tt.candidate.ID] = tt.candidate
			}

			resolver := &phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)}
			phase151GrantThreadAccess(resolver, memberID, channel, messenger.ChannelActionReadThread, true)
			service := messenger.ExportNewMessengerServiceWithChannelAccessTest(repo, resolver)
			app := phase151ThreadRouteApp(service, memberID)

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/messenger/channels/"+channelID.String()+"/threads/"+root.ID+"/read",
				strings.NewReader(`{"category":"followed_thread","message_id":"`+tt.candidateID+`"}`),
			)
			request.Header.Set("Content-Type", "application/json")
			response, err := app.Test(request)

			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
			assert.Equal(t, 1, repo.findMessageCalls[root.ID])
			if !tt.rootDeleted {
				assert.Equal(t, 1, repo.findMessageCalls[tt.candidateID])
			}
			assert.Zero(t, repo.threadReceiptAdvanceCalls, "missing/deleted resources must not advance the followed-thread cursor")
			body, readErr := io.ReadAll(response.Body)
			require.NoError(t, readErr)
			assert.NotContains(t, string(body), root.ID)
			assert.NotContains(t, string(body), tt.candidateID)
			assert.NotContains(t, string(body), "followed_thread")
		})
	}
}

func TestPhase151DeletedFollowRootIsOmittedFromUnreadThreadMap(t *testing.T) {
	repo := newPhase151ThreadUnreadRepository()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo.mockRepository, "private", false, ownerID, memberID)
	deletedRoot := &messenger.MessageDocument{
		ID:          "phase151-stale-unread-thread-root",
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "stale follow root",
		CreatedAt:   time.Now().UTC(),
	}
	deletedAt := time.Now().UTC()
	deletedRoot.DeletedAt = &deletedAt
	repo.messages[deletedRoot.ID] = deletedRoot
	repo.persistedThreadFollowRootIDs = []string{deletedRoot.ID}

	service := messenger.ExportNewMessengerServiceWithChannelAccessTest(
		repo,
		&phase151ThreadAccessResolver{grants: make(map[phase151ThreadAccessKey]messenger.ChannelAccess)},
	)
	app := phase151ThreadRouteApp(service, memberID)
	request := httptest.NewRequest(http.MethodGet, "/api/messenger/unread", http.NoBody)
	response, err := app.Test(request)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	var unread struct {
		Channels map[string]int64 `json:"channels"`
		DMs      map[string]int64 `json:"dms"`
		Mentions map[string]int64 `json:"mentions"`
		Threads  map[string]int64 `json:"threads"`
	}
	body, readErr := io.ReadAll(response.Body)
	require.NoError(t, readErr)
	require.NoError(t, json.Unmarshal(body, &unread))
	assert.NotNil(t, unread.Channels)
	assert.NotNil(t, unread.DMs)
	assert.NotNil(t, unread.Mentions)
	assert.NotNil(t, unread.Threads)
	assert.NotContains(t, unread.Threads, deletedRoot.ID)
	assert.Equal(t, 1, repo.threadFollowRootListCalls, "unread must enumerate persisted follows before it can omit a root deleted externally")
	assert.NotContains(t, string(body), deletedRoot.ID)
}

func decodePhase151ThreadJSON(t *testing.T, response *http.Response) map[string]json.RawMessage {
	t.Helper()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &payload))
	return payload
}
