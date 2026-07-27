package marketplace

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/marketplacelisting"
	"github.com/valerii/hub/backend/ent/user"
)

// Repository — доступ к данным marketplace.
type Repository interface {
	CreateListing(ctx context.Context, authorID uuid.UUID, req CreateListingRequest) (*ent.MarketplaceListing, error)
	GetListing(ctx context.Context, id uuid.UUID) (*ent.MarketplaceListing, error)
	ListApprovedListings(ctx context.Context, category string, page, perPage int) ([]*ent.MarketplaceListing, int, error)
	ListListingsByAuthor(ctx context.Context, authorID uuid.UUID) ([]*ent.MarketplaceListing, error)
	ListPendingListings(ctx context.Context) ([]*ent.MarketplaceListing, error)
	UpdateListing(ctx context.Context, id uuid.UUID, req UpdateListingRequest) (*ent.MarketplaceListing, error)
	SetListingStatus(ctx context.Context, id uuid.UUID, status marketplacelisting.Status, reviewerID *uuid.UUID, reviewNote string, expectedVersion int) (*ent.MarketplaceListing, error)
	SoftDeleteListing(ctx context.Context, id uuid.UUID) error
	IsSystemAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	ListAdminUserIDs(ctx context.Context) ([]uuid.UUID, error)
}

// EntRepository реализует Repository через Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт Ent-репозиторий marketplace.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

func (r *EntRepository) CreateListing(ctx context.Context, authorID uuid.UUID, req CreateListingRequest) (*ent.MarketplaceListing, error) {
	b := r.client.MarketplaceListing.Create().
		SetTitle(req.Title).
		SetDescription(req.Description).
		SetCategory(marketplacelisting.Category(req.Category)).
		SetAuthorID(authorID)
	if req.PriceCents != nil {
		b.SetPriceCents(*req.PriceCents)
	}
	if req.Currency != "" {
		b.SetCurrency(req.Currency)
	}
	return b.Save(ctx)
}

func (r *EntRepository) GetListing(ctx context.Context, id uuid.UUID) (*ent.MarketplaceListing, error) {
	return r.client.MarketplaceListing.Query().
		Where(
			marketplacelisting.ID(id),
			marketplacelisting.DeletedAtIsNil(),
		).
		WithAuthor().
		WithReviewer().
		Only(ctx)
}

func (r *EntRepository) ListApprovedListings(ctx context.Context, category string, page, perPage int) ([]*ent.MarketplaceListing, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}
	q := r.client.MarketplaceListing.Query().
		Where(
			marketplacelisting.StatusEQ(marketplacelisting.StatusApproved),
			marketplacelisting.DeletedAtIsNil(),
		)
	if category != "" {
		q = q.Where(marketplacelisting.CategoryEQ(marketplacelisting.Category(category)))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, err := q.
		Order(ent.Desc(marketplacelisting.FieldCreatedAt)).
		Offset((page - 1) * perPage).
		Limit(perPage).
		WithAuthor().
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *EntRepository) ListListingsByAuthor(ctx context.Context, authorID uuid.UUID) ([]*ent.MarketplaceListing, error) {
	return r.client.MarketplaceListing.Query().
		Where(
			marketplacelisting.HasAuthorWith(user.ID(authorID)),
			marketplacelisting.DeletedAtIsNil(),
		).
		Order(ent.Desc(marketplacelisting.FieldCreatedAt)).
		WithAuthor().
		WithReviewer().
		All(ctx)
}

func (r *EntRepository) ListPendingListings(ctx context.Context) ([]*ent.MarketplaceListing, error) {
	return r.client.MarketplaceListing.Query().
		Where(
			marketplacelisting.StatusEQ(marketplacelisting.StatusPending),
			marketplacelisting.DeletedAtIsNil(),
		).
		Order(ent.Desc(marketplacelisting.FieldCreatedAt)).
		WithAuthor().
		All(ctx)
}

func (r *EntRepository) UpdateListing(ctx context.Context, id uuid.UUID, req UpdateListingRequest) (*ent.MarketplaceListing, error) {
	// A6-003: Version всегда в WHERE — version:0 больше не обходит optimistic lock.
	upd := r.client.MarketplaceListing.UpdateOneID(id).
		Where(
			marketplacelisting.DeletedAtIsNil(),
			marketplacelisting.VersionEQ(req.Version),
		).
		AddVersion(1)
	if req.Title != nil {
		upd.SetTitle(*req.Title)
	}
	if req.Description != nil {
		upd.SetDescription(*req.Description)
	}
	if req.Category != nil {
		upd.SetCategory(marketplacelisting.Category(*req.Category))
	}
	if req.PriceCents != nil {
		upd.SetPriceCents(*req.PriceCents)
	}
	if req.Currency != nil {
		upd.SetCurrency(*req.Currency)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SetListingStatus(ctx context.Context, id uuid.UUID, status marketplacelisting.Status, reviewerID *uuid.UUID, reviewNote string, expectedVersion int) (*ent.MarketplaceListing, error) {
	now := time.Now().UTC()
	upd := r.client.MarketplaceListing.UpdateOneID(id).
		Where(
			marketplacelisting.DeletedAtIsNil(),
			marketplacelisting.Version(expectedVersion),
		).
		SetStatus(status).
		SetReviewedAt(now).
		AddVersion(1)
	if reviewerID != nil {
		upd.SetReviewerID(*reviewerID)
	}
	if reviewNote != "" {
		upd.SetReviewNote(reviewNote)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) SoftDeleteListing(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.MarketplaceListing.UpdateOneID(id).
		Where(marketplacelisting.DeletedAtIsNil()).
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

func (r *EntRepository) ListAdminUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	users, err := r.client.User.Query().
		Where(
			user.StatusEQ(user.StatusActive),
			user.Or(
				user.SystemRoleEQ(user.SystemRoleAdmin),
				user.SystemRoleEQ(user.SystemRoleFounder),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids, nil
}
