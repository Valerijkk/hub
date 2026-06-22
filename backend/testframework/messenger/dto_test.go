package messenger_test

// Тесты модуля мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

// TestToMessageResponse_MapsAllFields проверяет сценарий: to message response maps all fields.
func TestToMessageResponse_MapsAllFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	edited := now.Add(time.Minute)
	deleted := now.Add(2 * time.Minute)
	channelID := uuid.New()
	senderID := uuid.New()

	doc := &messenger.MessageDocument{
		ID:              "msg-1",
		ChannelID:       channelID,
		SenderID:        senderID,
		MessageType:     "poll",
		PollID:          "poll-1",
		ContentJSON:     map[string]interface{}{"text": "hello"},
		ContentText:     "hello",
		ThreadID:        "thread-1",
		ReplyCount:      3,
		ForwardedFromID: "orig-1",
		VoiceURL:        "https://example.com/voice",
		VoiceDurationMs: 1200,
		Attachments:     []messenger.AttachmentRef{{FileID: "f1", FileName: "a.png", FileSize: 100, MimeType: "image/png", URL: "/files/f1"}},
		Reactions:       []messenger.ReactionEntry{{Emoji: "thumbsup", UserIDs: []string{"u1"}}},
		Mentions:        []string{"u2"},
		MentionGroups:   []string{"all"},
		EditedAt:        &edited,
		DeletedAt:       &deleted,
		CreatedAt:       now,
	}

	resp := messenger.ExportToMessageResponse(doc)
	require.NotNil(t, resp)
	assert.Equal(t, doc.ID, resp.ID)
	assert.Equal(t, channelID.String(), resp.ChannelID)
	assert.Equal(t, senderID.String(), resp.SenderID)
	assert.Equal(t, "poll", resp.MessageType)
	assert.Equal(t, "poll-1", resp.PollID)
	assert.Equal(t, doc.ContentJSON, resp.ContentJSON)
	assert.Equal(t, "hello", resp.ContentText)
	assert.Equal(t, "thread-1", resp.ThreadID)
	assert.Equal(t, 3, resp.ReplyCount)
	assert.Equal(t, "orig-1", resp.ForwardedFromID)
	assert.Equal(t, doc.VoiceURL, resp.VoiceURL)
	assert.Equal(t, 1200, resp.VoiceDurationMs)
	assert.Len(t, resp.Attachments, 1)
	assert.Len(t, resp.Reactions, 1)
	assert.Equal(t, []string{"u2"}, resp.Mentions)
	assert.Equal(t, []string{"all"}, resp.MentionGroups)
	assert.Equal(t, edited, *resp.EditedAt)
	assert.Equal(t, deleted, *resp.DeletedAt)
	assert.Equal(t, now, resp.CreatedAt)
}

// TestToChannelResponse_WithAndWithoutProject проверяет сценарий: to channel response with and without project.
func TestToChannelResponse_WithAndWithoutProject(t *testing.T) {
	now := time.Now().UTC()
	projectID := uuid.New()
	userID := uuid.New()

	withProject := &ent.Channel{
		ID:          uuid.New(),
		Name:        "general",
		Description: "desc",
		IsReadonly:  true,
		CreatedBy:   userID,
		ProjectID:   &projectID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	resp := messenger.ExportToChannelResponse(withProject)
	assert.Equal(t, projectID.String(), resp.ProjectID)
	assert.Equal(t, "general", resp.Name)
	assert.True(t, resp.IsReadonly)

	withoutProject := &ent.Channel{
		ID:        uuid.New(),
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	resp2 := messenger.ExportToChannelResponse(withoutProject)
	assert.Empty(t, resp2.ProjectID)
}

// TestToChannelInviteResponse проверяет сценарий: to channel приглашение response.
func TestToChannelInviteResponse(t *testing.T) {
	now := time.Now().UTC()
	expires := now.Add(24 * time.Hour)
	invite := &ent.ChannelInvite{
		ID:        uuid.New(),
		ChannelID: uuid.New(),
		CreatorID: uuid.New(),
		Token:     "abc123",
		MaxUses:   10,
		UseCount:  2,
		ExpiresAt: &expires,
		IsRevoked: false,
		CreatedAt: now,
	}

	resp := messenger.ExportToChannelInviteResponse(invite)
	require.NotNil(t, resp)
	assert.Equal(t, invite.ID, resp.ID)
	assert.Equal(t, invite.ChannelID, resp.ChannelID)
	assert.Equal(t, "/invite/abc123", resp.InviteURL)
	assert.Equal(t, 10, resp.MaxUses)
	assert.Equal(t, 2, resp.UseCount)
	assert.Equal(t, expires, *resp.ExpiresAt)
}

// TestToPollResponse_AnonymousHidesVoters проверяет сценарий: to опрос response anonymous hides voters.
func TestToPollResponse_AnonymousHidesVoters(t *testing.T) {
	pollID := uuid.New()
	channelID := uuid.New()
	creatorID := uuid.New()
	optionID := uuid.New()
	voterID := uuid.New()
	otherVoter := uuid.New()

	p := &ent.Poll{
		ID:            pollID,
		ChannelID:     channelID,
		CreatorID:     creatorID,
		Question:      "Q?",
		IsAnonymous:   true,
		IsMultiChoice: false,
		IsClosed:      false,
		CreatedAt:     time.Now().UTC(),
	}
	options := []*ent.PollOption{{
		ID:       optionID,
		PollID:   pollID,
		Text:     "Yes",
		Position: 0,
	}}
	votes := []*ent.PollVote{
		{OptionID: optionID, UserID: voterID},
		{OptionID: optionID, UserID: otherVoter},
	}

	anon := messenger.ExportToPollResponse(p, options, votes, voterID, true)
	require.Len(t, anon.Options, 1)
	assert.True(t, anon.Options[0].Voted)
	assert.Empty(t, anon.Options[0].VoterIDs)
	assert.Equal(t, 2, anon.TotalVotes)

	public := messenger.ExportToPollResponse(p, options, votes, voterID, false)
	assert.Len(t, public.Options[0].VoterIDs, 2)
}
