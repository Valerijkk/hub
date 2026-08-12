package tracker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/issuemessagelink"
)

func (r *EntRepository) CreateIssueMessageLink(
	ctx context.Context,
	issueID, channelID, createdBy uuid.UUID,
	messageID string,
	mirrorToThread bool,
) (*ent.IssueMessageLink, error) {
	return r.client.IssueMessageLink.Create().
		SetIssueID(issueID).
		SetChannelID(channelID).
		SetMessageID(messageID).
		SetCreatedByID(createdBy).
		SetMirrorToThread(mirrorToThread).
		Save(ctx)
}

func (r *EntRepository) ListIssueMessageLinksByIssue(ctx context.Context, issueID uuid.UUID) ([]*ent.IssueMessageLink, error) {
	return r.client.IssueMessageLink.Query().
		Where(issuemessagelink.IssueIDEQ(issueID)).
		Order(ent.Desc(issuemessagelink.FieldCreatedAt)).
		All(ctx)
}

func (r *EntRepository) ListIssueMessageLinksByThread(
	ctx context.Context,
	channelID uuid.UUID,
	messageID string,
) ([]*ent.IssueMessageLink, error) {
	// A3W1-2-004: eager Issue+Project — без N+1 FindIssueByID в ListThreadIssueLinks.
	return r.client.IssueMessageLink.Query().
		Where(
			issuemessagelink.ChannelIDEQ(channelID),
			issuemessagelink.MessageIDEQ(messageID),
		).
		WithIssue(func(q *ent.IssueQuery) {
			q.WithProject()
		}).
		Order(ent.Desc(issuemessagelink.FieldCreatedAt)).
		All(ctx)
}

func (r *EntRepository) FindIssueMessageLink(ctx context.Context, linkID uuid.UUID) (*ent.IssueMessageLink, error) {
	return r.client.IssueMessageLink.Query().
		Where(issuemessagelink.IDEQ(linkID)).
		Only(ctx)
}

func (r *EntRepository) DeleteIssueMessageLink(ctx context.Context, linkID, issueID uuid.UUID) error {
	n, err := r.client.IssueMessageLink.Delete().
		Where(
			issuemessagelink.IDEQ(linkID),
			issuemessagelink.IssueIDEQ(issueID),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("issue message link not found")
	}
	return nil
}

// FindChannelInProject возвращает канал только если он привязан к projectID (P1-BE-04).
func (r *EntRepository) FindChannelInProject(ctx context.Context, projectID, channelID uuid.UUID) (*ent.Channel, error) {
	return r.client.Channel.Query().
		Where(
			channel.IDEQ(channelID),
			channel.DeletedAtIsNil(),
			channel.ProjectIDEQ(projectID),
		).
		Only(ctx)
}

// IsChannelMember проверяет, состоит ли пользователь в канале (P1-BE-04).
func (r *EntRepository) IsChannelMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	return r.client.ChannelMember.Query().
		Where(
			channelmember.ChannelIDEQ(channelID),
			channelmember.UserIDEQ(userID),
			channelmember.HasChannelWith(channel.DeletedAtIsNil()),
		).
		Exist(ctx)
}
