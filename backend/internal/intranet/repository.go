package intranet

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/kudo"
	"github.com/valerii/hub/backend/ent/newspost"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository — доступ к данным intranet.
type Repository interface {
	CreateNewsPost(ctx context.Context, authorID uuid.UUID, req CreateNewsRequest) (*ent.NewsPost, error)
	GetNewsPost(ctx context.Context, id uuid.UUID) (*ent.NewsPost, error)
	ListNewsPosts(ctx context.Context, publishedOnly bool, page, perPage int) ([]*ent.NewsPost, int, error)
	UpdateNewsPost(ctx context.Context, id uuid.UUID, req UpdateNewsRequest) (*ent.NewsPost, error)
	SoftDeleteNewsPost(ctx context.Context, id uuid.UUID) error
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	ListActiveUsers(ctx context.Context) ([]*ent.User, error)
	GetActiveUser(ctx context.Context, id uuid.UUID) (*ent.User, error)
	CreateKudo(ctx context.Context, senderID, recipientID uuid.UUID, message string, isPublic bool) (*ent.Kudo, error)
	GetKudo(ctx context.Context, id uuid.UUID) (*ent.Kudo, error)
	ListPublicKudos(ctx context.Context, page, perPage int) ([]*ent.Kudo, int, error)
	ListReceivedKudos(ctx context.Context, recipientID uuid.UUID, page, perPage int) ([]*ent.Kudo, int, error)
	GetCoffeeEnrollment(ctx context.Context, userID uuid.UUID) (*ent.CoffeeEnrollment, error)
	UpsertCoffeeEnrollment(ctx context.Context, userID uuid.UUID, active bool) (*ent.CoffeeEnrollment, error)
	ListActiveCoffeeEnrollments(ctx context.Context) ([]*ent.User, error)
	CreateCoffeeRound(ctx context.Context, adminID uuid.UUID, pairCount, unpairedCount int) (*ent.CoffeeRound, error)
	CreateCoffeePair(ctx context.Context, roundID, userA, userB uuid.UUID) (*ent.CoffeePair, error)
	GetLatestCoffeeRound(ctx context.Context) (*ent.CoffeeRound, error)
	GetPreviousRoundPairKeys(ctx context.Context) (map[string]struct{}, error)
	ListCoffeeRounds(ctx context.Context, limit int) ([]*ent.CoffeeRound, error)
	FindUserPairInRound(ctx context.Context, roundID, userID uuid.UUID) (*ent.CoffeePair, error)
}

// EntRepository реализует Repository через Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий intranet.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreateNewsPost(ctx context.Context, authorID uuid.UUID, req CreateNewsRequest) (*ent.NewsPost, error) {
	return r.client.NewsPost.
		Create().
		SetTitle(req.Title).
		SetBody(req.Body).
		SetPinned(req.Pinned).
		SetPublished(req.Published).
		SetAuthorID(authorID).
		Save(ctx)
}

func (r *EntRepository) GetNewsPost(ctx context.Context, id uuid.UUID) (*ent.NewsPost, error) {
	return r.client.NewsPost.
		Query().
		Where(
			newspost.ID(id),
			newspost.DeletedAtIsNil(),
		).
		WithAuthor().
		Only(ctx)
}

func (r *EntRepository) ListNewsPosts(ctx context.Context, publishedOnly bool, page, perPage int) ([]*ent.NewsPost, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	q := r.client.NewsPost.Query().
		Where(newspost.DeletedAtIsNil())
	if publishedOnly {
		q = q.Where(newspost.Published(true))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	items, err := q.
		Order(
			ent.Desc(newspost.FieldPinned),
			ent.Desc(newspost.FieldCreatedAt),
		).
		Offset((page - 1) * perPage).
		Limit(perPage).
		WithAuthor().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *EntRepository) UpdateNewsPost(ctx context.Context, id uuid.UUID, req UpdateNewsRequest) (*ent.NewsPost, error) {
	upd := r.client.NewsPost.UpdateOneID(id).
		Where(newspost.DeletedAtIsNil()).
		AddVersion(1)
	if req.Title != nil {
		upd.SetTitle(*req.Title)
	}
	if req.Body != nil {
		upd.SetBody(*req.Body)
	}
	if req.Pinned != nil {
		upd.SetPinned(*req.Pinned)
	}
	if req.Published != nil {
		upd.SetPublished(*req.Published)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteNewsPost(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.NewsPost.UpdateOneID(id).
		Where(newspost.DeletedAtIsNil()).
		SetDeletedAt(now).
		Save(ctx)
	return err
}

func (r *EntRepository) IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	u, err := r.client.User.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.SystemRole == user.SystemRoleAdmin || u.SystemRole == user.SystemRoleFounder, nil
}

func (r *EntRepository) ListActiveUsers(ctx context.Context) ([]*ent.User, error) {
	return r.client.User.Query().
		Where(user.StatusEQ(user.StatusActive)).
		Order(
			ent.Asc(user.FieldDepartment),
			ent.Asc(user.FieldDisplayName),
			ent.Asc(user.FieldName),
			ent.Asc(user.FieldEmail),
		).
		All(ctx)
}

func (r *EntRepository) GetActiveUser(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return r.client.User.Query().
		Where(
			user.ID(id),
			user.StatusEQ(user.StatusActive),
		).
		Only(ctx)
}

func (r *EntRepository) CreateKudo(ctx context.Context, senderID, recipientID uuid.UUID, message string, isPublic bool) (*ent.Kudo, error) {
	return r.client.Kudo.
		Create().
		SetMessage(message).
		SetIsPublic(isPublic).
		SetSenderID(senderID).
		SetRecipientID(recipientID).
		Save(ctx)
}

func (r *EntRepository) GetKudo(ctx context.Context, id uuid.UUID) (*ent.Kudo, error) {
	return r.client.Kudo.Query().
		Where(kudo.ID(id)).
		WithSender().
		WithRecipient().
		Only(ctx)
}

func (r *EntRepository) ListPublicKudos(ctx context.Context, page, perPage int) ([]*ent.Kudo, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	q := r.client.Kudo.Query().Where(kudo.IsPublic(true))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := q.
		Order(ent.Desc(kudo.FieldCreatedAt)).
		Offset((page - 1) * perPage).
		Limit(perPage).
		WithSender().
		WithRecipient().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *EntRepository) ListReceivedKudos(ctx context.Context, recipientID uuid.UUID, page, perPage int) ([]*ent.Kudo, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	q := r.client.Kudo.Query().Where(kudo.HasRecipientWith(user.ID(recipientID)))
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := q.
		Order(ent.Desc(kudo.FieldCreatedAt)).
		Offset((page - 1) * perPage).
		Limit(perPage).
		WithSender().
		WithRecipient().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
