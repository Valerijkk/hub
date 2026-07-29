package tracker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/user"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
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
		mentionIDs := make([]uuid.UUID, 0, len(input.Mentions))
		for _, raw := range input.Mentions {
			id, err := uuid.Parse(raw)
			if err != nil {
				return nil, fmt.Errorf("создание комментария: invalid mention id %q: %w", raw, err)
			}
			mentionIDs = append(mentionIDs, id)
		}
		create = create.SetMentions(mentionIDs)
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
			subj := fmt.Sprintf("project.%s.issue.comment_created", projectID)
			logOutboxPublishErr(ctx, subj, r.publisher.Publish(ctx, r.client, subj, CommentCreatedEvent{
				CommentID: c.ID.String(),
				IssueID:   issueID.String(),
				IssueKey:  iss.IssueKey,
				ProjectID: projectID,
				AuthorID:  authorID.String(),
				Mentions:  input.Mentions,
			}))
		}
	}

	// Перезагружаем с автором
	return r.client.Comment.Query().
		Where(comment.IDEQ(c.ID)).
		WithAuthor().
		WithReactions().
		Only(ctx)
}

// MaxIssueComments — P3-BE-06 / W1-A3-010: hard cap комментариев задачи.
// Mutation: снять Limit → unbounded All при N>cap.
const MaxIssueComments = 200

// ListComments возвращает комментарии к задаче с авторами и реакциями (cap MaxIssueComments).
// Soft-deleted (A8-006 / TZ 30-day trash) не попадают в список — реакции на них тоже скрыты.
func (r *EntRepository) ListComments(ctx context.Context, issueID uuid.UUID) ([]*ent.Comment, error) {
	return r.client.Comment.Query().
		Where(
			comment.HasIssueWith(issue.IDEQ(issueID)),
			comment.DeletedAtIsNil(),
		).
		WithAuthor().
		WithReactions(func(q *ent.ReactionQuery) {
			q.WithUser()
		}).
		Order(ent.Asc(comment.FieldCreatedAt)).
		Limit(MaxIssueComments).
		All(ctx)
}

// UpdateComment обновляет комментарий автора на задаче (A3-011).
func (r *EntRepository) UpdateComment(ctx context.Context, issueID, authorID, commentID uuid.UUID, input UpdateCommentInput) (*ent.Comment, error) {
	// TASK-095: VersionEQ + AddVersion — optimistic lock (NFR-24).
	affected, err := r.client.Comment.Update().
		Where(
			comment.IDEQ(commentID),
			comment.HasIssueWith(issue.IDEQ(issueID)),
			comment.HasAuthorWith(user.IDEQ(authorID)),
			comment.DeletedAtIsNil(),
			comment.VersionEQ(input.Version),
		).
		SetBodyJSON(input.BodyJSON).
		SetBodyText(input.BodyText).
		AddVersion(1).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление комментария: %w", err)
	}
	if affected == 0 {
		// A3W1-2-006: как UpdateIssue — alive+wrong version → 409; missing/soft-deleted/wrong author → 404.
		alive, existErr := r.client.Comment.Query().
			Where(
				comment.IDEQ(commentID),
				comment.HasIssueWith(issue.IDEQ(issueID)),
				comment.HasAuthorWith(user.IDEQ(authorID)),
				comment.DeletedAtIsNil(),
			).
			Exist(ctx)
		if existErr != nil {
			return nil, fmt.Errorf("проверка комментария после conflict: %w", existErr)
		}
		if !alive {
			return nil, apperrors.NewNotFound("комментарий не найден")
		}
		return nil, apperrors.NewConflict("version mismatch: комментарий был изменён другим пользователем")
	}
	return r.client.Comment.Query().
		Where(comment.IDEQ(commentID), comment.DeletedAtIsNil()).
		WithAuthor().
		Only(ctx)
}

// DeleteComment мягко удаляет комментарий автора (A8-006): SetDeletedAt, без hard Delete.
// Реакции не трогаем — они скрываются вместе с комментарием через ListComments DeletedAtIsNil.
func (r *EntRepository) DeleteComment(ctx context.Context, issueID, authorID, commentID uuid.UUID) error {
	affected, err := r.client.Comment.Update().
		Where(
			comment.IDEQ(commentID),
			comment.HasIssueWith(issue.IDEQ(issueID)),
			comment.HasAuthorWith(user.IDEQ(authorID)),
			comment.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("мягкое удаление комментария: %w", err)
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Реакции
// ---------------------------------------------------------------------------

// ToggleReaction добавляет или удаляет реакцию (toggle) на комментарий задачи (A3-011-react).
// Возвращает true если добавлена, false если удалена. Чужой issue → NotFound.
func (r *EntRepository) ToggleReaction(ctx context.Context, issueID, commentID, userID uuid.UUID, emoji string) (bool, error) {
	exists, err := r.client.Comment.Query().
		Where(
			comment.IDEQ(commentID),
			comment.HasIssueWith(issue.IDEQ(issueID)),
			comment.DeletedAtIsNil(),
		).
		Exist(ctx)
	if err != nil {
		return false, fmt.Errorf("проверка комментария: %w", err)
	}
	if !exists {
		return false, &ent.NotFoundError{}
	}

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
