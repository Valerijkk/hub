package marketplace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/marketplacelisting"
)

var (
	ErrNotFound        = errors.New("listing not found")
	ErrForbidden       = errors.New("admin access required")
	ErrInvalidInput    = errors.New("invalid listing input")
	ErrVersionConflict = errors.New("listing version conflict")
	ErrNotOwner        = errors.New("not listing owner")
	ErrInvalidStatus   = errors.New("invalid listing status for operation")
)

// Notifier — уведомления mod queue / author.
type Notifier interface {
	NotifyListingSubmitted(ctx context.Context, adminIDs []uuid.UUID, listing ListingResponse, authorID uuid.UUID) error
	NotifyListingDecision(ctx context.Context, authorID uuid.UUID, approved bool, listing ListingResponse, reviewerID uuid.UUID) error
}

// ChannelBroadcaster — post в #marketplace при approve.
type ChannelBroadcaster interface {
	BroadcastApprovedListing(ctx context.Context, listing ListingResponse) error
}

// Service — бизнес-логика marketplace.
type Service struct {
	repo         Repository
	notifier     Notifier
	broadcaster  ChannelBroadcaster
}

// NewService создаёт сервис marketplace.
func NewService(repo Repository, notifier Notifier, broadcaster ChannelBroadcaster) *Service {
	return &Service{repo: repo, notifier: notifier, broadcaster: broadcaster}
}

func (s *Service) ListApproved(ctx context.Context, category string, page, perPage int) (*ListingListResponse, error) {
	// P3-BE-09 / W7-A5-010: невалидная category → 400, не Ent 500.
	if category != "" && !isValidCategory(ListingCategory(category)) {
		return nil, ErrInvalidInput
	}
	items, total, err := s.repo.ListApprovedListings(ctx, category, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list approved listings: %w", err)
	}
	page, perPage = normalizePage(page, perPage)
	resp := make([]ListingResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toListingResponse(item))
	}
	return &ListingListResponse{
		Items:   resp,
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasMore: page*perPage < total,
	}, nil
}

func (s *Service) ListMine(ctx context.Context, userID uuid.UUID, page, perPage int) (*ListingListResponse, error) {
	items, total, err := s.repo.ListListingsByAuthor(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list my listings: %w", err)
	}
	page, perPage = normalizePage(page, perPage)
	out := make([]ListingResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toListingResponsePrivate(item))
	}
	return &ListingListResponse{
		Items:   out,
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasMore: page*perPage < total,
	}, nil
}

func (s *Service) ListPending(ctx context.Context, userID uuid.UUID, page, perPage int) (*ListingListResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	items, total, err := s.repo.ListPendingListings(ctx, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list pending listings: %w", err)
	}
	page, perPage = normalizePage(page, perPage)
	out := make([]ListingResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toListingResponsePrivate(item))
	}
	return &ListingListResponse{
		Items:   out,
		Total:   total,
		Page:    page,
		PerPage: perPage,
		HasMore: page*perPage < total,
	}, nil
}

func (s *Service) CreateListing(ctx context.Context, userID uuid.UUID, req CreateListingRequest) (*ListingResponse, error) {
	if err := validateListingInput(req.Title, req.Description, req.Category, req.PriceCents, req.Currency); err != nil {
		return nil, err
	}
	l, err := s.repo.CreateListing(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("create listing: %w", err)
	}
	l, err = s.repo.GetListing(ctx, l.ID)
	if err != nil {
		return nil, fmt.Errorf("reload listing: %w", err)
	}
	resp := toListingResponsePrivate(l)
	if s.notifier != nil {
		adminIDs, aErr := s.repo.ListAdminUserIDs(ctx)
		if aErr == nil {
			_ = s.notifier.NotifyListingSubmitted(ctx, adminIDs, resp, userID)
		}
	}
	return &resp, nil
}

func (s *Service) UpdateListing(ctx context.Context, userID, id uuid.UUID, req UpdateListingRequest) (*ListingResponse, error) {
	current, err := s.repo.GetListing(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if current.Edges.Author == nil || current.Edges.Author.ID != userID {
		return nil, ErrNotOwner
	}
	if current.Status != marketplacelisting.StatusPending && current.Status != marketplacelisting.StatusApproved {
		return nil, ErrInvalidStatus
	}
	if req.Version < 1 {
		return nil, ErrVersionConflict
	}
	title := current.Title
	desc := current.Description
	cat := ListingCategory(current.Category)
	price := current.PriceCents
	currency := current.Currency
	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		desc = *req.Description
	}
	if req.Category != nil {
		cat = *req.Category
	}
	if req.PriceCents != nil {
		price = req.PriceCents
	}
	if req.Currency != nil {
		currency = *req.Currency
	}
	if err := validateListingInput(title, desc, cat, price, currency); err != nil {
		return nil, err
	}

	var l *ent.MarketplaceListing
	if current.Status == marketplacelisting.StatusApproved {
		// W7-A5-004: один UPDATE — поля + pending (без окна approved+новый контент).
		l, err = s.repo.UpdateApprovedListingToPending(ctx, id, req)
	} else {
		l, err = s.repo.UpdateListing(ctx, id, req)
	}
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("update listing: %w", err)
	}
	l, err = s.repo.GetListing(ctx, l.ID)
	if err != nil {
		return nil, fmt.Errorf("reload listing: %w", err)
	}
	resp := toListingResponsePrivate(l)
	return &resp, nil
}

