package tracker_test

// Тесты сервиса трекера задач (волна 100): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

// TestLinkThreadToIssue_Success проверяет сценарий: link thread to issue успех.
func TestLinkThreadToIssue_Success(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	channelID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Linked", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	issue := repo.issues[uuid.MustParse(created.ID)]
	issue.Edges.Project = &ent.Project{ID: projectID}

	err = svc.LinkThreadToIssue(ctx, projectID, issue.ID, userID, channelID, "thread-root-1")
	require.NoError(t, err)
	require.Len(t, repo.issueMessageLinks, 1)
	assert.Equal(t, "thread-root-1", repo.issueMessageLinks[0].MessageID)
}

// TestLinkThreadToIssue_WrongProject проверяет отклонение при link thread to issue неверный project.
func TestLinkThreadToIssue_WrongProject(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	otherProject := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Wrong proj", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
	issue := repo.issues[uuid.MustParse(created.ID)]
	issue.Edges.Project = &ent.Project{ID: projectID}

	err = svc.LinkThreadToIssue(ctx, otherProject, issue.ID, userID, uuid.New(), "msg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in project")
}

// TestCreateIssueMessageLink_InvalidChannelID проверяет отклонение при создание issue message link некорректный channel id.
func TestCreateIssueMessageLink_InvalidChannelID(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Link me", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
	issue := repo.issues[uuid.MustParse(created.ID)]
	issue.Edges.Project = &ent.Project{ID: projectID}

	_, err = svc.CreateIssueMessageLink(ctx, projectID, created.IssueKey, userID, tracker.CreateIssueMessageLinkRequest{
		ChannelID: "bad", MessageID: "msg-1",
	})
	require.Error(t, err)
}

// TestListAndDeleteIssueMessageLinks проверяет сценарий: список and удаление issue message links.
func TestListAndDeleteIssueMessageLinks(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	channelID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Links", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
	issue := repo.issues[uuid.MustParse(created.ID)]
	issue.Edges.Project = &ent.Project{ID: projectID, Key: "TST"}

	linkResp, err := svc.CreateIssueMessageLink(ctx, projectID, created.IssueKey, userID, tracker.CreateIssueMessageLinkRequest{
		ChannelID: channelID.String(), MessageID: "root-msg",
	})
	require.NoError(t, err)

	list, err := svc.ListIssueMessageLinks(ctx, projectID, created.IssueKey)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, linkResp.ID, list[0].ID)

	threadLinks, err := svc.ListThreadIssueLinks(ctx, projectID, channelID, "root-msg")
	require.NoError(t, err)
	require.Len(t, threadLinks, 1)

	err = svc.DeleteIssueMessageLink(ctx, projectID, created.IssueKey, uuid.MustParse(linkResp.ID))
	require.NoError(t, err)
	list, err = svc.ListIssueMessageLinks(ctx, projectID, created.IssueKey)
	require.NoError(t, err)
	assert.Empty(t, list)
}
