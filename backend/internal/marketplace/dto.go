package marketplace

import (
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

type ListingCategory string

const (
	CategoryEquipment ListingCategory = "equipment"
	CategoryBooks     ListingCategory = "books"
	CategoryServices  ListingCategory = "services"
	CategoryFurniture ListingCategory = "furniture"
	CategoryOther     ListingCategory = "other"
)

type ListingStatus string

const (
	StatusPending   ListingStatus = "pending"
	StatusApproved  ListingStatus = "approved"
	StatusRejected  ListingStatus = "rejected"
	StatusSold      ListingStatus = "sold"
	StatusWithdrawn ListingStatus = "withdrawn"
)

// CreateListingRequest — новое объявление.
type CreateListingRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Category    ListingCategory `json:"category"`
	PriceCents  *int64          `json:"price_cents,omitempty"`
	Currency    string          `json:"currency,omitempty"`
}

// UpdateListingRequest — правка своего объявления.
type UpdateListingRequest struct {
	Title       *string          `json:"title,omitempty"`
	Description *string          `json:"description,omitempty"`
	Category    *ListingCategory `json:"category,omitempty"`
	PriceCents  *int64           `json:"price_cents,omitempty"`
	Currency    *string          `json:"currency,omitempty"`
	Version     int              `json:"version"`
}

// ReviewListingRequest — одобрение/отклонение админом.
type ReviewListingRequest struct {
	Version    int    `json:"version"`
	ReviewNote string `json:"review_note,omitempty"`
}

// VersionRequest — продажа/снятие/отмена.
type VersionRequest struct {
	Version int `json:"version"`
}

// ListingResponse — DTO объявления.
type ListingResponse struct {
	ID           uuid.UUID       `json:"id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Category     ListingCategory `json:"category"`
	Status       ListingStatus   `json:"status"`
	PriceCents   *int64          `json:"price_cents,omitempty"`
	Currency     string          `json:"currency"`
	ReviewNote   string          `json:"review_note,omitempty"`
	AuthorID     uuid.UUID       `json:"author_id"`
	AuthorName   string          `json:"author_name,omitempty"`
	ReviewerID   *uuid.UUID      `json:"reviewer_id,omitempty"`
	ReviewerName string          `json:"reviewer_name,omitempty"`
	Version      int             `json:"version"`
	ReviewedAt   *time.Time      `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ListingListResponse — пагинированный список.
type ListingListResponse struct {
	Items   []ListingResponse `json:"items"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	PerPage int               `json:"per_page"`
}

func toListingResponse(l *ent.MarketplaceListing) ListingResponse {
	resp := ListingResponse{
		ID:          l.ID,
		Title:       l.Title,
		Description: l.Description,
		Category:    ListingCategory(l.Category),
		Status:      ListingStatus(l.Status),
		PriceCents:  l.PriceCents,
		Currency:    l.Currency,
		ReviewNote:  l.ReviewNote,
		Version:     l.Version,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
	if l.ReviewedAt != nil {
		t := l.ReviewedAt.UTC()
		resp.ReviewedAt = &t
	}
	if l.Edges.Author != nil {
		resp.AuthorID = l.Edges.Author.ID
		resp.AuthorName = displayName(l.Edges.Author)
	}
	if l.Edges.Reviewer != nil {
		resp.ReviewerID = &l.Edges.Reviewer.ID
		resp.ReviewerName = displayName(l.Edges.Reviewer)
	}
	return resp
}

func displayName(u *ent.User) string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}

func isValidCategory(c ListingCategory) bool {
	switch c {
	case CategoryEquipment, CategoryBooks, CategoryServices, CategoryFurniture, CategoryOther:
		return true
	default:
		return false
	}
}

func formatPrice(price *int64, currency string) string {
	if price == nil {
		return "Договорная"
	}
	if *price == 0 {
		return "Бесплатно"
	}
	return currency + " " + formatMinorUnits(*price)
}

func formatMinorUnits(cents int64) string {
	rub := cents / 100
	kop := cents % 100
	if kop == 0 {
		return itoa(rub)
	}
	return itoa(rub) + "." + pad2(kop)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

func pad2(n int64) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}
