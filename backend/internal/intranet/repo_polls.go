package intranet

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/newspoll"
	"github.com/valerii/hub/backend/ent/newspolloption"
	"github.com/valerii/hub/backend/ent/newspollvote"
	"github.com/valerii/hub/backend/ent/newspost"
)

// CreateNewsPollInput — опрос при создании новости (FR-INTR-01).
type CreateNewsPollInput struct {
	Question string
	Options  []string
}

// UpdateNewsPollInput — правка опроса (FR-INTR-03).
type UpdateNewsPollInput struct {
	Question *string
	Options  []string // nil = не трогать; иначе полная замена (только без голосов)
	Closed   *bool
	Version  int
}

func (r *EntRepository) CreateNewsPost(ctx context.Context, authorID uuid.UUID, req CreateNewsRequest) (*ent.NewsPost, error) {
	if req.Poll == nil {
		return r.client.NewsPost.
			Create().
			SetTitle(req.Title).
			SetBody(req.Body).
			SetPinned(req.Pinned).
			SetPublished(req.Published).
			SetAuthorID(authorID).
			Save(ctx)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create news+poll tx: %w", err)
	}
	post, err := tx.NewsPost.
		Create().
		SetTitle(req.Title).
		SetBody(req.Body).
		SetPinned(req.Pinned).
		SetPublished(req.Published).
		SetAuthorID(authorID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	poll, err := tx.NewsPoll.Create().
		SetNewsPostID(post.ID).
		SetQuestion(strings.TrimSpace(req.Poll.Question)).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("create news poll: %w", err)
	}
	for i, text := range req.Poll.Options {
		_, err = tx.NewsPollOption.Create().
			SetPollID(poll.ID).
			SetText(strings.TrimSpace(text)).
			SetPosition(i).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("create news poll option: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create news+poll: %w", err)
	}
	return post, nil
}

func (r *EntRepository) SoftDeleteNewsPost(ctx context.Context, id, actorID uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin news soft-delete tx: %w", err)
	}
	now := time.Now().UTC()
	hadPoll, err := tx.NewsPoll.Query().
		Where(newspoll.NewsPostIDEQ(id)).
		Exist(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("check news poll: %w", err)
	}
	_, err = tx.NewsPost.UpdateOneID(id).
		Where(newspost.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.AuditLog.Create().
		SetAction("delete").
		SetEntityType("news_post").
		SetEntityID(id.String()).
		SetUserID(actorID).
		SetDetails(map[string]interface{}{
			"actor_id": actorID.String(),
			"post_id":  id.String(),
			"had_poll": hadPoll,
		}).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("audit news delete: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit news soft-delete: %w", err)
	}
	return nil
}

// GetNewsPollByPostID возвращает опрос поста (пост не в корзине — FR-INTR-07 hide).
func (r *EntRepository) GetNewsPollByPostID(ctx context.Context, newsPostID uuid.UUID) (*ent.NewsPoll, error) {
	return r.client.NewsPoll.Query().
		Where(
			newspoll.NewsPostIDEQ(newsPostID),
			newspoll.HasNewsPostWith(newspost.DeletedAtIsNil()),
		).
		WithOptions(func(q *ent.NewsPollOptionQuery) {
			q.Order(ent.Asc(newspolloption.FieldPosition))
		}).
		Only(ctx)
}

// GetNewsPollByPostIDIncludingDeleted — для restore IT (пост может быть в корзине).
func (r *EntRepository) GetNewsPollByPostIDIncludingDeleted(ctx context.Context, newsPostID uuid.UUID) (*ent.NewsPoll, error) {
	return r.client.NewsPoll.Query().
		Where(newspoll.NewsPostIDEQ(newsPostID)).
		WithOptions(func(q *ent.NewsPollOptionQuery) {
			q.Order(ent.Asc(newspolloption.FieldPosition))
		}).
		Only(ctx)
}

func (r *EntRepository) GetUserNewsPollVote(ctx context.Context, pollID, userID uuid.UUID) (*ent.NewsPollVote, error) {
	return r.client.NewsPollVote.Query().
		Where(
			newspollvote.PollIDEQ(pollID),
			newspollvote.UserIDEQ(userID),
		).
		Only(ctx)
}

func (r *EntRepository) CountNewsPollVotes(ctx context.Context, pollID uuid.UUID) (int, error) {
	return r.client.NewsPollVote.Query().
		Where(newspollvote.PollIDEQ(pollID)).
		Count(ctx)
}

func (r *EntRepository) VoteNewsPoll(ctx context.Context, pollID, optionID, userID uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin news poll vote tx: %w", err)
	}
	exists, err := tx.NewsPollVote.Query().
		Where(newspollvote.PollIDEQ(pollID), newspollvote.UserIDEQ(userID)).
		Exist(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if exists {
		_ = tx.Rollback()
		return ErrAlreadyVoted
	}
	optOK, err := tx.NewsPollOption.Query().
		Where(newspolloption.IDEQ(optionID), newspolloption.PollIDEQ(pollID)).
		Exist(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if !optOK {
		_ = tx.Rollback()
		return ErrInvalidInput
	}
	_, err = tx.NewsPollVote.Create().
		SetPollID(pollID).
		SetOptionID(optionID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		if ent.IsConstraintError(err) {
			return ErrAlreadyVoted
		}
		return fmt.Errorf("insert news poll vote: %w", err)
	}
	_, err = tx.NewsPollOption.UpdateOneID(optionID).
		AddVotesCount(1).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("inc news poll votes_count: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit news poll vote: %w", err)
	}
	return nil
}

func (r *EntRepository) UpdateNewsPoll(ctx context.Context, pollID uuid.UUID, in UpdateNewsPollInput) (*ent.NewsPoll, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update news poll tx: %w", err)
	}
	upd := tx.NewsPoll.UpdateOneID(pollID).AddVersion(1)
	if in.Question != nil {
		upd.SetQuestion(strings.TrimSpace(*in.Question))
	}
	if in.Closed != nil {
		upd.SetClosed(*in.Closed)
	}
	poll, err := upd.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if in.Options != nil {
		_, err = tx.NewsPollOption.Delete().Where(newspolloption.PollIDEQ(pollID)).Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("replace news poll options: %w", err)
		}
		for i, text := range in.Options {
			_, err = tx.NewsPollOption.Create().
				SetPollID(pollID).
				SetText(strings.TrimSpace(text)).
				SetPosition(i).
				Save(ctx)
			if err != nil {
				_ = tx.Rollback()
				return nil, fmt.Errorf("create replaced poll option: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update news poll: %w", err)
	}
	return r.GetNewsPollByPostID(ctx, poll.NewsPostID)
}
