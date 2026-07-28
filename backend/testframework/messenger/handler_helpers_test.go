package messenger_test

// Тесты HTTP-обработчика мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channelmember"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TestSplitComma проверяет сценарий: split comma.
func TestSplitComma(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, messenger.ExportSplitComma("a, b ,c"))
	assert.Equal(t, []string{"solo"}, messenger.ExportSplitComma("solo"))
	assert.Empty(t, messenger.ExportSplitComma("  ,  , "))
}

// TestQueryInt проверяет сценарий: query int.
func TestQueryInt(t *testing.T) {
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })

	t.Run("default when missing", func(t *testing.T) {
		var got int
		app.Get("/test", func(c fiber.Ctx) error {
			got = messenger.ExportQueryInt(c, "limit", 25)
			return c.SendStatus(200)
		})
		req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 25, got)
	})

	t.Run("parses valid int", func(t *testing.T) {
		var got int
		app.Get("/limit", func(c fiber.Ctx) error {
			got = messenger.ExportQueryInt(c, "limit", 25)
			return c.SendStatus(200)
		})
		req := httptest.NewRequest(fiber.MethodGet, "/limit?limit=10", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 10, got)
	})

	t.Run("invalid falls back to default", func(t *testing.T) {
		var got int
		app.Get("/bad", func(c fiber.Ctx) error {
			got = messenger.ExportQueryInt(c, "limit", 25)
			return c.SendStatus(200)
		})
		req := httptest.NewRequest(fiber.MethodGet, "/bad?limit=abc", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 25, got)
	})

	t.Run("negative falls back to default", func(t *testing.T) {
		var got int
		app.Get("/neg", func(c fiber.Ctx) error {
			got = messenger.ExportQueryInt(c, "limit", 25)
			return c.SendStatus(200)
		})
		req := httptest.NewRequest(fiber.MethodGet, "/neg?limit=-5", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, 25, got)
	})
}

type mockIssuePromoter struct {
	issueKey string
	issueID  string
	err      error
	called   bool
}

func (m *mockIssuePromoter) CreateIssueFromPromotion(
	_ context.Context, _, _ uuid.UUID, _, _ string, _ map[string]interface{},
) (string, string, error) {
	m.called = true
	if m.err != nil {
		return "", "", m.err
	}
	return m.issueKey, m.issueID, nil
}