func (s *Service) ApproveListing(ctx context.Context, reviewerID, id uuid.UUID, req ReviewListingRequest) (*ListingResponse, error) {
	if err := s.requireAdmin(ctx, reviewerID); err != nil {
		return nil, err
	}
	l, err := s.transitionReview(ctx, id, req, marketplacelisting.StatusApproved, reviewerID)
	if err != nil {
		return nil, err
	}
	resp := toListingResponse(l)
	if s.notifier != nil && l.Edges.Author != nil {
		_ = s.notifier.NotifyListingDecision(ctx, l.Edges.Author.ID, true, resp, reviewerID)
	}
	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastApprovedListing(ctx, resp)
	}
	return &resp, nil
}

func (s *Service) RejectListing(ctx context.Context, reviewerID, id uuid.UUID, req ReviewListingRequest) (*ListingResponse, error) {
	if err := s.requireAdmin(ctx, reviewerID); err != nil {
		return nil, err
	}
	l, err := s.transitionReview(ctx, id, req, marketplacelisting.StatusRejected, reviewerID)
	if err != nil {
		return nil, err
	}
	resp := toListingResponse(l)
	if s.notifier != nil && l.Edges.Author != nil {
		_ = s.notifier.NotifyListingDecision(ctx, l.Edges.Author.ID, false, resp, reviewerID)
	}
	return &resp, nil
}

func (s *Service) MarkSold(ctx context.Context, userID, id uuid.UUID, req VersionRequest) (*ListingResponse, error) {
	l, err := s.ownerStatusChange(ctx, userID, id, req.Version, marketplacelisting.StatusApproved, marketplacelisting.StatusSold)
	if err != nil {
		return nil, err
	}
	resp := toListingResponse(l)
	return &resp, nil
}

func (s *Service) WithdrawListing(ctx context.Context, userID, id uuid.UUID, req VersionRequest) (*ListingResponse, error) {
	current, err := s.repo.GetListing(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if current.Edges.Author == nil || current.Edges.Author.ID != userID {
		return nil, ErrNotOwner
	}
	if req.Version < 1 {
		return nil, ErrVersionConflict
	}
	if current.Status != marketplacelisting.StatusPending && current.Status != marketplacelisting.StatusApproved {
		return nil, ErrInvalidStatus
	}
	l, err := s.repo.SetListingStatus(ctx, id, marketplacelisting.StatusWithdrawn, nil, "", req.Version)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("withdraw listing: %w", err)
	}
	l, err = s.repo.GetListing(ctx, l.ID)
	if err != nil {
		return nil, fmt.Errorf("reload listing: %w", err)
	}
	resp := toListingResponse(l)
	return &resp, nil
}

func (s *Service) DeleteListing(ctx context.Context, userID, id uuid.UUID) error {
	current, err := s.repo.GetListing(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("get listing: %w", err)
	}
	isAdmin, err := s.repo.IsSystemAdmin(ctx, userID)
	if err != nil {
		// W7-A5-007 / P3-BE-01: DB blip → 500, не маскировать как Forbidden.
		return fmt.Errorf("check system admin: %w", err)
	}
	isOwner := current.Edges.Author != nil && current.Edges.Author.ID == userID
	if !isOwner && !isAdmin {
		// P3-BE-08 / W7-A5-008: stranger-delete — не admin-кейс.
		return ErrNotOwner
	}
	if err := s.repo.SoftDeleteListing(ctx, id, current.Version); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete listing: %w", err)
	}
	return nil
}

func (s *Service) transitionReview(ctx context.Context, id uuid.UUID, req ReviewListingRequest, status marketplacelisting.Status, reviewerID uuid.UUID) (*ent.MarketplaceListing, error) {
	current, err := s.repo.GetListing(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if current.Status != marketplacelisting.StatusPending {
		return nil, ErrInvalidStatus
	}
	if req.Version < 1 {
		return nil, ErrVersionConflict
	}
	l, err := s.repo.SetListingStatus(ctx, id, status, &reviewerID, req.ReviewNote, req.Version)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("review listing: %w", err)
	}
	return s.repo.GetListing(ctx, l.ID)
}

func (s *Service) ownerStatusChange(ctx context.Context, userID, id uuid.UUID, version int, from, to marketplacelisting.Status) (*ent.MarketplaceListing, error) {
	current, err := s.repo.GetListing(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get listing: %w", err)
	}
	if current.Edges.Author == nil || current.Edges.Author.ID != userID {
		return nil, ErrNotOwner
	}
	if version < 1 {
		return nil, ErrVersionConflict
	}
	if current.Status != from {
		return nil, ErrInvalidStatus
	}
	l, err := s.repo.SetListingStatus(ctx, id, to, nil, "", version)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("status change: %w", err)
	}
	return s.repo.GetListing(ctx, l.ID)
}

func (s *Service) requireAdmin(ctx context.Context, userID uuid.UUID) error {
	ok, err := s.repo.IsSystemAdmin(ctx, userID)
	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func validateListingInput(title, description string, category ListingCategory, priceCents *int64, currency string) error {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(description) == "" {
		return ErrInvalidInput
	}
	// P3-BE-04 / TASK-060: лимиты по рунам, не по байтам (кириллица/emoji).
	if utf8.RuneCountInString(title) > 200 || utf8.RuneCountInString(description) > 2000 {
		return ErrInvalidInput
	}
	if !isValidCategory(category) {
		return ErrInvalidInput
	}
	// W7-A5-006: price ≥ 0; currency allowlist (пусто → default RUB в схеме).
	if priceCents != nil && *priceCents < 0 {
		return ErrInvalidInput
	}
	if currency != "" && !isValidCurrency(currency) {
		return ErrInvalidInput
	}
	return nil
}
