package tracker_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tracker"
)

var _ tracker.Repository = (*issueMessageLinkListAuthzRepo)(nil)

type issueMessageLinkListAuthzRepo struct {
	tracker.Repository

	findErr          error
	projectMember    bool
	projectMemberErr error
	member           bool
	memberErr        error
	links            []*ent.IssueMessageLink
	listErr          error

	calls []string

	findProjectID          uuid.UUID
	findChannelID          uuid.UUID
	projectMemberProjectID uuid.UUID
	projectMemberActorID   uuid.UUID
	memberChannelID        uuid.UUID
	memberActorID          uuid.UUID
	listChannelID          uuid.UUID
	listMessageID          string
	memberCalls            int
	listCalls              int
}

func (r *issueMessageLinkListAuthzRepo) FindChannelInProject(
	_ context.Context,
	projectID uuid.UUID,
	channelID uuid.UUID,
) (*ent.Channel, error) {
	r.calls = append(r.calls, "find_channel_in_project")
	r.findProjectID = projectID
	r.findChannelID = channelID
	if r.findErr != nil {
		return nil, r.findErr
	}
	return nil, nil
}

func (r *issueMessageLinkListAuthzRepo) IsProjectMember(
	_ context.Context,
	projectID uuid.UUID,
	actorID uuid.UUID,
) (bool, error) {
	r.calls = append(r.calls, "is_project_member")
	r.projectMemberProjectID = projectID
	r.projectMemberActorID = actorID
	if r.projectMemberErr != nil {
		return false, r.projectMemberErr
	}
	return r.projectMember, nil
}

func (r *issueMessageLinkListAuthzRepo) IsChannelMember(
	_ context.Context,
	channelID uuid.UUID,
	actorID uuid.UUID,
) (bool, error) {
	r.calls = append(r.calls, "is_channel_member")
	r.memberCalls++
	r.memberChannelID = channelID
	r.memberActorID = actorID
	if r.memberErr != nil {
		return false, r.memberErr
	}
	return r.member, nil
}

func (r *issueMessageLinkListAuthzRepo) ListIssueMessageLinksByThread(
	_ context.Context,
	channelID uuid.UUID,
	messageID string,
) ([]*ent.IssueMessageLink, error) {
	r.calls = append(r.calls, "list_thread_links")
	r.listCalls++
	r.listChannelID = channelID
	r.listMessageID = messageID
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.links, nil
}

func TestListThreadIssueLinksAuthorizesBeforeLinkReadAndKeepsProjectRowFilter(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	foreignProjectID := uuid.New()
	channelID := uuid.New()
	actorID := uuid.New()
	messageID := "thread-root"

	matching := issueMessageLinkWithProject(projectID, channelID, actorID, messageID, "HUB-151")
	foreign := issueMessageLinkWithProject(foreignProjectID, channelID, actorID, messageID, "OTHER-1")
	missingIssue := &ent.IssueMessageLink{ID: uuid.New()}
	repo := &issueMessageLinkListAuthzRepo{
		projectMember: true,
		member:        true,
		links:         []*ent.IssueMessageLink{matching, foreign, missingIssue},
	}
	service := tracker.NewTrackerService(repo, nil)

	links, err := service.ListThreadIssueLinks(ctx, projectID, channelID, actorID, messageID)

	require.NoError(t, err)
	assert.Equal(t, []string{"find_channel_in_project", "is_project_member", "is_channel_member", "list_thread_links"}, repo.calls)
	assert.Equal(t, projectID, repo.findProjectID)
	assert.Equal(t, channelID, repo.findChannelID)
	assert.Equal(t, projectID, repo.projectMemberProjectID)
	assert.Equal(t, actorID, repo.projectMemberActorID)
	assert.Equal(t, channelID, repo.memberChannelID)
	assert.Equal(t, actorID, repo.memberActorID)
	assert.Equal(t, channelID, repo.listChannelID)
	assert.Equal(t, messageID, repo.listMessageID)
	require.Len(t, links, 1)
	assert.Equal(t, matching.ID.String(), links[0].ID)
	assert.Equal(t, "HUB-151", links[0].IssueKey)
	assert.Equal(t, "issue in requested project", links[0].IssueTitle)
}

func TestListThreadIssueLinksConcealsUnprovedChannelsBeforeLinkRead(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	channelID := uuid.New()
	actorID := uuid.New()

	for _, testCase := range []struct {
		name                string
		findErr             error
		projectMember       bool
		member              bool
		expectedMemberCalls int
		expectedCalls       []string
	}{
		{
			name:          "global channel",
			findErr:       &ent.NotFoundError{},
			expectedCalls: []string{"find_channel_in_project"},
		},
		{
			name:          "foreign project channel",
			findErr:       &ent.NotFoundError{},
			expectedCalls: []string{"find_channel_in_project"},
		},
		{
			name:          "deleted channel",
			findErr:       &ent.NotFoundError{},
			expectedCalls: []string{"find_channel_in_project"},
		},
		{
			name:          "inactive project member",
			member:        true,
			expectedCalls: []string{"find_channel_in_project", "is_project_member"},
		},
		{
			name:                "nonmember",
			projectMember:       true,
			member:              false,
			expectedMemberCalls: 1,
			expectedCalls:       []string{"find_channel_in_project", "is_project_member", "is_channel_member"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &issueMessageLinkListAuthzRepo{
				findErr:       testCase.findErr,
				projectMember: testCase.projectMember,
				member:        testCase.member,
			}
			service := tracker.NewTrackerService(repo, nil)

			links, err := service.ListThreadIssueLinks(ctx, projectID, channelID, actorID, "secret-thread")

			assert.Nil(t, links)
			requireAppError(t, err, fiber.StatusNotFound, "канал не найден")
			assert.Equal(t, testCase.expectedCalls, repo.calls)
			assert.Equal(t, testCase.expectedMemberCalls, repo.memberCalls, "concealed channel must not exceed its expected channel membership reads")
			assert.Zero(t, repo.listCalls, "concealed channel must not reach the link repository")
		})
	}
}