// TestPromoteToIssue_ValidationAndFlow проверяет сценарий: promote to issue валидация and поток.
func TestPromoteToIssue_ValidationAndFlow(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	projectID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	promoter := &mockIssuePromoter{issueKey: "HUB-99", issueID: uuid.New().String()}
	svc.SetIssuePromoter(promoter)
	svc.SetPromotePermissionChecker(&stubPromoteChecker{allowed: true})
	repo.displayNames[ownerID] = "Owner"

	doc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "promote this thread",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	t.Run("not configured", func(t *testing.T) {
		unconfigured, _, _ := newTestService()
		_, err := unconfigured.PromoteToIssue(ctx, doc.ID, ownerID, messenger.PromoteToIssueRequest{
			ProjectID: projectID.String(),
			Title:     "Task",
		})
		require.Error(t, err)
		appErr, ok := err.(*apperrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 500, appErr.Code)
	})

	t.Run("invalid project_id", func(t *testing.T) {
		_, err := svc.PromoteToIssue(ctx, doc.ID, ownerID, messenger.PromoteToIssueRequest{
			ProjectID: "bad",
			Title:     "Task",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "project_id")
	})

	t.Run("empty title", func(t *testing.T) {
		_, err := svc.PromoteToIssue(ctx, doc.ID, ownerID, messenger.PromoteToIssueRequest{
			ProjectID: projectID.String(),
			Title:     "   ",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("non-member forbidden", func(t *testing.T) {
		_, err := svc.PromoteToIssue(ctx, doc.ID, uuid.New(), messenger.PromoteToIssueRequest{
			ProjectID: projectID.String(),
			Title:     "Task",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "участником")
	})

	t.Run("nil promote checker forbidden", func(t *testing.T) {
		noChecker, noRepo, _ := newTestService()
		deniedPromoter := &mockIssuePromoter{issueKey: "HUB-0", issueID: uuid.New().String()}
		noChecker.SetIssuePromoter(deniedPromoter)
		// promoteCheck остаётся nil — fail-closed PrivEsc path.
		chID := setupChannelWithMembers(noRepo, "public", false, ownerID)
		msgID := uuid.New().String()
		noRepo.messages[msgID] = &messenger.MessageDocument{
			ID: msgID, ChannelID: chID, SenderID: ownerID,
			ContentText: "x", CreatedAt: time.Now(),
		}

		_, err := noChecker.PromoteToIssue(ctx, msgID, ownerID, messenger.PromoteToIssueRequest{
			ProjectID: projectID.String(),
			Title:     "Task",
		})
		require.Error(t, err)
		appErr, ok := err.(*apperrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 403, appErr.Code)
		assert.Contains(t, err.Error(), "проверка прав")
		assert.False(t, deniedPromoter.called)
	})

	t.Run("denied by promote checker", func(t *testing.T) {
		deniedSvc, deniedRepo, _ := newTestService()
		deniedPromoter := &mockIssuePromoter{issueKey: "HUB-0", issueID: uuid.New().String()}
		deniedSvc.SetIssuePromoter(deniedPromoter)
		deniedSvc.SetPromotePermissionChecker(&stubPromoteChecker{allowed: false})
		chID := setupChannelWithMembers(deniedRepo, "public", false, ownerID)
		msgID := uuid.New().String()
		deniedRepo.messages[msgID] = &messenger.MessageDocument{
			ID: msgID, ChannelID: chID, SenderID: ownerID,
			ContentText: "x", CreatedAt: time.Now(),
		}

		_, err := deniedSvc.PromoteToIssue(ctx, msgID, ownerID, messenger.PromoteToIssueRequest{
			ProjectID: projectID.String(),
			Title:     "Task",
		})
		require.Error(t, err)
		appErr, ok := err.(*apperrors.AppError)
		require.True(t, ok)
		assert.Equal(t, 403, appErr.Code)
		assert.Contains(t, err.Error(), "нет права создавать задачи")
		assert.False(t, deniedPromoter.called)
	})

	t.Run("success creates issue and backlink", func(t *testing.T) {
		promoter.called = false
		resp, err := svc.PromoteToIssue(ctx, doc.ID, ownerID, messenger.PromoteToIssueRequest{
			ProjectID: projectID.String(),
			Title:     "Promoted task",
			IssueType: "bug",
		})
		require.NoError(t, err)
		assert.Equal(t, "HUB-99", resp.IssueKey)
		assert.NotEmpty(t, resp.IssueID)
		assert.True(t, promoter.called)

		var backlinkFound bool
		for _, m := range repo.messages {
			if m.ThreadID == doc.ID && m.ContentText != doc.ContentText {
				backlinkFound = true
				assert.Contains(t, m.ContentText, "HUB-99")
			}
		}
		assert.True(t, backlinkFound)
	})
}

// TestVotePoll_ErrorPaths проверяет сценарий: vote опрос ошибка paths.
func TestVotePoll_ErrorPaths(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	pollID := uuid.New()
	opt1 := uuid.New()
	opt2 := uuid.New()
	repo.polls[pollID] = &ent.Poll{
		ID: pollID, ChannelID: channelID, CreatorID: ownerID,
		Question: "Q?", IsClosed: false, IsMultiChoice: false, CreatedAt: time.Now(),
	}
	repo.pollOptions[pollID] = []*ent.PollOption{
		{ID: opt1, PollID: pollID, Text: "A", Position: 0},
		{ID: opt2, PollID: pollID, Text: "B", Position: 1},
	}

	t.Run("closed poll", func(t *testing.T) {
		closedID := uuid.New()
		repo.polls[closedID] = &ent.Poll{
			ID: closedID, ChannelID: channelID, CreatorID: ownerID,
			Question: "Closed", IsClosed: true, CreatedAt: time.Now(),
		}
		repo.pollOptions[closedID] = []*ent.PollOption{{ID: uuid.New(), PollID: closedID, Text: "X"}}

		_, err := svc.VotePoll(ctx, closedID, ownerID, messenger.VotePollRequest{OptionIDs: []string{repo.pollOptions[closedID][0].ID.String()}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "закрыт")
	})

	t.Run("non-member", func(t *testing.T) {
		_, err := svc.VotePoll(ctx, pollID, outsiderID, messenger.VotePollRequest{OptionIDs: []string{opt1.String()}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "участником")
	})

	t.Run("invalid option id", func(t *testing.T) {
		_, err := svc.VotePoll(ctx, pollID, ownerID, messenger.VotePollRequest{OptionIDs: []string{"bad"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "некорректный")
	})

	t.Run("option from another poll", func(t *testing.T) {
		foreignOpt := uuid.New()
		_, err := svc.VotePoll(ctx, pollID, ownerID, messenger.VotePollRequest{OptionIDs: []string{foreignOpt.String()}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "не принадлежит")
	})

	t.Run("single choice only one option", func(t *testing.T) {
		_, err := svc.VotePoll(ctx, pollID, ownerID, messenger.VotePollRequest{
			OptionIDs: []string{opt1.String(), opt2.String()},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "один вариант")
	})
}

// TestClosePoll_ErrorPaths проверяет сценарий: close опрос ошибка paths.
func TestClosePoll_ErrorPaths(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	otherID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, otherID)

	pollID := uuid.New()
	repo.polls[pollID] = &ent.Poll{
		ID: pollID, ChannelID: channelID, CreatorID: ownerID,
		Question: "Close?", IsClosed: false, CreatedAt: time.Now(),
	}
	repo.pollOptions[pollID] = []*ent.PollOption{{ID: uuid.New(), PollID: pollID, Text: "A"}}

	t.Run("not creator", func(t *testing.T) {
		_, err := svc.ClosePoll(ctx, pollID, otherID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "создатель")
	})

	t.Run("already closed", func(t *testing.T) {
		closedID := uuid.New()
		repo.polls[closedID] = &ent.Poll{
			ID: closedID, ChannelID: channelID, CreatorID: ownerID,
			Question: "Done", IsClosed: true, CreatedAt: time.Now(),
		}
		_, err := svc.ClosePoll(ctx, closedID, ownerID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "уже закрыт")
	})
}

// TestListMessages_NonMemberForbidden проверяет отклонение при список сообщения non участник запрещённый.
func TestListMessages_NonMemberForbidden(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	_, err := svc.ListMessages(ctx, channelID, uuid.New(), nil, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "участником")
}

// TestResolveGroupMentions_OnlineRequiresPresence проверяет сценарий: resolve group упоминания online requires присутствие.
func TestResolveGroupMentions_OnlineRequiresPresence(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	mentions := messenger.ExportMessengerServiceResolveGroupMentions(svc, ctx, channelID, nil, []string{"online"}, nil, ownerID)
	assert.Empty(t, mentions)
}

// TestCreateInviteLink_RequiresAdmin проверяет сценарий: создание приглашение link requires админ.
func TestCreateInviteLink_RequiresAdmin(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "private", false, ownerID, memberID)

	_, err := svc.CreateInviteLink(ctx, channelID, memberID, messenger.CreateInviteLinkRequest{ExpiresIn: "1d"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner или admin")
}

// TestVotePoll_MultiChoiceAllowed проверяет сценарий: vote опрос multi choice allowed.
func TestVotePoll_MultiChoiceAllowed(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	pollID := uuid.New()
	opt1 := uuid.New()
	opt2 := uuid.New()
	repo.polls[pollID] = &ent.Poll{
		ID: pollID, ChannelID: channelID, CreatorID: ownerID,
		Question: "Multi", IsMultiChoice: true, CreatedAt: time.Now(),
	}
	repo.pollOptions[pollID] = []*ent.PollOption{
		{ID: opt1, PollID: pollID, Text: "A", Position: 0},
		{ID: opt2, PollID: pollID, Text: "B", Position: 1},
	}

	resp, err := svc.VotePoll(ctx, pollID, ownerID, messenger.VotePollRequest{
		OptionIDs: []string{opt1.String(), opt2.String()},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, resp.TotalVotes)
	assert.True(t, resp.Options[0].Voted)
	assert.True(t, resp.Options[1].Voted)
}

// TestAddMembers_AdminRoleAllowed проверяет сценарий: add members админ role allowed.
func TestAddMembers_AdminRoleAllowed(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	adminID := uuid.New()
	newUserID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)
	repo.members[channelID][adminID] = &ent.ChannelMember{
		ID: uuid.New(), ChannelID: channelID, UserID: adminID,
		Role: channelmember.RoleAdmin, JoinedAt: time.Now(),
	}

	require.NoError(t, svc.AddMembers(ctx, channelID, adminID, []string{newUserID.String()}))
	isMember, _ := repo.IsMember(ctx, channelID, newUserID)
	assert.True(t, isMember)
}
