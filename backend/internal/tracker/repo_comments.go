package tracker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/user"
)

// ---------------------------------------------------------------------------
// Комментарии
// ---------------------------------------------------------------------------

// CreateComment добавляет комментарий к задаче.
// Публикует outbox-событие project.{projectID}.issue.comment_created.
func (r *EntRepository) CreateComment(ctx context.Context, issueID, authorID uuid.UUID, input CreateCommentInput) (*ent.Comment, error) {
	create := r.client.Comment.Create().
		SetIssueID(issueID).
		SetAuthorID(authorID).
		SetBodyJSON(input.BodyJSON).
		SetBodyText(input.BodyText).
		SetIsInternal(input.IsInternal)

	if len(input.Mentions) > 0 {
		create = create.SetMentions(input.Mentions)
	}

	c, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание комментария: %w", err)
	}

	// Публикуем outbox-событие создания комментария.
	if r.publisher != nil {
		iss, err := r.client.Issue.Query().
			Where(issue.IDEQ(issueID)).
			WithProject().
			Only(ctx)
		if err == nil && iss.Edges.Project != nil {
			projectID := iss.Edges.Project.ID.String()
			_ = r.publisher.Publish(ctx, r.client, fmt.Sprintf("project.%s.issue.comment_created", projectID), CommentCreatedEvent{
				CommentID: c.ID.String(),
				IssueID:   issueID.String(),
				IssueKey:  iss.IssueKey,
				ProjectID: projectID,
				AuthorID:  authorID.String(),
				Mentions:  input.Mentions,
			})
		}
	}

	// Перезагружаем с автором
	return r.client.Comment.Query().
		Where(comment.IDEQ(c.ID)).
		WithAuthor().
		WithReactions().
		Only(ctx)
}

// ListComments возвращает все комментарии к задаче с авторами и реакциями.
func (r *EntRepository) ListComments(ctx context.Context, issueID uuid.UUID) ([]*ent.Comment, error) {
	return r.client.Comment.Query().
		Where(comment.HasIssueWith(issue.IDEQ(issueID))).
		WithAuthor().
		WithReactions(func(q *ent.ReactionQuery) {
			q.WithUser()
		}).
		Order(ent.Asc(comment.FieldCreatedAt)).
		All(ctx)
}

// UpdateComment обновляет комментарий автора на задаче (A3-011).
func (r *EntRepository) UpdateComment(ctx context.Context, issueID, authorID, commentID uuid.UUID, input UpdateCommentInput) (*ent.Comment, error) {
	affected, err := r.client.Comment.Update().
		Where(
			comment.IDEQ(commentID),
			comment.HasIssueWith(issue.IDEQ(issueID)),
			comment.HasAuthorWith(user.IDEQ(authorID)),
		).
		SetBodyJSON(input.BodyJSON).
		SetBodyText(input.BodyText).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление комментария: %w", err)
	}
	if affected == 0 {
		return nil, &ent.NotFoundError{}
	}
	return r.client.Comment.Query().
		Where(comment.IDEQ(commentID)).
		WithAuthor().
		Only(ctx)
}

// DeleteComment удаляет комментарий автора на задаче (A3-011).
func (r *EntRepository) DeleteComment(ctx context.Context, issueID, authorID, commentID uuid.UUID) error {
	affected, err := r.client.Comment.Delete().
		Where(
			comment.IDEQ(commentID),
			comment.HasIssueWith(issue.IDEQ(issueID)),
			comment.HasAuthorWith(user.IDEQ(authorID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Реакции
// ---------------------------------------------------------------------------

// ToggleReaction добавляет или удаляет реакцию (toggle). Возвращает true если добавлена, false если удалена.
func (r *EntRepository) ToggleReaction(ctx context.Context, commentID, userID uuid.UUID, emoji string) (bool, error) {
	// Ищем существующую реакцию
	existing, err := r.client.Reaction.Query().
		Where(
			reaction.HasCommentWith(comment.IDEQ(commentID)),
			reaction.HasUserWith(user.IDEQ(userID)),
			reaction.EmojiEQ(emoji),
		).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return false, fmt.Errorf("поиск реакции: %w", err)
	}

	if existing != nil {
		// Удаляем (toggle off)
		if err := r.client.Reaction.DeleteOneID(existing.ID).Exec(ctx); err != nil {
			return false, fmt.Errorf("удаление реакции: %w", err)
		}
		return false, nil
	}

	// Создаём (toggle on)
	_, err = r.client.Reaction.Create().
		SetCommentID(commentID).
		SetUserID(userID).
		SetEmoji(emoji).
		Save(ctx)
	if err != nil {
		return false, fmt.Errorf("создание реакции: %w", err)
	}
	return true, nil
}