func TestListThreadIssueLinksProofErrorsStopBeforeLinkRead(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	channelID := uuid.New()
	actorID := uuid.New()

	for _, testCase := range []struct {
		name                string
		findErr             error
		projectMember       bool
		projectMemberErr    error
		memberErr           error
		expectedMemberCalls int
		expectedCalls       []string
	}{
		{
			name:          "channel proof infrastructure error",
			findErr:       errors.New("channel lookup unavailable"),
			expectedCalls: []string{"find_channel_in_project"},
		},
		{
			name:             "project membership proof infrastructure error",
			projectMemberErr: errors.New("project membership lookup unavailable"),
			expectedCalls:    []string{"find_channel_in_project", "is_project_member"},
		},
		{
			name:                "membership proof infrastructure error",
			projectMember:       true,
			memberErr:           errors.New("membership lookup unavailable"),
			expectedMemberCalls: 1,
			expectedCalls:       []string{"find_channel_in_project", "is_project_member", "is_channel_member"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &issueMessageLinkListAuthzRepo{
				findErr:          testCase.findErr,
				projectMember:    testCase.projectMember,
				projectMemberErr: testCase.projectMemberErr,
				member:           true,
				memberErr:        testCase.memberErr,
			}
			service := tracker.NewTrackerService(repo, nil)

			links, err := service.ListThreadIssueLinks(ctx, projectID, channelID, actorID, "thread-root")

			assert.Nil(t, links)
			requireAppError(t, err, fiber.StatusInternalServerError, "ошибка проверки доступа к каналу")
			assert.Equal(t, testCase.expectedCalls, repo.calls)
			assert.Equal(t, testCase.expectedMemberCalls, repo.memberCalls, "proof failure must not exceed expected channel membership reads")
			assert.Zero(t, repo.listCalls, "proof infrastructure errors must not reach the link repository")
		})
	}
}

func TestListThreadIssueLinksPreservesGenericLinkListFailureAfterAuthorization(t *testing.T) {
	repo := &issueMessageLinkListAuthzRepo{
		projectMember: true,
		member:        true,
		listErr:       errors.New("link store unavailable"),
	}
	service := tracker.NewTrackerService(repo, nil)

	links, err := service.ListThreadIssueLinks(context.Background(), uuid.New(), uuid.New(), uuid.New(), "thread-root")

	assert.Nil(t, links)
	requireAppError(t, err, fiber.StatusInternalServerError, "ошибка получения связей")
	assert.Equal(t, []string{"find_channel_in_project", "is_project_member", "is_channel_member", "list_thread_links"}, repo.calls)
	assert.Equal(t, 1, repo.listCalls)
}

func TestListThreadIssueLinksHandlerRequiresActorAndForwardsIt(t *testing.T) {
	projectID := uuid.New()
	channelID := uuid.New()
	actorID := uuid.New()
	messageID := "thread-root"
	repo := &issueMessageLinkListAuthzRepo{projectMember: true, member: true}
	handler := tracker.NewHandler(tracker.NewTrackerService(repo, nil), nil)
	path := "/projects/" + projectID.String() + "/message-links?channel_id=" + channelID.String() + "&message_id=" + messageID

	withoutActor := fiber.New()
	withoutActor.Get("/projects/:projectId/message-links", handler.ListThreadIssueLinks)
	response, err := withoutActor.Test(httptest.NewRequest(http.MethodGet, path, nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	assert.Empty(t, repo.calls, "missing actor must stop before any service proof or link read")
	response.Body.Close()

	withActor := fiber.New()
	withActor.Use(func(c fiber.Ctx) error {
		c.Locals("userID", actorID.String())
		return c.Next()
	})
	withActor.Get("/projects/:projectId/message-links", handler.ListThreadIssueLinks)
	response, err = withActor.Test(httptest.NewRequest(http.MethodGet, path, nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, []string{"find_channel_in_project", "is_project_member", "is_channel_member", "list_thread_links"}, repo.calls)
	assert.Equal(t, actorID, repo.memberActorID)
	response.Body.Close()
}

func issueMessageLinkWithProject(
	projectID uuid.UUID,
	channelID uuid.UUID,
	actorID uuid.UUID,
	messageID string,
	issueKey string,
) *ent.IssueMessageLink {
	return &ent.IssueMessageLink{
		ID:          uuid.New(),
		IssueID:     uuid.New(),
		ChannelID:   channelID,
		MessageID:   messageID,
		CreatedByID: actorID,
		Edges: ent.IssueMessageLinkEdges{
			Issue: &ent.Issue{
				ID:       uuid.New(),
				IssueKey: issueKey,
				Title:    "issue in requested project",
				Edges: ent.IssueEdges{
					Project: &ent.Project{ID: projectID},
				},
			},
		},
	}
}

func requireAppError(t *testing.T, err error, expectedCode int, expectedMessage string) {
	t.Helper()
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok, "expected AppError, got %T", err)
	assert.Equal(t, expectedCode, appErr.Code)
	assert.Equal(t, expectedMessage, appErr.Message)
}
