package tracker_test

// Тесты сервиса трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

type stubThreadMirror struct {
	calls int
}

func (s *stubThreadMirror) MirrorIssueComment(_ context.Context, _, _ uuid.UUID, _, _ string) error {
	s.calls++
	return nil
}

// TestMirrorIssueCommentsToThreads_SkipsHubSyncPrefix проверяет сценарий: mirror issue comments to threads skips hub sync prefix.
func TestMirrorIssueCommentsToThreads_SkipsHubSyncPrefix(t *testing.T) {
	issueID := uuid.New()
	linkID := uuid.New()
	channelID := uuid.New()
	repo := &mockRepo{
		issueMessageLinks: []*ent.IssueMessageLink{
			{
				ID:             linkID,
				IssueID:        issueID,
				ChannelID:      channelID,
				MessageID:      "msg1",
				MirrorToThread: true,
			},
		},
	}
	mirror := &stubThreadMirror{}
	svc := tracker.NewTrackerService(repo, nil)
	svc.SetThreadMirror(mirror)

	tracker.ExportTrackerServiceMirrorIssueCommentsToThreads(svc, context.Background(), issueID, uuid.New(), "normal comment")
	assert.Equal(t, 1, mirror.calls)

	tracker.ExportTrackerServiceMirrorIssueCommentsToThreads(svc, context.Background(), issueID, uuid.New(), "[Hub sync] loop")
	assert.Equal(t, 1, mirror.calls)
}

// TestCreateIssueMessageLink_ValidatesProject проверяет сценарий: создание issue message link validates project.
func TestCreateIssueMessageLink_ValidatesProject(t *testing.T) {
	projectID := uuid.New()
	otherProject := uuid.New()
	issue := &ent.Issue{
		ID:       uuid.New(),
		IssueKey: "HUB-1",
		Title:    "Test",
		Edges: ent.IssueEdges{
			Project: &ent.Project{ID: otherProject},
		},
	}
	repo := &mockRepo{issuesByKey: map[string]*ent.Issue{"HUB-1": issue}}
	svc := tracker.NewTrackerService(repo, nil)

	_, err := svc.CreateIssueMessageLink(context.Background(), projectID, "HUB-1", uuid.New(), tracker.CreateIssueMessageLinkRequest{
		ChannelID: uuid.New().String(),
		MessageID: "abc",
	})
	require.Error(t, err)
}

// TestCreateIssueMessageLink_Success проверяет сценарий: создание issue message link успех.
func TestCreateIssueMessageLink_Success(t *testing.T) {
	projectID := uuid.New()
	issueID := uuid.New()
	channelID := uuid.New()
	userID := uuid.New()
	issue := &ent.Issue{
		ID:       issueID,
		IssueKey: "HUB-10",
		Title:    "Test",
		Edges: ent.IssueEdges{
			Project: &ent.Project{ID: projectID},
		},
	}
	repo := &mockRepo{
		issuesByKey: map[string]*ent.Issue{"HUB-10": issue},
	}
	svc := tracker.NewTrackerService(repo, nil)

	resp, err := svc.CreateIssueMessageLink(context.Background(), projectID, "HUB-10", userID, tracker.CreateIssueMessageLinkRequest{
		ChannelID: channelID.String(),
		MessageID: "thread-root",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "thread-root", resp.MessageID)
	assert.Len(t, repo.issueMessageLinks, 1)
}

// TestListIssueMessageLinks_Success проверяет сценарий: список issue message links успех.
func TestListIssueMessageLinks_Success(t *testing.T) {
	projectID := uuid.New()
	issueID := uuid.New()
	issue := &ent.Issue{
		ID: issueID, IssueKey: "HUB-11", Title: "List links",
		Edges: ent.IssueEdges{Project: &ent.Project{ID: projectID}},
	}
	linkID := uuid.New()
	repo := &mockRepo{
		issuesByKey: map[string]*ent.Issue{"HUB-11": issue},
		issueMessageLinks: []*ent.IssueMessageLink{{
			ID: linkID, IssueID: issueID, ChannelID: uuid.New(),
			MessageID: "m1", MirrorToThread: true, CreatedByID: uuid.New(),
		}},
	}
	svc := tracker.NewTrackerService(repo, nil)

	links, err := svc.ListIssueMessageLinks(context.Background(), projectID, "HUB-11")
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, linkID.String(), links[0].ID)
}

// TestDeleteIssueMessageLink_Success проверяет сценарий: удаление issue message link успех.
func TestDeleteIssueMessageLink_Success(t *testing.T) {
	projectID := uuid.New()
	issueID := uuid.New()
	linkID := uuid.New()
	issue := &ent.Issue{
		ID: issueID, IssueKey: "HUB-12",
		Edges: ent.IssueEdges{Project: &ent.Project{ID: projectID}},
	}
	repo := &mockRepo{
		issuesByKey: map[string]*ent.Issue{"HUB-12": issue},
		issueMessageLinks: []*ent.IssueMessageLink{{
			ID: linkID, IssueID: issueID, ChannelID: uuid.New(), MessageID: "m2",
		}},
	}
	svc := tracker.NewTrackerService(repo, nil)

	require.NoError(t, svc.DeleteIssueMessageLink(context.Background(), projectID, "HUB-12", linkID))
	assert.Empty(t, repo.issueMessageLinks)
}
